import { WebAuthnClient } from '@sonr.io/sdk';
import { startAuthentication, startRegistration } from '@simplewebauthn/browser';
import { useCallback, useEffect, useState, useMemo } from 'react';
import {
  OIDCClient,
  type OIDCClientConfig,
  type OIDCSession,
  clearSession,
  isSessionValid,
  loadSession,
  saveSession,
} from '../lib/oidc';
import { SIOPClient, type SIOPResponse, submitSIOPResponse } from '../lib/siop';

interface WebAuthnHookReturn {
  isLoading: boolean;
  error: string | null;
  session: OIDCSession | null;
  isAuthenticated: boolean;
  registerUser: (
    username: string,
    displayName?: string,
    email?: string,
    tel?: string,
    createVault?: boolean
  ) => Promise<WebAuthnRegistrationResult>;
  authenticateUser: (username: string) => Promise<WebAuthnAuthenticationResult>;
  handleSIOPRequest: (requestUrl: string) => Promise<SIOPResponse>;
  submitSIOPResponse: (
    response: SIOPResponse,
    redirectUri: string,
    responseMode?: 'form_post' | 'fragment' | 'query'
  ) => Promise<void>;
  initializeOIDC: (config: Partial<OIDCClientConfig>) => Promise<string>;
  handleOIDCCallback: (callbackUrl: string) => Promise<OIDCSession>;
  refreshSession: () => Promise<OIDCSession | null>;
  logout: () => void;
  getCredentials: (username: string) => Promise<WebAuthnCredential[]>;
  clearError: () => void;
}

interface WebAuthnRegistrationResult {
  success: boolean;
  userId?: string;
  did?: string;
  vaultId?: string;
  sessionId?: string;
  credential?: WebAuthnCredential;
}

interface WebAuthnAuthenticationResult {
  success: boolean;
  userId?: string;
  did?: string;
  vaultId?: string;
  sessionId?: string;
  user?: UserInfo;
}

interface WebAuthnCredential {
  id: string;
  rawId: string;
  type: string;
  publicKey: string;
  counter: number;
  createdAt: string;
}

interface UserInfo {
  sub: string;
  name?: string;
  preferred_username?: string;
  email?: string;
  email_verified?: boolean;
  did?: string;
  vault_id?: string;
}

interface ApiResponse {
  success?: boolean;
  error?: string;
  options?: unknown;
  userId?: string;
  sessionId?: string;
  user?: UserInfo;
  did?: string;
  vaultId?: string;
  credential?: WebAuthnCredential;
  credentials?: WebAuthnCredential[];
}

// Use bridge endpoints at port 8080 instead of direct 8787
const BRIDGE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

export function useWebAuthn(): WebAuthnHookReturn {
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [session, setSession] = useState<OIDCSession | null>(null);
  const [oidcClient, setOidcClient] = useState<OIDCClient | null>(null);
  const [siopClient, setSiopClient] = useState<SIOPClient | null>(null);

  // Initialize WebAuthn client from SDK
  const webAuthnClient = useMemo(() => {
    const apiUrl = process.env.NEXT_PUBLIC_CHAIN_API_URL || 'http://localhost:1317';
    return new WebAuthnClient(apiUrl);
  }, []);

  // Initialize session state
  useEffect(() => {
    const savedSession = loadSession();
    if (savedSession && isSessionValid(savedSession)) {
      setSession(savedSession);
    } else if (savedSession) {
      clearSession();
    }
  }, []);

  const clearError = useCallback(() => {
    setError(null);
  }, []);

  const isAuthenticated = session !== null && isSessionValid(session);

  const makeApiCall = useCallback(
    async (endpoint: string, data?: unknown, method = 'POST'): Promise<ApiResponse> => {
      const config: RequestInit = {
        method,
        headers: {
          'Content-Type': 'application/json',
          Accept: 'application/json',
        },
        credentials: 'include',
      };

      if (data && method !== 'GET') {
        config.body = JSON.stringify(data);
      }

      // Add session headers if available
      if (session?.accessToken) {
        config.headers = {
          ...config.headers,
          Authorization: `Bearer ${session.accessToken}`,
        };
      }

      const response = await fetch(`${BRIDGE_URL}${endpoint}`, config);

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({
          error: 'Request failed',
          error_description: `HTTP ${response.status} ${response.statusText}`,
        }));
        throw new Error(errorData.error_description || errorData.error || 'Request failed');
      }

      return response.json();
    },
    [session]
  );

  const registerUser = useCallback(
    async (
      username: string,
      displayName?: string,
      email?: string,
      tel?: string,
      createVault = true
    ): Promise<WebAuthnRegistrationResult> => {
      setIsLoading(true);
      setError(null);

      try {
        // Use SDK WebAuthn client for registration with email/tel support
        const result = await webAuthnClient.register({
          username,
          displayName,
          email,
          tel,
          createVault,
        });

        if (result.success) {
          // Fetch DID document to get full user info
          let userInfo: UserInfo | undefined;
          if (result.did) {
            try {
              const { document, metadata } = await webAuthnClient.getDIDDocument(result.did);
              userInfo = {
                sub: result.did,
                name: displayName || username,
                preferred_username: username,
                email: email,
                email_verified: !!email,
                did: result.did,
                vault_id: result.vaultId,
              };
            } catch (docError) {
              console.warn('Failed to fetch DID document:', docError);
            }
          }

          // Create OIDC session if we have user info
          if (userInfo && result.ucanToken) {
            const newSession: OIDCSession = {
              accessToken: result.ucanToken,
              tokenType: 'Bearer',
              expiresAt: Date.now() + 3600 * 1000, // 1 hour default
              userInfo,
            };

            saveSession(newSession);
            setSession(newSession);
          }

          return {
            success: true,
            userId: result.did?.split(':').pop(),
            did: result.did,
            vaultId: result.vaultId,
            sessionId: result.ucanToken,
            credential: result.credential,
          };
        }

        throw new Error(result.error || 'Registration failed');
      } catch (err) {
        console.error('WebAuthn registration error:', err);
        setError(err instanceof Error ? err.message : 'Registration failed');
        return {
          success: false,
        };
      } finally {
        setIsLoading(false);
      }
    },
    [webAuthnClient]
  );

  const authenticateUser = useCallback(
    async (username: string): Promise<WebAuthnAuthenticationResult> => {
      setIsLoading(true);
      setError(null);

      try {
        // Use SDK WebAuthn client for authentication
        const result = await webAuthnClient.authenticate({
          username,
        });

        if (result.success) {
          // Fetch DID document to get full user info
          let userInfo: UserInfo | undefined;
          if (result.did) {
            try {
              const { document, metadata } = await webAuthnClient.getDIDDocument(result.did);
              userInfo = {
                sub: result.did,
                name: username,
                preferred_username: username,
                did: result.did,
                vault_id: result.vaultId,
              };
            } catch (docError) {
              console.warn('Failed to fetch DID document:', docError);
            }
          }

          // Create OIDC session
          if (userInfo && result.sessionToken) {
            const newSession: OIDCSession = {
              accessToken: result.sessionToken,
              tokenType: 'Bearer',
              expiresAt: Date.now() + 3600 * 1000, // 1 hour default
              userInfo,
            };

            saveSession(newSession);
            setSession(newSession);
          }

          return {
            success: true,
            userId: result.did?.split(':').pop(),
            did: result.did,
            vaultId: result.vaultId,
            sessionId: result.sessionToken,
            user: userInfo,
          };
        }
        
        throw new Error(result.error || 'Authentication failed');
      } catch (err) {
        console.error('WebAuthn authentication error:', err);
        setError(err instanceof Error ? err.message : 'Authentication failed');
        return {
          success: false,
        };
      } finally {
        setIsLoading(false);
      }
    },
    [webAuthnClient]
  );

  // Get WebAuthn credentials for a user
  const getCredentials = useCallback(
    async (username: string): Promise<WebAuthnCredential[]> => {
      setIsLoading(true);
      setError(null);

      try {
        const response = await makeApiCall(`/webauthn/credentials/${username}`, undefined, 'GET');
        return response.credentials || [];
      } catch (err) {
        console.error('Get credentials error:', err);
        setError(err instanceof Error ? err.message : 'Failed to get credentials');
        return [];
      } finally {
        setIsLoading(false);
      }
    },
    [makeApiCall]
  );

  // Initialize OIDC client
  const initializeOIDC = useCallback(async (config: Partial<OIDCClientConfig>): Promise<string> => {
    setIsLoading(true);
    setError(null);

    try {
      const defaultConfig: OIDCClientConfig = {
        clientId: 'sonr-auth',
        redirectUri: `${window.location.origin}/callback`,
        scope: 'openid profile email',
        responseType: 'code',
        grantType: 'authorization_code',
        providerUrl: window.location.origin,
        usePKCE: true,
        ...config,
      };

      const client = new OIDCClient(defaultConfig);
      await client.initialize();
      setOidcClient(client);

      const authUrl = await client.authorize();
      return authUrl;
    } catch (err) {
      console.error('OIDC initialization error:', err);
      setError(err instanceof Error ? err.message : 'OIDC initialization failed');
      throw err;
    } finally {
      setIsLoading(false);
    }
  }, []);

  // Handle OIDC callback
  const handleOIDCCallback = useCallback(
    async (callbackUrl: string): Promise<OIDCSession> => {
      setIsLoading(true);
      setError(null);

      try {
        if (!oidcClient) {
          throw new Error('OIDC client not initialized');
        }

        const newSession = await oidcClient.handleCallback(callbackUrl);
        setSession(newSession);
        return newSession;
      } catch (err) {
        console.error('OIDC callback error:', err);
        setError(err instanceof Error ? err.message : 'OIDC callback failed');
        throw err;
      } finally {
        setIsLoading(false);
      }
    },
    [oidcClient]
  );

  // Refresh OIDC session
  const refreshSession = useCallback(async (): Promise<OIDCSession | null> => {
    setIsLoading(true);
    setError(null);

    try {
      if (!oidcClient) {
        return null;
      }

      const refreshedSession = await oidcClient.refreshSession();
      setSession(refreshedSession);
      return refreshedSession;
    } catch (err) {
      console.error('Session refresh error:', err);
      setError(err instanceof Error ? err.message : 'Session refresh failed');
      setSession(null);
      return null;
    } finally {
      setIsLoading(false);
    }
  }, [oidcClient]);

  // Handle SIOP request
  const handleSIOPRequest = useCallback(
    async (requestUrl: string): Promise<SIOPResponse> => {
      setIsLoading(true);
      setError(null);

      try {
        if (!session?.userInfo?.did) {
          throw new Error('No DID available for SIOP authentication');
        }

        // Initialize SIOP client if needed
        let client = siopClient;
        if (!client) {
          client = new SIOPClient(session.userInfo.did);
          await client.initialize();
          setSiopClient(client);
        }

        const response = await client.handleSIOPRequest(requestUrl);
        return response;
      } catch (err) {
        console.error('SIOP request error:', err);
        setError(err instanceof Error ? err.message : 'SIOP request failed');
        throw err;
      } finally {
        setIsLoading(false);
      }
    },
    [session, siopClient]
  );

  // Submit SIOP response
  const submitSIOPResponseCallback = useCallback(
    async (
      response: SIOPResponse,
      redirectUri: string,
      responseMode: 'form_post' | 'fragment' | 'query' = 'fragment'
    ): Promise<void> => {
      try {
        await submitSIOPResponse(response, redirectUri, responseMode);
      } catch (err) {
        console.error('SIOP response submission error:', err);
        setError(err instanceof Error ? err.message : 'SIOP response submission failed');
        throw err;
      }
    },
    []
  );

  // Logout and clear session
  const logout = useCallback(() => {
    clearSession();
    setSession(null);
    setOidcClient(null);
    setSiopClient(null);
    setError(null);
  }, []);

  return {
    isLoading,
    error,
    session,
    isAuthenticated,
    registerUser,
    authenticateUser,
    handleSIOPRequest,
    submitSIOPResponse: submitSIOPResponseCallback,
    initializeOIDC,
    handleOIDCCallback,
    refreshSession,
    logout,
    getCredentials,
    clearError,
  };
}
