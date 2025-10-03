/**
 * TypeScript type definitions for Motor Payment Gateway & OIDC Authorization
 * 
 * @packageDocumentation
 */

// ╭─────────────────────────────────────────────────────────╮
// │                  Payment Gateway Types                 │
// ╰─────────────────────────────────────────────────────────╯

/**
 * Payment instrument configuration
 */
export interface PaymentInstrument {
  /** Unique identifier for the instrument */
  id: string;
  /** Display name for the instrument */
  name: string;
  /** Payment method type (e.g., "card", "bank", "crypto") */
  type: string;
  /** Whether this instrument is enabled */
  enabled: boolean;
  /** Supported currencies */
  supportedCurrencies?: string[];
  /** Icon URL for the instrument */
  iconUrl?: string;
}

/**
 * Payment method details
 */
export interface PaymentMethod {
  /** Payment method identifier */
  supportedMethods: string;
  /** Method-specific data */
  data?: Record<string, unknown>;
}

/**
 * Payment details for a transaction
 */
export interface PaymentDetails {
  /** Unique transaction ID */
  id: string;
  /** Total amount */
  total: {
    label: string;
    amount: {
      currency: string;
      value: string;
    };
  };
  /** Display items */
  displayItems?: Array<{
    label: string;
    amount: {
      currency: string;
      value: string;
    };
  }>;
  /** Modifiers for specific payment methods */
  modifiers?: Array<{
    supportedMethods: string;
    total?: {
      label: string;
      amount: {
        currency: string;
        value: string;
      };
    };
    additionalDisplayItems?: Array<{
      label: string;
      amount: {
        currency: string;
        value: string;
      };
    }>;
    data?: Record<string, unknown>;
  }>;
}

/**
 * Request to check if payment can be made
 */
export interface CanMakePaymentRequest {
  /** Origin of the payment request */
  origin: string;
  /** Payment methods to check */
  methodData: PaymentMethod[];
  /** Payment modifiers */
  modifiers?: Array<{
    supportedMethods: string;
    data?: Record<string, unknown>;
  }>;
}

/**
 * Response for can make payment check
 */
export interface CanMakePaymentResponse {
  /** Whether payment can be made */
  canMakePayment: boolean;
}

/**
 * Payment request event data (W3C Payment Handler API)
 */
export interface PaymentRequestEvent {
  /** Payment request ID */
  paymentRequestId: string;
  /** Payment request origin */
  paymentRequestOrigin: string;
  /** Top-level origin */
  topOrigin: string;
  /** Payment method data */
  methodData: PaymentMethod[];
  /** Total amount */
  total: {
    currency: string;
    value: string;
  };
  /** Additional payment details */
  modifiers?: Array<{
    supportedMethods: string;
    total?: {
      currency: string;
      value: string;
    };
    data?: Record<string, unknown>;
  }>;
  /** Instrument key for the payment */
  instrumentKey?: string;
}

/**
 * Payment response from handler
 */
export interface PaymentHandlerResponse {
  /** Payment method name */
  methodName: string;
  /** Payment details */
  details: Record<string, unknown>;
}

/**
 * Request to process a payment
 */
export interface ProcessPaymentRequest {
  /** Payment method */
  method: string;
  /** Amount in smallest currency unit */
  amount: number;
  /** Currency code (e.g., "USD", "EUR") */
  currency: string;
  /** Payment description */
  description?: string;
  /** Customer information */
  customer?: {
    email?: string;
    name?: string;
    id?: string;
  };
  /** Additional metadata */
  metadata?: Record<string, unknown>;
  /** Idempotency key for duplicate prevention */
  idempotencyKey?: string;
}

/**
 * Response from payment processing
 */
export interface ProcessPaymentResponse {
  /** Payment ID */
  paymentId: string;
  /** Payment status */
  status: 'pending' | 'processing' | 'completed' | 'failed' | 'cancelled';
  /** Amount processed */
  amount: number;
  /** Currency used */
  currency: string;
  /** Transaction reference */
  transactionRef?: string;
  /** Processing timestamp */
  processedAt: number;
  /** Error message if failed */
  error?: string;
  /** Next action required (e.g., 3D Secure) */
  nextAction?: {
    type: string;
    redirectUrl?: string;
    data?: Record<string, unknown>;
  };
}

/**
 * Request to validate payment method
 */
export interface ValidatePaymentMethodRequest {
  /** Payment method type */
  method: string;
  /** Method-specific data to validate */
  data: Record<string, unknown>;
}

/**
 * Response from payment method validation
 */
export interface ValidatePaymentMethodResponse {
  /** Whether the payment method is valid */
  valid: boolean;
  /** Validation errors if any */
  errors?: string[];
  /** Sanitized/normalized data */
  normalizedData?: Record<string, unknown>;
}

/**
 * Payment status information
 */
export interface PaymentStatus {
  /** Payment ID */
  paymentId: string;
  /** Current status */
  status: 'pending' | 'processing' | 'completed' | 'failed' | 'cancelled' | 'refunded';
  /** Amount */
  amount: number;
  /** Currency */
  currency: string;
  /** Creation timestamp */
  createdAt: number;
  /** Last update timestamp */
  updatedAt: number;
  /** Transaction details */
  transaction?: {
    id: string;
    reference: string;
    method: string;
  };
  /** Refund information if applicable */
  refund?: {
    amount: number;
    reason?: string;
    refundedAt: number;
  };
}

/**
 * Request to refund a payment
 */
export interface RefundPaymentRequest {
  /** Payment ID to refund */
  paymentId: string;
  /** Amount to refund (partial refund if less than original) */
  amount?: number;
  /** Refund reason */
  reason?: string;
  /** Additional metadata */
  metadata?: Record<string, unknown>;
}

/**
 * Response from payment refund
 */
export interface RefundPaymentResponse {
  /** Refund ID */
  refundId: string;
  /** Payment ID */
  paymentId: string;
  /** Refund status */
  status: 'pending' | 'processing' | 'completed' | 'failed';
  /** Amount refunded */
  amount: number;
  /** Currency */
  currency: string;
  /** Refund timestamp */
  refundedAt: number;
  /** Error message if failed */
  error?: string;
}

// ╭─────────────────────────────────────────────────────────╮
// │                    OIDC Types                          │
// ╰─────────────────────────────────────────────────────────╯

/**
 * OIDC Discovery configuration
 */
export interface OIDCConfiguration {
  /** Issuer identifier */
  issuer: string;
  /** Authorization endpoint */
  authorization_endpoint: string;
  /** Token endpoint */
  token_endpoint: string;
  /** UserInfo endpoint */
  userinfo_endpoint: string;
  /** JWKS URI */
  jwks_uri: string;
  /** Registration endpoint (optional) */
  registration_endpoint?: string;
  /** Supported scopes */
  scopes_supported: string[];
  /** Supported response types */
  response_types_supported: string[];
  /** Supported response modes */
  response_modes_supported?: string[];
  /** Supported grant types */
  grant_types_supported?: string[];
  /** Supported ACR values */
  acr_values_supported?: string[];
  /** Subject types supported */
  subject_types_supported: string[];
  /** ID token signing algorithms */
  id_token_signing_alg_values_supported: string[];
  /** ID token encryption algorithms */
  id_token_encryption_alg_values_supported?: string[];
  /** ID token encryption methods */
  id_token_encryption_enc_values_supported?: string[];
  /** UserInfo signing algorithms */
  userinfo_signing_alg_values_supported?: string[];
  /** UserInfo encryption algorithms */
  userinfo_encryption_alg_values_supported?: string[];
  /** UserInfo encryption methods */
  userinfo_encryption_enc_values_supported?: string[];
  /** Request object signing algorithms */
  request_object_signing_alg_values_supported?: string[];
  /** Request object encryption algorithms */
  request_object_encryption_alg_values_supported?: string[];
  /** Request object encryption methods */
  request_object_encryption_enc_values_supported?: string[];
  /** Token endpoint auth methods */
  token_endpoint_auth_methods_supported?: string[];
  /** Token endpoint auth signing algorithms */
  token_endpoint_auth_signing_alg_values_supported?: string[];
  /** Display values supported */
  display_values_supported?: string[];
  /** Claim types supported */
  claim_types_supported?: string[];
  /** Claims supported */
  claims_supported?: string[];
  /** Service documentation */
  service_documentation?: string;
  /** Claims locales supported */
  claims_locales_supported?: string[];
  /** UI locales supported */
  ui_locales_supported?: string[];
  /** Claims parameter supported */
  claims_parameter_supported?: boolean;
  /** Request parameter supported */
  request_parameter_supported?: boolean;
  /** Request URI parameter supported */
  request_uri_parameter_supported?: boolean;
  /** Require request URI registration */
  require_request_uri_registration?: boolean;
  /** OP policy URI */
  op_policy_uri?: string;
  /** OP terms of service URI */
  op_tos_uri?: string;
  /** PKCE code challenge methods supported */
  code_challenge_methods_supported?: string[];
}

/**
 * OIDC Authorization request
 */
export interface OIDCAuthorizationRequest {
  /** Client identifier */
  client_id: string;
  /** Redirect URI */
  redirect_uri: string;
  /** Response type (e.g., "code", "token") */
  response_type: string;
  /** Requested scopes */
  scope: string;
  /** State parameter for CSRF protection */
  state?: string;
  /** Nonce for replay protection */
  nonce?: string;
  /** Response mode (e.g., "query", "fragment") */
  response_mode?: string;
  /** Display mode (e.g., "page", "popup") */
  display?: string;
  /** Authentication prompt (e.g., "none", "login") */
  prompt?: string;
  /** Maximum authentication age */
  max_age?: number;
  /** UI locales */
  ui_locales?: string;
  /** ID token hint */
  id_token_hint?: string;
  /** Login hint */
  login_hint?: string;
  /** ACR values */
  acr_values?: string;
  /** PKCE code challenge */
  code_challenge?: string;
  /** PKCE code challenge method */
  code_challenge_method?: string;
}

/**
 * OIDC Authorization response
 */
export interface OIDCAuthorizationResponse {
  /** Authorization code */
  code?: string;
  /** Access token (for implicit flow) */
  access_token?: string;
  /** Token type */
  token_type?: string;
  /** ID token */
  id_token?: string;
  /** State parameter */
  state?: string;
  /** Expiration time */
  expires_in?: number;
  /** Authorized scopes */
  scope?: string;
}

/**
 * OIDC Token request
 */
export interface OIDCTokenRequest {
  /** Grant type (e.g., "authorization_code", "refresh_token") */
  grant_type: string;
  /** Authorization code (for authorization_code grant) */
  code?: string;
  /** Redirect URI (for authorization_code grant) */
  redirect_uri?: string;
  /** Client ID */
  client_id: string;
  /** Client secret */
  client_secret?: string;
  /** Refresh token (for refresh_token grant) */
  refresh_token?: string;
  /** Requested scopes (for refresh_token grant) */
  scope?: string;
  /** PKCE code verifier */
  code_verifier?: string;
}

/**
 * OIDC Token response
 */
export interface OIDCTokenResponse {
  /** Access token */
  access_token: string;
  /** Token type (usually "Bearer") */
  token_type: string;
  /** Expiration time in seconds */
  expires_in: number;
  /** Refresh token */
  refresh_token?: string;
  /** ID token */
  id_token?: string;
  /** Authorized scopes */
  scope?: string;
  /** Error code */
  error?: string;
  /** Error description */
  error_description?: string;
}

/**
 * OIDC UserInfo response
 */
export interface OIDCUserInfo {
  /** Subject identifier */
  sub: string;
  /** Full name */
  name?: string;
  /** Given name */
  given_name?: string;
  /** Family name */
  family_name?: string;
  /** Middle name */
  middle_name?: string;
  /** Nickname */
  nickname?: string;
  /** Preferred username */
  preferred_username?: string;
  /** Profile URL */
  profile?: string;
  /** Profile picture URL */
  picture?: string;
  /** Website URL */
  website?: string;
  /** Email address */
  email?: string;
  /** Email verified flag */
  email_verified?: boolean;
  /** Gender */
  gender?: string;
  /** Birth date */
  birthdate?: string;
  /** Time zone */
  zoneinfo?: string;
  /** Locale */
  locale?: string;
  /** Phone number */
  phone_number?: string;
  /** Phone number verified flag */
  phone_number_verified?: boolean;
  /** Address */
  address?: {
    formatted?: string;
    street_address?: string;
    locality?: string;
    region?: string;
    postal_code?: string;
    country?: string;
  };
  /** Last updated timestamp */
  updated_at?: number;
  /** Additional custom claims */
  [key: string]: unknown;
}

/**
 * JSON Web Key Set
 */
export interface JWKS {
  /** Array of JSON Web Keys */
  keys: JWK[];
}

/**
 * JSON Web Key
 */
export interface JWK {
  /** Key type (e.g., "RSA", "EC") */
  kty: string;
  /** Key use (e.g., "sig", "enc") */
  use?: string;
  /** Key operations */
  key_ops?: string[];
  /** Algorithm */
  alg?: string;
  /** Key ID */
  kid?: string;
  /** X.509 URL */
  x5u?: string;
  /** X.509 certificate chain */
  x5c?: string[];
  /** X.509 certificate SHA-1 thumbprint */
  x5t?: string;
  /** X.509 certificate SHA-256 thumbprint */
  'x5t#S256'?: string;
  
  // RSA specific
  /** RSA modulus */
  n?: string;
  /** RSA exponent */
  e?: string;
  
  // EC specific
  /** Elliptic curve */
  crv?: string;
  /** EC x coordinate */
  x?: string;
  /** EC y coordinate */
  y?: string;
}

// ╭─────────────────────────────────────────────────────────╮
// │                Service Worker Types                    │
// ╰─────────────────────────────────────────────────────────╯

/**
 * Configuration for the Motor service worker
 */
export interface MotorServiceWorkerConfig {
  /** URL where the Motor WASM service worker is available */
  worker_url?: string;
  /** Timeout for HTTP requests in milliseconds */
  timeout?: number;
  /** Maximum number of retry attempts */
  max_retries?: number;
  /** Whether to enable debug logging */
  debug?: boolean;
}

/**
 * Health check response from the Motor service worker
 */
export interface HealthCheckResponse {
  /** Service status */
  status: string;
  /** Service name */
  service: string;
  /** Service version */
  version?: string;
  /** Current timestamp */
  time?: number;
  /** Additional health details */
  details?: {
    payment_gateway?: boolean;
    oidc_provider?: boolean;
    memory_usage?: number;
    uptime?: number;
  };
}

/**
 * Service information response
 */
export interface ServiceInfoResponse {
  /** Service name */
  service: string;
  /** Service version */
  version: string;
  /** Service description */
  description: string;
  /** Available endpoints by category */
  endpoints: {
    payment: string[];
    oidc: string[];
    health: string[];
  };
  /** Service capabilities */
  capabilities?: {
    payment_methods?: string[];
    oidc_flows?: string[];
    supported_currencies?: string[];
  };
}

/**
 * Generic error response structure
 */
export interface ErrorResponse {
  /** Error message */
  error: string;
  /** Error code */
  code?: string;
  /** Additional error details */
  details?: Record<string, unknown>;
}

// ╭─────────────────────────────────────────────────────────╮
// │                   Plugin Interface                     │
// ╰─────────────────────────────────────────────────────────╯

/**
 * Main interface for the Motor plugin providing Payment Gateway and OIDC services
 */
export interface MotorPlugin {
  // Payment Gateway Operations
  
  /**
   * Get available payment instruments
   * @returns Promise resolving to payment instruments
   */
  getPaymentInstruments(): Promise<PaymentInstrument[]>;
  
  /**
   * Check if payment can be made
   * @param request Can make payment request
   * @returns Promise resolving to whether payment can be made
   */
  canMakePayment(request: CanMakePaymentRequest): Promise<CanMakePaymentResponse>;
  
  /**
   * Handle payment request event (W3C Payment Handler API)
   * @param event Payment request event
   * @returns Promise resolving to payment response
   */
  handlePaymentRequest(event: PaymentRequestEvent): Promise<PaymentHandlerResponse>;
  
  /**
   * Process a payment transaction
   * @param request Payment processing request
   * @returns Promise resolving to payment response
   */
  processPayment(request: ProcessPaymentRequest): Promise<ProcessPaymentResponse>;
  
  /**
   * Validate a payment method
   * @param request Payment method validation request
   * @returns Promise resolving to validation response
   */
  validatePaymentMethod(request: ValidatePaymentMethodRequest): Promise<ValidatePaymentMethodResponse>;
  
  /**
   * Get payment status
   * @param paymentId Payment identifier
   * @returns Promise resolving to payment status
   */
  getPaymentStatus(paymentId: string): Promise<PaymentStatus>;
  
  /**
   * Process a refund
   * @param request Refund request
   * @returns Promise resolving to refund response
   */
  refundPayment(request: RefundPaymentRequest): Promise<RefundPaymentResponse>;
  
  // OIDC Operations
  
  /**
   * Get OIDC configuration
   * @returns Promise resolving to OIDC configuration
   */
  getOIDCConfiguration(): Promise<OIDCConfiguration>;
  
  /**
   * Handle authorization request
   * @param request Authorization request
   * @returns Promise resolving to authorization response
   */
  authorize(request: OIDCAuthorizationRequest): Promise<OIDCAuthorizationResponse>;
  
  /**
   * Exchange authorization code for tokens
   * @param request Token request
   * @returns Promise resolving to token response
   */
  token(request: OIDCTokenRequest): Promise<OIDCTokenResponse>;
  
  /**
   * Get user information
   * @param accessToken Access token
   * @returns Promise resolving to user info
   */
  getUserInfo(accessToken: string): Promise<OIDCUserInfo>;
  
  /**
   * Get JSON Web Key Set
   * @returns Promise resolving to JWKS
   */
  getJWKS(): Promise<JWKS>;
  
  // Health & Status
  
  /**
   * Check service health
   * @returns Promise resolving to health status
   */
  healthCheck(): Promise<HealthCheckResponse>;
  
  /**
   * Get service information
   * @returns Promise resolving to service info
   */
  getServiceInfo(): Promise<ServiceInfoResponse>;
}

// ╭─────────────────────────────────────────────────────────╮
// │                  Utility Types                         │
// ╰─────────────────────────────────────────────────────────╯

/**
 * Environment detection results
 */
export interface EnvironmentInfo {
  /** Whether running in a browser environment */
  is_browser: boolean;
  /** Whether running in Node.js */
  is_node: boolean;
  /** Whether service workers are supported */
  supports_service_worker: boolean;
  /** Whether WebAssembly is supported */
  supports_wasm: boolean;
  /** Whether Payment Handler API is supported */
  supports_payment_handler?: boolean;
}

/**
 * Service worker registration status
 */
export interface ServiceWorkerStatus {
  /** Whether a service worker is registered */
  registered: boolean;
  /** Service worker state */
  state?: string;
  /** Service worker URL */
  url?: string;
  /** Registration timestamp */
  registered_at?: number;
  /** Whether service worker is ready */
  ready?: boolean;
}