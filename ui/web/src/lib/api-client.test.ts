import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { API_BASE_URL, ApiError, fetchApi } from '@/lib/api-client';
import { clearTokens, getToken, saveToken } from '@/lib/auth-token';

interface JsonResponseInit {
  status?: number;
  body?: unknown;
  bodyIsInvalidJson?: boolean;
}

const mockResponse = ({ status = 200, body = {}, bodyIsInvalidJson = false }: JsonResponseInit = {}) =>
  ({
    ok: status >= 200 && status < 300,
    status,
    statusText: `status ${status}`,
    json: bodyIsInvalidJson
      ? vi.fn().mockRejectedValue(new SyntaxError('Unexpected token'))
      : vi.fn().mockResolvedValue(body),
  }) as unknown as Response;

const mockFetch = (response: Response) => vi.spyOn(globalThis, 'fetch').mockResolvedValue(response);

const lastRequestInit = (fetchSpy: ReturnType<typeof mockFetch>): RequestInit =>
  fetchSpy.mock.calls[0][1] as RequestInit;

beforeEach(() => {
  clearTokens();
  vi.spyOn(console, 'error').mockImplementation(() => {});
  vi.spyOn(console, 'warn').mockImplementation(() => {});
  vi.spyOn(console, 'log').mockImplementation(() => {});
});

afterEach(() => {
  clearTokens();
});

describe('fetchApi', () => {
  it('API_BASE_URL を前置した URL にリクエストする', async () => {
    const fetchSpy = mockFetch(mockResponse({ body: { users: [] } }));

    await expect(fetchApi('/users')).resolves.toEqual({ users: [] });
    expect(fetchSpy).toHaveBeenCalledWith(`${API_BASE_URL}/users`, expect.anything());
  });

  it('トークンがあれば Authorization ヘッダーを付与する', async () => {
    saveToken('access-token');
    const fetchSpy = mockFetch(mockResponse());

    await fetchApi('/users');

    expect(lastRequestInit(fetchSpy).headers).toMatchObject({
      Authorization: 'Bearer access-token',
      'Content-Type': 'application/json',
    });
  });

  it('トークンが無ければ Authorization ヘッダーを付けない', async () => {
    const fetchSpy = mockFetch(mockResponse());

    await fetchApi('/users');

    expect(lastRequestInit(fetchSpy).headers).not.toHaveProperty('Authorization');
  });

  it('呼び出し側のヘッダーで既定値を上書きできる', async () => {
    const fetchSpy = mockFetch(mockResponse());

    await fetchApi('/users', { headers: { 'Content-Type': 'text/plain' } });

    expect(lastRequestInit(fetchSpy).headers).toMatchObject({ 'Content-Type': 'text/plain' });
  });

  it('204 No Content では空オブジェクトを返す', async () => {
    const response = mockResponse({ status: 204 });
    mockFetch(response);

    await expect(fetchApi('/users/1', { method: 'DELETE' })).resolves.toEqual({});
    expect(response.json).not.toHaveBeenCalled();
  });
});

describe('fetchApi の 401 ハンドリング', () => {
  it('トークンを破棄し api-unauthorized イベントを発火して ApiError を投げる', async () => {
    saveToken('access-token', 'refresh-token', 300);
    mockFetch(mockResponse({ status: 401, body: { message: 'token expired' } }));

    const onUnauthorized = vi.fn();
    window.addEventListener('api-unauthorized', onUnauthorized);

    try {
      const error = await fetchApi('/users').catch((e: unknown) => e);

      expect(error).toBeInstanceOf(ApiError);
      expect(error).toMatchObject({ name: 'ApiError', status: 401, message: 'token expired' });
      expect((error as ApiError).data).toEqual({ message: 'token expired' });
      expect(getToken()).toBeNull();
      expect(onUnauthorized).toHaveBeenCalledTimes(1);
    } finally {
      window.removeEventListener('api-unauthorized', onUnauthorized);
    }
  });

  it('401 のボディが JSON でなくても statusText をメッセージにして ApiError を投げる', async () => {
    saveToken('access-token');
    mockFetch(mockResponse({ status: 401, bodyIsInvalidJson: true }));

    const error = await fetchApi('/users').catch((e: unknown) => e);

    expect(error).toBeInstanceOf(ApiError);
    expect(error).toMatchObject({ status: 401, message: 'status 401' });
    expect((error as ApiError).data).toBeNull();
    expect(getToken()).toBeNull();
  });
});

describe('fetchApi の 401 以外のエラー', () => {
  it('ApiError ではない Error を投げ、トークンは破棄しない', async () => {
    saveToken('access-token');
    mockFetch(mockResponse({ status: 500, body: { message: 'internal error' } }));

    const onUnauthorized = vi.fn();
    window.addEventListener('api-unauthorized', onUnauthorized);

    try {
      const error = await fetchApi('/users').catch((e: unknown) => e);

      expect(error).toBeInstanceOf(Error);
      expect(error).not.toBeInstanceOf(ApiError);
      expect((error as Error).message).toBe('internal error');
      expect(getToken()).toBe('access-token');
      expect(onUnauthorized).not.toHaveBeenCalled();
    } finally {
      window.removeEventListener('api-unauthorized', onUnauthorized);
    }
  });

  it('エラーボディに message が無ければ既定メッセージを使う', async () => {
    mockFetch(mockResponse({ status: 400, body: {} }));

    await expect(fetchApi('/users')).rejects.toThrow('An error occurred');
  });
});
