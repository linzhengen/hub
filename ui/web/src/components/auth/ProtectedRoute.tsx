import React, { ReactNode, useEffect, useRef } from 'react';
import { useAuth } from '@/providers/AuthProvider';

interface ProtectedRouteProps {
  children: ReactNode;
}

const FullScreenMessage: React.FC<{ children: ReactNode }> = ({ children }) => (
  <div className="flex items-center justify-center min-h-screen">
    <div className="text-center">
      <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto"></div>
      <p className="mt-4 text-gray-600">{children}</p>
    </div>
  </div>
);

export const ProtectedRoute: React.FC<ProtectedRouteProps> = ({ children }) => {
  const { isAuthenticated, isLoading, login } = useAuth();
  const loginTriggered = useRef(false);

  // Every route in this app sits behind this guard, and there is no in-app
  // login page — Keycloak hosts it. Redirecting to an in-app path (previously
  // `<Navigate to="/" replace />`) would therefore re-enter this same guard,
  // so send the user to Keycloak instead.
  //
  // `login` is recreated on every AuthProvider render, so the effect can re-run
  // on unrelated renders; the ref keeps keycloak.login() to a single call.
  useEffect(() => {
    if (isLoading || isAuthenticated || loginTriggered.current) {
      return;
    }
    loginTriggered.current = true;
    login();
  }, [isLoading, isAuthenticated, login]);

  if (isLoading) {
    return <FullScreenMessage>Loading authentication...</FullScreenMessage>;
  }

  if (!isAuthenticated) {
    return <FullScreenMessage>Redirecting to sign in...</FullScreenMessage>;
  }

  return <>{children}</>;
};
