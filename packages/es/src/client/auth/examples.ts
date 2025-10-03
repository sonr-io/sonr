/**
 * Example usage of Passkey authentication methods
 *
 * These examples demonstrate how to use the passkey helpers
 * with a Sonr blockchain node running locally or remotely.
 */

import {
  registerWithPasskey,
  loginWithPasskey,
  isWebAuthnSupported,
  isWebAuthnAvailable,
  isConditionalMediationAvailable,
} from './webauthn';

// Example 1: Check WebAuthn/Passkey availability
export async function checkPasskeySupport() {
  const supported = isWebAuthnSupported();
  console.log('Passkey supported:', supported);

  if (supported) {
    const available = await isWebAuthnAvailable();
    console.log('Platform authenticator available:', available);

    const conditionalMediation = await isConditionalMediationAvailable();
    console.log('Conditional mediation available:', conditionalMediation);
  }
}

// Example 2: Register a new user with passkey and email
export async function registerUserWithEmail(username: string, email: string) {
  const apiUrl = 'http://localhost:1317'; // Your Sonr node API

  try {
    const result = await registerWithPasskey(apiUrl, {
      username,
      email,
      rpId: 'localhost',
      rpName: 'Sonr Network',
      displayName: username,
      createVault: true,
    });

    if (result.success) {
      console.log('Registration successful!');
      console.log('DID:', result.did);
      console.log('Vault ID:', result.vaultId);
      console.log('Assertion methods:', result.assertionMethods);
    } else {
      console.error('Registration failed:', result.error);
    }
  } catch (error) {
    console.error('Registration error:', error);
  }
}

// Example 3: Register a new user with passkey and phone number
export async function registerUserWithPhone(username: string, phoneNumber: string) {
  const apiUrl = 'http://localhost:1317'; // Your Sonr node API

  try {
    const result = await registerWithPasskey(apiUrl, {
      username,
      tel: phoneNumber,
      rpId: 'localhost',
      rpName: 'Sonr Network',
      displayName: username,
      createVault: true,
    });

    if (result.success) {
      console.log('Registration successful!');
      console.log('DID:', result.did);
      console.log('Vault ID:', result.vaultId);
      console.log('Assertion methods:', result.assertionMethods);
    } else {
      console.error('Registration failed:', result.error);
    }
  } catch (error) {
    console.error('Registration error:', error);
  }
}

// Example 4: Authenticate a user with passkey
export async function authenticateUser(username: string) {
  const apiUrl = 'http://localhost:1317'; // Your Sonr node API

  try {
    const result = await loginWithPasskey(apiUrl, {
      username,
      rpId: 'localhost',
      rpName: 'Sonr Network',
    });

    if (result.success) {
      console.log('Authentication successful!');
      console.log('DID:', result.did);
      console.log('Vault ID:', result.vaultId);
      console.log('Session token:', result.sessionToken);
    } else {
      console.error('Authentication failed:', result.error);
    }
  } catch (error) {
    console.error('Authentication error:', error);
  }
}

// Example 5: Registration with custom server configuration
export async function registerWithCustomServer(
  username: string,
  email: string,
  serverUrl: string,
  rpId: string,
  rpName: string
) {
  try {
    const result = await registerWithPasskey(serverUrl, {
      username,
      email,
      rpId,
      rpName,
      displayName: username,
      createVault: true,
      timeout: 120000, // 2 minutes
    });

    if (result.success) {
      console.log(`User ${username} registered successfully on ${rpName}`);
      console.log('DID:', result.did);
      console.log('Vault ID:', result.vaultId);
    } else {
      console.error('Custom registration failed:', result.error);
    }
  } catch (error) {
    console.error('Custom registration error:', error);
  }
}

// Example 6: Conditional UI authentication (autofill)
export async function setupConditionalAuthentication() {
  const supported = await isConditionalMediationAvailable();

  if (supported) {
    console.log('Conditional mediation is available');
    // You can now use conditional UI for seamless authentication
    // This allows the browser to suggest available credentials
    // in form fields marked with autocomplete="username webauthn"
  } else {
    console.log('Conditional mediation not supported');
    // Fall back to traditional authentication button
  }
}

// Example 7: Full registration flow with error handling
export async function fullRegistrationFlow(
  username: string,
  email?: string,
  phoneNumber?: string
) {
  const apiUrl = 'http://localhost:1317';

  // Check if passkeys are supported
  if (!isWebAuthnSupported()) {
    console.error('Passkeys are not supported in this browser');
    return;
  }

  // Check if platform authenticator is available
  const platformAvailable = await isWebAuthnAvailable();
  if (!platformAvailable) {
    console.warn('No platform authenticator available, using cross-platform');
  }

  try {
    const result = await registerWithPasskey(apiUrl, {
      username,
      email,
      tel: phoneNumber,
      rpId: 'localhost',
      rpName: 'Sonr Network',
      displayName: `${username} User`,
      createVault: true,
    });

    if (result.success) {
      console.log('✅ Registration successful!');
      console.log('📝 DID Document:', result.did);
      console.log('🔐 Vault ID:', result.vaultId);
      console.log('🔑 Assertion Methods:', result.assertionMethods);
      console.log('🎫 UCAN Token:', result.ucanToken);
      
      // Store the UCAN token for future operations
      localStorage.setItem('ucan_token', result.ucanToken || '');
      localStorage.setItem('user_did', result.did || '');
      localStorage.setItem('vault_id', result.vaultId || '');
      
      return result;
    } else {
      console.error('❌ Registration failed:', result.error);
      return null;
    }
  } catch (error) {
    console.error('❌ Unexpected error during registration:', error);
    return null;
  }
}