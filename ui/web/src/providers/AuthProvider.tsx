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
    // id_token_hint を明示的に渡すことで Keycloak がセッションを確実に破棄する。
    // keycloak-js が idToken を持っていない場合は手動でログアウト URL を組み立てる。
    const idToken = keycloak.idToken;
    const keycloakBaseUrl = import.meta.env.VITE_KEYCLOAK_URL || 'http://localhost:8080';
    const realm = import.meta.env.VITE_KEYCLOAK_REALM || 'hub';
    const redirectUri = window.location.origin;

    clearTokens();

    if (idToken) {
      const logoutOptions: KeycloakLogoutOptions = { redirectUri };
      await keycloak.logout(logoutOptions);
    } else {
      // idToken がない場合は直接ログアウトエンドポイントへ遷移する
      const logoutUrl =
        `${keycloakBaseUrl}/realms/${realm}/protocol/openid-connect/logout` +
        `?post_logout_redirect_uri=${encodeURIComponent(redirectUri)}` +
        `&client_id=hub-web`;
      window.location.href = logoutUrl;
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
