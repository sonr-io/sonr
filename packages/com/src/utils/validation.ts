/**
 * Validation Utilities
 * Functions for validating common formats and patterns
 */

import { z } from 'zod';

/**
 * Validate email address
 */
export function isValidEmail(email: string): boolean {
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
  return emailRegex.test(email);
}

/**
 * Validate URL
 */
export function isValidUrl(url: string): boolean {
  try {
    new URL(url);
    return true;
  } catch {
    return false;
  }
}

/**
 * Validate domain
 */
export function isValidDomain(domain: string): boolean {
  const domainRegex =
    /^(?:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\.)*[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?$/i;
  return domainRegex.test(domain);
}

/**
 * Validate API key format
 */
export function isValidApiKey(key: string): boolean {
  const apiKeyRegex = /^sk_(test|live)_[a-zA-Z0-9]{24,}$/;
  return apiKeyRegex.test(key);
}

/**
 * Validate DID format
 */
export function isValidDID(did: string): boolean {
  // W3C DID format: did:method:method-specific-id
  const didRegex = /^did:[a-z0-9]+:[a-zA-Z0-9:.-]+$/;
  return didRegex.test(did);
}

/**
 * Validate phone number
 */
export function isValidPhoneNumber(phone: string): boolean {
  // Basic international format
  const phoneRegex = /^\+?[1-9]\d{1,14}$/;
  return phoneRegex.test(phone.replace(/[\s-()]/g, ''));
}

// ============================================================================
// Zod Validation Schemas
// ============================================================================

/**
 * Common validation schemas
 */
export const schemas = {
  /**
   * Email schema
   */
  email: z.string().email('Invalid email address'),

  /**
   * URL schema
   */
  url: z.string().url('Invalid URL'),

  /**
   * Domain schema
   */
  domain: z.string().refine(isValidDomain, 'Invalid domain'),

  /**
   * API key schema
   */
  apiKey: z.string().refine(isValidApiKey, 'Invalid API key format'),

  /**
   * DID schema
   */
  did: z.string().refine(isValidDID, 'Invalid DID format'),

  /**
   * Phone number schema
   */
  phoneNumber: z.string().refine(isValidPhoneNumber, 'Invalid phone number'),

  /**
   * Service registration schema
   */
  serviceRegistration: z.object({
    name: z.string().min(3, 'Name must be at least 3 characters').max(50),
    description: z.string().optional(),
    domain: z.string().refine(isValidDomain, 'Invalid domain').optional(),
    permissions: z.array(z.string()).optional(),
    tags: z.array(z.string()).max(10, 'Maximum 10 tags allowed').optional(),
  }),

  /**
   * Domain verification schema
   */
  domainVerification: z.object({
    domain: z.string().refine(isValidDomain, 'Invalid domain'),
    method: z.enum(['dns', 'http']),
    autoVerify: z.boolean().optional(),
  }),

  /**
   * User profile schema
   */
  userProfile: z.object({
    username: z
      .string()
      .min(3)
      .max(30)
      .regex(/^[a-zA-Z0-9_-]+$/, 'Username can only contain letters, numbers, - and _'),
    email: z.string().email().optional(),
    name: z.string().max(100).optional(),
    avatar: z.string().url().optional(),
  }),

  /**
   * Pagination schema
   */
  pagination: z.object({
    page: z.number().int().positive(),
    pageSize: z.number().int().positive().max(100),
  }),

  /**
   * Time range schema
   */
  timeRange: z.object({
    start: z.date(),
    end: z.date(),
    preset: z
      .enum(['today', 'yesterday', 'last7days', 'last30days', 'thisMonth', 'lastMonth', 'custom'])
      .optional(),
  }),

  /**
   * Permission request schema
   */
  permissionRequest: z.object({
    resource: z.string().min(1),
    action: z.string().min(1),
    effect: z.enum(['allow', 'deny']),
    conditions: z
      .array(
        z.object({
          field: z.string(),
          operator: z.enum(['equals', 'notEquals', 'contains', 'in', 'notIn']),
          value: z.any(),
        })
      )
      .optional(),
  }),
};

/**
 * Create a validator function from a Zod schema
 */
export function createValidator<T>(schema: z.ZodSchema<T>) {
  return (data: unknown): { success: boolean; data?: T; errors?: z.ZodError } => {
    const result = schema.safeParse(data);
    if (result.success) {
      return { success: true, data: result.data };
    }
    return { success: false, errors: result.error };
  };
}
