/**
 * Formatting Utilities
 * Functions for number and string formatting
 */

import type { Metric, Status } from '../types';

/**
 * Format a number with commas
 */
export function formatNumber(value: number, decimals = 0): string {
  return value.toLocaleString(undefined, {
    minimumFractionDigits: decimals,
    maximumFractionDigits: decimals,
  });
}

/**
 * Format a number as currency
 */
export function formatCurrency(value: number, currency = 'USD'): string {
  return value.toLocaleString(undefined, {
    style: 'currency',
    currency,
  });
}

/**
 * Format a number as a percentage
 */
export function formatPercentage(value: number, decimals = 1): string {
  return `${(value * 100).toFixed(decimals)}%`;
}

/**
 * Format bytes to human-readable size
 */
export function formatBytes(bytes: number, decimals = 2): string {
  if (bytes === 0) return '0 Bytes';

  const k = 1024;
  const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB', 'PB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));

  return `${Number.parseFloat((bytes / k ** i).toFixed(decimals))} ${sizes[i]}`;
}

/**
 * Abbreviate large numbers (e.g., 1.2K, 3.4M)
 */
export function abbreviateNumber(value: number): string {
  if (value < 1000) return value.toString();

  const suffixes = ['', 'K', 'M', 'B', 'T'];
  const suffixNum = Math.floor(`${value}`.length / 3);
  const shortValue = Number.parseFloat(
    (suffixNum !== 0 ? value / 1000 ** suffixNum : value).toPrecision(2)
  );

  if (shortValue % 1 !== 0) {
    return shortValue.toFixed(1) + suffixes[suffixNum];
  }

  return shortValue + suffixes[suffixNum];
}

/**
 * Truncate a string to a maximum length
 */
export function truncate(str: string, maxLength: number, suffix = '...'): string {
  if (str.length <= maxLength) return str;
  return str.slice(0, maxLength - suffix.length) + suffix;
}

/**
 * Convert a string to title case
 */
export function toTitleCase(str: string): string {
  return str.replace(/\w\S*/g, (txt) => txt.charAt(0).toUpperCase() + txt.substr(1).toLowerCase());
}

/**
 * Convert a string to slug format
 */
export function toSlug(str: string): string {
  return str
    .toLowerCase()
    .trim()
    .replace(/[^\w\s-]/g, '')
    .replace(/[\s_-]+/g, '-')
    .replace(/^-+|-+$/g, '');
}

/**
 * Generate a random ID
 */
export function generateId(prefix = ''): string {
  const random = Math.random().toString(36).substr(2, 9);
  const timestamp = Date.now().toString(36);
  return prefix ? `${prefix}_${timestamp}_${random}` : `${timestamp}_${random}`;
}

/**
 * Extract initials from a name
 */
export function getInitials(name: string): string {
  return name
    .split(' ')
    .map((word) => word[0])
    .join('')
    .toUpperCase()
    .slice(0, 2);
}

/**
 * Get status color/variant
 */
export function getStatusColor(status: Status): string {
  const colors: Record<Status, string> = {
    active: 'green',
    inactive: 'gray',
    pending: 'yellow',
    error: 'red',
    warning: 'orange',
    success: 'green',
  };
  return colors[status] || 'gray';
}

/**
 * Get status icon
 */
export function getStatusIcon(status: Status): string {
  const icons: Record<Status, string> = {
    active: 'check-circle',
    inactive: 'x-circle',
    pending: 'clock',
    error: 'alert-circle',
    warning: 'alert-triangle',
    success: 'check-circle',
  };
  return icons[status] || 'circle';
}

/**
 * Format metric with trend
 */
export function formatMetricWithTrend(metric: Metric): string {
  let result = String(metric.value);

  if (metric.unit) {
    result += ` ${metric.unit}`;
  }

  if (metric.change !== undefined) {
    const symbol = metric.change > 0 ? '↑' : metric.change < 0 ? '↓' : '→';
    const changeStr = Math.abs(metric.change).toFixed(1);
    result += ` ${symbol} ${changeStr}%`;
  }

  return result;
}
