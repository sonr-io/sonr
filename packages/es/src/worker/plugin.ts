/**
 * Motor plugin implementation for Payment Gateway and OIDC Authorization
 * 
 * @packageDocumentation
 */

import type {
  MotorPlugin,
  MotorServiceWorkerConfig,
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
  // Service types
  HealthCheckResponse,
  ServiceInfoResponse,
} from './types';

import { MotorClient } from './client';
import { MotorServiceWorkerManager } from './worker';
import { PaymentGatewayClient } from './payment';
import { OIDCClient } from './oidc';

/**
 * Configuration for the Motor plugin
 */
export interface MotorPluginConfig extends MotorServiceWorkerConfig {
  /** Whether to automatically register the service worker */
  auto_register_worker?: boolean;
  /** Whether to use service worker when available */
  prefer_service_worker?: boolean;
  /** Fallback configuration for direct HTTP calls */
  fallback_url?: string;
  /** OIDC client ID */
  oidc_client_id?: string;
  /** OIDC redirect URI */
  oidc_redirect_uri?: string;
}

/**
 * Default plugin configuration
 */
const DEFAULT_PLUGIN_CONFIG: Required<MotorPluginConfig> = {
  worker_url: '/motor-worker',
  timeout: 30000,
  max_retries: 3,
  debug: false,
  auto_register_worker: true,
  prefer_service_worker: true,
  fallback_url: 'http://localhost:8080',
  oidc_client_id: 'motor-client',
  oidc_redirect_uri: window?.location?.origin || 'http://localhost:3000',
};

/**
 * Motor plugin implementation
 */
export class MotorPluginImpl implements MotorPlugin {
  private readonly config: Required<MotorPluginConfig>;
  private client: MotorClient;
  private paymentClient: PaymentGatewayClient;
  private oidcClient?: OIDCClient;
  private workerManager?: MotorServiceWorkerManager;
  private isInitialized = false;

  constructor(config: Partial<MotorPluginConfig> = {}) {
    this.config = { ...DEFAULT_PLUGIN_CONFIG, ...config };
    this.client = new MotorClient(this.config);
    this.paymentClient = new PaymentGatewayClient(this.config);

    if (this.config.oidc_client_id) {
      this.oidcClient = new OIDCClient({
        ...this.config,
        client_id: this.config.oidc_client_id,
        redirect_uri: this.config.oidc_redirect_uri,
      });
    }

    if (this.config.debug) {
      console.debug('[MotorPlugin] Initialized with config:', this.config);
    }
  }

  /**
   * Initialize the plugin
   */
  async initialize(): Promise<void> {
    if (this.isInitialized) {
      return;
    }

    if (this.config.debug) {
      console.debug('[MotorPlugin] Initializing...');
    }

    if (this.config.prefer_service_worker && MotorServiceWorkerManager.isSupported()) {
      try {
        await this.setupServiceWorker();
      } catch (error) {
        if (this.config.debug) {
          console.warn('[MotorPlugin] Service worker setup failed, using fallback:', error);
        }
        this.setupFallback();
      }
    } else {
      this.setupFallback();
    }

    this.isInitialized = true;

    if (this.config.debug) {
      console.debug('[MotorPlugin] Initialization complete');
    }
  }

  /**
   * Setup service worker
   */
  private async setupServiceWorker(): Promise<void> {
    this.workerManager = new MotorServiceWorkerManager({
      worker_script: '/motor-worker.js',
      scope: '/',
      debug: this.config.debug,
    });

    if (this.config.auto_register_worker) {
      const registration = await this.workerManager.register();
      if (!registration) {
        throw new Error('Failed to register service worker');
      }
    }

    // Wait for service worker to be ready
    await navigator.serviceWorker.ready;
    
    this.client.updateConfig({
      worker_url: this.config.worker_url,
    });
  }

  /**
   * Setup fallback HTTP client
   */
  private setupFallback(): void {
    this.client.updateConfig({
      worker_url: this.config.fallback_url,
    });
  }

  // Payment Gateway Operations

  async getPaymentInstruments(): Promise<PaymentInstrument[]> {
    await this.ensureInitialized();
    return this.paymentClient.getPaymentInstruments();
  }

  async canMakePayment(request: CanMakePaymentRequest): Promise<CanMakePaymentResponse> {
    await this.ensureInitialized();
    return this.paymentClient.canMakePayment(request);
  }

  async handlePaymentRequest(event: PaymentRequestEvent): Promise<PaymentHandlerResponse> {
    await this.ensureInitialized();
    return this.paymentClient.handlePaymentRequest(event);
  }

  async processPayment(request: ProcessPaymentRequest): Promise<ProcessPaymentResponse> {
    await this.ensureInitialized();
    return this.paymentClient.processPayment(request);
  }

  async validatePaymentMethod(request: ValidatePaymentMethodRequest): Promise<ValidatePaymentMethodResponse> {
    await this.ensureInitialized();
    return this.paymentClient.validatePaymentMethod(request);
  }

  async getPaymentStatus(paymentId: string): Promise<PaymentStatus> {
    await this.ensureInitialized();
    return this.paymentClient.getPaymentStatus(paymentId);
  }

  async refundPayment(request: RefundPaymentRequest): Promise<RefundPaymentResponse> {
    await this.ensureInitialized();
    return this.paymentClient.refundPayment(request);
  }

  // OIDC Operations

  async getOIDCConfiguration(): Promise<OIDCConfiguration> {
    await this.ensureInitialized();
    if (!this.oidcClient) {
      throw new Error('OIDC client not configured');
    }
    return this.oidcClient.getConfiguration();
  }

  async authorize(request: OIDCAuthorizationRequest): Promise<OIDCAuthorizationResponse> {
    await this.ensureInitialized();
    if (!this.oidcClient) {
      throw new Error('OIDC client not configured');
    }
    const authUrl = await this.oidcClient.buildAuthorizationUrl(request);
    window.location.href = authUrl;
    return {} as OIDCAuthorizationResponse;
  }

  async token(request: OIDCTokenRequest): Promise<OIDCTokenResponse> {
    await this.ensureInitialized();
    return this.client.token(request);
  }

  async getUserInfo(accessToken: string): Promise<OIDCUserInfo> {
    await this.ensureInitialized();
    if (!this.oidcClient) {
      throw new Error('OIDC client not configured');
    }
    this.oidcClient.setTokens({ access_token: accessToken });
    return this.oidcClient.getUserInfo();
  }

  async getJWKS(): Promise<JWKS> {
    await this.ensureInitialized();
    if (!this.oidcClient) {
      throw new Error('OIDC client not configured');
    }
    return this.oidcClient.getJWKS();
  }

  // Health & Status

  async healthCheck(): Promise<HealthCheckResponse> {
    await this.ensureInitialized();
    return this.client.healthCheck();
  }

  async getServiceInfo(): Promise<ServiceInfoResponse> {
    await this.ensureInitialized();
    return this.client.getServiceInfo();
  }

  // Utility Methods

  /**
   * Ensure plugin is initialized
   */
  private async ensureInitialized(): Promise<void> {
    if (!this.isInitialized) {
      await this.initialize();
    }
  }

  /**
   * Get the HTTP client
   */
  getClient(): MotorClient {
    return this.client;
  }

  /**
   * Get the payment client
   */
  getPaymentClient(): PaymentGatewayClient {
    return this.paymentClient;
  }

  /**
   * Get the OIDC client
   */
  getOIDCClient(): OIDCClient | undefined {
    return this.oidcClient;
  }

  /**
   * Get service worker manager
   */
  getServiceWorkerManager(): MotorServiceWorkerManager | undefined {
    return this.workerManager;
  }

  /**
   * Check if plugin is initialized
   */
  isReady(): boolean {
    return this.isInitialized;
  }
}

// Factory Functions

/**
 * Create a Motor plugin with auto-detection
 */
export async function createMotorPlugin(
  config?: Partial<MotorPluginConfig>
): Promise<MotorPlugin> {
  const plugin = new MotorPluginImpl(config);
  await plugin.initialize();
  return plugin;
}

/**
 * Create a Motor plugin for browser environment
 */
export async function createMotorPluginForBrowser(
  workerUrl = '/motor-worker',
  config?: Partial<MotorPluginConfig>
): Promise<MotorPlugin> {
  const plugin = new MotorPluginImpl({
    ...config,
    worker_url: workerUrl,
    prefer_service_worker: true,
  });
  await plugin.initialize();
  return plugin;
}

/**
 * Create a Motor plugin for Node.js environment
 */
export async function createMotorPluginForNode(
  serverUrl = 'http://localhost:8080',
  config?: Partial<MotorPluginConfig>
): Promise<MotorPlugin> {
  const plugin = new MotorPluginImpl({
    ...config,
    worker_url: serverUrl,
    prefer_service_worker: false,
    auto_register_worker: false,
  });
  await plugin.initialize();
  return plugin;
}