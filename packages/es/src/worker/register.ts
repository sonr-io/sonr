/**
 * Service Worker Registration Helper for Motor WASM
 * Provides utilities for registering and managing the Motor service worker in web applications.
 */

import type { MotorServiceWorkerConfig } from './types';

export interface RegistrationOptions {
  /** URL to the service worker script */
  workerUrl?: string;
  /** Service worker scope */
  scope?: string;
  /** Whether to skip waiting during update */
  skipWaiting?: boolean;
  /** Whether to enable debug logging */
  debug?: boolean;
  /** Callback when registration succeeds */
  onSuccess?: (registration: ServiceWorkerRegistration) => void;
  /** Callback when registration fails */
  onError?: (error: Error) => void;
  /** Callback when an update is found */
  onUpdateFound?: (registration: ServiceWorkerRegistration) => void;
  /** Callback when the worker is ready */
  onReady?: (registration: ServiceWorkerRegistration) => void;
}

const DEFAULT_OPTIONS: Required<RegistrationOptions> = {
  workerUrl: '/motor-worker.js',
  scope: '/',
  skipWaiting: true,
  debug: false,
  onSuccess: () => {},
  onError: () => {},
  onUpdateFound: () => {},
  onReady: () => {},
};

/**
 * Registers the Motor service worker with comprehensive lifecycle management.
 * @param options Registration options
 * @returns Promise resolving to the service worker registration
 */
export async function registerMotorServiceWorker(
  options: RegistrationOptions = {}
): Promise<ServiceWorkerRegistration | null> {
  const config = { ...DEFAULT_OPTIONS, ...options };
  
  // Check browser support
  if (!('serviceWorker' in navigator)) {
    const error = new Error('Service workers are not supported in this browser');
    config.onError(error);
    if (config.debug) {
      console.error('[Motor Register]', error.message);
    }
    return null;
  }

  // Check for secure context (HTTPS or localhost)
  if (!window.isSecureContext) {
    const error = new Error('Service workers require a secure context (HTTPS or localhost)');
    config.onError(error);
    if (config.debug) {
      console.error('[Motor Register]', error.message);
    }
    return null;
  }

  try {
    if (config.debug) {
      console.log('[Motor Register] Registering service worker:', config.workerUrl);
    }

    // Register the service worker
    const registration = await navigator.serviceWorker.register(config.workerUrl, {
      scope: config.scope,
      type: 'classic',
      updateViaCache: 'none', // Always check for updates
    });

    if (config.debug) {
      console.log('[Motor Register] Service worker registered successfully');
    }

    // Set up event listeners
    setupEventListeners(registration, config);

    // Handle immediate activation for first install
    if (registration.installing) {
      await waitForActivation(registration, config);
    }

    // Check for updates
    await registration.update();

    // Call success callback
    config.onSuccess(registration);

    // Wait for the service worker to be ready
    const readyRegistration = await navigator.serviceWorker.ready;
    config.onReady(readyRegistration);

    return registration;
  } catch (error) {
    const err = error instanceof Error ? error : new Error(String(error));
    config.onError(err);
    if (config.debug) {
      console.error('[Motor Register] Registration failed:', err);
    }
    return null;
  }
}

/**
 * Sets up event listeners for service worker lifecycle events.
 */
function setupEventListeners(
  registration: ServiceWorkerRegistration,
  config: Required<RegistrationOptions>
): void {
  // Listen for update found
  registration.addEventListener('updatefound', () => {
    if (config.debug) {
      console.log('[Motor Register] New service worker update found');
    }
    
    const newWorker = registration.installing;
    if (newWorker) {
      newWorker.addEventListener('statechange', () => {
        if (config.debug) {
          console.log('[Motor Register] Service worker state changed:', newWorker.state);
        }

        if (newWorker.state === 'installed' && navigator.serviceWorker.controller) {
          // New update available
          config.onUpdateFound(registration);
          
          if (config.skipWaiting) {
            // Tell the new worker to skip waiting
            newWorker.postMessage({ type: 'SKIP_WAITING' });
          }
        }
      });
    }
  });

  // Listen for controller change (new worker activated)
  navigator.serviceWorker.addEventListener('controllerchange', () => {
    if (config.debug) {
      console.log('[Motor Register] Controller changed, reloading may be needed');
    }
  });

  // Listen for messages from the service worker
  navigator.serviceWorker.addEventListener('message', (event) => {
    if (config.debug) {
      console.log('[Motor Register] Message from service worker:', event.data);
    }
  });
}

/**
 * Waits for the service worker to activate.
 */
async function waitForActivation(
  registration: ServiceWorkerRegistration,
  config: Required<RegistrationOptions>
): Promise<void> {
  return new Promise((resolve) => {
    const worker = registration.installing || registration.waiting;
    
    if (!worker) {
      resolve();
      return;
    }

    if (worker.state === 'activated') {
      resolve();
      return;
    }

    worker.addEventListener('statechange', function onStateChange() {
      if (worker.state === 'activated') {
        worker.removeEventListener('statechange', onStateChange);
        if (config.debug) {
          console.log('[Motor Register] Service worker activated');
        }
        resolve();
      }
    });
  });
}

/**
 * Unregisters all Motor service workers.
 * @returns Promise resolving to whether unregistration was successful
 */
export async function unregisterMotorServiceWorker(): Promise<boolean> {
  if (!('serviceWorker' in navigator)) {
    return false;
  }

  try {
    const registrations = await navigator.serviceWorker.getRegistrations();
    
    const unregistrations = registrations
      .filter(reg => reg.active?.scriptURL.includes('motor'))
      .map(reg => reg.unregister());
    
    const results = await Promise.all(unregistrations);
    return results.some(result => result === true);
  } catch (error) {
    console.error('[Motor Register] Failed to unregister:', error);
    return false;
  }
}

/**
 * Checks if a Motor service worker is currently registered and active.
 * @returns Promise resolving to registration status
 */
export async function getMotorServiceWorkerStatus(): Promise<{
  registered: boolean;
  active: boolean;
  waiting: boolean;
  installing: boolean;
  registration?: ServiceWorkerRegistration;
}> {
  if (!('serviceWorker' in navigator)) {
    return {
      registered: false,
      active: false,
      waiting: false,
      installing: false,
    };
  }

  try {
    const registrations = await navigator.serviceWorker.getRegistrations();
    const motorRegistration = registrations.find(reg => 
      reg.active?.scriptURL.includes('motor') ||
      reg.waiting?.scriptURL.includes('motor') ||
      reg.installing?.scriptURL.includes('motor')
    );

    if (!motorRegistration) {
      return {
        registered: false,
        active: false,
        waiting: false,
        installing: false,
      };
    }

    return {
      registered: true,
      active: !!motorRegistration.active,
      waiting: !!motorRegistration.waiting,
      installing: !!motorRegistration.installing,
      registration: motorRegistration,
    };
  } catch (error) {
    console.error('[Motor Register] Failed to get status:', error);
    return {
      registered: false,
      active: false,
      waiting: false,
      installing: false,
    };
  }
}

/**
 * Updates the Motor service worker if an update is available.
 * @returns Promise resolving to whether an update was performed
 */
export async function updateMotorServiceWorker(): Promise<boolean> {
  const status = await getMotorServiceWorkerStatus();
  
  if (!status.registered || !status.registration) {
    return false;
  }

  try {
    await status.registration.update();
    
    // Check if there's a waiting worker
    if (status.registration.waiting) {
      // Tell the waiting worker to skip waiting
      status.registration.waiting.postMessage({ type: 'SKIP_WAITING' });
      return true;
    }
    
    return false;
  } catch (error) {
    console.error('[Motor Register] Failed to update:', error);
    return false;
  }
}

/**
 * React hook for Motor service worker registration (if using React).
 * This is a conditional export that only works when React is available.
 * @param options Registration options
 * @returns Service worker registration state
 */
export function useMotorServiceWorker(options: RegistrationOptions = {}) {
  if (typeof window === 'undefined') {
    return {
      registration: null,
      isRegistered: false,
      isActive: false,
      error: null,
    };
  }

  // Check if React is available
  let useState: any;
  let useEffect: any;
  
  try {
    // Try to get React from the global scope if it exists
    const globalReact = (globalThis as any).React || (window as any).React;
    useState = globalReact?.useState;
    useEffect = globalReact?.useEffect;
  } catch {
    // React not available
  }

  // If React hooks aren't available, return a static state
  if (!useState || !useEffect) {
    const staticState = {
      registration: null,
      isRegistered: false,
      isActive: false,
      error: null,
    };
    
    // Still attempt registration but without React state management
    registerMotorServiceWorker(options).catch(() => {});
    
    return staticState;
  }

  const [state, setState] = useState({
    registration: null as ServiceWorkerRegistration | null,
    isRegistered: false,
    isActive: false,
    error: null as Error | null,
  });

  useEffect(() => {
    let mounted = true;

    const register = async () => {
      try {
        const registration = await registerMotorServiceWorker({
          ...options,
          onSuccess: (reg) => {
            if (mounted) {
              setState({
                registration: reg,
                isRegistered: true,
                isActive: !!reg.active,
                error: null,
              });
            }
            options.onSuccess?.(reg);
          },
          onError: (err) => {
            if (mounted) {
              setState((prev: any) => ({ ...prev, error: err }));
            }
            options.onError?.(err);
          },
        });

        if (mounted && registration) {
          setState({
            registration,
            isRegistered: true,
            isActive: !!registration.active,
            error: null,
          });
        }
      } catch (error) {
        if (mounted) {
          const err = error instanceof Error ? error : new Error(String(error));
          setState((prev: any) => ({ ...prev, error: err }));
        }
      }
    };

    register();

    return () => {
      mounted = false;
    };
  }, []);

  return state;
}

/**
 * Utility to create a Motor service worker update prompt for the user.
 * @param options Prompt options
 * @returns Update prompt handler
 */
export function createUpdatePrompt(options: {
  message?: string;
  confirmText?: string;
  cancelText?: string;
  onUpdate?: () => void;
  onCancel?: () => void;
} = {}) {
  const {
    message = 'A new version of Motor is available. Would you like to update?',
    confirmText = 'Update',
    cancelText = 'Later',
    onUpdate = () => window.location.reload(),
    onCancel = () => {},
  } = options;

  return async (registration: ServiceWorkerRegistration) => {
    if (!registration.waiting) {
      return;
    }

    // You can implement your own UI here
    // This is a simple confirm dialog example
    const shouldUpdate = window.confirm(message);

    if (shouldUpdate) {
      registration.waiting.postMessage({ type: 'SKIP_WAITING' });
      
      // Listen for the controller change
      navigator.serviceWorker.addEventListener('controllerchange', () => {
        onUpdate();
      });
    } else {
      onCancel();
    }
  };
}

// Export a default registration function for easy use
export default registerMotorServiceWorker;