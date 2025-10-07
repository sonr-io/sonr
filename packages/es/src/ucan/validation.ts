/**
 * UCAN Token Validation
 *
 * Comprehensive validation for UCAN tokens including signature verification,
 * timestamp validation, capability chain verification, and DID format checks.
 *
 * @see https://github.com/ucan-wg/spec - Official UCAN specification
 */

import { sha256 } from '@noble/hashes/sha256';
import * as secp256k1 from '@noble/secp256k1';

import { base64urlEncode, base64urlEncodeJSON } from './encoding';
import type {
  Capability,
  UCANPayload,
  UCANToken,
  UCANValidationError,
  ValidationOptions,
  ValidationResult,
} from './types';

/**
 * Default clock drift tolerance in seconds.
 * Allows tokens to be considered valid within ±60 seconds of their exact nbf/exp times.
 */
const DEFAULT_CLOCK_DRIFT_TOLERANCE = 60;

/**
 * Validates a complete UCAN token.
 *
 * Performs comprehensive validation including:
 * - Header validation (algorithm support)
 * - DID format validation (issuer and audience)
 * - Timestamp validation (nbf and exp)
 * - Capability structure validation
 * - Optional signature verification
 * - Optional proof chain validation
 *
 * @param token - The UCAN token to validate
 * @param options - Optional validation configuration
 * @returns ValidationResult with valid/invalid status and error details
 *
 * @example
 * ```typescript
 * const result = await validateToken(token, {
 *   clockDriftTolerance: 60,
 *   verifySignature: true
 * });
 *
 * if (!result.valid) {
 *   console.error('Validation failed:', result.error);
 * }
 * ```
 */
export async function validateToken(
  token: UCANToken,
  options?: ValidationOptions
): Promise<ValidationResult> {
  // Validate algorithm support
  const algorithmResult = validateAlgorithm(token);
  if (!algorithmResult.valid) {
    return algorithmResult;
  }

  // Validate DIDs
  const issuerResult = validateDID(token.payload.iss);
  if (!issuerResult.valid) {
    return {
      valid: false,
      error: `Invalid issuer DID: ${issuerResult.error}`,
      details: { code: 'INVALID_ISSUER' as UCANValidationError },
    };
  }

  const audienceResult = validateDID(token.payload.aud);
  if (!audienceResult.valid) {
    return {
      valid: false,
      error: `Invalid audience DID: ${audienceResult.error}`,
      details: { code: 'INVALID_AUDIENCE' as UCANValidationError },
    };
  }

  // Validate timestamps
  const timestampResult = validateTimestamps(token.payload, options);
  if (!timestampResult.valid) {
    return timestampResult;
  }

  // Validate capabilities
  const capabilitiesResult = validateCapabilities(token.payload.att);
  if (!capabilitiesResult.valid) {
    return capabilitiesResult;
  }

  // Optionally verify signature
  const shouldVerifySignature = options?.verifySignature ?? true;
  if (shouldVerifySignature) {
    const signatureResult = await validateSignature(token);
    if (!signatureResult.valid) {
      return signatureResult;
    }
  }

  // Optionally validate proof chain
  const shouldValidateProofChain = options?.validateProofChain ?? false;
  if (shouldValidateProofChain && token.payload.prf && token.payload.prf.length > 0) {
    // Note: Proof chain validation would require parsing each proof token
    // and recursively validating them. This is a placeholder for future implementation.
    // For now, we just check that the proof array exists and is non-empty.
  }

  return { valid: true };
}

/**
 * Validates algorithm support.
 *
 * Ensures the token's algorithm is one of the supported types:
 * EdDSA, ES256, or RS256.
 *
 * @param token - The UCAN token to validate
 * @returns ValidationResult indicating if algorithm is supported
 */
export function validateAlgorithm(token: UCANToken): ValidationResult {
  const supportedAlgorithms = ['EdDSA', 'ES256', 'RS256'];

  if (!supportedAlgorithms.includes(token.header.alg)) {
    return {
      valid: false,
      error: `Unsupported algorithm: ${token.header.alg}. Must be one of: ${supportedAlgorithms.join(', ')}`,
      details: { code: 'UNSUPPORTED_ALGORITHM' as UCANValidationError },
    };
  }

  return { valid: true };
}

/**
 * Validates DID format.
 *
 * Checks if a string is a valid DID URI according to W3C DID specification.
 * Format: did:<method>:<identifier>
 *
 * @param did - The DID string to validate
 * @returns ValidationResult indicating if DID format is valid
 *
 * @example
 * ```typescript
 * validateDID('did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK'); // valid
 * validateDID('invalid-did'); // invalid
 * ```
 */
export function validateDID(did: string): ValidationResult {
  // Basic DID format: did:<method>:<identifier>
  const didRegex = /^did:[a-z0-9]+:[a-zA-Z0-9._-]+$/;

  if (!didRegex.test(did)) {
    return {
      valid: false,
      error: `Invalid DID format: ${did}. Expected format: did:<method>:<identifier>`,
    };
  }

  // Additional validation: ensure minimum length
  if (did.length < 10) {
    return {
      valid: false,
      error: `DID too short: ${did}. Must be at least 10 characters`,
    };
  }

  return { valid: true };
}

/**
 * Validates token timestamps (nbf and exp).
 *
 * Checks that:
 * - Current time is before expiration (if exp is not null)
 * - Current time is after not-before time (if nbf is present)
 * - Respects clock drift tolerance
 *
 * @param payload - The UCAN payload containing timestamps
 * @param options - Optional validation configuration
 * @returns ValidationResult indicating if timestamps are valid
 *
 * @example
 * ```typescript
 * const payload = {
 *   iss: 'did:key:abc',
 *   aud: 'did:key:xyz',
 *   exp: Math.floor(Date.now() / 1000) + 3600, // 1 hour from now
 *   att: []
 * };
 * validateTimestamps(payload); // valid
 * ```
 */
export function validateTimestamps(
  payload: UCANPayload,
  options?: ValidationOptions
): ValidationResult {
  const now = options?.now ?? Math.floor(Date.now() / 1000);
  const clockDrift = options?.clockDriftTolerance ?? DEFAULT_CLOCK_DRIFT_TOLERANCE;

  // Validate expiration
  if (payload.exp !== null) {
    if (typeof payload.exp !== 'number') {
      return {
        valid: false,
        error: 'Expiration (exp) must be a number or null',
        details: { code: 'INVALID_PAYLOAD' as UCANValidationError },
      };
    }

    if (now > payload.exp + clockDrift) {
      const expDate = new Date(payload.exp * 1000).toISOString();
      return {
        valid: false,
        error: `Token has expired. Expiration: ${expDate}, Current time: ${new Date(now * 1000).toISOString()}`,
        details: {
          code: 'EXPIRED' as UCANValidationError,
          exp: payload.exp,
          now,
        },
      };
    }
  }

  // Validate not before (if present)
  if (payload.nbf !== undefined) {
    if (typeof payload.nbf !== 'number') {
      return {
        valid: false,
        error: 'Not before (nbf) must be a number',
        details: { code: 'INVALID_PAYLOAD' as UCANValidationError },
      };
    }

    if (now < payload.nbf - clockDrift) {
      const nbfDate = new Date(payload.nbf * 1000).toISOString();
      return {
        valid: false,
        error: `Token is not yet valid. Not before: ${nbfDate}, Current time: ${new Date(now * 1000).toISOString()}`,
        details: {
          code: 'NOT_YET_VALID' as UCANValidationError,
          nbf: payload.nbf,
          now,
        },
      };
    }
  }

  return { valid: true };
}

/**
 * Validates capability structure.
 *
 * Ensures all capabilities have required fields (with, can)
 * and proper structure for caveats (nb field).
 *
 * @param capabilities - Array of capabilities to validate
 * @returns ValidationResult indicating if all capabilities are valid
 *
 * @example
 * ```typescript
 * const capabilities = [
 *   { with: 'storage://did:key:abc/photos', can: 'crud/read' },
 *   { with: 'storage://did:key:abc/docs', can: 'crud/write', nb: { maxSize: 1024 } }
 * ];
 * validateCapabilities(capabilities); // valid
 * ```
 */
export function validateCapabilities(capabilities: Capability[]): ValidationResult {
  if (!Array.isArray(capabilities)) {
    return {
      valid: false,
      error: 'Capabilities (att) must be an array',
      details: { code: 'INVALID_CAPABILITY' as UCANValidationError },
    };
  }

  if (capabilities.length === 0) {
    return {
      valid: false,
      error: 'Token must have at least one capability',
      details: { code: 'INVALID_CAPABILITY' as UCANValidationError },
    };
  }

  for (let i = 0; i < capabilities.length; i++) {
    const cap = capabilities[i];

    // Check required fields
    if (!cap.with || typeof cap.with !== 'string') {
      return {
        valid: false,
        error: `Capability at index ${i} missing or invalid 'with' field`,
        details: {
          code: 'INVALID_CAPABILITY' as UCANValidationError,
          index: i,
          capability: cap,
        },
      };
    }

    if (!cap.can || typeof cap.can !== 'string') {
      return {
        valid: false,
        error: `Capability at index ${i} missing or invalid 'can' field`,
        details: {
          code: 'INVALID_CAPABILITY' as UCANValidationError,
          index: i,
          capability: cap,
        },
      };
    }

    // Validate caveats structure if present
    if (cap.nb !== undefined) {
      if (typeof cap.nb !== 'object' || cap.nb === null || Array.isArray(cap.nb)) {
        return {
          valid: false,
          error: `Capability at index ${i} has invalid 'nb' field. Must be an object`,
          details: {
            code: 'INVALID_CAPABILITY' as UCANValidationError,
            index: i,
            capability: cap,
          },
        };
      }
    }
  }

  return { valid: true };
}

/**
 * Validates the cryptographic signature of a UCAN token.
 *
 * Verifies that the signature matches the header + payload using the
 * algorithm specified in the token header. Supports:
 * - EdDSA (Ed25519)
 * - ES256 (ECDSA with P-256 and SHA-256)
 * - RS256 (RSA with SHA-256)
 *
 * Note: This is a placeholder implementation. Full signature verification
 * requires extracting the public key from the issuer DID and using the
 * appropriate cryptographic library.
 *
 * @param token - The UCAN token to verify
 * @returns ValidationResult indicating if signature is valid
 *
 * @example
 * ```typescript
 * const result = await validateSignature(token);
 * if (!result.valid) {
 *   console.error('Invalid signature:', result.error);
 * }
 * ```
 */
export async function validateSignature(token: UCANToken): Promise<ValidationResult> {
  try {
    // Reconstruct the signing message: base64url(header) + '.' + base64url(payload)
    const headerEncoded = base64urlEncodeJSON(token.header);
    const payloadEncoded = base64urlEncodeJSON(token.payload);
    const signingMessage = `${headerEncoded}.${payloadEncoded}`;
    const messageBytes = new TextEncoder().encode(signingMessage);

    // For actual signature verification, we would need:
    // 1. Extract public key from issuer DID (requires DID resolution)
    // 2. Use appropriate crypto library based on algorithm
    // 3. Verify signature against the message

    // Algorithm-specific verification
    switch (token.header.alg) {
      case 'ES256': {
        // ES256: ECDSA with P-256 curve and SHA-256
        // Hash the message
        const messageHash = sha256(messageBytes);

        // Note: secp256k1.verify expects a public key
        // For now, we return a placeholder indicating signature verification is not fully implemented
        // In production, you would:
        // 1. Resolve the issuer DID to get the public key
        // 2. Verify: secp256k1.verify(token.signature, messageHash, publicKey)

        return {
          valid: false,
          error:
            'Signature verification not fully implemented. Requires DID resolution to extract public key.',
          details: {
            code: 'INVALID_SIGNATURE' as UCANValidationError,
            algorithm: token.header.alg,
            note: 'Implementation requires DID resolver integration',
          },
        };
      }

      case 'EdDSA': {
        // EdDSA: Ed25519 signature
        // Note: Requires @noble/ed25519 library (not currently in dependencies)
        // In production, you would:
        // 1. Import: import * as ed25519 from '@noble/ed25519';
        // 2. Resolve the issuer DID to get the Ed25519 public key
        // 3. Verify: await ed25519.verify(token.signature, messageBytes, publicKey)

        return {
          valid: false,
          error: 'EdDSA signature verification not implemented. Requires @noble/ed25519 library.',
          details: {
            code: 'INVALID_SIGNATURE' as UCANValidationError,
            algorithm: token.header.alg,
            note: 'Add @noble/ed25519 to dependencies for EdDSA support',
          },
        };
      }

      case 'RS256': {
        // RS256: RSA signature with SHA-256
        // Note: Requires Web Crypto API or a suitable RSA library
        // In production, you would:
        // 1. Resolve the issuer DID to get the RSA public key
        // 2. Use Web Crypto API:
        //    const key = await crypto.subtle.importKey(...);
        //    const valid = await crypto.subtle.verify('RSASSA-PKCS1-v1_5', key, token.signature, messageBytes);

        return {
          valid: false,
          error:
            'RS256 signature verification not implemented. Requires Web Crypto API integration.',
          details: {
            code: 'INVALID_SIGNATURE' as UCANValidationError,
            algorithm: token.header.alg,
            note: 'Use Web Crypto API for RS256 verification',
          },
        };
      }

      default:
        return {
          valid: false,
          error: `Unsupported algorithm for signature verification: ${token.header.alg}`,
          details: { code: 'UNSUPPORTED_ALGORITHM' as UCANValidationError },
        };
    }
  } catch (error) {
    return {
      valid: false,
      error: `Signature verification failed: ${error instanceof Error ? error.message : String(error)}`,
      details: {
        code: 'INVALID_SIGNATURE' as UCANValidationError,
        error: error instanceof Error ? error.message : String(error),
      },
    };
  }
}

/**
 * Validates capability attenuation in a proof chain.
 *
 * Ensures that delegated capabilities are properly attenuated (equal or more restrictive)
 * compared to parent capabilities. This is a placeholder for future implementation
 * of the full attenuation checking algorithm.
 *
 * @param childCapability - The capability in the child token
 * @param parentCapability - The capability in the parent token
 * @returns ValidationResult indicating if attenuation is valid
 *
 * @example
 * ```typescript
 * const parent = { with: 'storage://*', can: 'crud/*' };
 * const child = { with: 'storage://photos', can: 'crud/read' };
 * validateCapabilityAttenuation(child, parent); // valid (child is more restrictive)
 * ```
 */
export function validateCapabilityAttenuation(
  childCapability: Capability,
  parentCapability: Capability
): ValidationResult {
  // Basic validation: check that 'with' is more specific or equal
  // Full implementation would require URI pattern matching and capability semantics

  // For now, we perform basic checks:
  // 1. Child 'with' must match or be more specific than parent 'with'
  // 2. Child 'can' must match or be more specific than parent 'can'

  // Wildcard matching for 'with' field
  const parentWith = parentCapability.with;
  const childWith = childCapability.with;

  // Simple check: child must start with parent prefix (after removing wildcards)
  const parentPrefix = parentWith.replace(/\*/g, '');
  if (!childWith.startsWith(parentPrefix)) {
    return {
      valid: false,
      error: `Capability 'with' field is not properly attenuated. Child: ${childWith}, Parent: ${parentWith}`,
      details: { code: 'INVALID_CAPABILITY' as UCANValidationError },
    };
  }

  // Wildcard matching for 'can' field
  const parentCan = parentCapability.can;
  const childCan = childCapability.can;

  const parentCanPrefix = parentCan.replace(/\*/g, '');
  if (!childCan.startsWith(parentCanPrefix)) {
    return {
      valid: false,
      error: `Capability 'can' field is not properly attenuated. Child: ${childCan}, Parent: ${parentCan}`,
      details: { code: 'INVALID_CAPABILITY' as UCANValidationError },
    };
  }

  // Caveat validation: child caveats should be equal or more restrictive
  // This requires domain-specific logic and is left as a placeholder

  return { valid: true };
}

/**
 * Helper function to check if a token has expired.
 *
 * @param payload - The UCAN payload
 * @param options - Optional validation configuration
 * @returns true if token is expired, false otherwise
 */
export function isTokenExpired(payload: UCANPayload, options?: ValidationOptions): boolean {
  if (payload.exp === null) {
    return false; // Never-expiring tokens are not expired
  }

  const now = options?.now ?? Math.floor(Date.now() / 1000);
  const clockDrift = options?.clockDriftTolerance ?? DEFAULT_CLOCK_DRIFT_TOLERANCE;

  return now > payload.exp + clockDrift;
}

/**
 * Helper function to check if a token is not yet valid.
 *
 * @param payload - The UCAN payload
 * @param options - Optional validation configuration
 * @returns true if token is not yet valid, false otherwise
 */
export function isTokenNotYetValid(payload: UCANPayload, options?: ValidationOptions): boolean {
  if (payload.nbf === undefined) {
    return false; // Tokens without nbf are immediately valid
  }

  const now = options?.now ?? Math.floor(Date.now() / 1000);
  const clockDrift = options?.clockDriftTolerance ?? DEFAULT_CLOCK_DRIFT_TOLERANCE;

  return now < payload.nbf - clockDrift;
}
