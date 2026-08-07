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

/**
 * Keycloak のセッションを初期化し、トークン保存のイベント配線を行う。
 *
 * keycloak はモジュールスコープのシングルトンで、そのハンドラを差し替えるのは
 * React の管理外にある外部システムへの副作用なので、コンポーネント本体ではなく
 * ここで行う（React Compiler の immutability ルール）。
 */
export const initializeKeycloak = async (handlers: KeycloakSessionHandlers): Promise<boolean> => {
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
    handlers.onLoggedOut();
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
      handlers.onAuthenticated(keycloak.tokenParsed);
    }
  }

  return authenticated;
};

export default keycloak;
