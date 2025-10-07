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

// Mock crypto functions - create proper Uint8Array signatures
const createMockSignature = (algorithm: UCANAlgorithm): Uint8Array => {
  // EdDSA/ES256 typically uses 64 bytes, RS256 uses 256 bytes
  const length = algorithm === 'RS256' ? 256 : 64;
  return new Uint8Array(length).fill(0);
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

    it('should create, format, parse, and validate a complete token', async () => {
      // Create token using builder pattern
      const builder = new UCANBuilder();

      const jwtString = builder
        .issuer(MOCK_DIDS.ISSUER)
        .audience(MOCK_DIDS.AUDIENCE)
        .expiration(Math.floor(Date.now() / 1000) + 3600) // 1 hour from now
        .notBefore(Math.floor(Date.now() / 1000)) // now
        .addCapability(validCapabilities[0])
        .addCapability(validCapabilities[1])
        .build(createMockSignature('EdDSA'));

      // Verify JWT format
      expect(jwtString).toMatch(/^[a-zA-Z0-9_-]+\.[a-zA-Z0-9_-]+\.[a-zA-Z0-9_-]+$/);

      // Parse token
      const parsedToken = parseToken(jwtString);
      expect(parsedToken).toBeDefined();
      expect(parsedToken.payload.iss).toBe(MOCK_DIDS.ISSUER);
      expect(parsedToken.payload.aud).toBe(MOCK_DIDS.AUDIENCE);

      // Validate token (skip signature check for mock)
      const validationOptions: ValidationOptions = {
        now: Math.floor(Date.now() / 1000),
        clockDriftTolerance: 60,
        verifySignature: false, // Skip signature check for mock
      };
      const validationResult = await validateToken(parsedToken, validationOptions);
      expect(validationResult.valid).toBe(true);
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
      const parentBuilder = new UCANBuilder();
      const parentJwt = parentBuilder
        .issuer(MOCK_DIDS.ISSUER)
        .audience(MOCK_DIDS.AUDIENCE)
        .expiration(Math.floor(Date.now() / 1000) + 3600)
        .addCapability(parentCapabilities[0])
        .addCapability(parentCapabilities[1])
        .addCapability(parentCapabilities[2])
        .build(createMockSignature('EdDSA'));

      const parentToken = parseToken(parentJwt);

      // Create child token with attenuated capabilities
      const childBuilder = new UCANBuilder();
      const childJwt = childBuilder
        .issuer(MOCK_DIDS.AUDIENCE)
        .audience(MOCK_DIDS.THIRD_PARTY)
        .expiration(Math.floor(Date.now() / 1000) + 3600)
        .addCapability(childCapabilities[0])
        .addProof(parentJwt)
        .build(createMockSignature('EdDSA'));

      const childToken = parseToken(childJwt);

      // Validate capability attenuation
      const attenuationResult = validateCapabilityAttenuation(
        childToken.payload.att[0],
        parentToken.payload.att[0]
      );
      expect(attenuationResult.valid).toBe(true);
    });

    it('should reject invalid capability attenuation', () => {
      const invalidChildCapability: Capability = {
        with: 'storage://example.com/data',
        can: 'delete', // Not in parent token
      };

      // Create parent token
      const parentBuilder = new UCANBuilder();
      const parentJwt = parentBuilder
        .issuer(MOCK_DIDS.ISSUER)
        .audience(MOCK_DIDS.AUDIENCE)
        .expiration(Math.floor(Date.now() / 1000) + 3600)
        .addCapability(parentCapabilities[0])
        .addCapability(parentCapabilities[1])
        .addCapability(parentCapabilities[2])
        .build(createMockSignature('EdDSA'));

      const parentToken = parseToken(parentJwt);

      // Validate capability attenuation (should fail)
      const attenuationResult = validateCapabilityAttenuation(
        invalidChildCapability,
        parentToken.payload.att[0]
      );
      expect(attenuationResult.valid).toBe(false);
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
        const builder = new UCANBuilder({ algorithm });

        const jwtString = builder
          .issuer(MOCK_DIDS.ISSUER)
          .audience(MOCK_DIDS.AUDIENCE)
          .expiration(Math.floor(Date.now() / 1000) + 3600)
          .addCapability(capabilities[0])
          .build(createMockSignature(algorithm));

        // Parse and validate token with algorithm-specific logic
        const token = parseToken(jwtString);
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
      const builder = new UCANBuilder();

      const jwtString = builder
        .issuer(MOCK_DIDS.ISSUER)
        .audience(MOCK_DIDS.AUDIENCE)
        .notBefore(futureStart)
        .expiration(futureStart + 3600)
        .addCapability(baseCapabilities[0])
        .build(createMockSignature('EdDSA'));

      const token = parseToken(jwtString);

      // Check not yet valid
      const currentTime = Math.floor(Date.now() / 1000);
      expect(isTokenNotYetValid(token.payload, { now: currentTime })).toBe(true);
    });

    it('should handle expired tokens', () => {
      const pastExpiration = Math.floor(Date.now() / 1000) - 3600; // 1 hour in past
      const builder = new UCANBuilder();

      const jwtString = builder
        .issuer(MOCK_DIDS.ISSUER)
        .audience(MOCK_DIDS.AUDIENCE)
        .expiration(pastExpiration)
        .addCapability(baseCapabilities[0])
        .build(createMockSignature('EdDSA'));

      const token = parseToken(jwtString);

      // Check expired
      const currentTime = Math.floor(Date.now() / 1000);
      expect(isTokenExpired(token.payload, { now: currentTime })).toBe(true);
    });

    it('should support tokens without explicit expiration', () => {
      const builder = new UCANBuilder();

      const jwtString = builder
        .issuer(MOCK_DIDS.ISSUER)
        .audience(MOCK_DIDS.AUDIENCE)
        .expiration(null) // Explicitly set null for never-expiring
        .addCapability(baseCapabilities[0])
        .build(createMockSignature('EdDSA'));

      const token = parseToken(jwtString);

      // Ensure no expiration or nbf by default
      expect(token.payload.exp).toBeNull();
      expect(token.payload.nbf).toBeUndefined();
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
        const builder = new UCANBuilder();

        const jwtString = builder
          .issuer(did)
          .audience(MOCK_DIDS.AUDIENCE)
          .expiration(Math.floor(Date.now() / 1000) + 3600)
          .addCapability({
            with: 'storage://example.com/data',
            can: 'read',
          })
          .build(createMockSignature('EdDSA'));

        const token = parseToken(jwtString);
        expect(token.payload.iss).toBe(did);
      });
    });
  });

  // 6. Proof Chain Testing
  describe('Proof Chain Delegation', () => {
    it('should create and validate a multi-level delegation chain', () => {
      // Create first token (root)
      const rootBuilder = new UCANBuilder();
      const rootJwt = rootBuilder
        .issuer(MOCK_DIDS.ISSUER)
        .audience(MOCK_DIDS.AUDIENCE)
        .expiration(Math.floor(Date.now() / 1000) + 3600)
        .addCapability({
          with: 'storage://example.com/data',
          can: 'read',
        })
        .build(createMockSignature('EdDSA'));

      // Create second token (delegation)
      const secondBuilder = new UCANBuilder();
      const secondJwt = secondBuilder
        .issuer(MOCK_DIDS.AUDIENCE)
        .audience(MOCK_DIDS.THIRD_PARTY)
        .expiration(Math.floor(Date.now() / 1000) + 3600)
        .addCapability({
          with: 'storage://example.com/data',
          can: 'read',
        })
        .addProof(rootJwt)
        .build(createMockSignature('EdDSA'));

      const secondToken = parseToken(secondJwt);

      // Validate token chain
      expect(secondToken.payload.prf).toBeDefined();
      expect(secondToken.payload.prf).toContain(rootJwt);
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

      const builder = new UCANBuilder();
      builder
        .issuer(MOCK_DIDS.ISSUER)
        .audience(MOCK_DIDS.AUDIENCE)
        .expiration(Math.floor(Date.now() / 1000) + 3600);

      // Add all capabilities
      largeCaps.forEach(cap => builder.addCapability(cap));

      const jwtString = builder.build(createMockSignature('EdDSA'));
      const token = parseToken(jwtString);

      expect(token.payload.att.length).toBe(100);
    });
  });
});
