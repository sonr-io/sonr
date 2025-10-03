/**
 * Unit tests for MPC enclave manager
 */

import { describe, it, expect, beforeEach, vi } from 'vitest';
import { EnclaveIPFSManager, createEnclaveIPFSManager } from '../enclave';
import type { IPFSClient } from '../../../client/services/ipfs';
import type { EnclaveDataWithCID } from '../enclave';

// Mock IPFS client
const mockIPFSClient: IPFSClient = {
  initialize: vi.fn(),
  addEnclaveData: vi.fn().mockResolvedValue({
    cid: 'QmTestCID123',
    size: 100,
    timestamp: Date.now(),
  }),
  getEnclaveData: vi.fn().mockResolvedValue(new Uint8Array([1, 2, 3, 4, 5])),
  verifiedFetch: vi.fn(),
  pin: vi.fn(),
  unpin: vi.fn(),
  isPinned: vi.fn().mockResolvedValue(true),
  listPins: vi.fn().mockResolvedValue(['QmCID1', 'QmCID2']),
  getNodeStatus: vi.fn(),
  isInitialized: vi.fn().mockReturnValue(true),
  cleanup: vi.fn(),
  addString: vi.fn(),
  getString: vi.fn(),
} as any;

describe('EnclaveIPFSManager', () => {
  let manager: EnclaveIPFSManager;
  const validEnclaveData: EnclaveDataWithCID = {
    publicKey: 'public-key-123',
    privateKeyShares: ['share1', 'share2', 'share3'],
    threshold: 2,
    parties: 3,
    encryptionMetadata: {
      algorithm: 'AES-256-GCM',
      keyVersion: 1,
      consensusHeight: 100,
      nonce: 'test-nonce',
    },
  };

  beforeEach(() => {
    vi.clearAllMocks();
    manager = new EnclaveIPFSManager(mockIPFSClient, {
      encryptionRequired: true,
      pinningEnabled: true,
      redundancy: 3,
      maxRetries: 2,
      operationTimeout: 5000,
    });
  });

  describe('storeEnclaveData', () => {
    it('should store encrypted enclave data', async () => {
      const encryptedPayload = new Uint8Array([10, 20, 30]);

      const result = await manager.storeEnclaveData(
        validEnclaveData,
        encryptedPayload
      );

      expect(result.cid).toBe('QmTestCID123');
      expect(result.isPinned).toBe(true);
      expect(result.size).toBe(100);
      expect(mockIPFSClient.addEnclaveData).toHaveBeenCalledWith(encryptedPayload);
      expect(mockIPFSClient.pin).toHaveBeenCalledWith('QmTestCID123');
    });

    it('should require encryption metadata when encryption is required', async () => {
      const dataWithoutMetadata: EnclaveDataWithCID = {
        ...validEnclaveData,
        encryptionMetadata: undefined,
      };

      await expect(
        manager.storeEnclaveData(dataWithoutMetadata, new Uint8Array())
      ).rejects.toThrow('Encryption metadata required for enclave data storage');
    });

    it('should validate enclave data structure', async () => {
      const invalidData: EnclaveDataWithCID = {
        ...validEnclaveData,
        threshold: 5, // Invalid: threshold > parties
      };

      await expect(
        manager.storeEnclaveData(invalidData, new Uint8Array())
      ).rejects.toThrow('Invalid threshold value');
    });

    it('should retry on failure', async () => {
      (mockIPFSClient.addEnclaveData as any)
        .mockRejectedValueOnce(new Error('Network error'))
        .mockResolvedValueOnce({
          cid: 'QmRetryCID',
          size: 50,
          timestamp: Date.now(),
        });

      const result = await manager.storeEnclaveData(
        validEnclaveData,
        new Uint8Array()
      );

      expect(result.cid).toBe('QmRetryCID');
      expect(mockIPFSClient.addEnclaveData).toHaveBeenCalledTimes(2);
    });

    it('should fail after max retries', async () => {
      (mockIPFSClient.addEnclaveData as any).mockRejectedValue(
        new Error('Persistent error')
      );

      await expect(
        manager.storeEnclaveData(validEnclaveData, new Uint8Array())
      ).rejects.toThrow('Failed to store enclave data after 2 attempts');
    });
  });

  describe('retrieveEnclaveData', () => {
    it('should retrieve enclave data', async () => {
      const data = await manager.retrieveEnclaveData('QmTestCID123');

      expect(data).toBeInstanceOf(Uint8Array);
      expect(Array.from(data)).toEqual([1, 2, 3, 4, 5]);
      expect(mockIPFSClient.getEnclaveData).toHaveBeenCalledWith('QmTestCID123');
    });

    it('should retry on failure', async () => {
      (mockIPFSClient.getEnclaveData as any)
        .mockRejectedValueOnce(new Error('Network error'))
        .mockResolvedValueOnce(new Uint8Array([9, 8, 7]));

      const data = await manager.retrieveEnclaveData('QmTestCID123');

      expect(Array.from(data)).toEqual([9, 8, 7]);
      expect(mockIPFSClient.getEnclaveData).toHaveBeenCalledTimes(2);
    });
  });

  describe('verifyEnclaveDataIntegrity', () => {
    it('should verify data integrity successfully', async () => {
      const expectedData = new Uint8Array([1, 2, 3, 4, 5]);
      (mockIPFSClient.getEnclaveData as any).mockResolvedValue(expectedData);

      const isValid = await manager.verifyEnclaveDataIntegrity(
        'QmTestCID123',
        expectedData
      );

      expect(isValid).toBe(true);
    });

    it('should detect data mismatch', async () => {
      const expectedData = new Uint8Array([1, 2, 3]);
      const actualData = new Uint8Array([4, 5, 6]);
      (mockIPFSClient.getEnclaveData as any).mockResolvedValue(actualData);

      const isValid = await manager.verifyEnclaveDataIntegrity(
        'QmTestCID123',
        expectedData
      );

      expect(isValid).toBe(false);
    });

    it('should handle retrieval errors', async () => {
      (mockIPFSClient.getEnclaveData as any).mockRejectedValue(
        new Error('Retrieval failed')
      );

      const isValid = await manager.verifyEnclaveDataIntegrity(
        'QmTestCID123',
        new Uint8Array()
      );

      expect(isValid).toBe(false);
    });
  });

  describe('batch operations', () => {
    it('should batch store multiple enclaves', async () => {
      const enclaves = [
        { data: validEnclaveData, payload: new Uint8Array([1, 2, 3]) },
        { data: validEnclaveData, payload: new Uint8Array([4, 5, 6]) },
      ];

      (mockIPFSClient.addEnclaveData as any)
        .mockResolvedValueOnce({ cid: 'QmCID1', size: 10, timestamp: Date.now() })
        .mockResolvedValueOnce({ cid: 'QmCID2', size: 20, timestamp: Date.now() });

      const results = await manager.batchStoreEnclaves(enclaves);

      expect(results).toHaveLength(2);
      expect(results[0].cid).toBe('QmCID1');
      expect(results[1].cid).toBe('QmCID2');
    });

    it('should continue batch operation even if one fails', async () => {
      const enclaves = [
        { data: validEnclaveData, payload: new Uint8Array([1, 2, 3]) },
        { data: { ...validEnclaveData, threshold: 10 }, payload: new Uint8Array() }, // Invalid
        { data: validEnclaveData, payload: new Uint8Array([4, 5, 6]) },
      ];

      (mockIPFSClient.addEnclaveData as any)
        .mockResolvedValueOnce({ cid: 'QmCID1', size: 10, timestamp: Date.now() })
        .mockResolvedValueOnce({ cid: 'QmCID3', size: 30, timestamp: Date.now() });

      const results = await manager.batchStoreEnclaves(enclaves);

      expect(results).toHaveLength(2);
      expect(results[0].cid).toBe('QmCID1');
      expect(results[1].cid).toBe('QmCID3');
    });
  });

  describe('pin management', () => {
    it('should list pinned enclaves', async () => {
      const pins = await manager.listPinnedEnclaves();

      expect(pins).toEqual(['QmCID1', 'QmCID2']);
      expect(mockIPFSClient.listPins).toHaveBeenCalled();
    });

    it('should remove enclave data', async () => {
      await manager.removeEnclaveData('QmTestCID123');

      expect(mockIPFSClient.isPinned).toHaveBeenCalledWith('QmTestCID123');
      expect(mockIPFSClient.unpin).toHaveBeenCalledWith('QmTestCID123');
    });

    it('should skip unpinning if not pinned', async () => {
      (mockIPFSClient.isPinned as any).mockResolvedValue(false);

      await manager.removeEnclaveData('QmTestCID123');

      expect(mockIPFSClient.unpin).not.toHaveBeenCalled();
    });
  });

  describe('getEnclaveStatus', () => {
    it('should get enclave status', async () => {
      // Set timeout to 0 for instant response
      manager = new EnclaveIPFSManager(mockIPFSClient, {
        encryptionRequired: true,
        pinningEnabled: true,
        redundancy: 3,
        maxRetries: 1,
        operationTimeout: 0,
      });

      // Mock the getEnclaveData to return a valid response
      (mockIPFSClient.getEnclaveData as any).mockResolvedValue(new Uint8Array([1, 2, 3, 4, 5]));
      (mockIPFSClient.isPinned as any).mockResolvedValue(true);

      const status = await manager.getEnclaveStatus('QmTestCID123');

      expect(status.exists).toBe(true);
      expect(status.isPinned).toBe(true);
      expect(status.size).toBe(5);
    });

    it('should handle non-existent enclave', async () => {
      // Set timeout to 0 for instant response
      manager = new EnclaveIPFSManager(mockIPFSClient, {
        encryptionRequired: true,
        pinningEnabled: true,
        redundancy: 3,
        maxRetries: 1,
        operationTimeout: 0,
      });

      (mockIPFSClient.getEnclaveData as any).mockRejectedValue(
        new Error('Not found')
      );
      (mockIPFSClient.isPinned as any).mockResolvedValue(false);

      const status = await manager.getEnclaveStatus('QmNonExistent');

      expect(status.exists).toBe(false);
      expect(status.isPinned).toBe(false);
    });
  });
});

describe('createEnclaveIPFSManager factory', () => {
  it('should create manager with initialized IPFS client', async () => {
    const manager = await createEnclaveIPFSManager(mockIPFSClient);

    expect(manager).toBeInstanceOf(EnclaveIPFSManager);
  });

  it('should throw if IPFS client not initialized', async () => {
    (mockIPFSClient.isInitialized as any).mockReturnValue(false);

    await expect(createEnclaveIPFSManager(mockIPFSClient)).rejects.toThrow(
      'IPFS client must be initialized before creating enclave manager'
    );
  });
});