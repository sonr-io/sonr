/**
 * @sonr.io/com - Utility Functions
 * Central export for all utility functions
 */

// Date & time utilities
export * from './date';

// Formatting utilities
export * from './format';

// Validation utilities
export * from './validation';

// Array & object utilities
export * from './array';

// Chart & analytics utilities
export * from './chart';

// Export utility collections for convenience
export {
  formatDate,
  formatTime,
  formatDateTime,
  getRelativeTime,
  getDateRangeFromPreset,
  formatDuration,
} from './date';

export {
  formatNumber,
  formatCurrency,
  formatPercentage,
  formatBytes,
  abbreviateNumber,
  truncate,
  toTitleCase,
  toSlug,
  generateId,
  getInitials,
  getStatusColor,
  getStatusIcon,
  formatMetricWithTrend,
} from './format';

export {
  isValidEmail,
  isValidUrl,
  isValidDomain,
  isValidApiKey,
  isValidDID,
  isValidPhoneNumber,
  schemas,
  createValidator,
} from './validation';

export {
  groupBy,
  sortBy,
  filterBy,
  deepClone,
  deepMerge,
  unique,
  uniqueBy,
  chunk,
  flatten,
  pick,
  omit,
} from './array';

export {
  generateChartColors,
  calculatePercentageChange,
  aggregateDataPoints,
  calculateStatistics,
  generateTrendLine,
  calculateMovingAverage,
} from './chart';
