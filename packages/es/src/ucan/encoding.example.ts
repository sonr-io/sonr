/**
 * Example usage of UCAN base64url encoding utilities.
 *
 * This file demonstrates how to use the encoding functions
 * for UCAN JWT token formatting.
 */

import {
  base64urlDecode,
  base64urlDecodeJSON,
  base64urlDecodeString,
  base64urlEncode,
  base64urlEncodeJSON,
  base64urlEncodeString,
  isValidBase64url,
} from './encoding';
import type { UCANHeader, UCANPayload } from './types';

// Example 1: Encoding a UCAN header
export function encodeUCANHeader() {
  const header: UCANHeader = {
    alg: 'EdDSA',
    typ: 'JWT',
    ucv: '0.10.0',
  };

  const encoded = base64urlEncodeJSON(header);
  console.log('Encoded header:', encoded);

  return encoded;
}

// Example 2: Encoding a UCAN payload
export function encodeUCANPayload() {
  const payload: UCANPayload = {
    iss: 'did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK',
    aud: 'did:key:z6MkrZ1r5XBFZjBU34qyD8fueMbMRkKw17BZaq2ivKFjnz2z',
    exp: Math.floor(Date.now() / 1000) + 3600,
    att: [
      {
        with: 'storage://did:key:z6Mk.../photos',
        can: 'crud/read',
        nb: {
          maxSize: 1048576,
        },
      },
    ],
  };

  const encoded = base64urlEncodeJSON(payload);
  console.log('Encoded payload:', encoded);

  return encoded;
}

// Example 3: Creating a complete JWT token structure
export function createJWTToken(signature: Uint8Array) {
  const headerEncoded = encodeUCANHeader();
  const payloadEncoded = encodeUCANPayload();
  const signatureEncoded = base64urlEncode(signature);

  const jwt = headerEncoded + '.' + payloadEncoded + '.' + signatureEncoded;
  console.log('Complete JWT:', jwt);

  return jwt;
}

// Example 4: Parsing a JWT token
export function parseJWTToken(jwt: string) {
  const parts = jwt.split('.');

  if (parts.length !== 3) {
    throw new Error('Invalid JWT format');
  }

  const headerEncoded = parts[0];
  const payloadEncoded = parts[1];
  const signatureEncoded = parts[2];

  // Validate all parts
  if (
    !isValidBase64url(headerEncoded) ||
    !isValidBase64url(payloadEncoded) ||
    !isValidBase64url(signatureEncoded)
  ) {
    throw new Error('Invalid base64url encoding in JWT');
  }

  // Decode components
  const header = base64urlDecodeJSON<UCANHeader>(headerEncoded);
  const payload = base64urlDecodeJSON<UCANPayload>(payloadEncoded);
  const signature = base64urlDecode(signatureEncoded);

  console.log('Decoded header:', header);
  console.log('Decoded payload:', payload);
  console.log('Signature length:', signature.length);

  return { header, payload, signature };
}

// Example 5: String encoding/decoding
export function encodeDecodeStrings() {
  const message = 'Hello, UCAN!';

  const encoded = base64urlEncodeString(message);
  console.log('Encoded message:', encoded);

  const decoded = base64urlDecodeString(encoded);
  console.log('Decoded message:', decoded);

  console.assert(message === decoded, 'Roundtrip failed');
}

// Example 6: Binary data encoding
export function encodeBinaryData() {
  const binaryData = new Uint8Array([
    0x30, 0x44, 0x02, 0x20, 0x1a, 0x2b, 0x3c, 0x4d, 0x5e, 0x6f, 0x70, 0x81, 0x92, 0xa3, 0xb4, 0xc5,
    0xd6, 0xe7, 0xf8, 0x09,
  ]);

  const encoded = base64urlEncode(binaryData);
  console.log('Encoded binary:', encoded);

  const decoded = base64urlDecode(encoded);
  console.log('Decoded binary:', decoded);

  const match = binaryData.every((byte, index) => byte === decoded[index]);
  console.assert(match, 'Binary roundtrip failed');
}

// Example 7: Validation
export function validateBase64urlStrings() {
  const testCases = [
    { input: 'SGVsbG8', expected: true, description: 'valid unpadded' },
    { input: 'SGVsbG8=', expected: true, description: 'valid with padding' },
    { input: 'hello+world', expected: false, description: 'contains +' },
    { input: 'hello/world', expected: false, description: 'contains /' },
    { input: 'hello world', expected: false, description: 'contains space' },
    { input: '', expected: true, description: 'empty string (valid)' },
  ];

  for (const tc of testCases) {
    const result = isValidBase64url(tc.input);
    const status = result === tc.expected ? 'PASS' : 'FAIL';
    console.log(tc.description + ': ' + status);
  }
}
