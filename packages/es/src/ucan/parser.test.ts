/**
 * Tests for UCAN JWT token parser.
 */

import { describe, expect, it } from 'vitest';
import { base64urlEncodeJSON } from './encoding.js';
import {
  extractHeader,
  extractPayload,
  isValidJWTFormat,
  parseToken,
  tryParseToken,
} from './parser.js';
import type { UCANHeader, UCANPayload } from './types.js';

describe('parseToken', () => {
  it('should parse a valid UCAN token', () => {
    const header: UCANHeader = {
      alg: 'EdDSA',
      typ: 'JWT',
      ucv: '0.10.0',
    };

    const payload: UCANPayload = {
      iss: 'did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK',
      aud: 'did:key:z6MkrZ1r5XBFZjBU34qyD8fueMbMRkKw17BZaq2ivKFjnz2z',
      exp: 1735689600,
      att: [
        {
          with: 'storage://did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK',
          can: 'crud/read',
        },
      ],
    };

    const headerEncoded = base64urlEncodeJSON(header);
    const payloadEncoded = base64urlEncodeJSON(payload);
    const signatureEncoded = 'dGVzdC1zaWduYXR1cmU'; // "test-signature" in base64url

    const token = `${headerEncoded}.${payloadEncoded}.${signatureEncoded}`;
    const parsed = parseToken(token);

    expect(parsed.header).toEqual(header);
    expect(parsed.payload).toEqual(payload);
    expect(parsed.signature).toBeInstanceOf(Uint8Array);
  });

  it('should parse a token with optional payload fields', () => {
    const header: UCANHeader = {
      alg: 'ES256',
      typ: 'JWT',
      ucv: '0.10.0',
    };

    const payload: UCANPayload = {
      iss: 'did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK',
      aud: 'did:key:z6MkrZ1r5XBFZjBU34qyD8fueMbMRkKw17BZaq2ivKFjnz2z',
      exp: 1735689600,
      nbf: 1704067200,
      nnc: 'unique-nonce-123',
      fct: [{ key: 'value' }],
      prf: ['parent.ucan.token'],
      att: [
        {
          with: 'mailto:user@example.com',
          can: 'msg/send',
          nb: { maxSize: 1024 },
        },
      ],
    };

    const token = `${base64urlEncodeJSON(header)}.${base64urlEncodeJSON(payload)}.dGVzdA`;
    const parsed = parseToken(token);

    expect(parsed.payload.nbf).toBe(1704067200);
    expect(parsed.payload.nnc).toBe('unique-nonce-123');
    expect(parsed.payload.fct).toEqual([{ key: 'value' }]);
    expect(parsed.payload.prf).toEqual(['parent.ucan.token']);
    expect(parsed.payload.att[0].nb).toEqual({ maxSize: 1024 });
  });

  it('should parse a token with RS256 algorithm', () => {
    const header: UCANHeader = {
      alg: 'RS256',
      typ: 'JWT',
      ucv: '0.10.0',
    };

    const payload: UCANPayload = {
      iss: 'did:web:example.com',
      aud: 'did:web:service.com',
      exp: null,
      att: [{ with: 'https://example.com/api', can: 'api/call' }],
    };

    const token = `${base64urlEncodeJSON(header)}.${base64urlEncodeJSON(payload)}.c2lnbmF0dXJl`;
    const parsed = parseToken(token);

    expect(parsed.header.alg).toBe('RS256');
    expect(parsed.payload.exp).toBeNull();
  });

  it('should throw on empty string', () => {
    expect(() => parseToken('')).toThrow('UCAN token must be a non-empty string');
  });

  it('should throw on non-string input', () => {
    expect(() => parseToken(null as unknown as string)).toThrow(
      'UCAN token must be a non-empty string'
    );
    expect(() => parseToken(undefined as unknown as string)).toThrow(
      'UCAN token must be a non-empty string'
    );
  });

  it('should throw on invalid JWT format - too few parts', () => {
    expect(() => parseToken('header.payload')).toThrow('expected 3 parts');
  });

  it('should throw on invalid JWT format - too many parts', () => {
    expect(() => parseToken('header.payload.signature.extra')).toThrow('expected 3 parts');
  });

  it('should throw on empty header segment', () => {
    expect(() => parseToken('.payload.signature')).toThrow('must be non-empty');
  });

  it('should throw on empty payload segment', () => {
    expect(() => parseToken('header..signature')).toThrow('must be non-empty');
  });

  it('should throw on empty signature segment', () => {
    expect(() => parseToken('header.payload.')).toThrow('must be non-empty');
  });

  it('should throw on invalid base64url in header', () => {
    expect(() => parseToken('invalid+base64.payload.signature')).toThrow(
      'Failed to decode JWT header'
    );
  });

  it('should throw on invalid base64url in payload', () => {
    const header = base64urlEncodeJSON({ alg: 'EdDSA', typ: 'JWT', ucv: '0.10.0' });
    expect(() => parseToken(`${header}.invalid+base64.signature`)).toThrow(
      'Failed to decode JWT payload'
    );
  });

  it('should throw on invalid JSON in header', () => {
    const invalidHeader = 'aW52YWxpZC1qc29u'; // "invalid-json" in base64url
    expect(() => parseToken(`${invalidHeader}.payload.signature`)).toThrow(
      'Failed to decode JWT header'
    );
  });

  it('should throw on invalid JSON in payload', () => {
    const header = base64urlEncodeJSON({ alg: 'EdDSA', typ: 'JWT', ucv: '0.10.0' });
    const invalidPayload = 'aW52YWxpZC1qc29u'; // "invalid-json"
    expect(() => parseToken(`${header}.${invalidPayload}.signature`)).toThrow(
      'Failed to decode JWT payload'
    );
  });

  it('should throw on missing header algorithm', () => {
    const invalidHeader = { typ: 'JWT', ucv: '0.10.0' };
    const payload: UCANPayload = {
      iss: 'did:key:abc',
      aud: 'did:key:xyz',
      exp: 1735689600,
      att: [{ with: 'resource', can: 'read' }],
    };

    const token = `${base64urlEncodeJSON(invalidHeader)}.${base64urlEncodeJSON(payload)}.c2ln`;
    expect(() => parseToken(token)).toThrow('Invalid UCAN header');
  });

  it('should throw on invalid header algorithm', () => {
    const invalidHeader = { alg: 'HS256', typ: 'JWT', ucv: '0.10.0' };
    const payload: UCANPayload = {
      iss: 'did:key:abc',
      aud: 'did:key:xyz',
      exp: 1735689600,
      att: [{ with: 'resource', can: 'read' }],
    };

    const token = `${base64urlEncodeJSON(invalidHeader)}.${base64urlEncodeJSON(payload)}.c2ln`;
    expect(() => parseToken(token)).toThrow('Invalid UCAN header');
  });

  it('should throw on invalid header typ', () => {
    const invalidHeader = { alg: 'EdDSA', typ: 'INVALID', ucv: '0.10.0' };
    const payload: UCANPayload = {
      iss: 'did:key:abc',
      aud: 'did:key:xyz',
      exp: 1735689600,
      att: [{ with: 'resource', can: 'read' }],
    };

    const token = `${base64urlEncodeJSON(invalidHeader)}.${base64urlEncodeJSON(payload)}.c2ln`;
    expect(() => parseToken(token)).toThrow('Invalid UCAN header');
  });

  it('should throw on missing header ucv', () => {
    const invalidHeader = { alg: 'EdDSA', typ: 'JWT' };
    const payload: UCANPayload = {
      iss: 'did:key:abc',
      aud: 'did:key:xyz',
      exp: 1735689600,
      att: [{ with: 'resource', can: 'read' }],
    };

    const token = `${base64urlEncodeJSON(invalidHeader)}.${base64urlEncodeJSON(payload)}.c2ln`;
    expect(() => parseToken(token)).toThrow('Invalid UCAN header');
  });

  it('should throw on missing payload issuer', () => {
    const header: UCANHeader = { alg: 'EdDSA', typ: 'JWT', ucv: '0.10.0' };
    const invalidPayload = {
      aud: 'did:key:xyz',
      exp: 1735689600,
      att: [{ with: 'resource', can: 'read' }],
    };

    const token = `${base64urlEncodeJSON(header)}.${base64urlEncodeJSON(invalidPayload)}.c2ln`;
    expect(() => parseToken(token)).toThrow("missing or invalid 'iss'");
  });

  it('should throw on missing payload audience', () => {
    const header: UCANHeader = { alg: 'EdDSA', typ: 'JWT', ucv: '0.10.0' };
    const invalidPayload = {
      iss: 'did:key:abc',
      exp: 1735689600,
      att: [{ with: 'resource', can: 'read' }],
    };

    const token = `${base64urlEncodeJSON(header)}.${base64urlEncodeJSON(invalidPayload)}.c2ln`;
    expect(() => parseToken(token)).toThrow("missing or invalid 'aud'");
  });

  it('should throw on missing payload capabilities', () => {
    const header: UCANHeader = { alg: 'EdDSA', typ: 'JWT', ucv: '0.10.0' };
    const invalidPayload = {
      iss: 'did:key:abc',
      aud: 'did:key:xyz',
      exp: 1735689600,
    };

    const token = `${base64urlEncodeJSON(header)}.${base64urlEncodeJSON(invalidPayload)}.c2ln`;
    expect(() => parseToken(token)).toThrow("missing or empty 'att'");
  });

  it('should throw on empty payload capabilities array', () => {
    const header: UCANHeader = { alg: 'EdDSA', typ: 'JWT', ucv: '0.10.0' };
    const invalidPayload = {
      iss: 'did:key:abc',
      aud: 'did:key:xyz',
      exp: 1735689600,
      att: [],
    };

    const token = `${base64urlEncodeJSON(header)}.${base64urlEncodeJSON(invalidPayload)}.c2ln`;
    expect(() => parseToken(token)).toThrow("missing or empty 'att'");
  });

  it('should throw on invalid capability structure', () => {
    const header: UCANHeader = { alg: 'EdDSA', typ: 'JWT', ucv: '0.10.0' };
    const invalidPayload = {
      iss: 'did:key:abc',
      aud: 'did:key:xyz',
      exp: 1735689600,
      att: [{ with: 'resource' }], // missing 'can'
    };

    const token = `${base64urlEncodeJSON(header)}.${base64urlEncodeJSON(invalidPayload)}.c2ln`;
    expect(() => parseToken(token)).toThrow('Invalid UCAN payload');
  });

  it('should throw on empty signature', () => {
    const header: UCANHeader = { alg: 'EdDSA', typ: 'JWT', ucv: '0.10.0' };
    const payload: UCANPayload = {
      iss: 'did:key:abc',
      aud: 'did:key:xyz',
      exp: 1735689600,
      att: [{ with: 'resource', can: 'read' }],
    };

    // Empty base64url string
    const token = `${base64urlEncodeJSON(header)}.${base64urlEncodeJSON(payload)}.`;
    expect(() => parseToken(token)).toThrow('must be non-empty');
  });
});

describe('tryParseToken', () => {
  it('should return parsed token on success', () => {
    const header: UCANHeader = { alg: 'EdDSA', typ: 'JWT', ucv: '0.10.0' };
    const payload: UCANPayload = {
      iss: 'did:key:abc',
      aud: 'did:key:xyz',
      exp: 1735689600,
      att: [{ with: 'resource', can: 'read' }],
    };

    const token = `${base64urlEncodeJSON(header)}.${base64urlEncodeJSON(payload)}.dGVzdA`;
    const result = tryParseToken(token);

    expect(result).not.toBeNull();
    expect(result?.header.alg).toBe('EdDSA');
    expect(result?.payload.iss).toBe('did:key:abc');
  });

  it('should return null on invalid token', () => {
    expect(tryParseToken('invalid')).toBeNull();
    expect(tryParseToken('')).toBeNull();
    expect(tryParseToken('header.payload')).toBeNull();
  });

  it('should return null on malformed JSON', () => {
    const result = tryParseToken('aW52YWxpZA.aW52YWxpZA.aW52YWxpZA');
    expect(result).toBeNull();
  });
});

describe('isValidJWTFormat', () => {
  it('should return true for valid JWT format', () => {
    expect(isValidJWTFormat('eyJhbGc.eyJpc3M.dGVzdA')).toBe(true);
    expect(isValidJWTFormat('abc123.def456.ghi789')).toBe(true);
    expect(isValidJWTFormat('a-b_c.d-e_f.g-h_i')).toBe(true);
  });

  it('should return false for invalid JWT format', () => {
    expect(isValidJWTFormat('')).toBe(false);
    expect(isValidJWTFormat('   ')).toBe(false);
    expect(isValidJWTFormat('header.payload')).toBe(false);
    expect(isValidJWTFormat('header.payload.signature.extra')).toBe(false);
  });

  it('should return false for empty parts', () => {
    expect(isValidJWTFormat('.payload.signature')).toBe(false);
    expect(isValidJWTFormat('header..signature')).toBe(false);
    expect(isValidJWTFormat('header.payload.')).toBe(false);
  });

  it('should return false for invalid base64url characters', () => {
    expect(isValidJWTFormat('header+plus.payload.signature')).toBe(false);
    expect(isValidJWTFormat('header.payload/slash.signature')).toBe(false);
    expect(isValidJWTFormat('header.payload.signature=')).toBe(false);
  });

  it('should return false for non-string input', () => {
    expect(isValidJWTFormat(null as unknown as string)).toBe(false);
    expect(isValidJWTFormat(undefined as unknown as string)).toBe(false);
    expect(isValidJWTFormat(123 as unknown as string)).toBe(false);
  });
});

describe('extractPayload', () => {
  it('should extract payload without validation', () => {
    const payload = {
      iss: 'did:key:abc',
      aud: 'did:key:xyz',
      exp: 1735689600,
      att: [{ with: 'resource', can: 'read' }],
    };

    const token = `header.${base64urlEncodeJSON(payload)}.signature`;
    const extracted = extractPayload(token);

    expect(extracted).toEqual(payload);
  });

  it('should throw on invalid format', () => {
    expect(() => extractPayload('invalid')).toThrow('expected 3 parts');
  });

  it('should throw on invalid base64url', () => {
    expect(() => extractPayload('header.invalid+base64.signature')).toThrow(
      'Failed to decode payload'
    );
  });
});

describe('extractHeader', () => {
  it('should extract header without validation', () => {
    const header = { alg: 'EdDSA', typ: 'JWT', ucv: '0.10.0' };
    const token = `${base64urlEncodeJSON(header)}.payload.signature`;
    const extracted = extractHeader(token);

    expect(extracted).toEqual(header);
  });

  it('should throw on invalid format', () => {
    expect(() => extractHeader('invalid')).toThrow('expected 3 parts');
  });

  it('should throw on invalid base64url', () => {
    expect(() => extractHeader('invalid+base64.payload.signature')).toThrow(
      'Failed to decode header'
    );
  });
});
