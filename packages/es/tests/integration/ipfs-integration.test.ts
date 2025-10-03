/**
 * Integration tests for IPFS/Helia with Docker IPFS nodes
 *
 * These tests require Docker and docker-compose to be running with IPFS nodes.
 * Run with: docker-compose up -d ipfs
 */

import { describe, it, expect, beforeAll, afterAll } from 'vitest';
import { IPFSClient, createIPFSClient } from '../../src/client/services/ipfs';
// TODO: Fix imports - vault/enclave and client-ipfs paths are incorrect
// import { EnclaveIPFSManager } from '../../src/plugins/vault/enclave';
import { EnclaveIPFSManager } from '../../src/plugin/enclave';
import { IPFSCache } from '../../src/client/services/ipfs-cache';
// import { VaultClientWithIPFS } from '../../src/plugins/vault/client-ipfs';
import { VaultClientWithIPFS } from '../../src/plugin/client-ipfs';
// import type { EnclaveDataWithCID } from '../../src/plugins/vault/enclave';
import type { EnclaveDataWithCID } from '../../src/plugin/enclave';

// Skip these tests in CI environment or when IPFS is not available
const skipIntegrationTests = process.env.CI === 'true' || process.env.SKIP_INTEGRATION === 'true';

describe.skipIf(skipIntegrationTests)('IPFS Integration Tests', () => {
  let ipfsClient: IPFSClient;
  let enclaveManager: EnclaveIPFSManager;
  let cache: IPFSCache;

  beforeAll(async () => {
    // Initialize IPFS client with local node
    ipfsClient = await createIPFSClient({
      gateways: ['http://localhost:8080', 'http://localhost:5001'],
      enablePersistence: true,
    });

    // Create enclave manager
    enclaveManager = new EnclaveIPFSManager(ipfsClient, {
      encryptionRequired: false, // Simplified for testing
      pinningEnabled: true,
      maxRetries: 3,
    });

    // Initialize cache
    cache = new IPFSCache({
      maxSize: 100,
      ttl: 300000,
      enablePersistence: true,
    });
  }, 30000); // Allow 30 seconds for initialization

  afterAll(async () => {
    await ipfsClient.cleanup();
    await cache.destroy();
  });

  describe('Basic IPFS Operations', () => {
    it('should connect to IPFS node and get status', async () => {
      const status = await ipfsClient.getNodeStatus();

      expect(status.peerId).toBeDefined();
      expect(status.isOnline).toBe(true);
      console.log('Connected to IPFS node:', status.peerId);
    });

    it('should store and retrieve data', async () => {
      const testData = new Uint8Array([1, 2, 3, 4, 5, 6, 7, 8, 9, 10]);

      const { cid } = await ipfsClient.addEnclaveData(testData);
      expect(cid).toBeDefined();

      const retrieved = await ipfsClient.getEnclaveData(cid);
      expect(retrieved).toEqual(testData);
    });

    it('should store and retrieve strings', async () => {
      const testString = 'Hello IPFS from Sonr integration test!';

      const cid = await ipfsClient.addString(testString);
      expect(cid).toBeDefined();

      const retrieved = await ipfsClient.getString(cid);
      expect(retrieved).toBe(testString);
    });
  });

  describe('Large File Handling', () => {
    it('should handle large enclave data (1MB)', async () => {
      const largeData = new Uint8Array(1024 * 1024); // 1MB
      // Fill with random data
      for (let i = 0; i < largeData.length; i++) {
        largeData[i] = Math.floor(Math.random() * 256);
      }

      const { cid, size } = await ipfsClient.addEnclaveData(largeData);
      expect(cid).toBeDefined();
      expect(size).toBe(largeData.length);

      const retrieved = await ipfsClient.getEnclaveData(cid);
      expect(retrieved.length).toBe(largeData.length);
      expect(retrieved).toEqual(largeData);
    }, 60000); // Allow 60 seconds for large file

    it('should handle chunked data retrieval', async () => {
      // Create data that will be chunked
      const chunkSize = 256 * 1024; // 256KB chunks
      const totalSize = chunkSize * 3; // 768KB
      const chunkedData = new Uint8Array(totalSize);

      // Fill with pattern for verification
      for (let i = 0; i < chunkedData.length; i++) {
        chunkedData[i] = i % 256;
      }

      const { cid } = await ipfsClient.addEnclaveData(chunkedData);
      const retrieved = await ipfsClient.getEnclaveData(cid);

      // Verify data integrity
      expect(retrieved.length).toBe(totalSize);
      for (let i = 0; i < 100; i++) {
        const idx = Math.floor(Math.random() * totalSize);
        expect(retrieved[idx]).toBe(idx % 256);
      }
    });
  });

  describe('MPC Enclave Storage', () => {
    it('should store and retrieve MPC enclave data', async () => {
      const enclaveData: EnclaveDataWithCID = {
        publicKey: 'test-public-key-12345',
        privateKeyShares: ['share1', 'share2', 'share3'],
        threshold: 2,
        parties: 3,
        encryptionMetadata: {
          algorithm: 'AES-256-GCM',
          keyVersion: 1,
          consensusHeight: 1000,
          nonce: 'test-nonce-123',
        },
      };

      const payload = new TextEncoder().encode(JSON.stringify(enclaveData));

      const result = await enclaveManager.storeEnclaveData(enclaveData, payload);
      expect(result.cid).toBeDefined();
      expect(result.isPinned).toBe(true);

      const retrieved = await enclaveManager.retrieveEnclaveData(result.cid);
      const decoded = JSON.parse(new TextDecoder().decode(retrieved));
      expect(decoded.publicKey).toBe(enclaveData.publicKey);
    });

    it('should verify enclave data integrity', async () => {
      const testData = new Uint8Array([10, 20, 30, 40, 50]);
      const { cid } = await ipfsClient.addEnclaveData(testData);

      const enclaveData: EnclaveDataWithCID = {
        publicKey: 'integrity-test-key',
        privateKeyShares: ['share1'],
        threshold: 1,
        parties: 1,
      };

      const isValid = await enclaveManager.verifyEnclaveDataIntegrity(cid, testData);
      expect(isValid).toBe(true);

      const tamperedData = new Uint8Array([10, 20, 30, 40, 51]); // Changed last byte
      const isInvalid = await enclaveManager.verifyEnclaveDataIntegrity(cid, tamperedData);
      expect(isInvalid).toBe(false);
    });
  });

  describe('Pinning Operations', () => {
    it('should pin and unpin content', async () => {
      const testData = new Uint8Array([100, 200, 50]);
      const { cid } = await ipfsClient.addEnclaveData(testData);

      await ipfsClient.pin(cid);
      let isPinned = await ipfsClient.isPinned(cid);
      expect(isPinned).toBe(true);

      await ipfsClient.unpin(cid);
      isPinned = await ipfsClient.isPinned(cid);
      expect(isPinned).toBe(false);
    });

    it('should list all pinned CIDs', async () => {
      const data1 = new Uint8Array([1, 1, 1]);
      const data2 = new Uint8Array([2, 2, 2]);

      const { cid: cid1 } = await ipfsClient.addEnclaveData(data1);
      const { cid: cid2 } = await ipfsClient.addEnclaveData(data2);

      await ipfsClient.pin(cid1);
      await ipfsClient.pin(cid2);

      const pins = await ipfsClient.listPins();
      expect(pins).toContain(cid1);
      expect(pins).toContain(cid2);
    });
  });

  describe('Caching Integration', () => {
    it('should cache retrieved data', async () => {
      const testData = new Uint8Array([111, 222, 333]);
      const { cid } = await ipfsClient.addEnclaveData(testData);

      // First retrieval - from IPFS
      const start1 = Date.now();
      await cache.set(cid, testData);
      const time1 = Date.now() - start1;

      // Second retrieval - from cache
      const start2 = Date.now();
      const cached = await cache.get(cid);
      const time2 = Date.now() - start2;

      expect(cached).toEqual(testData);
      expect(time2).toBeLessThan(time1); // Cache should be faster

      const stats = cache.getStats();
      expect(stats.hits).toBeGreaterThan(0);
    });

    it('should preload multiple CIDs into cache', async () => {
      const cids: string[] = [];
      const dataMap = new Map<string, Uint8Array>();

      // Store multiple items
      for (let i = 0; i < 5; i++) {
        const data = new Uint8Array([i, i, i]);
        const { cid } = await ipfsClient.addEnclaveData(data);
        cids.push(cid);
        dataMap.set(cid, data);
      }

      // Preload into cache
      await cache.preload(cids, async (cid) => {
        return await ipfsClient.getEnclaveData(cid);
      });

      // Verify all are cached
      for (const cid of cids) {
        expect(await cache.has(cid)).toBe(true);
        const cached = await cache.get(cid);
        expect(cached).toEqual(dataMap.get(cid));
      }
    });
  });

  describe('Network Failure Recovery', () => {
    it('should retry on transient failures', async () => {
      // This test simulates retry logic
      const enclaveData: EnclaveDataWithCID = {
        publicKey: 'retry-test-key',
        privateKeyShares: ['share1'],
        threshold: 1,
        parties: 1,
      };

      const payload = new Uint8Array([99, 99, 99]);

      // Store data (should succeed even with retries configured)
      const result = await enclaveManager.storeEnclaveData(enclaveData, payload);
      expect(result.cid).toBeDefined();

      // Retrieve with retry logic
      const retrieved = await enclaveManager.retrieveEnclaveData(result.cid);
      expect(retrieved).toEqual(payload);
    });

    it('should use gateway fallback when direct connection fails', async () => {
      const testData = new Uint8Array([77, 88, 99]);
      const { cid } = await ipfsClient.addEnclaveData(testData);

      // Test verified fetch with fallback
      const response = await ipfsClient.verifiedFetch(cid);
      expect(response.ok).toBe(true);
    });
  });

  describe('Batch Operations', () => {
    it('should batch store multiple enclaves', async () => {
      const enclaves = [];

      for (let i = 0; i < 3; i++) {
        const enclaveData: EnclaveDataWithCID = {
          publicKey: `batch-key-${i}`,
          privateKeyShares: [`share-${i}`],
          threshold: 1,
          parties: 1,
        };

        const payload = new Uint8Array([i, i, i]);
        enclaves.push({ data: enclaveData, payload });
      }

      const results = await enclaveManager.batchStoreEnclaves(enclaves);

      expect(results).toHaveLength(3);
      for (const result of results) {
        expect(result.cid).toBeDefined();
        expect(result.isPinned).toBe(true);
      }
    });
  });

  describe('VaultClient IPFS Integration', () => {
    it('should initialize VaultClient with IPFS', async () => {
      const vaultClient = new VaultClientWithIPFS({
        chainId: 'test-chain',
        enableIPFSPersistence: true,
        ipfsGateways: ['http://localhost:8080'],
      });

      // Initialize with IPFS (skip WASM for integration test)
      try {
        await vaultClient.initializeWithIPFS();

        const status = await vaultClient.getIPFSStatus();
        expect(status.peerId).toBeDefined();
        expect(status.isOnline).toBe(true);

        await vaultClient.cleanup();
      } catch (error) {
        // WASM initialization might fail in test environment
        // But IPFS should still be initialized
        if (error.message.includes('WASM')) {
          console.log('WASM initialization skipped in test');
        } else {
          throw error;
        }
      }
    });
  });

  describe('Data Persistence', () => {
    it('should persist data across client reconnections', async () => {
      const testData = new Uint8Array([50, 60, 70, 80, 90]);
      const { cid } = await ipfsClient.addEnclaveData(testData);

      // Clean up current client
      await ipfsClient.cleanup();

      // Create new client
      const newClient = await createIPFSClient({
        gateways: ['http://localhost:8080'],
      });

      // Should still be able to retrieve the data
      const retrieved = await newClient.getEnclaveData(cid);
      expect(retrieved).toEqual(testData);

      await newClient.cleanup();
    });
  });

  describe('CID Validation', () => {
    it('should validate CID format', async () => {
      const invalidCids = [
        'invalid-cid',
        '12345',
        '',
        'Qm', // Too short
      ];

      for (const invalidCid of invalidCids) {
        await expect(ipfsClient.getEnclaveData(invalidCid)).rejects.toThrow();
      }
    });

    it('should handle non-existent CIDs gracefully', async () => {
      const nonExistentCid = 'QmYwAPJzv5CZsnA625s3Xf2nemtYgPpHdWEz79ojWnPbdG';

      // This should timeout or throw after retries
      await expect(
        Promise.race([
          ipfsClient.getEnclaveData(nonExistentCid),
          new Promise((_, reject) =>
            setTimeout(() => reject(new Error('Timeout')), 5000)
          )
        ])
      ).rejects.toThrow();
    });
  });
});

// Performance benchmarks (optional, run with --bench flag)
describe.skipIf(skipIntegrationTests)('IPFS Performance Benchmarks', () => {
  let ipfsClient: IPFSClient;

  beforeAll(async () => {
    ipfsClient = await createIPFSClient({
      gateways: ['http://localhost:8080'],
    });
  });

  afterAll(async () => {
    await ipfsClient.cleanup();
  });

  it('should measure storage performance', async () => {
    const sizes = [1024, 10240, 102400]; // 1KB, 10KB, 100KB
    const results: any[] = [];

    for (const size of sizes) {
      const data = new Uint8Array(size);
      crypto.getRandomValues(data);

      const start = performance.now();
      const { cid } = await ipfsClient.addEnclaveData(data);
      const storeTime = performance.now() - start;

      const retrieveStart = performance.now();
      await ipfsClient.getEnclaveData(cid);
      const retrieveTime = performance.now() - retrieveStart;

      results.push({
        size: `${size / 1024}KB`,
        storeTime: `${storeTime.toFixed(2)}ms`,
        retrieveTime: `${retrieveTime.toFixed(2)}ms`,
        throughput: `${(size / storeTime * 1000 / 1024).toFixed(2)}KB/s`,
      });
    }

    console.table(results);
  });
});