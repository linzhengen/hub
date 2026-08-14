import { useCallback, useEffect, useRef, useState } from 'react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { chatService } from '@/services/chat';
import type { Message, SendMessageResponse, ToolCall, ToolProposal } from '@/services/chat';

export const chatMessagesQueryKey = (sessionId: string) => ['chatMessages', sessionId];

/** The turn in flight, tagged with the session it belongs to. */
interface InFlight {
  sessionId: string;
  /** What the user sent. Empty while the assistant carries on after a decision. */
  content: string;
  answer: string;
  /** The lookups reported so far, oldest first. */
  tools: ToolCall[];
}

/** A failed turn, tagged the same way. */
interface Failure {
  sessionId: string;
  message: string;
}

/** A change waiting on the user, tagged the same way. */
interface Awaiting {
  sessionId: string;
  proposal: ToolProposal;
  /** What the assistant said on its way to asking. */
  answer: string;
}

/**
 * Holds one session's conversation, including the answer still being streamed
 * and any change waiting to be approved.
 *
 * The persisted messages come from the server; the turn in flight does not
 * exist there yet, so it is kept here as `pendingContent` (what the user just
 * sent) and `streamingText` (what has arrived so far). Once the stream ends
 * both are dropped and the query is refetched, so the rendered history always
 * settles on what the server actually stored.
 *
 * A stream can also end by asking permission. The server cannot read from a
 * server-streaming rpc, so approving is a second request rather than a reply
 * into the first: the proposal is held here until the user answers, and the
 * answer opens a new stream that carries on where this one stopped.
 */
export const useChatConversation = (sessionId: string | null) => {
  const queryClient = useQueryClient();
  const [inFlight, setInFlight] = useState<InFlight | null>(null);
  const [failure, setFailure] = useState<Failure | null>(null);
  const [awaiting, setAwaiting] = useState<Awaiting | null>(null);
  const abortRef = useRef<AbortController | null>(null);

  const { data: messages = [], isLoading } = useQuery<Message[]>({
    queryKey: chatMessagesQueryKey(sessionId ?? ''),
    queryFn: async () => (await chatService.listMessages(sessionId as string)).messages ?? [],
    enabled: sessionId !== null,
  });

  const abort = useCallback(() => {
    abortRef.current?.abort();
    abortRef.current = null;
  }, []);

  // Leaving a session mid-answer must stop the request rather than let it stream
  // into a component nobody is rendering. Only the cleanup runs here - the
  // transient state below is tagged with its session and derived, so switching
  // sessions drops it without an effect writing state.
  useEffect(() => abort, [sessionId, abort]);

  /**
   * Consumes one answer stream.
   *
   * Sending a message and answering a proposal differ only in which request
   * opens the stream: what arrives on it, and what it means, is the same. They
   * share this so the conversation cannot come to behave differently either
   * side of an approval.
   */
  const consume = useCallback(
    async (
      session: string,
      content: string,
      open: (signal: AbortSignal) => AsyncIterable<SendMessageResponse>,
    ) => {
      abort();
      const controller = new AbortController();
      abortRef.current = controller;

      setInFlight({ sessionId: session, content, answer: '', tools: [] });
      setFailure(null);
      setAwaiting(null);

      let proposed: Awaiting | null = null;
      try {
        let answer = '';
        const tools: ToolCall[] = [];
        for await (const response of open(controller.signal)) {
          // A frame reports one thing: a piece of the answer, a lookup, or a
          // change to approve. They accumulate independently.
          if (response.toolCall) {
            tools.push(response.toolCall);
            setInFlight({ sessionId: session, content, answer, tools: [...tools] });
          }
          if (response.delta) {
            answer += response.delta;
            setInFlight({ sessionId: session, content, answer, tools: [...tools] });
          }
          if (response.toolProposal) {
            // The stream ends here. Nothing has been changed, and nothing will
            // be until the user answers.
            proposed = { sessionId: session, proposal: response.toolProposal, answer };
            break;
          }
          if (response.done) break;
        }
      } catch (cause) {
        // An aborted request is the user's own doing, not a failure to report.
        if (!controller.signal.aborted) {
          setFailure({
            sessionId: session,
            message: cause instanceof Error ? cause.message : 'failed to send the message',
          });
        }
      } finally {
        if (abortRef.current === controller) abortRef.current = null;
        setInFlight(null);
        if (proposed !== null) setAwaiting(proposed);
        // The server stores the user's message before it calls the model, so it
        // is on the server even when the answer failed - refetch either way
        // rather than guessing which local state to keep.
        if (!controller.signal.aborted) {
          await queryClient.invalidateQueries({ queryKey: chatMessagesQueryKey(session) });
        }
      }
    },
    [abort, queryClient],
  );

  const send = useCallback(
    async (content: string) => {
      if (sessionId === null || content.trim() === '') return;
      await consume(sessionId, content, (signal) => chatService.sendMessage(sessionId, content, signal));
    },
    [sessionId, consume],
  );

  /**
   * Answers the change the assistant proposed.
   *
   * `content` is empty because the user did not type anything: the transcript
   * shows the proposal card and its outcome, not a synthetic "yes" turn that
   * they never wrote.
   */
  const confirm = useCallback(
    async (approved: boolean) => {
      if (sessionId === null || awaiting === null || awaiting.sessionId !== sessionId) return;
      const proposalId = awaiting.proposal.id;
      if (proposalId === undefined) return;
      await consume(sessionId, '', (signal) => chatService.confirmToolCall(proposalId, approved, signal));
    },
    [sessionId, awaiting, consume],
  );

  // Tagged state belonging to another session is not this session's business.
  const active = inFlight?.sessionId === sessionId ? inFlight : null;
  const activeFailure = failure?.sessionId === sessionId ? failure : null;
  const activeAwaiting = awaiting?.sessionId === sessionId ? awaiting : null;

  return {
    messages,
    isLoading,
    /** The message just sent, not yet persisted - null when nothing is in flight. */
    pendingContent: active !== null && active.content !== '' ? active.content : null,
    /** The answer so far. Empty until the first delta arrives. */
    streamingText: active?.answer ?? '',
    /** The lookups made while answering, oldest first. */
    toolCalls: active?.tools ?? [],
    isStreaming: active !== null,
    /** The change waiting on the user, or null. Nothing has been changed yet. */
    proposal: activeAwaiting?.proposal ?? null,
    /** What the assistant said before asking, kept on screen beside the card. */
    proposalPreamble: activeAwaiting?.answer ?? '',
    error: activeFailure?.message ?? null,
    send,
    confirm,
    abort,
  };
};
