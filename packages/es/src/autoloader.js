/**
 * @sonr.io/es - ESM Autoloader for Browser
 * 
 * This file automatically loads and exposes all Sonr ES modules
 * for browser usage via script tag or dynamic import.
 * 
 * Usage:
 * <script type="module" src="https://unpkg.com/@sonr.io/es/dist/autoloader.js"></script>
 * 
 * All exports are available on window.Sonr namespace
 */

// Import all modules
import * as auth from './client/auth/index.js';
import * as client from './client/index.js';
import * as codec from './codec/index.js';
import * as wallet from './wallet/index.js';
import * as registry from './registry/index.js';
import * as plugins from './plugins/index.js';

// Import specific auth utilities for convenience
import {
  registerWithPasskey,
  loginWithPasskey,
  isWebAuthnSupported,
  isWebAuthnAvailable,
  isConditionalMediationAvailable,
  bufferToBase64url,
  base64urlToBuffer,
  checkConditionalMediationSupport,
  createRegistrationButton,
  createLoginButton,
  DEFAULT_WEBAUTHN_CONFIG,
  WEBAUTHN_PRESETS
} from './client/auth/webauthn.js';

// Create the main Sonr namespace
const Sonr = {
  // Core modules
  auth,
  client,
  codec,
  wallet,
  registry,
  plugins,
  
  // Convenience shortcuts for common operations
  webauthn: {
    register: registerWithPasskey,
    login: loginWithPasskey,
    isSupported: isWebAuthnSupported,
    isAvailable: isWebAuthnAvailable,
    isConditionalAvailable: isConditionalMediationAvailable,
    bufferToBase64url,
    base64urlToBuffer,
    checkSupport: checkConditionalMediationSupport,
    createRegistrationButton,
    createLoginButton,
    
    // Configuration
    config: DEFAULT_WEBAUTHN_CONFIG,
    presets: WEBAUTHN_PRESETS
  },
  
  // Plugin shortcuts
  motor: plugins.motor,
  vault: plugins.vault,
  
  // Factory functions for plugins
  createMotorPlugin: plugins.createMotorPlugin,
  createVaultClient: plugins.createVaultClient,
  
  // Version info
  version: '0.0.8',
  
  // Initialization function for custom configuration
  init: async (config = {}) => {
    console.log('[Sonr] Initializing with config:', config);
    
    // Initialize Motor plugin if service worker is available
    if ('serviceWorker' in navigator && config.enableMotor !== false) {
      try {
        const motorPlugin = await plugins.createMotorPluginForBrowser({
          wasmUrl: config.motorWasmUrl || '/motor.wasm',
          ...config.motor
        });
        Sonr.motor.instance = motorPlugin;
        console.log('[Sonr] Motor plugin initialized');
      } catch (error) {
        console.warn('[Sonr] Motor plugin initialization failed:', error);
      }
    }
    
    // Initialize Vault client if requested
    if (config.enableVault) {
      try {
        const vaultClient = await plugins.createVaultClient(config.vault);
        Sonr.vault.instance = vaultClient;
        console.log('[Sonr] Vault client initialized');
      } catch (error) {
        console.warn('[Sonr] Vault client initialization failed:', error);
      }
    }
    
    // Check WebAuthn availability
    if (await isWebAuthnAvailable()) {
      console.log('[Sonr] WebAuthn is available');
      Sonr.webauthn.available = true;
      
      // Check for conditional mediation (autofill)
      if (await isConditionalMediationAvailable()) {
        console.log('[Sonr] Conditional mediation (autofill) is available');
        Sonr.webauthn.conditionalAvailable = true;
      }
    }
    
    return Sonr;
  },
  
  // Helper to check if running in browser
  isBrowser: typeof window !== 'undefined',
  
  // Helper to check if running in Node.js
  isNode: typeof process !== 'undefined' && process.versions && process.versions.node,
  
  // Helper to get environment info
  getEnvironment: () => {
    if (typeof window !== 'undefined') {
      return {
        type: 'browser',
        userAgent: navigator.userAgent,
        platform: navigator.platform,
        language: navigator.language,
        online: navigator.onLine,
        serviceWorker: 'serviceWorker' in navigator,
        webauthn: 'credentials' in navigator
      };
    } else if (typeof process !== 'undefined') {
      return {
        type: 'node',
        version: process.version,
        platform: process.platform,
        arch: process.arch
      };
    }
    return { type: 'unknown' };
  }
};

// Auto-initialize with default settings if in browser
if (typeof window !== 'undefined') {
  // Make Sonr globally available
  window.Sonr = Sonr;
  
  // Auto-init on DOMContentLoaded if not already loaded
  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', async () => {
      if (!window.Sonr.initialized) {
        await Sonr.init();
        window.Sonr.initialized = true;
        console.log('[Sonr] Auto-initialized on DOMContentLoaded');
        
        // Dispatch custom event
        window.dispatchEvent(new CustomEvent('sonr:ready', { detail: Sonr }));
      }
    });
  } else {
    // DOM already loaded, init immediately
    (async () => {
      if (!window.Sonr.initialized) {
        await Sonr.init();
        window.Sonr.initialized = true;
        console.log('[Sonr] Auto-initialized (DOM already loaded)');
        
        // Dispatch custom event
        window.dispatchEvent(new CustomEvent('sonr:ready', { detail: Sonr }));
      }
    })();
  }
  
  // Log availability
  console.log('[Sonr] Library loaded. Access via window.Sonr or import modules directly.');
  console.log('[Sonr] Environment:', Sonr.getEnvironment());
}

// Export everything for ES module usage
export default Sonr;
export {
  auth,
  client,
  codec,
  wallet,
  registry,
  plugins,
  // WebAuthn shortcuts
  registerWithPasskey,
  loginWithPasskey,
  isWebAuthnSupported,
  isWebAuthnAvailable,
  isConditionalMediationAvailable,
  bufferToBase64url,
  base64urlToBuffer
};