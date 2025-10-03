/**
 * Date & Time Utilities
 * Functions for date/time formatting and manipulation
 */

import type { TimeRange, Timestamp } from '../types';

/**
 * Format a timestamp to a human-readable date string
 */
export function formatDate(timestamp: Timestamp, options?: Intl.DateTimeFormatOptions): string {
  const date = new Date(timestamp);
  return date.toLocaleDateString(
    undefined,
    options || {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
    }
  );
}

/**
 * Format a timestamp to a human-readable time string
 */
export function formatTime(timestamp: Timestamp, options?: Intl.DateTimeFormatOptions): string {
  const date = new Date(timestamp);
  return date.toLocaleTimeString(
    undefined,
    options || {
      hour: '2-digit',
      minute: '2-digit',
    }
  );
}

/**
 * Format a timestamp to a human-readable date and time string
 */
export function formatDateTime(timestamp: Timestamp): string {
  return `${formatDate(timestamp)} ${formatTime(timestamp)}`;
}

/**
 * Get a relative time string (e.g., "2 hours ago")
 */
export function getRelativeTime(timestamp: Timestamp): string {
  const now = Date.now();
  const then = new Date(timestamp).getTime();
  const diff = now - then;

  const seconds = Math.floor(diff / 1000);
  const minutes = Math.floor(seconds / 60);
  const hours = Math.floor(minutes / 60);
  const days = Math.floor(hours / 24);
  const weeks = Math.floor(days / 7);
  const months = Math.floor(days / 30);
  const years = Math.floor(days / 365);

  if (seconds < 60) return `${seconds} second${seconds !== 1 ? 's' : ''} ago`;
  if (minutes < 60) return `${minutes} minute${minutes !== 1 ? 's' : ''} ago`;
  if (hours < 24) return `${hours} hour${hours !== 1 ? 's' : ''} ago`;
  if (days < 7) return `${days} day${days !== 1 ? 's' : ''} ago`;
  if (weeks < 4) return `${weeks} week${weeks !== 1 ? 's' : ''} ago`;
  if (months < 12) return `${months} month${months !== 1 ? 's' : ''} ago`;
  return `${years} year${years !== 1 ? 's' : ''} ago`;
}

/**
 * Get date range from preset
 */
export function getDateRangeFromPreset(preset: TimeRange['preset']): TimeRange {
  const now = new Date();
  const today = new Date(now.getFullYear(), now.getMonth(), now.getDate());

  switch (preset) {
    case 'today':
      return {
        start: today,
        end: now,
        preset,
      };
    case 'yesterday': {
      const yesterday = new Date(today);
      yesterday.setDate(yesterday.getDate() - 1);
      return {
        start: yesterday,
        end: today,
        preset,
      };
    }
    case 'last7days': {
      const weekAgo = new Date(today);
      weekAgo.setDate(weekAgo.getDate() - 7);
      return {
        start: weekAgo,
        end: now,
        preset,
      };
    }
    case 'last30days': {
      const monthAgo = new Date(today);
      monthAgo.setDate(monthAgo.getDate() - 30);
      return {
        start: monthAgo,
        end: now,
        preset,
      };
    }
    case 'thisMonth': {
      const startOfMonth = new Date(now.getFullYear(), now.getMonth(), 1);
      return {
        start: startOfMonth,
        end: now,
        preset,
      };
    }
    case 'lastMonth': {
      const startOfLastMonth = new Date(now.getFullYear(), now.getMonth() - 1, 1);
      const endOfLastMonth = new Date(now.getFullYear(), now.getMonth(), 0);
      return {
        start: startOfLastMonth,
        end: endOfLastMonth,
        preset,
      };
    }
    default:
      return {
        start: today,
        end: now,
        preset: 'today',
      };
  }
}

/**
 * Format duration in milliseconds to human-readable string
 */
export function formatDuration(ms: number): string {
  if (ms < 1000) return `${ms}ms`;

  const seconds = Math.floor(ms / 1000);
  if (seconds < 60) return `${seconds}s`;

  const minutes = Math.floor(seconds / 60);
  if (minutes < 60) return `${minutes}m ${seconds % 60}s`;

  const hours = Math.floor(minutes / 60);
  if (hours < 24) return `${hours}h ${minutes % 60}m`;

  const days = Math.floor(hours / 24);
  return `${days}d ${hours % 24}h`;
}
