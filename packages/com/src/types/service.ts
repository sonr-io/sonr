/**
 * Service-related Types
 * Types for Sonr Services (x/svc module)
 */

import type { ID, Status, Timestamp } from './common';
import type { Permission } from './permission';

/**
 * Service entity
 */
export interface Service {
  id: ID;
  name: string;
  description?: string;
  domain: string;
  status: 'active' | 'inactive' | 'pending';
  apiKey: string;
  createdAt: Timestamp;
  updatedAt: Timestamp;
  permissions?: Permission[];
  metrics?: ServiceMetrics;
  owner: string;
  tags?: string[];
  endpoints?: ServiceEndpoint[];
  lastActive?: Timestamp;
  domainVerificationStatus?: 'verified' | 'pending' | 'failed' | 'unverified';
}

/**
 * Service endpoint configuration
 */
export interface ServiceEndpoint {
  url: string;
  method: 'GET' | 'POST' | 'PUT' | 'DELETE' | 'PATCH';
  description?: string;
  authenticated: boolean;
  rateLimit?: number;
}

/**
 * Service metrics
 */
export interface ServiceMetrics {
  requests: number;
  errors: number;
  latencyP50: number;
  latencyP95: number;
  latencyP99: number;
  uptime: number;
  lastActivity?: Timestamp;
}

/**
 * Service creation/update request
 */
export interface ServiceRequest {
  name: string;
  description?: string;
  domain?: string;
  permissions?: string[];
  tags?: string[];
  endpoints?: ServiceEndpoint[];
}

/**
 * Service status
 */
export type ServiceStatus = 'active' | 'inactive' | 'suspended' | 'pending';

/**
 * Service category
 */
export type ServiceCategory = 'api' | 'webapp' | 'mobile' | 'iot' | 'blockchain' | 'other';
