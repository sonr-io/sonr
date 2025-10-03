/**
 * Auth & Permission Types
 * Types for authentication, authorization, and UCAN tokens
 */

import type { ID, Timestamp } from './common';

/**
 * Permission entity
 */
export interface Permission {
  id: ID;
  resource: string;
  action: string;
  effect: 'allow' | 'deny';
  conditions?: PermissionCondition[];
  description?: string;
}

/**
 * Permission condition
 */
export interface PermissionCondition {
  field: string;
  operator: 'equals' | 'notEquals' | 'contains' | 'in' | 'notIn';
  value: any;
}

/**
 * Permission template for quick setup
 */
export interface PermissionTemplate {
  id: ID;
  name: string;
  description: string;
  permissions: Permission[];
  category: 'basic' | 'standard' | 'advanced' | 'custom';
}

/**
 * UCAN token capability
 */
export interface UCANCapability {
  with: string;
  can: string;
  nb?: Record<string, any>;
}

/**
 * UCAN token structure
 */
export interface UCANToken {
  iss: string;
  aud: string;
  exp?: number;
  nbf?: number;
  nnc?: string;
  att: UCANCapability[];
  prf?: string[];
  fct?: Record<string, any>;
}

/**
 * Audit log entry
 */
export interface AuditLogEntry {
  id: ID;
  timestamp: Timestamp;
  actor: string;
  action: string;
  resource: string;
  result: 'success' | 'failure';
  details?: Record<string, any>;
  ip?: string;
  userAgent?: string;
}

/**
 * Permission request data
 */
export interface PermissionRequestData {
  requester: string;
  resource: string;
  actions: string[];
  reason?: string;
  duration?: number;
  conditions?: PermissionCondition[];
}
