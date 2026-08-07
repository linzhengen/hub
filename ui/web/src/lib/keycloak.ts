import Keycloak from 'keycloak-js';
import type { KeycloakTokenParsed } from 'keycloak-js';
import { clearTokens, saveToken } from '@/lib/auth-token';

const keycloakConfig = {
  url: import.meta.env.VITE_KEYCLOAK_URL || 'http://localhost:8080',
  realm: import.meta.env.VITE_KEYCLOAK_REALM || 'hub',
  clientId: import.meta.env.VITE_KEYCLOAK_CLIENT_ID || 'hub-web',
};

const keycloak = new Keycloak(keycloakConfig);

/**
 * 現在のトークンをメモリに保存する。
 *
 * 有効期限は「exp - 現在時刻」の秒数として渡す。`Date.now()` を読むため、
 * この処理はレンダー中に呼んではならない（React Compiler の purity ルール）。
 */
const persistCurrentToken = () => {
  if (!keycloak.token) return;

  saveToken(
    keycloak.token,
    keycloak.refreshToken,
    keycloak.tokenParsed?.exp ? keycloak.tokenParsed.exp - Math.floor(Date.now() / 1000) : undefined,
  );
};

export interface KeycloakSessionHandlers {
  onAuthenticated: (tokenParsed: KeycloakTokenParsed) => void;
  onLoggedOut: () => void;
}

// keycloak.init() は 1 インスタンスにつき一度しか呼べず、二度目は必ず throw する。
// StrictMode は開発ビルドで effect を二重に実行するため、素直に呼ぶと二度目が
// 失敗して未認証と判定され、ログイン画面へのリダイレクトループになる。
// 実行中/完了済みの初期化を使い回して、init 自体は一度だけにする。
let initialization: Promise<boolean> | undefined;

// ハンドラは常に最新の呼び出し元のものを使う。初期化を共有する都合上、
// 先行する呼び出しのクロージャを掴んだままにすると、破棄済みの effect に
// 通知してしまい状態が反映されない。
let currentHandlers: KeycloakSessionHandlers | undefined;

/**
 * Keycloak のセッションを初期化し、トークン保存のイベント配線を行う。
 *
 * keycloak はモジュールスコープのシングルトンで、そのハンドラを差し替えるのは
 * React の管理外にある外部システムへの副作用なので、コンポーネント本体ではなく
 * ここで行う（React Compiler の immutability ルール）。
 */
export const initializeKeycloak = (handlers: KeycloakSessionHandlers): Promise<boolean> => {
  currentHandlers = handlers;
  initialization ??= startInitialization();
  return initialization;
};

/** テスト用: モジュール状態を初期化前に戻す */
export const resetKeycloakInitializationForTests = () => {
  initialization = undefined;
  currentHandlers = undefined;
};

const startInitialization = async (): Promise<boolean> => {
  console.log('Initializing Keycloak...');

  keycloak.onAuthSuccess = () => {
    console.log('Authentication successful');
    persistCurrentToken();
  };

  keycloak.onAuthRefreshSuccess = () => {
    console.log('Token refresh successful');
    persistCurrentToken();
  };

  keycloak.onAuthLogout = () => {
    console.log('Logout detected');
    clearTokens();
    currentHandlers?.onLoggedOut();
  };

  keycloak.onTokenExpired = () => {
    console.log('Token expired, attempting refresh');
    keycloak.updateToken(30).catch((error) => {
      console.error('Token refresh failed:', error);
      keycloak.login();
    });
  };

  // トークンはメモリにのみ保持されページリロードで失われるため、毎回初期化する。
  // 有効な SSO セッション（Cookie）があれば login-required はリダイレクトなしで
  // サイレントに再認証される。
  const authenticated = await keycloak.init({
    onLoad: 'login-required',
    checkLoginIframe: false,
    pkceMethod: 'S256',
  });

  if (authenticated && keycloak.token) {
    persistCurrentToken();

    if (keycloak.tokenParsed) {
      currentHandlers?.onAuthenticated(keycloak.tokenParsed);
    }
  }

  return authenticated;
};

export default keycloak;
