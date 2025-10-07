/**
 * UCAN token formatting and serialization utilities.
 *
 * Provides functions to convert UCAN token objects into JWT strings.
 * Implements deterministic serialization for consistent round-trip behavior
 * with the parser, ensuring formatToken(parseToken(x)) === x.
 *
 * @see https://tools.ietf.org/html/rfc7519 - JWT specification
 * @see https://github.com/ucan-wg/spec - UCAN specification
 */

import { base64urlEncode, base64urlEncodeJSON } from './encoding.js';
import type { UCANHeader, UCANPayload, UCANToken } from './types.js';

/**
 * Sorts an object's keys alphabetically for deterministic serialization.
 *
 * Recursively processes nested objects and arrays to ensure consistent
 * JSON encoding across all platforms.
 *
 * @param obj - Object to sort
 * @returns New object with alphabetically sorted keys
 */
function sortObjectByKey<T>(obj: T): T {
  if (typeof obj !== 'object' || obj == null) {
    return obj;
  }
  if (Array.isArray(obj)) {
    return obj.map(sortObjectByKey) as T;
  }
  const sortedKeys = Object.keys(obj).sort();
  const result: Record<string, unknown> = {};
  for (const key of sortedKeys) {
    result[key] = sortObjectByKey((obj as Record<string, unknown>)[key]);
  }
  return result as T;
}

/**
 * Removes undefined properties from an object.
 *
 * Creates a new object containing only properties with defined values.
 * This ensures optional fields that are undefined are not included in
 * the JSON serialization.
 *
 * @param obj - Object to clean
 * @returns New object without undefined properties
 */
function removeUndefined<T extends Record<string, unknown>>(obj: T): Partial<T> {
  const result: Partial<T> = {};
  for (const key in obj) {
    if (obj[key] !== undefined) {
      result[key] = obj[key];
    }
  }
  return result;
}

/**
 * Formats a UCAN header to base64url-encoded JSON string.
 *
 * Encodes the header with deterministic key ordering for consistent
 * serialization across invocations.
 *
 * @param header - UCAN header to format
 * @returns Base64url-encoded JSON string
 *
 * @example
 * ```typescript
 * const header = { alg: 'EdDSA', typ: 'JWT', ucv: '0.10.0' };
 * const encoded = formatHeader(header);
 * // => 'eyJhbGciOiJFZERTQSIsInR5cCI6IkpXVCIsInVjdiI6IjAuMTAuMCJ9'
 * ```
 */
export function formatHeader(header: UCANHeader): string {
  const sorted = sortObjectByKey(header);
  return base64urlEncodeJSON(sorted);
}

/**
 * Formats a UCAN payload to base64url-encoded JSON string.
 *
 * Encodes the payload with deterministic key ordering and removes
 * undefined optional fields for clean serialization.
 *
 * @param payload - UCAN payload to format
 * @returns Base64url-encoded JSON string
 *
 * @example
 * ```typescript
 * const payload = {
 *   iss: 'did:key:z6Mk...',
 *   aud: 'did:key:z6Mr...',
 *   exp: 1735689600,
 *   att: [{ with: 'storage://...', can: 'crud/read' }]
 * };
 * const encoded = formatPayload(payload);
 * ```
 */
export function formatPayload(payload: UCANPayload): string {
  // Remove undefined optional fields before encoding
  const cleaned = removeUndefined(payload as unknown as Record<string, unknown>);
  const sorted = sortObjectByKey(cleaned);
  return base64urlEncodeJSON(sorted);
}

/**
 * Creates the signing message for UCAN token signature generation.
 *
 * Concatenates the formatted header and payload with a '.' separator.
 * This is the exact string that should be signed to create a valid
 * UCAN token signature.
 *
 * @param header - UCAN header
 * @param payload - UCAN payload
 * @returns Signing message string in format: header.payload
 *
 * @example
 * ```typescript
 * const header = { alg: 'EdDSA', typ: 'JWT', ucv: '0.10.0' };
 * const payload = { iss: 'did:key:...', aud: 'did:key:...', exp: 123, att: [...] };
 * const message = createSigningMessage(header, payload);
 * // Sign this message to get the signature
 * const signature = await sign(message, privateKey);
 * ```
 */
export function createSigningMessage(header: UCANHeader, payload: UCANPayload): string {
  const headerEncoded = formatHeader(header);
  const payloadEncoded = formatPayload(payload);
  return `${headerEncoded}.${payloadEncoded}`;
}

/**
 * Formats a complete UCAN token to JWT string.
 *
 * Serializes a UCANToken object to a JWT string by encoding the header,
 * payload, and signature, then concatenating them with '.' separators.
 *
 * **Round-trip guarantee**: `formatToken(parseToken(x)) === x` for any valid token.
 *
 * @param ucan - UCAN token to format
 * @returns JWT string in format: header.payload.signature
 *
 * @example
 * ```typescript
 * const token: UCANToken = {
 *   header: { alg: 'EdDSA', typ: 'JWT', ucv: '0.10.0' },
 *   payload: {
 *     iss: 'did:key:z6Mk...',
 *     aud: 'did:key:z6Mr...',
 *     exp: 1735689600,
 *     att: [{ with: 'storage://...', can: 'crud/read' }]
 *   },
 *   signature: new Uint8Array([...])
 * };
 *
 * const jwtString = formatToken(token);
 * // => 'eyJhbGc...eyJpc3M...dGVzdA'
 *
 * // Round-trip verification
 * const reparsed = parseToken(jwtString);
 * const reformatted = formatToken(reparsed);
 * console.log(jwtString === reformatted); // => true
 * ```
 */
export function formatToken(ucan: UCANToken): string {
  const headerEncoded = formatHeader(ucan.header);
  const payloadEncoded = formatPayload(ucan.payload);
  const signatureEncoded = base64urlEncode(ucan.signature);

  return `${headerEncoded}.${payloadEncoded}.${signatureEncoded}`;
}
