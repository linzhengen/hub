import { afterEach, describe, expect, it, vi } from 'vitest';
import type { ApiOperation } from '@/api/operations';
import { userServiceOperations } from '@/api/operations';
import { apiRequest, buildPath, buildQueryString } from '@/api/request';
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
