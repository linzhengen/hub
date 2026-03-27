import React, { createContext, useContext, useEffect, useState } from 'react';
import keycloak from '@/lib/keycloak';

interface AuthContextType {
  isAuthenticated: boolean;
  isInitialized: boolean;
  token: string | undefined;
  login: () => void;
  logout: () => void;
}

const AuthContext = createContext<AuthContextType>({
  isAuthenticated: false,
  isInitialized: false,
  token: undefined,
  login: () => {},
  logout: () => {},
});

export const useAuth = () => useContext(AuthContext);

export const AuthProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [isInitialized, setIsInitialized] = useState(false);
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [token, setToken] = useState<string | undefined>(undefined);

  useEffect(() => {
    const initKeycloak = async () => {
      try {
        const authenticated = await keycloak.init({
          onLoad: 'login-required',
          checkLoginIframe: false,
        });

        setIsAuthenticated(authenticated);
        setToken(keycloak.token);
        setIsInitialized(true);

        // Optional: setup token refresh
        keycloak.onTokenExpired = () => {
          keycloak.updateToken(30).then((refreshed) => {
            if (refreshed) {
              setToken(keycloak.token);
            }
          }).catch(() => {
            keycloak.login();
          });
        };
      } catch (error) {
        // console.error('Failed to initialize Keycloak', error);
        setIsInitialized(true);
      }
    };

    initKeycloak();
  }, []);

  const login = () => keycloak.login();
  const logout = () => keycloak.logout();

  return (
    <AuthContext.Provider value={{ isAuthenticated, isInitialized, token, login, logout }}>
      {children}
    </AuthContext.Provider>
  );
};
