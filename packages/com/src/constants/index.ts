/**
 * @sonr.io/com - Constants
 * Shared constants and configurations
 */

/**
 * Status constants
 */
export const STATUS = {
  ACTIVE: 'active',
  INACTIVE: 'inactive',
  PENDING: 'pending',
  ERROR: 'error',
  WARNING: 'warning',
  SUCCESS: 'success',
} as const;

/**
 * HTTP methods
 */
export const HTTP_METHOD = {
  GET: 'GET',
  POST: 'POST',
  PUT: 'PUT',
  DELETE: 'DELETE',
  PATCH: 'PATCH',
} as const;

/**
 * Time intervals in milliseconds
 */
export const TIME = {
  SECOND: 1000,
  MINUTE: 60 * 1000,
  HOUR: 60 * 60 * 1000,
  DAY: 24 * 60 * 60 * 1000,
  WEEK: 7 * 24 * 60 * 60 * 1000,
} as const;

/**
 * Default pagination settings
 */
export const PAGINATION = {
  DEFAULT_PAGE: 1,
  DEFAULT_PAGE_SIZE: 20,
  MAX_PAGE_SIZE: 100,
} as const;

/**
 * API error codes
 */
export const ERROR_CODE = {
  NETWORK_ERROR: 'NETWORK_ERROR',
  TIMEOUT: 'TIMEOUT',
  UNAUTHORIZED: 'UNAUTHORIZED',
  FORBIDDEN: 'FORBIDDEN',
  NOT_FOUND: 'NOT_FOUND',
  SERVER_ERROR: 'SERVER_ERROR',
  VALIDATION_ERROR: 'VALIDATION_ERROR',
  RATE_LIMIT: 'RATE_LIMIT',
} as const;

/**
 * User roles
 */
export const USER_ROLE = {
  OWNER: 'owner',
  ADMIN: 'admin',
  DEVELOPER: 'developer',
  VIEWER: 'viewer',
} as const;

/**
 * Permission effects
 */
export const PERMISSION_EFFECT = {
  ALLOW: 'allow',
  DENY: 'deny',
} as const;

/**
 * Domain verification methods
 */
export const VERIFICATION_METHOD = {
  DNS: 'dns',
  HTTP: 'http',
} as const;

/**
 * Chart types
 */
export const CHART_TYPE = {
  LINE: 'line',
  AREA: 'area',
  BAR: 'bar',
  PIE: 'pie',
  DONUT: 'donut',
  RADAR: 'radar',
} as const;

/**
 * Theme options
 */
export const THEME = {
  LIGHT: 'light',
  DARK: 'dark',
  SYSTEM: 'system',
} as const;

/**
 * Notification frequency
 */
export const NOTIFICATION_FREQUENCY = {
  REALTIME: 'realtime',
  DAILY: 'daily',
  WEEKLY: 'weekly',
  NEVER: 'never',
} as const;

/**
 * Performance status thresholds
 */
export const PERFORMANCE_THRESHOLD = {
  GOOD: 'good',
  WARNING: 'warning',
  CRITICAL: 'critical',
} as const;

/**
 * DNS record types
 */
export const DNS_RECORD_TYPE = {
  TXT: 'TXT',
  CNAME: 'CNAME',
  A: 'A',
  AAAA: 'AAAA',
} as const;

/**
 * Service categories
 */
export const SERVICE_CATEGORY = {
  API: 'api',
  WEBAPP: 'webapp',
  MOBILE: 'mobile',
  IOT: 'iot',
  BLOCKCHAIN: 'blockchain',
  OTHER: 'other',
} as const;

/**
 * Permission categories
 */
export const PERMISSION_CATEGORY = {
  BASIC: 'basic',
  STANDARD: 'standard',
  ADVANCED: 'advanced',
  CUSTOM: 'custom',
} as const;

/**
 * Time range presets
 */
export const TIME_PRESET = {
  TODAY: 'today',
  YESTERDAY: 'yesterday',
  LAST_7_DAYS: 'last7days',
  LAST_30_DAYS: 'last30days',
  THIS_MONTH: 'thisMonth',
  LAST_MONTH: 'lastMonth',
  CUSTOM: 'custom',
} as const;

/**
 * Regular expressions for validation
 */
export const REGEX = {
  EMAIL: /^[^\s@]+@[^\s@]+\.[^\s@]+$/,
  URL: /^https?:\/\/.+/,
  DOMAIN: /^(?:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\.)*[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?$/i,
  API_KEY: /^sk_(test|live)_[a-zA-Z0-9]{24,}$/,
  DID: /^did:[a-z0-9]+:[a-zA-Z0-9:.-]+$/,
  USERNAME: /^[a-zA-Z0-9_-]+$/,
  PHONE: /^\+?[1-9]\d{1,14}$/,
} as const;

/**
 * Default chart colors
 */
export const CHART_COLORS = [
  '#3b82f6', // blue
  '#10b981', // emerald
  '#f59e0b', // amber
  '#ef4444', // red
  '#8b5cf6', // violet
  '#ec4899', // pink
  '#14b8a6', // teal
  '#f97316', // orange
] as const;
