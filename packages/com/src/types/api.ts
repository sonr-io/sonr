/**
 * API Types
 * Types for API requests, responses, and errors
 */

import type { Timestamp } from './common';

/**
 * API response wrapper
 */
export interface ApiResponse<T> {
  data?: T;
  error?: ApiError;
  status: number;
  timestamp: Timestamp;
}

/**
 * API error
 */
export interface ApiError {
  code: string;
  message: string;
  details?: Record<string, any>;
  stack?: string;
}

/**
 * API request configuration
 */
export interface ApiRequestConfig {
  method: 'GET' | 'POST' | 'PUT' | 'DELETE' | 'PATCH';
  headers?: Record<string, string>;
  params?: Record<string, any>;
  body?: any;
  timeout?: number;
  retry?: {
    attempts: number;
    delay: number;
    backoff?: 'linear' | 'exponential';
  };
}

/**
 * Paginated API response
 */
export interface PaginatedResponse<T> {
  items: T[];
  pagination: {
    page: number;
    pageSize: number;
    total: number;
    totalPages: number;
  };
}

/**
 * API endpoint metadata
 */
export interface ApiEndpoint {
  path: string;
  method: 'GET' | 'POST' | 'PUT' | 'DELETE' | 'PATCH';
  description?: string;
  authenticated: boolean;
  rateLimit?: {
    requests: number;
    window: number;
  };
}
