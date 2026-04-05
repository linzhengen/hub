import { AuthProvider } from '@refinedev/core';
import keycloak from '@/lib/keycloak';
import { saveToken, clearTokens, isTokenValid, debugTokenInfo, getParsedToken } from '@/lib/auth-token';

// Keycloakの初期化状態を追跡
let keycloakInitialized = false;
// 初期化中であることを追跡（複数回初期化を防ぐ）
let keycloakInitializing = false;
// ログアウト状態を追跡（ログアウト直後は認証チェックをスキップ）
let isLoggingOut = false;

// Keycloakの初期化状態をリセットする関数（ログアウト時に使用）
function resetKeycloakInitialization() {
  console.log('AuthProvider: Resetting Keycloak initialization flag');
  keycloakInitialized = false;
  keycloakInitializing = false;
}

// ログアウト状態を設定する関数
function setLoggingOut(state: boolean) {
  console.log('AuthProvider: Setting isLoggingOut to', state);
  isLoggingOut = state;
}

// Keycloakを初期化する関数
async function initializeKeycloak() {
  // ログアウト中の場合は初期化をスキップ
  if (isLoggingOut) {
    console.log('AuthProvider: Logout in progress, skipping Keycloak initialization');
    return false;
  }

  // 既に初期化中の場合は待機
  if (keycloakInitializing) {
    console.log('AuthProvider: Keycloak initialization already in progress, waiting...');
    // 短い待機後に再試行
    await new Promise(resolve => setTimeout(resolve, 100));
    // 再帰的に呼び出して状態を確認
    return initializeKeycloak();
  }

  if (!keycloakInitialized) {
    try {
      console.log('AuthProvider: Initializing Keycloak with login-required mode');
      keycloakInitializing = true;

      // トークン更新時のコールバックを設定（トークンをローカルストレージに保存）
      keycloak.onAuthSuccess = () => {
        console.log('AuthProvider: Authentication successful, saving token to localStorage');
        if (keycloak.token) {
          saveToken(
            keycloak.token,
            keycloak.refreshToken,
            keycloak.tokenParsed?.exp ? keycloak.tokenParsed.exp - Math.floor(Date.now() / 1000) : undefined
          );
          debugTokenInfo();
        }
      };

      keycloak.onAuthRefreshSuccess = () => {
        console.log('AuthProvider: Token refresh successful, updating localStorage');
        if (keycloak.token) {
          saveToken(
            keycloak.token,
            keycloak.refreshToken,
            keycloak.tokenParsed?.exp ? keycloak.tokenParsed.exp - Math.floor(Date.now() / 1000) : undefined
          );
          debugTokenInfo();
        }
      };

      keycloak.onAuthLogout = () => {
        console.log('AuthProvider: Logout detected, clearing tokens');
        clearTokens();
      };

      // トークン更新の設定
      keycloak.onTokenExpired = () => {
        console.log('AuthProvider: Token expired, attempting refresh');
        keycloak.updateToken(30).catch((error) => {
          console.error('AuthProvider: Token refresh failed:', error);
          // トークン更新失敗時は明示的なログインが必要
          keycloak.login();
        });
      };

      // 常にlogin-requiredモードを使用
      const authenticated = await keycloak.init({
        onLoad: 'login-required', // 未認証時は常にログイン画面へ
        checkLoginIframe: false,
        pkceMethod: 'S256',
      });

      keycloakInitialized = true;
      keycloakInitializing = false;

      // 初期化成功時にトークンを保存
      if (authenticated && keycloak.token) {
        saveToken(
          keycloak.token,
          keycloak.refreshToken,
          keycloak.tokenParsed?.exp ? keycloak.tokenParsed.exp - Math.floor(Date.now() / 1000) : undefined
        );
        debugTokenInfo();
      }

      return authenticated;
    } catch (error) {
      console.error('Keycloak initialization failed:', error);
      // 初期化失敗時はフラグをリセットして次回の初期化を許可
      keycloakInitialized = false;
      keycloakInitializing = false;
      return false;
    }
  }

  // 既に初期化済みの場合は、ローカルストレージのトークンをチェック
  console.log('AuthProvider: Keycloak already initialized, checking localStorage token');

  // ローカルストレージに有効なトークンがある場合は認証済みとみなす
  if (isTokenValid()) {
    console.log('AuthProvider: Valid token found in localStorage');
    return true;
  }

  console.log('AuthProvider: No valid token in localStorage');
  return false;
}

export const authProvider: AuthProvider = {
  login: async () => {
    console.log('AuthProvider: login() called');
    try {
      // login-requiredモードではinitializeKeycloak()が既にログインを要求する
      // 明示的なログインが必要な場合はkeycloak.login()を直接呼び出す
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
      console.log('AuthProvider: Keycloak authenticated:', keycloak.authenticated);
      console.log('AuthProvider: Keycloak token exists:', !!keycloak.token);

      // ログアウト状態を設定
      setLoggingOut(true);

      // まずローカルストレージのトークンをクリア
      clearTokens();
      console.log('AuthProvider: Local storage tokens cleared');

      // Keycloakの初期化状態をリセット
      resetKeycloakInitialization();

      // Keycloakのログアウトを実行
      // redirectUriを指定しないとKeycloakのログイン画面にリダイレクトされる
      // ログアウト後はKeycloakのログイン画面に直接リダイレクトする
      // シンプルに現在のオリジンにリダイレクト（SPAのルートに戻る）
      console.log('AuthProvider: Calling keycloak.logout() with redirectUri:', window.location.origin);
      const logoutOptions: any = {
        redirectUri: window.location.origin,
      };

      // id_token_hintを追加（Keycloakが要求する場合がある）
      if (keycloak.idToken) {
        logoutOptions.idTokenHint = keycloak.idToken;
        console.log('AuthProvider: Added idTokenHint to logout options');
      }

      console.log('AuthProvider: Logout options:', logoutOptions);

      await keycloak.logout(logoutOptions);

      console.log('AuthProvider: Logout completed successfully');
      // ログアウト状態を解除（リダイレクトが発生するのでここまで来ることは稀）
      setLoggingOut(false);

      // ログアウト成功、リダイレクトが発生するはず
      return { success: true };
    } catch (error) {
      console.error('AuthProvider: Logout failed:', error);
      // エラーが発生してもローカルストレージはクリアする
      clearTokens();

      // ログアウト状態を解除（エラー時も解除する）
      setLoggingOut(false);

      // ログアウトAPIが失敗した場合、手動でログイン画面にリダイレクト
      console.log('AuthProvider: Manually redirecting to login page');
      try {
        // KeycloakのログインURLを直接構築
        const keycloakBaseUrl = import.meta.env.VITE_KEYCLOAK_URL || 'http://localhost:8080';
        const realm = import.meta.env.VITE_KEYCLOAK_REALM || 'hub';
        const clientId = import.meta.env.VITE_KEYCLOAK_CLIENT_ID || 'hub-web';
        const redirectUri = encodeURIComponent(window.location.origin);
        const loginUrl = `${keycloakBaseUrl}/realms/${realm}/protocol/openid-connect/auth?client_id=${clientId}&redirect_uri=${redirectUri}&response_type=code&scope=openid`;

        console.log('AuthProvider: Redirecting to login URL:', loginUrl);
        window.location.href = loginUrl;
      } catch (redirectError) {
        console.error('AuthProvider: Failed to redirect to login:', redirectError);
        // 最終手段として現在のページをリロード
        window.location.reload();
      }

      return { success: false };
    }
  },

  check: async () => {
    try {
      console.log('AuthProvider: check() called - SPA page navigation check');
      console.log('AuthProvider: isLoggingOut =', isLoggingOut);
      console.log('AuthProvider: keycloakInitialized =', keycloakInitialized);
      console.log('AuthProvider: keycloakInitializing =', keycloakInitializing);

      // ログアウト中の場合は認証チェックをスキップ
      if (isLoggingOut) {
        console.log('AuthProvider: Logout in progress, skipping authentication check');
        return { authenticated: false };
      }

      // デバッグ情報を表示
      debugTokenInfo();

      // ローカルストレージに有効なトークンがある場合は認証済みとみなす
      if (isTokenValid()) {
        console.log('AuthProvider: Valid token found in localStorage, authentication confirmed');
        return { authenticated: true };
      }

      console.log('AuthProvider: No valid token in localStorage, initializing Keycloak with login-required');

      // ローカルストレージに有効なトークンがない場合はKeycloakを初期化
      // login-requiredモードなので、未認証の場合は自動的にログイン画面へ
      const authenticated = await initializeKeycloak();
      console.log('AuthProvider: Keycloak initialization result =', authenticated);

      return { authenticated };
    } catch (error) {
      console.log('AuthProvider: check() error:', error);
      return { authenticated: false };
    }
  },

  onError: async (error) => {
    console.error('AuthProvider: onError() called with error:', error);

    // 認証エラーの場合はトークンをクリア（次回のcheck()でKeycloakログイン画面へ）
    if (error?.status === 401 || error?.status === 403) {
      console.log('AuthProvider: Authentication error detected in onError - clearing tokens');
      clearTokens();
    }

    return { error };
  },

  getIdentity: async () => {
    console.log('AuthProvider: getIdentity() called');
    console.log('AuthProvider: isLoggingOut =', isLoggingOut);

    // ログアウト中の場合はnullを返す
    if (isLoggingOut) {
      console.log('AuthProvider: Logout in progress, returning null identity');
      return null;
    }

    // まずローカルストレージのトークン有効性をチェック
    if (!isTokenValid()) {
      console.log('AuthProvider: No valid token in localStorage, checking Keycloak');
      await initializeKeycloak();
    }

    console.log('AuthProvider: keycloak.tokenParsed =', keycloak.tokenParsed);
    console.log('AuthProvider: keycloak.authenticated =', keycloak.authenticated);
    console.log('AuthProvider: keycloak.token =', keycloak.token ? `exists (${keycloak.token.length} chars)` : 'null');

    // まずKeycloakのトークンパース情報を試す
    if (keycloak.tokenParsed) {
      const identity = {
        id: keycloak.tokenParsed.sub,
        name: keycloak.tokenParsed.name || keycloak.tokenParsed.preferred_username,
        avatar: keycloak.tokenParsed.picture,
        email: keycloak.tokenParsed.email,
      };
      console.log('AuthProvider: Returning identity from keycloak.tokenParsed:', identity);
      return identity;
    }

    // Keycloakのトークンパース情報がない場合は、ローカルストレージのトークンから取得を試みる
    const parsedToken = getParsedToken();
    console.log('AuthProvider: Parsed token from localStorage:', parsedToken);

    if (parsedToken) {
      const identity = {
        id: parsedToken.sub,
        name: parsedToken.name || parsedToken.preferred_username,
        avatar: parsedToken.picture,
        email: parsedToken.email,
      };
      console.log('AuthProvider: Returning identity from localStorage token:', identity);
      return identity;
    }

    console.log('AuthProvider: No token parsed from either source, returning null');
    return null;
  },

  getPermissions: async () => {
    console.log('AuthProvider: getPermissions() called');
    console.log('AuthProvider: isLoggingOut =', isLoggingOut);

    // ログアウト中の場合は空の配列を返す
    if (isLoggingOut) {
      console.log('AuthProvider: Logout in progress, returning empty permissions');
      return [];
    }

    // まずローカルストレージのトークン有効性をチェック
    if (!isTokenValid()) {
      console.log('AuthProvider: No valid token in localStorage, checking Keycloak');
      await initializeKeycloak();
    }

    console.log('AuthProvider: keycloak.tokenParsed for permissions =', keycloak.tokenParsed);

    // まずKeycloakのトークンパース情報から権限を取得
    let roles: string[] = [];
    let resourceRoles: any = {};

    if (keycloak.tokenParsed) {
      roles = keycloak.tokenParsed?.realm_access?.roles || [];
      resourceRoles = keycloak.tokenParsed?.resource_access || {};
    } else {
      // Keycloakのトークンパース情報がない場合は、ローカルストレージのトークンから取得
      const parsedToken = getParsedToken();
      console.log('AuthProvider: Parsed token for permissions:', parsedToken);

      if (parsedToken) {
        roles = parsedToken?.realm_access?.roles || [];
        resourceRoles = parsedToken?.resource_access || {};
      }
    }

    // すべての権限を結合
    const allPermissions = [...roles];
    Object.values(resourceRoles).forEach((resource: any) => {
      if (resource.roles) {
        allPermissions.push(...resource.roles);
      }
    });

    console.log('AuthProvider: Returning permissions:', allPermissions);
    return allPermissions;
  },

};
