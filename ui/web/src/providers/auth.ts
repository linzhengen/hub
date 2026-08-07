import { createContext, useContext } from "react";

export interface AppUser {
  id?: string;
  name?: string;
  email?: string;
  emailVerified: boolean;
  roles: string[];
}

export interface AuthContextType {
  isAuthenticated: boolean;
  isLoading: boolean;
  user: AppUser | null;
  login: () => Promise<void>;
  logout: () => Promise<void>;
}

export const AuthContext = createContext<AuthContextType | undefined>(undefined);

// AuthProvider と別ファイルにしているのは、コンポーネント以外を同居させると
// Fast Refresh が効かなくなるため（react-refresh/only-export-components）。
export const useAuth = () => {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
};
