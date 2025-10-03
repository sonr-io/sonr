// Temporary stub to fix module resolution - TODO: Replace with actual @sonr.io/es/client import
class RpcClient {
  constructor(_endpoint: string) {
    console.warn('Using stub implementation for RpcClient');
  }
}

import type {
  ApiResponse,
  DomainVerification,
  Service,
  ServiceCapability,
} from '@sonr.io/com/types';

/**
 * Service Module API Client
 * Handles all interactions with the x/svc module
 */
export class SvcApiClient {
  private rpcClient: RpcClient;
  private baseUrl: string;

  constructor(rpcEndpoint: string) {
    this.rpcClient = new RpcClient(rpcEndpoint);
    this.baseUrl = rpcEndpoint.replace('/rpc', '');
  }

  /**
   * Query module parameters
   */
  async getParams(): Promise<ApiResponse<any>> {
    try {
      const response = await fetch(`${this.baseUrl}/svc/v1/params`);
      const data = await response.json();
      return {
        success: true,
        data: data.params,
      };
    } catch (error) {
      return {
        success: false,
        error: error instanceof Error ? error.message : 'Failed to fetch params',
      };
    }
  }

  /**
   * Get domain verification status
   */
  async getDomainVerification(domain: string): Promise<ApiResponse<DomainVerification>> {
    try {
      const response = await fetch(`${this.baseUrl}/svc/v1/domain/${encodeURIComponent(domain)}`);

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      const data = await response.json();
      return {
        success: true,
        data: data.verification,
      };
    } catch (error) {
      return {
        success: false,
        error: error instanceof Error ? error.message : 'Failed to fetch domain verification',
      };
    }
  }

  /**
   * Get service by ID
   */
  async getService(serviceId: string): Promise<ApiResponse<Service>> {
    try {
      const response = await fetch(
        `${this.baseUrl}/svc/v1/service/${encodeURIComponent(serviceId)}`
      );

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      const data = await response.json();
      return {
        success: true,
        data: data.service,
      };
    } catch (error) {
      return {
        success: false,
        error: error instanceof Error ? error.message : 'Failed to fetch service',
      };
    }
  }

  /**
   * Get all services owned by an address
   */
  async getServicesByOwner(owner: string): Promise<ApiResponse<Service[]>> {
    try {
      const response = await fetch(
        `${this.baseUrl}/svc/v1/services/owner/${encodeURIComponent(owner)}`
      );

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      const data = await response.json();
      return {
        success: true,
        data: data.services || [],
      };
    } catch (error) {
      return {
        success: false,
        error: error instanceof Error ? error.message : 'Failed to fetch services by owner',
      };
    }
  }

  /**
   * Get services bound to a domain
   */
  async getServicesByDomain(domain: string): Promise<ApiResponse<Service[]>> {
    try {
      const response = await fetch(
        `${this.baseUrl}/svc/v1/services/domain/${encodeURIComponent(domain)}`
      );

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      const data = await response.json();
      return {
        success: true,
        data: data.services || [],
      };
    } catch (error) {
      return {
        success: false,
        error: error instanceof Error ? error.message : 'Failed to fetch services by domain',
      };
    }
  }

  /**
   * Initiate domain verification
   * This would typically broadcast a transaction
   */
  async initiateDomainVerification(_domain: string, _signer: any): Promise<ApiResponse<string>> {
    try {
      // TODO: Use @sonr.io/es transaction broadcasting
      // This requires proper message construction with protobuf types
      // For now, returning a placeholder
      return {
        success: false,
        error: 'Transaction broadcasting not yet implemented',
      };
    } catch (error) {
      return {
        success: false,
        error: error instanceof Error ? error.message : 'Failed to initiate domain verification',
      };
    }
  }

  /**
   * Complete domain verification
   */
  async verifyDomain(_domain: string, _signer: any): Promise<ApiResponse<boolean>> {
    try {
      // TODO: Use @sonr.io/es transaction broadcasting
      // This requires proper message construction with protobuf types
      return {
        success: false,
        error: 'Transaction broadcasting not yet implemented',
      };
    } catch (error) {
      return {
        success: false,
        error: error instanceof Error ? error.message : 'Failed to verify domain',
      };
    }
  }

  /**
   * Register a new service
   */
  async registerService(
    _serviceData: {
      name: string;
      domain: string;
      description: string;
      capabilities: ServiceCapability[];
    },
    _signer: any
  ): Promise<ApiResponse<Service>> {
    try {
      // TODO: Use @sonr.io/es transaction broadcasting
      // This requires proper message construction with protobuf types
      return {
        success: false,
        error: 'Transaction broadcasting not yet implemented',
      };
    } catch (error) {
      return {
        success: false,
        error: error instanceof Error ? error.message : 'Failed to register service',
      };
    }
  }

  /**
   * Helper method to check if the API is reachable
   */
  async healthCheck(): Promise<boolean> {
    try {
      const response = await fetch(`${this.baseUrl}/cosmos/base/tendermint/v1beta1/node_info`);
      return response.ok;
    } catch {
      return false;
    }
  }
}

// Singleton instance with configuration
let apiClient: SvcApiClient | null = null;

/**
 * Get or create the service API client
 */
export function getSvcApiClient(endpoint?: string): SvcApiClient {
  const rpcEndpoint = endpoint || process.env.NEXT_PUBLIC_RPC_ENDPOINT || 'http://localhost:26657';

  if (!apiClient || endpoint) {
    apiClient = new SvcApiClient(rpcEndpoint);
  }

  return apiClient;
}

/**
 * Service API helper functions for common operations
 */
export const svcApi = {
  /**
   * Get all services for the current user
   */
  async getMyServices(ownerAddress: string): Promise<Service[]> {
    const client = getSvcApiClient();
    const result = await client.getServicesByOwner(ownerAddress);
    return result.success ? result.data : [];
  },

  /**
   * Get service details with capabilities
   */
  async getServiceDetails(serviceId: string): Promise<Service | null> {
    const client = getSvcApiClient();
    const result = await client.getService(serviceId);
    return result.success ? result.data : null;
  },

  /**
   * Check domain verification status
   */
  async checkDomainStatus(domain: string): Promise<DomainVerification | null> {
    const client = getSvcApiClient();
    const result = await client.getDomainVerification(domain);
    return result.success ? result.data : null;
  },

  /**
   * Poll domain verification status
   */
  async pollDomainVerification(
    domain: string,
    intervalMs = 5000,
    maxAttempts = 60
  ): Promise<DomainVerification | null> {
    const client = getSvcApiClient();
    let attempts = 0;

    return new Promise((resolve) => {
      const checkStatus = async () => {
        attempts++;
        const result = await client.getDomainVerification(domain);

        if (result.success && result.data) {
          const verification = result.data;

          if (
            verification.status === 'DOMAIN_VERIFICATION_STATUS_VERIFIED' ||
            verification.status === 'DOMAIN_VERIFICATION_STATUS_FAILED' ||
            attempts >= maxAttempts
          ) {
            resolve(verification);
            return;
          }
        }

        if (attempts >= maxAttempts) {
          resolve(null);
          return;
        }

        setTimeout(checkStatus, intervalMs);
      };

      checkStatus();
    });
  },
};

export default svcApi;
