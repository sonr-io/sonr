import type { DomainVerification } from '@sonr.io/com/types';
import { useCallback, useEffect, useRef, useState } from 'react';
import { svcApi } from '../lib/api';

interface UseDomainVerificationReturn {
  verification: DomainVerification | null;
  isVerifying: boolean;
  isPolling: boolean;
  error: Error | null;
  startVerification: (domain: string) => Promise<void>;
  checkStatus: (domain: string) => Promise<void>;
  startPolling: (domain: string) => void;
  stopPolling: () => void;
}

interface PollingConfig {
  interval?: number;
  maxAttempts?: number;
  onSuccess?: (verification: DomainVerification) => void;
  onFailure?: (error: Error) => void;
}

/**
 * Hook for domain verification with polling support
 */
export function useDomainVerification(
  initialDomain?: string,
  config?: PollingConfig
): UseDomainVerificationReturn {
  const [verification, setVerification] = useState<DomainVerification | null>(null);
  const [isVerifying, setIsVerifying] = useState(false);
  const [isPolling, setIsPolling] = useState(false);
  const [error, setError] = useState<Error | null>(null);

  const pollingIntervalRef = useRef<NodeJS.Timeout | null>(null);
  const pollingAttemptsRef = useRef(0);
  const currentDomainRef = useRef<string | null>(initialDomain || null);

  const { interval = 5000, maxAttempts = 60, onSuccess, onFailure } = config || {};

  /**
   * Check domain verification status once
   */
  const checkStatus = useCallback(
    async (domain: string) => {
      setError(null);

      try {
        const status = await svcApi.checkDomainStatus(domain);

        if (!status) {
          throw new Error('Failed to fetch domain status');
        }

        setVerification(status);

        // Check if verification is complete
        if (status.status === 'DOMAIN_VERIFICATION_STATUS_VERIFIED') {
          onSuccess?.(status);
          stopPolling();
        } else if (status.status === 'DOMAIN_VERIFICATION_STATUS_FAILED') {
          const error = new Error('Domain verification failed');
          setError(error);
          onFailure?.(error);
          stopPolling();
        }

        return status;
      } catch (err) {
        const error = err instanceof Error ? err : new Error('Failed to check domain status');
        setError(error);
        throw error;
      }
    },
    [onSuccess, onFailure]
  );

  /**
   * Start domain verification process
   */
  const startVerification = useCallback(
    async (domain: string) => {
      setIsVerifying(true);
      setError(null);
      currentDomainRef.current = domain;

      try {
        // In a real implementation, this would initiate the verification
        // For now, we just start polling the status
        await checkStatus(domain);
        startPolling(domain);
      } catch (err) {
        const error = err instanceof Error ? err : new Error('Failed to start verification');
        setError(error);
        onFailure?.(error);
      } finally {
        setIsVerifying(false);
      }
    },
    [checkStatus, onFailure]
  );

  /**
   * Start polling for verification status
   */
  const startPolling = useCallback(
    (domain: string) => {
      // Clear any existing polling
      stopPolling();

      setIsPolling(true);
      pollingAttemptsRef.current = 0;
      currentDomainRef.current = domain;

      // Set up polling interval
      pollingIntervalRef.current = setInterval(async () => {
        pollingAttemptsRef.current++;

        // Check if we've exceeded max attempts
        if (pollingAttemptsRef.current >= maxAttempts) {
          const error = new Error('Domain verification timeout');
          setError(error);
          onFailure?.(error);
          stopPolling();
          return;
        }

        try {
          await checkStatus(domain);
        } catch (err) {
          // Continue polling even if individual checks fail
          console.error('Polling check failed:', err);

          // Stop polling after multiple consecutive failures
          if (pollingAttemptsRef.current > 3) {
            const error = err instanceof Error ? err : new Error('Polling failed');
            setError(error);
            onFailure?.(error);
            stopPolling();
          }
        }
      }, interval);
    },
    [interval, maxAttempts, checkStatus, onFailure]
  );

  /**
   * Stop polling
   */
  const stopPolling = useCallback(() => {
    if (pollingIntervalRef.current) {
      clearInterval(pollingIntervalRef.current);
      pollingIntervalRef.current = null;
    }
    setIsPolling(false);
    pollingAttemptsRef.current = 0;
  }, []);

  /**
   * Clean up on unmount
   */
  useEffect(() => {
    return () => {
      stopPolling();
    };
  }, [stopPolling]);

  /**
   * Auto-start polling if initial domain is provided
   */
  useEffect(() => {
    if (initialDomain) {
      startVerification(initialDomain);
    }
  }, [initialDomain, startVerification]);

  return {
    verification,
    isVerifying,
    isPolling,
    error,
    startVerification,
    checkStatus,
    startPolling,
    stopPolling,
  };
}

/**
 * Hook for managing multiple domain verifications
 */
export function useDomainVerifications() {
  const [verifications, setVerifications] = useState<Map<string, DomainVerification>>(new Map());
  const [activePolls, setActivePolls] = useState<Set<string>>(new Set());

  const addVerification = useCallback((domain: string, verification: DomainVerification) => {
    setVerifications((prev) => new Map(prev).set(domain, verification));
  }, []);

  const removeVerification = useCallback((domain: string) => {
    setVerifications((prev) => {
      const next = new Map(prev);
      next.delete(domain);
      return next;
    });
  }, []);

  const startPollingFor = useCallback((domain: string) => {
    setActivePolls((prev) => new Set(prev).add(domain));
  }, []);

  const stopPollingFor = useCallback((domain: string) => {
    setActivePolls((prev) => {
      const next = new Set(prev);
      next.delete(domain);
      return next;
    });
  }, []);

  return {
    verifications: Array.from(verifications.values()),
    activePolls: Array.from(activePolls),
    addVerification,
    removeVerification,
    startPollingFor,
    stopPollingFor,
    isPolling: (domain: string) => activePolls.has(domain),
    getVerification: (domain: string) => verifications.get(domain),
  };
}
