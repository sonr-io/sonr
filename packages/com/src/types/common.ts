/**
 * Core Common Types
 * Fundamental types used across the Sonr ecosystem
 */

/**
 * Common ID type for entities
 */
export type ID = string;

/**
 * Timestamp in ISO 8601 format
 */
export type Timestamp = string;

/**
 * Status types used across components
 */
export type Status = 'active' | 'inactive' | 'pending' | 'error' | 'warning' | 'success';

/**
 * Sort direction for tables and lists
 */
export type SortDirection = 'asc' | 'desc';

/**
 * Common pagination props
 */
export interface Pagination {
  page: number;
  pageSize: number;
  total: number;
  totalPages: number;
}

/**
 * Common filter props
 */
export interface Filter {
  field: string;
  operator: 'equals' | 'contains' | 'startsWith' | 'endsWith' | 'gt' | 'lt' | 'gte' | 'lte';
  value: string | number | boolean;
}

/**
 * Navigation item
 */
export interface NavigationItem {
  id: ID;
  label: string;
  href: string;
  icon?: string;
  badge?: string | number;
  disabled?: boolean;
  external?: boolean;
  children?: NavigationItem[];
}

/**
 * Breadcrumb item
 */
export interface BreadcrumbItem {
  label: string;
  href?: string;
  current?: boolean;
}

/**
 * Notification
 */
export interface Notification {
  id: ID;
  type: 'info' | 'success' | 'warning' | 'error';
  title: string;
  message?: string;
  timestamp: Timestamp;
  read?: boolean;
  action?: {
    label: string;
    href?: string;
    onClick?: () => void;
  };
}

/**
 * Activity event
 */
export interface ActivityEvent {
  id: ID;
  type: string;
  actor: string;
  action: string;
  target?: string;
  timestamp: Timestamp;
  details?: Record<string, any>;
  icon?: string;
}
