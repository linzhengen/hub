import { useCallback, useEffect, useRef, useState } from 'react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { chatService } from '@/services/chat';
import type { Message } from '@/services/chat';

export const chatMessagesQueryKey = (sessionId: string) => ['chatMessages', sessionId];

/**
 * Holds one session's conversation, including the answer still being streamed.
 *
 * The persisted messages come from the server; the turn in flight does not
 * exist there yet, so it is kept here as `pendingContent` (what the user just
 * sent) and `streamingText` (what has arrived so far). Once the stream ends
 * both are dropped and the query is refetched, so the rendered history always
 * settles on what the server actually stored.
 */
/** The turn in flight, tagged with the session it belongs to. */
interface InFlight {
  sessionId: string;
  content: string;
  answer: string;
}

/** A failed turn, tagged the same way. */
interface Failure {
  sessionId: string;
  message: string;
}

export const useChatConversation = (sessionId: string | null) => {
  const queryClient = useQueryClient();
  const [inFlight, setInFlight] = useState<InFlight | null>(null);
  const [failure, setFailure] = useState<Failure | null>(null);
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

  const send = useCallback(
    async (content: string) => {
      if (sessionId === null || content.trim() === '') return;

      abort();
      const controller = new AbortController();
      abortRef.current = controller;

      setInFlight({ sessionId, content, answer: '' });
      setFailure(null);

      try {
        let answer = '';
        for await (const response of chatService.sendMessage(sessionId, content, controller.signal)) {
          if (response.delta) {
            answer += response.delta;
            setInFlight({ sessionId, content, answer });
          }
          if (response.done) break;
        }
      } catch (cause) {
        // An aborted request is the user's own doing, not a failure to report.
        if (!controller.signal.aborted) {
          setFailure({
            sessionId,
            message: cause instanceof Error ? cause.message : 'failed to send the message',
          });
        }
      } finally {
        if (abortRef.current === controller) abortRef.current = null;
        setInFlight(null);
        // The server stores the user's message before it calls the model, so it
        // is on the server even when the answer failed - refetch either way
        // rather than guessing which local state to keep.
        if (!controller.signal.aborted) {
          await queryClient.invalidateQueries({ queryKey: chatMessagesQueryKey(sessionId) });
        }
      }
    },
    [sessionId, abort, queryClient],
  );

  // Tagged state belonging to another session is not this session's business.
  const active = inFlight?.sessionId === sessionId ? inFlight : null;
  const activeFailure = failure?.sessionId === sessionId ? failure : null;

  return {
    messages,
    isLoading,
    /** The message just sent, not yet persisted - null when nothing is in flight. */
    pendingContent: active?.content ?? null,
    /** The answer so far. Empty until the first delta arrives. */
    streamingText: active?.answer ?? '',
    isStreaming: active !== null,
    error: activeFailure?.message ?? null,
    send,
    abort,
  };
};
