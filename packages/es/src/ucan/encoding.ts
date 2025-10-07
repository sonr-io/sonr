/**
 * Base64url encoding utilities for UCAN JWT tokens.
 *
 * Implements RFC 4648 Section 5 compliant base64url encoding/decoding
 * for use in UCAN token formatting. Uses URL-safe character set and
 * handles padding appropriately for JWT compatibility.
 *
 * @see https://tools.ietf.org/html/rfc4648#section-5 - Base64url encoding
 * @see https://tools.ietf.org/html/rfc7515#appendix-C - JWT base64url notes
 */

import { base64url, base64urlnopad } from '@scure/base';

/**
 * Encodes a Uint8Array to base64url string (URL-safe, no padding).
 *
 * This is the standard encoding for JWT tokens as per RFC 7515.
 * Uses URL-safe alphabet: '-' and '_' instead of '+' and '/'.
 * Omits padding ('=') characters for compactness.
 *
 * @param data - Binary data to encode
 * @returns Base64url encoded string without padding
 *
 * @example
 * ```typescript
 * const data = new TextEncoder().encode('Hello, UCAN!');
 * const encoded = base64urlEncode(data);
 * console.log(encoded); // => 'SGVsbG8sIFVDQU4h'
 * ```
 */
export function base64urlEncode(data: Uint8Array): string {
  return base64urlnopad.encode(data);
}

/**
 * Decodes a base64url string to Uint8Array.
 *
 * Handles both padded and unpadded base64url strings.
 * Accepts URL-safe alphabet: '-' and '_' are valid characters.
 *
 * @param encoded - Base64url encoded string (with or without padding)
 * @returns Decoded binary data
 * @throws {Error} If the input string is invalid base64url
 *
 * @example
 * ```typescript
 * const decoded = base64urlDecode('SGVsbG8sIFVDQU4h');
 * const text = new TextDecoder().decode(decoded);
 * console.log(text); // => 'Hello, UCAN!'
 * ```
 */
export function base64urlDecode(encoded: string): Uint8Array {
  try {
    // First try decoding without padding (most common for JWTs)
    return base64urlnopad.decode(encoded);
  } catch {
    // Fallback to padded decoder if the string has padding
    return base64url.decode(encoded);
  }
}

/**
 * Encodes a JavaScript object to base64url JSON string.
 *
 * Convenience method for encoding JWT header/payload objects.
 * Serializes to JSON then encodes to base64url.
 *
 * @param obj - JavaScript object to encode
 * @returns Base64url encoded JSON string
 *
 * @example
 * ```typescript
 * const header = { alg: 'EdDSA', typ: 'JWT', ucv: '0.10.0' };
 * const encoded = base64urlEncodeJSON(header);
 * // Can be used directly in JWT token
 * ```
 */
export function base64urlEncodeJSON(obj: unknown): string {
  const json = JSON.stringify(obj);
  const bytes = new TextEncoder().encode(json);
  return base64urlEncode(bytes);
}

/**
 * Decodes a base64url string to a JavaScript object.
 *
 * Convenience method for decoding JWT header/payload segments.
 * Decodes base64url then parses as JSON.
 *
 * @param encoded - Base64url encoded JSON string
 * @returns Parsed JavaScript object
 * @throws {Error} If the input is invalid base64url or invalid JSON
 *
 * @example
 * ```typescript
 * const header = base64urlDecodeJSON(encodedHeader);
 * console.log(header.alg); // => 'EdDSA'
 * ```
 */
export function base64urlDecodeJSON<T = unknown>(encoded: string): T {
  const bytes = base64urlDecode(encoded);
  const json = new TextDecoder().decode(bytes);
  return JSON.parse(json) as T;
}

/**
 * Validates if a string is valid base64url format.
 *
 * Checks if the string contains only valid base64url characters.
 * Valid characters: A-Z, a-z, 0-9, '-', '_', and optionally '=' for padding.
 *
 * @param str - String to validate
 * @returns true if valid base64url format
 *
 * @example
 * ```typescript
 * isValidBase64url('SGVsbG8'); // => true
 * isValidBase64url('Hello+World'); // => false (contains '+')
 * isValidBase64url('Hello World'); // => false (contains space)
 * ```
 */
export function isValidBase64url(str: string): boolean {
  // Base64url regex: alphanumeric, hyphen, underscore, optional padding
  // Allow empty string (valid encoding of empty Uint8Array)
  const base64urlRegex = /^[A-Za-z0-9_-]*(={0,2})?$/;
  return base64urlRegex.test(str);
}

/**
 * Encodes a UTF-8 string to base64url.
 *
 * Convenience method for encoding text strings directly.
 *
 * @param str - UTF-8 string to encode
 * @returns Base64url encoded string
 *
 * @example
 * ```typescript
 * const encoded = base64urlEncodeString('Hello, UCAN!');
 * console.log(encoded); // => 'SGVsbG8sIFVDQU4h'
 * ```
 */
export function base64urlEncodeString(str: string): string {
  const bytes = new TextEncoder().encode(str);
  return base64urlEncode(bytes);
}

/**
 * Decodes a base64url string to UTF-8 string.
 *
 * Convenience method for decoding to text strings.
 *
 * @param encoded - Base64url encoded string
 * @returns Decoded UTF-8 string
 * @throws {Error} If the input is invalid base64url
 *
 * @example
 * ```typescript
 * const decoded = base64urlDecodeString('SGVsbG8sIFVDQU4h');
 * console.log(decoded); // => 'Hello, UCAN!'
 * ```
 */
export function base64urlDecodeString(encoded: string): string {
  const bytes = base64urlDecode(encoded);
  return new TextDecoder().decode(bytes);
}
