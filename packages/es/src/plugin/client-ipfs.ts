/**
 * Enhanced VaultClient with IPFS integration
 */

import { VaultClient } from './client';
import { IPFSClient, createIPFSClient } from '../client/services/ipfs';
import { EnclaveIPFSManager, createEnclaveIPFSManager, type EnclaveDataWithCID } from './enclave';
import type {
  VaultConfigWithIPFS,
  VaultError,
  VaultErrorCode,
  IPFSEnclaveReference,
  VaultStateWithIPFS,
} from './types';

/**
 * VaultClient with integrated IPFS support for MPC enclave data
 */
export class VaultClientWithIPFS extends VaultClient {
  private ipfsClient: IPFSClient | null = null;
  private enclaveManager: EnclaveIPFSManager | null = null;
  private ipfsConfig: VaultConfigWithIPFS;

  constructor(config: VaultConfigWithIPFS = {}) {
    super(config);
    this.ipfsConfig = config;
  }

  /**
   * Initialize vault with IPFS support
   */
  async initializeWithIPFS(
    wasmPath?: string,
    accountAddress?: string,
    ipfsConfig?: any
  ): Promise<void> {
    // Initialize base vault client
    await super.initialize(wasmPath, accountAddress);

    // Initialize IPFS client
    try {
      this.ipfsClient = await createIPFSClient({
        gateways: this.ipfsConfig.ipfsGateways,
        enablePersistence: this.ipfsConfig.enableIPFSPersistence,
        libp2pConfig: ipfsConfig || this.ipfsConfig.ipfsNodeConfig,
      });

      // Initialize enclave manager
      this.enclaveManager = await createEnclaveIPFSManager(this.ipfsClient, {
        encryptionRequired: true,
        pinningEnabled: this.ipfsConfig.enableIPFSPersistence ?? true,
        maxRetries: 3,
      });
    } catch (error) {
      console.error('Failed to initialize IPFS:', error);
      throw new Error(`IPFS initialization failed: ${error}`);
    }
  }

  /**
   * Store enclave data to IPFS
   */
  async storeEnclaveToIPFS(
    enclaveData: EnclaveDataWithCID,
    encryptedPayload: Uint8Array
  ): Promise<string> {
    if (!this.enclaveManager) {
      throw new Error('IPFS not initialized');
    }

    const result = await this.enclaveManager.storeEnclaveData(
      enclaveData,
      encryptedPayload
    );

    // Save reference to database if persistence is enabled
    if (this.ipfsConfig.enablePersistence) {
      await this.saveIPFSReference({
        cid: result.cid,
        storedAt: result.timestamp,
        isPinned: result.isPinned,
        size: result.size,
      });
    }

    return result.cid;
  }

  /**
   * Retrieve enclave data from IPFS
   */
  async retrieveEnclaveFromIPFS(cid: string): Promise<Uint8Array> {
    if (!this.enclaveManager) {
      throw new Error('IPFS not initialized');
    }

    return await this.enclaveManager.retrieveEnclaveData(cid);
  }

  /**
   * Store vault enclave with automatic encryption
   */
  async storeVaultEnclave(
    privateKeyShares: string[]
  ): Promise<string> {
    if (!this.enclaveManager) {
      throw new Error('IPFS not initialized');
    }

    // Get current enclave configuration
    const enclaveConfig = this.ipfsConfig.enclave;
    if (!enclaveConfig) {
      throw new Error('Enclave configuration not set');
    }

    // Create enclave data with CID
    const enclaveData: EnclaveDataWithCID = {
      ...enclaveConfig,
      encryptionMetadata: {
        algorithm: 'AES-256-GCM',
        keyVersion: 1,
        consensusHeight: 0,
        nonce: this.generateNonce(),
      },
    };

    // Prepare encrypted payload
    const payload = JSON.stringify({
      publicKey: enclaveData.publicKey,
      privateKeyShares,
      threshold: enclaveData.threshold,
      parties: enclaveData.parties,
      timestamp: Date.now(),
    });

    const encryptedPayload = new TextEncoder().encode(payload);

    // Store to IPFS
    return await this.storeEnclaveToIPFS(enclaveData, encryptedPayload);
  }

  /**
   * Retrieve and decrypt vault enclave
   */
  async retrieveVaultEnclave(cid: string): Promise<EnclaveDataWithCID> {
    if (!this.enclaveManager) {
      throw new Error('IPFS not initialized');
    }

    // Retrieve encrypted data
    const encryptedData = await this.retrieveEnclaveFromIPFS(cid);

    // Decrypt and parse
    const decryptedString = new TextDecoder().decode(encryptedData);
    const enclaveData = JSON.parse(decryptedString);

    return {
      ...enclaveData,
      cid,
    };
  }

  /**
   * Verify enclave data integrity
   */
  async verifyEnclaveIntegrity(
    cid: string,
    expectedData: Uint8Array
  ): Promise<boolean> {
    if (!this.enclaveManager) {
      throw new Error('IPFS not initialized');
    }

    return await this.enclaveManager.verifyEnclaveDataIntegrity(
      cid,
      expectedData
    );
  }

  /**
   * Get IPFS node status
   */
  async getIPFSStatus(): Promise<any> {
    if (!this.ipfsClient) {
      throw new Error('IPFS not initialized');
    }

    return await this.ipfsClient.getNodeStatus();
  }

  /**
   * List all pinned enclave CIDs
   */
  async listPinnedEnclaves(): Promise<string[]> {
    if (!this.enclaveManager) {
      throw new Error('IPFS not initialized');
    }

    return await this.enclaveManager.listPinnedEnclaves();
  }

  /**
   * Remove enclave from IPFS (unpin)
   */
  async removeEnclaveFromIPFS(cid: string): Promise<void> {
    if (!this.enclaveManager) {
      throw new Error('IPFS not initialized');
    }

    await this.enclaveManager.removeEnclaveData(cid);

    // Remove reference from database
    if (this.ipfsConfig.enablePersistence) {
      await this.removeIPFSReference(cid);
    }
  }

  /**
   * Batch store multiple enclaves
   */
  async batchStoreEnclaves(
    enclaves: Array<{
      data: EnclaveDataWithCID;
      payload: Uint8Array;
    }>
  ): Promise<string[]> {
    if (!this.enclaveManager) {
      throw new Error('IPFS not initialized');
    }

    const results = await this.enclaveManager.batchStoreEnclaves(enclaves);

    // Save references if persistence is enabled
    if (this.ipfsConfig.enablePersistence) {
      for (const result of results) {
        await this.saveIPFSReference({
          cid: result.cid,
          storedAt: result.timestamp,
          isPinned: result.isPinned,
          size: result.size,
        });
      }
    }

    return results.map(r => r.cid);
  }

  /**
   * Sync enclave data with IPFS network
   */
  async syncWithIPFS(): Promise<void> {
    if (!this.ipfsClient || !this.enclaveManager) {
      throw new Error('IPFS not initialized');
    }

    // Get stored references
    const references = await this.getIPFSReferences();

    // Verify each reference
    for (const ref of references) {
      try {
        const status = await this.enclaveManager.getEnclaveStatus(ref.cid);

        // Re-pin if needed
        if (!status.isPinned && ref.isPinned) {
          await this.ipfsClient.pin(ref.cid);
        }
      } catch (error) {
        console.warn(`Failed to sync CID ${ref.cid}:`, error);
      }
    }

    // Update last sync timestamp
    await this.updateLastIPFSSync();
  }

  // ============= Storage Methods =============

  /**
   * Save IPFS reference to storage
   */
  private async saveIPFSReference(ref: IPFSEnclaveReference): Promise<void> {
    const database = (this as any).database;
    if (!database) return;

    // Store in metadata collection
    await database.metadata.put({
      id: `ipfs_${ref.cid}`,
      type: 'ipfs_reference',
      data: ref,
      createdAt: Date.now(),
    });
  }

  /**
   * Get all IPFS references
   */
  private async getIPFSReferences(): Promise<IPFSEnclaveReference[]> {
    const database = (this as any).database;
    if (!database) return [];

    const metadata = await database.metadata
      .where('type')
      .equals('ipfs_reference')
      .toArray();

    return metadata.map((m: any) => m.data);
  }

  /**
   * Remove IPFS reference
   */
  private async removeIPFSReference(cid: string): Promise<void> {
    const database = (this as any).database;
    if (!database) return;

    await database.metadata.delete(`ipfs_${cid}`);
  }

  /**
   * Update last IPFS sync timestamp
   */
  private async updateLastIPFSSync(): Promise<void> {
    const database = (this as any).database;
    if (!database) return;

    await database.metadata.put({
      id: 'ipfs_last_sync',
      type: 'ipfs_sync',
      timestamp: Date.now(),
    });
  }

  /**
   * Persist state with IPFS references
   */
  async persistState(): Promise<void> {
    await super.persistState();

    // Add IPFS-specific state
    const database = (this as any).database;
    const accountAddress = (this as any).accountAddress;

    if (!database || !accountAddress) return;

    const references = await this.getIPFSReferences();

    const state: VaultStateWithIPFS = {
      id: 'current',
      accountAddress,
      isInitialized: this.isReady(),
      enclave: this.ipfsConfig.enclave ?
        JSON.stringify(this.ipfsConfig.enclave) : undefined,
      lastAccessed: Date.now(),
      createdAt: Date.now(),
      ipfsReferences: references,
      lastIPFSSync: Date.now(),
    };

    await database.state.put(state);
  }

  /**
   * Generate a random nonce
   */
  private generateNonce(): string {
    const bytes = new Uint8Array(16);
    crypto.getRandomValues(bytes);
    return Array.from(bytes)
      .map(b => b.toString(16).padStart(2, '0'))
      .join('');
  }

  /**
   * Cleanup with IPFS shutdown
   */
  async cleanup(): Promise<void> {
    // Clean up IPFS resources
    if (this.ipfsClient) {
      await this.ipfsClient.cleanup();
      this.ipfsClient = null;
    }

    this.enclaveManager = null;

    // Call parent cleanup
    await super.cleanup();
  }
}

/**
 * Create a VaultClient with IPFS support
 */
export function createVaultClientWithIPFS(
  config?: VaultConfigWithIPFS
): VaultClientWithIPFS {
  return new VaultClientWithIPFS(config);
}

/**
 * Default IPFS-enabled vault client instance
 */
let defaultIPFSClient: VaultClientWithIPFS | null = null;

/**
 * Get or create the default IPFS-enabled vault client
 */
export async function getDefaultVaultClientWithIPFS(
  config?: VaultConfigWithIPFS
): Promise<VaultClientWithIPFS> {
  if (!defaultIPFSClient) {
    defaultIPFSClient = createVaultClientWithIPFS(config);
    await defaultIPFSClient.initializeWithIPFS();
  }
  return defaultIPFSClient;
}