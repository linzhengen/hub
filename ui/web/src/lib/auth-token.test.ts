import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import {
  clearTokens,
  getParsedToken,
  getRefreshToken,
  getSecondsUntilExpiry,
  getToken,
  getTokenExpiry,
  isTokenValid,
  parseToken,
  saveToken,
} from '@/lib/auth-token';

/** saveToken が期限計算に使うマージン（auth-token.ts と同じ値） */
const EXPIRY_MARGIN_MS = 30_000;

const encodeSegment = (value: object): string =>
  btoa(JSON.stringify(value)).replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '');

const buildJwt = (payload: object): string =>
  `${encodeSegment({ alg: 'RS256', typ: 'JWT' })}.${encodeSegment(payload)}.signature`;

beforeEach(() => {
  clearTokens();
});

afterEach(() => {
  vi.useRealTimers();
  clearTokens();
});

describe('saveToken', () => {
  it('アクセストークンとリフレッシュトークンをメモリに保持する', () => {
    saveToken('access-token', 'refresh-token');

    expect(getToken()).toBe('access-token');
    expect(getRefreshToken()).toBe('refresh-token');
  });

  it('空文字のトークンは保存しない', () => {
    const warn = vi.spyOn(console, 'warn').mockImplementation(() => {});

    saveToken('');

    expect(getToken()).toBeNull();
    expect(warn).toHaveBeenCalled();
  });

  it('リフレッシュトークン未指定時は既存の値を保持する', () => {
    saveToken('access-token-1', 'refresh-token');
    saveToken('access-token-2');

    expect(getToken()).toBe('access-token-2');
    expect(getRefreshToken()).toBe('refresh-token');
  });

  it('expiresIn から 30 秒のマージンを引いた期限を計算する', () => {
    vi.useFakeTimers();
    const now = new Date('2026-01-01T00:00:00.000Z');
    vi.setSystemTime(now);

    saveToken('access-token', undefined, 300);

    expect(getTokenExpiry()).toBe(now.getTime() + 300_000 - EXPIRY_MARGIN_MS);
  });

  it('expiresIn 未指定時は期限を設定しない', () => {
    saveToken('access-token');

    expect(getTokenExpiry()).toBeNull();
  });
});

describe('isTokenValid', () => {
  it('トークンが無い場合は false', () => {
    expect(isTokenValid()).toBe(false);
  });

  it('期限情報が無い場合は有効とみなす', () => {
    saveToken('access-token');

    expect(isTokenValid()).toBe(true);
  });

  it('期限内であれば true', () => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date('2026-01-01T00:00:00.000Z'));
    saveToken('access-token', undefined, 300);

    vi.advanceTimersByTime(100_000);

    expect(isTokenValid()).toBe(true);
  });

  it('マージンを含む期限を過ぎたら false', () => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date('2026-01-01T00:00:00.000Z'));
    saveToken('access-token', undefined, 300);

    // 実際の失効は 300 秒後だが、30 秒のマージンにより 270 秒で無効になる
    vi.advanceTimersByTime(270_000);

    expect(isTokenValid()).toBe(false);
  });
});

describe('getSecondsUntilExpiry', () => {
  it('期限が未設定なら null', () => {
    saveToken('access-token');

    expect(getSecondsUntilExpiry()).toBeNull();
  });

  it('残り秒数を切り捨てで返す', () => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date('2026-01-01T00:00:00.000Z'));
    saveToken('access-token', undefined, 300);

    vi.advanceTimersByTime(10_500);

    // 300 - 30 - 10.5 = 259.5 秒 -> 切り捨てて 259
    expect(getSecondsUntilExpiry()).toBe(259);
  });

  it('期限切れ後は負値ではなく 0 を返す', () => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date('2026-01-01T00:00:00.000Z'));
    saveToken('access-token', undefined, 300);

    vi.advanceTimersByTime(600_000);

    expect(getSecondsUntilExpiry()).toBe(0);
  });
});

describe('clearTokens', () => {
  it('トークン・リフレッシュトークン・期限をすべて破棄する', () => {
    saveToken('access-token', 'refresh-token', 300);

    clearTokens();

    expect(getToken()).toBeNull();
    expect(getRefreshToken()).toBeNull();
    expect(getTokenExpiry()).toBeNull();
    expect(isTokenValid()).toBe(false);
  });
});

describe('parseToken', () => {
  it('base64url エンコードされたペイロードをデコードする', () => {
    const payload = { sub: 'user-1', preferred_username: 'linzhengen', realm_access: { roles: ['admin'] } };

    expect(parseToken(buildJwt(payload))).toEqual(payload);
  });

  it('3 セグメントでない場合は null', () => {
    vi.spyOn(console, 'warn').mockImplementation(() => {});

    expect(parseToken('not-a-jwt')).toBeNull();
  });

  it('ペイロードが JSON として不正な場合は null', () => {
    vi.spyOn(console, 'error').mockImplementation(() => {});

    expect(parseToken('header.bm90LWpzb24.signature')).toBeNull();
  });

  it('空文字は null', () => {
    expect(parseToken('')).toBeNull();
  });
});

describe('getParsedToken', () => {
  it('トークン未保存なら null', () => {
    expect(getParsedToken()).toBeNull();
  });

  it('保存済みトークンのペイロードを返す', () => {
    saveToken(buildJwt({ sub: 'user-1', email: 'user@example.com' }));

    expect(getParsedToken()).toEqual({ sub: 'user-1', email: 'user@example.com' });
  });
});
