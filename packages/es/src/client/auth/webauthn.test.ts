/**
 * Integration tests for WebAuthn/Passkey authentication
 */

import { describe, it, expect, beforeAll, afterAll, vi } from 'vitest';
import {
  registerWithPasskey,
  loginWithPasskey,
  isWebAuthnSupported,
  isWebAuthnAvailable,
  bufferToBase64url,
  base64urlToBuffer,
} from './webauthn';

// Mock @simplewebauthn/browser for testing
vi.mock('@simplewebauthn/browser', () => ({
  browserSupportsWebAuthn: () => true,
  platformAuthenticatorIsAvailable: () => Promise.resolve(true),
  browserSupportsWebAuthnAutofill: () => Promise.resolve(true),
  startRegistration: vi.fn(),
  startAuthentication: vi.fn(),
  bufferToBase64URLString: (buffer: ArrayBuffer) => {
    const bytes = new Uint8Array(buffer);
    let binary = '';
    for (let i = 0; i < bytes.byteLength; i++) {
      binary += String.fromCharCode(bytes[i]);
    }
    return btoa(binary).replace(/\+/g, '-').replace(/\//g, '_').replace(/=/g, '');
  },
  base64URLStringToBuffer: (base64url: string) => {
    const base64 = base64url.replace(/-/g, '+').replace(/_/g, '/');
    const binary = atob(base64);
    const bytes = new Uint8Array(binary.length);
    for (let i = 0; i < binary.length; i++) {
      bytes[i] = binary.charCodeAt(i);
    }
    return bytes.buffer;
  },
}));

// Mock fetch for API calls
global.fetch = vi.fn();

describe('WebAuthn/Passkey Authentication', () => {
  const mockApiUrl = 'http://localhost:1317';
  
  beforeAll(() => {
    // Setup any global mocks or test data
  });

  afterAll(() => {
    vi.clearAllMocks();
  });

  describe('Utility Functions', () => {
    it('should check WebAuthn support', () => {
      const supported = isWebAuthnSupported();
      expect(supported).toBe(true);
    });

    it('should check platform authenticator availability', async () => {
      const available = await isWebAuthnAvailable();
      expect(available).toBe(true);
    });

    it('should convert buffer to base64url and back', () => {
      const testString = 'Hello, WebAuthn!';
      const encoder = new TextEncoder();
      const buffer = encoder.encode(testString).buffer;
      
      const base64url = bufferToBase64url(buffer);
      expect(base64url).toBeTruthy();
      expect(base64url).not.toContain('+');
      expect(base64url).not.toContain('/');
      expect(base64url).not.toContain('=');
      
      const decodedBuffer = base64urlToBuffer(base64url);
      const decoder = new TextDecoder();
      const decodedString = decoder.decode(decodedBuffer);
      expect(decodedString).toBe(testString);
    });
  });

  describe('Registration with Passkey', () => {
    it('should successfully register with email assertion', async () => {
      const { startRegistration } = await import('@simplewebauthn/browser');
      
      // Mock RegisterStart response
      (global.fetch as any).mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          challenge: 'test-challenge',
          rp: { id: 'localhost', name: 'Sonr Network' },
          user: { id: 'user-id', name: 'alice', displayName: 'Alice' },
        }),
      });

      // Mock startRegistration
      (startRegistration as any).mockResolvedValueOnce({
        id: 'credential-id',
        rawId: 'credential-raw-id',
        response: {
          publicKey: 'mock-public-key',
          attestationObject: 'mock-attestation',
          clientDataJSON: 'mock-client-data',
        },
        authenticatorAttachment: 'platform',
        type: 'public-key',
      });

      // Mock registration submission response
      (global.fetch as any).mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          did: 'did:email:abc123',
          vault_id: 'vault-123',
          ucan_token: 'ucan-token-123',
          credential: { id: 'credential-id' },
        }),
      });

      const result = await registerWithPasskey(mockApiUrl, {
        username: 'alice',
        email: 'alice@example.com',
        rpId: 'localhost',
        rpName: 'Sonr Network',
        displayName: 'Alice Smith',
        createVault: true,
      });

      expect(result.success).toBe(true);
      expect(result.did).toBe('did:email:abc123');
      expect(result.vaultId).toBe('vault-123');
      expect(result.ucanToken).toBe('ucan-token-123');
      expect(result.assertionMethods).toEqual([
        'did:sonr:alice',
        'did:email:alice@example.com',
      ]);
    });

    it('should successfully register with phone assertion', async () => {
      const { startRegistration } = await import('@simplewebauthn/browser');
      
      // Mock RegisterStart response
      (global.fetch as any).mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          challenge: 'test-challenge',
          rp: { id: 'localhost', name: 'Sonr Network' },
          user: { id: 'user-id', name: 'bob', displayName: 'Bob' },
        }),
      });

      // Mock startRegistration
      (startRegistration as any).mockResolvedValueOnce({
        id: 'credential-id-2',
        rawId: 'credential-raw-id-2',
        response: {
          publicKey: 'mock-public-key-2',
          attestationObject: 'mock-attestation-2',
          clientDataJSON: 'mock-client-data-2',
        },
        authenticatorAttachment: 'platform',
        type: 'public-key',
      });

      // Mock registration submission response
      (global.fetch as any).mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          did: 'did:tel:xyz789',
          vault_id: 'vault-456',
          ucan_token: 'ucan-token-456',
          credential: { id: 'credential-id-2' },
        }),
      });

      const result = await registerWithPasskey(mockApiUrl, {
        username: 'bob',
        tel: '+1234567890',
        rpId: 'localhost',
        rpName: 'Sonr Network',
        displayName: 'Bob Johnson',
        createVault: true,
      });

      expect(result.success).toBe(true);
      expect(result.did).toBe('did:tel:xyz789');
      expect(result.vaultId).toBe('vault-456');
      expect(result.assertionMethods).toEqual([
        'did:sonr:bob',
        'did:tel:+1234567890',
      ]);
    });

    it('should handle registration failure gracefully', async () => {
      // Mock RegisterStart failure
      (global.fetch as any).mockResolvedValueOnce({
        ok: false,
        json: async () => ({ error: 'Invalid origin' }),
      });

      const result = await registerWithPasskey(mockApiUrl, {
        username: 'charlie',
        email: 'charlie@example.com',
        rpId: 'malicious.com',
        rpName: 'Malicious Site',
        createVault: false,
      });

      expect(result.success).toBe(false);
      expect(result.error).toContain('Invalid origin');
    });

    it('should handle WebAuthn ceremony cancellation', async () => {
      const { startRegistration } = await import('@simplewebauthn/browser');
      
      // Mock RegisterStart success
      (global.fetch as any).mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          challenge: 'test-challenge',
          rp: { id: 'localhost', name: 'Sonr Network' },
          user: { id: 'user-id', name: 'dave', displayName: 'Dave' },
        }),
      });

      // Mock user cancellation
      (startRegistration as any).mockRejectedValueOnce(
        new Error('User cancelled the ceremony')
      );

      const result = await registerWithPasskey(mockApiUrl, {
        username: 'dave',
        email: 'dave@example.com',
        rpId: 'localhost',
        rpName: 'Sonr Network',
        createVault: true,
      });

      expect(result.success).toBe(false);
      expect(result.error).toContain('User cancelled');
    });
  });

  describe('Login with Passkey', () => {
    it('should successfully authenticate with passkey', async () => {
      const { startAuthentication } = await import('@simplewebauthn/browser');
      
      // Mock LoginStart response
      (global.fetch as any).mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          challenge: 'login-challenge',
          rpId: 'localhost',
          allowCredentials: [
            { id: 'credential-id', type: 'public-key' },
          ],
          userVerification: 'preferred',
        }),
      });

      // Mock startAuthentication
      (startAuthentication as any).mockResolvedValueOnce({
        id: 'credential-id',
        rawId: 'credential-raw-id',
        response: {
          authenticatorData: 'mock-auth-data',
          clientDataJSON: 'mock-client-data',
          signature: 'mock-signature',
        },
        type: 'public-key',
      });

      // Mock login finish response
      (global.fetch as any).mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          did: 'did:email:abc123',
          vault_id: 'vault-123',
          session_token: 'session-token-xyz',
        }),
      });

      const result = await loginWithPasskey(mockApiUrl, {
        username: 'alice',
        rpId: 'localhost',
      });

      expect(result.success).toBe(true);
      expect(result.did).toBe('did:email:abc123');
      expect(result.vaultId).toBe('vault-123');
      expect(result.sessionToken).toBe('session-token-xyz');
    });

    it('should handle authentication failure', async () => {
      // Mock LoginStart failure
      (global.fetch as any).mockResolvedValueOnce({
        ok: false,
        text: async () => 'User not found',
      });

      const result = await loginWithPasskey(mockApiUrl, {
        username: 'nonexistent',
        rpId: 'localhost',
      });

      expect(result.success).toBe(false);
      expect(result.error).toContain('User not found');
    });

    it('should handle invalid credentials', async () => {
      const { startAuthentication } = await import('@simplewebauthn/browser');
      
      // Mock LoginStart success
      (global.fetch as any).mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          challenge: 'login-challenge',
          rpId: 'localhost',
          allowCredentials: [],
        }),
      });

      // Mock authentication with wrong credential
      (startAuthentication as any).mockResolvedValueOnce({
        id: 'wrong-credential-id',
        rawId: 'wrong-credential-raw-id',
        response: {
          authenticatorData: 'mock-auth-data',
          clientDataJSON: 'mock-client-data',
          signature: 'mock-signature',
        },
        type: 'public-key',
      });

      // Mock login finish failure
      (global.fetch as any).mockResolvedValueOnce({
        ok: false,
        json: async () => ({ error: 'Invalid credential' }),
      });

      const result = await loginWithPasskey(mockApiUrl, {
        username: 'alice',
        rpId: 'localhost',
      });

      expect(result.success).toBe(false);
      expect(result.error).toContain('Invalid credential');
    });
  });

  describe('Edge Cases and Error Handling', () => {
    it('should handle network errors', async () => {
      // Mock network error
      (global.fetch as any).mockRejectedValueOnce(
        new Error('Network request failed')
      );

      const result = await registerWithPasskey(mockApiUrl, {
        username: 'network-test',
        email: 'test@example.com',
        rpId: 'localhost',
        rpName: 'Sonr Network',
        createVault: false,
      });

      expect(result.success).toBe(false);
      expect(result.error).toContain('Network request failed');
    });

    it('should handle malformed API responses', async () => {
      // Mock malformed response
      (global.fetch as any).mockResolvedValueOnce({
        ok: true,
        json: async () => { throw new Error('Invalid JSON'); },
      });

      const result = await registerWithPasskey(mockApiUrl, {
        username: 'malformed-test',
        email: 'test@example.com',
        rpId: 'localhost',
        rpName: 'Sonr Network',
        createVault: false,
      });

      expect(result.success).toBe(false);
      expect(result.error).toBeTruthy();
    });

    it('should handle timeout scenarios', async () => {
      // Mock timeout
      (global.fetch as any).mockImplementationOnce(() => 
        new Promise((_, reject) => 
          setTimeout(() => reject(new Error('Request timeout')), 100)
        )
      );

      const result = await registerWithPasskey(mockApiUrl, {
        username: 'timeout-test',
        email: 'test@example.com',
        rpId: 'localhost',
        rpName: 'Sonr Network',
        timeout: 50, // Very short timeout
        createVault: false,
      });

      expect(result.success).toBe(false);
      expect(result.error).toContain('timeout');
    });
  });
});