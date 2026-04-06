import React, { createContext, useContext, useState, useEffect, ReactNode } from 'react';
import keycloak from '@/lib/keycloak';
import { saveToken, clearTokens, isTokenValid } from '@/lib/auth-token';

interface AuthContextType {
  isAuthenticated: boolean;
  isLoading: boolean;
  user: any | null;
  login: () => Promise<void>;
  logout: () => Promise<void>;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

interface AuthProviderProps {
  children: ReactNode;
}

export const AuthProvider: React.FC<AuthProviderProps> = ({ children }) => {
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [isLoading, setIsLoading] = useState(true);
  const [user, setUser] = useState<any>(null);

  // Keycloak初期化関数
  const initializeKeycloak = async () => {
    try {
      console.log('Initializing Keycloak...');

      // トークン更新時のコールバックを設定
      keycloak.onAuthSuccess = () => {
        console.log('Authentication successful');
        if (keycloak.token) {
          saveToken(
            keycloak.token,
            keycloak.refreshToken,
            keycloak.tokenParsed?.exp ? keycloak.tokenParsed.exp - Math.floor(Date.now() / 1000) : undefined
          );
        }
      };

      keycloak.onAuthRefreshSuccess = () => {
        console.log('Token refresh successful');
        if (keycloak.token) {
          saveToken(
            keycloak.token,
            keycloak.refreshToken,
            keycloak.tokenParsed?.exp ? keycloak.tokenParsed.exp - Math.floor(Date.now() / 1000) : undefined
          );
        }
      };

      keycloak.onAuthLogout = () => {
        console.log('Logout detected');
        clearTokens();
        setIsAuthenticated(false);
        setUser(null);
      };

      // トークン更新の設定
      keycloak.onTokenExpired = () => {
        console.log('Token expired, attempting refresh');
        keycloak.updateToken(30).catch((error) => {
          console.error('Token refresh failed:', error);
          keycloak.login();
        });
      };

      // login-requiredモードで初期化
      const authenticated = await keycloak.init({
        onLoad: 'login-required',
        checkLoginIframe: false,
        pkceMethod: 'S256',
      });

      if (authenticated && keycloak.token) {
        saveToken(
          keycloak.token,
          keycloak.refreshToken,
          keycloak.tokenParsed?.exp ? keycloak.tokenParsed.exp - Math.floor(Date.now() / 1000) : undefined
        );

        // ユーザー情報を設定
        if (keycloak.tokenParsed) {
          // 表示名の構築: given_name + family_name があれば結合、なければ既存の name または preferred_username
          let displayName = keycloak.tokenParsed.name || keycloak.tokenParsed.preferred_username;
          if (keycloak.tokenParsed.given_name && keycloak.tokenParsed.family_name) {
            displayName = `${keycloak.tokenParsed.given_name} ${keycloak.tokenParsed.family_name}`;
          } else if (keycloak.tokenParsed.given_name) {
            displayName = keycloak.tokenParsed.given_name;
          } else if (keycloak.tokenParsed.family_name) {
            displayName = keycloak.tokenParsed.family_name;
          }

          setUser({
            id: keycloak.tokenParsed.sub,
            name: displayName,
            email: keycloak.tokenParsed.email,
            emailVerified: keycloak.tokenParsed.email_verified || false,
            roles: keycloak.tokenParsed?.realm_access?.roles || [],
          });
        }

        setIsAuthenticated(true);
      } else {
        setIsAuthenticated(false);
      }

      return authenticated;
    } catch (error) {
      console.error('Keycloak initialization failed:', error);
      setIsAuthenticated(false);
      return false;
    } finally {
      setIsLoading(false);
    }
  };

  // 認証状態のチェック
  const checkAuth = async () => {
    // ローカルストレージに有効なトークンがある場合は認証済みとみなす
    if (isTokenValid()) {
      console.log('Valid token found in localStorage');
      setIsAuthenticated(true);
      setIsLoading(false);
      return;
    }

    // トークンがない場合はKeycloakを初期化
    console.log('No valid token, initializing Keycloak');
    await initializeKeycloak();
  };

  useEffect(() => {
    checkAuth();
  }, []);

  const login = async () => {
    try {
      await keycloak.login();
    } catch (error) {
      console.error('Login failed:', error);
    }
  };

  const logout = async () => {
    try {
      clearTokens();
      setIsAuthenticated(false);
      setUser(null);

      const logoutOptions: any = {
        redirectUri: window.location.origin,
      };

      if (keycloak.idToken) {
        logoutOptions.idTokenHint = keycloak.idToken;
      }

      await keycloak.logout(logoutOptions);
    } catch (error) {
      console.error('Logout failed:', error);
      // エラーが発生してもローカルストレージはクリアする
      clearTokens();
      setIsAuthenticated(false);
      setUser(null);

      // 手動でログイン画面にリダイレクト
      const keycloakBaseUrl = import.meta.env.VITE_KEYCLOAK_URL || 'http://localhost:8080';
      const realm = import.meta.env.VITE_KEYCLOAK_REALM || 'hub';
      const clientId = import.meta.env.VITE_KEYCLOAK_CLIENT_ID || 'hub-web';
      const redirectUri = encodeURIComponent(window.location.origin);
      const loginUrl = `${keycloakBaseUrl}/realms/${realm}/protocol/openid-connect/auth?client_id=${clientId}&redirect_uri=${redirectUri}&response_type=code&scope=openid`;

      window.location.href = loginUrl;
    }
  };

  const value = {
    isAuthenticated,
    isLoading,
    user,
    login,
    logout,
  };

  return (
    <AuthContext.Provider value={value}>
      {children}
    </AuthContext.Provider>
  );
};

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
};
