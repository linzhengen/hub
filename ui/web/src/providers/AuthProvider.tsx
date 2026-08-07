import React, { useState, useEffect, ReactNode } from 'react';
import type { KeycloakTokenParsed, KeycloakLogoutOptions } from 'keycloak-js';
import keycloak, { initializeKeycloak } from '@/lib/keycloak';
import { clearTokens } from '@/lib/auth-token';
import { AuthContext, type AppUser } from '@/providers/auth';

interface AuthProviderProps {
  children: ReactNode;
}

// 表示名の構築: given_name + family_name があれば結合、なければ既存の name または preferred_username
const toAppUser = (tokenParsed: KeycloakTokenParsed): AppUser => {
  let displayName = tokenParsed.name || tokenParsed.preferred_username;
  if (tokenParsed.given_name && tokenParsed.family_name) {
    displayName = `${tokenParsed.given_name} ${tokenParsed.family_name}`;
  } else if (tokenParsed.given_name) {
    displayName = tokenParsed.given_name;
  } else if (tokenParsed.family_name) {
    displayName = tokenParsed.family_name;
  }

  return {
    id: tokenParsed.sub,
    name: displayName,
    email: tokenParsed.email,
    emailVerified: tokenParsed.email_verified || false,
    roles: tokenParsed?.realm_access?.roles || [],
  };
};

export const AuthProvider: React.FC<AuthProviderProps> = ({ children }) => {
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [isLoading, setIsLoading] = useState(true);
  const [user, setUser] = useState<AppUser | null>(null);

  useEffect(() => {
    let cancelled = false;

    initializeKeycloak({
      onAuthenticated: (tokenParsed) => {
        if (cancelled) return;
        setUser(toAppUser(tokenParsed));
      },
      onLoggedOut: () => {
        if (cancelled) return;
        setIsAuthenticated(false);
        setUser(null);
      },
    })
      .then((authenticated) => {
        if (cancelled) return;
        setIsAuthenticated(authenticated);
      })
      .catch((error) => {
        console.error('Keycloak initialization failed:', error);
        if (cancelled) return;
        setIsAuthenticated(false);
      })
      .finally(() => {
        if (cancelled) return;
        setIsLoading(false);
      });

    return () => {
      cancelled = true;
    };
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
      console.log('Logging out...');

      // keycloak-js は id_token_hint を this.idToken から自動的に付与するため、
      // options 経由で渡す必要はない。
      const logoutOptions: KeycloakLogoutOptions = {
        redirectUri: window.location.origin,
      };

      // トークンを先にクリアする（もしリダイレクトに失敗しても再認証が必要な状態にするため）
      clearTokens();

      // keycloak.logout() はリダイレクトを伴うため、ここで処理が中断される可能性がある
      await keycloak.logout(logoutOptions);

      // keycloak.logout() がリダイレクトしない場合に備えて状態を更新
      setIsAuthenticated(false);
      setUser(null);
    } catch (error) {
      console.error('Logout failed:', error);
      // エラーが発生してもローカルの状態はクリアする
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
