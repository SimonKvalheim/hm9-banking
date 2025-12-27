import { createContext, useContext, useState, useEffect, type ReactNode } from 'react';
import { setAccessToken, logout as apiLogout, tryRefreshToken } from '@/api/client';

interface AuthContextType {
  isAuthenticated: boolean;
  isLoading: boolean;
  login: (token: string) => void;
  logout: () => Promise<void>;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [isLoading, setIsLoading] = useState(true);

  // On mount, try to refresh token to check if user is logged in
  useEffect(() => {
    async function checkAuth() {
      try {
        const success = await tryRefreshToken();
        setIsAuthenticated(success);
      } catch {
        // Not logged in, that's fine
      } finally {
        setIsLoading(false);
      }
    }
    checkAuth();
  }, []);

  const login = (token: string) => {
    setAccessToken(token);
    setIsAuthenticated(true);
  };

  const logout = async () => {
    await apiLogout();
    setIsAuthenticated(false);
  };

  return (
    <AuthContext.Provider value={{ isAuthenticated, isLoading, login, logout }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
}
