import type { Service } from '@sonr.io/com';
import { useCallback, useEffect, useRef, useState } from 'react';
import { svcApi } from '../lib/api';

interface UseServiceReturn {
  service: Service | null;
  isLoading: boolean;
  error: Error | null;
  refetch: () => Promise<void>;
  mutate: (updates: Partial<Service>) => void;
  subscribe: (callback: (service: Service) => void) => () => void;
}

// Service cache for optimistic updates
const serviceCache = new Map<string, Service>();

// Subscribers for real-time updates
const serviceSubscribers = new Map<string, Set<(service: Service) => void>>();

/**
 * Hook for individual service data with real-time updates
 */
export function useService(serviceId: string): UseServiceReturn {
  const [service, setService] = useState<Service | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);
  const retryTimeoutRef = useRef<NodeJS.Timeout | null>(null);

  /**
   * Fetch service with error recovery
   */
  const fetchService = useCallback(async () => {
    if (!serviceId) {
      setError(new Error('Service ID is required'));
      setIsLoading(false);
      return;
    }

    setIsLoading(true);
    setError(null);

    // Check cache first
    const cached = serviceCache.get(serviceId);
    if (cached) {
      setService(cached);
      setIsLoading(false);
    }

    try {
      const fetchedService = await svcApi.getServiceDetails(serviceId);

      if (!fetchedService) {
        throw new Error('Service not found');
      }

      // Update cache
      serviceCache.set(serviceId, fetchedService);

      // Update state
      setService(fetchedService);

      // Notify subscribers
      const subscribers = serviceSubscribers.get(serviceId);
      if (subscribers) {
        for (const callback of subscribers) {
          callback(fetchedService);
        }
      }
    } catch (err) {
      setError(err instanceof Error ? err : new Error('Unknown error'));

      // Retry with exponential backoff
      if (!retryTimeoutRef.current) {
        retryTimeoutRef.current = setTimeout(() => {
          retryTimeoutRef.current = null;
          fetchService();
        }, 5000);
      }

      // Use mock data for development
      if (process.env.NODE_ENV === 'development') {
        const mockService = getMockService(serviceId);
        setService(mockService);
        serviceCache.set(serviceId, mockService);
      }
    } finally {
      setIsLoading(false);
    }
  }, [serviceId]);

  /**
   * Optimistic update
   */
  const mutate = useCallback(
    (updates: Partial<Service>) => {
      if (!service) return;

      const updated = { ...service, ...updates };

      // Update cache
      serviceCache.set(serviceId, updated);

      // Update state
      setService(updated);

      // Notify subscribers
      const subscribers = serviceSubscribers.get(serviceId);
      if (subscribers) {
        for (const callback of subscribers) {
          callback(updated);
        }
      }

      // Sync with server in background
      svcApi
        .getServiceDetails(serviceId)
        .then((freshService) => {
          if (freshService) {
            serviceCache.set(serviceId, freshService);
            setService(freshService);
          }
        })
        .catch(console.error);
    },
    [service, serviceId]
  );

  /**
   * Subscribe to real-time updates
   */
  const subscribe = useCallback(
    (callback: (service: Service) => void) => {
      if (!serviceSubscribers.has(serviceId)) {
        serviceSubscribers.set(serviceId, new Set());
      }

      const subscribers = serviceSubscribers.get(serviceId);
      if (subscribers) {
        subscribers.add(callback);
      }

      // Return unsubscribe function
      return () => {
        subscribers.delete(callback);
        if (subscribers.size === 0) {
          serviceSubscribers.delete(serviceId);
        }
      };
    },
    [serviceId]
  );

  // Initial fetch
  useEffect(() => {
    fetchService();
  }, [fetchService]);

  // Clean up on unmount
  useEffect(() => {
    return () => {
      if (retryTimeoutRef.current) {
        clearTimeout(retryTimeoutRef.current);
      }
    };
  }, []);

  // Set up real-time updates (WebSocket in production)
  useEffect(() => {
    if (process.env.NODE_ENV === 'production') {
      // TODO: Implement WebSocket connection for real-time updates
      // const ws = new WebSocket(`${wsEndpoint}/services/${serviceId}`);
      // ws.onmessage = (event) => {
      //   const updatedService = JSON.parse(event.data);
      //   serviceCache.set(serviceId, updatedService);
      //   setService(updatedService);
      // };
    }

    // Simulate real-time updates in development
    if (process.env.NODE_ENV === 'development') {
      const interval = setInterval(() => {
        if (service) {
          const updated = {
            ...service,
            metadata: {
              ...service.metadata,
              totalRequests:
                (service.metadata?.totalRequests || 0) + Math.floor(Math.random() * 100),
            },
          };
          mutate(updated);
        }
      }, 30000); // Update every 30 seconds

      return () => clearInterval(interval);
    }
  }, [service, mutate]);

  return {
    service,
    isLoading,
    error,
    refetch: fetchService,
    mutate,
    subscribe,
  };
}

/**
 * Mock service data for development
 */
function getMockService(serviceId: string): Service {
  return {
    id: serviceId,
    name: 'My API Service',
    description: 'A powerful API service for data processing and management',
    domain: 'api.example.com',
    status: 'active',
    owner: 'did:sonr:alice',
    createdAt: new Date(Date.now() - 30 * 24 * 60 * 60 * 1000).toISOString(),
    updatedAt: new Date().toISOString(),
    permissions: [
      {
        id: 'perm_1',
        name: 'read:data',
        description: 'Read user data',
        scope: 'data',
        granted: true,
      },
      {
        id: 'perm_2',
        name: 'write:data',
        description: 'Write user data',
        scope: 'data',
        granted: true,
      },
      {
        id: 'perm_3',
        name: 'manage:vault',
        description: 'Manage vault contents',
        scope: 'vault',
        granted: false,
      },
    ],
    apiKeys: [
      {
        id: 'key_1',
        name: 'Production Key',
        lastUsed: new Date(Date.now() - 2 * 60 * 60 * 1000).toISOString(),
      },
    ],
    domainVerificationStatus: 'verified',
    metadata: {
      totalRequests: 142523,
      activeUsers: 1250,
      averageLatency: 45,
    },
  };
}
