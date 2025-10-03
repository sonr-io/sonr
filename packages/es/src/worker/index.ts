/**
 * Motor WASM service worker integration for @sonr.io/es
 * 
 * This module provides TypeScript wrapper methods for the Motor WASM service worker,
 * enabling type-safe interactions with Payment Gateway and OIDC Authorization services.
 * 
 * ## Features
 * 
 * - **Payment Gateway**: W3C Payment Handler API implementation
 * - **OIDC Authorization**: OpenID Connect authentication flows
 * - **Service Worker Management**: Automatic registration and lifecycle handling
 * - **Cross-Environment Support**: Works in browser and Node.js environments
 * 
 * @packageDocumentation
 */

// Type Exports

// Core types
export type {
  MotorPlugin,
  MotorServiceWorkerConfig,
} from './types';

// Plugin-specific config type
export type { MotorPluginConfig } from './plugin';

// Payment types
export type {
  PaymentInstrument,
  PaymentMethod,
  PaymentDetails,
  CanMakePaymentRequest,
  CanMakePaymentResponse,
  PaymentRequestEvent,
  PaymentHandlerResponse,
  ProcessPaymentRequest,
  ProcessPaymentResponse,
  ValidatePaymentMethodRequest,
  ValidatePaymentMethodResponse,
  PaymentStatus,
  RefundPaymentRequest,
  RefundPaymentResponse,
} from './types';

// OIDC types
export type {
  OIDCConfiguration,
  OIDCAuthorizationRequest,
  OIDCAuthorizationResponse,
  OIDCTokenRequest,
  OIDCTokenResponse,
  OIDCUserInfo,
  JWKS,
  JWK,
} from './types';

// Service worker types
export type {
  ServiceWorkerStatus,
  EnvironmentInfo,
  HealthCheckResponse,
  ServiceInfoResponse,
  ErrorResponse,
} from './types';

// Registration types
export type { RegistrationOptions } from './register';

// OIDC client types
export type { OIDCClientConfig } from './oidc';

// Class Exports

// Main plugin implementation
export { MotorPluginImpl } from './plugin';

// Service worker manager
export { MotorServiceWorkerManager } from './worker';

// HTTP client
export { MotorClient } from './client';

// Payment Gateway client
export { PaymentGatewayClient, MotorPaymentHandler, MotorPaymentRequest } from './payment';

// OIDC client
export { OIDCClient } from './oidc';

// Factory Functions

// Main factory functions
export {
  createMotorPlugin,
  createMotorPluginForNode,
  createMotorPluginForBrowser,
} from './plugin';

// Payment factory functions
export {
  createPaymentGatewayClient,
  createPaymentHandler,
  isPaymentHandlerSupported,
  isPaymentRequestSupported,
} from './payment';

// OIDC factory functions
export {
  createOIDCClient,
  isOIDCCallback,
  autoHandleOIDCCallback,
} from './oidc';

// Service worker registration utilities
export {
  registerMotorServiceWorker,
  unregisterMotorServiceWorker,
  getMotorServiceWorkerStatus,
  updateMotorServiceWorker,
  createUpdatePrompt,
  useMotorServiceWorker,
} from './register';

// Service worker lifecycle helpers
export {
  getDefaultServiceWorkerManager,
  waitForServiceWorker,
  detectEnvironment,
} from './worker';

// Default Export

/**
 * Default Motor plugin instance
 */
let defaultPlugin: import('./types').MotorPlugin | null = null;

/**
 * Gets or creates the default Motor plugin instance
 */
export default async function getDefaultMotorPlugin(
  config?: Partial<import('./plugin').MotorPluginConfig>
): Promise<import('./types').MotorPlugin> {
  if (!defaultPlugin) {
    const { createMotorPlugin } = await import('./plugin');
    defaultPlugin = await createMotorPlugin(config);
  }
  return defaultPlugin;
}

// Auto-Registration

/**
 * Automatically register the Motor service worker in browser environments
 */
if (typeof window !== 'undefined' && process.env.MOTOR_DISABLE_AUTO_REGISTER !== 'true') {
  import('./worker').then(({ MotorServiceWorkerManager }) => {
    if (MotorServiceWorkerManager.isSupported()) {
      const manager = new MotorServiceWorkerManager({
        worker_script: '/motor-worker.js',
        scope: '/motor-worker',
        debug: process.env.NODE_ENV === 'development',
      });

      manager.register().catch((error) => {
        console.warn('[Motor] Auto-registration failed:', error);
      });
    }
  }).catch(() => {
    // Silently ignore import errors
  });
}

// Browser Events Setup

/**
 * Set up global event listeners for Motor service worker events
 */
if (typeof window !== 'undefined') {
  window.addEventListener('motor-service-worker-update', ((event: CustomEvent) => {
    console.log('[Motor] Service worker update available:', event.detail);
    
    window.dispatchEvent(new CustomEvent('motor-update-available', {
      detail: event.detail,
    }));
  }) as EventListener);

  window.addEventListener('motor-service-worker-health-failure', ((event: CustomEvent) => {
    console.warn('[Motor] Service worker health check failed:', event.detail);
    
    window.dispatchEvent(new CustomEvent('motor-health-failure', {
      detail: event.detail,
    }));
  }) as EventListener);
}

// Utility Functions

/**
 * Check if Motor is supported in the current environment
 */
export function isMotorSupported(): boolean {
  try {
    const { MotorServiceWorkerManager } = require('./worker');
    const env = MotorServiceWorkerManager.detectEnvironment();
    return env.supports_wasm && (env.supports_service_worker || env.is_node);
  } catch {
    return typeof WebAssembly !== 'undefined';
  }
}

/**
 * Get Motor environment information
 */
export function getMotorEnvironment(): import('./types').EnvironmentInfo {
  try {
    const { MotorServiceWorkerManager } = require('./worker');
    return MotorServiceWorkerManager.detectEnvironment();
  } catch {
    return {
      is_browser: typeof window !== 'undefined',
      is_node: typeof process !== 'undefined',
      supports_service_worker: false,
      supports_wasm: typeof WebAssembly !== 'undefined',
      supports_payment_handler: typeof window !== 'undefined' && 'PaymentManager' in window,
    };
  }
}