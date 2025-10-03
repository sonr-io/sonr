/**
 * Motor Service Worker
 * Payment Gateway & OpenID Connect Authorization Beacon
 * 
 * This service worker uses go-wasm-http-server to embed a Go HTTP server
 * that handles payment processing and OIDC authentication flows.
 */

// Import wasm_exec.js for Go 1.23.4 and go-wasm-http-server
importScripts('https://cdn.jsdelivr.net/gh/golang/go@go1.23.4/misc/wasm/wasm_exec.js');
importScripts('https://cdn.jsdelivr.net/gh/nlepage/go-wasm-http-server@v2.2.1/sw.js');

// Configuration
const WASM_MODULE_URL = '/motor.wasm';
const CACHE_NAME = 'motor-gateway-v1';

/**
 * Service Worker lifecycle - Install event
 */
self.addEventListener('install', (event) => {
  console.log('[Motor Gateway] Installing service worker...');
  
  event.waitUntil(
    (async () => {
      try {
        // Pre-cache the WASM module
        const cache = await caches.open(CACHE_NAME);
        await cache.add(WASM_MODULE_URL);
        
        console.log('[Motor Gateway] WASM module cached');
        
        // Skip waiting to activate immediately
        await self.skipWaiting();
      } catch (error) {
        console.error('[Motor Gateway] Installation failed:', error);
      }
    })()
  );
});

/**
 * Service Worker lifecycle - Activate event
 */
self.addEventListener('activate', (event) => {
  console.log('[Motor Gateway] Activating service worker...');
  
  event.waitUntil(
    (async () => {
      // Clean up old caches
      const cacheNames = await caches.keys();
      await Promise.all(
        cacheNames
          .filter(name => name.startsWith('motor-') && name !== CACHE_NAME)
          .map(name => caches.delete(name))
      );
      
      // Take control of all clients
      await self.clients.claim();
      
      console.log('[Motor Gateway] Service worker activated');
    })()
  );
});

/**
 * Register the Motor WASM HTTP server
 * This handles all HTTP requests through the Go WASM module
 */
registerWasmHTTPListener(WASM_MODULE_URL, {
  // Base path for API endpoints
  base: '/',
  
  // Cache strategies for different endpoints
  cacheStrategies: {
    // Health endpoints - always fetch fresh
    '/health': 'network-first',
    '/status': 'network-first',
    
    // OIDC discovery - cache for performance
    '/.well-known/openid-configuration': 'cache-first',
    '/.well-known/jwks.json': 'cache-first',
    
    // Payment endpoints - always fetch fresh for security
    '/api/payment/process': 'network-only',
    '/api/payment/validate': 'network-only',
    '/api/payment/status': 'network-first',
    '/api/payment/refund': 'network-only',
    
    // Auth endpoints - no caching for security
    '/authorize': 'network-only',
    '/token': 'network-only',
    '/userinfo': 'network-only',
  }
});

/**
 * Handle messages from clients
 */
self.addEventListener('message', async (event) => {
  const { type, data } = event.data;
  
  try {
    let result;
    
    switch (type) {
      case 'HEALTH_CHECK':
        // Check health via the WASM handler
        try {
          const response = await fetch('/health');
          result = await response.json();
        } catch (error) {
          result = { 
            status: 'error', 
            error: error.message,
            service: 'motor-gateway'
          };
        }
        break;
        
      case 'CLEAR_CACHE':
        // Clear the cache
        await caches.delete(CACHE_NAME);
        result = { success: true, message: 'Cache cleared' };
        break;
        
      case 'SKIP_WAITING':
        // Skip waiting for activation
        await self.skipWaiting();
        result = { success: true };
        break;
        
      default:
        result = { error: `Unknown message type: ${type}` };
    }
    
    // Send response back to client
    if (event.ports && event.ports[0]) {
      event.ports[0].postMessage(result);
    }
  } catch (error) {
    if (event.ports && event.ports[0]) {
      event.ports[0].postMessage({
        error: error.message || 'Service worker error',
      });
    }
  }
});

console.log('[Motor Gateway] Service worker loaded');