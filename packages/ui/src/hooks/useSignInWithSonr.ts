'use client';

/**
 * React hook for Sonr OAuth2 authentication
 */

import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import {
  OAuth2Client,
  OAuth2ClientError,
  type OAuth2Config,
  type OAuth2Token,
  type OAuth2UserInfo,
  isTokenExpired,
  parseCallbackUrl,
} from '../lib/oauth';

/**
 * Hook state
 */
export interface UseSignInWithSonrState {
  isLoading: boolean;
  isAuthenticated: boolean;
  user: OAuth2UserInfo | null;
  token: OAuth2Token | null;
  error: Error | null;
}

/**
 * Hook options
 */
export interface UseSignInWithSonrOptions extends Partial<OAuth2Config> {
  /**
   * Auto refresh token before expiry
   */
  autoRefresh?: boolean;
  /**
   * Auto refresh buffer time in seconds
   */
  refreshBuffer?: number;
  /**
   * Callback on successful authentication
   */
  onSuccess?: (user: OAuth2UserInfo, token: OAuth2Token) => void;
  /**
   * Callback on authentication error
   */
  onError?: (error: Error) => void;
  /**
   * Callback on logout
   */
  onLogout?: () => void;
  /**
   * Storage key prefix
   */
  storageKeyPrefix?: string;
}

/**
 * Hook return type
 */
export interface UseSignInWithSonrReturn extends UseSignInWithSonrState {
  /**
   * Initiate OAuth2 authorization flow
   */
  signIn: (state?: string) => Promise<void>;
  /**
   * Handle OAuth2 callback
   */
  handleCallback: (callbackUrl?: string) => Promise<OAuth2UserInfo | null>;
  /**
   * Refresh access token
   */
  refresh: () => Promise<OAuth2Token | null>;
  /**
   * Get user info
   */
  getUser: () => Promise<OAuth2UserInfo | null>;
  /**
   * Logout and clear tokens
   */
  signOut: () => Promise<void>;
  /**
   * Get access token
   */
  getAccessToken: () => string | null;
  /**
   * OAuth2 client instance
   */
  client: OAuth2Client;
}

/**
 * React hook for Sonr OAuth2 authentication
 */
export function useSignInWithSonr(
  config: OAuth2Config,
  options: UseSignInWithSonrOptions = {}
): UseSignInWithSonrReturn {
  const {
    autoRefresh = true,
    refreshBuffer = 60,
    onSuccess,
    onError,
    onLogout,
    storageKeyPrefix = 'sonr_oauth',
    ...configOverrides
  } = options;

  // OAuth2 client
  const client = useMemo(
    () => new OAuth2Client({ ...config, ...configOverrides }),
    [config, configOverrides]
  );

  // State
  const [state, setState] = useState<UseSignInWithSonrState>({
    isLoading: false,
    isAuthenticated: false,
    user: null,
    token: null,
    error: null,
  });

  // Refs for callbacks
  const refreshTimeoutRef = useRef<NodeJS.Timeout>();
  const isRefreshingRef = useRef(false);

  /**
   * Load stored token and user on mount
   */
  useEffect(() => {
    const loadStoredAuth = async () => {
      try {
        const storedToken = localStorage.getItem(`${storageKeyPrefix}_token`);
        const storedUser = localStorage.getItem(`${storageKeyPrefix}_user`);

        if (storedToken && storedUser) {
          const token = JSON.parse(storedToken) as OAuth2Token;
          const user = JSON.parse(storedUser) as OAuth2UserInfo;

          // Check if token is expired
          if (!isTokenExpired(token)) {
            setState((prev) => ({
              ...prev,
              isAuthenticated: true,
              token,
              user,
            }));

            // Setup auto refresh
            if (autoRefresh && token.expires_in) {
              scheduleTokenRefresh(token.expires_in);
            }
          } else {
            // Try to refresh if we have a refresh token
            if (token.refresh_token) {
              await refresh();
            } else {
              // Clear expired auth
              await signOut();
            }
          }
        }
      } catch (error) {
        console.error('Failed to load stored auth:', error);
      }
    };

    loadStoredAuth();
  }, []);

  /**
   * Schedule token refresh
   */
  const scheduleTokenRefresh = useCallback(
    (expiresIn: number) => {
      if (refreshTimeoutRef.current) {
        clearTimeout(refreshTimeoutRef.current);
      }

      // Refresh token before it expires
      const refreshIn = Math.max(0, (expiresIn - refreshBuffer) * 1000);

      refreshTimeoutRef.current = setTimeout(async () => {
        if (!isRefreshingRef.current) {
          await refresh();
        }
      }, refreshIn);
    },
    [refreshBuffer]
  );

  /**
   * Clear refresh timeout
   */
  const clearRefreshTimeout = useCallback(() => {
    if (refreshTimeoutRef.current) {
      clearTimeout(refreshTimeoutRef.current);
      refreshTimeoutRef.current = undefined;
    }
  }, []);

  /**
   * Sign in - initiate OAuth2 flow
   */
  const signIn = useCallback(
    async (authState?: string) => {
      setState((prev) => ({ ...prev, isLoading: true, error: null }));

      try {
        const authUrl = await client.getAuthorizationUrl(authState);
        window.location.href = authUrl;
      } catch (error) {
        const err = error as Error;
        setState((prev) => ({ ...prev, isLoading: false, error: err }));
        onError?.(err);
      }
    },
    [client, onError]
  );

  /**
   * Handle OAuth2 callback
   */
  const handleCallback = useCallback(
    async (callbackUrl?: string): Promise<OAuth2UserInfo | null> => {
      setState((prev) => ({ ...prev, isLoading: true, error: null }));

      try {
        // Parse callback URL
        const url = callbackUrl || window.location.href;
        const params = parseCallbackUrl(url);

        // Check for errors
        if (params.error) {
          throw new OAuth2ClientError(params.error, params.error_description);
        }

        // Exchange code for tokens
        if (!params.code) {
          throw new OAuth2ClientError('invalid_request', 'No authorization code in callback');
        }

        const token = await client.exchangeCode(params.code, params.state);

        // Get user info
        const user = await client.getUserInfo(token.access_token);

        // Store auth data
        localStorage.setItem(`${storageKeyPrefix}_token`, JSON.stringify(token));
        localStorage.setItem(`${storageKeyPrefix}_user`, JSON.stringify(user));

        // Update state
        setState((prev) => ({
          ...prev,
          isLoading: false,
          isAuthenticated: true,
          token,
          user,
        }));

        // Setup auto refresh
        if (autoRefresh && token.expires_in) {
          scheduleTokenRefresh(token.expires_in);
        }

        // Call success callback
        onSuccess?.(user, token);

        // Clear callback params from URL
        if (!callbackUrl) {
          window.history.replaceState({}, document.title, window.location.pathname);
        }

        return user;
      } catch (error) {
        const err = error as Error;
        setState((prev) => ({
          ...prev,
          isLoading: false,
          isAuthenticated: false,
          error: err,
        }));
        onError?.(err);
        return null;
      }
    },
    [client, storageKeyPrefix, autoRefresh, scheduleTokenRefresh, onSuccess, onError]
  );

  /**
   * Refresh access token
   */
  const refresh = useCallback(async (): Promise<OAuth2Token | null> => {
    if (isRefreshingRef.current) {
      return state.token;
    }

    isRefreshingRef.current = true;
    setState((prev) => ({ ...prev, isLoading: true, error: null }));

    try {
      const token = await client.refreshToken();

      // Get updated user info
      const user = await client.getUserInfo(token.access_token);

      // Store updated auth data
      localStorage.setItem(`${storageKeyPrefix}_token`, JSON.stringify(token));
      localStorage.setItem(`${storageKeyPrefix}_user`, JSON.stringify(user));

      // Update state
      setState((prev) => ({
        ...prev,
        isLoading: false,
        token,
        user,
      }));

      // Reschedule auto refresh
      if (autoRefresh && token.expires_in) {
        scheduleTokenRefresh(token.expires_in);
      }

      return token;
    } catch (error) {
      const err = error as Error;
      setState((prev) => ({
        ...prev,
        isLoading: false,
        error: err,
      }));

      // If refresh fails, sign out
      await signOut();

      onError?.(err);
      return null;
    } finally {
      isRefreshingRef.current = false;
    }
  }, [client, state.token, storageKeyPrefix, autoRefresh, scheduleTokenRefresh, onError]);

  /**
   * Get user info
   */
  const getUser = useCallback(async (): Promise<OAuth2UserInfo | null> => {
    if (!state.isAuthenticated) {
      return null;
    }

    setState((prev) => ({ ...prev, isLoading: true, error: null }));

    try {
      const user = await client.getUserInfo();

      // Update stored user
      localStorage.setItem(`${storageKeyPrefix}_user`, JSON.stringify(user));

      // Update state
      setState((prev) => ({
        ...prev,
        isLoading: false,
        user,
      }));

      return user;
    } catch (error) {
      const err = error as Error;
      setState((prev) => ({ ...prev, isLoading: false, error: err }));
      onError?.(err);
      return null;
    }
  }, [client, state.isAuthenticated, storageKeyPrefix, onError]);

  /**
   * Sign out
   */
  const signOut = useCallback(async () => {
    setState((prev) => ({ ...prev, isLoading: true, error: null }));

    try {
      // Revoke tokens
      await client.logout();
    } catch (error) {
      console.error('Failed to revoke tokens:', error);
    } finally {
      // Clear refresh timeout
      clearRefreshTimeout();

      // Clear stored auth data
      localStorage.removeItem(`${storageKeyPrefix}_token`);
      localStorage.removeItem(`${storageKeyPrefix}_user`);
      sessionStorage.removeItem(`${storageKeyPrefix}_code_verifier`);

      // Reset state
      setState({
        isLoading: false,
        isAuthenticated: false,
        user: null,
        token: null,
        error: null,
      });

      // Call logout callback
      onLogout?.();
    }
  }, [client, storageKeyPrefix, clearRefreshTimeout, onLogout]);

  /**
   * Get access token
   */
  const getAccessToken = useCallback((): string | null => {
    return state.token?.access_token || null;
  }, [state.token]);

  /**
   * Cleanup on unmount
   */
  useEffect(() => {
    return () => {
      clearRefreshTimeout();
    };
  }, [clearRefreshTimeout]);

  return {
    ...state,
    signIn,
    handleCallback,
    refresh,
    getUser,
    signOut,
    getAccessToken,
    client,
  };
}

/**
 * Default export
 */
export default useSignInWithSonr;
