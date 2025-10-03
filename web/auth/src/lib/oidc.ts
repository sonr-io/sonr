/**
 * OpenID Connect (OIDC) Utilities
 *
 * This module provides comprehensive OIDC client utilities for the Next.js auth application.
 * It supports standard OIDC flows with PKCE, token validation, and session management.
 *
 * Features:
 * - Authorization Code Flow with PKCE
 * - Token validation and management
 * - Discovery document fetching
 * - JWT decoding and validation
 * - Session management
 * - CORS-enabled fetch wrapper
 */

// Utility function to generate random string for PKCE and nonces
export function generateRandomString(length: number): string {
  const charset = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-._~';
  const randomValues = new Uint8Array(length);
  crypto.getRandomValues(randomValues);
  return Array.from(randomValues, (byte) => charset[byte % charset.length]).join('');
}

// ============================================================================
// Types and Interfaces
// ============================================================================

/**
 * OIDC Provider Configuration from Discovery Document
 */
export interface OIDCProviderConfig {
  issuer: string;
  authorization_endpoint: string;
  token_endpoint: string;
  userinfo_endpoint: string;
  jwks_uri: string;
  registration_endpoint?: string;
  scopes_supported: string[];
  response_types_supported: string[];
  response_modes_supported?: string[];
  grant_types_supported: string[];
  subject_types_supported: string[];
  id_token_signing_alg_values_supported: string[];
  token_endpoint_auth_methods_supported?: string[];
  claims_supported?: string[];
  code_challenge_methods_supported?: string[];
  // SIOP specific fields
  subject_syntax_types_supported?: string[];
  id_token_types_supported?: string[];
  request_object_signing_alg_values_supported?: string[];
}

/**
 * OIDC Client Configuration
 */
export interface OIDCClientConfig {
  clientId: string;
  clientSecret?: string;
  redirectUri: string;
  scope: string;
  responseType: string;
  grantType: string;
  providerUrl: string;
  usePKCE: boolean;
  additionalParams?: Record<string, string>;
}

/**
 * PKCE Parameters for Authorization Code Flow
 */
export interface PKCEParams {
  codeVerifier: string;
  codeChallenge: string;
  codeChallengeMethod: string;
}

/**
 * Authorization Request Parameters
 */
export interface AuthorizationParams {
  response_type: string;
  client_id: string;
  redirect_uri: string;
  scope: string;
  state: string;
  nonce: string;
  code_challenge?: string;
  code_challenge_method?: string;
  response_mode?: string;
  prompt?: string;
  max_age?: number;
  ui_locales?: string;
  claims?: string;
}

/**
 * Token Request Parameters
 */
export interface TokenRequestParams {
  grant_type: string;
  code?: string;
  redirect_uri?: string;
  client_id: string;
  client_secret?: string;
  code_verifier?: string;
  refresh_token?: string;
  scope?: string;
}

/**
 * OIDC Token Response
 */
export interface TokenResponse {
  access_token: string;
  token_type: string;
  expires_in: number;
  refresh_token?: string;
  id_token?: string;
  scope?: string;
}

/**
 * OIDC UserInfo Response
 */
export interface UserInfo {
  sub: string;
  name?: string;
  preferred_username?: string;
  email?: string;
  email_verified?: boolean;
  did?: string;
  vault_id?: string;
  updated_at?: number;
  claims?: Record<string, unknown>;
}

/**
 * JWT Header
 */
export interface JWTHeader {
  alg: string;
  typ: string;
  kid?: string;
  jku?: string;
  x5u?: string;
}

/**
 * JWT Payload (ID Token Claims)
 */
export interface JWTPayload {
  iss: string;
  sub: string;
  aud: string | string[];
  exp: number;
  iat: number;
  nbf?: number;
  jti?: string;
  auth_time?: number;
  nonce?: string;
  acr?: string;
  amr?: string[];
  azp?: string;
  // Standard claims
  name?: string;
  preferred_username?: string;
  email?: string;
  email_verified?: boolean;
  // Custom claims
  did?: string;
  vault_id?: string;
  [key: string]: unknown;
}

/**
 * OIDC Session Data
 */
export interface OIDCSession {
  accessToken: string;
  refreshToken?: string;
  idToken?: string;
  expiresAt: number;
  userInfo?: UserInfo;
  tokenType: string;
  scope?: string;
}

/**
 * OIDC Error Response
 */
export interface OIDCError {
  error: string;
  error_description?: string;
  error_uri?: string;
  state?: string;
}

// ============================================================================
// Constants and Configuration
// ============================================================================

const OIDC_DEFAULTS = {
  SCOPE: 'openid profile email',
  RESPONSE_TYPE: 'code',
  GRANT_TYPE: 'authorization_code',
  CODE_CHALLENGE_METHOD: 'S256',
  TOKEN_TYPE: 'Bearer',
  SESSION_STORAGE_KEY: 'oidc_session',
  STATE_STORAGE_KEY: 'oidc_state',
  NONCE_STORAGE_KEY: 'oidc_nonce',
  PKCE_STORAGE_KEY: 'oidc_pkce',
} as const;

// ============================================================================
// CORS-Enabled Fetch Wrapper
// ============================================================================

/**
 * CORS-enabled fetch wrapper for OIDC endpoints
 */
export async function oidcFetch(url: string, options: RequestInit = {}): Promise<Response> {
  const defaultHeaders = {
    'Content-Type': 'application/json',
    Accept: 'application/json',
  };

  const config: RequestInit = {
    ...options,
    headers: {
      ...defaultHeaders,
      ...options.headers,
    },
    credentials: 'include',
    mode: 'cors',
  };

  try {
    const response = await fetch(url, config);
    return response;
  } catch (error) {
    console.error('OIDC Fetch Error:', error);
    throw new Error(`Network error: ${error instanceof Error ? error.message : 'Unknown error'}`);
  }
}

// ============================================================================
// PKCE Utilities
// ============================================================================

/**
 * Generate PKCE parameters for authorization code flow
 */
export function generatePKCE(): PKCEParams {
  const codeVerifier = generateRandomString(128);

  // Synchronous fallback using base64url encoding
  const codeChallenge = btoa(codeVerifier)
    .replace(/\+/g, '-')
    .replace(/\//g, '_')
    .replace(/=/g, '');

  return {
    codeVerifier,
    codeChallenge,
    codeChallengeMethod: OIDC_DEFAULTS.CODE_CHALLENGE_METHOD,
  };
}

/**
 * Store PKCE parameters in session storage
 */
export function storePKCE(pkce: PKCEParams): void {
  if (typeof window !== 'undefined') {
    sessionStorage.setItem(OIDC_DEFAULTS.PKCE_STORAGE_KEY, JSON.stringify(pkce));
  }
}

/**
 * Retrieve and clear PKCE parameters from session storage
 */
export function retrieveAndClearPKCE(): PKCEParams | null {
  if (typeof window === 'undefined') return null;

  const stored = sessionStorage.getItem(OIDC_DEFAULTS.PKCE_STORAGE_KEY);
  if (!stored) return null;

  sessionStorage.removeItem(OIDC_DEFAULTS.PKCE_STORAGE_KEY);

  try {
    return JSON.parse(stored);
  } catch {
    return null;
  }
}

// ============================================================================
// Discovery Document
// ============================================================================

/**
 * Fetch OIDC discovery document from provider
 */
export async function fetchDiscoveryDocument(providerUrl: string): Promise<OIDCProviderConfig> {
  const discoveryUrl = `${providerUrl}/.well-known/openid-configuration`;

  const response = await oidcFetch(discoveryUrl, {
    method: 'GET',
    headers: {
      'Cache-Control': 'max-age=3600',
    },
  });

  if (!response.ok) {
    throw new Error(
      `Failed to fetch discovery document: ${response.status} ${response.statusText}`
    );
  }

  const config = (await response.json()) as OIDCProviderConfig;

  // Validate required endpoints
  if (!config.authorization_endpoint || !config.token_endpoint) {
    throw new Error('Invalid discovery document: missing required endpoints');
  }

  return config;
}

// ============================================================================
// Authorization URL Builder
// ============================================================================

/**
 * Build authorization URL with PKCE support
 */
export async function buildAuthorizationUrl(
  config: OIDCClientConfig,
  providerConfig?: OIDCProviderConfig
): Promise<{ url: string; state: string; nonce: string; pkce?: PKCEParams }> {
  // Fetch discovery document if not provided
  let discovery = providerConfig;
  if (!discovery) {
    discovery = await fetchDiscoveryDocument(config.providerUrl);
  }

  // Generate state and nonce
  const state = generateRandomString(32);
  const nonce = generateRandomString(32);

  // Generate PKCE parameters if enabled
  let pkce: PKCEParams | undefined;
  if (config.usePKCE) {
    pkce = generatePKCE();
    storePKCE(pkce);
  }

  // Store state and nonce
  if (typeof window !== 'undefined') {
    sessionStorage.setItem(OIDC_DEFAULTS.STATE_STORAGE_KEY, state);
    sessionStorage.setItem(OIDC_DEFAULTS.NONCE_STORAGE_KEY, nonce);
  }

  // Build authorization parameters
  const authParams: AuthorizationParams = {
    response_type: config.responseType,
    client_id: config.clientId,
    redirect_uri: config.redirectUri,
    scope: config.scope,
    state,
    nonce,
    ...config.additionalParams,
  };

  // Add PKCE parameters if enabled
  if (pkce) {
    authParams.code_challenge = pkce.codeChallenge;
    authParams.code_challenge_method = pkce.codeChallengeMethod;
  }

  // Build query string
  const params = new URLSearchParams();
  Object.entries(authParams).forEach(([key, value]) => {
    if (value !== undefined && value !== '') {
      params.append(key, value.toString());
    }
  });

  const url = `${discovery.authorization_endpoint}?${params.toString()}`;

  return { url, state, nonce, pkce };
}

// ============================================================================
// Token Exchange
// ============================================================================

/**
 * Exchange authorization code for tokens
 */
export async function exchangeCodeForTokens(
  code: string,
  config: OIDCClientConfig,
  providerConfig?: OIDCProviderConfig
): Promise<TokenResponse> {
  // Fetch discovery document if not provided
  let discovery = providerConfig;
  if (!discovery) {
    discovery = await fetchDiscoveryDocument(config.providerUrl);
  }

  // Retrieve PKCE parameters if used
  const pkce = config.usePKCE ? retrieveAndClearPKCE() : undefined;

  // Build token request parameters
  const tokenParams: TokenRequestParams = {
    grant_type: config.grantType,
    code,
    redirect_uri: config.redirectUri,
    client_id: config.clientId,
  };

  // Add client secret if provided
  if (config.clientSecret) {
    tokenParams.client_secret = config.clientSecret;
  }

  // Add PKCE code verifier if used
  if (pkce) {
    tokenParams.code_verifier = pkce.codeVerifier;
  }

  // Prepare form data (OIDC spec requires form encoding)
  const formData = new URLSearchParams();
  Object.entries(tokenParams).forEach(([key, value]) => {
    if (value !== undefined && value !== '') {
      formData.append(key, value);
    }
  });

  const response = await oidcFetch(discovery.token_endpoint, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/x-www-form-urlencoded',
    },
    body: formData,
  });

  if (!response.ok) {
    const errorData = await response.json().catch(() => ({
      error: 'token_error',
      error_description: 'Token exchange failed',
    }));
    throw new Error(`Token exchange failed: ${errorData.error_description || errorData.error}`);
  }

  return response.json() as Promise<TokenResponse>;
}

/**
 * Refresh access token using refresh token
 */
export async function refreshAccessToken(
  refreshToken: string,
  config: OIDCClientConfig,
  providerConfig?: OIDCProviderConfig
): Promise<TokenResponse> {
  // Fetch discovery document if not provided
  let discovery = providerConfig;
  if (!discovery) {
    discovery = await fetchDiscoveryDocument(config.providerUrl);
  }

  const tokenParams: TokenRequestParams = {
    grant_type: 'refresh_token',
    refresh_token: refreshToken,
    client_id: config.clientId,
    scope: config.scope,
  };

  if (config.clientSecret) {
    tokenParams.client_secret = config.clientSecret;
  }

  const formData = new URLSearchParams();
  Object.entries(tokenParams).forEach(([key, value]) => {
    if (value !== undefined && value !== '') {
      formData.append(key, value);
    }
  });

  const response = await oidcFetch(discovery.token_endpoint, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/x-www-form-urlencoded',
    },
    body: formData,
  });

  if (!response.ok) {
    const errorData = await response.json().catch(() => ({
      error: 'refresh_error',
      error_description: 'Token refresh failed',
    }));
    throw new Error(`Token refresh failed: ${errorData.error_description || errorData.error}`);
  }

  return response.json() as Promise<TokenResponse>;
}

// ============================================================================
// UserInfo Fetching
// ============================================================================

/**
 * Fetch user information using access token
 */
export async function fetchUserInfo(
  accessToken: string,
  config: OIDCClientConfig,
  providerConfig?: OIDCProviderConfig
): Promise<UserInfo> {
  // Fetch discovery document if not provided
  let discovery = providerConfig;
  if (!discovery) {
    discovery = await fetchDiscoveryDocument(config.providerUrl);
  }

  const response = await oidcFetch(discovery.userinfo_endpoint, {
    method: 'GET',
    headers: {
      Authorization: `Bearer ${accessToken}`,
    },
  });

  if (!response.ok) {
    const errorData = await response.json().catch(() => ({
      error: 'userinfo_error',
      error_description: 'UserInfo request failed',
    }));
    throw new Error(`UserInfo request failed: ${errorData.error_description || errorData.error}`);
  }

  return response.json() as Promise<UserInfo>;
}

// ============================================================================
// JWT Utilities
// ============================================================================

/**
 * Decode JWT token without verification (for inspection only)
 * WARNING: This does not validate the token signature - use validateJWT for production
 */
export function decodeJWT(token: string): {
  header: JWTHeader;
  payload: JWTPayload;
  signature: string;
} {
  const parts = token.split('.');
  if (parts.length !== 3) {
    throw new Error('Invalid JWT format');
  }

  const [headerPart, payloadPart, signature] = parts;

  if (!headerPart || !payloadPart || !signature) {
    throw new Error('Invalid JWT format: missing parts');
  }

  try {
    const header = JSON.parse(atob(headerPart.replace(/-/g, '+').replace(/_/g, '/'))) as JWTHeader;
    const payload = JSON.parse(
      atob(payloadPart.replace(/-/g, '+').replace(/_/g, '/'))
    ) as JWTPayload;

    return { header, payload, signature };
  } catch {
    throw new Error('Failed to decode JWT: Invalid encoding');
  }
}

/**
 * Validate JWT token (basic validation without signature verification)
 * For full validation, use a proper JWT library with JWKS support
 */
export function validateJWTClaims(
  token: string,
  expectedIssuer: string,
  expectedAudience: string,
  expectedNonce?: string
): JWTPayload {
  const { payload } = decodeJWT(token);

  // Validate issuer
  if (payload.iss !== expectedIssuer) {
    throw new Error(`Invalid issuer: expected ${expectedIssuer}, got ${payload.iss}`);
  }

  // Validate audience
  const audiences = Array.isArray(payload.aud) ? payload.aud : [payload.aud];
  if (!audiences.includes(expectedAudience)) {
    throw new Error(`Invalid audience: expected ${expectedAudience}, got ${audiences.join(', ')}`);
  }

  // Validate expiration
  const now = Math.floor(Date.now() / 1000);
  if (payload.exp && payload.exp < now) {
    throw new Error('Token has expired');
  }

  // Validate not before
  if (payload.nbf && payload.nbf > now) {
    throw new Error('Token not yet valid');
  }

  // Validate nonce if provided
  if (expectedNonce && payload.nonce !== expectedNonce) {
    throw new Error('Invalid nonce');
  }

  return payload;
}

// ============================================================================
// Session Management
// ============================================================================

/**
 * Save OIDC session to storage
 */
export function saveSession(session: OIDCSession): void {
  if (typeof window !== 'undefined') {
    localStorage.setItem(OIDC_DEFAULTS.SESSION_STORAGE_KEY, JSON.stringify(session));
  }
}

/**
 * Load OIDC session from storage
 */
export function loadSession(): OIDCSession | null {
  if (typeof window === 'undefined') return null;

  const stored = localStorage.getItem(OIDC_DEFAULTS.SESSION_STORAGE_KEY);
  if (!stored) return null;

  try {
    const session = JSON.parse(stored) as OIDCSession;

    // Check if session is expired
    if (session.expiresAt && session.expiresAt < Date.now()) {
      clearSession();
      return null;
    }

    return session;
  } catch {
    clearSession();
    return null;
  }
}

/**
 * Clear OIDC session from storage
 */
export function clearSession(): void {
  if (typeof window !== 'undefined') {
    localStorage.removeItem(OIDC_DEFAULTS.SESSION_STORAGE_KEY);
    sessionStorage.removeItem(OIDC_DEFAULTS.STATE_STORAGE_KEY);
    sessionStorage.removeItem(OIDC_DEFAULTS.NONCE_STORAGE_KEY);
    sessionStorage.removeItem(OIDC_DEFAULTS.PKCE_STORAGE_KEY);
  }
}

/**
 * Check if current session is valid and not expired
 */
export function isSessionValid(session?: OIDCSession): boolean {
  const currentSession = session || loadSession();
  if (!currentSession) return false;

  return currentSession.expiresAt > Date.now();
}

/**
 * Get stored state parameter for validation
 */
export function getStoredState(): string | null {
  if (typeof window === 'undefined') return null;

  const state = sessionStorage.getItem(OIDC_DEFAULTS.STATE_STORAGE_KEY);
  sessionStorage.removeItem(OIDC_DEFAULTS.STATE_STORAGE_KEY);
  return state;
}

/**
 * Get stored nonce parameter for validation
 */
export function getStoredNonce(): string | null {
  if (typeof window === 'undefined') return null;

  const nonce = sessionStorage.getItem(OIDC_DEFAULTS.NONCE_STORAGE_KEY);
  sessionStorage.removeItem(OIDC_DEFAULTS.NONCE_STORAGE_KEY);
  return nonce;
}

// ============================================================================
// High-Level OIDC Client
// ============================================================================

/**
 * High-level OIDC Client class
 */
export class OIDCClient {
  private config: OIDCClientConfig;
  private providerConfig?: OIDCProviderConfig;

  constructor(config: OIDCClientConfig) {
    this.config = config;
  }

  /**
   * Initialize the client by fetching provider configuration
   */
  async initialize(): Promise<void> {
    this.providerConfig = await fetchDiscoveryDocument(this.config.providerUrl);
  }

  /**
   * Start authorization flow
   */
  async authorize(): Promise<string> {
    const { url } = await buildAuthorizationUrl(this.config, this.providerConfig);
    return url;
  }

  /**
   * Handle authorization callback
   */
  async handleCallback(callbackUrl: string): Promise<OIDCSession> {
    const url = new URL(callbackUrl);
    const code = url.searchParams.get('code');
    const state = url.searchParams.get('state');
    const error = url.searchParams.get('error');

    // Check for authorization errors
    if (error) {
      const errorDescription = url.searchParams.get('error_description');
      throw new Error(
        `Authorization error: ${error}${errorDescription ? ` - ${errorDescription}` : ''}`
      );
    }

    // Validate required parameters
    if (!code) {
      throw new Error('Missing authorization code');
    }

    // Validate state parameter
    const storedState = getStoredState();
    if (!state || state !== storedState) {
      throw new Error('Invalid state parameter');
    }

    // Exchange code for tokens
    const tokens = await exchangeCodeForTokens(code, this.config, this.providerConfig);

    // Validate ID token if present
    let userInfo: UserInfo | undefined;
    if (tokens.id_token) {
      const storedNonce = getStoredNonce();
      const payload = validateJWTClaims(
        tokens.id_token,
        this.providerConfig?.issuer ?? '',
        this.config.clientId,
        storedNonce || undefined
      );

      // Extract user info from ID token
      userInfo = {
        sub: payload.sub,
        name: payload.name,
        preferred_username: payload.preferred_username,
        email: payload.email,
        email_verified: payload.email_verified,
        did: payload.did,
        vault_id: payload.vault_id,
      };
    } else {
      // Fetch user info from UserInfo endpoint
      userInfo = await fetchUserInfo(tokens.access_token, this.config, this.providerConfig);
    }

    // Create session
    const session: OIDCSession = {
      accessToken: tokens.access_token,
      refreshToken: tokens.refresh_token,
      idToken: tokens.id_token,
      expiresAt: Date.now() + tokens.expires_in * 1000,
      userInfo,
      tokenType: tokens.token_type,
      scope: tokens.scope,
    };

    // Save session
    saveSession(session);

    return session;
  }

  /**
   * Refresh current session
   */
  async refreshSession(): Promise<OIDCSession | null> {
    const currentSession = loadSession();
    if (!currentSession?.refreshToken) {
      return null;
    }

    try {
      const tokens = await refreshAccessToken(
        currentSession.refreshToken,
        this.config,
        this.providerConfig
      );

      const newSession: OIDCSession = {
        ...currentSession,
        accessToken: tokens.access_token,
        refreshToken: tokens.refresh_token || currentSession.refreshToken,
        expiresAt: Date.now() + tokens.expires_in * 1000,
      };

      saveSession(newSession);
      return newSession;
    } catch (error) {
      console.error('Failed to refresh session:', error);
      clearSession();
      return null;
    }
  }

  /**
   * Logout and clear session
   */
  logout(): void {
    clearSession();
  }

  /**
   * Get current session
   */
  getSession(): OIDCSession | null {
    return loadSession();
  }

  /**
   * Check if user is authenticated
   */
  isAuthenticated(): boolean {
    return isSessionValid();
  }
}

// ============================================================================
// Error Handling Utilities
// ============================================================================

/**
 * Check if error is an OIDC error response
 */
export function isOIDCError(error: unknown): error is OIDCError {
  return (
    typeof error === 'object' &&
    error !== null &&
    'error' in error &&
    typeof (error as Record<string, unknown>).error === 'string'
  );
}

/**
 * Format OIDC error for display
 */
export function formatOIDCError(error: OIDCError): string {
  return `${error.error}${error.error_description ? `: ${error.error_description}` : ''}`;
}
