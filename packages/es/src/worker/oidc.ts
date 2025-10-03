/**
 * OpenID Connect (OIDC) client for Motor Authorization Beacon
 * 
 * @packageDocumentation
 */

import type {
  OIDCConfiguration,
  OIDCAuthorizationRequest,
  OIDCAuthorizationResponse,
  OIDCTokenRequest,
  OIDCTokenResponse,
  OIDCUserInfo,
  JWKS,
  MotorServiceWorkerConfig,
} from './types';

/**
 * OIDC client configuration
 */
export interface OIDCClientConfig extends MotorServiceWorkerConfig {
  /** Client ID for OIDC */
  client_id: string;
  /** Redirect URI for authorization callback */
  redirect_uri: string;
  /** Client secret (for confidential clients) */
  client_secret?: string;
  /** Requested scopes */
  scope?: string;
  /** Response type */
  response_type?: string;
  /** Use PKCE for authorization code flow */
  use_pkce?: boolean;
  /** Auto-refresh tokens */
  auto_refresh?: boolean;
}

/**
 * OIDC Authorization client for Motor
 */
export class OIDCClient {
  private readonly baseUrl: string;
  private readonly config: OIDCClientConfig;
  private discoveryCache?: OIDCConfiguration;
  private accessToken?: string;
  private refreshToken?: string;
  private idToken?: string;
  private tokenExpiresAt?: number;
  private codeVerifier?: string;
  private codeChallenge?: string;

  constructor(config: OIDCClientConfig) {
    this.config = {
      worker_url: '/api',
      timeout: 30000,
      max_retries: 3,
      debug: false,
      scope: 'openid profile email',
      response_type: 'code',
      use_pkce: true,
      auto_refresh: true,
      ...config,
    };
    this.baseUrl = this.config.worker_url || '/api';
  }

  /**
   * Make an API request
   */
  private async request<T>(
    method: string,
    endpoint: string,
    data?: unknown,
    headers?: Record<string, string>
  ): Promise<T> {
    const url = `${this.baseUrl}${endpoint}`;
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), this.config.timeout || 30000);

    try {
      const response = await fetch(url, {
        method,
        headers: {
          'Content-Type': 'application/json',
          ...headers,
        },
        body: data ? JSON.stringify(data) : undefined,
        signal: controller.signal,
      });

      clearTimeout(timeoutId);

      if (!response.ok) {
        const error = await response.text();
        throw new Error(`OIDC API error: ${response.status} - ${error}`);
      }

      return await response.json();
    } catch (error) {
      clearTimeout(timeoutId);
      if (error instanceof Error && error.name === 'AbortError') {
        throw new Error('OIDC request timed out');
      }
      throw error;
    }
  }

  /**
   * Get OIDC discovery configuration
   */
  async getConfiguration(): Promise<OIDCConfiguration> {
    if (!this.discoveryCache) {
      this.discoveryCache = await this.request<OIDCConfiguration>(
        'GET',
        '/.well-known/openid-configuration'
      );
    }
    return this.discoveryCache;
  }

  /**
   * Generate PKCE code verifier and challenge
   */
  private generatePKCE(): { verifier: string; challenge: string } {
    const verifier = this.generateRandomString(128);
    const challenge = this.base64UrlEncode(
      crypto.subtle.digest('SHA-256', new TextEncoder().encode(verifier))
    );
    
    return { verifier, challenge };
  }

  /**
   * Generate random string for state/nonce/PKCE
   */
  private generateRandomString(length: number): string {
    const array = new Uint8Array(length);
    crypto.getRandomValues(array);
    return this.base64UrlEncode(array);
  }

  /**
   * Base64 URL encode
   */
  private base64UrlEncode(data: ArrayBuffer | Uint8Array | Promise<ArrayBuffer>): string {
    if (data instanceof Promise) {
      throw new Error('Cannot encode Promise directly');
    }
    
    const bytes = data instanceof ArrayBuffer ? new Uint8Array(data) : data;
    let binary = '';
    bytes.forEach(byte => binary += String.fromCharCode(byte));
    
    return btoa(binary)
      .replace(/\+/g, '-')
      .replace(/\//g, '_')
      .replace(/=/g, '');
  }

  /**
   * Build authorization URL
   */
  async buildAuthorizationUrl(options: Partial<OIDCAuthorizationRequest> = {}): Promise<string> {
    const config = await this.getConfiguration();
    const state = this.generateRandomString(32);
    const nonce = this.generateRandomString(32);

    const params: OIDCAuthorizationRequest = {
      client_id: this.config.client_id,
      redirect_uri: this.config.redirect_uri,
      response_type: this.config.response_type || 'code',
      scope: this.config.scope || 'openid profile email',
      state,
      nonce,
      ...options,
    };

    // Add PKCE if enabled
    if (this.config.use_pkce) {
      const pkce = this.generatePKCE();
      this.codeVerifier = pkce.verifier;
      this.codeChallenge = pkce.challenge;
      params.code_challenge = pkce.challenge;
      params.code_challenge_method = 'S256';
    }

    // Store state and nonce for validation
    sessionStorage.setItem('oidc_state', state);
    sessionStorage.setItem('oidc_nonce', nonce);

    const url = new URL(config.authorization_endpoint);
    Object.entries(params).forEach(([key, value]) => {
      if (value !== undefined) {
        url.searchParams.append(key, String(value));
      }
    });

    return url.toString();
  }

  /**
   * Start authorization flow
   */
  async authorize(options: Partial<OIDCAuthorizationRequest> = {}): Promise<void> {
    const authUrl = await this.buildAuthorizationUrl(options);
    window.location.href = authUrl;
  }

  /**
   * Open authorization in popup
   */
  async authorizePopup(options: Partial<OIDCAuthorizationRequest> = {}): Promise<OIDCAuthorizationResponse> {
    const authUrl = await this.buildAuthorizationUrl(options);
    
    const popup = window.open(
      authUrl,
      'oidc_auth',
      'width=500,height=600,menubar=no,toolbar=no'
    );

    if (!popup) {
      throw new Error('Failed to open authorization popup');
    }

    return new Promise((resolve, reject) => {
      const checkInterval = setInterval(() => {
        try {
          if (popup.closed) {
            clearInterval(checkInterval);
            reject(new Error('Authorization popup was closed'));
            return;
          }

          // Check if redirected back
          if (popup.location.href.startsWith(this.config.redirect_uri)) {
            clearInterval(checkInterval);
            
            const url = new URL(popup.location.href);
            const response = this.parseAuthorizationResponse(url);
            
            popup.close();
            resolve(response);
          }
        } catch (e) {
          // Cross-origin, ignore
        }
      }, 500);
    });
  }

  /**
   * Parse authorization response from URL
   */
  private parseAuthorizationResponse(url: URL): OIDCAuthorizationResponse {
    const params = url.searchParams;
    const hashParams = new URLSearchParams(url.hash.slice(1));

    // Check for error
    const error = params.get('error') || hashParams.get('error');
    if (error) {
      throw new Error(`Authorization error: ${error} - ${params.get('error_description') || ''}`);
    }

    // Validate state
    const state = params.get('state') || hashParams.get('state');
    const savedState = sessionStorage.getItem('oidc_state');
    if (state !== savedState) {
      throw new Error('State mismatch - possible CSRF attack');
    }

    return {
      code: params.get('code') || undefined,
      access_token: hashParams.get('access_token') || undefined,
      token_type: hashParams.get('token_type') || undefined,
      id_token: hashParams.get('id_token') || undefined,
      state: state || undefined,
      expires_in: hashParams.get('expires_in') ? parseInt(hashParams.get('expires_in')!) : undefined,
      scope: hashParams.get('scope') || undefined,
    };
  }

  /**
   * Handle authorization callback
   */
  async handleCallback(url?: string): Promise<OIDCTokenResponse> {
    const callbackUrl = new URL(url || window.location.href);
    const response = this.parseAuthorizationResponse(callbackUrl);

    if (response.code) {
      // Exchange code for tokens
      return await this.exchangeCode(response.code);
    } else if (response.access_token) {
      // Implicit flow
      this.accessToken = response.access_token;
      this.idToken = response.id_token;
      if (response.expires_in) {
        this.tokenExpiresAt = Date.now() + response.expires_in * 1000;
      }
      
      return {
        access_token: response.access_token,
        token_type: response.token_type || 'Bearer',
        expires_in: response.expires_in || 3600,
        id_token: response.id_token,
        scope: response.scope,
      };
    }

    throw new Error('No authorization code or access token in response');
  }

  /**
   * Exchange authorization code for tokens
   */
  async exchangeCode(code: string): Promise<OIDCTokenResponse> {
    const request: OIDCTokenRequest = {
      grant_type: 'authorization_code',
      code,
      redirect_uri: this.config.redirect_uri,
      client_id: this.config.client_id,
      client_secret: this.config.client_secret,
    };

    // Add PKCE verifier if used
    if (this.codeVerifier) {
      request.code_verifier = this.codeVerifier;
    }

    const response = await this.request<OIDCTokenResponse>('POST', '/token', request);

    // Store tokens
    this.accessToken = response.access_token;
    this.refreshToken = response.refresh_token;
    this.idToken = response.id_token;
    if (response.expires_in) {
      this.tokenExpiresAt = Date.now() + response.expires_in * 1000;
    }

    // Clear PKCE values
    this.codeVerifier = undefined;
    this.codeChallenge = undefined;

    return response;
  }

  /**
   * Refresh access token
   */
  async refreshAccessToken(): Promise<OIDCTokenResponse> {
    if (!this.refreshToken) {
      throw new Error('No refresh token available');
    }

    const request: OIDCTokenRequest = {
      grant_type: 'refresh_token',
      refresh_token: this.refreshToken,
      client_id: this.config.client_id,
      client_secret: this.config.client_secret,
      scope: this.config.scope,
    };

    const response = await this.request<OIDCTokenResponse>('POST', '/token', request);

    // Update tokens
    this.accessToken = response.access_token;
    if (response.refresh_token) {
      this.refreshToken = response.refresh_token;
    }
    if (response.expires_in) {
      this.tokenExpiresAt = Date.now() + response.expires_in * 1000;
    }

    return response;
  }

  /**
   * Get user information
   */
  async getUserInfo(): Promise<OIDCUserInfo> {
    if (!this.accessToken) {
      throw new Error('No access token available');
    }

    // Auto-refresh if expired
    if (this.config.auto_refresh && this.isTokenExpired()) {
      await this.refreshAccessToken();
    }

    return await this.request<OIDCUserInfo>(
      'GET',
      '/userinfo',
      undefined,
      {
        'Authorization': `Bearer ${this.accessToken}`,
      }
    );
  }

  /**
   * Get JSON Web Key Set
   */
  async getJWKS(): Promise<JWKS> {
    const config = await this.getConfiguration();
    const jwksUrl = new URL(config.jwks_uri);
    
    // If JWKS is on the same origin, use our request method
    if (jwksUrl.origin === window.location.origin) {
      return await this.request<JWKS>('GET', jwksUrl.pathname);
    }
    
    // Otherwise fetch directly
    const response = await fetch(config.jwks_uri);
    if (!response.ok) {
      throw new Error(`Failed to fetch JWKS: ${response.status}`);
    }
    
    return await response.json();
  }

  /**
   * Logout / revoke tokens
   */
  async logout(): Promise<void> {
    // Clear local tokens
    this.accessToken = undefined;
    this.refreshToken = undefined;
    this.idToken = undefined;
    this.tokenExpiresAt = undefined;
    
    // Clear session storage
    sessionStorage.removeItem('oidc_state');
    sessionStorage.removeItem('oidc_nonce');
    
    // TODO: Call revocation endpoint if available
  }

  /**
   * Check if token is expired
   */
  isTokenExpired(): boolean {
    if (!this.tokenExpiresAt) {
      return true;
    }
    // Consider expired 60 seconds before actual expiry
    return Date.now() >= this.tokenExpiresAt - 60000;
  }

  /**
   * Get current access token
   */
  getAccessToken(): string | undefined {
    return this.accessToken;
  }

  /**
   * Get current ID token
   */
  getIdToken(): string | undefined {
    return this.idToken;
  }

  /**
   * Get current refresh token
   */
  getRefreshToken(): string | undefined {
    return this.refreshToken;
  }

  /**
   * Set tokens manually (for testing or restoration)
   */
  setTokens(tokens: {
    access_token?: string;
    refresh_token?: string;
    id_token?: string;
    expires_in?: number;
  }): void {
    this.accessToken = tokens.access_token;
    this.refreshToken = tokens.refresh_token;
    this.idToken = tokens.id_token;
    if (tokens.expires_in) {
      this.tokenExpiresAt = Date.now() + tokens.expires_in * 1000;
    }
  }
}

/**
 * Create an OIDC client
 */
export function createOIDCClient(config: OIDCClientConfig): OIDCClient {
  return new OIDCClient(config);
}

/**
 * Check if we're on a callback URL
 */
export function isOIDCCallback(redirectUri: string): boolean {
  const currentUrl = window.location.href;
  return currentUrl.startsWith(redirectUri) && 
         (currentUrl.includes('code=') || currentUrl.includes('access_token=') || currentUrl.includes('error='));
}

/**
 * Auto-handle OIDC callback if on callback URL
 */
export async function autoHandleOIDCCallback(
  client: OIDCClient,
  redirectUri: string,
  onSuccess?: (tokens: OIDCTokenResponse) => void,
  onError?: (error: Error) => void
): Promise<void> {
  if (!isOIDCCallback(redirectUri)) {
    return;
  }

  try {
    const tokens = await client.handleCallback();
    if (onSuccess) {
      onSuccess(tokens);
    }
    
    // Clean up URL
    window.history.replaceState({}, document.title, redirectUri);
  } catch (error) {
    if (onError) {
      onError(error instanceof Error ? error : new Error(String(error)));
    }
  }
}