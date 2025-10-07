/**
 * UCAN (User Controlled Authorization Network) module.
 *
 * Provides utilities for working with UCAN tokens including:
 * - Base64url encoding/decoding (RFC 4648 Section 5)
 * - Type definitions for UCAN tokens
 * - JWT token parsing and validation
 * - JWT-compatible token formatting
 *
 * @see https://github.com/ucan-wg/spec - UCAN specification
 */

// Export encoding utilities
export {
  base64urlDecode,
  base64urlDecodeJSON,
  base64urlDecodeString,
  base64urlEncode,
  base64urlEncodeJSON,
  base64urlEncodeString,
  isValidBase64url,
} from './encoding';

// Export parser utilities
export {
  extractHeader,
  extractPayload,
  isValidJWTFormat,
  parseToken,
  tryParseToken,
} from './parser';

// Export builder utilities
export { UCANBuilder } from './builder';

// Export validation utilities
export {
  isTokenExpired,
  isTokenNotYetValid,
  validateAlgorithm,
  validateCapabilities,
  validateCapabilityAttenuation,
  validateDID,
  validateSignature,
  validateTimestamps,
  validateToken,
} from './validation';

// Export token formatting utilities
export {
  createSigningMessage,
  formatHeader,
  formatPayload,
  formatToken,
} from './token';

// Export type definitions
export type {
  Capability,
  CapabilityComparison,
  Fact,
  PartialUCANPayload,
  TimeBounds,
  UCANAlgorithm,
  UCANBuilderOptions,
  UCANHeader,
  UCANPayload,
  UCANToken,
  UCANVersion,
  ValidationOptions,
  ValidationResult,
} from './types';

// Export type guards and enums
export {
  isCapability,
  isUCANAlgorithm,
  isUCANHeader,
  isUCANPayload,
  isUCANToken,
  UCANValidationError,
} from './types';
