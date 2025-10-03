import { z } from 'zod';

/**
 * Service status enum
 */
export enum ServiceStatus {
  ACTIVE = 'active',
  PENDING = 'pending',
  SUSPENDED = 'suspended',
  INACTIVE = 'inactive',
}

/**
 * Domain verification status enum
 */
export enum DomainVerificationStatus {
  UNVERIFIED = 'unverified',
  PENDING = 'pending',
  VERIFIED = 'verified',
  FAILED = 'failed',
  EXPIRED = 'expired',
}

/**
 * Permission scope enum
 */
export enum PermissionScope {
  DATA = 'data',
  VAULT = 'vault',
  PROFILE = 'profile',
  SERVICE = 'service',
  ADMIN = 'admin',
}

/**
 * Service capability interface
 */
export interface ServiceCapability {
  id: string;
  name: string;
  description: string;
  scope: PermissionScope | string;
  granted: boolean;
  expiresAt?: string;
  constraints?: Record<string, any>;
}

/**
 * Service permission interface
 */
export interface ServicePermission {
  id: string;
  name: string;
  description: string;
  scope: string;
  granted: boolean;
  grantedAt?: string;
  grantedBy?: string;
  revokedAt?: string;
  revokedBy?: string;
}

/**
 * Service API key interface
 */
export interface ServiceApiKey {
  id: string;
  name: string;
  key?: string; // Only returned on creation
  prefix?: string; // Key prefix for identification
  createdAt: string;
  lastUsed?: string;
  expiresAt?: string;
  status: 'active' | 'expired' | 'revoked';
  permissions?: string[];
}

/**
 * Service metadata interface
 */
export interface ServiceMetadata {
  totalRequests?: number;
  activeUsers?: number;
  averageLatency?: number;
  errorRate?: number;
  uptime?: number;
  lastHealthCheck?: string;
  version?: string;
  environment?: string;
  [key: string]: any;
}

/**
 * Service domain interface
 */
export interface ServiceDomain {
  domain: string;
  verificationStatus: DomainVerificationStatus;
  verifiedAt?: string;
  txtRecord?: string;
  challengeToken?: string;
  expiresAt?: string;
}

/**
 * Service configuration interface
 */
export interface ServiceConfig {
  webhookUrl?: string;
  callbackUrl?: string;
  allowedOrigins?: string[];
  rateLimits?: {
    requestsPerMinute?: number;
    requestsPerHour?: number;
    requestsPerDay?: number;
  };
  features?: {
    webhooksEnabled?: boolean;
    analyticsEnabled?: boolean;
    loggingEnabled?: boolean;
  };
  customSettings?: Record<string, any>;
}

/**
 * Main Service interface
 */
export interface Service {
  id: string;
  name: string;
  description: string;
  domain: string;
  domains?: ServiceDomain[];
  status: ServiceStatus | string;
  owner: string;
  createdAt: string;
  updatedAt: string;
  permissions?: ServicePermission[] | string[];
  capabilities?: ServiceCapability[];
  apiKeys?: ServiceApiKey[];
  domainVerificationStatus?: DomainVerificationStatus | string;
  metadata?: ServiceMetadata;
  config?: ServiceConfig;
  tags?: string[];
}

/**
 * Service creation request
 */
export interface ServiceCreateRequest {
  name: string;
  description: string;
  domain: string;
  permissions?: string[];
  config?: ServiceConfig;
  tags?: string[];
}

/**
 * Service update request
 */
export interface ServiceUpdateRequest {
  name?: string;
  description?: string;
  status?: ServiceStatus;
  permissions?: string[];
  config?: ServiceConfig;
  tags?: string[];
}

/**
 * Service validation schemas using Zod
 */

export const ServiceStatusSchema = z.enum([
  ServiceStatus.ACTIVE,
  ServiceStatus.PENDING,
  ServiceStatus.SUSPENDED,
  ServiceStatus.INACTIVE,
]);

export const DomainVerificationStatusSchema = z.enum([
  DomainVerificationStatus.UNVERIFIED,
  DomainVerificationStatus.PENDING,
  DomainVerificationStatus.VERIFIED,
  DomainVerificationStatus.FAILED,
  DomainVerificationStatus.EXPIRED,
]);

export const ServiceCapabilitySchema = z.object({
  id: z.string(),
  name: z.string().min(1).max(100),
  description: z.string().max(500),
  scope: z.string(),
  granted: z.boolean(),
  expiresAt: z.string().optional(),
  constraints: z.record(z.string(), z.any()).optional(),
});

export const ServiceApiKeySchema = z.object({
  id: z.string(),
  name: z.string().min(1).max(100),
  key: z.string().optional(),
  prefix: z.string().optional(),
  createdAt: z.string(),
  lastUsed: z.string().optional(),
  expiresAt: z.string().optional(),
  status: z.enum(['active', 'expired', 'revoked']),
  permissions: z.array(z.string()).optional(),
});

export const ServiceDomainSchema = z.object({
  domain: z.string(),
  verificationStatus: DomainVerificationStatusSchema,
  verifiedAt: z.string().optional(),
  txtRecord: z.string().optional(),
  challengeToken: z.string().optional(),
  expiresAt: z.string().optional(),
});

export const ServiceConfigSchema = z.object({
  webhookUrl: z.string().url().optional(),
  callbackUrl: z.string().url().optional(),
  allowedOrigins: z.array(z.string()).optional(),
  rateLimits: z
    .object({
      requestsPerMinute: z.number().min(1).max(10000).optional(),
      requestsPerHour: z.number().min(1).max(100000).optional(),
      requestsPerDay: z.number().min(1).max(1000000).optional(),
    })
    .optional(),
  features: z
    .object({
      webhooksEnabled: z.boolean().optional(),
      analyticsEnabled: z.boolean().optional(),
      loggingEnabled: z.boolean().optional(),
    })
    .optional(),
  customSettings: z.record(z.any()).optional(),
});

export const ServiceSchema = z.object({
  id: z.string(),
  name: z.string().min(3).max(100),
  description: z.string().min(10).max(500),
  domain: z.string(),
  domains: z.array(ServiceDomainSchema).optional(),
  status: z.union([ServiceStatusSchema, z.string()]),
  owner: z.string(),
  createdAt: z.string(),
  updatedAt: z.string(),
  permissions: z.array(z.union([z.string(), z.any()])).optional(),
  capabilities: z.array(ServiceCapabilitySchema).optional(),
  apiKeys: z.array(ServiceApiKeySchema).optional(),
  domainVerificationStatus: z.union([DomainVerificationStatusSchema, z.string()]).optional(),
  metadata: z.record(z.any()).optional(),
  config: ServiceConfigSchema.optional(),
  tags: z.array(z.string()).optional(),
});

export const ServiceCreateRequestSchema = z.object({
  name: z.string().min(3).max(100),
  description: z.string().min(10).max(500),
  domain: z.string(),
  permissions: z.array(z.string()).optional(),
  config: ServiceConfigSchema.optional(),
  tags: z.array(z.string()).optional(),
});

export const ServiceUpdateRequestSchema = z.object({
  name: z.string().min(3).max(100).optional(),
  description: z.string().min(10).max(500).optional(),
  status: ServiceStatusSchema.optional(),
  permissions: z.array(z.string()).optional(),
  config: ServiceConfigSchema.optional(),
  tags: z.array(z.string()).optional(),
});

/**
 * Type guards
 */
export function isServiceActive(service: Service): boolean {
  return service.status === ServiceStatus.ACTIVE || service.status === 'active';
}

export function isDomainVerified(domain: ServiceDomain | string): boolean {
  if (typeof domain === 'string') {
    return false;
  }
  return domain.verificationStatus === DomainVerificationStatus.VERIFIED;
}

export function hasPermission(service: Service, permission: string): boolean {
  if (!service.permissions) return false;

  if (Array.isArray(service.permissions)) {
    return service.permissions.some((p) => {
      if (typeof p === 'string') {
        return p === permission;
      }
      return p.name === permission && p.granted;
    });
  }

  return false;
}

/**
 * Utility functions
 */
export function formatServiceStatus(status: ServiceStatus | string): string {
  const statusMap: Record<string, string> = {
    [ServiceStatus.ACTIVE]: 'Active',
    [ServiceStatus.PENDING]: 'Pending',
    [ServiceStatus.SUSPENDED]: 'Suspended',
    [ServiceStatus.INACTIVE]: 'Inactive',
  };
  return statusMap[status] || status;
}

export function formatDomainStatus(status: DomainVerificationStatus | string): string {
  const statusMap: Record<string, string> = {
    [DomainVerificationStatus.UNVERIFIED]: 'Unverified',
    [DomainVerificationStatus.PENDING]: 'Pending Verification',
    [DomainVerificationStatus.VERIFIED]: 'Verified',
    [DomainVerificationStatus.FAILED]: 'Verification Failed',
    [DomainVerificationStatus.EXPIRED]: 'Expired',
  };
  return statusMap[status] || status;
}

export function getServiceStatusColor(status: ServiceStatus | string): string {
  const colorMap: Record<string, string> = {
    [ServiceStatus.ACTIVE]: 'green',
    [ServiceStatus.PENDING]: 'yellow',
    [ServiceStatus.SUSPENDED]: 'orange',
    [ServiceStatus.INACTIVE]: 'gray',
  };
  return colorMap[status] || 'gray';
}

export function getDomainStatusColor(status: DomainVerificationStatus | string): string {
  const colorMap: Record<string, string> = {
    [DomainVerificationStatus.VERIFIED]: 'green',
    [DomainVerificationStatus.PENDING]: 'yellow',
    [DomainVerificationStatus.FAILED]: 'red',
    [DomainVerificationStatus.EXPIRED]: 'orange',
    [DomainVerificationStatus.UNVERIFIED]: 'gray',
  };
  return colorMap[status] || 'gray';
}
