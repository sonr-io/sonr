/**
 * OAuth 2.0 client utilities for Sonr authentication
 */

/**
 * OAuth2 configuration
 */
export interface OAuth2Config {
  clientId: string;
  redirectUri: string;
  authorizationUrl?: string;
  tokenUrl?: string;
  userInfoUrl?: string;
  scopes?: string[];
  responseType?: string;
  grantType?: string;
  pkce?: boolean;
}

/**
 * OAuth2 token response
 */
export interface OAuth2Token {
  access_token: string;
  token_type: string;
  expires_in?: number;
  refresh_token?: string;
  scope?: string;
  id_token?: string;
  ucan_token?: string;
}

/**
 * OAuth2 error response
 */
export interface OAuth2Error {
  error: string;
  error_description?: string;
  error_uri?: string;
  state?: string;
}

/**
 * User info from OAuth2 provider
 */
export interface OAuth2UserInfo {
  sub: string;
  name?: string;
  email?: string;
  email_verified?: boolean;
  picture?: string;
  did?: string;
  vault_address?: string;
  capabilities?: string[];
}

/**
 * OAuth2 client class
 */
export class OAuth2Client {
  private config: OAuth2Config;
  private codeVerifier?: string;

  constructor(config: OAuth2Config) {
    this.config = {
      authorizationUrl: '/oauth2/authorize',
      tokenUrl: '/oauth2/token',
      userInfoUrl: '/oauth2/userinfo',
      responseType: 'code',
      grantType: 'authorization_code',
      pkce: true,
      scopes: ['openid', 'profile'],
      ...config,
    };
  }

  /**
   * Generate authorization URL
   */
  async getAuthorizationUrl(state?: string): Promise<string> {
    const params = new URLSearchParams({
      response_type: this.config.responseType || 'code',
      client_id: this.config.clientId,
      redirect_uri: this.config.redirectUri,
      scope: (this.config.scopes || []).join(' '),
      state: state || this.generateState(),
    });

    // Add PKCE challenge if enabled
    if (this.config.pkce) {
      const { verifier, challenge } = await this.generatePKCE();
      this.codeVerifier = verifier;

      // Store verifier in session storage for later use
      sessionStorage.setItem('sonr_oauth_code_verifier', verifier);

      params.append('code_challenge', challenge);
      params.append('code_challenge_method', 'S256');
    }

    return `${this.config.authorizationUrl}?${params.toString()}`;
  }

  /**
   * Exchange authorization code for tokens
   */
  async exchangeCode(code: string, state?: string): Promise<OAuth2Token> {
    const params = new URLSearchParams({
      grant_type: this.config.grantType || 'authorization_code',
      code,
      client_id: this.config.clientId,
      redirect_uri: this.config.redirectUri,
    });

    // Add PKCE verifier if available
    if (this.config.pkce) {
      const verifier = sessionStorage.getItem('sonr_oauth_code_verifier');
      if (verifier) {
        params.append('code_verifier', verifier);
        sessionStorage.removeItem('sonr_oauth_code_verifier');
      }
    }

    // Add state if provided
    if (state) {
      params.append('state', state);
    }

    const response = await fetch(this.config.tokenUrl!, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/x-www-form-urlencoded',
      },
      body: params.toString(),
    });

    if (!response.ok) {
      const error = await response.json();
      throw new OAuth2ClientError(error.error, error.error_description);
    }

    const token = await response.json();

    // Store tokens
    this.storeTokens(token);

    return token;
  }

  /**
   * Refresh access token
   */
  async refreshToken(refreshToken?: string): Promise<OAuth2Token> {
    const storedRefreshToken = refreshToken || this.getStoredToken()?.refresh_token;

    if (!storedRefreshToken) {
      throw new OAuth2ClientError('invalid_grant', 'No refresh token available');
    }

    const params = new URLSearchParams({
      grant_type: 'refresh_token',
      refresh_token: storedRefreshToken,
      client_id: this.config.clientId,
    });

    const response = await fetch(this.config.tokenUrl!, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/x-www-form-urlencoded',
      },
      body: params.toString(),
    });

    if (!response.ok) {
      const error = await response.json();
      throw new OAuth2ClientError(error.error, error.error_description);
    }

    const token = await response.json();

    // Store new tokens
    this.storeTokens(token);

    return token;
  }

  /**
   * Get user info
   */
  async getUserInfo(accessToken?: string): Promise<OAuth2UserInfo> {
    const token = accessToken || this.getStoredToken()?.access_token;

    if (!token) {
      throw new OAuth2ClientError('invalid_request', 'No access token available');
    }

    const response = await fetch(this.config.userInfoUrl!, {
      headers: {
        Authorization: `Bearer ${token}`,
      },
    });

    if (!response.ok) {
      const error = await response.json();
      throw new OAuth2ClientError(error.error, error.error_description);
    }

    return response.json();
  }

  /**
   * Revoke token
   */
  async revokeToken(
    token: string,
    tokenType: 'access_token' | 'refresh_token' = 'access_token'
  ): Promise<void> {
    const params = new URLSearchParams({
      token,
      token_type_hint: tokenType,
      client_id: this.config.clientId,
    });

    const response = await fetch('/oauth2/revoke', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/x-www-form-urlencoded',
      },
      body: params.toString(),
    });

    if (!response.ok && response.status !== 200) {
      const error = await response.json();
      throw new OAuth2ClientError(error.error, error.error_description);
    }

    // Clear stored tokens if revoking refresh token
    if (tokenType === 'refresh_token') {
      this.clearStoredTokens();
    }
  }

  /**
   * Logout
   */
  async logout(): Promise<void> {
    const token = this.getStoredToken();

    if (token?.access_token) {
      try {
        await this.revokeToken(token.access_token, 'access_token');
      } catch (error) {
        console.error('Failed to revoke access token:', error);
      }
    }

    if (token?.refresh_token) {
      try {
        await this.revokeToken(token.refresh_token, 'refresh_token');
      } catch (error) {
        console.error('Failed to revoke refresh token:', error);
      }
    }

    this.clearStoredTokens();
  }

  /**
   * Check if user is authenticated
   */
  isAuthenticated(): boolean {
    const token = this.getStoredToken();
    if (!token?.access_token) {
      return false;
    }

    // Check if token is expired
    const expiresAt = localStorage.getItem('sonr_oauth_expires_at');
    if (expiresAt && Date.now() > Number.parseInt(expiresAt)) {
      return false;
    }

    return true;
  }

  /**
   * Get stored access token
   */
  getAccessToken(): string | null {
    return this.getStoredToken()?.access_token || null;
  }

  /**
   * Get stored refresh token
   */
  getRefreshToken(): string | null {
    return this.getStoredToken()?.refresh_token || null;
  }

  /**
   * Get current code verifier (for testing/debugging)
   */
  getCodeVerifier(): string | undefined {
    return this.codeVerifier;
  }

  /**
   * Private methods
   */

  private generateState(): string {
    const array = new Uint8Array(32);
    crypto.getRandomValues(array);
    return this.base64UrlEncode(array);
  }

  private async generatePKCE(): Promise<{ verifier: string; challenge: string }> {
    const verifier = this.generateCodeVerifier();
    const challenge = await this.generateCodeChallenge(verifier);
    return { verifier, challenge };
  }

  private generateCodeVerifier(): string {
    const array = new Uint8Array(32);
    crypto.getRandomValues(array);
    return this.base64UrlEncode(array);
  }

  private async generateCodeChallenge(verifier: string): Promise<string> {
    const encoder = new TextEncoder();
    const data = encoder.encode(verifier);
    const digest = await crypto.subtle.digest('SHA-256', data);
    return this.base64UrlEncode(new Uint8Array(digest));
  }

  private base64UrlEncode(array: Uint8Array): string {
    return btoa(String.fromCharCode(...array))
      .replace(/\+/g, '-')
      .replace(/\//g, '_')
      .replace(/=/g, '');
  }

  private storeTokens(token: OAuth2Token): void {
    localStorage.setItem('sonr_oauth_token', JSON.stringify(token));

    // Calculate and store expiration time
    if (token.expires_in) {
      const expiresAt = Date.now() + token.expires_in * 1000;
      localStorage.setItem('sonr_oauth_expires_at', expiresAt.toString());
    }
  }

  private getStoredToken(): OAuth2Token | null {
    const tokenString = localStorage.getItem('sonr_oauth_token');
    if (!tokenString) {
      return null;
    }

    try {
      return JSON.parse(tokenString);
    } catch {
      return null;
    }
  }

  private clearStoredTokens(): void {
    localStorage.removeItem('sonr_oauth_token');
    localStorage.removeItem('sonr_oauth_expires_at');
    sessionStorage.removeItem('sonr_oauth_code_verifier');
  }
}

/**
 * OAuth2 client error
 */
export class OAuth2ClientError extends Error {
  constructor(
    public code: string,
    message?: string,
    public uri?: string
  ) {
    super(message || code);
    this.name = 'OAuth2ClientError';
  }
}

/**
 * Parse OAuth2 callback URL
 */
export function parseCallbackUrl(url: string): {
  code?: string;
  state?: string;
  error?: string;
  error_description?: string;
} {
  const urlObj = new URL(url);
  const params = new URLSearchParams(urlObj.search);

  return {
    code: params.get('code') || undefined,
    state: params.get('state') || undefined,
    error: params.get('error') || undefined,
    error_description: params.get('error_description') || undefined,
  };
}

/**
 * Check if token is expired
 */
export function isTokenExpired(_token: OAuth2Token): boolean {
  const expiresAt = localStorage.getItem('sonr_oauth_expires_at');
  if (!expiresAt) {
    return false;
  }

  return Date.now() > Number.parseInt(expiresAt);
}

/**
 * Calculate token expiration time
 */
export function calculateTokenExpiry(expiresIn: number): Date {
  return new Date(Date.now() + expiresIn * 1000);
}
