import { describe, expect, test } from 'vitest';
import {
  base64urlDecode,
  base64urlDecodeJSON,
  base64urlDecodeString,
  base64urlEncode,
  base64urlEncodeJSON,
  base64urlEncodeString,
  isValidBase64url,
} from './encoding';

describe('base64url encoding utilities', () => {
  describe('base64urlEncode', () => {
    test('encodes empty Uint8Array', () => {
      const data = new Uint8Array([]);
      const encoded = base64urlEncode(data);
      expect(encoded).toBe('');
    });

    test('encodes simple binary data', () => {
      const data = new Uint8Array([0x12, 0xab, 0xcd]);
      const encoded = base64urlEncode(data);
      expect(encoded).toBeTruthy();
      expect(encoded).not.toContain('='); // No padding
      expect(encoded).not.toContain('+'); // No + character
      expect(encoded).not.toContain('/'); // No / character
    });

    test('encodes UTF-8 text data', () => {
      const text = 'Hello, UCAN!';
      const data = new TextEncoder().encode(text);
      const encoded = base64urlEncode(data);
      expect(encoded).toBe('SGVsbG8sIFVDQU4h');
    });

    test('produces URL-safe output', () => {
      // Create data that would produce + or / in standard base64
      const data = new Uint8Array([0xff, 0xfe, 0xfd]);
      const encoded = base64urlEncode(data);
      expect(encoded).toMatch(/^[A-Za-z0-9_-]+$/);
    });
  });

  describe('base64urlDecode', () => {
    test('decodes empty string', () => {
      const decoded = base64urlDecode('');
      expect(decoded).toEqual(new Uint8Array([]));
    });

    test('decodes unpadded base64url', () => {
      const encoded = 'SGVsbG8sIFVDQU4h';
      const decoded = base64urlDecode(encoded);
      const text = new TextDecoder().decode(decoded);
      expect(text).toBe('Hello, UCAN!');
    });

    test('decodes padded base64url', () => {
      // Use a proper padded base64url string (using standard base64url with padding)
      const encoded = 'SGk='; // "Hi" with padding
      const decoded = base64urlDecode(encoded);
      const text = new TextDecoder().decode(decoded);
      expect(text).toBe('Hi');
    });

    test('handles URL-safe characters', () => {
      // Encode data that produces - and _ characters
      const data = new Uint8Array([0xff, 0xfe, 0xfd]);
      const encoded = base64urlEncode(data);
      const decoded = base64urlDecode(encoded);
      expect(decoded).toEqual(data);
    });

    test('throws on invalid base64url', () => {
      expect(() => base64urlDecode('invalid base64url!!!')).toThrow();
    });

    test('roundtrip encoding/decoding', () => {
      const original = new Uint8Array([1, 2, 3, 4, 5, 6, 7, 8, 9, 10]);
      const encoded = base64urlEncode(original);
      const decoded = base64urlDecode(encoded);
      expect(decoded).toEqual(original);
    });
  });

  describe('base64urlEncodeJSON', () => {
    test('encodes simple object', () => {
      const obj = { foo: 'bar', baz: 42 };
      const encoded = base64urlEncodeJSON(obj);
      expect(encoded).toBeTruthy();
      expect(encoded).not.toContain('=');
    });

    test('encodes UCAN header', () => {
      const header = {
        alg: 'EdDSA',
        typ: 'JWT',
        ucv: '0.10.0',
      };
      const encoded = base64urlEncodeJSON(header);
      expect(encoded).toMatch(/^[A-Za-z0-9_-]+$/);
    });

    test('encodes nested objects', () => {
      const obj = {
        outer: {
          inner: {
            value: 'nested',
          },
        },
      };
      const encoded = base64urlEncodeJSON(obj);
      expect(encoded).toBeTruthy();
    });

    test('encodes arrays', () => {
      const arr = [1, 2, 3, 'four', { five: 5 }];
      const encoded = base64urlEncodeJSON(arr);
      expect(encoded).toBeTruthy();
    });
  });

  describe('base64urlDecodeJSON', () => {
    test('decodes simple object', () => {
      const original = { foo: 'bar', baz: 42 };
      const encoded = base64urlEncodeJSON(original);
      const decoded = base64urlDecodeJSON(encoded);
      expect(decoded).toEqual(original);
    });

    test('decodes UCAN header with type safety', () => {
      interface UCANHeader {
        alg: string;
        typ: string;
        ucv: string;
      }
      const original: UCANHeader = {
        alg: 'EdDSA',
        typ: 'JWT',
        ucv: '0.10.0',
      };
      const encoded = base64urlEncodeJSON(original);
      const decoded = base64urlDecodeJSON<UCANHeader>(encoded);
      expect(decoded.alg).toBe('EdDSA');
      expect(decoded.typ).toBe('JWT');
      expect(decoded.ucv).toBe('0.10.0');
    });

    test('roundtrip with nested objects', () => {
      const original = {
        outer: {
          inner: {
            value: 'nested',
            number: 123,
          },
        },
      };
      const encoded = base64urlEncodeJSON(original);
      const decoded = base64urlDecodeJSON(encoded);
      expect(decoded).toEqual(original);
    });

    test('throws on invalid JSON', () => {
      const invalidJSON = base64urlEncodeString('not valid json {');
      expect(() => base64urlDecodeJSON(invalidJSON)).toThrow();
    });
  });

  describe('base64urlEncodeString', () => {
    test('encodes empty string', () => {
      const encoded = base64urlEncodeString('');
      expect(encoded).toBe('');
    });

    test('encodes ASCII text', () => {
      const encoded = base64urlEncodeString('Hello, UCAN!');
      expect(encoded).toBe('SGVsbG8sIFVDQU4h');
    });

    test('encodes Unicode text', () => {
      const text = 'Hello, 世界! 🌍';
      const encoded = base64urlEncodeString(text);
      expect(encoded).toBeTruthy();
      expect(encoded).toMatch(/^[A-Za-z0-9_-]+$/);
    });

    test('encodes special characters', () => {
      const text = 'Special: @#$%^&*()';
      const encoded = base64urlEncodeString(text);
      expect(encoded).toBeTruthy();
    });
  });

  describe('base64urlDecodeString', () => {
    test('decodes empty string', () => {
      const decoded = base64urlDecodeString('');
      expect(decoded).toBe('');
    });

    test('decodes ASCII text', () => {
      const decoded = base64urlDecodeString('SGVsbG8sIFVDQU4h');
      expect(decoded).toBe('Hello, UCAN!');
    });

    test('roundtrip with Unicode', () => {
      const original = 'Hello, 世界! 🌍';
      const encoded = base64urlEncodeString(original);
      const decoded = base64urlDecodeString(encoded);
      expect(decoded).toBe(original);
    });

    test('roundtrip with special characters', () => {
      const original = 'Special: @#$%^&*()';
      const encoded = base64urlEncodeString(original);
      const decoded = base64urlDecodeString(encoded);
      expect(decoded).toBe(original);
    });
  });

  describe('isValidBase64url', () => {
    test('accepts valid base64url without padding', () => {
      expect(isValidBase64url('SGVsbG8')).toBe(true);
      expect(isValidBase64url('SGVsbG8sIFVDQU4h')).toBe(true);
    });

    test('accepts valid base64url with padding', () => {
      expect(isValidBase64url('SGVsbG8=')).toBe(true);
      expect(isValidBase64url('SGVsbG8==')).toBe(true);
    });

    test('accepts URL-safe characters', () => {
      expect(isValidBase64url('abc-def_ghi')).toBe(true);
      expect(isValidBase64url('ABC123-_')).toBe(true);
    });

    test('rejects invalid characters', () => {
      expect(isValidBase64url('hello+world')).toBe(false); // + not allowed
      expect(isValidBase64url('hello/world')).toBe(false); // / not allowed
      expect(isValidBase64url('hello world')).toBe(false); // space not allowed
      expect(isValidBase64url('hello!')).toBe(false); // ! not allowed
    });

    test('accepts empty string (valid for empty array)', () => {
      expect(isValidBase64url('')).toBe(true);
    });

    test('rejects too much padding', () => {
      expect(isValidBase64url('SGVsbG8===')).toBe(false);
    });

    test('accepts empty array encoding', () => {
      const encoded = base64urlEncode(new Uint8Array([]));
      expect(isValidBase64url(encoded)).toBe(true);
    });
  });

  describe('JWT compatibility', () => {
    test('encodes JWT header correctly', () => {
      const header = {
        alg: 'EdDSA',
        typ: 'JWT',
        ucv: '0.10.0',
      };
      const encoded = base64urlEncodeJSON(header);

      // JWT spec requires no padding
      expect(encoded).not.toContain('=');

      // Verify can decode back
      const decoded = base64urlDecodeJSON(encoded);
      expect(decoded).toEqual(header);
    });

    test('creates valid JWT token structure', () => {
      const header = { alg: 'EdDSA', typ: 'JWT', ucv: '0.10.0' };
      const payload = {
        iss: 'did:key:z6Mk...',
        aud: 'did:key:z6Mk...',
        exp: 1234567890,
        att: [],
      };
      const signature = new Uint8Array([1, 2, 3, 4, 5]);

      const encodedHeader = base64urlEncodeJSON(header);
      const encodedPayload = base64urlEncodeJSON(payload);
      const encodedSignature = base64urlEncode(signature);

      const jwt = `${encodedHeader}.${encodedPayload}.${encodedSignature}`;

      // Verify JWT structure
      const parts = jwt.split('.');
      expect(parts).toHaveLength(3);
      expect(parts.every((part) => isValidBase64url(part))).toBe(true);
    });
  });

  describe('edge cases', () => {
    test('handles large payloads', () => {
      const largeArray = new Uint8Array(10000);
      for (let i = 0; i < largeArray.length; i++) {
        largeArray[i] = i % 256;
      }

      const encoded = base64urlEncode(largeArray);
      const decoded = base64urlDecode(encoded);
      expect(decoded).toEqual(largeArray);
    });

    test('handles all byte values', () => {
      const allBytes = new Uint8Array(256);
      for (let i = 0; i < 256; i++) {
        allBytes[i] = i;
      }

      const encoded = base64urlEncode(allBytes);
      const decoded = base64urlDecode(encoded);
      expect(decoded).toEqual(allBytes);
    });

    test('handles deep JSON structures', () => {
      const deep = {
        level1: {
          level2: {
            level3: {
              level4: {
                level5: {
                  value: 'deep',
                  array: [1, 2, 3],
                  bool: true,
                  null: null,
                },
              },
            },
          },
        },
      };

      const encoded = base64urlEncodeJSON(deep);
      const decoded = base64urlDecodeJSON(encoded);
      expect(decoded).toEqual(deep);
    });
  });
});
