'use client';

import { useEffect, useState } from 'react';

interface AuthWrapperProps {
  children: React.ReactNode;
}

export function AuthWrapper({ children }: AuthWrapperProps) {
  const [isAuthenticated, setIsAuthenticated] = useState<boolean | null>(null);

  useEffect(() => {
    // Check authentication status
    const checkAuth = () => {
      // In development mode, skip authentication for easier testing
      if (process.env.NODE_ENV === 'development') {
        setIsAuthenticated(true);
        return;
      }

      // Check for auth token in localStorage or cookie
      const authToken =
        typeof window !== 'undefined' ? localStorage.getItem('sonr_auth_token') : null;

      if (!authToken) {
        // Redirect to auth app
        window.location.href = process.env.NEXT_PUBLIC_AUTH_URL || 'https://auth.sonr.io';
        return;
      }

      setIsAuthenticated(true);
    };

    checkAuth();
  }, []);

  // Show loading while checking auth
  if (isAuthenticated === null) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900" />
      </div>
    );
  }

  // Show children if authenticated
  if (isAuthenticated) {
    return <>{children}</>;
  }

  // Should not reach here due to redirect
  return null;
}
