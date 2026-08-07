import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

/**
 * keycloak-js の実インスタンスの代わりに、ハンドラ代入と init 呼び出しを
 * 記録するだけのスタブを使う。lib/keycloak.ts はモジュール読み込み時に
 * `new Keycloak(...)` するため、モジュールモックで差し替える必要がある。
 */
const keycloakStub = vi.hoisted(() => ({
  token: undefined as string | undefined,
  refreshToken: undefined as string | undefined,
  tokenParsed: undefined as Record<string, unknown> | undefined,
  onAuthSuccess: undefined as (() => void) | undefined,
  onAuthRefreshSuccess: undefined as (() => void) | undefined,
  onAuthLogout: undefined as (() => void) | undefined,
  onTokenExpired: undefined as (() => void) | undefined,
  init: vi.fn(),
  login: vi.fn(),
  updateToken: vi.fn(),
}));

vi.mock('keycloak-js', () => ({
  // アロー関数は new できないため、通常の関数で構築する
  default: function Keycloak() {
    return keycloakStub;
  },
}));

const { initializeKeycloak, resetKeycloakInitializationForTests } = await import('@/lib/keycloak');
const { clearTokens, getToken, getSecondsUntilExpiry } = await import('@/lib/auth-token');

/** exp は秒精度なので、固定時刻を使って残り秒数を決定的にする */
const NOW = new Date('2026-01-01T00:00:00.000Z');
const expIn = (seconds: number) => Math.floor(NOW.getTime() / 1000) + seconds;

const handlers = () => ({ onAuthenticated: vi.fn(), onLoggedOut: vi.fn() });

beforeEach(() => {
  vi.useFakeTimers();
  vi.setSystemTime(NOW);
  vi.spyOn(console, 'log').mockImplementation(() => {});
  vi.spyOn(console, 'error').mockImplementation(() => {});
  clearTokens();
  resetKeycloakInitializationForTests();
  keycloakStub.token = 'access-token';
  keycloakStub.refreshToken = 'refresh-token';
  keycloakStub.tokenParsed = { sub: 'user-1', exp: expIn(300) };
  keycloakStub.init.mockResolvedValue(true);
});

afterEach(() => {
  vi.useRealTimers();
  clearTokens();
});

describe('initializeKeycloak', () => {
  it('認証済みならトークンを保存し、有効期限を exp から計算する', async () => {
    await initializeKeycloak(handlers());

    expect(getToken()).toBe('access-token');
    // saveToken は 30 秒のマージンを引くので 300 - 30 = 270
    expect(getSecondsUntilExpiry()).toBe(270);
  });

  it('認証済みなら tokenParsed を onAuthenticated に渡す', async () => {
    const h = handlers();

    await expect(initializeKeycloak(h)).resolves.toBe(true);
    expect(h.onAuthenticated).toHaveBeenCalledWith(keycloakStub.tokenParsed);
  });

  it('未認証なら onAuthenticated を呼ばずトークンも保存しない', async () => {
    keycloakStub.init.mockResolvedValue(false);
    const h = handlers();

    await expect(initializeKeycloak(h)).resolves.toBe(false);
    expect(h.onAuthenticated).not.toHaveBeenCalled();
    expect(getToken()).toBeNull();
  });

  it('login-required / PKCE S256 で初期化する', async () => {
    await initializeKeycloak(handlers());

    expect(keycloakStub.init).toHaveBeenCalledWith({
      onLoad: 'login-required',
      checkLoginIframe: false,
      pkceMethod: 'S256',
    });
  });

  it('トークンリフレッシュ成功時に新しいトークンを保存し直す', async () => {
    await initializeKeycloak(handlers());

    keycloakStub.token = 'refreshed-token';
    keycloakStub.tokenParsed = { sub: 'user-1', exp: expIn(600) };
    keycloakStub.onAuthRefreshSuccess?.();

    expect(getToken()).toBe('refreshed-token');
    expect(getSecondsUntilExpiry()).toBe(570);
  });

  it('ログアウト検知でトークンを破棄し onLoggedOut を呼ぶ', async () => {
    const h = handlers();
    await initializeKeycloak(h);

    keycloakStub.onAuthLogout?.();

    expect(getToken()).toBeNull();
    expect(h.onLoggedOut).toHaveBeenCalledTimes(1);
  });

  it('トークン期限切れでリフレッシュを試み、失敗したらログインへ送る', async () => {
    keycloakStub.updateToken.mockRejectedValue(new Error('refresh failed'));
    await initializeKeycloak(handlers());

    keycloakStub.onTokenExpired?.();
    await vi.waitFor(() => expect(keycloakStub.login).toHaveBeenCalledTimes(1));
    expect(keycloakStub.updateToken).toHaveBeenCalledWith(30);
  });

  it('init が失敗したら呼び出し側に伝播する', async () => {
    keycloakStub.init.mockRejectedValue(new Error('keycloak down'));

    await expect(initializeKeycloak(handlers())).rejects.toThrow('keycloak down');
  });
});

describe('二重初期化の抑止', () => {
  it('複数回呼ばれても keycloak.init は一度しか実行しない', async () => {
    // StrictMode は開発ビルドで effect を二重に実行する。keycloak.init() は
    // 二度目に必ず throw するため、素直に呼ぶと未認証と判定されてしまう。
    const first = handlers();
    const second = handlers();

    const results = await Promise.all([initializeKeycloak(first), initializeKeycloak(second)]);

    expect(keycloakStub.init).toHaveBeenCalledTimes(1);
    expect(results).toEqual([true, true]);
  });

  it('通知先は最後に渡されたハンドラになる', async () => {
    // 先行する effect は cleanup 済みで、その setState は捨てられる。
    // 古いクロージャに通知すると認証状態が画面に反映されない。
    const stale = handlers();
    const latest = handlers();

    const pending = initializeKeycloak(stale);
    await initializeKeycloak(latest);
    await pending;

    expect(latest.onAuthenticated).toHaveBeenCalledWith(keycloakStub.tokenParsed);
    expect(stale.onAuthenticated).not.toHaveBeenCalled();

    keycloakStub.onAuthLogout?.();
    expect(latest.onLoggedOut).toHaveBeenCalledTimes(1);
    expect(stale.onLoggedOut).not.toHaveBeenCalled();
  });
});
