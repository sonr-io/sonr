/**
 * Vault client module for interacting with the MPC-based vault WASM module
 *
 * This module provides secure key management and cryptographic operations
 * through a WebAssembly-based vault that uses Multi-Party Computation (MPC)
 * for enhanced security.
 *
 * Now includes IPFS integration for distributed enclave data storage.
 */

export * from './types';
export * from './client';
export * from './loader';
export * from './storage';

// IPFS-enhanced components
export * from './client-ipfs';
export * from './enclave';