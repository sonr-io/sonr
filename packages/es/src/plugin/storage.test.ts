import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { VaultStorageManager, AccountVaultDatabase } from './storage';
import type { StoredVaultState, StoredUCANToken, VaultStorageConfig } from './types';
import Dexie from 'dexie';

describe('VaultStorageManager', () => {
  let storageManager: VaultStorageManager;
  const testAccountAddress = 'sonr1test123abc';

  beforeEach(() => {
    storageManager = new VaultStorageManager({
      enablePersistence: true,
      autoCleanup: false, // Disable auto cleanup for tests
    });
  });

  afterEach(async () => {
    await storageManager.closeAll();
  });

  describe('Database Management', () => {
    it('should create a database for an account', async () => {
      const db = await storageManager.getDatabase(testAccountAddress);
      expect(db).toBeDefined();
      expect(db).toBeInstanceOf(AccountVaultDatabase);
    });

    it('should reuse existing database for the same account', async () => {
      const db1 = await storageManager.getDatabase(testAccountAddress);
      const db2 = await storageManager.getDatabase(testAccountAddress);
      expect(db1).toBe(db2);
    });

    it('should create separate databases for different accounts', async () => {
      const account1 = 'sonr1account1';
      const account2 = 'sonr1account2';

      const db1 = await storageManager.getDatabase(account1);
      const db2 = await storageManager.getDatabase(account2);

      expect(db1).not.toBe(db2);
    });

    it('should throw error when account address is not provided', async () => {
      await expect(storageManager.getDatabase('')).rejects.toThrow('Account address is required');
    });

    it('should remove a database for an account', async () => {
      const db = await storageManager.getDatabase(testAccountAddress);
      expect(db).toBeDefined();

      await storageManager.removeDatabase(testAccountAddress);

      // Getting the database again should create a new instance
      const newDb = await storageManager.getDatabase(testAccountAddress);
      expect(newDb).not.toBe(db);
    });
  });

  describe('Storage Persistence', () => {
    it('should request persistent storage when enabled', async () => {
      const mockPersist = vi.fn().mockResolvedValue(true);
      Object.defineProperty(global, 'navigator', {
        value: {
          storage: {
            persist: mockPersist,
          },
        },
        configurable: true,
      });

      const result = await storageManager.requestPersistentStorage();
      expect(result).toBe(true);
      expect(mockPersist).toHaveBeenCalled();
    });

    it('should handle missing storage API gracefully', async () => {
      Object.defineProperty(global, 'navigator', {
        value: {},
        configurable: true,
      });

      const result = await storageManager.requestPersistentStorage();
      expect(result).toBe(false);
    });

    it('should check if storage is persisted', async () => {
      const mockPersisted = vi.fn().mockResolvedValue(true);
      Object.defineProperty(global, 'navigator', {
        value: {
          storage: {
            persisted: mockPersisted,
          },
        },
        configurable: true,
      });

      const result = await storageManager.isStoragePersisted();
      expect(result).toBe(true);
      expect(mockPersisted).toHaveBeenCalled();
    });

    it('should get storage estimate', async () => {
      const mockEstimate = {
        usage: 1024 * 1024 * 10, // 10MB
        quota: 1024 * 1024 * 100, // 100MB
      };

      Object.defineProperty(global, 'navigator', {
        value: {
          storage: {
            estimate: vi.fn().mockResolvedValue(mockEstimate),
          },
        },
        configurable: true,
      });

      const estimate = await storageManager.getStorageEstimate();
      expect(estimate).toEqual(mockEstimate);
    });
  });

  describe('Cleanup Operations', () => {
    it('should clean up expired data', async () => {
      const db = await storageManager.getDatabase(testAccountAddress);

      // Mock the database tables
      const mockTokensDelete = vi.fn().mockResolvedValue(2);
      const mockSessionsDelete = vi.fn().mockResolvedValue(1);
      const mockStateModify = vi.fn().mockResolvedValue(1);

      db.tokens = {
        where: vi.fn().mockReturnThis(),
        below: vi.fn().mockReturnThis(),
        delete: mockTokensDelete,
      } as any;

      db.sessions = {
        where: vi.fn().mockReturnThis(),
        below: vi.fn().mockReturnThis(),
        delete: mockSessionsDelete,
      } as any;

      db.state = {
        where: vi.fn().mockReturnThis(),
        equals: vi.fn().mockReturnThis(),
        modify: mockStateModify,
      } as any;

      await storageManager.cleanupExpiredData();

      expect(mockTokensDelete).toHaveBeenCalled();
      expect(mockSessionsDelete).toHaveBeenCalled();
      expect(mockStateModify).toHaveBeenCalled();
    });

    it('should close all databases', async () => {
      const db1 = await storageManager.getDatabase('account1');
      const db2 = await storageManager.getDatabase('account2');

      const mockClose1 = vi.fn();
      const mockClose2 = vi.fn();

      db1.close = mockClose1;
      db2.close = mockClose2;

      await storageManager.closeAll();

      expect(mockClose1).toHaveBeenCalled();
      expect(mockClose2).toHaveBeenCalled();
    });
  });

  describe('Persistence Status', () => {
    it('should return "persisted" when storage is already persisted', async () => {
      Object.defineProperty(global, 'navigator', {
        value: {
          storage: {
            persisted: vi.fn().mockResolvedValue(true),
            persist: vi.fn().mockResolvedValue(true),
          },
        },
        configurable: true,
      });

      const status = await storageManager.tryPersistWithoutPromptingUser();
      expect(status).toBe('persisted');
    });

    it('should return "prompt" when persistence requires user interaction', async () => {
      Object.defineProperty(global, 'navigator', {
        value: {
          storage: {
            persisted: vi.fn().mockResolvedValue(false),
            persist: vi.fn().mockResolvedValue(false),
          },
        },
        configurable: true,
      });

      const status = await storageManager.tryPersistWithoutPromptingUser();
      expect(status).toBe('prompt');
    });

    it('should return "never" when storage API is not available', async () => {
      Object.defineProperty(global, 'navigator', {
        value: {},
        configurable: true,
      });

      const status = await storageManager.tryPersistWithoutPromptingUser();
      expect(status).toBe('never');
    });
  });
});

describe('AccountVaultDatabase', () => {
  const testAccountAddress = 'sonr1test123abc';

  it('should create database with correct name', () => {
    const db = new AccountVaultDatabase(testAccountAddress);
    expect(db.name).toBe(`vault_${testAccountAddress}`);
  });

  it('should have correct table definitions', () => {
    const db = new AccountVaultDatabase(testAccountAddress);

    // Verify table properties exist
    expect(db).toHaveProperty('state');
    expect(db).toHaveProperty('tokens');
    expect(db).toHaveProperty('sessions');
    expect(db).toHaveProperty('metadata');
  });
});