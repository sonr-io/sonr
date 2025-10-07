import { describe, it, expect } from 'vitest';
import {
  UCANBuilder,
  parseToken,
  validateToken,
  formatToken,
  formatPayload,
  isTokenExpired,
  isTokenNotYetValid,
  validateCapabilityAttenuation,
} from './index';
import type {
  UCANToken,
  Capability,
  ValidationOptions,
  UCANAlgorithm,
} from './types';

// Mock DIDs for testing
const MOCK_DIDS = {
  ISSUER: 'did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGTsKovqrifQ',
  AUDIENCE: 'did:web:example.com',
  THIRD_PARTY: 'did:key:z6Mkw1nNDPqFXZ2D4CQCCkWMqTJPRvBL8QrQsMvHGMhxr',
};

// Mock crypto functions (would be replaced with real implementations)
const mockSignature = {
  EdDSA: 'mock-signature-eddsa',
  ES256: 'mock-signature-es256',
  RS256: 'mock-signature-rs256',
};

describe('UCAN Integration Tests', () => {
  // 1. End-to-End Token Lifecycle
  describe('Token Lifecycle', () => {
    const validCapabilities: Capability[] = [
      {
        with: 'storage://example.com/data',
        can: 'read',
      },
      {
        with: 'storage://example.com/data',
        can: 'write',
      }
    ];

    it('should create, format, parse, and validate a complete token', () => {
      // Create token
      const builder = new UCANBuilder({
        issuer: MOCK_DIDS.ISSUER,
        audience: MOCK_DIDS.AUDIENCE,
        capabilities: validCapabilities,
      });

      // Set default options
      const token = builder
        .setExpiration(Math.floor(Date.now() / 1000) + 3600) // 1 hour from now
        .setNotBefore(Math.floor(Date.now() / 1000)) // now
        .build();

      // Format token
      const formattedToken = formatToken(token);
      expect(formattedToken).toMatch(/^[a-zA-Z0-9_-]+\.[a-zA-Z0-9_-]+\.[a-zA-Z0-9_-]+$/);

      // Parse token
      const parsedToken = parseToken(formattedToken);
      expect(parsedToken).toBeDefined();

      // Validate token
      const validationOptions: ValidationOptions = {
        now: Math.floor(Date.now() / 1000),
        checkTimestamps: true,
        checkSignature: false, // Skip signature check for mock
      };
      const validationResult = validateToken(parsedToken, validationOptions);
      expect(validationResult.ok).toBe(true);
    });
  });

  // 2. Capability Chain Testing
  describe('Capability Attenuation', () => {
    const parentCapabilities: Capability[] = [
      {
        with: 'storage://example.com/data',
        can: 'read',
      },
      {
        with: 'storage://example.com/data',
        can: 'write',
      },
      {
        with: 'storage://example.com/metadata',
        can: 'read',
      }
    ];

    const childCapabilities: Capability[] = [
      {
        with: 'storage://example.com/data',
        can: 'read',
      }
    ];

    it('should validate proper capability attenuation', () => {
      // Create parent token
      const parentBuilder = new UCANBuilder({
        issuer: MOCK_DIDS.ISSUER,
        audience: MOCK_DIDS.AUDIENCE,
        capabilities: parentCapabilities,
      });

      const parentToken = parentBuilder.build();

      // Create child token with attenuated capabilities
      const childBuilder = new UCANBuilder({
        issuer: MOCK_DIDS.AUDIENCE,
        audience: MOCK_DIDS.THIRD_PARTY,
        capabilities: childCapabilities,
        proofs: [parentToken],
      });

      const childToken = childBuilder.build();

      // Validate capability attenuation
      const attenuationResult = validateCapabilityAttenuation(
        childToken.payload.caps,
        parentToken.payload.caps
      );
      expect(attenuationResult).toBe(true);
    });

    it('should reject invalid capability attenuation', () => {
      const invalidChildCapabilities: Capability[] = [
        {
          with: 'storage://example.com/data',
          can: 'delete', // Not in parent token
        }
      ];

      // Create parent token
      const parentBuilder = new UCANBuilder({
        issuer: MOCK_DIDS.ISSUER,
        audience: MOCK_DIDS.AUDIENCE,
        capabilities: parentCapabilities,
      });

      const parentToken = parentBuilder.build();

      // Try to create child token with broader capabilities
      const childBuilder = new UCANBuilder({
        issuer: MOCK_DIDS.AUDIENCE,
        audience: MOCK_DIDS.THIRD_PARTY,
        capabilities: invalidChildCapabilities,
        proofs: [parentToken],
      });

      const childToken = childBuilder.build();

      // Validate capability attenuation (should fail)
      const attenuationResult = validateCapabilityAttenuation(
        childToken.payload.caps,
        parentToken.payload.caps
      );
      expect(attenuationResult).toBe(false);
    });
  });

  // 3. Multi-Algorithm Support
  describe('Multi-Algorithm Token Support', () => {
    const algorithms: UCANAlgorithm[] = ['EdDSA', 'ES256', 'RS256'];
    const capabilities: Capability[] = [{
      with: 'storage://example.com/data',
      can: 'read',
    }];

    algorithms.forEach(algorithm => {
      it(`should create and validate token with ${algorithm} algorithm`, () => {
        const builder = new UCANBuilder({
          issuer: MOCK_DIDS.ISSUER,
          audience: MOCK_DIDS.AUDIENCE,
          capabilities,
          algorithm, // Specify algorithm
        });

        const token = builder
          .setExpiration(Math.floor(Date.now() / 1000) + 3600)
          .build();

        // Validate token with algorithm-specific logic
        expect(token.header.alg).toBe(algorithm);
      });
    });
  });

  // 4. Timestamp Scenarios
  describe('Token Timestamp Scenarios', () => {
    const baseCapabilities: Capability[] = [{
      with: 'storage://example.com/data',
      can: 'read',
    }];

    it('should handle tokens with future start time', () => {
      const futureStart = Math.floor(Date.now() / 1000) + 3600; // 1 hour in future
      const builder = new UCANBuilder({
        issuer: MOCK_DIDS.ISSUER,
        audience: MOCK_DIDS.AUDIENCE,
        capabilities: baseCapabilities,
      });

      const token = builder
        .setNotBefore(futureStart)
        .setExpiration(futureStart + 3600)
        .build();

      // Check not yet valid
      const currentTime = Math.floor(Date.now() / 1000);
      expect(isTokenNotYetValid(token, currentTime)).toBe(true);
    });

    it('should handle expired tokens', () => {
      const pastExpiration = Math.floor(Date.now() / 1000) - 3600; // 1 hour in past
      const builder = new UCANBuilder({
        issuer: MOCK_DIDS.ISSUER,
        audience: MOCK_DIDS.AUDIENCE,
        capabilities: baseCapabilities,
      });

      const token = builder
        .setExpiration(pastExpiration)
        .build();

      // Check expired
      const currentTime = Math.floor(Date.now() / 1000);
      expect(isTokenExpired(token, currentTime)).toBe(true);
    });

    it('should support tokens without explicit expiration', () => {
      const builder = new UCANBuilder({
        issuer: MOCK_DIDS.ISSUER,
        audience: MOCK_DIDS.AUDIENCE,
        capabilities: baseCapabilities,
      });

      const token = builder.build();

      // Ensure no expiration or nbf by default
      expect(token.payload.exp).toBeNull();
      expect(token.payload.nbf).toBeNull();
    });
  });

  // 5. DID Integration (Simplified mock scenarios)
  describe('DID Integration', () => {
    const didFormats = [
      'did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGTsKovqrifQ',
      'did:web:example.com',
      'did:pkh:eip155:1:0x1234567890123456789012345678901234567890',
    ];

    didFormats.forEach(did => {
      it(`should handle DID format: ${did}`, () => {
        const builder = new UCANBuilder({
          issuer: did,
          audience: MOCK_DIDS.AUDIENCE,
          capabilities: [{
            with: 'storage://example.com/data',
            can: 'read',
          }],
        });

        const token = builder.build();
        expect(token.payload.iss).toBe(did);
      });
    });
  });

  // 6. Proof Chain Testing
  describe('Proof Chain Delegation', () => {
    it('should create and validate a multi-level delegation chain', () => {
      // Create first token (root)
      const rootBuilder = new UCANBuilder({
        issuer: MOCK_DIDS.ISSUER,
        audience: MOCK_DIDS.AUDIENCE,
        capabilities: [{
          with: 'storage://example.com/data',
          can: 'read',
        }],
      });
      const rootToken = rootBuilder.build();

      // Create second token (delegation)
      const secondBuilder = new UCANBuilder({
        issuer: MOCK_DIDS.AUDIENCE,
        audience: MOCK_DIDS.THIRD_PARTY,
        capabilities: [{
          with: 'storage://example.com/data',
          can: 'read',
        }],
        proofs: [rootToken],
      });
      const secondToken = secondBuilder.build();

      // Validate token chain
      expect(secondToken.payload.prf).toContain(formatToken(rootToken));
    });
  });

  // 7. Error Recovery
  describe('Error Handling', () => {
    it('should reject malformed tokens', () => {
      const malformedTokens = [
        'incomplete.token',
        'malformed.token.parts.with.too.many.sections',
        'invalid.base64.token!@#',
      ];

      malformedTokens.forEach(malformedToken => {
        expect(() => parseToken(malformedToken)).toThrow();
      });
    });
  });

  // 8. Performance Testing (Simplified)
  describe('Large Token Scenarios', () => {
    it('should handle large capability lists', () => {
      // Generate 100 capabilities
      const largeCaps = Array.from({ length: 100 }, (_, i) => ({
        with: `storage://example.com/data/${i}`,
        can: 'read',
      }));

      const builder = new UCANBuilder({
        issuer: MOCK_DIDS.ISSUER,
        audience: MOCK_DIDS.AUDIENCE,
        capabilities: largeCaps,
      });

      const token = builder.build();

      expect(token.payload.caps.length).toBe(100);
    });
  });
});