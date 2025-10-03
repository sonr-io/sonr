/**
 * Service worker lifecycle management for Motor WASM
 */

import type { ServiceWorkerStatus, EnvironmentInfo } from './types';

/**
 * Detects the current environment capabilities
 */
export function detectEnvironment(): EnvironmentInfo {
  const isBrowser = typeof window !== 'undefined' && typeof window.document !== 'undefined';
  const isNode = typeof process !== 'undefined' && process.versions && process.versions.node;
  const supportsServiceWorker = isBrowser && 'serviceWorker' in navigator;
  const supportsWasm = typeof WebAssembly !== 'undefined';

  return {
    is_browser: isBrowser,
    is_node: !!isNode,
    supports_service_worker: supportsServiceWorker,
    supports_wasm: supportsWasm,
  };
}

/**
 * Registers the Motor service worker
 */
export async function registerMotorServiceWorker(
  workerUrl: string,
  options?: RegistrationOptions
): Promise<ServiceWorkerRegistration> {
  const env = detectEnvironment();
  
  if (!env.supports_service_worker) {
    throw new Error('Service workers are not supported in this environment');
  }

  if (!env.supports_wasm) {
    throw new Error('WebAssembly is not supported in this environment');
  }

  try {
    // Register the service worker
    const registration = await navigator.serviceWorker.register(workerUrl, {
      scope: options?.scope || '/',
      type: options?.type || 'classic',
      updateViaCache: options?.updateViaCache || 'imports',
    });

    // Wait for the service worker to be ready
    await navigator.serviceWorker.ready;

    console.log('Motor service worker registered successfully');
    return registration;
  } catch (error) {
    console.error('Failed to register Motor service worker:', error);
    throw error;
  }
}

/**
 * Unregisters the Motor service worker
 */
export async function unregisterMotorServiceWorker(): Promise<boolean> {
  const env = detectEnvironment();
  
  if (!env.supports_service_worker) {
    return false;
  }

  try {
    const registrations = await navigator.serviceWorker.getRegistrations();
    
    for (const registration of registrations) {
      // Check if this is the Motor service worker
      if (registration.active?.scriptURL.includes('motr')) {
        const success = await registration.unregister();
        if (success) {
          console.log('Motor service worker unregistered successfully');
        }
        return success;
      }
    }
    
    return false;
  } catch (error) {
    console.error('Failed to unregister Motor service worker:', error);
    return false;
  }
}

/**
 * Gets the current Motor service worker status
 */
export async function getMotorServiceWorkerStatus(): Promise<ServiceWorkerStatus> {
  const env = detectEnvironment();
  
  if (!env.supports_service_worker) {
    return { registered: false };
  }

  try {
    const registrations = await navigator.serviceWorker.getRegistrations();
    
    for (const registration of registrations) {
      // Check if this is the Motor service worker
      if (registration.active?.scriptURL.includes('motr')) {
        return {
          registered: true,
          state: registration.active.state,
          url: registration.active.scriptURL,
          registered_at: Date.now(),
        };
      }
    }
    
    return { registered: false };
  } catch (error) {
    console.error('Failed to get Motor service worker status:', error);
    return { registered: false };
  }
}

/**
 * Configuration for MotorServiceWorkerManager
 */
export interface MotorServiceWorkerManagerConfig {
  worker_script: string;
  scope?: string;
  debug?: boolean;
  type?: WorkerType;
  updateViaCache?: ServiceWorkerUpdateViaCache;
}

/**
 * Service Worker Manager class for advanced lifecycle management
 */
export class MotorServiceWorkerManager {
  private registration?: ServiceWorkerRegistration;
  private config: MotorServiceWorkerManagerConfig;
  private updateCheckInterval?: number;
  private debug: boolean;

  constructor(config: MotorServiceWorkerManagerConfig) {
    this.config = config;
    this.debug = config.debug || false;
  }

  /**
   * Checks if service workers are supported in the current environment
   */
  static isSupported(): boolean {
    const env = detectEnvironment();
    return env.supports_service_worker;
  }

  /**
   * Detects the current environment capabilities
   */
  static detectEnvironment(): EnvironmentInfo {
    return detectEnvironment();
  }

  /**
   * Gets browser compatibility information
   */
  static getBrowserCompatibility(): {
    compatible: boolean;
    issues: string[];
    recommendations: string[];
  } {
    const env = detectEnvironment();
    const issues: string[] = [];
    const recommendations: string[] = [];

    if (!env.is_browser && !env.is_node) {
      issues.push('Unknown environment');
      recommendations.push('Use a modern browser or Node.js');
    }

    if (env.is_browser && !env.supports_service_worker) {
      issues.push('Service workers not supported');
      recommendations.push('Use a modern browser with service worker support');
    }

    if (!env.supports_wasm) {
      issues.push('WebAssembly not supported');
      recommendations.push('Update your browser or Node.js version');
    }

    return {
      compatible: issues.length === 0,
      issues,
      recommendations,
    };
  }

  /**
   * Registers and initializes the service worker
   */
  async register(): Promise<ServiceWorkerRegistration> {
    if (this.registration) {
      return this.registration;
    }

    const options: RegistrationOptions = {
      scope: this.config.scope,
      type: this.config.type,
      updateViaCache: this.config.updateViaCache,
    };

    this.registration = await registerMotorServiceWorker(this.config.worker_script, options);
    
    // Set up update checking
    this.startUpdateChecking();
    
    // Listen for service worker events
    this.attachEventListeners();
    
    return this.registration;
  }

  /**
   * Unregisters the service worker
   */
  async unregister(): Promise<boolean> {
    this.stopUpdateChecking();
    
    if (this.registration) {
      const success = await this.registration.unregister();
      if (success) {
        this.registration = undefined;
      }
      return success;
    }
    
    return await unregisterMotorServiceWorker();
  }

  /**
   * Checks for service worker updates
   */
  async checkForUpdates(): Promise<void> {
    if (!this.registration) {
      return;
    }

    try {
      await this.registration.update();
      if (this.debug) {
        console.log('Checked for Motor service worker updates');
      }
    } catch (error) {
      console.error('Failed to check for updates:', error);
    }
  }

  /**
   * Gets the current status
   */
  async getStatus(): Promise<ServiceWorkerStatus> {
    if (!this.registration) {
      return { registered: false };
    }

    const worker = this.registration.active || this.registration.waiting || this.registration.installing;
    
    if (!worker) {
      return { registered: false };
    }

    return {
      registered: true,
      state: worker.state,
      url: worker.scriptURL,
      registered_at: Date.now(),
    };
  }

  /**
   * Sends a message to the service worker
   */
  async sendMessage(message: any): Promise<void> {
    if (!this.registration?.active) {
      throw new Error('Service worker is not active');
    }

    this.registration.active.postMessage(message);
  }

  /**
   * Starts periodic update checking
   */
  private startUpdateChecking(intervalMs: number = 3600000): void {
    if (this.updateCheckInterval) {
      return;
    }

    this.updateCheckInterval = window.setInterval(() => {
      this.checkForUpdates();
    }, intervalMs);
  }

  /**
   * Stops periodic update checking
   */
  private stopUpdateChecking(): void {
    if (this.updateCheckInterval) {
      clearInterval(this.updateCheckInterval);
      this.updateCheckInterval = undefined;
    }
  }

  /**
   * Attaches event listeners for service worker events
   */
  private attachEventListeners(): void {
    if (!this.registration) {
      return;
    }

    this.registration.addEventListener('updatefound', () => {
      if (this.debug) {
        console.log('Motor service worker update found');
      }
      
      const newWorker = this.registration!.installing;
      if (newWorker) {
        newWorker.addEventListener('statechange', () => {
          if (this.debug) {
            console.log('Motor service worker state changed:', newWorker.state);
          }
          
          if (newWorker.state === 'activated') {
            console.log('Motor service worker updated and activated');
          }
        });
      }
    });

    // Listen for messages from the service worker
    navigator.serviceWorker.addEventListener('message', (event) => {
      if (this.debug) {
        console.log('Message from Motor service worker:', event.data);
      }
    });
  }
}

/**
 * Default service worker manager instance
 */
let defaultManager: MotorServiceWorkerManager | undefined;

/**
 * Gets or creates the default service worker manager
 */
export function getDefaultServiceWorkerManager(
  config?: MotorServiceWorkerManagerConfig
): MotorServiceWorkerManager {
  if (!defaultManager && config) {
    defaultManager = new MotorServiceWorkerManager(config);
  }
  
  if (!defaultManager) {
    throw new Error('Motor service worker manager not initialized');
  }
  
  return defaultManager;
}

/**
 * Waits for the service worker to be ready
 */
export async function waitForServiceWorker(timeoutMs: number = 30000): Promise<ServiceWorkerRegistration> {
  const env = detectEnvironment();
  
  if (!env.supports_service_worker) {
    throw new Error('Service workers are not supported');
  }

  return new Promise((resolve, reject) => {
    const timeout = setTimeout(() => {
      reject(new Error('Service worker registration timed out'));
    }, timeoutMs);

    navigator.serviceWorker.ready.then((registration) => {
      clearTimeout(timeout);
      resolve(registration);
    }).catch((error) => {
      clearTimeout(timeout);
      reject(error);
    });
  });
}