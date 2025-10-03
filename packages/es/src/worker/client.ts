/**
 * HTTP client for Motor Payment Gateway and OIDC Authorization
 * 
 * @packageDocumentation
 */

import type {
  MotorServiceWorkerConfig,
  HealthCheckResponse,
  ServiceInfoResponse,
  ErrorResponse,
  // Payment types
  PaymentInstrument,
  CanMakePaymentRequest,
  CanMakePaymentResponse,
  PaymentRequestEvent,
  PaymentHandlerResponse,
  ProcessPaymentRequest,
  ProcessPaymentResponse,
  ValidatePaymentMethodRequest,
  ValidatePaymentMethodResponse,
  PaymentStatus,
  RefundPaymentRequest,
  RefundPaymentResponse,
  // OIDC types
  OIDCConfiguration,
  OIDCAuthorizationRequest,
  OIDCAuthorizationResponse,
  OIDCTokenRequest,
  OIDCTokenResponse,
  OIDCUserInfo,
  JWKS,
} from './types';

/**
 * Default configuration for the Motor client
 */
const DEFAULT_CONFIG: Required<MotorServiceWorkerConfig> = {
  worker_url: '/motor-worker',
  timeout: 30000,
  max_retries: 3,
  debug: false,
};

/**
 * HTTP client for Motor Payment Gateway and OIDC services
 */
export class MotorClient {
  private readonly config: Required<MotorServiceWorkerConfig>;
  private readonly baseUrl: string;
  private serviceWorkerReady: Promise<boolean> | null = null;
  private useServiceWorker = false;

  constructor(config: Partial<MotorServiceWorkerConfig> = {}) {
    this.config = { ...DEFAULT_CONFIG, ...config };
    this.baseUrl = this.config.worker_url;
    
    this.detectServiceWorker();
    
    if (this.config.debug) {
      console.debug('[MotorClient] Initialized with config:', this.config);
    }
  }

  /**
   * Detects if a Motor service worker is available
   */
  private detectServiceWorker(): void {
    if (typeof navigator !== 'undefined' && 'serviceWorker' in navigator) {
      this.serviceWorkerReady = navigator.serviceWorker.ready.then(registration => {
        if (registration.active?.scriptURL.includes('motor')) {
          this.useServiceWorker = true;
          if (this.config.debug) {
            console.debug('[MotorClient] Motor service worker detected');
          }
          return true;
        }
        return false;
      }).catch(() => false);
    }
  }

  // Core HTTP Methods

  /**
   * Makes an HTTP request with automatic retries
   */
  private async request<T>(
    method: string,
    endpoint: string,
    data?: unknown
  ): Promise<T> {
    const url = `${this.baseUrl}${endpoint}`;
    let lastError: Error | null = null;

    for (let attempt = 1; attempt <= this.config.max_retries; attempt++) {
      try {
        if (this.config.debug) {
          console.debug(`[MotorClient] ${method} ${url} (attempt ${attempt})`);
        }

        const response = await this.performRequest(method, url, data);
        
        if (!response.ok) {
          const errorData = await this.parseErrorResponse(response);
          throw new Error(`HTTP ${response.status}: ${errorData.error || response.statusText}`);
        }

        return await this.parseResponse<T>(response);
      } catch (error) {
        lastError = error instanceof Error ? error : new Error(String(error));
        
        if (this.config.debug) {
          console.debug(`[MotorClient] Attempt ${attempt} failed:`, lastError.message);
        }

        if (lastError.message.includes('HTTP 4')) {
          break;
        }

        if (attempt < this.config.max_retries) {
          const delay = Math.min(1000 * Math.pow(2, attempt - 1), 5000);
          await this.sleep(delay);
        }
      }
    }

    throw lastError || new Error('Request failed after all retries');
  }

  /**
   * Performs the actual HTTP request
   */
  private async performRequest(
    method: string,
    url: string,
    data?: unknown
  ): Promise<Response> {
    if (this.useServiceWorker && this.serviceWorkerReady) {
      const swReady = await this.serviceWorkerReady;
      if (swReady) {
        return this.performServiceWorkerRequest(method, url, data);
      }
    }

    const headers: HeadersInit = {
      'Content-Type': 'application/json',
    };

    const requestInit: RequestInit = {
      method,
      headers,
      mode: 'cors',
    };

    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), this.config.timeout);
    requestInit.signal = controller.signal;

    try {
      if (data && method !== 'GET') {
        requestInit.body = JSON.stringify(this.serializeData(data));
      }

      return await fetch(url, requestInit);
    } finally {
      clearTimeout(timeoutId);
    }
  }

  /**
   * Performs a request through the service worker
   */
  private async performServiceWorkerRequest(
    method: string,
    url: string,
    data?: unknown
  ): Promise<Response> {
    return new Promise((resolve, reject) => {
      if (!navigator.serviceWorker.controller) {
        return this.performRequest(method, url, data).then(resolve).catch(reject);
      }

      const messageChannel = new MessageChannel();
      const timeout = setTimeout(() => {
        reject(new Error('Service worker request timed out'));
      }, this.config.timeout);

      messageChannel.port1.onmessage = (event) => {
        clearTimeout(timeout);
        
        if (event.data.error) {
          reject(new Error(event.data.error));
        } else {
          const response = new Response(
            JSON.stringify(event.data),
            {
              status: event.data.error ? 500 : 200,
              headers: { 'Content-Type': 'application/json' },
            }
          );
          resolve(response);
        }
      };

      navigator.serviceWorker.controller.postMessage(
        {
          type: 'API_REQUEST',
          method,
          url,
          data: data ? this.serializeData(data) : undefined,
        },
        [messageChannel.port2]
      );
    });
  }

  /**
   * Parses a successful response
   */
  private async parseResponse<T>(response: Response): Promise<T> {
    const text = await response.text();
    
    if (!text) {
      return {} as T;
    }

    try {
      const parsed = JSON.parse(text);
      return this.deserializeData<T>(parsed);
    } catch (error) {
      throw new Error(`Failed to parse response JSON: ${error}`);
    }
  }

  /**
   * Parses an error response
   */
  private async parseErrorResponse(response: Response): Promise<ErrorResponse> {
    try {
      const text = await response.text();
      if (text) {
        return JSON.parse(text) as ErrorResponse;
      }
    } catch {
      // Ignore parsing errors
    }
    
    return { error: response.statusText || 'Unknown error' };
  }

  /**
   * Serializes data for sending
   */
  private serializeData(data: unknown): unknown {
    if (data instanceof Uint8Array) {
      return Array.from(data);
    }
    
    if (data && typeof data === 'object') {
      const result: Record<string, unknown> = {};
      for (const [key, value] of Object.entries(data)) {
        result[key] = this.serializeData(value);
      }
      return result;
    }
    
    return data;
  }

  /**
   * Deserializes received data
   */
  private deserializeData<T>(data: unknown): T {
    if (Array.isArray(data) && data.every(item => typeof item === 'number')) {
      return new Uint8Array(data) as unknown as T;
    }
    
    if (data && typeof data === 'object') {
      const result: Record<string, unknown> = {};
      for (const [key, value] of Object.entries(data)) {
        result[key] = this.deserializeData(value);
      }
      return result as T;
    }
    
    return data as T;
  }

  /**
   * Sleep utility
   */
  private sleep(ms: number): Promise<void> {
    return new Promise(resolve => setTimeout(resolve, ms));
  }

  // Health & Status

  /**
   * Check service health
   */
  async healthCheck(): Promise<HealthCheckResponse> {
    return this.request<HealthCheckResponse>('GET', '/health');
  }

  /**
   * Get service information
   */
  async getServiceInfo(): Promise<ServiceInfoResponse> {
    return this.request<ServiceInfoResponse>('GET', '/status');
  }

  // Payment Gateway API

  /**
   * Get available payment instruments
   */
  async getPaymentInstruments(): Promise<PaymentInstrument[]> {
    return this.request<PaymentInstrument[]>('GET', '/payment/instruments');
  }

  /**
   * Check if payment can be made
   */
  async canMakePayment(request: CanMakePaymentRequest): Promise<CanMakePaymentResponse> {
    return this.request<CanMakePaymentResponse>('POST', '/payment/canmakepayment', request);
  }

  /**
   * Handle payment request event
   */
  async handlePaymentRequest(event: PaymentRequestEvent): Promise<PaymentHandlerResponse> {
    return this.request<PaymentHandlerResponse>('POST', '/payment/paymentrequest', event);
  }

  /**
   * Process a payment
   */
  async processPayment(request: ProcessPaymentRequest): Promise<ProcessPaymentResponse> {
    return this.request<ProcessPaymentResponse>('POST', '/api/payment/process', request);
  }

  /**
   * Validate payment method
   */
  async validatePaymentMethod(request: ValidatePaymentMethodRequest): Promise<ValidatePaymentMethodResponse> {
    return this.request<ValidatePaymentMethodResponse>('POST', '/api/payment/validate', request);
  }

  /**
   * Get payment status
   */
  async getPaymentStatus(paymentId: string): Promise<PaymentStatus> {
    return this.request<PaymentStatus>('GET', `/api/payment/status/${paymentId}`);
  }

  /**
   * Process refund
   */
  async refundPayment(request: RefundPaymentRequest): Promise<RefundPaymentResponse> {
    return this.request<RefundPaymentResponse>('POST', '/api/payment/refund', request);
  }

  // OIDC API

  /**
   * Get OIDC configuration
   */
  async getOIDCConfiguration(): Promise<OIDCConfiguration> {
    return this.request<OIDCConfiguration>('GET', '/.well-known/openid-configuration');
  }

  /**
   * Handle authorization
   */
  async authorize(request: OIDCAuthorizationRequest): Promise<OIDCAuthorizationResponse> {
    return this.request<OIDCAuthorizationResponse>('GET', '/authorize', request);
  }

  /**
   * Exchange code for tokens
   */
  async token(request: OIDCTokenRequest): Promise<OIDCTokenResponse> {
    return this.request<OIDCTokenResponse>('POST', '/token', request);
  }

  /**
   * Get user info
   */
  async getUserInfo(accessToken: string): Promise<OIDCUserInfo> {
    return this.request<OIDCUserInfo>(
      'GET',
      '/userinfo',
      undefined
    );
  }

  /**
   * Get JWKS
   */
  async getJWKS(): Promise<JWKS> {
    return this.request<JWKS>('GET', '/.well-known/jwks.json');
  }

  // Utility Methods

  /**
   * Test connectivity
   */
  async testConnection(): Promise<boolean> {
    try {
      await this.healthCheck();
      return true;
    } catch {
      return false;
    }
  }

  /**
   * Get configuration
   */
  getConfig(): Required<MotorServiceWorkerConfig> {
    return { ...this.config };
  }

  /**
   * Update configuration
   */
  updateConfig(newConfig: Partial<MotorServiceWorkerConfig>): void {
    Object.assign(this.config, newConfig);
    
    if (this.config.debug) {
      console.debug('[MotorClient] Updated config:', this.config);
    }
  }

  /**
   * Send message to service worker
   */
  async sendServiceWorkerMessage(type: string, data?: unknown): Promise<unknown> {
    if (!navigator.serviceWorker?.controller) {
      throw new Error('No service worker controller available');
    }

    return new Promise((resolve, reject) => {
      const messageChannel = new MessageChannel();
      const timeout = setTimeout(() => {
        reject(new Error(`Service worker message timeout: ${type}`));
      }, this.config.timeout);

      messageChannel.port1.onmessage = (event) => {
        clearTimeout(timeout);
        
        if (event.data.error) {
          reject(new Error(event.data.error));
        } else {
          resolve(event.data);
        }
      };

      navigator.serviceWorker.controller.postMessage(
        { type, data },
        [messageChannel.port2]
      );
    });
  }

  /**
   * Clear service worker cache
   */
  async clearServiceWorkerCache(): Promise<void> {
    if (this.useServiceWorker) {
      await this.sendServiceWorkerMessage('CLEAR_CACHE');
    }
  }

  /**
   * Check if service worker is ready
   */
  async isServiceWorkerReady(): Promise<boolean> {
    if (!this.serviceWorkerReady) {
      return false;
    }
    return this.serviceWorkerReady;
  }
}