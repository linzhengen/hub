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
      confirmToolCall: vi.fn(),
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
const confirmToolCall = vi.mocked(chatService.confirmToolCall);

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

  it('ツールを呼んでいる間その内容を表示する', async () => {
    let release: (() => void) | undefined;
    const blocked = new Promise<void>((resolve) => {
      release = resolve;
    });
    sendMessage.mockReturnValue(
      (async function* () {
        yield { toolCall: { name: 'list_group', arguments: '{"groupName":"admin"}' } };
        await blocked;
        yield { done: true };
      })() as ReturnType<typeof chatService.sendMessage>,
    );

    renderChat();
    await screen.findByText('First chat');

    fireEvent.change(await screen.findByLabelText('Message'), { target: { value: 'admin には誰がいる?' } });
    fireEvent.click(screen.getByRole('button', { name: /send/i }));

    expect(await screen.findByText('list_group')).toBeTruthy();
    expect(await screen.findByText('{"groupName":"admin"}')).toBeTruthy();

    release?.();
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

// sendDraft types a message and sends it, which every test below does first.
const sendDraft = async (text: string) => {
  await screen.findByText('First chat');
  fireEvent.change(await screen.findByLabelText('Message'), { target: { value: text } });
  fireEvent.click(screen.getByRole('button', { name: /send/i }));
};

describe('a change the assistant proposes', () => {
  const proposal = {
    id: 'p1',
    name: 'add_users_to_group',
    arguments: '{"id":"admin-group","userIds":["u1","u2"]}',
  };

  beforeEach(() => {
    listSessions.mockResolvedValue({ sessions });
    listMessages.mockResolvedValue({ messages: [] });
    sendMessage.mockReturnValue(streamOf([{ toolProposal: proposal }]));
  });

  // Approving something you cannot read is not approving it, so the operation
  // and its arguments have to be on screen - not a summary of them.
  it('shows the operation and the arguments it will be called with', async () => {
    renderChat();
    await sendDraft('add them to the admin group');

    await waitFor(() => {
      expect(screen.getByText(/The assistant wants to make a change/)).toBeTruthy();
    });
    const card = screen.getByText(/add_users_to_group/);
    expect(card.textContent).toContain('admin-group');
    expect(card.textContent).toContain('u1');
    expect(card.textContent).toContain('u2');
  });

  it('says plainly that nothing has been changed yet', async () => {
    renderChat();
    await sendDraft('add them to the admin group');

    await waitFor(() => {
      expect(screen.getByText(/Nothing has been changed/)).toBeTruthy();
    });
    expect(confirmToolCall).not.toHaveBeenCalled();
  });

  it('runs the change only when approved', async () => {
    confirmToolCall.mockReturnValue(
      streamOf([{ delta: 'Added them.' }, { done: true }]) as ReturnType<
        typeof chatService.confirmToolCall
      >,
    );
    renderChat();
    await sendDraft('add them to the admin group');

    await waitFor(() => expect(screen.getByRole('button', { name: /Approve/ })).toBeTruthy());
    fireEvent.click(screen.getByRole('button', { name: /Approve/ }));

    await waitFor(() => {
      expect(confirmToolCall).toHaveBeenCalledWith('p1', true, expect.anything());
    });
  });

  it('declines without running the change', async () => {
    confirmToolCall.mockReturnValue(
      streamOf([{ delta: 'Left it alone.' }, { done: true }]) as ReturnType<
        typeof chatService.confirmToolCall
      >,
    );
    renderChat();
    await sendDraft('add them to the admin group');

    await waitFor(() => expect(screen.getByRole('button', { name: /Decline/ })).toBeTruthy());
    fireEvent.click(screen.getByRole('button', { name: /Decline/ }));

    await waitFor(() => {
      expect(confirmToolCall).toHaveBeenCalledWith('p1', false, expect.anything());
    });
  });

  // Typing past an unanswered proposal would leave it hanging, and the next
  // message would arrive with the conversation still stopped.
  it('does not accept another message until the change is answered', async () => {
    renderChat();
    await sendDraft('add them to the admin group');

    await waitFor(() => expect(screen.getByText(/Nothing has been changed/)).toBeTruthy());
    expect(screen.getByLabelText('Message')).toHaveProperty('disabled', true);
  });
});
