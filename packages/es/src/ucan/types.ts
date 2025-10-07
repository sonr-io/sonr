/**
 * UCAN (User Controlled Authorization Network) Type Definitions
 *
 * UCAN tokens are JWT-based capability tokens that enable decentralized
 * authorization without backend infrastructure. This module defines the
 * complete type system for UCAN token structure, validation, and operations.
 *
 * @see https://github.com/ucan-wg/spec - Official UCAN specification
 * @see https://tools.ietf.org/html/rfc7519 - JWT specification
 */

/**
 * Supported cryptographic algorithms for UCAN token signatures.
 *
 * - EdDSA: Edwards-curve Digital Signature Algorithm (Ed25519)
 * - ES256: ECDSA using P-256 curve and SHA-256
 * - RS256: RSA signature with SHA-256
 */
export type UCANAlgorithm = 'EdDSA' | 'ES256' | 'RS256';

/**
 * UCAN version identifier. Current specification is version "0.10.0".
 */
export type UCANVersion = '0.10.0' | string;

/**
 * UCAN token header following JWT standard with UCAN-specific extensions.
 *
 * The header contains metadata about the token's signature algorithm and format.
 * Must be base64url encoded as the first segment of the JWT.
 */
export interface UCANHeader {
  /** Algorithm used to sign the token */
  alg: UCANAlgorithm;

  /** Token type, must be 'JWT' for UCAN tokens */
  typ: 'JWT';

  /** UCAN specification version */
  ucv: UCANVersion;
}

/**
 * Resource capability with optional caveats (constraints).
 *
 * Capabilities define what actions can be performed on specific resources.
 * Caveats allow fine-grained control over when and how capabilities can be exercised.
 *
 * @example
 * ```typescript
 * const capability: Capability = {
 *   with: 'storage://did:key:abc123/photos',
 *   can: 'crud/read',
 *   nb: {
 *     maxSize: 1048576,
 *     fileTypes: ['image/jpeg', 'image/png']
 *   }
 * };
 * ```
 */
export interface Capability {
  /** Resource identifier (URI) this capability applies to */
  with: string;

  /** Action or permission granted (e.g., 'crud/read', 'msg/send') */
  can: string;

  /**
   * Caveats (constraints) on the capability.
   * Key-value pairs that limit when/how the capability can be used.
   */
  nb?: Record<string, unknown>;
}

/**
 * Fact assertion for including additional context in UCAN tokens.
 *
 * Facts are signed statements that can provide additional information
 * needed to evaluate capabilities. They are not capabilities themselves.
 *
 * @example
 * ```typescript
 * const fact: Fact = {
 *   'user-type': 'premium',
 *   'region': 'us-west-2'
 * };
 * ```
 */
export type Fact = Record<string, unknown>;

/**
 * UCAN token payload containing authorization claims and capabilities.
 *
 * The payload is the core of the UCAN token, defining who issued it,
 * who can use it, what capabilities it grants, and when it's valid.
 */
export interface UCANPayload {
  /**
   * Issuer DID - the entity that created and signed this token.
   * Must be a valid DID URI (e.g., 'did:key:z6Mk...')
   */
  iss: string;

  /**
   * Audience DID - the entity that can use this token.
   * Must be a valid DID URI.
   */
  aud: string;

  /**
   * Expiration time in seconds since Unix epoch (UTC).
   * Token is invalid after this time.
   * Range: -2^53 - 1 to 2^53 - 1
   *
   * Can be null for never-expiring tokens (not recommended).
   */
  exp: number | null;

  /**
   * Not before time in seconds since Unix epoch (UTC).
   * Token is invalid before this time.
   * Optional - if omitted, token is valid immediately upon issuance.
   */
  nbf?: number;

  /**
   * Nonce for uniqueness and replay protection.
   * Optional - recommended for single-use tokens.
   */
  nnc?: string;

  /**
   * Facts - additional assertions included in the token.
   * Optional contextual information that may be needed for capability evaluation.
   */
  fct?: Fact[];

  /**
   * Proof chain - array of parent UCAN tokens (as JWT strings).
   * Each entry is a complete UCAN token that grants capabilities to the issuer.
   * Enables delegation chains where capabilities can be attenuated.
   */
  prf?: string[];

  /**
   * Attenuations/capabilities - permissions granted by this token.
   * Must contain at least one capability.
   */
  att: Capability[];
}

/**
 * Complete UCAN token structure with decoded components.
 *
 * Represents a parsed UCAN token with header, payload, and signature
 * separated and decoded. This is the primary type for working with
 * UCAN tokens in application code.
 */
export interface UCANToken {
  /** Decoded token header */
  header: UCANHeader;

  /** Decoded token payload */
  payload: UCANPayload;

  /** Raw signature bytes */
  signature: Uint8Array;
}

/**
 * Result of UCAN token validation.
 *
 * Provides detailed information about validation success or failure,
 * including specific error messages for debugging.
 */
export interface ValidationResult {
  /** Whether the token passed all validation checks */
  valid: boolean;

  /**
   * Error message if validation failed.
   * Undefined if validation succeeded.
   */
  error?: string;

  /**
   * Additional context about validation errors.
   * May include details about which check failed and why.
   */
  details?: Record<string, unknown>;
}

/**
 * Options for UCAN token validation.
 *
 * Allows customization of validation behavior for different use cases.
 */
export interface ValidationOptions {
  /**
   * Current time for timestamp validation (seconds since Unix epoch).
   * If not provided, uses current system time.
   */
  now?: number;

  /**
   * Clock drift tolerance in seconds.
   * Allows tokens to be considered valid within this window before/after
   * their exact nbf/exp times to account for clock skew.
   *
   * @default 60
   */
  clockDriftTolerance?: number;

  /**
   * Whether to verify the signature cryptographically.
   * If false, only structural and timestamp validation is performed.
   *
   * @default true
   */
  verifySignature?: boolean;

  /**
   * Whether to validate the proof chain recursively.
   * If true, all parent UCANs in the proof chain are validated.
   *
   * @default false
   */
  validateProofChain?: boolean;
}

/**
 * Capability comparison result for attenuation checking.
 *
 * Used to determine if a delegated capability is properly attenuated
 * (equal or more restrictive) compared to the parent capability.
 */
export type CapabilityComparison = 'equal' | 'attenuated' | 'expanded' | 'unrelated';

/**
 * UCAN builder configuration options.
 */
export interface UCANBuilderOptions {
  /** UCAN version to use. Defaults to '0.10.0' */
  version?: UCANVersion;

  /** Algorithm for signature. Defaults to 'EdDSA' */
  algorithm?: UCANAlgorithm;
}

/**
 * Time bounds for UCAN token validity.
 */
export interface TimeBounds {
  /** Not before time (optional) */
  notBefore?: number;

  /** Expiration time (required) */
  expiration: number;
}

/**
 * Type guard to check if a value is a valid UCAN algorithm.
 */
export function isUCANAlgorithm(value: unknown): value is UCANAlgorithm {
  return value === 'EdDSA' || value === 'ES256' || value === 'RS256';
}

/**
 * Type guard to check if a value is a valid UCAN header.
 */
export function isUCANHeader(value: unknown): value is UCANHeader {
  if (typeof value !== 'object' || value === null) {
    return false;
  }

  const header = value as Partial<UCANHeader>;
  return isUCANAlgorithm(header.alg) && header.typ === 'JWT' && typeof header.ucv === 'string';
}

/**
 * Type guard to check if a value is a valid capability.
 */
export function isCapability(value: unknown): value is Capability {
  if (typeof value !== 'object' || value === null) {
    return false;
  }

  const cap = value as Partial<Capability>;
  return (
    typeof cap.with === 'string' &&
    typeof cap.can === 'string' &&
    (cap.nb === undefined || (typeof cap.nb === 'object' && cap.nb !== null))
  );
}

/**
 * Type guard to check if a value is a valid UCAN payload.
 */
export function isUCANPayload(value: unknown): value is UCANPayload {
  if (typeof value !== 'object' || value === null) {
    return false;
  }

  const payload = value as Partial<UCANPayload>;

  // Check required fields
  if (
    typeof payload.iss !== 'string' ||
    typeof payload.aud !== 'string' ||
    (payload.exp !== null && typeof payload.exp !== 'number') ||
    !Array.isArray(payload.att) ||
    payload.att.length === 0
  ) {
    return false;
  }

  // Validate all capabilities
  if (!payload.att.every(isCapability)) {
    return false;
  }

  // Check optional fields if present
  if (payload.nbf !== undefined && typeof payload.nbf !== 'number') {
    return false;
  }

  if (payload.nnc !== undefined && typeof payload.nnc !== 'string') {
    return false;
  }

  if (payload.fct !== undefined && !Array.isArray(payload.fct)) {
    return false;
  }

  if (
    payload.prf !== undefined &&
    (!Array.isArray(payload.prf) || !payload.prf.every((p) => typeof p === 'string'))
  ) {
    return false;
  }

  return true;
}

/**
 * Type guard to check if a value is a valid UCAN token.
 */
export function isUCANToken(value: unknown): value is UCANToken {
  if (typeof value !== 'object' || value === null) {
    return false;
  }

  const token = value as Partial<UCANToken>;
  return (
    isUCANHeader(token.header) &&
    isUCANPayload(token.payload) &&
    token.signature instanceof Uint8Array
  );
}

/**
 * Utility type for partial UCAN payload during construction.
 */
export type PartialUCANPayload = Partial<UCANPayload> & {
  iss?: string;
  aud?: string;
  exp?: number | null;
  att?: Capability[];
};

/**
 * Error codes for UCAN validation failures.
 */
export enum UCANValidationError {
  INVALID_FORMAT = 'INVALID_FORMAT',
  INVALID_HEADER = 'INVALID_HEADER',
  INVALID_PAYLOAD = 'INVALID_PAYLOAD',
  INVALID_SIGNATURE = 'INVALID_SIGNATURE',
  EXPIRED = 'EXPIRED',
  NOT_YET_VALID = 'NOT_YET_VALID',
  INVALID_ISSUER = 'INVALID_ISSUER',
  INVALID_AUDIENCE = 'INVALID_AUDIENCE',
  INVALID_CAPABILITY = 'INVALID_CAPABILITY',
  INVALID_PROOF_CHAIN = 'INVALID_PROOF_CHAIN',
  UNSUPPORTED_ALGORITHM = 'UNSUPPORTED_ALGORITHM',
}
