import { afterEach, describe, expect, it, vi } from 'vitest';
import type { ApiOperation } from '@/api/operations';
import { chatServiceOperations, userServiceOperations } from '@/api/operations';
import { apiRequest, ApiStreamError, apiStream, buildPath, buildQueryString } from '@/api/request';
import { API_BASE_URL } from '@/lib/api-client';

const mockFetch = () =>
  vi.spyOn(globalThis, 'fetch').mockResolvedValue({
    ok: true,
    status: 200,
    statusText: 'OK',
    json: vi.fn().mockResolvedValue({}),
  } as unknown as Response);

afterEach(() => {
  vi.restoreAllMocks();
});

describe('buildQueryString', () => {
  it('繰り返しフィールドをキーの繰り返しとして送る', () => {
    // grpc-gateway はカンマ区切りを単一要素として解釈するため、キーを繰り返す。
    expect(buildQueryString({ userIds: ['a', 'b'] })).toBe('?userIds=a&userIds=b');
  });

  it('undefined と null のパラメータを除外する', () => {
    expect(buildQueryString({ limit: 10, offset: undefined, userName: null })).toBe('?limit=10');
  });

  it('パラメータが無ければ空文字を返す', () => {
    expect(buildQueryString(undefined)).toBe('');
    expect(buildQueryString({})).toBe('');
    expect(buildQueryString({ limit: undefined })).toBe('');
  });

  it('値をエスケープする', () => {
    expect(buildQueryString({ userName: 'a b&c' })).toBe('?userName=a+b%26c');
  });
});

describe('buildPath', () => {
  it('プレースホルダを置換する', () => {
    expect(buildPath(userServiceOperations.getUser, { id: 'u1' })).toBe('/users/u1');
  });

  it('パス値をエスケープする', () => {
    expect(buildPath(userServiceOperations.getUser, { id: 'a/b' })).toBe('/users/a%2Fb');
  });

  it('パスパラメータが欠けていれば投げる', () => {
    expect(() => buildPath(userServiceOperations.getUser, {})).toThrow('missing path parameter "id"');
  });

  it('プレースホルダの無いパスはそのまま返す', () => {
    expect(buildPath(userServiceOperations.listUser)).toBe('/users');
  });
});

describe('apiRequest', () => {
  it('生成された verb とパスでリクエストする', async () => {
    const fetchSpy = mockFetch();

    await apiRequest(userServiceOperations.deleteUser, { path: { id: 'u1' } });

    expect(fetchSpy).toHaveBeenCalledWith(
      `${API_BASE_URL}/users/u1`,
      expect.objectContaining({ method: 'DELETE' }),
    );
  });

  it('クエリとボディを組み立てる', async () => {
    const fetchSpy = mockFetch();

    await apiRequest(userServiceOperations.listUser, { query: { limit: 10, groupIds: ['g1', 'g2'] } });
    expect(fetchSpy).toHaveBeenCalledWith(`${API_BASE_URL}/users?limit=10&groupIds=g1&groupIds=g2`, expect.anything());

    await apiRequest(userServiceOperations.createUser, { body: { username: 'taro' } });
    expect(fetchSpy).toHaveBeenLastCalledWith(
      `${API_BASE_URL}/users`,
      expect.objectContaining({ method: 'POST', body: '{"username":"taro"}' }),
    );
  });

  it('ボディの無い書き込みには空オブジェクトを送る', async () => {
    const fetchSpy = mockFetch();

    await apiRequest(userServiceOperations.sendMeVerifyEmail);

    expect(fetchSpy).toHaveBeenCalledWith(
      `${API_BASE_URL}/me/verify-email`,
      expect.objectContaining({ method: 'POST', body: '{}' }),
    );
  });

  it('GET にはボディを付けない', async () => {
    const fetchSpy = mockFetch();

    await apiRequest(userServiceOperations.getMe);

    expect((fetchSpy.mock.calls[0][1] as RequestInit).body).toBeUndefined();
  });
});

describe('生成された操作テーブル', () => {
  it('pathParams はパスのプレースホルダと一致する', () => {
    const operations: ApiOperation[] = Object.values(userServiceOperations);

    for (const operation of operations) {
      const placeholders = [...operation.path.matchAll(/\{(\w+)}/g)].map(([, name]) => name);
      expect(placeholders).toEqual([...operation.pathParams]);
    }
  });
});

/** Builds a Response whose body streams the given chunks, as fetch would. */
const mockStream = (chunks: readonly string[]) => {
  const encoder = new TextEncoder();
  const cancel = vi.fn().mockResolvedValue(undefined);
  let i = 0;

  const body = {
    getReader: () => ({
      read: vi.fn().mockImplementation(() => {
        if (i >= chunks.length) return Promise.resolve({ done: true, value: undefined });
        const value = encoder.encode(chunks[i]);
        i += 1;
        return Promise.resolve({ done: false, value });
      }),
      cancel,
    }),
  };

  const fetchSpy = vi.spyOn(globalThis, 'fetch').mockResolvedValue({
    ok: true,
    status: 200,
    statusText: 'OK',
    body,
  } as unknown as Response);

  return { fetchSpy, cancel };
};

const collect = async <T>(stream: AsyncGenerator<T>): Promise<T[]> => {
  const out: T[] = [];
  for await (const message of stream) out.push(message);
  return out;
};

interface Delta {
  delta?: string;
  done?: boolean;
}

describe('apiStream', () => {
  it('grpc-gateway のラッパを剥がして 1 メッセージずつ返す', async () => {
    mockStream(['{"result":{"delta":"He"}}\n{"result":{"delta":"llo"}}\n{"result":{"done":true}}\n']);

    expect(await collect(apiStream<Delta>(chatServiceOperations.sendMessage, { path: { id: 's1' } }))).toEqual([
      { delta: 'He' },
      { delta: 'llo' },
      { done: true },
    ]);
  });

  it('チャンクの途中で切れた行を次のチャンクとつなぐ', async () => {
    // ネットワークの区切りとメッセージの区切りは一致しないので、
    // 行が 2 つのチャンクにまたがっても落としてはいけない。
    mockStream(['{"result":{"del', 'ta":"He"}}\n{"result":{"delta":"llo"}}\n']);

    expect(await collect(apiStream<Delta>(chatServiceOperations.sendMessage, { path: { id: 's1' } }))).toEqual([
      { delta: 'He' },
      { delta: 'llo' },
    ]);
  });

  it('末尾の改行が無い最後のメッセージも返す', async () => {
    mockStream(['{"result":{"delta":"Hi"}}']);

    expect(await collect(apiStream<Delta>(chatServiceOperations.sendMessage, { path: { id: 's1' } }))).toEqual([
      { delta: 'Hi' },
    ]);
  });

  it('途中のエラーフレームを ApiStreamError として投げる', async () => {
    mockStream(['{"result":{"delta":"Hi"}}\n{"error":{"message":"stream error","grpc_code":13,"http_code":500}}\n']);

    const stream = apiStream<Delta>(chatServiceOperations.sendMessage, { path: { id: 's1' } });
    await expect(collect(stream)).rejects.toThrow(ApiStreamError);
  });

  it('途中で読むのをやめたらレスポンスをキャンセルする', async () => {
    const { cancel } = mockStream(['{"result":{"delta":"a"}}\n{"result":{"delta":"b"}}\n']);

    for await (const _ of apiStream<Delta>(chatServiceOperations.sendMessage, { path: { id: 's1' } })) {
      break;
    }

    expect(cancel).toHaveBeenCalled();
  });

  it('signal を fetch に渡す', async () => {
    const { fetchSpy } = mockStream(['{"result":{"done":true}}\n']);
    const controller = new AbortController();

    await collect(
      apiStream<Delta>(chatServiceOperations.sendMessage, { path: { id: 's1' }, signal: controller.signal }),
    );

    expect(fetchSpy).toHaveBeenCalledWith(
      `${API_BASE_URL}/chat/sessions/s1/messages`,
      expect.objectContaining({ method: 'POST', signal: controller.signal }),
    );
  });
});
