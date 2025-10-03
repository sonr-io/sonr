import type { Service } from '@sonr.io/com';
import { useCallback, useEffect, useRef, useState } from 'react';
import { authApi, svcApi } from '../lib/api';

interface UseServicesReturn {
  services: Service[] | null;
  isLoading: boolean;
  error: Error | null;
  refetch: () => Promise<void>;
  mutate: (updater: (services: Service[]) => Service[]) => void;
}

interface CacheEntry {
  data: Service[];
  timestamp: number;
}

// Cache configuration
const CACHE_DURATION = 5 * 60 * 1000; // 5 minutes
const servicesCache = new Map<string, CacheEntry>();

export function useServices(): UseServicesReturn {
  const [services, setServices] = useState<Service[] | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);
  const retryCountRef = useRef(0);
  const maxRetries = 3;

  /**
   * Fetch services with caching and error recovery
   */
  const fetchServices = useCallback(async (forceRefresh = false) => {
    const user = authApi.getCurrentUser();
    if (!user?.address) {
      setError(new Error('Authentication required'));
      setIsLoading(false);
      return;
    }

    const cacheKey = `services-${user.address}`;

    // Check cache unless force refresh
    if (!forceRefresh) {
      const cached = servicesCache.get(cacheKey);
      if (cached && Date.now() - cached.timestamp < CACHE_DURATION) {
        setServices(cached.data);
        setIsLoading(false);
        return;
      }
    }

    setIsLoading(true);
    setError(null);

    try {
      // Use the API client to fetch services
      const fetchedServices = await svcApi.getMyServices(user.address);

      // Update cache
      servicesCache.set(cacheKey, {
        data: fetchedServices,
        timestamp: Date.now(),
      });

      setServices(fetchedServices);
      retryCountRef.current = 0; // Reset retry count on success
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to fetch services';

      // Implement exponential backoff for retries
      if (retryCountRef.current < maxRetries) {
        retryCountRef.current++;
        const delay = Math.min(1000 * 2 ** retryCountRef.current, 10000);

        setTimeout(() => {
          fetchServices(forceRefresh);
        }, delay);

        setError(
          new Error(`${errorMessage}. Retrying... (${retryCountRef.current}/${maxRetries})`)
        );
      } else {
        setError(new Error(errorMessage));

        // Fallback to cached data if available
        const cached = servicesCache.get(cacheKey);
        if (cached) {
          setServices(cached.data);
        } else {
          // Use mock data as last resort in development
          if (process.env.NODE_ENV === 'development') {
            setServices(getMockServices());
          }
        }
      }
    } finally {
      setIsLoading(false);
    }
  }, []);

  /**
   * Optimistic update function
   */
  const mutate = useCallback((updater: (services: Service[]) => Service[]) => {
    setServices((current) => {
      if (!current) return null;
      const updated = updater(current);

      // Update cache with mutated data
      const user = authApi.getCurrentUser();
      if (user?.address) {
        servicesCache.set(`services-${user.address}`, {
          data: updated,
          timestamp: Date.now(),
        });
      }

      return updated;
    });
  }, []);

  // Initial fetch on mount
  useEffect(() => {
    fetchServices();
  }, [fetchServices]);

  // Set up periodic refresh
  useEffect(() => {
    const interval = setInterval(() => {
      fetchServices(false); // Use cache if valid
    }, CACHE_DURATION);

    return () => clearInterval(interval);
  }, [fetchServices]);

  return {
    services,
    isLoading,
    error,
    refetch: () => fetchServices(true), // Force refresh on manual refetch
    mutate,
  };
}

/**
 * Mock data for development
 */
function getMockServices(): Service[] {
  return [
    {
      id: 'svc_1',
      name: 'My API Service',
      description: 'A powerful API service for data processing',
      domain: 'api.example.com',
      status: 'active',
      owner: 'did:sonr:alice',
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
      permissions: ['read:data', 'write:data'],
      apiKeys: [],
      domainVerificationStatus: 'verified',
    },
    {
      id: 'svc_2',
      name: 'Web Application',
      description: 'Main web application frontend',
      domain: 'app.example.com',
      status: 'pending',
      owner: 'did:sonr:alice',
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
      permissions: ['read:profile', 'manage:vault'],
      apiKeys: [],
      domainVerificationStatus: 'pending',
    },
  ];
}
