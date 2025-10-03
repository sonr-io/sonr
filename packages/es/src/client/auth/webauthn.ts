import {
  base64URLStringToBuffer,
  browserSupportsWebAuthn,
  browserSupportsWebAuthnAutofill,
  bufferToBase64URLString,
  platformAuthenticatorIsAvailable,
  startAuthentication,
  startRegistration,
} from '@simplewebauthn/browser';
import type {
  AuthenticationResponseJSON,
  PublicKeyCredentialCreationOptionsJSON,
  PublicKeyCredentialRequestOptionsJSON,
  RegistrationResponseJSON,
} from '@simplewebauthn/types';

// Configuration for WebAuthn operations
export interface WebAuthnConfig {
  // Authenticator preferences
  authenticatorSelection?: {
    authenticatorAttachment?: 'platform' | 'cross-platform';
    requireResidentKey?: boolean;
    residentKey?: 'required' | 'preferred' | 'discouraged';
    userVerification?: 'required' | 'preferred' | 'discouraged';
  };
  
  // Attestation preference
  attestation?: 'none' | 'indirect' | 'direct' | 'enterprise';
  
  // Algorithm preferences (in order of preference)
  algorithms?: number[];
  
  // UI/UX options
  showQROption?: boolean;
  preferPlatformAuthenticator?: boolean;
  
  // Callbacks
  onStart?: (options: any) => void | Promise<void>;
  onComplete?: (credential: any) => void | Promise<void>;
  onError?: (error: Error) => void | Promise<void>;
  onStatusUpdate?: (status: string, type: 'info' | 'success' | 'error' | 'warning') => void;
}

// Default configuration with broad compatibility
export const DEFAULT_WEBAUTHN_CONFIG: WebAuthnConfig = {
  authenticatorSelection: {
    // No authenticatorAttachment to allow both platform and cross-platform
    requireResidentKey: false,
    residentKey: 'preferred',
    userVerification: 'preferred', // Broad compatibility
  },
  attestation: 'none', // Simplest option for broad compatibility
  algorithms: [
    -7,   // ES256 (most common)
    -257, // RS256
    -8,   // EdDSA
  ],
  showQROption: true,
  preferPlatformAuthenticator: false,
};

// Preset configurations for common use cases
export const WEBAUTHN_PRESETS = {
  // Maximum compatibility - works with most devices
  BROAD_COMPATIBILITY: DEFAULT_WEBAUTHN_CONFIG,
  
  // Platform only - for native app feel
  PLATFORM_ONLY: {
    ...DEFAULT_WEBAUTHN_CONFIG,
    authenticatorSelection: {
      authenticatorAttachment: 'platform' as const,
      requireResidentKey: true,
      residentKey: 'required' as const,
      userVerification: 'required' as const,
    },
    showQROption: false,
    preferPlatformAuthenticator: true,
  },
  
  // Security key focused
  SECURITY_KEY: {
    ...DEFAULT_WEBAUTHN_CONFIG,
    authenticatorSelection: {
      authenticatorAttachment: 'cross-platform' as const,
      requireResidentKey: false,
      residentKey: 'discouraged' as const,
      userVerification: 'preferred' as const,
    },
    attestation: 'direct' as const,
    showQROption: false,
  },
  
  // Mobile friendly with QR codes
  MOBILE_FRIENDLY: {
    ...DEFAULT_WEBAUTHN_CONFIG,
    authenticatorSelection: {
      requireResidentKey: false,
      residentKey: 'preferred' as const,
      userVerification: 'preferred' as const,
    },
    showQROption: true,
    preferPlatformAuthenticator: false,
  },
  
  // High security (enterprise)
  HIGH_SECURITY: {
    ...DEFAULT_WEBAUTHN_CONFIG,
    authenticatorSelection: {
      requireResidentKey: true,
      residentKey: 'required' as const,
      userVerification: 'required' as const,
    },
    attestation: 'direct' as const,
  },
};

// Types for passkey registration and login
export interface PasskeyRegistrationOptions {
  username: string;
  rpId?: string;
  rpName?: string;
  email?: string;
  tel?: string;
  displayName?: string;
  createVault?: boolean;
  timeout?: number;
  config?: WebAuthnConfig;
}

export interface PasskeyLoginOptions {
  username: string;
  rpId?: string;
  rpName?: string;
  timeout?: number;
  config?: WebAuthnConfig;
}

export interface PasskeyRegistrationResult {
  success: boolean;
  did?: string;
  vaultId?: string;
  assertionMethods?: string[];
  ucanToken?: string;
  credential?: any;
  error?: string;
}

export interface PasskeyLoginResult {
  success: boolean;
  did?: string;
  vaultId?: string;
  sessionToken?: string;
  error?: string;
}

// Utility exports
export const bufferToBase64url = bufferToBase64URLString;
export const base64urlToBuffer = base64URLStringToBuffer;
export const isWebAuthnSupported = browserSupportsWebAuthn;
export const isWebAuthnAvailable = platformAuthenticatorIsAvailable;
export const isConditionalMediationAvailable = browserSupportsWebAuthnAutofill;

/**
 * Register with a passkey (WebAuthn)
 * Supports email/tel assertion methods for Sonr blockchain
 */
export async function registerWithPasskey(
  apiUrl: string,
  options: PasskeyRegistrationOptions
): Promise<PasskeyRegistrationResult> {
  const config = { ...DEFAULT_WEBAUTHN_CONFIG, ...options.config };
  
  try {
    // Check WebAuthn support
    if (!browserSupportsWebAuthn()) {
      throw new Error('WebAuthn is not supported in this browser');
    }
    
    // Check platform authenticator if preferred
    if (config.preferPlatformAuthenticator) {
      const hasPlatform = await platformAuthenticatorIsAvailable();
      if (!hasPlatform) {
        config.onStatusUpdate?.('Platform authenticator not available, using cross-platform options', 'warning');
      }
    }
    
    config.onStatusUpdate?.('Preparing registration...', 'info');
    
    // If using custom config, build options directly
    if (options.config) {
      const registrationOptions: PublicKeyCredentialCreationOptionsJSON = {
        challenge: generateChallenge(),
        rp: {
          id: options.rpId || window.location.hostname,
          name: options.rpName || 'Sonr Identity',
        },
        user: {
          id: btoa(options.username),
          name: options.username,
          displayName: options.displayName || options.username,
        },
        pubKeyCredParams: (config.algorithms || DEFAULT_WEBAUTHN_CONFIG.algorithms!).map(alg => ({
          type: 'public-key' as const,
          alg,
        })),
        authenticatorSelection: config.authenticatorSelection,
        timeout: options.timeout || 60000,
        attestation: config.attestation as AttestationConveyancePreference,
      };
      
      // Call start callback
      await config.onStart?.(registrationOptions);
      
      config.onStatusUpdate?.('Please interact with your authenticator...', 'info');
      
      // Create credential with WebAuthn
      const credential = await startRegistration(registrationOptions);
      
      // Call complete callback
      await config.onComplete?.(credential);
      
      config.onStatusUpdate?.('Registration successful!', 'success');
      
      // For custom config, return simplified result
      return {
        success: true,
        credential,
        did: `did:sonr:${options.username}`, // Placeholder
      };
    }
    
    // Original flow for Sonr blockchain integration
    const registrationOptions = await beginRegistrationPasskey(apiUrl, options);
    await config.onStart?.(registrationOptions);
    
    config.onStatusUpdate?.('Please interact with your authenticator...', 'info');
    const credential = await startRegistration(registrationOptions);
    await config.onComplete?.(credential);
    
    const result = await finishRegistrationPasskey(
      apiUrl,
      options,
      credential,
      registrationOptions.challenge
    );
    
    config.onStatusUpdate?.('Registration successful!', 'success');
    return result;
  } catch (error) {
    const err = error as Error;
    await config.onError?.(err);
    config.onStatusUpdate?.(`Registration failed: ${err.message}`, 'error');
    
    console.error('Passkey registration failed:', error);
    return {
      success: false,
      error: err.message,
    };
  }
}

/**
 * Login with a passkey (WebAuthn)
 */
export async function loginWithPasskey(
  apiUrl: string,
  options: PasskeyLoginOptions
): Promise<PasskeyLoginResult> {
  const config = { ...DEFAULT_WEBAUTHN_CONFIG, ...options.config };
  
  try {
    // Check WebAuthn support
    if (!browserSupportsWebAuthn()) {
      throw new Error('WebAuthn is not supported in this browser');
    }
    
    config.onStatusUpdate?.('Preparing authentication...', 'info');
    
    // If using custom config, build options directly
    if (options.config) {
      const authOptions: PublicKeyCredentialRequestOptionsJSON = {
        challenge: generateChallenge(),
        rpId: options.rpId || window.location.hostname,
        timeout: options.timeout || 60000,
        userVerification: config.authenticatorSelection?.userVerification || 'preferred',
      };
      
      // Call start callback
      await config.onStart?.(authOptions);
      
      config.onStatusUpdate?.('Please authenticate with your passkey...', 'info');
      
      // Authenticate with WebAuthn
      const credential = await startAuthentication(authOptions);
      
      // Call complete callback
      await config.onComplete?.(credential);
      
      config.onStatusUpdate?.('Authentication successful!', 'success');
      
      // For custom config, return simplified result
      return {
        success: true,
        did: `did:sonr:${options.username}`, // Placeholder
        sessionToken: credential.id,
      };
    }
    
    // Original flow for Sonr blockchain integration
    const loginOptions = await beginLoginPasskey(apiUrl, options);
    await config.onStart?.(loginOptions);
    
    config.onStatusUpdate?.('Please authenticate with your passkey...', 'info');
    const credential = await startAuthentication(loginOptions);
    await config.onComplete?.(credential);
    
    const result = await finishLoginPasskey(
      apiUrl,
      options.username,
      credential,
      loginOptions.challenge
    );
    
    config.onStatusUpdate?.('Authentication successful!', 'success');
    return result;
  } catch (error) {
    const err = error as Error;
    await config.onError?.(err);
    config.onStatusUpdate?.(`Authentication failed: ${err.message}`, 'error');
    
    console.error('Passkey authentication failed:', error);
    return {
      success: false,
      error: err.message,
    };
  }
}

/**
 * Generate a random challenge (for demo purposes)
 * In production, this should come from the server
 */
function generateChallenge(): string {
  const array = new Uint8Array(32);
  crypto.getRandomValues(array);
  return btoa(String.fromCharCode(...array))
    .replace(/\+/g, '-')
    .replace(/\//g, '_')
    .replace(/=/g, '');
}

/**
 * Utility to create a button with WebAuthn registration
 */
export function createRegistrationButton(
  buttonElement: HTMLButtonElement,
  apiUrl: string,
  options: PasskeyRegistrationOptions
): void {
  buttonElement.addEventListener('click', async () => {
    buttonElement.disabled = true;
    const originalText = buttonElement.textContent;
    
    // Default status update if not provided
    const config = options.config || {};
    if (!config.onStatusUpdate) {
      config.onStatusUpdate = (status, type) => {
        buttonElement.textContent = status;
        buttonElement.className = `webauthn-button webauthn-${type}`;
      };
    }
    
    const result = await registerWithPasskey(apiUrl, { ...options, config });
    
    if (result.success) {
      buttonElement.textContent = '✓ Registered';
    } else {
      buttonElement.textContent = originalText;
      buttonElement.disabled = false;
    }
  });
}

/**
 * Utility to create a button with WebAuthn login
 */
export function createLoginButton(
  buttonElement: HTMLButtonElement,
  apiUrl: string,
  options: PasskeyLoginOptions
): void {
  buttonElement.addEventListener('click', async () => {
    buttonElement.disabled = true;
    const originalText = buttonElement.textContent;
    
    // Default status update if not provided
    const config = options.config || {};
    if (!config.onStatusUpdate) {
      config.onStatusUpdate = (status, type) => {
        buttonElement.textContent = status;
        buttonElement.className = `webauthn-button webauthn-${type}`;
      };
    }
    
    const result = await loginWithPasskey(apiUrl, { ...options, config });
    
    if (result.success) {
      buttonElement.textContent = '✓ Logged In';
    } else {
      buttonElement.textContent = originalText;
      buttonElement.disabled = false;
    }
  });
}

/**
 * Check if conditional mediation (autofill) is available
 */
export async function checkConditionalMediationSupport(): Promise<{
  supported: boolean;
  available: boolean;
  platformAuthenticator: boolean;
}> {
  const supported = browserSupportsWebAuthn();
  const available = supported && await browserSupportsWebAuthnAutofill();
  const platformAuthenticator = supported && await platformAuthenticatorIsAvailable();
  
  return {
    supported,
    available,
    platformAuthenticator,
  };
}

// Internal helper functions

async function beginRegistrationPasskey(
  apiUrl: string,
  options: PasskeyRegistrationOptions
): Promise<PublicKeyCredentialCreationOptionsJSON> {
  // Determine assertion type and value
  const assertionValue = options.email || options.tel || options.username;
  const assertionType = options.email ? 'email' : options.tel ? 'tel' : 'username';
  const serviceOrigin = typeof window !== 'undefined' ? window.location.origin : options.rpId;

  // Call Sonr's RegisterStart query
  const response = await fetch(`${apiUrl}/did/v1/register/start`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      assertion_value: assertionValue,
      assertion_type: assertionType,
      service_origin: serviceOrigin,
    }),
  });

  if (!response.ok) {
    const error = await response
      .json()
      .catch(() => ({ error: 'Failed to get registration options' }));
    throw new Error(error.error || 'Failed to get registration options');
  }

  const data = await response.json();

  // Generate a challenge if not provided
  const challenge = data.challenge || generateChallenge();

  // Create WebAuthn options
  const publicKeyOptions: PublicKeyCredentialCreationOptionsJSON = {
    challenge,
    rp: {
      id: data.rp?.id || options.rpId,
      name: data.rp?.name || options.rpName,
    },
    user: {
      id: data.user?.id || generateUserId(),
      name: options.username,
      displayName: options.displayName || options.username,
    },
    pubKeyCredParams: data.pubKeyCredParams || [
      { alg: -7, type: 'public-key' },  // ES256
      { alg: -257, type: 'public-key' }, // RS256
    ],
    timeout: data.timeout || options.timeout || 60000,
    attestation: data.attestation || 'direct',
    authenticatorSelection: data.authenticatorSelection || {
      authenticatorAttachment: 'platform',
      requireResidentKey: false,
      userVerification: 'preferred',
    },
  };

  return publicKeyOptions;
}

async function finishRegistrationPasskey(
  apiUrl: string,
  options: PasskeyRegistrationOptions,
  credential: RegistrationResponseJSON,
  challenge: string
): Promise<PasskeyRegistrationResult> {
  const assertionValue = options.email || options.tel || options.username;
  const assertionType = options.email ? 'email' : options.tel ? 'tel' : 'username';

  const response = await fetch(`${apiUrl}/did/v1/tx/register-webauthn-credential`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      username: options.username,
      assertion_value: assertionValue,
      assertion_type: assertionType,
      webauthn_credential: {
        credential_id: credential.id,
        public_key: credential.response.publicKey,
        attestation_object: credential.response.attestationObject,
        client_data_json: credential.response.clientDataJSON,
        authenticator_attachment: credential.authenticatorAttachment,
      },
      create_vault: options.createVault ?? true,
      challenge,
    }),
  });

  if (!response.ok) {
    const error = await response
      .json()
      .catch(() => ({ error: 'Registration submission failed' }));
    throw new Error(error.error || 'Registration submission failed');
  }

  const result = await response.json();
  return {
    success: true,
    did: result.did,
    vaultId: result.vault_id,
    assertionMethods: [
      `did:sonr:${options.username}`,
      `did:${assertionType}:${assertionValue}`,
    ],
    ucanToken: result.ucan_token,
    credential: result.credential,
  };
}

async function beginLoginPasskey(
  apiUrl: string,
  options: PasskeyLoginOptions
): Promise<PublicKeyCredentialRequestOptionsJSON> {
  const url = new URL(`${apiUrl}/did/v1/login/start`);
  
  const response = await fetch(url.toString(), {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ username: options.username }),
  });

  if (!response.ok) {
    const error = await response.text();
    throw new Error('Failed to get authentication options: ' + error);
  }

  const loginOptions = await response.json();

  const publicKeyOptions: PublicKeyCredentialRequestOptionsJSON = {
    challenge: loginOptions.challenge,
    rpId: loginOptions.rpId || options.rpId,
    allowCredentials: loginOptions.allowCredentials,
    userVerification: loginOptions.userVerification || 'preferred',
    timeout: loginOptions.timeout || options.timeout || 30000,
  };

  return publicKeyOptions;
}

async function finishLoginPasskey(
  apiUrl: string,
  username: string,
  credential: AuthenticationResponseJSON,
  challenge: string
): Promise<PasskeyLoginResult> {
  const response = await fetch(`${apiUrl}/did/v1/login/finish`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      username,
      credential,
      challenge,
    }),
  });

  if (!response.ok) {
    const error = await response
      .json()
      .catch(() => ({ error: 'Authentication verification failed' }));
    throw new Error(error.error || 'Authentication verification failed');
  }

  const result = await response.json();
  return {
    success: true,
    did: result.did,
    vaultId: result.vault_id,
    sessionToken: result.session_token,
  };
}


// Helper function to generate a random user ID
function generateUserId(): string {
  const array = new Uint8Array(16);
  if (typeof window !== 'undefined' && window.crypto) {
    window.crypto.getRandomValues(array);
  } else {
    // Fallback for Node.js environment
    for (let i = 0; i < array.length; i++) {
      array[i] = Math.floor(Math.random() * 256);
    }
  }
  return bufferToBase64URLString(array.buffer);
}