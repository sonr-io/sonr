/**
 * @sonr.io/com
 * Shared types, utilities, and constants for the Sonr ecosystem
 */

// Export all types
export * from './types';

// Export all utilities
export * from './utils';

// Export all constants
export * from './constants';

// Re-export commonly used items at top level for convenience
export type {
  // Core types
  ID,
  Timestamp,
  Status,
  Pagination,
  Filter,
  // Service types
  Service,
  ServiceMetrics,
  ServiceRequest,
  // Domain types
  Domain,
  DomainStatus,
  DNSRecord,
  // Permission types
  Permission,
  UCANToken,
  UCANCapability,
  // User types
  User,
  UserRole,
  UserPreferences,
  // API types
  ApiResponse,
  ApiError,
  // Analytics types
  TimeRange,
  DataPoint,
  Metric,
} from './types';

export {
  // Date utilities
  formatDate,
  formatTime,
  formatDateTime,
  getRelativeTime,
  // Format utilities
  formatNumber,
  formatCurrency,
  formatPercentage,
  formatBytes,
  abbreviateNumber,
  // Validation utilities
  isValidEmail,
  isValidUrl,
  isValidDomain,
  schemas,
  // Array utilities
  groupBy,
  sortBy,
  filterBy,
  unique,
} from './utils';

export {
  // Constants
  STATUS,
  HTTP_METHOD,
  ERROR_CODE,
  USER_ROLE,
  THEME,
} from './constants';
