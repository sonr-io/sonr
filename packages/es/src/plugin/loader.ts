/**
 * WASM loader utilities for the vault module
 */

import { VaultError, VaultErrorCode } from './types';

/**
 * Options for loading WASM module
 */
export interface WASMLoadOptions {
  /** URL to load WASM from */
  url?: string;
  /** Use CDN for loading (jsDelivr) */
  useCDN?: boolean;
  /** Package version for CDN loading */
  version?: string;
  /** Timeout for loading in milliseconds */
  timeout?: number;
}

/**
 * Default CDN configuration
 */
const CDN_BASE = 'https://cdn.jsdelivr.net/npm/@sonr.io/es';
const DEFAULT_TIMEOUT = 30000; // 30 seconds

/**
 * Load vault WASM module
 */
export async function loadVaultWASM(options: WASMLoadOptions = {}): Promise<ArrayBuffer> {
  const {
    url,
    useCDN = false,
    version = 'latest',
    timeout = DEFAULT_TIMEOUT,
  } = options;

  let wasmUrl: string;

  if (url) {
    // Use provided URL
    wasmUrl = url;
  } else if (useCDN) {
    // Use jsDelivr CDN
    wasmUrl = `${CDN_BASE}@${version}/dist/plugins/vault/plugin.wasm`;
  } else {
    // Use local path
    wasmUrl = '/plugin.wasm';
  }

  try {
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), timeout);

    const response = await fetch(wasmUrl, {
      signal: controller.signal,
      headers: {
        'Accept': 'application/wasm',
      },
    });

    clearTimeout(timeoutId);

    if (!response.ok) {
      throw new Error(`Failed to load WASM: ${response.status} ${response.statusText}`);
    }

    const contentType = response.headers.get('content-type');
    if (contentType && !contentType.includes('wasm') && !contentType.includes('octet-stream')) {
      console.warn(`Unexpected content type for WASM: ${contentType}`);
    }

    return await response.arrayBuffer();
  } catch (error: any) {
    if (error.name === 'AbortError') {
      throw new VaultError(
        VaultErrorCode.TIMEOUT,
        `WASM loading timed out after ${timeout}ms`,
        { url: wasmUrl }
      );
    }

    throw new VaultError(
      VaultErrorCode.WASM_NOT_LOADED,
      `Failed to load WASM from ${wasmUrl}: ${error.message}`,
      error
    );
  }
}

/**
 * Verify WASM module is valid
 */
export async function verifyWASM(wasmBuffer: ArrayBuffer): Promise<boolean> {
  try {
    // Check WASM magic number (0x00 0x61 0x73 0x6D)
    const view = new DataView(wasmBuffer);
    const magic = view.getUint32(0, true);
    
    if (magic !== 0x6D736100) {
      throw new Error('Invalid WASM magic number');
    }

    // Try to compile the module
    await WebAssembly.compile(wasmBuffer);
    
    return true;
  } catch (error) {
    console.error('WASM verification failed:', error);
    return false;
  }
}

/**
 * Get WASM module info
 */
export async function getWASMInfo(wasmBuffer: ArrayBuffer): Promise<{
  size: number;
  exports: string[];
  imports: string[];
}> {
  const module = await WebAssembly.compile(wasmBuffer);
  
  const exports = WebAssembly.Module.exports(module).map(exp => exp.name);
  const imports = WebAssembly.Module.imports(module).map(imp => `${imp.module}.${imp.name}`);

  return {
    size: wasmBuffer.byteLength,
    exports,
    imports,
  };
}

/**
 * Cache for loaded WASM modules
 */
class WASMCache {
  private cache = new Map<string, ArrayBuffer>();

  set(key: string, buffer: ArrayBuffer): void {
    this.cache.set(key, buffer);
  }

  get(key: string): ArrayBuffer | undefined {
    return this.cache.get(key);
  }

  has(key: string): boolean {
    return this.cache.has(key);
  }

  clear(): void {
    this.cache.clear();
  }

  remove(key: string): boolean {
    return this.cache.delete(key);
  }
}

/**
 * Global WASM cache instance
 */
export const wasmCache = new WASMCache();

/**
 * Load WASM with caching
 */
export async function loadVaultWASMCached(
  cacheKey: string,
  options: WASMLoadOptions = {}
): Promise<ArrayBuffer> {
  // Check cache first
  if (wasmCache.has(cacheKey)) {
    const cached = wasmCache.get(cacheKey);
    if (cached) {
      return cached;
    }
  }

  // Load WASM
  const wasmBuffer = await loadVaultWASM(options);

  // Verify and cache
  if (await verifyWASM(wasmBuffer)) {
    wasmCache.set(cacheKey, wasmBuffer);
  }

  return wasmBuffer;
}

/**
 * Preload vault WASM module for faster initialization
 */
export async function preloadVaultWASM(options: WASMLoadOptions = {}): Promise<void> {
  try {
    await loadVaultWASMCached('vault-default', options);
  } catch (error) {
    console.error('Failed to preload vault WASM:', error);
  }
}