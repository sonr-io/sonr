// Temporary stub to fix module resolution - TODO: Replace with actual @sonr.io/es/client import
const getAccount = async (params: { address: string; rpcEndpoint: string }): Promise<any> => {
  console.warn('Using stub implementation for getAccount');
  return {
    address: params.address,
    accountNumber: '1',
    sequence: '0',
    pubKey: null,
  };
};

import type { ApiResponse, AuthStatus, User } from '@sonr.io/com/types';

/**
 * Authentication API Client
 * Integrates with the web/auth WebAuthn authentication system
 */
export class AuthApiClient {
  private authUrl: string;
  private rpcEndpoint: string;

  constructor(authUrl: string, rpcEndpoint: string) {
    this.authUrl = authUrl;
    this.rpcEndpoint = rpcEndpoint;
  }

  /**
   * Check if user is authenticated
   */
  async checkAuthStatus(): Promise<AuthStatus> {
    try {
      // Check for stored session
      const session = this.getStoredSession();

      if (!session) {
        return {
          isAuthenticated: false,
          user: null,
        };
      }

      // Verify session is still valid by checking account on chain
      const account = await getAccount({
        address: session.address,
        rpcEndpoint: this.rpcEndpoint,
      });

      if (account) {
        return {
          isAuthenticated: true,
          user: {
            id: session.userId,
            address: session.address,
            username: session.username,
            did: session.did,
            createdAt: session.createdAt,
          },
        };
      }

      // Session invalid, clear it
      this.clearSession();
      return {
        isAuthenticated: false,
        user: null,
      };
    } catch (error) {
      console.error('Auth status check failed:', error);
      return {
        isAuthenticated: false,
        user: null,
      };
    }
  }

  /**
   * Redirect to authentication app
   */
  redirectToAuth(returnUrl?: string): void {
    const currentUrl = returnUrl || window.location.href;
    const authRedirectUrl = `${this.authUrl}/login?returnUrl=${encodeURIComponent(currentUrl)}`;
    window.location.href = authRedirectUrl;
  }

  /**
   * Handle authentication callback
   */
  async handleAuthCallback(params: URLSearchParams): Promise<ApiResponse<User>> {
    try {
      const token = params.get('token');
      const address = params.get('address');
      const username = params.get('username');
      const did = params.get('did');

      if (!token || !address) {
        return {
          success: false,
          error: 'Missing authentication parameters',
        };
      }

      // Verify the token with the auth service
      const verified = await this.verifyAuthToken(token, address);

      if (!verified) {
        return {
          success: false,
          error: 'Invalid authentication token',
        };
      }

      // Store session
      const user: User = {
        id: did || address,
        address,
        username: username || 'User',
        did: did || null,
        createdAt: new Date().toISOString(),
      };

      this.storeSession({
        userId: user.id,
        address: user.address,
        username: user.username,
        did: user.did,
        token,
        createdAt: user.createdAt,
      });

      return {
        success: true,
        data: user,
      };
    } catch (error) {
      return {
        success: false,
        error: error instanceof Error ? error.message : 'Authentication failed',
      };
    }
  }

  /**
   * Verify authentication token with auth service
   */
  private async verifyAuthToken(token: string, address: string): Promise<boolean> {
    try {
      const response = await fetch(`${this.authUrl}/api/verify`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ token, address }),
      });

      if (!response.ok) {
        return false;
      }

      const data = await response.json();
      return data.valid === true;
    } catch (error) {
      console.error('Token verification failed:', error);
      return false;
    }
  }

  /**
   * Sign out the current user
   */
  async signOut(): Promise<void> {
    try {
      // Clear local session
      this.clearSession();

      // Notify auth service
      await fetch(`${this.authUrl}/api/logout`, {
        method: 'POST',
        credentials: 'include',
      });
    } catch (error) {
      console.error('Sign out error:', error);
    }
  }

  /**
   * Get current user from session
   */
  getCurrentUser(): User | null {
    const session = this.getStoredSession();

    if (!session) {
      return null;
    }

    return {
      id: session.userId,
      address: session.address,
      username: session.username,
      did: session.did,
      createdAt: session.createdAt,
    };
  }

  /**
   * Store session in localStorage
   */
  private storeSession(session: any): void {
    if (typeof window !== 'undefined') {
      localStorage.setItem('sonr_auth_session', JSON.stringify(session));
    }
  }

  /**
   * Get stored session from localStorage
   */
  private getStoredSession(): any {
    if (typeof window === 'undefined') {
      return null;
    }

    try {
      const sessionStr = localStorage.getItem('sonr_auth_session');
      return sessionStr ? JSON.parse(sessionStr) : null;
    } catch {
      return null;
    }
  }

  /**
   * Clear stored session
   */
  private clearSession(): void {
    if (typeof window !== 'undefined') {
      localStorage.removeItem('sonr_auth_session');
    }
  }
}

// Singleton instance
let authClient: AuthApiClient | null = null;

/**
 * Get or create the auth API client
 */
export function getAuthApiClient(): AuthApiClient {
  const authUrl = process.env.NEXT_PUBLIC_AUTH_URL || 'http://localhost:3001';
  const rpcEndpoint = process.env.NEXT_PUBLIC_RPC_ENDPOINT || 'http://localhost:26657';

  if (!authClient) {
    authClient = new AuthApiClient(authUrl, rpcEndpoint);
  }

  return authClient;
}

/**
 * Authentication helper functions
 */
export const authApi = {
  /**
   * Check if user is authenticated
   */
  async isAuthenticated(): Promise<boolean> {
    const client = getAuthApiClient();
    const status = await client.checkAuthStatus();
    return status.isAuthenticated;
  },

  /**
   * Get current user
   */
  getCurrentUser(): User | null {
    const client = getAuthApiClient();
    return client.getCurrentUser();
  },

  /**
   * Require authentication (redirect if not authenticated)
   */
  async requireAuth(): Promise<User | null> {
    const client = getAuthApiClient();
    const status = await client.checkAuthStatus();

    if (!status.isAuthenticated) {
      client.redirectToAuth();
      return null;
    }

    return status.user;
  },

  /**
   * Sign in (redirect to auth app)
   */
  signIn(returnUrl?: string): void {
    const client = getAuthApiClient();
    client.redirectToAuth(returnUrl);
  },

  /**
   * Sign out
   */
  async signOut(): Promise<void> {
    const client = getAuthApiClient();
    await client.signOut();
    window.location.href = '/';
  },

  /**
   * Handle OAuth-style callback
   */
  async handleCallback(): Promise<User | null> {
    const client = getAuthApiClient();
    const params = new URLSearchParams(window.location.search);

    if (!params.has('token')) {
      return null;
    }

    const result = await client.handleAuthCallback(params);

    if (result.success) {
      // Clear URL parameters
      window.history.replaceState({}, '', window.location.pathname);
      return result.data;
    }

    return null;
  },
};

/**
 * React hook helper for authentication
 */
export function useAuthCheck() {
  if (typeof window === 'undefined') {
    return { loading: true, authenticated: false, user: null };
  }

  const user = authApi.getCurrentUser();

  return {
    loading: false,
    authenticated: !!user,
    user,
  };
}

export default authApi;
