/**
 * Examples of UCAN token formatting and serialization.
 *
 * This module demonstrates how to use the token formatting utilities
 * to serialize UCAN tokens to JWT strings.
 */

import { parseToken } from './parser.js';
import {
  createSigningMessage,
  formatHeader,
  formatPayload,
  formatToken,
} from './token.js';
import type { UCANHeader, UCANPayload, UCANToken } from './types.js';

/**
 * Example 1: Format a complete UCAN token to JWT string
 */
export function formatCompleteToken() {
  const token: UCANToken = {
    header: {
      alg: 'EdDSA',
      typ: 'JWT',
      ucv: '0.10.0',
    },
    payload: {
      iss: 'did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK',
      aud: 'did:key:z6MkrZ1r5XBFZjBU34qyD8fueMbMRkKw17BZaq2ivKFjnz2z',
      exp: Math.floor(Date.now() / 1000) + 3600, // Expires in 1 hour
      att: [
        {
          with: 'storage://did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK/photos',
          can: 'crud/read',
        },
      ],
    },
    signature: new Uint8Array([1, 2, 3, 4, 5]), // Mock signature
  };

  const jwtString = formatToken(token);
  console.log('Formatted UCAN token:', jwtString);

  return jwtString;
}

/**
 * Example 2: Create a signing message for signature generation
 */
export function createSigningMessageExample() {
  const header: UCANHeader = {
    alg: 'EdDSA',
    typ: 'JWT',
    ucv: '0.10.0',
  };

  const payload: UCANPayload = {
    iss: 'did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK',
    aud: 'did:key:z6MkrZ1r5XBFZjBU34qyD8fueMbMRkKw17BZaq2ivKFjnz2z',
    exp: Math.floor(Date.now() / 1000) + 3600,
    att: [
      {
        with: 'storage://bucket/private',
        can: 'crud/write',
        nb: { maxSize: 1048576 },
      },
    ],
  };

  // Create the message that should be signed
  const signingMessage = createSigningMessage(header, payload);
  console.log('Message to sign:', signingMessage);

  // In a real application, you would sign this message:
  // const signature = await sign(signingMessage, privateKey);
  // const token: UCANToken = { header, payload, signature };
  // const jwt = formatToken(token);

  return signingMessage;
}

/**
 * Example 3: Round-trip formatting and parsing
 */
export function roundTripExample() {
  const originalToken: UCANToken = {
    header: {
      alg: 'ES256',
      typ: 'JWT',
      ucv: '0.10.0',
    },
    payload: {
      iss: 'did:web:example.com',
      aud: 'did:web:service.com',
      exp: 1735689600,
      nbf: 1704067200,
      nnc: 'unique-nonce-abc123',
      att: [
        {
          with: 'mailto:user@example.com',
          can: 'msg/send',
          nb: { maxRecipients: 10 },
        },
      ],
    },
    signature: new Uint8Array([10, 20, 30, 40, 50]),
  };

  // Format to JWT string
  const jwtString = formatToken(originalToken);
  console.log('Original JWT:', jwtString);

  // Parse back to token object
  const parsedToken = parseToken(jwtString);

  // Format again
  const reformattedJwt = formatToken(parsedToken);
  console.log('Reformatted JWT:', reformattedJwt);

  // Verify round-trip guarantee
  const isIdentical = jwtString === reformattedJwt;
  console.log('Round-trip successful:', isIdentical); // Should be true

  return { original: jwtString, reformatted: reformattedJwt, isIdentical };
}

/**
 * Example 4: Format token with delegation chain (proof chain)
 */
export function formatTokenWithProofChain() {
  // First, a parent token grants broad permissions
  const parentToken: UCANToken = {
    header: { alg: 'EdDSA', typ: 'JWT', ucv: '0.10.0' },
    payload: {
      iss: 'did:key:z6MkrootAuthority',
      aud: 'did:key:z6MkintermediarySevice',
      exp: Math.floor(Date.now() / 1000) + 86400, // 24 hours
      att: [
        {
          with: 'storage://*',
          can: 'crud/*',
        },
      ],
    },
    signature: new Uint8Array([100, 101, 102]),
  };

  const parentJwt = formatToken(parentToken);
  console.log('Parent UCAN:', parentJwt);

  // Then, create a delegated token with attenuated (reduced) permissions
  const delegatedToken: UCANToken = {
    header: { alg: 'EdDSA', typ: 'JWT', ucv: '0.10.0' },
    payload: {
      iss: 'did:key:z6MkintermediarySevice',
      aud: 'did:key:z6MkendUser',
      exp: Math.floor(Date.now() / 1000) + 3600, // 1 hour (less than parent)
      prf: [parentJwt], // Reference to parent token
      att: [
        {
          with: 'storage://specific-bucket', // More specific than parent
          can: 'crud/read', // More restrictive than parent
        },
      ],
    },
    signature: new Uint8Array([200, 201, 202]),
  };

  const delegatedJwt = formatToken(delegatedToken);
  console.log('Delegated UCAN:', delegatedJwt);

  return { parentJwt, delegatedJwt };
}

/**
 * Example 5: Format token with facts (assertions)
 */
export function formatTokenWithFacts() {
  const token: UCANToken = {
    header: { alg: 'EdDSA', typ: 'JWT', ucv: '0.10.0' },
    payload: {
      iss: 'did:key:z6Mkservice',
      aud: 'did:key:z6Mkuser',
      exp: Math.floor(Date.now() / 1000) + 3600,
      fct: [
        {
          'user-type': 'premium',
          'subscription-tier': 'gold',
          'region': 'us-west-2',
        },
      ],
      att: [
        {
          with: 'api://premium-features',
          can: 'api/access',
        },
      ],
    },
    signature: new Uint8Array([50, 51, 52]),
  };

  const jwtString = formatToken(token);
  console.log('UCAN with facts:', jwtString);

  return jwtString;
}

/**
 * Example 6: Format header and payload separately
 */
export function formatSeparateComponents() {
  const header: UCANHeader = {
    alg: 'RS256',
    typ: 'JWT',
    ucv: '0.10.0',
  };

  const payload: UCANPayload = {
    iss: 'did:key:z6Mkissuer',
    aud: 'did:key:z6Mkaudience',
    exp: null, // Never expires
    att: [
      {
        with: 'https://api.example.com',
        can: 'api/call',
      },
    ],
  };

  // Format components separately
  const headerEncoded = formatHeader(header);
  const payloadEncoded = formatPayload(payload);

  console.log('Encoded header:', headerEncoded);
  console.log('Encoded payload:', payloadEncoded);

  return { headerEncoded, payloadEncoded };
}

/**
 * Example 7: Deterministic serialization
 */
export function demonstrateDeterministicSerialization() {
  // Create two payloads with same content but different property order
  const payload1: UCANPayload = {
    iss: 'did:key:abc',
    aud: 'did:key:xyz',
    exp: 123456,
    att: [{ with: 'resource', can: 'read' }],
  };

  const payload2: UCANPayload = {
    att: [{ with: 'resource', can: 'read' }],
    exp: 123456,
    aud: 'did:key:xyz',
    iss: 'did:key:abc',
  };

  const encoded1 = formatPayload(payload1);
  const encoded2 = formatPayload(payload2);

  console.log('Payload 1 encoded:', encoded1);
  console.log('Payload 2 encoded:', encoded2);
  console.log('Encodings are identical:', encoded1 === encoded2); // true

  return { encoded1, encoded2, areEqual: encoded1 === encoded2 };
}

// Run examples if this file is executed directly
if (import.meta.url === `file://${process.argv[1]}`) {
  console.log('\n=== Example 1: Format Complete Token ===');
  formatCompleteToken();

  console.log('\n=== Example 2: Create Signing Message ===');
  createSigningMessageExample();

  console.log('\n=== Example 3: Round-trip Formatting ===');
  roundTripExample();

  console.log('\n=== Example 4: Token with Proof Chain ===');
  formatTokenWithProofChain();

  console.log('\n=== Example 5: Token with Facts ===');
  formatTokenWithFacts();

  console.log('\n=== Example 6: Format Separate Components ===');
  formatSeparateComponents();

  console.log('\n=== Example 7: Deterministic Serialization ===');
  demonstrateDeterministicSerialization();
}
