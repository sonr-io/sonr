// Export passkey authentication functions
export {
  registerWithPasskey,
  loginWithPasskey,
  // Utility functions
  bufferToBase64url,
  base64urlToBuffer,
  isWebAuthnSupported,
  isWebAuthnAvailable,
  isConditionalMediationAvailable,
  // Enhanced utilities
  checkConditionalMediationSupport,
  createRegistrationButton,
  createLoginButton,
  // Configuration and presets
  DEFAULT_WEBAUTHN_CONFIG,
  WEBAUTHN_PRESETS,
} from './webauthn';

// Export types
export type {
  PasskeyRegistrationOptions,
  PasskeyLoginOptions,
  PasskeyRegistrationResult,
  PasskeyLoginResult,
  WebAuthnConfig,
} from './webauthn';
