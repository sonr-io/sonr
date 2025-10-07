/**
 * UCAN Token Validation Examples
 *
 * Demonstrates comprehensive validation of UCAN tokens including
 * timestamp checking, DID validation, capability verification, and
 * signature validation.
 */

import type { UCANToken } from './types';
import {
  isTokenExpired,
  isTokenNotYetValid,
  validateAlgorithm,
  validateCapabilities,
  validateCapabilityAttenuation,
  validateDID,
  validateTimestamps,
  validateToken,
} from './validation';

// Example 1: Complete token validation
async function exampleCompleteValidation() {
  const token: UCANToken = {
    header: {
      alg: 'EdDSA',
      typ: 'JWT',
      ucv: '0.10.0',
    },
    payload: {
      iss: 'did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK',
      aud: 'did:key:z6MkrZ1r5XBFZjBU34qyD8fueMbMRkKw17BZaq2ivKFjnz2z',
      exp: Math.floor(Date.now() / 1000) + 3600, // 1 hour from now
      att: [
        {
          with: 'storage://did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK/photos',
          can: 'crud/read',
        },
      ],
    },
    signature: new Uint8Array(64), // Placeholder signature
  };

  // Validate the complete token
  const result = await validateToken(token, {
    clockDriftTolerance: 60,
    verifySignature: false, // Skip signature verification for this example
  });

  if (result.valid) {
    console.log('Token is valid!');
  } else {
    console.error('Token validation failed:', result.error);
  }
}

// Example 2: DID format validation
function exampleDIDValidation() {
  // Valid DIDs
  console.log(
    'Valid DID:',
    validateDID('did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK')
  );
  console.log('Valid DID:', validateDID('did:web:example.com'));
  console.log('Valid DID:', validateDID('did:sonr:abc123xyz'));

  // Invalid DIDs
  console.log('Invalid DID (bad format):', validateDID('not-a-did'));
  console.log('Invalid DID (too short):', validateDID('did:x:y'));
}

// Example 3: Timestamp validation
function exampleTimestampValidation() {
  const now = Math.floor(Date.now() / 1000);

  // Valid token (not expired, no nbf)
  const validPayload = {
    iss: 'did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK',
    aud: 'did:key:z6MkrZ1r5XBFZjBU34qyD8fueMbMRkKw17BZaq2ivKFjnz2z',
    exp: now + 3600, // Expires in 1 hour
    att: [],
  };
  console.log('Valid timestamps:', validateTimestamps(validPayload));

  // Expired token
  const expiredPayload = {
    ...validPayload,
    exp: now - 3600, // Expired 1 hour ago
  };
  console.log('Expired token:', validateTimestamps(expiredPayload));

  // Not yet valid token
  const notYetValidPayload = {
    ...validPayload,
    nbf: now + 3600, // Not valid until 1 hour from now
  };
  console.log('Not yet valid:', validateTimestamps(notYetValidPayload));

  // Never-expiring token
  const neverExpiringPayload = {
    ...validPayload,
    exp: null,
  };
  console.log('Never expires:', validateTimestamps(neverExpiringPayload));
}

// Example 4: Capability validation
function exampleCapabilityValidation() {
  // Valid capabilities
  const validCapabilities = [
    {
      with: 'storage://did:key:abc123/photos',
      can: 'crud/read',
    },
    {
      with: 'storage://did:key:abc123/docs',
      can: 'crud/write',
      nb: { maxSize: 1048576 },
    },
  ];
  console.log('Valid capabilities:', validateCapabilities(validCapabilities));

  // Invalid capabilities (missing 'with' field)
  const invalidCapabilities = [
    {
      can: 'crud/read',
    } as any,
  ];
  console.log('Invalid capabilities:', validateCapabilities(invalidCapabilities));

  // Empty capabilities array
  console.log('Empty capabilities:', validateCapabilities([]));
}

// Example 5: Algorithm validation
function exampleAlgorithmValidation() {
  const tokenEdDSA: UCANToken = {
    header: { alg: 'EdDSA', typ: 'JWT', ucv: '0.10.0' },
    payload: {
      iss: 'did:key:abc',
      aud: 'did:key:xyz',
      exp: null,
      att: [{ with: 'storage://*', can: 'crud/*' }],
    },
    signature: new Uint8Array(64),
  };
  console.log('EdDSA algorithm:', validateAlgorithm(tokenEdDSA));

  const tokenES256: UCANToken = {
    ...tokenEdDSA,
    header: { alg: 'ES256', typ: 'JWT', ucv: '0.10.0' },
  };
  console.log('ES256 algorithm:', validateAlgorithm(tokenES256));

  const tokenRS256: UCANToken = {
    ...tokenEdDSA,
    header: { alg: 'RS256', typ: 'JWT', ucv: '0.10.0' },
  };
  console.log('RS256 algorithm:', validateAlgorithm(tokenRS256));
}

// Example 6: Capability attenuation validation
function exampleCapabilityAttenuation() {
  // Valid attenuation (child is more restrictive)
  const parentCapability = {
    with: 'storage://*',
    can: 'crud/*',
  };
  const childCapability = {
    with: 'storage://photos',
    can: 'crud/read',
  };
  console.log(
    'Valid attenuation:',
    validateCapabilityAttenuation(childCapability, parentCapability)
  );

  // Invalid attenuation (child is more permissive)
  const invalidChildCapability = {
    with: 'different://resource',
    can: 'crud/*',
  };
  console.log(
    'Invalid attenuation:',
    validateCapabilityAttenuation(invalidChildCapability, parentCapability)
  );
}

// Example 7: Helper functions for token expiry checks
function exampleHelperFunctions() {
  const now = Math.floor(Date.now() / 1000);

  const payload = {
    iss: 'did:key:abc',
    aud: 'did:key:xyz',
    exp: now + 3600, // Expires in 1 hour
    nbf: now - 60, // Valid since 1 minute ago
    att: [],
  };

  // Check if token is expired
  console.log('Is token expired?', isTokenExpired(payload));

  // Check if token is not yet valid
  console.log('Is token not yet valid?', isTokenNotYetValid(payload));

  // Check expired token
  const expiredPayload = { ...payload, exp: now - 3600 };
  console.log('Is expired token expired?', isTokenExpired(expiredPayload));

  // Check not-yet-valid token
  const futurePayload = { ...payload, nbf: now + 3600 };
  console.log('Is future token not yet valid?', isTokenNotYetValid(futurePayload));
}

// Example 8: Validation with custom options
async function exampleCustomOptions() {
  const token: UCANToken = {
    header: { alg: 'EdDSA', typ: 'JWT', ucv: '0.10.0' },
    payload: {
      iss: 'did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK',
      aud: 'did:key:z6MkrZ1r5XBFZjBU34qyD8fueMbMRkKw17BZaq2ivKFjnz2z',
      exp: Math.floor(Date.now() / 1000) + 3600,
      att: [{ with: 'storage://*', can: 'crud/*' }],
    },
    signature: new Uint8Array(64),
  };

  // Validate with custom clock drift tolerance
  const result1 = await validateToken(token, {
    clockDriftTolerance: 120, // 2 minutes
    verifySignature: false,
  });
  console.log('Validation with custom clock drift:', result1);

  // Validate at a specific time (useful for testing)
  const futureTime = Math.floor(Date.now() / 1000) + 7200; // 2 hours from now
  const result2 = await validateToken(token, {
    now: futureTime,
    verifySignature: false,
  });
  console.log('Validation at future time (should fail):', result2);

  // Skip signature verification (default is true)
  const result3 = await validateToken(token, {
    verifySignature: false,
  });
  console.log('Validation without signature check:', result3);
}

// Run all examples
async function runAllExamples() {
  console.log('\n=== Example 1: Complete Token Validation ===');
  await exampleCompleteValidation();

  console.log('\n=== Example 2: DID Format Validation ===');
  exampleDIDValidation();

  console.log('\n=== Example 3: Timestamp Validation ===');
  exampleTimestampValidation();

  console.log('\n=== Example 4: Capability Validation ===');
  exampleCapabilityValidation();

  console.log('\n=== Example 5: Algorithm Validation ===');
  exampleAlgorithmValidation();

  console.log('\n=== Example 6: Capability Attenuation ===');
  exampleCapabilityAttenuation();

  console.log('\n=== Example 7: Helper Functions ===');
  exampleHelperFunctions();

  console.log('\n=== Example 8: Custom Validation Options ===');
  await exampleCustomOptions();
}

// Uncomment to run examples
// runAllExamples().catch(console.error);
