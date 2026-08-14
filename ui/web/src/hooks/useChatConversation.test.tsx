import React from 'react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { useChatConversation } from '@/hooks/useChatConversation';
import { chatService } from '@/services/chat';
import type { Message, SendMessageResponse } from '@/services/chat';

vi.mock('@/services/chat', async (importOriginal) => {
  const actual = await importOriginal<typeof import('@/services/chat')>();
  return {
    ...actual,
    chatService: {
      listMessages: vi.fn(),
      sendMessage: vi.fn(),
    },
  };
});

const listMessages = vi.mocked(chatService.listMessages);
const sendMessage = vi.mocked(chatService.sendMessage);

/** Turns a list of responses into the async iterable sendMessage returns. */
const streamOf = (responses: readonly SendMessageResponse[]) =>
  (async function* () {
    for (const response of responses) yield response;
  })() as ReturnType<typeof chatService.sendMessage>;

const wrapper = ({ children }: { children: React.ReactNode }) => {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return <QueryClientProvider client={client}>{children}</QueryClientProvider>;
};

const stored: Message[] = [{ id: 'm1', role: 'user', content: 'hi' }];

beforeEach(() => {
  listMessages.mockResolvedValue({ messages: stored });
});

afterEach(() => {
  vi.clearAllMocks();
});

describe('useChatConversation', () => {
  it('セッション未選択なら履歴を取りに行かない', () => {
    renderHook(() => useChatConversation(null), { wrapper });

    expect(listMessages).not.toHaveBeenCalled();
  });

  it('delta を積み上げて途中経過を見せる', async () => {
    sendMessage.mockReturnValue(streamOf([{ delta: 'He' }, { delta: 'llo' }, { done: true }]));

    const { result } = renderHook(() => useChatConversation('s1'), { wrapper });
    await waitFor(() => expect(result.current.messages).toHaveLength(1));

    await act(async () => {
      await result.current.send('こんにちは');
    });

    expect(sendMessage).toHaveBeenCalledWith('s1', 'こんにちは', expect.any(AbortSignal));
    // ストリームが終われば、表示はサーバーの保存内容に委ねる。
    expect(result.current.streamingText).toBe('');
    expect(result.current.pendingContent).toBeNull();
    expect(result.current.isStreaming).toBe(false);
  });

  it('送信中は入力をそのまま表示し続ける', async () => {
    let release: (() => void) | undefined;
    const blocked = new Promise<void>((resolve) => {
      release = resolve;
    });
    sendMessage.mockReturnValue(
      (async function* () {
        yield { delta: 'partial' };
        await blocked;
        yield { done: true };
      })() as ReturnType<typeof chatService.sendMessage>,
    );

    const { result } = renderHook(() => useChatConversation('s1'), { wrapper });
    await waitFor(() => expect(result.current.messages).toHaveLength(1));

    let sending: Promise<void>;
    act(() => {
      sending = result.current.send('待って');
    });

    await waitFor(() => expect(result.current.streamingText).toBe('partial'));
    expect(result.current.pendingContent).toBe('待って');
    expect(result.current.isStreaming).toBe(true);

    await act(async () => {
      release?.();
      await sending;
    });
  });

  it('ストリームのエラーを表面化し、履歴を取り直す', async () => {
    sendMessage.mockReturnValue(
      (async function* () {
        yield { delta: 'partial' };
        throw new Error('stream error');
      })() as ReturnType<typeof chatService.sendMessage>,
    );

    const { result } = renderHook(() => useChatConversation('s1'), { wrapper });
    await waitFor(() => expect(result.current.messages).toHaveLength(1));
    listMessages.mockClear();

    await act(async () => {
      await result.current.send('落ちる');
    });

    expect(result.current.error).toBe('stream error');
    expect(result.current.isStreaming).toBe(false);
    // ユーザーの発言はモデル呼び出し前にサーバーへ保存されるため、取り直す。
    await waitFor(() => expect(listMessages).toHaveBeenCalled());
  });

  it('空の入力は送らない', async () => {
    const { result } = renderHook(() => useChatConversation('s1'), { wrapper });

    await act(async () => {
      await result.current.send('   ');
    });

    expect(sendMessage).not.toHaveBeenCalled();
  });

  it('セッションを切り替えたら進行中の応答を捨てる', async () => {
    sendMessage.mockReturnValue(
      (async function* () {
        yield { delta: 'partial' };
        await new Promise(() => undefined); // 終わらないストリーム
      })() as ReturnType<typeof chatService.sendMessage>,
    );

    const { result, rerender } = renderHook(({ id }) => useChatConversation(id), {
      wrapper,
      initialProps: { id: 's1' as string | null },
    });
    await waitFor(() => expect(result.current.messages).toHaveLength(1));

    act(() => {
      void result.current.send('置いていく');
    });
    await waitFor(() => expect(result.current.streamingText).toBe('partial'));

    rerender({ id: 's2' });

    expect(result.current.streamingText).toBe('');
    expect(result.current.pendingContent).toBeNull();
  });
});
