/**
 * Type definitions for Vault WASM module
 * Mirrors the Go plugin interface from cmd/vault
 */

/**
 * MPC Enclave data for vault initialization
 */
export interface EnclaveData {
  publicKey: string;
  privateKeyShares: string[];
  threshold: number;
  parties: number;
}

/**
 * Vault configuration options
 */
export interface VaultConfig {
  chainId?: string;
  enclave?: EnclaveData;
  [key: string]: any;
}

/**
 * Request for creating a new origin UCAN token
 */
export interface NewOriginTokenRequest {
  audience_did: string;
  attenuations?: Record<string, any>[];
  facts?: string[];
  not_before?: number;
  expires_at?: number;
}

/**
 * Request for creating a new attenuated UCAN token
 */
export interface NewAttenuatedTokenRequest {
  parent_token: string;
  audience_did: string;
  attenuations?: Record<string, any>[];
  facts?: string[];
  not_before?: number;
  expires_at?: number;
}

/**
 * UCAN token response
 */
export interface UCANTokenResponse {
  token: string;
  issuer: string;
  address: string;
  error?: string;
}

/**
 * Request for signing data
 */
export interface SignDataRequest {
  data: Uint8Array;
}

/**
 * Response from signing data
 */
export interface SignDataResponse {
  signature: Uint8Array;
  error?: string;
}

/**
 * Request for verifying data
 */
export interface VerifyDataRequest {
  data: Uint8Array;
  signature: Uint8Array;
}

/**
 * Response from verifying data
 */
export interface VerifyDataResponse {
  valid: boolean;
  error?: string;
}

/**
 * Response for getting issuer DID
 */
export interface GetIssuerDIDResponse {
  issuer_did: string;
  address: string;
  chain_code: string;
  error?: string;
}

/**
 * Vault plugin interface matching the WASM exports
 */
export interface VaultPlugin {
  newOriginToken(request: NewOriginTokenRequest): Promise<UCANTokenResponse>;
  newAttenuatedToken(request: NewAttenuatedTokenRequest): Promise<UCANTokenResponse>;
  signData(request: SignDataRequest): Promise<SignDataResponse>;
  verifyData(request: VerifyDataRequest): Promise<VerifyDataResponse>;
  getIssuerDID(): Promise<GetIssuerDIDResponse>;
}

/**
 * Error codes for vault operations
 */
export enum VaultErrorCode {
  NOT_INITIALIZED = 'VAULT_NOT_INITIALIZED',
  ALREADY_INITIALIZED = 'VAULT_ALREADY_INITIALIZED',
  LOCKED = 'VAULT_LOCKED',
  KEY_NOT_FOUND = 'KEY_NOT_FOUND',
  INVALID_KEY_TYPE = 'INVALID_KEY_TYPE',
  OPERATION_FAILED = 'OPERATION_FAILED',
  INVALID_PASSPHRASE = 'INVALID_PASSPHRASE',
  WASM_NOT_LOADED = 'WASM_NOT_LOADED',
  TIMEOUT = 'TIMEOUT',
}

/**
 * Vault error class
 */
export class VaultError extends Error {
  constructor(
    public code: VaultErrorCode,
    message: string,
    public details?: any
  ) {
    super(message);
    this.name = 'VaultError';
  }
}

/**
 * Vault event types
 */
export enum VaultEventType {
  INITIALIZED = 'vault:initialized',
  LOCKED = 'vault:locked',
  UNLOCKED = 'vault:unlocked',
  KEY_GENERATED = 'vault:key_generated',
  KEY_DELETED = 'vault:key_deleted',
  EXPORTED = 'vault:exported',
  IMPORTED = 'vault:imported',
  ERROR = 'vault:error',
}

/**
 * Vault event
 */
export interface VaultEvent {
  type: VaultEventType;
  timestamp: number;
  data?: any;
}

/**
 * Vault event listener
 */
export type VaultEventListener = (event: VaultEvent) => void;

/**
 * Storage configuration for vault
 */
export interface VaultStorageConfig {
  enablePersistence?: boolean;
  storageQuotaRequest?: number;
  autoCleanup?: boolean;
  cleanupInterval?: number;
}

/**
 * Enhanced vault configuration with storage options
 */
export interface VaultConfigWithStorage extends VaultConfig, VaultStorageConfig {}

/**
 * Stored vault state in IndexedDB
 */
export interface StoredVaultState {
  id: string;
  accountAddress: string;
  isInitialized: boolean;
  enclave?: string;
  lastAccessed: number;
  createdAt: number;
}

/**
 * Stored UCAN token in IndexedDB
 */
export interface StoredUCANToken {
  id: string;
  token: string;
  type: 'origin' | 'attenuated';
  issuer: string;
  audience: string;
  capabilities?: string;
  expiresAt?: number;
  createdAt: number;
}

/**
 * Storage persistence status
 */
export type StoragePersistenceStatus = 'persisted' | 'prompt' | 'never';

/**
 * Storage statistics
 */
export interface StorageStats {
  accountCount: number;
  tokenCount: number;
  sessionCount: number;
  storageUsed?: number;
  storageQuota?: number;
  isPersisted: boolean;
}

/**
 * IPFS-specific vault configuration
 */
export interface VaultIPFSConfig {
  /** IPFS gateway URLs for fallback */
  ipfsGateways?: string[];
  /** Enable IPFS persistence */
  enableIPFSPersistence?: boolean;
  /** Custom IPFS node configuration */
  ipfsNodeConfig?: any;
}

/**
 * Enhanced vault configuration with IPFS support
 */
export interface VaultConfigWithIPFS extends VaultConfigWithStorage, VaultIPFSConfig {}

/**
 * IPFS-stored enclave reference
 */
export interface IPFSEnclaveReference {
  /** CID of the encrypted enclave data */
  cid: string;
  /** Timestamp when stored */
  storedAt: number;
  /** Whether the data is pinned */
  isPinned: boolean;
  /** Size of the encrypted data */
  size: number;
}

/**
 * Vault state with IPFS references
 */
export interface VaultStateWithIPFS extends StoredVaultState {
  /** IPFS CID references for enclave data */
  ipfsReferences?: IPFSEnclaveReference[];
  /** Last IPFS sync timestamp */
  lastIPFSSync?: number;
}