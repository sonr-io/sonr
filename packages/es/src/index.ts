/**
 * @sonr.io/es - Sonr ES Module
 * Main entry point for browser/CDN usage
 */
// Motor Plugin - Service worker and WebAssembly runtime
// Re-export main interfaces and classes for convenience
export {
  MotorPluginImpl,
  createMotorPlugin,
  createMotorPluginForNode,
  createMotorPluginForBrowser,
} from './worker';

export { VaultClient, createVaultClient, getDefaultVaultClient } from './plugin';

// Re-export Motor types
export type {
  MotorPlugin,
  MotorPluginConfig,
  MotorServiceWorkerConfig,
  // Payment types
  PaymentInstrument,
  PaymentMethod,
  PaymentDetails,
  ProcessPaymentRequest,
  ProcessPaymentResponse,
  PaymentStatus,
  // OIDC types
  OIDCConfiguration,
  OIDCTokenRequest,
  OIDCTokenResponse,
  OIDCUserInfo,
} from './worker';

// Re-export Service worker types
export type {
  ServiceWorkerStatus,
  EnvironmentInfo,
  HealthCheckResponse,
  ServiceInfoResponse,
  ErrorResponse,
} from './worker';

// Re-export Vault types
export type {
  VaultConfig,
  VaultPlugin,
  EnclaveData,
  NewOriginTokenRequest,
  NewAttenuatedTokenRequest,
  UCANTokenResponse,
  SignDataRequest,
  SignDataResponse,
  VerifyDataRequest,
  VerifyDataResponse,
  GetIssuerDIDResponse,
} from './plugin';

// Re-export error classes
export { VaultError, VaultErrorCode } from './plugin';
// Re-export auth functions directly for CDN usage
export {
  registerWithPasskey,
  loginWithPasskey,
  isWebAuthnSupported,
  isWebAuthnAvailable,
  isConditionalMediationAvailable,
  bufferToBase64url,
  base64urlToBuffer,
} from './client/auth/webauthn';

// Re-export client functionality
export * from './client';

// Re-export codec utilities
export * from './codec';

// Re-export wallet functionality (be selective to avoid conflicts)
export type { ChainInfo, ConnectedWallet, WalletType } from './wallet';

// Re-export registry
export * from './registry';

// Re-export protobufs
export * from './protobufs';

// Export IPFS services namespace
export * as ipfs from './client/services';

// Vault Plugin - MPC-based cryptographic vault
export * as vault from './plugin';
export * as motor from './worker';

