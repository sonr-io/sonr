/**
 * MPC Enclave manager for IPFS-based vault data operations
 */

import type { IPFSClient } from '../client/services/ipfs'
import { EnclaveData } from './types'

/**
 * Extended enclave data with CID reference
 */
export interface EnclaveDataWithCID extends EnclaveData {
  /** IPFS CID for the encrypted enclave data */
  cid?: string
  /** Encryption metadata for consensus-based encryption */
  encryptionMetadata?: EncryptionMetadata
}

/**
 * Encryption metadata for consensus-based encryption
 */
export interface EncryptionMetadata {
  /** Encryption algorithm used */
  algorithm: string
  /** Version of the encryption key */
  keyVersion: number
  /** Consensus height at encryption time */
  consensusHeight: number
  /** Nonce for encryption */
  nonce: string
}

/**
 * Configuration for enclave storage operations
 */
export interface EnclaveStorageConfig {
  /** Whether encryption is required for all enclave data */
  encryptionRequired: boolean
  /** Enable automatic pinning of enclave data */
  pinningEnabled: boolean
  /** Number of redundant copies to maintain */
  redundancy: number
  /** Maximum retry attempts for failed operations */
  maxRetries: number
  /** Timeout for operations in milliseconds */
  operationTimeout?: number
}

/**
 * Result of enclave storage operation
 */
export interface EnclaveStorageResult {
  /** CID of stored data */
  cid: string
  /** Whether data was pinned */
  isPinned: boolean
  /** Size of stored data */
  size: number
  /** Timestamp of storage */
  timestamp: number
}

/**
 * Manages MPC enclave data operations with IPFS
 */
export class EnclaveIPFSManager {
  private ipfsClient: IPFSClient
  private config: EnclaveStorageConfig

  constructor(
    ipfsClient: IPFSClient,
    config: Partial<EnclaveStorageConfig> = {}
  ) {
    this.ipfsClient = ipfsClient
    this.config = {
      encryptionRequired: true,
      pinningEnabled: true,
      redundancy: 3,
      maxRetries: 3,
      operationTimeout: 30000,
      ...config
    }
  }

  /**
   * Store encrypted enclave data to IPFS
   */
  async storeEnclaveData(
    enclaveData: EnclaveDataWithCID,
    encryptedPayload: Uint8Array
  ): Promise<EnclaveStorageResult> {
    // Validate encryption requirement
    if (this.config.encryptionRequired && !enclaveData.encryptionMetadata) {
      throw new Error('Encryption metadata required for enclave data storage')
    }

    // Validate enclave data structure
    this.validateEnclaveData(enclaveData)

    // Store with retry logic
    let lastError: Error | null = null
    let result: EnclaveStorageResult | null = null

    for (let attempt = 0; attempt < this.config.maxRetries; attempt++) {
      try {
        // Store encrypted data to IPFS
        const { cid, size, timestamp } = await this.ipfsClient.addEnclaveData(encryptedPayload)

        // Pin the content if enabled
        let isPinned = false
        if (this.config.pinningEnabled) {
          await this.ipfsClient.pin(cid)
          isPinned = true
        }

        result = {
          cid,
          isPinned,
          size,
          timestamp
        }

        break
      } catch (error) {
        lastError = error as Error
        console.warn(`Attempt ${attempt + 1} failed:`, error)

        // Exponential backoff
        if (attempt < this.config.maxRetries - 1) {
          await this.delay(Math.pow(2, attempt) * 1000)
        }
      }
    }

    if (!result) {
      throw new Error(
        `Failed to store enclave data after ${this.config.maxRetries} attempts: ${lastError?.message}`
      )
    }

    return result
  }

  /**
   * Retrieve encrypted enclave data from IPFS
   */
  async retrieveEnclaveData(cid: string): Promise<Uint8Array> {
    let lastError: Error | null = null

    for (let attempt = 0; attempt < this.config.maxRetries; attempt++) {
      try {
        // Create a race between operation and timeout
        const timeoutPromise = this.createTimeoutPromise()
        const dataPromise = this.ipfsClient.getEnclaveData(cid)

        const data = await Promise.race([dataPromise, timeoutPromise])

        if (data === null) {
          throw new Error('Operation timed out')
        }

        return data as Uint8Array
      } catch (error) {
        lastError = error as Error
        console.warn(`Retrieval attempt ${attempt + 1} failed:`, error)

        // Exponential backoff
        if (attempt < this.config.maxRetries - 1) {
          await this.delay(Math.pow(2, attempt) * 1000)
        }
      }
    }

    throw new Error(
      `Failed to retrieve enclave data after ${this.config.maxRetries} attempts: ${lastError?.message}`
    )
  }

  /**
   * Verify enclave data integrity by comparing CID
   */
  async verifyEnclaveDataIntegrity(
    cid: string,
    expectedData: Uint8Array
  ): Promise<boolean> {
    try {
      const retrievedData = await this.retrieveEnclaveData(cid)

      // Compare byte arrays
      if (retrievedData.length !== expectedData.length) {
        return false
      }

      for (let i = 0; i < retrievedData.length; i++) {
        if (retrievedData[i] !== expectedData[i]) {
          return false
        }
      }

      return true
    } catch (error) {
      console.error('Failed to verify enclave data integrity:', error)
      return false
    }
  }

  /**
   * Store enclave data with metadata
   */
  async storeEnclaveWithMetadata(
    enclaveData: EnclaveDataWithCID,
    privateKeyShares: Uint8Array[],
    encryptionKey: Uint8Array
  ): Promise<EnclaveStorageResult> {
    // Prepare the complete enclave payload
    const payload = this.prepareEnclavePayload(enclaveData, privateKeyShares)

    // Encrypt the payload
    const encryptedPayload = await this.encryptPayload(payload, encryptionKey)

    // Store to IPFS
    return await this.storeEnclaveData(enclaveData, encryptedPayload)
  }

  /**
   * Batch store multiple enclave data
   */
  async batchStoreEnclaves(
    enclaves: Array<{
      data: EnclaveDataWithCID
      payload: Uint8Array
    }>
  ): Promise<EnclaveStorageResult[]> {
    const results: EnclaveStorageResult[] = []

    for (const enclave of enclaves) {
      try {
        const result = await this.storeEnclaveData(enclave.data, enclave.payload)
        results.push(result)
      } catch (error) {
        console.error('Failed to store enclave:', error)
        // Continue with other enclaves even if one fails
      }
    }

    return results
  }

  /**
   * List all pinned enclave CIDs
   */
  async listPinnedEnclaves(): Promise<string[]> {
    return await this.ipfsClient.listPins()
  }

  /**
   * Remove enclave data from IPFS (unpin)
   */
  async removeEnclaveData(cid: string): Promise<void> {
    try {
      // Check if pinned before unpinning
      const isPinned = await this.ipfsClient.isPinned(cid)
      if (isPinned) {
        await this.ipfsClient.unpin(cid)
      }
    } catch (error) {
      console.error('Failed to remove enclave data:', error)
      throw error
    }
  }

  /**
   * Get enclave storage status
   */
  async getEnclaveStatus(cid: string): Promise<{
    exists: boolean
    isPinned: boolean
    size?: number
  }> {
    try {
      const isPinned = await this.ipfsClient.isPinned(cid)

      // Try to retrieve to check existence
      let exists = false
      let size: number | undefined

      try {
        const data = await this.retrieveEnclaveData(cid)
        exists = true
        size = data.length
      } catch {
        // Data doesn't exist or is not accessible
      }

      return {
        exists,
        isPinned,
        size
      }
    } catch (error) {
      console.error('Failed to get enclave status:', error)
      return {
        exists: false,
        isPinned: false
      }
    }
  }

  /**
   * Validate enclave data structure
   */
  private validateEnclaveData(data: EnclaveDataWithCID): void {
    if (!data.publicKey) {
      throw new Error('Public key is required for enclave data')
    }

    if (!data.privateKeyShares || data.privateKeyShares.length === 0) {
      throw new Error('Private key shares are required for enclave data')
    }

    if (data.threshold < 1 || data.threshold > data.parties) {
      throw new Error('Invalid threshold value')
    }

    if (data.parties !== data.privateKeyShares.length) {
      throw new Error('Number of parties must match number of key shares')
    }
  }

  /**
   * Prepare enclave payload for storage
   */
  private prepareEnclavePayload(
    enclaveData: EnclaveDataWithCID,
    privateKeyShares: Uint8Array[]
  ): Uint8Array {
    // Create JSON representation
    const payload = {
      publicKey: enclaveData.publicKey,
      privateKeyShares: privateKeyShares.map(share =>
        Buffer.from(share).toString('base64')
      ),
      threshold: enclaveData.threshold,
      parties: enclaveData.parties,
      encryptionMetadata: enclaveData.encryptionMetadata
    }

    // Convert to bytes
    const jsonString = JSON.stringify(payload)
    return new TextEncoder().encode(jsonString)
  }

  /**
   * Encrypt payload (placeholder - actual implementation would use consensus keys)
   */
  private async encryptPayload(
    payload: Uint8Array,
    encryptionKey: Uint8Array
  ): Promise<Uint8Array> {
    // TODO: Implement actual consensus-based encryption
    // For now, return the payload as-is
    // In production, this would use the consensus encryption key
    return payload
  }

  /**
   * Create a timeout promise
   */
  private createTimeoutPromise(): Promise<null> {
    if (!this.config.operationTimeout) {
      return new Promise(() => {}) // Never resolves
    }

    return new Promise((resolve) => {
      setTimeout(() => resolve(null), this.config.operationTimeout)
    })
  }

  /**
   * Delay helper for retry logic
   */
  private delay(ms: number): Promise<void> {
    return new Promise(resolve => setTimeout(resolve, ms))
  }
}

/**
 * Factory function to create an enclave manager
 */
export async function createEnclaveIPFSManager(
  ipfsClient: IPFSClient,
  config?: Partial<EnclaveStorageConfig>
): Promise<EnclaveIPFSManager> {
  if (!ipfsClient.isInitialized()) {
    throw new Error('IPFS client must be initialized before creating enclave manager')
  }

  return new EnclaveIPFSManager(ipfsClient, config)
}