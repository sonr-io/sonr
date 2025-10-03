import { useCallback, useEffect, useState } from 'react';

interface User {
  id: string;
  username: string;
  displayName?: string;
  createdAt: string;
}

interface SessionHookReturn {
  user: User | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  error: string | null;
  logout: () => Promise<void>;
  checkSession: () => Promise<void>;
}

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

export function useSession(): SessionHookReturn {
  const [user, setUser] = useState<User | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const isAuthenticated = user !== null;

  const makeAuthenticatedRequest = useCallback(
    async (endpoint: string, options: RequestInit = {}) => {
      const sessionToken = localStorage.getItem('highway_session');

      if (!sessionToken) {
        throw new Error('No session token found');
      }

      const response = await fetch(`${API_BASE_URL}${endpoint}`, {
        ...options,
        headers: {
          ...options.headers,
          Authorization: `Bearer ${sessionToken}`,
          'Content-Type': 'application/json',
        },
      });

      if (!response.ok) {
        if (response.status === 401) {
          // Session expired or invalid
          localStorage.removeItem('highway_session');
          setUser(null);
          throw new Error('Session expired');
        }
        const errorData = await response.json();
        throw new Error(errorData.error || 'Request failed');
      }

      return response.json();
    },
    []
  );

  const checkSession = useCallback(async () => {
    setIsLoading(true);
    setError(null);

    try {
      const sessionToken = localStorage.getItem('highway_session');

      if (!sessionToken) {
        setUser(null);
        return;
      }

      // Validate session by getting user profile
      const response = await makeAuthenticatedRequest('/auth/profile');

      if (response.user) {
        setUser(response.user);
      } else {
        setUser(null);
        localStorage.removeItem('highway_session');
      }
    } catch (err) {
      console.error('Session check error:', err);
      setError(err instanceof Error ? err.message : 'Session check failed');
      setUser(null);
      localStorage.removeItem('highway_session');
    } finally {
      setIsLoading(false);
    }
  }, [makeAuthenticatedRequest]);

  const logout = useCallback(async () => {
    setIsLoading(true);
    setError(null);

    try {
      // Call logout endpoint to invalidate session on server
      await makeAuthenticatedRequest('/auth/logout', {
        method: 'POST',
      });
    } catch (err) {
      console.error('Logout error:', err);
      // Continue with local logout even if server call fails
    } finally {
      // Always clear local session
      localStorage.removeItem('highway_session');
      setUser(null);
      setIsLoading(false);
    }
  }, [makeAuthenticatedRequest]);

  // Check session on mount and when the component is focused
  useEffect(() => {
    checkSession();

    const handleFocus = () => {
      checkSession();
    };

    const handleStorageChange = (e: StorageEvent) => {
      if (e.key === 'highway_session') {
        if (e.newValue === null) {
          setUser(null);
        } else {
          checkSession();
        }
      }
    };

    window.addEventListener('focus', handleFocus);
    window.addEventListener('storage', handleStorageChange);

    return () => {
      window.removeEventListener('focus', handleFocus);
      window.removeEventListener('storage', handleStorageChange);
    };
  }, [checkSession]);

  return {
    user,
    isAuthenticated,
    isLoading,
    error,
    logout,
    checkSession,
  };
}
