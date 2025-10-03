/**
 * WebAuthn client for Sonr SDK
 */

import {
  startAuthentication,
  startRegistration,
} from '@simplewebauthn/browser';
import type {
  PublicKeyCredentialCreationOptionsJSON,
  PublicKeyCredentialRequestOptionsJSON,
  RegistrationResponseJSON,
  AuthenticationResponseJSON,
} from '@simplewebauthn/types';

import type {
  WebAuthnRegistrationOptions,
  WebAuthnAuthenticationOptions,
  WebAuthnRegistrationResult,
  WebAuthnAuthenticationResult,
  RegisterStartRequest,
  RegisterStartResponse,
  DIDDocument,
  DIDDocumentMetadata,
} from './types';

export class WebAuthnClient {
  private apiUrl: string;
  private origin: string;

  constructor(apiUrl: string, origin?: string) {
    this.apiUrl = apiUrl;
    this.origin = origin || 'http://localhost:3000';
  }

  /**
   * Register a new user with WebAuthn
   */
  async register(options: WebAuthnRegistrationOptions): Promise<WebAuthnRegistrationResult> {
    try {
      // Step 1: Call RegisterStart query to get WebAuthn options
      const startRequest: RegisterStartRequest = {
        assertion_value: options.email || options.tel || options.username,
        assertion_type: options.email ? 'email' : options.tel ? 'tel' : 'username',
        service_origin: options.origin || this.origin,
      };

      const startResponse = await this.callRegisterStart(startRequest);

      // Step 2: Create WebAuthn credential creation options
      const creationOptions: PublicKeyCredentialCreationOptionsJSON = {
        challenge: startResponse.challenge,
        rp: startResponse.rp,
        user: {
          id: startResponse.user.id,
          name: options.username,
          displayName: options.displayName || options.username,
        },
        pubKeyCredParams: (startResponse.pubKeyCredParams as PublicKeyCredentialCreationOptionsJSON['pubKeyCredParams']) || [
          { type: 'public-key' as const, alg: -7 },  // ES256
          { type: 'public-key' as const, alg: -257 }, // RS256
        ],
        timeout: startResponse.timeout || 60000,
        attestation: startResponse.attestation || 'direct',
        authenticatorSelection: startResponse.authenticatorSelection || {
          authenticatorAttachment: 'platform',
          requireResidentKey: false,
          residentKey: 'preferred',
          userVerification: 'preferred',
        },
      };

      // Step 3: Start WebAuthn registration ceremony
      const credential = await startRegistration(creationOptions);

      // Step 4: Submit registration to blockchain
      const result = await this.submitRegistration({
        ...options,
        credential,
        challenge: startResponse.challenge,
      });

      return result;
    } catch (error) {
      console.error('WebAuthn registration error:', error);
      return {
        success: false,
        error: error instanceof Error ? error.message : 'Registration failed',
      };
    }
  }

  /**
   * Authenticate a user with WebAuthn
   */
  async authenticate(options: WebAuthnAuthenticationOptions): Promise<WebAuthnAuthenticationResult> {
    try {
      // Step 1: Get authentication options from the server
      const authOptions = await this.getAuthenticationOptions(options.username);

      // Step 2: Create WebAuthn authentication options
      const requestOptions: PublicKeyCredentialRequestOptionsJSON = {
        challenge: authOptions.challenge,
        rpId: authOptions.rpId,
        timeout: authOptions.timeout || 60000,
        userVerification: authOptions.userVerification || 'preferred',
        allowCredentials: authOptions.allowCredentials,
      };

      // Step 3: Start WebAuthn authentication ceremony
      const credential = await startAuthentication(requestOptions);

      // Step 4: Verify authentication with the server
      const result = await this.verifyAuthentication({
        username: options.username,
        credential,
        challenge: authOptions.challenge,
      });

      return result;
    } catch (error) {
      console.error('WebAuthn authentication error:', error);
      return {
        success: false,
        error: error instanceof Error ? error.message : 'Authentication failed',
      };
    }
  }

  /**
   * Call RegisterStart query
   */
  private async callRegisterStart(request: RegisterStartRequest): Promise<RegisterStartResponse> {
    const response = await fetch(`${this.apiUrl}/did/v1/register/start`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(request),
    });

    if (!response.ok) {
      const error = await response.json().catch(() => ({ error: 'RegisterStart failed' })) as { error: string };
      throw new Error(error.error || 'RegisterStart failed');
    }

    const data = await response.json() as any;

    // Transform the response to match WebAuthn options format
    return {
      challenge: data.challenge || this.generateChallenge(),
      rp: {
        id: data.rp?.id || new URL(this.origin).hostname,
        name: data.rp?.name || 'Sonr Network',
      },
      user: {
        id: data.user?.id || this.generateUserId(),
        name: data.user?.name || request.assertion_value,
        displayName: data.user?.displayName || request.assertion_value,
      },
      pubKeyCredParams: data.pubKeyCredParams || [
        { type: 'public-key' as const, alg: -7 },  // ES256
        { type: 'public-key' as const, alg: -257 }, // RS256
      ],
      timeout: data.timeout,
      attestation: data.attestation,
      authenticatorSelection: data.authenticatorSelection,
    };
  }

  /**
   * Submit registration to blockchain
   */
  private async submitRegistration(data: {
    username: string;
    email?: string;
    tel?: string;
    createVault?: boolean;
    credential: RegistrationResponseJSON;
    challenge: string;
  }): Promise<WebAuthnRegistrationResult> {
    const response = await fetch(`${this.apiUrl}/did/v1/tx/register-webauthn-credential`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        username: data.username,
        assertion_value: data.email || data.tel || data.username,
        assertion_type: data.email ? 'email' : data.tel ? 'tel' : 'username',
        webauthn_credential: {
          credential_id: data.credential.id,
          public_key: data.credential.response.publicKey,
          attestation_object: data.credential.response.attestationObject,
          client_data_json: data.credential.response.clientDataJSON,
          authenticator_attachment: data.credential.authenticatorAttachment,
        },
        create_vault: data.createVault ?? true,
        challenge: data.challenge,
      }),
    });

    if (!response.ok) {
      const error = await response.json().catch(() => ({ error: 'Registration submission failed' })) as { error: string };
      throw new Error(error.error || 'Registration submission failed');
    }

    const result = await response.json() as any;
    return {
      success: true,
      did: result.did,
      vaultId: result.vault_id,
      credential: result.credential,
      ucanToken: result.ucan_token,
    };
  }

  /**
   * Get authentication options
   */
  private async getAuthenticationOptions(username: string): Promise<any> {
    const response = await fetch(`${this.apiUrl}/did/v1/login/start`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ username }),
    });

    if (!response.ok) {
      const error = await response.json().catch(() => ({ error: 'Failed to get authentication options' })) as { error: string };
      throw new Error(error.error || 'Failed to get authentication options');
    }

    return response.json();
  }

  /**
   * Verify authentication
   */
  private async verifyAuthentication(data: {
    username: string;
    credential: AuthenticationResponseJSON;
    challenge: string;
  }): Promise<WebAuthnAuthenticationResult> {
    const response = await fetch(`${this.apiUrl}/did/v1/login/finish`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        username: data.username,
        credential: data.credential,
        challenge: data.challenge,
      }),
    });

    if (!response.ok) {
      const error = await response.json().catch(() => ({ error: 'Authentication verification failed' })) as { error: string };
      throw new Error(error.error || 'Authentication verification failed');
    }

    const result = await response.json() as any;
    return {
      success: true,
      did: result.did,
      vaultId: result.vault_id,
      sessionToken: result.session_token,
    };
  }

  /**
   * Get DID document
   */
  async getDIDDocument(did: string): Promise<{ document: DIDDocument; metadata: DIDDocumentMetadata }> {
    const response = await fetch(`${this.apiUrl}/did/v1/document/${did}`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
      },
    });

    if (!response.ok) {
      const error = await response.json().catch(() => ({ error: 'Failed to get DID document' })) as { error: string };
      throw new Error(error.error || 'Failed to get DID document');
    }

    const data = await response.json() as { did_document: DIDDocument; did_document_metadata: DIDDocumentMetadata };
    return {
      document: data.did_document,
      metadata: data.did_document_metadata,
    };
  }

  /**
   * Generate a random challenge
   */
  private generateChallenge(): string {
    const array = new Uint8Array(32);
    crypto.getRandomValues(array);
    return btoa(String.fromCharCode(...array))
      .replace(/\+/g, '-')
      .replace(/\//g, '_')
      .replace(/=/g, '');
  }

  /**
   * Generate a random user ID
   */
  private generateUserId(): string {
    const array = new Uint8Array(16);
    crypto.getRandomValues(array);
    return btoa(String.fromCharCode(...array))
      .replace(/\+/g, '-')
      .replace(/\//g, '_')
      .replace(/=/g, '');
  }
}