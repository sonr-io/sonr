/**
 * UCAN JWT Token Parser
 *
 * Parses UCAN JWT tokens from string format into structured UCANToken objects.
 * Validates token structure, header, payload, and signature format according to
 * UCAN specification and JWT standards (RFC 7519).
 *
 * @see https://github.com/ucan-wg/spec - UCAN specification
 * @see https://tools.ietf.org/html/rfc7519 - JWT specification
 */

import { base64urlDecode, base64urlDecodeJSON } from './encoding.js';
import type { UCANToken } from './types.js';
import { isUCANHeader, isUCANPayload } from './types.js';

/**
 * Parses a UCAN JWT token string into a structured UCANToken object.
 *
 * Validates the token format, decodes base64url segments, and validates
 * header and payload structure using type guards. Does NOT verify the
 * cryptographic signature - use a separate verification function for that.
 *
 * @param token - JWT token string in format: header.payload.signature
 * @returns Parsed and validated UCANToken object
 * @throws {Error} If token format is invalid, segments cannot be decoded, or structure is invalid
 *
 * @example
 * ```typescript
 * const tokenString = 'eyJhbGc...eyJpc3M...abc123';
 * const token = parseToken(tokenString);
 * console.log(token.payload.iss); // => 'did:key:...'
 * console.log(token.header.alg); // => 'EdDSA'
 * ```
 */
export function parseToken(token: string): UCANToken {
  // Validate input
  if (typeof token !== 'string' || token.trim() === '') {
    throw new Error('UCAN token must be a non-empty string');
  }

  // Split JWT on '.' separator - must have exactly 3 parts
  const parts = token.split('.');
  if (parts.length !== 3) {
    throw new Error(
      `Invalid JWT format: expected 3 parts (header.payload.signature), got ${parts.length} parts`
    );
  }

  const [headerEncoded, payloadEncoded, signatureEncoded] = parts;

  // Validate that each part is non-empty
  if (!headerEncoded || !payloadEncoded || !signatureEncoded) {
    throw new Error(
      'Invalid JWT format: header, payload, and signature segments must be non-empty'
    );
  }

  // Decode header
  let header: unknown;
  try {
    header = base64urlDecodeJSON(headerEncoded);
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error);
    throw new Error(`Failed to decode JWT header: ${message}`);
  }

  // Validate header structure using type guard
  if (!isUCANHeader(header)) {
    throw new Error(
      `Invalid UCAN header: must contain 'alg' (EdDSA|ES256|RS256), 'typ' (JWT), and 'ucv' (version). Got: ${JSON.stringify(header)}`
    );
  }

  // Decode payload
  let payload: unknown;
  try {
    payload = base64urlDecodeJSON(payloadEncoded);
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error);
    throw new Error(`Failed to decode JWT payload: ${message}`);
  }

  // Validate payload structure using type guard
  if (!isUCANPayload(payload)) {
    const errors: string[] = [];

    if (typeof payload !== 'object' || payload === null) {
      errors.push('payload must be an object');
    } else {
      const p = payload as Record<string, unknown>;
      if (typeof p.iss !== 'string') errors.push("missing or invalid 'iss' (issuer DID)");
      if (typeof p.aud !== 'string') errors.push("missing or invalid 'aud' (audience DID)");
      if (p.exp !== null && typeof p.exp !== 'number') errors.push("invalid 'exp' (expiration)");
      if (!Array.isArray(p.att) || p.att.length === 0) {
        errors.push("missing or empty 'att' (capabilities array)");
      }
    }

    throw new Error(`Invalid UCAN payload: ${errors.join(', ')}. Got: ${JSON.stringify(payload)}`);
  }

  // Decode signature
  let signature: Uint8Array;
  try {
    signature = base64urlDecode(signatureEncoded);
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error);
    throw new Error(`Failed to decode JWT signature: ${message}`);
  }

  // Validate signature is not empty
  if (signature.length === 0) {
    throw new Error('Invalid JWT signature: signature cannot be empty');
  }

  return {
    header,
    payload,
    signature,
  };
}

/**
 * Safely attempts to parse a UCAN token, returning null on failure.
 *
 * Useful when you want to check if a string is a valid UCAN token
 * without throwing exceptions.
 *
 * @param token - JWT token string to parse
 * @returns Parsed UCANToken object or null if parsing fails
 *
 * @example
 * ```typescript
 * const token = tryParseToken(userInput);
 * if (token) {
 *   console.log('Valid UCAN token:', token.payload.iss);
 * } else {
 *   console.log('Invalid token format');
 * }
 * ```
 */
export function tryParseToken(token: string): UCANToken | null {
  try {
    return parseToken(token);
  } catch {
    return null;
  }
}

/**
 * Validates a JWT token string format without full parsing.
 *
 * Performs lightweight validation to check if a string looks like a valid JWT:
 * - Has exactly 3 parts separated by '.'
 * - Each part is non-empty and valid base64url
 *
 * Does NOT decode or validate header/payload contents.
 *
 * @param token - String to validate
 * @returns true if string has valid JWT format
 *
 * @example
 * ```typescript
 * if (isValidJWTFormat(input)) {
 *   const token = parseToken(input);
 *   // ... use token
 * }
 * ```
 */
export function isValidJWTFormat(token: string): boolean {
  if (typeof token !== 'string' || token.trim() === '') {
    return false;
  }

  const parts = token.split('.');
  if (parts.length !== 3) {
    return false;
  }

  // Check that each part is non-empty and looks like base64url
  const base64urlRegex = /^[A-Za-z0-9_-]+$/;
  return parts.every((part) => part.length > 0 && base64urlRegex.test(part));
}

/**
 * Extracts the raw payload JSON from a JWT token without full validation.
 *
 * Useful for quickly inspecting token contents or debugging.
 * Does NOT validate the payload structure.
 *
 * @param token - JWT token string
 * @returns Decoded payload as unknown type (needs validation)
 * @throws {Error} If token format is invalid or payload cannot be decoded
 *
 * @example
 * ```typescript
 * const payload = extractPayload(tokenString);
 * console.log('Issuer:', payload.iss);
 * console.log('Audience:', payload.aud);
 * ```
 */
export function extractPayload(token: string): unknown {
  const parts = token.split('.');
  if (parts.length !== 3) {
    throw new Error('Invalid JWT format: expected 3 parts');
  }

  try {
    return base64urlDecodeJSON(parts[1]);
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error);
    throw new Error(`Failed to decode payload: ${message}`);
  }
}

/**
 * Extracts the raw header JSON from a JWT token without full validation.
 *
 * Useful for quickly inspecting token algorithm or debugging.
 * Does NOT validate the header structure.
 *
 * @param token - JWT token string
 * @returns Decoded header as unknown type (needs validation)
 * @throws {Error} If token format is invalid or header cannot be decoded
 *
 * @example
 * ```typescript
 * const header = extractHeader(tokenString);
 * console.log('Algorithm:', header.alg);
 * console.log('Version:', header.ucv);
 * ```
 */
export function extractHeader(token: string): unknown {
  const parts = token.split('.');
  if (parts.length !== 3) {
    throw new Error('Invalid JWT format: expected 3 parts');
  }

  try {
    return base64urlDecodeJSON(parts[0]);
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error);
    throw new Error(`Failed to decode header: ${message}`);
  }
}
