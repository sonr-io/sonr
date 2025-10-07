/**
 * Tests for UCAN token builder.
 */

import { describe, expect, it } from 'vitest';
import { UCANBuilder } from './builder.js';
import { parseToken } from './parser.js';

describe('UCANBuilder', () => {
  const testSignature = new Uint8Array([1, 2, 3, 4, 5]);
  const validIssuer = 'did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK';
  const validAudience = 'did:key:z6MkrZ1r5XBFZjBU34qyD8fueMbMRkKw17BZaq2ivKFjnz2z';

  describe('constructor', () => {
    it('should create builder with default options', () => {
      const builder = new UCANBuilder();
      const token = builder
        .issuer(validIssuer)
        .audience(validAudience)
        .expiresIn(3600)
        .addCapability({ with: 'resource', can: 'read' })
        .build(testSignature);

      const parsed = parseToken(token);
      expect(parsed.header.alg).toBe('EdDSA');
      expect(parsed.header.ucv).toBe('0.10.0');
    });

    it('should create builder with custom version', () => {
      const builder = new UCANBuilder({ version: '0.9.0' });
      const token = builder
        .issuer(validIssuer)
        .audience(validAudience)
        .expiresIn(3600)
        .addCapability({ with: 'resource', can: 'read' })
        .build(testSignature);

      const parsed = parseToken(token);
      expect(parsed.header.ucv).toBe('0.9.0');
    });

    it('should create builder with custom algorithm', () => {
      const builder = new UCANBuilder({ algorithm: 'ES256' });
      const token = builder
        .issuer(validIssuer)
        .audience(validAudience)
        .expiresIn(3600)
        .addCapability({ with: 'resource', can: 'read' })
        .build(testSignature);

      const parsed = parseToken(token);
      expect(parsed.header.alg).toBe('ES256');
    });
  });

  describe('issuer', () => {
    it('should set valid issuer DID', () => {
      const token = new UCANBuilder()
        .issuer(validIssuer)
        .audience(validAudience)
        .expiresIn(3600)
        .addCapability({ with: 'resource', can: 'read' })
        .build(testSignature);

      const parsed = parseToken(token);
      expect(parsed.payload.iss).toBe(validIssuer);
    });

    it('should throw on invalid DID format', () => {
      expect(() => {
        new UCANBuilder().issuer('not-a-did');
      }).toThrow('Invalid issuer DID format');
    });

    it('should throw on empty DID', () => {
      expect(() => {
        new UCANBuilder().issuer('');
      }).toThrow('Invalid issuer DID format');
    });

    it('should accept various DID methods', () => {
      const builder = new UCANBuilder();
      expect(() => builder.issuer('did:key:abc123')).not.toThrow();
      expect(() => builder.issuer('did:web:example.com')).not.toThrow();
      expect(() => builder.issuer('did:pkh:eth:0x123')).not.toThrow();
    });
  });

  describe('audience', () => {
    it('should set valid audience DID', () => {
      const token = new UCANBuilder()
        .issuer(validIssuer)
        .audience(validAudience)
        .expiresIn(3600)
        .addCapability({ with: 'resource', can: 'read' })
        .build(testSignature);

      const parsed = parseToken(token);
      expect(parsed.payload.aud).toBe(validAudience);
    });

    it('should throw on invalid DID format', () => {
      expect(() => {
        new UCANBuilder().audience('invalid');
      }).toThrow('Invalid audience DID format');
    });
  });

  describe('expiration', () => {
    it('should set valid expiration timestamp', () => {
      const exp = Math.floor(Date.now() / 1000) + 3600;
      const token = new UCANBuilder()
        .issuer(validIssuer)
        .audience(validAudience)
        .expiration(exp)
        .addCapability({ with: 'resource', can: 'read' })
        .build(testSignature);

      const parsed = parseToken(token);
      expect(parsed.payload.exp).toBe(exp);
    });

    it('should accept null expiration', () => {
      const token = new UCANBuilder()
        .issuer(validIssuer)
        .audience(validAudience)
        .expiration(null)
        .addCapability({ with: 'resource', can: 'read' })
        .build(testSignature);

      const parsed = parseToken(token);
      expect(parsed.payload.exp).toBeNull();
    });

    it('should throw on non-integer timestamp', () => {
      expect(() => {
        new UCANBuilder().expiration(123.456);
      }).toThrow('Invalid expiration timestamp');
    });

    it('should throw on out-of-range timestamp', () => {
      expect(() => {
        new UCANBuilder().expiration(Number.MAX_SAFE_INTEGER + 1);
      }).toThrow('Invalid expiration timestamp');
    });
  });

  describe('expiresIn', () => {
    it('should set expiration relative to current time', () => {
      const beforeBuild = Math.floor(Date.now() / 1000);
      const token = new UCANBuilder()
        .issuer(validIssuer)
        .audience(validAudience)
        .expiresIn(3600)
        .addCapability({ with: 'resource', can: 'read' })
        .build(testSignature);
      const afterBuild = Math.floor(Date.now() / 1000);

      const parsed = parseToken(token);
      expect(parsed.payload.exp).toBeGreaterThanOrEqual(beforeBuild + 3600);
      expect(parsed.payload.exp).toBeLessThanOrEqual(afterBuild + 3600);
    });

    it('should throw on zero seconds', () => {
      expect(() => {
        new UCANBuilder().expiresIn(0);
      }).toThrow('Invalid expiresIn value');
    });

    it('should throw on negative seconds', () => {
      expect(() => {
        new UCANBuilder().expiresIn(-100);
      }).toThrow('Invalid expiresIn value');
    });

    it('should throw on infinity', () => {
      expect(() => {
        new UCANBuilder().expiresIn(Number.POSITIVE_INFINITY);
      }).toThrow('Invalid expiresIn value');
    });
  });

  describe('notBefore', () => {
    it('should set valid notBefore timestamp', () => {
      const nbf = Math.floor(Date.now() / 1000);
      const token = new UCANBuilder()
        .issuer(validIssuer)
        .audience(validAudience)
        .notBefore(nbf)
        .expiresIn(3600)
        .addCapability({ with: 'resource', can: 'read' })
        .build(testSignature);

      const parsed = parseToken(token);
      expect(parsed.payload.nbf).toBe(nbf);
    });

    it('should throw on non-integer timestamp', () => {
      expect(() => {
        new UCANBuilder().notBefore(123.456);
      }).toThrow('Invalid notBefore timestamp');
    });
  });

  describe('nonce', () => {
    it('should set valid nonce', () => {
      const nonce = 'unique-nonce-123';
      const token = new UCANBuilder()
        .issuer(validIssuer)
        .audience(validAudience)
        .expiresIn(3600)
        .nonce(nonce)
        .addCapability({ with: 'resource', can: 'read' })
        .build(testSignature);

      const parsed = parseToken(token);
      expect(parsed.payload.nnc).toBe(nonce);
    });

    it('should throw on empty nonce', () => {
      expect(() => {
        new UCANBuilder().nonce('');
      }).toThrow('Nonce must be a non-empty string');
    });
  });

  describe('algorithm', () => {
    it('should override algorithm', () => {
      const token = new UCANBuilder()
        .algorithm('RS256')
        .issuer(validIssuer)
        .audience(validAudience)
        .expiresIn(3600)
        .addCapability({ with: 'resource', can: 'read' })
        .build(testSignature);

      const parsed = parseToken(token);
      expect(parsed.header.alg).toBe('RS256');
    });
  });

  describe('version', () => {
    it('should override version', () => {
      const token = new UCANBuilder()
        .version('0.9.0')
        .issuer(validIssuer)
        .audience(validAudience)
        .expiresIn(3600)
        .addCapability({ with: 'resource', can: 'read' })
        .build(testSignature);

      const parsed = parseToken(token);
      expect(parsed.header.ucv).toBe('0.9.0');
    });
  });

  describe('addCapability', () => {
    it('should add single capability', () => {
      const cap = { with: 'storage://path', can: 'crud/read' };
      const token = new UCANBuilder()
        .issuer(validIssuer)
        .audience(validAudience)
        .expiresIn(3600)
        .addCapability(cap)
        .build(testSignature);

      const parsed = parseToken(token);
      expect(parsed.payload.att).toHaveLength(1);
      expect(parsed.payload.att[0]).toEqual(cap);
    });

    it('should add multiple capabilities', () => {
      const cap1 = { with: 'storage://path1', can: 'crud/read' };
      const cap2 = { with: 'mailto:user@example.com', can: 'msg/send' };
      const token = new UCANBuilder()
        .issuer(validIssuer)
        .audience(validAudience)
        .expiresIn(3600)
        .addCapability(cap1)
        .addCapability(cap2)
        .build(testSignature);

      const parsed = parseToken(token);
      expect(parsed.payload.att).toHaveLength(2);
      expect(parsed.payload.att[0]).toEqual(cap1);
      expect(parsed.payload.att[1]).toEqual(cap2);
    });

    it('should add capability with caveats', () => {
      const cap = {
        with: 'storage://path',
        can: 'crud/write',
        nb: { maxSize: 1024, fileTypes: ['image/jpeg'] },
      };
      const token = new UCANBuilder()
        .issuer(validIssuer)
        .audience(validAudience)
        .expiresIn(3600)
        .addCapability(cap)
        .build(testSignature);

      const parsed = parseToken(token);
      expect(parsed.payload.att[0].nb).toEqual(cap.nb);
    });

    it('should throw on invalid capability structure', () => {
      expect(() => {
        new UCANBuilder().addCapability({ with: 'resource' } as any);
      }).toThrow('Invalid capability structure');
    });
  });

  describe('addFact', () => {
    it('should add single fact', () => {
      const fact = { key: 'value' };
      const token = new UCANBuilder()
        .issuer(validIssuer)
        .audience(validAudience)
        .expiresIn(3600)
        .addCapability({ with: 'resource', can: 'read' })
        .addFact(fact)
        .build(testSignature);

      const parsed = parseToken(token);
      expect(parsed.payload.fct).toHaveLength(1);
      expect(parsed.payload.fct?.[0]).toEqual(fact);
    });

    it('should add multiple facts', () => {
      const fact1 = { 'user-type': 'premium' };
      const fact2 = { region: 'us-west' };
      const token = new UCANBuilder()
        .issuer(validIssuer)
        .audience(validAudience)
        .expiresIn(3600)
        .addCapability({ with: 'resource', can: 'read' })
        .addFact(fact1)
        .addFact(fact2)
        .build(testSignature);

      const parsed = parseToken(token);
      expect(parsed.payload.fct).toHaveLength(2);
    });

    it('should throw on non-object fact', () => {
      expect(() => {
        new UCANBuilder().addFact('not-an-object' as any);
      }).toThrow('Fact must be a non-null object');
    });

    it('should throw on array fact', () => {
      expect(() => {
        new UCANBuilder().addFact([1, 2, 3] as any);
      }).toThrow('Fact must be a non-null object');
    });
  });

  describe('addProof', () => {
    it('should add single proof', () => {
      const proof = 'eyJhbGciOiJFZERTQSIsInR5cCI6IkpXVCJ9.eyJpc3MiOi4uLg';
      const token = new UCANBuilder()
        .issuer(validIssuer)
        .audience(validAudience)
        .expiresIn(3600)
        .addCapability({ with: 'resource', can: 'read' })
        .addProof(proof)
        .build(testSignature);

      const parsed = parseToken(token);
      expect(parsed.payload.prf).toHaveLength(1);
      expect(parsed.payload.prf?.[0]).toBe(proof);
    });

    it('should add multiple proofs', () => {
      const proof1 = 'eyJhbGciOiJFZERTQSIsInR5cCI6IkpXVCJ9.proof1';
      const proof2 = 'eyJhbGciOiJFZERTQSIsInR5cCI6IkpXVCJ9.proof2';
      const token = new UCANBuilder()
        .issuer(validIssuer)
        .audience(validAudience)
        .expiresIn(3600)
        .addCapability({ with: 'resource', can: 'read' })
        .addProof(proof1)
        .addProof(proof2)
        .build(testSignature);

      const parsed = parseToken(token);
      expect(parsed.payload.prf).toHaveLength(2);
    });

    it('should throw on empty proof', () => {
      expect(() => {
        new UCANBuilder().addProof('');
      }).toThrow('Proof UCAN must be a non-empty string');
    });
  });

  describe('build', () => {
    it('should build valid token with all required fields', () => {
      const token = new UCANBuilder()
        .issuer(validIssuer)
        .audience(validAudience)
        .expiresIn(3600)
        .addCapability({ with: 'resource', can: 'read' })
        .build(testSignature);

      expect(typeof token).toBe('string');
      expect(token.split('.')).toHaveLength(3);

      const parsed = parseToken(token);
      expect(parsed.payload.iss).toBe(validIssuer);
      expect(parsed.payload.aud).toBe(validAudience);
      expect(parsed.payload.att).toHaveLength(1);
    });

    it('should throw when issuer is missing', () => {
      expect(() => {
        new UCANBuilder()
          .audience(validAudience)
          .expiresIn(3600)
          .addCapability({ with: 'resource', can: 'read' })
          .build(testSignature);
      }).toThrow('issuer DID is required');
    });

    it('should throw when audience is missing', () => {
      expect(() => {
        new UCANBuilder()
          .issuer(validIssuer)
          .expiresIn(3600)
          .addCapability({ with: 'resource', can: 'read' })
          .build(testSignature);
      }).toThrow('audience DID is required');
    });

    it('should throw when expiration is missing', () => {
      expect(() => {
        new UCANBuilder()
          .issuer(validIssuer)
          .audience(validAudience)
          .addCapability({ with: 'resource', can: 'read' })
          .build(testSignature);
      }).toThrow('expiration is required');
    });

    it('should throw when capabilities are missing', () => {
      expect(() => {
        new UCANBuilder()
          .issuer(validIssuer)
          .audience(validAudience)
          .expiresIn(3600)
          .build(testSignature);
      }).toThrow('at least one capability is required');
    });

    it('should throw when notBefore is after expiration', () => {
      const now = Math.floor(Date.now() / 1000);
      expect(() => {
        new UCANBuilder()
          .issuer(validIssuer)
          .audience(validAudience)
          .notBefore(now + 3600)
          .expiration(now)
          .addCapability({ with: 'resource', can: 'read' })
          .build(testSignature);
      }).toThrow('notBefore');
    });

    it('should throw on invalid signature type', () => {
      expect(() => {
        new UCANBuilder()
          .issuer(validIssuer)
          .audience(validAudience)
          .expiresIn(3600)
          .addCapability({ with: 'resource', can: 'read' })
          .build('not-a-uint8array' as any);
      }).toThrow('Signature must be a Uint8Array');
    });

    it('should throw on empty signature', () => {
      expect(() => {
        new UCANBuilder()
          .issuer(validIssuer)
          .audience(validAudience)
          .expiresIn(3600)
          .addCapability({ with: 'resource', can: 'read' })
          .build(new Uint8Array([]));
      }).toThrow('Signature cannot be empty');
    });
  });

  describe('method chaining', () => {
    it('should support fluent API', () => {
      const token = new UCANBuilder()
        .issuer(validIssuer)
        .audience(validAudience)
        .expiresIn(3600)
        .notBefore(Math.floor(Date.now() / 1000))
        .nonce('nonce-123')
        .addCapability({ with: 'resource1', can: 'read' })
        .addCapability({ with: 'resource2', can: 'write' })
        .addFact({ key: 'value' })
        .addProof('proof-token')
        .algorithm('ES256')
        .version('0.10.0')
        .build(testSignature);

      const parsed = parseToken(token);
      expect(parsed.payload.iss).toBe(validIssuer);
      expect(parsed.payload.aud).toBe(validAudience);
      expect(parsed.payload.att).toHaveLength(2);
      expect(parsed.payload.fct).toHaveLength(1);
      expect(parsed.payload.prf).toHaveLength(1);
      expect(parsed.header.alg).toBe('ES256');
    });
  });

  describe('round-trip parsing', () => {
    it('should produce token that can be parsed back', () => {
      const originalCap = {
        with: 'storage://path',
        can: 'crud/read',
        nb: { maxSize: 1024 },
      };

      const token = new UCANBuilder()
        .issuer(validIssuer)
        .audience(validAudience)
        .expiresIn(3600)
        .addCapability(originalCap)
        .build(testSignature);

      const parsed = parseToken(token);
      expect(parsed.payload.iss).toBe(validIssuer);
      expect(parsed.payload.aud).toBe(validAudience);
      expect(parsed.payload.att[0]).toEqual(originalCap);
      expect(parsed.signature).toBeInstanceOf(Uint8Array);
      expect(Array.from(parsed.signature)).toEqual(Array.from(testSignature));
    });
  });
});
