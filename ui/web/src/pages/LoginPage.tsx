import React from 'react';
import { useAuth } from '../contexts/AuthContext';

const LoginPage: React.FC = () => {
  const { login, isLoading, error } = useAuth();

  return (
    <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100vh' }}>
      <div style={{ textAlign: 'center' }}>
        <h1>Login</h1>
        <p>Please log in to access the application.</p>
        {error && <p style={{ color: 'red' }}>Error: {error.message}</p>}
        <button onClick={login} disabled={isLoading}>
          {isLoading ? 'Loading...' : 'Login with Keycloak'}
        </button>
      </div>
    </div>
  );
};

export default LoginPage;
