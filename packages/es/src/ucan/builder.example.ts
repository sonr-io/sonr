/**
 * UCANBuilder usage examples.
 *
 * Demonstrates how to use the fluent builder API to construct UCAN tokens
 * with various configurations and optional fields.
 */

import { UCANBuilder } from './builder.js';

/**
 * Example 1: Basic UCAN token with required fields only.
 *
 * Creates a simple token with issuer, audience, expiration, and one capability.
 */
function basicExample(): void {
  const signature = new Uint8Array([1, 2, 3, 4, 5]); // Placeholder signature

  const token = new UCANBuilder()
    .issuer('did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK')
    .audience('did:key:z6MkrZ1r5XBFZjBU34qyD8fueMbMRkKw17BZaq2ivKFjnz2z')
    .expiresIn(3600) // 1 hour from now
    .addCapability({
      with: 'storage://did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK/photos',
      can: 'crud/read',
    })
    .build(signature);

  console.log('Basic UCAN token:', token);
}

/**
 * Example 2: UCAN token with all optional fields.
 *
 * Demonstrates usage of notBefore, nonce, facts, proofs, and caveats.
 */
function fullExample(): void {
  const signature = new Uint8Array([1, 2, 3, 4, 5]); // Placeholder signature
  const now = Math.floor(Date.now() / 1000);

  const token = new UCANBuilder()
    .issuer('did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK')
    .audience('did:key:z6MkrZ1r5XBFZjBU34qyD8fueMbMRkKw17BZaq2ivKFjnz2z')
    .expiration(now + 3600) // Absolute timestamp
    .notBefore(now) // Valid starting now
    .nonce('unique-nonce-123')
    .addCapability({
      with: 'storage://did:key:abc/photos',
      can: 'crud/write',
      nb: {
        maxSize: 1048576, // 1MB max file size
        fileTypes: ['image/jpeg', 'image/png'],
      },
    })
    .addCapability({
      with: 'mailto:user@example.com',
      can: 'msg/send',
    })
    .addFact({
      'user-type': 'premium',
      region: 'us-west-2',
    })
    .addProof('eyJhbGciOiJFZERTQSIsInR5cCI6IkpXVCJ9.eyJpc3MiOi...') // Parent UCAN
    .build(signature);

  console.log('Full UCAN token:', token);
}

/**
 * Example 3: Customizing UCAN version and algorithm.
 *
 * Shows how to override default version and signature algorithm.
 */
function customConfigExample(): void {
  const signature = new Uint8Array([1, 2, 3, 4, 5]); // Placeholder signature

  // Set custom defaults in constructor
  const token = new UCANBuilder({
    version: '0.10.0',
    algorithm: 'ES256',
  })
    .issuer('did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK')
    .audience('did:web:example.com')
    .expiresIn(86400) // 24 hours
    .addCapability({
      with: 'https://api.example.com/v1',
      can: 'api/invoke',
    })
    .build(signature);

  console.log('Custom config UCAN token:', token);
}

/**
 * Example 4: Building delegation chain with attenuation.
 *
 * Demonstrates creating a delegated token with reduced capabilities.
 */
function delegationExample(): void {
  const signature = new Uint8Array([1, 2, 3, 4, 5]); // Placeholder signature

  // Child token with attenuated (more restrictive) capabilities
  const delegatedToken = new UCANBuilder()
    .issuer('did:key:z6MkrZ1r5XBFZjBU34qyD8fueMbMRkKw17BZaq2ivKFjnz2z')
    .audience('did:key:z6MksomeOtherDID123456789')
    .expiresIn(1800) // 30 minutes (shorter than parent)
    .addCapability({
      with: 'storage://did:key:abc/photos/public', // More specific resource
      can: 'crud/read', // Read-only (parent had write)
      nb: {
        maxSize: 524288, // 512KB (smaller than parent's 1MB)
      },
    })
    .addProof('eyJhbGciOiJFZERTQSIsInR5cCI6IkpXVCJ9.eyJpc3MiOi...') // Parent UCAN
    .build(signature);

  console.log('Delegated UCAN token:', delegatedToken);
}

/**
 * Example 5: Error handling and validation.
 *
 * Shows validation errors when required fields are missing.
 */
function errorHandlingExample(): void {
  try {
    const signature = new Uint8Array([1, 2, 3, 4, 5]);

    // This will throw an error - missing required fields
    new UCANBuilder()
      .issuer('did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK')
      // Missing: audience, expiration, capabilities
      .build(signature);
  } catch (error) {
    console.error('Validation error:', error);
    // Expected: "UCAN validation failed: audience DID is required; expiration is required; at least one capability is required"
  }

  try {
    // Invalid DID format
    new UCANBuilder().issuer('not-a-valid-did');
  } catch (error) {
    console.error('DID validation error:', error);
    // Expected: "Invalid issuer DID format: 'not-a-valid-did'. Expected format: did:<method>:<identifier>"
  }

  try {
    const signature = new Uint8Array([1, 2, 3, 4, 5]);
    const now = Math.floor(Date.now() / 1000);

    // Invalid: notBefore after expiration
    new UCANBuilder()
      .issuer('did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK')
      .audience('did:key:z6MkrZ1r5XBFZjBU34qyD8fueMbMRkKw17BZaq2ivKFjnz2z')
      .notBefore(now + 3600)
      .expiration(now) // Expires before it becomes valid
      .addCapability({ with: 'resource', can: 'read' })
      .build(signature);
  } catch (error) {
    console.error('Timestamp validation error:', error);
    // Expected: "UCAN validation failed: notBefore (...) must be before expiration (...)"
  }
}

/**
 * Example 6: Method chaining with runtime configuration.
 *
 * Shows how to build tokens with conditional logic.
 */
function conditionalBuildingExample(isPremiumUser: boolean, includeProof: boolean): void {
  const signature = new Uint8Array([1, 2, 3, 4, 5]);

  let builder = new UCANBuilder()
    .issuer('did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK')
    .audience('did:key:z6MkrZ1r5XBFZjBU34qyD8fueMbMRkKw17BZaq2ivKFjnz2z')
    .expiresIn(3600)
    .addCapability({
      with: 'storage://did:key:abc/files',
      can: 'crud/read',
    });

  // Add premium capabilities
  if (isPremiumUser) {
    builder = builder
      .addCapability({
        with: 'storage://did:key:abc/files',
        can: 'crud/write',
      })
      .addFact({ 'user-type': 'premium' });
  }

  // Add proof if provided
  if (includeProof) {
    builder = builder.addProof('eyJhbGciOiJFZERTQSIsInR5cCI6IkpXVCJ9...');
  }

  const token = builder.build(signature);
  console.log('Conditional UCAN token:', token);
}

// Uncomment to run examples
// basicExample();
// fullExample();
// customConfigExample();
// delegationExample();
// errorHandlingExample();
// conditionalBuildingExample(true, false);
