import React from 'react';
import { Link } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';

const NavBar: React.FC = () => {
  const { isAuthenticated, logout } = useAuth();

  if (!isAuthenticated) {
    return null;
  }

  return (
    <nav style={{ padding: '1rem', backgroundColor: '#f0f0f0', display: 'flex', justifyContent: 'space-between' }}>
      <div>
        <Link to="/users" style={{ marginRight: '1rem' }}>Users</Link>
        <Link to="/systems" style={{ marginRight: '1rem' }}>Systems</Link>
      </div>
      <button onClick={logout}>Logout</button>
    </nav>
  );
};

export default NavBar;
