import React from 'react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { App } from 'antd';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { Chat } from '@/pages/Chat';
import { chatService } from '@/services/chat';
import type { SendMessageResponse, Session } from '@/services/chat';

vi.mock('@/services/chat', async (importOriginal) => {
  const actual = await importOriginal<typeof import('@/services/chat')>();
  return {
    ...actual,
    chatService: {
      listSessions: vi.fn(),
      createSession: vi.fn(),
      deleteSession: vi.fn(),
      listMessages: vi.fn(),
      sendMessage: vi.fn(),
    },
  };
});

const toastSuccess = vi.hoisted(() => vi.fn());
const toastError = vi.hoisted(() => vi.fn());
vi.mock('@/hooks/useNotify', () => ({
  useNotify: () => ({ success: toastSuccess, error: toastError }),
}));

const listSessions = vi.mocked(chatService.listSessions);
const listMessages = vi.mocked(chatService.listMessages);
const sendMessage = vi.mocked(chatService.sendMessage);
const createSession = vi.mocked(chatService.createSession);

const streamOf = (responses: readonly SendMessageResponse[]) =>
  (async function* () {
    for (const response of responses) yield response;
  })() as ReturnType<typeof chatService.sendMessage>;

const sessions: Session[] = [{ id: 's1', title: 'First chat' }];

const renderChat = () => {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={client}>
      <App>
        <Chat />
      </App>
    </QueryClientProvider>,
  );
};

beforeEach(() => {
  listSessions.mockResolvedValue({ sessions });
  listMessages.mockResolvedValue({ messages: [] });
});

afterEach(() => {
  vi.clearAllMocks();
});

describe('Chat', () => {
  it('セッション一覧を表示し、最新のものを選ぶ', async () => {
    renderChat();

    expect(await screen.findByText('First chat')).toBeTruthy();
    await waitFor(() => expect(listMessages).toHaveBeenCalledWith('s1'));
  });

  it('送信した内容と応答を順に表示する', async () => {
    sendMessage.mockReturnValue(streamOf([{ delta: 'Hello ' }, { delta: 'there' }, { done: true }]));
    listMessages
      .mockResolvedValueOnce({ messages: [] })
      .mockResolvedValue({
        messages: [
          { id: 'm1', role: 'user', content: 'やあ' },
          { id: 'm2', role: 'assistant', content: 'Hello there' },
        ],
      });

    renderChat();
    await screen.findByText('First chat');

    fireEvent.change(await screen.findByLabelText('Message'), { target: { value: 'やあ' } });
    fireEvent.click(screen.getByRole('button', { name: /send/i }));

    expect(await screen.findByText('やあ')).toBeTruthy();
    expect(await screen.findByText('Hello there')).toBeTruthy();
    expect(sendMessage).toHaveBeenCalledWith('s1', 'やあ', expect.any(AbortSignal));
  });

  it('ストリームが失敗したらエラーを表示する', async () => {
    // 最初の読み取りで失敗するストリーム。generator ではなく手書きの
    // async iterable にしているのは、yield を一度もしない generator が
    // require-yield に触れるため。
    sendMessage.mockReturnValue({
      [Symbol.asyncIterator]: () => ({
        next: () => Promise.reject(new Error('simulated upstream failure')),
      }),
    } as unknown as ReturnType<typeof chatService.sendMessage>);

    renderChat();
    await screen.findByText('First chat');

    fireEvent.change(await screen.findByLabelText('Message'), { target: { value: 'こわれて' } });
    fireEvent.click(screen.getByRole('button', { name: /send/i }));

    expect(await screen.findByText('simulated upstream failure')).toBeTruthy();
  });

  it('空のままでは送信できない', async () => {
    renderChat();
    await screen.findByText('First chat');

    expect(screen.getByRole('button', { name: /send/i })).toHaveProperty('disabled', true);
  });

  it('新規セッションを作って選択する', async () => {
    createSession.mockResolvedValue({ session: { id: 's2', title: 'New chat' } });
    listSessions.mockResolvedValueOnce({ sessions }).mockResolvedValue({
      sessions: [{ id: 's2', title: 'New chat' }, ...sessions],
    });

    renderChat();
    await screen.findByText('First chat');

    fireEvent.click(screen.getByRole('button', { name: /new chat/i }));

    await waitFor(() => expect(createSession).toHaveBeenCalled());
    await waitFor(() => expect(listMessages).toHaveBeenCalledWith('s2'));
  });

  it('セッションが無ければ案内を出す', async () => {
    listSessions.mockResolvedValue({ sessions: [] });

    renderChat();

    expect(await screen.findByText('Select a session, or start a new chat')).toBeTruthy();
  });
});
