import { useCallback, useEffect, useState } from 'react';

interface ServiceMetrics {
  totalRequests: number;
  successRate: number;
  averageLatency: number;
  errorRate: number;
  requestsPerMinute: number[];
  topEndpoints: Array<{
    endpoint: string;
    count: number;
    averageLatency: number;
  }>;
  dailyStats: Array<{
    date: string;
    requests: number;
    errors: number;
  }>;
}

interface UseServiceMetricsReturn {
  metrics: ServiceMetrics | null;
  isLoading: boolean;
  error: Error | null;
  refetch: () => Promise<void>;
}

export function useServiceMetrics(serviceId: string): UseServiceMetricsReturn {
  const [metrics, setMetrics] = useState<ServiceMetrics | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  const fetchMetrics = useCallback(async () => {
    if (!serviceId) {
      setError(new Error('Service ID is required'));
      setIsLoading(false);
      return;
    }

    setIsLoading(true);
    setError(null);

    try {
      // TODO: Replace with actual API call
      const response = await fetch(`/api/services/${serviceId}/metrics`);

      if (!response.ok) {
        throw new Error('Failed to fetch metrics');
      }

      const data = await response.json();
      setMetrics(data.metrics);
    } catch (err) {
      setError(err instanceof Error ? err : new Error('Unknown error'));
      // Mock data for development
      setMetrics({
        totalRequests: 142523,
        successRate: 99.2,
        averageLatency: 45,
        errorRate: 0.8,
        requestsPerMinute: Array.from({ length: 60 }, () => Math.floor(Math.random() * 100) + 20),
        topEndpoints: [
          { endpoint: '/api/v1/data', count: 45230, averageLatency: 32 },
          { endpoint: '/api/v1/auth', count: 28450, averageLatency: 28 },
          { endpoint: '/api/v1/users', count: 18230, averageLatency: 45 },
          { endpoint: '/api/v1/vault', count: 12450, averageLatency: 67 },
          { endpoint: '/api/v1/credentials', count: 8230, averageLatency: 52 },
        ],
        dailyStats: Array.from({ length: 30 }, (_, i) => {
          const date = new Date();
          date.setDate(date.getDate() - (29 - i));
          return {
            date: date.toISOString().split('T')[0],
            requests: Math.floor(Math.random() * 5000) + 3000,
            errors: Math.floor(Math.random() * 50) + 10,
          };
        }),
      });
    } finally {
      setIsLoading(false);
    }
  }, [serviceId]);

  useEffect(() => {
    fetchMetrics();

    // Poll for updates every 30 seconds
    const interval = setInterval(fetchMetrics, 30000);

    return () => clearInterval(interval);
  }, [fetchMetrics]);

  return {
    metrics,
    isLoading,
    error,
    refetch: fetchMetrics,
  };
}
