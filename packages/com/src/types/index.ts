/**
 * @sonr.io/com - Type Definitions
 * Central export for all shared types
 */

// Core common types
export * from './common';

// Service types
export * from './service';

// Domain & DNS types
export * from './domain';

// Permission & auth types
export * from './permission';

// User types
export * from './user';

// Analytics types
export * from './analytics';

// API types
export * from './api';

// Re-export commonly used types for convenience
export type {
  ID,
  Timestamp,
  Status,
  SortDirection,
  Pagination,
  Filter,
} from './common';

export type {
  Service,
  ServiceMetrics,
  ServiceEndpoint,
  ServiceRequest,
} from './service';

export type {
  Domain,
  DomainStatus,
  DNSRecord,
  DomainVerificationRequest,
} from './domain';

export type {
  Permission,
  UCANToken,
  UCANCapability,
  AuditLogEntry,
} from './permission';

export type {
  User,
  UserRole,
  UserPreferences,
} from './user';

export type {
  TimeRange,
  DataPoint,
  Metric,
  PerformanceMetric,
} from './analytics';

export type {
  ApiResponse,
  ApiError,
  ApiRequestConfig,
} from './api';
