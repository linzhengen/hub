import React, { useEffect, useMemo, useRef, useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Alert, Button, Empty, Input, Spin, Tag, Typography } from 'antd';
import { DeleteOutlined, PlusOutlined, SendOutlined } from '@ant-design/icons';
import { useNotify } from '@/hooks/useNotify';
import { useDangerConfirm } from '@/hooks/useDangerConfirm';
import { useChatConversation } from '@/hooks/useChatConversation';
import { chatService, ROLE_ASSISTANT } from '@/services/chat';
import type { Message, Session, ToolCall } from '@/services/chat';

const SESSIONS_QUERY_KEY = ['chatSessions'];

/**
 * One turn in the transcript.
 *
 * The content is rendered as preformatted text rather than Markdown: the model
 * answers in Markdown, but rendering it needs a dependency this project has not
 * taken yet, and showing the source is honest where a half-built renderer would
 * silently drop formatting. `whitespace-pre-wrap` at least keeps the line breaks
 * and indentation the model intended.
 */
const Turn: React.FC<{ role: string; content: string; pending?: boolean }> = ({
  role,
  content,
  pending = false,
}) => {
  const isAssistant = role === ROLE_ASSISTANT;
  return (
    <div className={`flex ${isAssistant ? 'justify-start' : 'justify-end'} mb-4`}>
      <div
        className={`max-w-[80%] rounded-lg px-4 py-3 ${
          isAssistant ? 'bg-gray-100 dark:bg-gray-800' : 'bg-blue-600 text-white'
        }`}
      >
        <div className="whitespace-pre-wrap break-words font-mono text-sm">
          {content}
          {pending && <span className="ml-1 animate-pulse">▍</span>}
        </div>
      </div>
    </div>
  );
};

/**
 * The lookups the assistant made while answering.
 *
 * Shown while the answer streams so a pause has a visible reason, and dropped
 * once the turn is stored - the saved conversation is the answer, not how it
 * was reached. Arguments are included because "which group is it looking at" is
 * the part that makes the line informative; results are not, since the answer
 * summarises them anyway.
 */
const ToolTrace: React.FC<{ calls: ToolCall[] }> = ({ calls }) => {
  if (calls.length === 0) return null;
  return (
    <div className="mb-4 flex justify-start">
      <ul className="m-0 list-none space-y-1 p-0 text-xs text-gray-500 dark:text-gray-400">
        {calls.map((call, i) => (
          <li key={`${call.name}-${i}`} className="font-mono">
            <Spin size="small" className="mr-2" />
            {call.name}
            {call.arguments && call.arguments !== '{}' && (
              <span className="ml-1 opacity-70">{call.arguments}</span>
            )}
          </li>
        ))}
      </ul>
    </div>
  );
};

export function Chat() {
  const queryClient = useQueryClient();
  const notify = useNotify();
  const confirmDanger = useDangerConfirm();
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [draft, setDraft] = useState('');
  const transcriptEndRef = useRef<HTMLDivElement | null>(null);

  const { data: sessions = [], isLoading: isSessionsLoading } = useQuery<Session[]>({
    queryKey: SESSIONS_QUERY_KEY,
    queryFn: async () => (await chatService.listSessions()).sessions ?? [],
  });

  // Which session is open follows from the user's choice and the list, so it is
  // derived at render rather than copied into state by an effect: `selectedId`
  // holds only the explicit choice, and until there is one the newest session is
  // shown so the page is not empty for a user who already has a history.
  const openId = selectedId ?? sessions[0]?.id ?? null;

  const {
    messages,
    isLoading: isMessagesLoading,
    pendingContent,
    streamingText,
    toolCalls,
    isStreaming,
    error,
    send,
  } = useChatConversation(openId);

  // Follow the answer as it streams in.
  useEffect(() => {
    transcriptEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages, streamingText, pendingContent, toolCalls]);

  const createMutation = useMutation({
    mutationFn: () => chatService.createSession({ title: 'New chat' }),
    onSuccess: async (response) => {
      await queryClient.invalidateQueries({ queryKey: SESSIONS_QUERY_KEY });
      setSelectedId(response.session?.id ?? null);
    },
    onError: () => notify.error('Failed to create the chat session'),
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) => chatService.deleteSession(id),
    onSuccess: async (_, id) => {
      await queryClient.invalidateQueries({ queryKey: SESSIONS_QUERY_KEY });
      if (openId === id) setSelectedId(null);
      notify.success('Deleted the chat session');
    },
    onError: () => notify.error('Failed to delete the chat session'),
  });

  const transcript = useMemo<Message[]>(() => messages, [messages]);

  const handleSend = async () => {
    const content = draft.trim();
    if (content === '' || isStreaming) return;
    setDraft('');
    await send(content);
  };

  return (
    <div className="flex h-[calc(100vh-12rem)] gap-4">
      <aside className="flex w-64 shrink-0 flex-col rounded-lg border border-gray-200 dark:border-gray-700">
        <div className="border-b border-gray-200 p-3 dark:border-gray-700">
          <Button
            type="primary"
            icon={<PlusOutlined />}
            block
            loading={createMutation.isPending}
            onClick={() => createMutation.mutate()}
          >
            New chat
          </Button>
        </div>
        <div className="flex-1 overflow-y-auto">
          {isSessionsLoading ? (
            <div className="flex justify-center p-6">
              <Spin />
            </div>
          ) : (
            <ul className="m-0 list-none p-0">
              {sessions.length === 0 && <Empty description="No sessions yet" className="mt-6" />}
              {sessions.map((session) => (
                <li
                  key={session.id}
                  className={`flex items-center justify-between gap-1 border-b border-gray-100 px-3 py-2 dark:border-gray-800 ${
                    session.id === openId ? 'bg-gray-100 dark:bg-gray-800' : ''
                  }`}
                >
                  <button
                    type="button"
                    className="min-w-0 flex-1 cursor-pointer truncate bg-transparent text-left"
                    onClick={() => setSelectedId(session.id ?? null)}
                  >
                    {session.title || 'Untitled'}
                  </button>
                  <Button
                    type="text"
                    size="small"
                    danger
                    icon={<DeleteOutlined />}
                    aria-label={`Delete ${session.title || 'session'}`}
                    onClick={() =>
                      confirmDanger(
                        {
                          title: 'Delete this chat session?',
                          content: 'The session and all its messages are removed. This cannot be undone.',
                        },
                        () => deleteMutation.mutate(session.id as string),
                      )
                    }
                  />
                </li>
              ))}
            </ul>
          )}
        </div>
      </aside>

      <section className="flex min-w-0 flex-1 flex-col rounded-lg border border-gray-200 dark:border-gray-700">
        {openId === null ? (
          <div className="flex flex-1 items-center justify-center">
            <Empty description="Select a session, or start a new chat" />
          </div>
        ) : (
          <>
            <div className="flex-1 overflow-y-auto p-4">
              {isMessagesLoading ? (
                <div className="flex justify-center p-6">
                  <Spin />
                </div>
              ) : (
                <>
                  {transcript.length === 0 && pendingContent === null && (
                    <div className="flex h-full items-center justify-center">
                      <Empty description="Send a message to start" />
                    </div>
                  )}
                  {transcript.map((message) => (
                    <Turn key={message.id} role={message.role ?? ''} content={message.content ?? ''} />
                  ))}
                  {pendingContent !== null && <Turn role="user" content={pendingContent} />}
                  {isStreaming && <ToolTrace calls={toolCalls} />}
                  {isStreaming && <Turn role={ROLE_ASSISTANT} content={streamingText} pending />}
                  <div ref={transcriptEndRef} />
                </>
              )}
            </div>

            {error !== null && (
              <div className="px-4 pb-2">
                <Alert type="error" showIcon message={error} />
              </div>
            )}

            <div className="border-t border-gray-200 p-3 dark:border-gray-700">
              <div className="flex gap-2">
                <Input.TextArea
                  value={draft}
                  onChange={(event) => setDraft(event.target.value)}
                  onPressEnter={(event) => {
                    if (event.shiftKey) return;
                    event.preventDefault();
                    void handleSend();
                  }}
                  autoSize={{ minRows: 1, maxRows: 6 }}
                  placeholder="Send a message (Enter to send, Shift+Enter for a new line)"
                  aria-label="Message"
                  disabled={isStreaming}
                />
                <Button
                  type="primary"
                  icon={<SendOutlined />}
                  loading={isStreaming}
                  disabled={draft.trim() === ''}
                  onClick={() => void handleSend()}
                >
                  Send
                </Button>
              </div>
              <Typography.Text type="secondary" className="mt-1 block text-xs">
                <Tag color="default">mock</Tag>
                Local development answers from the mock backend when ANTHROPIC_MOCK=true.
              </Typography.Text>
            </div>
          </>
        )}
      </section>
    </div>
  );
}
