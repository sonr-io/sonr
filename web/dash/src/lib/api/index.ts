import { authApi } from './auth';
import { svcApi } from './svc';

/**
 * Centralized API configuration and error handling
 */
export class ApiConfig {
  private static instance: ApiConfig;

  public rpcEndpoint: string;
  public restEndpoint: string;
  public authUrl: string;
  public chainId: string;
  public requestInterceptors: Array<(config: any) => any> = [];
  public responseInterceptors: Array<(response: any) => any> = [];

  private constructor() {
    // Load from environment variables with defaults
    this.rpcEndpoint = process.env.NEXT_PUBLIC_RPC_ENDPOINT || 'http://localhost:26657';
    this.restEndpoint = process.env.NEXT_PUBLIC_REST_ENDPOINT || 'http://localhost:1317';
    this.authUrl = process.env.NEXT_PUBLIC_AUTH_URL || 'http://localhost:3001';
    this.chainId = process.env.NEXT_PUBLIC_CHAIN_ID || 'sonrtest_1-1';
  }

  static getInstance(): ApiConfig {
    if (!ApiConfig.instance) {
      ApiConfig.instance = new ApiConfig();
    }
    return ApiConfig.instance;
  }

  /**
   * Configure endpoints
   */
  configure(
    config: Partial<{
      rpcEndpoint: string;
      restEndpoint: string;
      authUrl: string;
      chainId: string;
    }>
  ): void {
    Object.assign(this, config);
  }

  /**
   * Add request interceptor
   */
  addRequestInterceptor(interceptor: (config: any) => any): void {
    this.requestInterceptors.push(interceptor);
  }

  /**
   * Add response interceptor
   */
  addResponseInterceptor(interceptor: (response: any) => any): void {
    this.responseInterceptors.push(interceptor);
  }

  /**
   * Apply interceptors to fetch config
   */
  applyRequestInterceptors(config: RequestInit): RequestInit {
    return this.requestInterceptors.reduce((acc, interceptor) => interceptor(acc), config);
  }

  /**
   * Apply interceptors to response
   */
  applyResponseInterceptors(response: Response): Response {
    return this.responseInterceptors.reduce((acc, interceptor) => interceptor(acc), response);
  }
}

/**
 * Error handling utilities
 */
export class ApiError extends Error {
  public code: string;
  public statusCode?: number;
  public details?: any;

  constructor(message: string, code: string, statusCode?: number, details?: any) {
    super(message);
    this.name = 'ApiError';
    this.code = code;
    this.statusCode = statusCode;
    this.details = details;
  }

  static fromResponse(response: Response, body?: any): ApiError {
    const message = body?.message || body?.error || `HTTP ${response.status}`;
    const code = body?.code || 'API_ERROR';
    return new ApiError(message, code, response.status, body);
  }

  static networkError(error: Error): ApiError {
    return new ApiError('Network request failed', 'NETWORK_ERROR', undefined, {
      originalError: error.message,
    });
  }

  static validationError(message: string, details?: any): ApiError {
    return new ApiError(message, 'VALIDATION_ERROR', 400, details);
  }
}

/**
 * Enhanced fetch with interceptors and error handling
 */
export async function apiFetch(url: string, options?: RequestInit): Promise<Response> {
  const config = ApiConfig.getInstance();

  try {
    // Apply request interceptors
    const requestConfig = config.applyRequestInterceptors({
      ...options,
      headers: {
        'Content-Type': 'application/json',
        ...options?.headers,
      },
    });

    // Make the request
    const response = await fetch(url, requestConfig);

    // Apply response interceptors
    const processedResponse = config.applyResponseInterceptors(response);

    // Check for errors
    if (!processedResponse.ok) {
      const body = await processedResponse.json().catch(() => null);
      throw ApiError.fromResponse(processedResponse, body);
    }

    return processedResponse;
  } catch (error) {
    if (error instanceof ApiError) {
      throw error;
    }
    if (error instanceof TypeError && error.message === 'Failed to fetch') {
      throw ApiError.networkError(error);
    }
    throw new ApiError(error instanceof Error ? error.message : 'Unknown error', 'UNKNOWN_ERROR');
  }
}

/**
 * Setup default interceptors
 */
export function setupDefaultInterceptors(): void {
  const config = ApiConfig.getInstance();

  // Add authentication header
  config.addRequestInterceptor((requestConfig) => {
    const user = authApi.getCurrentUser();
    if (user?.address) {
      return {
        ...requestConfig,
        headers: {
          ...requestConfig.headers,
          'X-User-Address': user.address,
        },
      };
    }
    return requestConfig;
  });

  // Handle authentication errors
  config.addResponseInterceptor((response) => {
    if (response.status === 401) {
      // Clear session and redirect to auth
      authApi.signIn();
    }
    return response;
  });

  // Add request ID for tracking
  config.addRequestInterceptor((requestConfig) => {
    return {
      ...requestConfig,
      headers: {
        ...requestConfig.headers,
        'X-Request-ID': generateRequestId(),
      },
    };
  });
}

/**
 * Generate unique request ID
 */
function generateRequestId(): string {
  return `${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
}

/**
 * API client factory
 */
export const api = {
  auth: authApi,
  svc: svcApi,
  config: ApiConfig.getInstance(),
  fetch: apiFetch,
  Error: ApiError,
  setupInterceptors: setupDefaultInterceptors,
};

export type { AuthApiClient } from './auth';
// Export individual APIs for convenience
export { authApi } from './auth';
export type { SvcApiClient } from './svc';
export { svcApi } from './svc';

export default api;
