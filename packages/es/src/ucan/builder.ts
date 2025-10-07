/**
 * UCAN Token Builder
 *
 * Provides a fluent API for constructing UCAN JWT tokens with comprehensive
 * validation and type safety. Supports method chaining for ergonomic token
 * creation and validates all inputs before token assembly.
 *
 * @see https://github.com/ucan-wg/spec - UCAN specification
 */

import { base64urlEncode, base64urlEncodeJSON } from './encoding.js';
import type {
  Capability,
  Fact,
  UCANAlgorithm,
  UCANBuilderOptions,
  UCANHeader,
  UCANPayload,
  UCANVersion,
} from './types.js';
import { isCapability } from './types.js';

/**
 * Maximum safe integer for JavaScript timestamps (2^53 - 1).
 * Ensures timestamps remain precise in all JavaScript environments.
 */
const MAX_SAFE_TIMESTAMP = Number.MAX_SAFE_INTEGER;

/**
 * Minimum safe integer for JavaScript timestamps (-2^53 + 1).
 */
const MIN_SAFE_TIMESTAMP = Number.MIN_SAFE_INTEGER;

/**
 * Validates a DID string format.
 *
 * Performs basic DID URI validation according to W3C DID specification.
 * Checks for 'did:' prefix and method:identifier structure.
 *
 * @param did - DID string to validate
 * @returns true if DID format is valid
 */
function isValidDID(did: string): boolean {
  if (typeof did !== 'string' || did.length === 0) {
    return false;
  }

  // DID format: did:<method>:<method-specific-id>
  // Method: lowercase letters, digits
  // Method-specific-id: any characters except whitespace
  const didRegex = /^did:[a-z0-9]+:[^\s]+$/;
  return didRegex.test(did);
}

/**
 * Validates a timestamp is within safe JavaScript integer range.
 *
 * UCAN timestamps must be in seconds since Unix epoch (UTC) and
 * within the range [-2^53 - 1, 2^53 - 1] for precision.
 *
 * @param timestamp - Timestamp in seconds
 * @returns true if timestamp is valid
 */
function isValidTimestamp(timestamp: number): boolean {
  return (
    typeof timestamp === 'number' &&
    Number.isInteger(timestamp) &&
    timestamp >= MIN_SAFE_TIMESTAMP &&
    timestamp <= MAX_SAFE_TIMESTAMP
  );
}

/**
 * Fluent builder for constructing UCAN tokens.
 *
 * Provides a chainable API for setting token properties with validation
 * at each step. Ensures all required fields are set before token assembly.
 *
 * @example
 * ```typescript
 * const token = new UCANBuilder()
 *   .issuer('did:key:z6Mk...')
 *   .audience('did:key:z6Mk...')
 *   .expiresIn(3600)
 *   .addCapability({ with: 'storage://...', can: 'crud/read' })
 *   .build(signature);
 * ```
 */
export class UCANBuilder {
  private _version: UCANVersion;
  private _algorithm: UCANAlgorithm;
  private _issuer?: string;
  private _audience?: string;
  private _expiration?: number | null;
  private _notBefore?: number;
  private _nonce?: string;
  private _facts: Fact[] = [];
  private _proofs: string[] = [];
  private _capabilities: Capability[] = [];

  /**
   * Creates a new UCAN token builder.
   *
   * @param options - Optional configuration for UCAN version and algorithm
   */
  constructor(options?: UCANBuilderOptions) {
    this._version = options?.version ?? '0.10.0';
    this._algorithm = options?.algorithm ?? 'EdDSA';
  }

  /**
   * Sets the issuer DID (the entity creating and signing the token).
   *
   * @param did - Issuer DID (e.g., 'did:key:z6Mk...')
   * @returns this builder for chaining
   * @throws {Error} If DID format is invalid
   */
  issuer(did: string): this {
    if (!isValidDID(did)) {
      throw new Error(
        `Invalid issuer DID format: '${did}'. Expected format: did:<method>:<identifier>`
      );
    }
    this._issuer = did;
    return this;
  }

  /**
   * Sets the audience DID (the entity that can use this token).
   *
   * @param did - Audience DID (e.g., 'did:key:z6Mk...')
   * @returns this builder for chaining
   * @throws {Error} If DID format is invalid
   */
  audience(did: string): this {
    if (!isValidDID(did)) {
      throw new Error(
        `Invalid audience DID format: '${did}'. Expected format: did:<method>:<identifier>`
      );
    }
    this._audience = did;
    return this;
  }

  /**
   * Sets the expiration time (Unix timestamp in seconds).
   *
   * Token will be invalid after this time. Set to null for never-expiring
   * tokens (not recommended for security).
   *
   * @param timestamp - Expiration time in seconds since Unix epoch, or null
   * @returns this builder for chaining
   * @throws {Error} If timestamp is invalid or out of safe range
   */
  expiration(timestamp: number | null): this {
    if (timestamp !== null && !isValidTimestamp(timestamp)) {
      throw new Error(
        `Invalid expiration timestamp: ${timestamp}. Must be integer in range [${MIN_SAFE_TIMESTAMP}, ${MAX_SAFE_TIMESTAMP}]`
      );
    }
    this._expiration = timestamp;
    return this;
  }

  /**
   * Sets the not-before time (Unix timestamp in seconds).
   *
   * Token will be invalid before this time. Optional field for
   * time-delayed token activation.
   *
   * @param timestamp - Not-before time in seconds since Unix epoch
   * @returns this builder for chaining
   * @throws {Error} If timestamp is invalid or out of safe range
   */
  notBefore(timestamp: number): this {
    if (!isValidTimestamp(timestamp)) {
      throw new Error(
        `Invalid notBefore timestamp: ${timestamp}. Must be integer in range [${MIN_SAFE_TIMESTAMP}, ${MAX_SAFE_TIMESTAMP}]`
      );
    }
    this._notBefore = timestamp;
    return this;
  }

  /**
   * Helper method to set expiration relative to current time.
   *
   * Convenient alternative to setting absolute expiration timestamps.
   *
   * @param seconds - Number of seconds from now until expiration
   * @returns this builder for chaining
   * @throws {Error} If seconds is not a positive number
   *
   * @example
   * ```typescript
   * builder.expiresIn(3600); // Expires in 1 hour
   * builder.expiresIn(86400); // Expires in 1 day
   * ```
   */
  expiresIn(seconds: number): this {
    if (typeof seconds !== 'number' || seconds <= 0 || !Number.isFinite(seconds)) {
      throw new Error(`Invalid expiresIn value: ${seconds}. Must be a positive number`);
    }

    const now = Math.floor(Date.now() / 1000);
    const expiration = now + Math.floor(seconds);

    if (!isValidTimestamp(expiration)) {
      throw new Error(
        `Calculated expiration ${expiration} is out of safe range. Reduce expiresIn value.`
      );
    }

    this._expiration = expiration;
    return this;
  }

  /**
   * Sets the nonce for uniqueness and replay protection.
   *
   * Optional field recommended for single-use tokens.
   *
   * @param nonce - Unique nonce string
   * @returns this builder for chaining
   * @throws {Error} If nonce is empty
   */
  nonce(nonce: string): this {
    if (typeof nonce !== 'string' || nonce.length === 0) {
      throw new Error('Nonce must be a non-empty string');
    }
    this._nonce = nonce;
    return this;
  }

  /**
   * Sets the signature algorithm.
   *
   * Overrides the default algorithm set in constructor options.
   *
   * @param alg - Signature algorithm (EdDSA, ES256, or RS256)
   * @returns this builder for chaining
   */
  algorithm(alg: UCANAlgorithm): this {
    this._algorithm = alg;
    return this;
  }

  /**
   * Sets the UCAN version.
   *
   * Overrides the default version set in constructor options.
   *
   * @param version - UCAN specification version
   * @returns this builder for chaining
   */
  version(version: UCANVersion): this {
    this._version = version;
    return this;
  }

  /**
   * Adds a capability (permission) to the token.
   *
   * Capabilities define what actions can be performed on specific resources.
   * At least one capability is required for a valid UCAN token.
   *
   * @param capability - Capability with resource (with) and action (can)
   * @returns this builder for chaining
   * @throws {Error} If capability structure is invalid
   *
   * @example
   * ```typescript
   * builder.addCapability({
   *   with: 'storage://did:key:abc/photos',
   *   can: 'crud/read',
   *   nb: { maxSize: 1048576 }
   * });
   * ```
   */
  addCapability(capability: Capability): this {
    if (!isCapability(capability)) {
      throw new Error(
        `Invalid capability structure. Must have 'with' (string) and 'can' (string) fields. Got: ${JSON.stringify(capability)}`
      );
    }
    this._capabilities.push(capability);
    return this;
  }

  /**
   * Adds a fact (assertion) to the token.
   *
   * Facts are signed statements that provide additional context for
   * capability evaluation. Optional field.
   *
   * @param fact - Fact object with key-value pairs
   * @returns this builder for chaining
   * @throws {Error} If fact is not a valid object
   *
   * @example
   * ```typescript
   * builder.addFact({ 'user-type': 'premium', 'region': 'us-west' });
   * ```
   */
  addFact(fact: Fact): this {
    if (typeof fact !== 'object' || fact === null || Array.isArray(fact)) {
      throw new Error('Fact must be a non-null object');
    }
    this._facts.push(fact);
    return this;
  }

  /**
   * Adds a proof (parent UCAN token) to the proof chain.
   *
   * Proofs establish delegation chains where capabilities can be
   * attenuated (reduced) from parent tokens.
   *
   * @param ucan - Complete parent UCAN token (JWT string)
   * @returns this builder for chaining
   * @throws {Error} If UCAN token is not a non-empty string
   *
   * @example
   * ```typescript
   * builder.addProof('eyJhbGc...'); // Parent UCAN token
   * ```
   */
  addProof(ucan: string): this {
    if (typeof ucan !== 'string' || ucan.length === 0) {
      throw new Error('Proof UCAN must be a non-empty string');
    }
    this._proofs.push(ucan);
    return this;
  }

  /**
   * Validates that all required fields are set.
   *
   * @throws {Error} If any required field is missing or invalid
   */
  private _validate(): void {
    const errors: string[] = [];

    if (!this._issuer) {
      errors.push('issuer DID is required (use .issuer() method)');
    }

    if (!this._audience) {
      errors.push('audience DID is required (use .audience() method)');
    }

    if (this._expiration === undefined) {
      errors.push('expiration is required (use .expiration() or .expiresIn() method)');
    }

    if (this._capabilities.length === 0) {
      errors.push('at least one capability is required (use .addCapability() method)');
    }

    // Validate notBefore is before expiration if both are set
    if (
      this._notBefore !== undefined &&
      this._expiration !== null &&
      this._expiration !== undefined &&
      this._notBefore >= this._expiration
    ) {
      errors.push(`notBefore (${this._notBefore}) must be before expiration (${this._expiration})`);
    }

    if (errors.length > 0) {
      throw new Error(`UCAN validation failed: ${errors.join('; ')}`);
    }
  }

  /**
   * Assembles and encodes the UCAN token.
   *
   * Validates all required fields, constructs header and payload,
   * encodes them with the signature, and returns the complete JWT string.
   *
   * @param signature - Cryptographic signature as raw bytes
   * @returns Complete UCAN JWT token string (header.payload.signature)
   * @throws {Error} If validation fails or signature is invalid
   *
   * @example
   * ```typescript
   * const signature = new Uint8Array([...]); // From signing operation
   * const token = builder.build(signature);
   * console.log(token); // => 'eyJhbGc...eyJpc3M...abc123'
   * ```
   */
  build(signature: Uint8Array): string {
    // Validate all required fields are set
    this._validate();

    // Validate signature
    if (!(signature instanceof Uint8Array)) {
      throw new Error('Signature must be a Uint8Array');
    }
    if (signature.length === 0) {
      throw new Error('Signature cannot be empty');
    }

    // Construct header
    const header: UCANHeader = {
      alg: this._algorithm,
      typ: 'JWT',
      ucv: this._version,
    };

    // Construct payload with required fields
    const payload: UCANPayload = {
      iss: this._issuer!,
      aud: this._audience!,
      exp: this._expiration!,
      att: this._capabilities,
    };

    // Add optional fields if set
    if (this._notBefore !== undefined) {
      payload.nbf = this._notBefore;
    }
    if (this._nonce !== undefined) {
      payload.nnc = this._nonce;
    }
    if (this._facts.length > 0) {
      payload.fct = this._facts;
    }
    if (this._proofs.length > 0) {
      payload.prf = this._proofs;
    }

    // Encode header and payload
    const headerEncoded = base64urlEncodeJSON(header);
    const payloadEncoded = base64urlEncodeJSON(payload);
    const signatureEncoded = base64urlEncode(signature);

    // Assemble JWT: header.payload.signature
    return `${headerEncoded}.${payloadEncoded}.${signatureEncoded}`;
  }
}
