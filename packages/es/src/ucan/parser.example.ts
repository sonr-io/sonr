/**
 * Example usage of UCAN JWT parser
 *
 * This file demonstrates how to parse and work with UCAN JWT tokens.
 */

import { parseToken, tryParseToken, isValidJWTFormat, extractPayload } from './parser.js';
import { base64urlEncodeJSON } from './encoding.js';
import type { UCANHeader, UCANPayload } from './types.js';

/**
 * Example 1: Creating and parsing a basic UCAN token
 */
function example1_BasicParsing() {
  console.log('Example 1: Basic UCAN Token Parsing\n');

  // Create a sample UCAN token
  const header: UCANHeader = {
    alg: 'EdDSA',
    typ: 'JWT',
    ucv: '0.10.0',
  };

  const payload: UCANPayload = {
    iss: 'did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK',
    aud: 'did:key:z6MkrZ1r5XBFZjBU34qyD8fueMbMRkKw17BZaq2ivKFjnz2z',
    exp: Math.floor(Date.now() / 1000) + 3600, // Expires in 1 hour
    att: [
      {
        with: 'storage://did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK/photos',
        can: 'crud/read',
      },
      {
        with: 'storage://did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK/documents',
        can: 'crud/write',
        nb: { maxSize: 1048576 }, // 1 MB limit
      },
    ],
  };

  // Encode token (in production, signature would be cryptographically generated)
  const tokenString = `${base64urlEncodeJSON(header)}.${base64urlEncodeJSON(payload)}.mock-signature-bytes`;

  // Parse the token
  const token = parseToken(tokenString);

  console.log('Parsed UCAN Token:');
  console.log('Algorithm:', token.header.alg);
  console.log('Issuer:', token.payload.iss);
  console.log('Audience:', token.payload.aud);
  console.log('Expires:', new Date(token.payload.exp! * 1000).toISOString());
  console.log('Capabilities:', token.payload.att.length);
  console.log('\n');
}

/**
 * Example 2: Safe parsing with tryParseToken
 */
function example2_SafeParsing() {
  console.log('Example 2: Safe Parsing with tryParseToken\n');

  const validToken =
    'eyJhbGciOiJFZERTQSIsInR5cCI6IkpXVCIsInVjdiI6IjAuMTAuMCJ9.eyJpc3MiOiJkaWQ6a2V5OmFiYyIsImF1ZCI6ImRpZDprZXk6eHl6IiwiZXhwIjoxNzM1Njg5NjAwLCJhdHQiOlt7IndpdGgiOiJyZXNvdXJjZSIsImNhbiI6InJlYWQifV19.c2lnbmF0dXJl';
  const invalidToken = 'not.a.valid.token';

  // Try parsing valid token
  const result1 = tryParseToken(validToken);
  if (result1) {
    console.log('Valid token parsed successfully!');
    console.log('Issuer:', result1.payload.iss);
  } else {
    console.log('Failed to parse valid token');
  }

  // Try parsing invalid token - returns null instead of throwing
  const result2 = tryParseToken(invalidToken);
  if (result2) {
    console.log('Invalid token somehow parsed');
  } else {
    console.log('Invalid token correctly rejected (returned null)');
  }
  console.log('\n');
}

/**
 * Example 3: Pre-validation with isValidJWTFormat
 */
function example3_Prevalidation() {
  console.log('Example 3: Pre-validation with isValidJWTFormat\n');

  const testStrings = [
    'header.payload.signature', // Valid format
    'only.two.parts.extra', // Too many parts
    'only.two', // Too few parts
    'has spaces.in.parts', // Invalid characters
    'header+plus.payload.sig', // Invalid base64url
  ];

  for (const str of testStrings) {
    const isValid = isValidJWTFormat(str);
    console.log(`"${str}": ${isValid ? 'Valid format' : 'Invalid format'}`);
  }
  console.log('\n');
}

/**
 * Example 4: Quick payload inspection
 */
function example4_QuickInspection() {
  console.log('Example 4: Quick Payload Inspection\n');

  const header: UCANHeader = {
    alg: 'ES256',
    typ: 'JWT',
    ucv: '0.10.0',
  };

  const payload: UCANPayload = {
    iss: 'did:web:example.com',
    aud: 'did:web:service.com',
    exp: 1735689600,
    nbf: 1704067200,
    att: [
      {
        with: 'https://api.example.com/v1',
        can: 'api/invoke',
      },
    ],
  };

  const tokenString = `${base64urlEncodeJSON(header)}.${base64urlEncodeJSON(payload)}.signature`;

  // Quick inspection without full validation
  const extractedPayload = extractPayload(tokenString);
  console.log('Extracted payload (unvalidated):');
  console.log(JSON.stringify(extractedPayload, null, 2));
  console.log('\n');
}

/**
 * Example 5: Error handling
 */
function example5_ErrorHandling() {
  console.log('Example 5: Error Handling\n');

  const invalidTokens = [
    { token: 'not-a-jwt', error: 'Invalid format' },
    { token: 'invalid+base64.payload.signature', error: 'Invalid base64url' },
    { token: '', error: 'Empty string' },
  ];

  for (const { token, error } of invalidTokens) {
    try {
      parseToken(token);
      console.log(`${error}: Should have thrown, but didn't!`);
    } catch (e) {
      const message = e instanceof Error ? e.message : String(e);
      console.log(`${error}: Caught error - ${message.split('\n')[0]}`);
    }
  }
  console.log('\n');
}

/**
 * Example 6: Working with capabilities
 */
function example6_Capabilities() {
  console.log('Example 6: Working with Capabilities\n');

  const header: UCANHeader = {
    alg: 'EdDSA',
    typ: 'JWT',
    ucv: '0.10.0',
  };

  const payload: UCANPayload = {
    iss: 'did:key:issuer123',
    aud: 'did:key:audience456',
    exp: null, // Never expires
    att: [
      {
        with: 'storage://photos',
        can: 'crud/read',
      },
      {
        with: 'storage://photos',
        can: 'crud/write',
        nb: { maxFileSize: 5242880, allowedTypes: ['image/jpeg', 'image/png'] },
      },
      {
        with: 'mailto:user@example.com',
        can: 'msg/send',
        nb: { rateLimit: 100 },
      },
    ],
  };

  const tokenString = `${base64urlEncodeJSON(header)}.${base64urlEncodeJSON(payload)}.signature`;
  const token = parseToken(tokenString);

  console.log(`Token grants ${token.payload.att.length} capabilities:\n`);

  for (const cap of token.payload.att) {
    console.log(`- Resource: ${cap.with}`);
    console.log(`  Action: ${cap.can}`);
    if (cap.nb) {
      console.log(`  Constraints: ${JSON.stringify(cap.nb)}`);
    }
    console.log('');
  }
}

/**
 * Run all examples
 */
function runAllExamples() {
  example1_BasicParsing();
  example2_SafeParsing();
  example3_Prevalidation();
  example4_QuickInspection();
  example5_ErrorHandling();
  example6_Capabilities();
}

// Uncomment to run examples:
// runAllExamples();

// Export examples for use in documentation or testing
export {
  example1_BasicParsing,
  example2_SafeParsing,
  example3_Prevalidation,
  example4_QuickInspection,
  example5_ErrorHandling,
  example6_Capabilities,
  runAllExamples,
};
