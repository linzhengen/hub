import { chatServiceOperations as ops } from '@/api/operations';
import { apiRequest, apiStream } from '@/api/request';
import type { components } from '@/api/schema/ai-chat-v1-service';

export type Session = components['schemas']['v1Session'];
export type Message = components['schemas']['v1Message'];
export type CreateSessionRequest = components['schemas']['v1CreateSessionRequest'];
export type CreateSessionResponse = components['schemas']['v1CreateSessionResponse'];
export type ListSessionsResponse = components['schemas']['v1ListSessionsResponse'];
export type DeleteSessionResponse = components['schemas']['v1DeleteSessionResponse'];
export type ListMessagesResponse = components['schemas']['v1ListMessagesResponse'];
export type SendMessageResponse = components['schemas']['v1SendMessageResponse'];
export type ToolCall = components['schemas']['v1ToolCall'];

/** The role the server tags a message with. */
export const ROLE_USER = 'user';
export const ROLE_ASSISTANT = 'assistant';

export const chatService = {
  listSessions: () => apiRequest<ListSessionsResponse>(ops.listSessions),
  createSession: (data: CreateSessionRequest) =>
    apiRequest<CreateSessionResponse>(ops.createSession, { body: data }),
  deleteSession: (id: string) => apiRequest<DeleteSessionResponse>(ops.deleteSession, { path: { id } }),
  listMessages: (id: string) => apiRequest<ListMessagesResponse>(ops.listMessages, { path: { id } }),

  /**
   * Sends a message and yields the answer as it is generated.
   *
   * SendMessage is a server-streaming rpc, so this goes through apiStream: each
   * yielded response carries either the next `delta` of text or `done`.
   */
  sendMessage: (id: string, content: string, signal?: AbortSignal) =>
    apiStream<SendMessageResponse>(ops.sendMessage, { path: { id }, body: { content }, signal }),
};
