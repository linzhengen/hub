import { AuthProvider } from '@refinedev/core';
import keycloak from '@/lib/keycloak';

// Keycloakの初期化状態を追跡
let keycloakInitialized = false;

// Keycloakを初期化する関数
async function initializeKeycloak() {
  if (!keycloakInitialized) {
    try {
      const authenticated = await keycloak.init({
        onLoad: 'login-required', // 未認証時は自動的にログイン画面へ
        checkLoginIframe: false,
        pkceMethod: 'S256',
      });

      keycloakInitialized = true;

      // トークン更新の設定
      keycloak.onTokenExpired = () => {
        keycloak.updateToken(30).catch(() => {
          keycloak.login();
        });
      };

      return authenticated;
    } catch (error) {
      console.error('Keycloak initialization failed:', error);
      // AuthProvider.tsxと同様に、エラー時も初期化済みとしてマーク
      keycloakInitialized = true;
      return false;
    }
  }
  return keycloak.authenticated || false;
}

export const authProvider: AuthProvider = {
  login: async () => {
    console.log('AuthProvider: login() called');
    // AuthProvider.tsxと同じシンプルな実装
    try {
      await keycloak.login();
      console.log('AuthProvider: login() completed');
      return { success: true };
    } catch (error) {
      console.error('AuthProvider: login() failed:', error);
      return {
        success: false,
        error: {
          name: 'Login Error',
          message: 'Login failed',
        },
      };
    }
  },

  logout: async () => {
    try {
      console.log('AuthProvider: Starting logout...');
      console.log('AuthProvider: Current location:', window.location.href);
      console.log('AuthProvider: Keycloak instance:', keycloak);

      // ログアウト後にログイン画面にリダイレクトするためにredirectUriを指定
      await keycloak.logout({
        redirectUri: window.location.origin,
      });

      console.log('AuthProvider: Logout completed successfully');
      return { success: true };
    } catch (error) {
      console.error('AuthProvider: Logout failed:', error);
      return { success: false };
    }
  },

  check: async () => {
    try {
      console.log('AuthProvider: check() called');
      const authenticated = await initializeKeycloak();
      console.log('AuthProvider: authenticated =', authenticated);
      console.log('AuthProvider: keycloak.authenticated =', keycloak.authenticated);
      return { authenticated };
    } catch (error) {
      console.log('AuthProvider: check() error:', error);
      return { authenticated: false };
    }
  },

  onError: async (error) => {
    console.error('AuthProvider: onError() called with error:', error);

    // 認証エラーの場合はログアウト処理を実行
    if (error?.status === 401 || error?.status === 403) {
      console.log('AuthProvider: Authentication error detected, triggering logout');
      return {
        logout: true,
        redirectTo: '/',
        error: error,
      };
    }

    return { error };
  },

  getIdentity: async () => {
    await initializeKeycloak();

    if (keycloak.tokenParsed) {
      return {
        id: keycloak.tokenParsed.sub,
        name: keycloak.tokenParsed.name || keycloak.tokenParsed.preferred_username,
        avatar: keycloak.tokenParsed.picture,
        email: keycloak.tokenParsed.email,
      };
    }
    return null;
  },

  getPermissions: async () => {
    await initializeKeycloak();

    // Keycloakのロールや権限を取得
    const roles = keycloak.tokenParsed?.realm_access?.roles || [];
    const resourceRoles = keycloak.tokenParsed?.resource_access || {};

    // すべての権限を結合
    const allPermissions = [...roles];
    Object.values(resourceRoles).forEach((resource: any) => {
      if (resource.roles) {
        allPermissions.push(...resource.roles);
      }
    });

    return allPermissions;
  },

};
