import { z } from 'zod';
import type { Service, ServiceDomain } from './service';

/**
 * API error codes
 */
export enum ApiErrorCode {
  UNKNOWN = 'UNKNOWN',
  NETWORK_ERROR = 'NETWORK_ERROR',
  VALIDATION_ERROR = 'VALIDATION_ERROR',
  AUTHENTICATION_ERROR = 'AUTHENTICATION_ERROR',
  AUTHORIZATION_ERROR = 'AUTHORIZATION_ERROR',
  NOT_FOUND = 'NOT_FOUND',
  RATE_LIMIT = 'RATE_LIMIT',
  SERVER_ERROR = 'SERVER_ERROR',
  TIMEOUT = 'TIMEOUT',
  CONFLICT = 'CONFLICT',
  PRECONDITION_FAILED = 'PRECONDITION_FAILED',
}

/**
 * API response status
 */
export enum ApiResponseStatus {
  SUCCESS = 'success',
  ERROR = 'error',
  PARTIAL = 'partial',
}

/**
 * Base API response interface
 */
export interface ApiResponse<T = any> {
  data?: T;
  error?: ApiError;
  status?: ApiResponseStatus;
  timestamp?: string;
  requestId?: string;
}

/**
 * API error interface
 */
export interface ApiError {
  code: ApiErrorCode | string;
  message: string;
  details?: any;
  statusCode?: number;
  timestamp?: string;
  path?: string;
  requestId?: string;
}

/**
 * Paginated API response
 */
export interface PaginatedResponse<T> extends ApiResponse<T[]> {
  pagination: {
    page: number;
    limit: number;
    total: number;
    totalPages: number;
    hasNext: boolean;
    hasPrev: boolean;
  };
}

/**
 * Batch API response
 */
export interface BatchResponse<T> extends ApiResponse {
  results: Array<{
    id: string;
    success: boolean;
    data?: T;
    error?: ApiError;
  }>;
  summary: {
    total: number;
    successful: number;
    failed: number;
    partial: number;
  };
}

/**
 * Authentication response
 */
export interface AuthResponse extends ApiResponse {
  data?: {
    user: AuthUser;
    token: string;
    refreshToken?: string;
    expiresAt: string;
  };
}

/**
 * Authenticated user
 */
export interface AuthUser {
  id: string;
  username: string;
  did?: string;
  address?: string;
  email?: string;
  createdAt: string;
  lastLogin?: string;
  roles?: string[];
  permissions?: string[];
}

/**
 * Authentication status
 */
export interface AuthStatus {
  isAuthenticated: boolean;
  user: AuthUser | null;
  expiresAt?: string;
}

/**
 * Service API responses
 */
export interface ServiceListResponse extends PaginatedResponse<Service> {
  filters?: {
    status?: string[];
    owner?: string;
    domain?: string;
    tags?: string[];
  };
}

export interface ServiceDetailResponse extends ApiResponse<Service> {
  related?: {
    domains?: ServiceDomain[];
    apiKeys?: number;
    permissions?: string[];
  };
}

export interface ServiceCreateResponse extends ApiResponse<Service> {
  verificationRequired?: boolean;
  verificationInstructions?: {
    txtRecord: string;
    domain: string;
    ttl: number;
  };
}

export interface ServiceUpdateResponse extends ApiResponse<Service> {
  changes?: {
    field: string;
    oldValue: any;
    newValue: any;
  }[];
}

export interface ServiceDeleteResponse extends ApiResponse {
  data?: {
    id: string;
    deletedAt: string;
    cascaded?: string[];
  };
}

/**
 * Domain verification responses
 */
export interface DomainVerificationInitResponse extends ApiResponse {
  data?: {
    domain: string;
    challengeToken: string;
    txtRecord: string;
    expiresAt: string;
    status: string;
  };
}

export interface DomainVerificationStatusResponse extends ApiResponse {
  data?: {
    domain: string;
    status: string;
    verifiedAt?: string;
    lastChecked: string;
    dnsRecordsFound: boolean;
    expectedRecord: string;
    actualRecord?: string;
  };
}

/**
 * API key responses
 */
export interface ApiKeyCreateResponse extends ApiResponse {
  data?: {
    id: string;
    name: string;
    key: string; // Full key only returned on creation
    prefix: string;
    createdAt: string;
    expiresAt?: string;
    permissions: string[];
  };
}

export interface ApiKeyListResponse
  extends PaginatedResponse<{
    id: string;
    name: string;
    prefix: string;
    createdAt: string;
    lastUsed?: string;
    status: string;
  }> {}

/**
 * Analytics responses
 */
export interface AnalyticsResponse extends ApiResponse {
  data?: {
    timeRange: {
      start: string;
      end: string;
    };
    metrics: {
      totalRequests: number;
      uniqueUsers: number;
      averageLatency: number;
      errorRate: number;
      successRate: number;
    };
    timeSeries?: Array<{
      timestamp: string;
      requests: number;
      errors: number;
      latency: number;
    }>;
    breakdown?: {
      byEndpoint?: Record<string, number>;
      byStatus?: Record<string, number>;
      byUser?: Record<string, number>;
    };
  };
}

/**
 * WebSocket message types
 */
export interface WebSocketMessage<T = any> {
  type: 'update' | 'delete' | 'create' | 'error' | 'ping' | 'pong';
  channel: string;
  data?: T;
  timestamp: string;
  id?: string;
}

/**
 * Request configuration
 */
export interface ApiRequestConfig {
  method?: 'GET' | 'POST' | 'PUT' | 'DELETE' | 'PATCH';
  headers?: Record<string, string>;
  params?: Record<string, any>;
  body?: any;
  timeout?: number;
  retries?: number;
  cache?: boolean;
  signal?: AbortSignal;
}

/**
 * API validation schemas
 */
export const ApiErrorSchema = z.object({
  code: z.string(),
  message: z.string(),
  details: z.any().optional(),
  statusCode: z.number().optional(),
  timestamp: z.string().optional(),
  path: z.string().optional(),
  requestId: z.string().optional(),
});

export const ApiResponseSchema = z.object({
  data: z.any().optional(),
  error: ApiErrorSchema.optional(),
  status: z.enum(['success', 'error', 'partial']).optional(),
  timestamp: z.string().optional(),
  requestId: z.string().optional(),
});

export const PaginatedResponseSchema = <T extends z.ZodType>(itemSchema: T) =>
  z.object({
    data: z.array(itemSchema).optional(),
    error: ApiErrorSchema.optional(),
    status: z.enum(['success', 'error', 'partial']).optional(),
    timestamp: z.string().optional(),
    requestId: z.string().optional(),
    pagination: z.object({
      page: z.number(),
      limit: z.number(),
      total: z.number(),
      totalPages: z.number(),
      hasNext: z.boolean(),
      hasPrev: z.boolean(),
    }),
  });

/**
 * Response type guards
 */
export function isApiError(response: any): response is ApiError {
  return response && typeof response.code === 'string' && typeof response.message === 'string';
}

export function isSuccessResponse<T>(response: ApiResponse<T>): boolean {
  return !response.error && response.status !== 'error';
}

export function isPaginatedResponse<T>(response: any): response is PaginatedResponse<T> {
  return response && typeof response.pagination === 'object' && Array.isArray(response.data);
}

/**
 * Error factory functions
 */
export function createApiError(
  code: ApiErrorCode | string,
  message: string,
  details?: any
): ApiError {
  return {
    code,
    message,
    details,
    timestamp: new Date().toISOString(),
  };
}

export function createNetworkError(message = 'Network request failed'): ApiError {
  return createApiError(ApiErrorCode.NETWORK_ERROR, message);
}

export function createValidationError(message: string, details?: any): ApiError {
  return createApiError(ApiErrorCode.VALIDATION_ERROR, message, details);
}

export function createAuthError(message = 'Authentication required'): ApiError {
  return createApiError(ApiErrorCode.AUTHENTICATION_ERROR, message);
}

/**
 * Response builder functions
 */
export function createSuccessResponse<T>(data: T, requestId?: string): ApiResponse<T> {
  return {
    data,
    status: ApiResponseStatus.SUCCESS,
    timestamp: new Date().toISOString(),
    requestId,
  };
}

export function createErrorResponse(error: ApiError, requestId?: string): ApiResponse {
  return {
    error,
    status: ApiResponseStatus.ERROR,
    timestamp: new Date().toISOString(),
    requestId,
  };
}

export function createPaginatedResponse<T>(
  data: T[],
  page: number,
  limit: number,
  total: number
): PaginatedResponse<T> {
  const totalPages = Math.ceil(total / limit);
  return {
    data,
    status: ApiResponseStatus.SUCCESS,
    timestamp: new Date().toISOString(),
    pagination: {
      page,
      limit,
      total,
      totalPages,
      hasNext: page < totalPages,
      hasPrev: page > 1,
    },
  };
}
