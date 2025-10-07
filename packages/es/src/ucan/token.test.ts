/**
 * Tests for UCAN token formatting utilities.
 */

import { describe, expect, it } from 'vitest';
import { base64urlDecode, base64urlDecodeJSON, base64urlEncode } from './encoding.js';
import { parseToken } from './parser.js';
import {
  createSigningMessage,
  formatHeader,
  formatPayload,
  formatToken,
} from './token.js';
import type { UCANHeader, UCANPayload, UCANToken } from './types.js';

describe('formatHeader', () => {
  it('should format a valid UCAN header', () => {
    const header: UCANHeader = {
      alg: 'EdDSA',
      typ: 'JWT',
      ucv: '0.10.0',
    };

    const encoded = formatHeader(header);
    const decoded = base64urlDecodeJSON<UCANHeader>(encoded);

    expect(decoded).toEqual(header);
  });

  it('should format header with deterministic key ordering', () => {
    const header1: UCANHeader = { alg: 'EdDSA', typ: 'JWT', ucv: '0.10.0' };
    const header2: UCANHeader = { ucv: '0.10.0', alg: 'EdDSA', typ: 'JWT' };

    const encoded1 = formatHeader(header1);
    const encoded2 = formatHeader(header2);

    // Same content, different property order -> same encoding
    expect(encoded1).toBe(encoded2);
  });

  it('should format headers with different algorithms', () => {
    const algorithms: Array<'EdDSA' | 'ES256' | 'RS256'> = ['EdDSA', 'ES256', 'RS256'];

    for (const alg of algorithms) {
      const header: UCANHeader = { alg, typ: 'JWT', ucv: '0.10.0' };
      const encoded = formatHeader(header);
      const decoded = base64urlDecodeJSON<UCANHeader>(encoded);

      expect(decoded.alg).toBe(alg);
      expect(decoded.typ).toBe('JWT');
      expect(decoded.ucv).toBe('0.10.0');
    }
  });
});

describe('formatPayload', () => {
  it('should format a minimal payload', () => {
    const payload: UCANPayload = {
      iss: 'did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK',
      aud: 'did:key:z6MkrZ1r5XBFZjBU34qyD8fueMbMRkKw17BZaq2ivKFjnz2z',
      exp: 1735689600,
      att: [{ with: 'storage://did:key:abc', can: 'crud/read' }],
    };

    const encoded = formatPayload(payload);
    const decoded = base64urlDecodeJSON<UCANPayload>(encoded);

    expect(decoded).toEqual(payload);
  });

  it('should format payload with all optional fields', () => {
    const payload: UCANPayload = {
      iss: 'did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK',
      aud: 'did:key:z6MkrZ1r5XBFZjBU34qyD8fueMbMRkKw17BZaq2ivKFjnz2z',
      exp: 1735689600,
      nbf: 1704067200,
      nnc: 'unique-nonce-123',
      fct: [{ key: 'value', nested: { prop: 123 } }],
      prf: ['parent.ucan.token'],
      att: [
        {
          with: 'mailto:user@example.com',
          can: 'msg/send',
          nb: { maxSize: 1024 },
        },
      ],
    };

    const encoded = formatPayload(payload);
    const decoded = base64urlDecodeJSON<UCANPayload>(encoded);

    expect(decoded).toEqual(payload);
  });

  it('should omit undefined optional fields', () => {
    const payload: UCANPayload = {
      iss: 'did:key:abc',
      aud: 'did:key:xyz',
      exp: 1735689600,
      nbf: undefined,
      nnc: undefined,
      fct: undefined,
      prf: undefined,
      att: [{ with: 'resource', can: 'read' }],
    };

    const encoded = formatPayload(payload);
    const decoded = base64urlDecodeJSON<Record<string, unknown>>(encoded);

    // Undefined fields should not be present in encoded JSON
    expect(decoded.nbf).toBeUndefined();
    expect(decoded.nnc).toBeUndefined();
    expect(decoded.fct).toBeUndefined();
    expect(decoded.prf).toBeUndefined();

    // Only defined fields should be present
    expect(decoded.iss).toBe('did:key:abc');
    expect(decoded.aud).toBe('did:key:xyz');
    expect(decoded.exp).toBe(1735689600);
    expect(decoded.att).toBeDefined();
  });

  it('should format payload with deterministic key ordering', () => {
    const payload1: UCANPayload = {
      iss: 'did:key:abc',
      aud: 'did:key:xyz',
      exp: 123,
      att: [{ with: 'resource', can: 'read' }],
    };

    const payload2: UCANPayload = {
      att: [{ with: 'resource', can: 'read' }],
      exp: 123,
      aud: 'did:key:xyz',
      iss: 'did:key:abc',
    };

    const encoded1 = formatPayload(payload1);
    const encoded2 = formatPayload(payload2);

    // Same content, different property order -> same encoding
    expect(encoded1).toBe(encoded2);
  });

  it('should handle null expiration', () => {
    const payload: UCANPayload = {
      iss: 'did:key:abc',
      aud: 'did:key:xyz',
      exp: null,
      att: [{ with: 'resource', can: 'read' }],
    };

    const encoded = formatPayload(payload);
    const decoded = base64urlDecodeJSON<UCANPayload>(encoded);

    expect(decoded.exp).toBeNull();
  });

  it('should sort nested objects and arrays deterministically', () => {
    const payload: UCANPayload = {
      iss: 'did:key:abc',
      aud: 'did:key:xyz',
      exp: 123,
      fct: [
        { z_key: 'last', a_key: 'first', m_key: 'middle' },
        { nested: { z: 3, a: 1, m: 2 } },
      ],
      att: [
        {
          with: 'resource',
          can: 'read',
          nb: { z_param: 'last', a_param: 'first' },
        },
      ],
    };

    const encoded = formatPayload(payload);
    const decoded = base64urlDecodeJSON<UCANPayload>(encoded);

    // Verify nested objects are sorted
    expect(Object.keys(decoded.fct![0])).toEqual(['a_key', 'm_key', 'z_key']);
    expect(Object.keys(decoded.fct![1].nested as Record<string, unknown>)).toEqual([
      'a',
      'm',
      'z',
    ]);
    expect(Object.keys(decoded.att[0].nb!)).toEqual(['a_param', 'z_param']);
  });
});

describe('createSigningMessage', () => {
  it('should create a signing message from header and payload', () => {
    const header: UCANHeader = {
      alg: 'EdDSA',
      typ: 'JWT',
      ucv: '0.10.0',
    };

    const payload: UCANPayload = {
      iss: 'did:key:abc',
      aud: 'did:key:xyz',
      exp: 1735689600,
      att: [{ with: 'resource', can: 'read' }],
    };

    const message = createSigningMessage(header, payload);

    // Should be in format: header.payload (two base64url parts separated by '.')
    const parts = message.split('.');
    expect(parts).toHaveLength(2);

    // Verify parts can be decoded
    const decodedHeader = base64urlDecodeJSON<UCANHeader>(parts[0]);
    const decodedPayload = base64urlDecodeJSON<UCANPayload>(parts[1]);

    expect(decodedHeader).toEqual(header);
    expect(decodedPayload).toEqual(payload);
  });

  it('should create deterministic signing messages', () => {
    const header: UCANHeader = { alg: 'EdDSA', typ: 'JWT', ucv: '0.10.0' };
    const payload: UCANPayload = {
      iss: 'did:key:abc',
      aud: 'did:key:xyz',
      exp: 123,
      att: [{ with: 'r', can: 'read' }],
    };

    const message1 = createSigningMessage(header, payload);
    const message2 = createSigningMessage(header, payload);

    // Multiple invocations with same input should produce identical output
    expect(message1).toBe(message2);
  });
});

describe('formatToken', () => {
  it('should format a complete UCAN token', () => {
    const token: UCANToken = {
      header: {
        alg: 'EdDSA',
        typ: 'JWT',
        ucv: '0.10.0',
      },
      payload: {
        iss: 'did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK',
        aud: 'did:key:z6MkrZ1r5XBFZjBU34qyD8fueMbMRkKw17BZaq2ivKFjnz2z',
        exp: 1735689600,
        att: [{ with: 'storage://did:key:abc', can: 'crud/read' }],
      },
      signature: new TextEncoder().encode('test-signature'),
    };

    const formatted = formatToken(token);

    // Should be in JWT format: header.payload.signature
    const parts = formatted.split('.');
    expect(parts).toHaveLength(3);

    // Verify each part
    const decodedHeader = base64urlDecodeJSON<UCANHeader>(parts[0]);
    const decodedPayload = base64urlDecodeJSON<UCANPayload>(parts[1]);
    const decodedSignature = base64urlDecode(parts[2]);

    expect(decodedHeader).toEqual(token.header);
    expect(decodedPayload).toEqual(token.payload);
    expect(new TextDecoder().decode(decodedSignature)).toBe('test-signature');
  });

  it('should maintain round-trip compatibility with parser', () => {
    const original: UCANToken = {
      header: {
        alg: 'ES256',
        typ: 'JWT',
        ucv: '0.10.0',
      },
      payload: {
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
      },
      signature: new Uint8Array([1, 2, 3, 4, 5]),
    };

    const formatted = formatToken(original);
    const parsed = parseToken(formatted);
    const reformatted = formatToken(parsed);

    // Round-trip guarantee: formatToken(parseToken(x)) === x
    expect(reformatted).toBe(formatted);

    // Verify deep equality of token components
    expect(parsed.header).toEqual(original.header);
    expect(parsed.payload).toEqual(original.payload);
    expect(parsed.signature).toEqual(original.signature);
  });

  it('should format token with RS256 algorithm', () => {
    const token: UCANToken = {
      header: { alg: 'RS256', typ: 'JWT', ucv: '0.10.0' },
      payload: {
        iss: 'did:web:example.com',
        aud: 'did:web:service.com',
        exp: null,
        att: [{ with: 'https://api.example.com', can: 'api/call' }],
      },
      signature: new Uint8Array([255, 128, 64, 32, 16]),
    };

    const formatted = formatToken(token);
    const parsed = parseToken(formatted);

    expect(parsed.header.alg).toBe('RS256');
    expect(parsed.payload.exp).toBeNull();
  });

  it('should handle empty signature', () => {
    const token: UCANToken = {
      header: { alg: 'EdDSA', typ: 'JWT', ucv: '0.10.0' },
      payload: {
        iss: 'did:key:abc',
        aud: 'did:key:xyz',
        exp: 123,
        att: [{ with: 'r', can: 'read' }],
      },
      signature: new Uint8Array([]),
    };

    const formatted = formatToken(token);
    const parts = formatted.split('.');

    // Should still have 3 parts, signature will be empty base64url string
    expect(parts).toHaveLength(3);

    // Parser will reject this as invalid, but formatter should handle it
    const decodedSig = base64urlDecode(parts[2]);
    expect(decodedSig).toHaveLength(0);
  });

  it('should produce deterministic output for identical tokens', () => {
    const token: UCANToken = {
      header: { alg: 'EdDSA', typ: 'JWT', ucv: '0.10.0' },
      payload: {
        iss: 'did:key:abc',
        aud: 'did:key:xyz',
        exp: 123,
        att: [{ with: 'r', can: 'read' }],
      },
      signature: new Uint8Array([1, 2, 3]),
    };

    const formatted1 = formatToken(token);
    const formatted2 = formatToken(token);

    expect(formatted1).toBe(formatted2);
  });

  it('should format token with complex capability caveats', () => {
    const token: UCANToken = {
      header: { alg: 'EdDSA', typ: 'JWT', ucv: '0.10.0' },
      payload: {
        iss: 'did:key:abc',
        aud: 'did:key:xyz',
        exp: 123,
        att: [
          {
            with: 'storage://bucket',
            can: 'crud/write',
            nb: {
              maxSize: 1048576,
              allowedTypes: ['image/jpeg', 'image/png'],
              metadata: { owner: 'did:key:abc', tags: ['public', 'photos'] },
            },
          },
        ],
      },
      signature: new Uint8Array([10, 20, 30]),
    };

    const formatted = formatToken(token);
    const parsed = parseToken(formatted);

    // Verify complex nested caveat structure is preserved
    expect(parsed.payload.att[0].nb).toEqual(token.payload.att[0].nb);
  });

  it('should handle multiple capabilities', () => {
    const token: UCANToken = {
      header: { alg: 'EdDSA', typ: 'JWT', ucv: '0.10.0' },
      payload: {
        iss: 'did:key:abc',
        aud: 'did:key:xyz',
        exp: 123,
        att: [
          { with: 'storage://bucket1', can: 'crud/read' },
          { with: 'storage://bucket2', can: 'crud/write', nb: { maxSize: 1024 } },
          { with: 'mailto:user@example.com', can: 'msg/send' },
        ],
      },
      signature: new Uint8Array([5, 10, 15]),
    };

    const formatted = formatToken(token);
    const parsed = parseToken(formatted);

    expect(parsed.payload.att).toHaveLength(3);
    expect(parsed.payload.att[0].with).toBe('storage://bucket1');
    expect(parsed.payload.att[1].nb).toEqual({ maxSize: 1024 });
    expect(parsed.payload.att[2].can).toBe('msg/send');
  });

  it('should format token with proof chain', () => {
    const token: UCANToken = {
      header: { alg: 'EdDSA', typ: 'JWT', ucv: '0.10.0' },
      payload: {
        iss: 'did:key:abc',
        aud: 'did:key:xyz',
        exp: 123,
        prf: [
          'eyJhbGc.eyJpc3M.abc123',
          'eyJhbGc.eyJpc3M.def456',
        ],
        att: [{ with: 'resource', can: 'read' }],
      },
      signature: new TextEncoder().encode('sig'),
    };

    const formatted = formatToken(token);
    const parsed = parseToken(formatted);

    expect(parsed.payload.prf).toHaveLength(2);
    expect(parsed.payload.prf).toEqual(token.payload.prf);
  });

  it('should handle large signatures', () => {
    const largeSignature = new Uint8Array(256); // RSA-2048 signature size
    for (let i = 0; i < largeSignature.length; i++) {
      largeSignature[i] = i % 256;
    }

    const token: UCANToken = {
      header: { alg: 'RS256', typ: 'JWT', ucv: '0.10.0' },
      payload: {
        iss: 'did:key:abc',
        aud: 'did:key:xyz',
        exp: 123,
        att: [{ with: 'r', can: 'read' }],
      },
      signature: largeSignature,
    };

    const formatted = formatToken(token);
    const parsed = parseToken(formatted);

    expect(parsed.signature).toEqual(largeSignature);
  });
});
