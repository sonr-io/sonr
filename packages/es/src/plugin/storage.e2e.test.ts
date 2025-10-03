/**
 * End-to-end tests for Dexie.js integration with VaultClient
 * Tests real IndexedDB functionality without mocks
 */

import { describe, it, expect, beforeEach, afterEach, beforeAll, afterAll, vi } from 'vitest';

// Setup fake IndexedDB for Node.js environment BEFORE importing Dexie
import 'fake-indexeddb/auto';
import FDBFactory from 'fake-indexeddb/lib/FDBFactory';
import FDBKeyRange from 'fake-indexeddb/lib/FDBKeyRange';

// Set up globals before Dexie import
globalThis.indexedDB = new FDBFactory();
globalThis.IDBKeyRange = FDBKeyRange;

import Dexie from 'dexie';
import { VaultStorageManager, AccountVaultDatabase } from './storage';
import { VaultClient, createVaultClient } from './client';
import type { StoredVaultState, StoredUCANToken, VaultStorageConfig } from './types';

describe('VaultClient with Dexie.js Storage - End to End', () => {
  let storageManager: VaultStorageManager;
  let vaultClient: VaultClient;
  const testAccount1 = 'sonr1testaccount123';
  const testAccount2 = 'sonr1testaccount456';

  beforeAll(async () => {
    // Clean up any existing test databases
    const dbs = await Dexie.getDatabaseNames();
    for (const dbName of dbs) {
      if (dbName.startsWith('vault_')) {
        await Dexie.delete(dbName);
      }
    }
  });

  afterAll(async () => {
    // Final cleanup
    const dbs = await Dexie.getDatabaseNames();
    for (const dbName of dbs) {
      if (dbName.startsWith('vault_')) {
        await Dexie.delete(dbName);
      }
    }
  });

  describe('Storage Manager Initialization', () => {
    it('should create storage manager with default config', () => {
      const manager = new VaultStorageManager();
      expect(manager).toBeInstanceOf(VaultStorageManager);
    });

    it('should create storage manager with custom config', () => {
      const config: VaultStorageConfig = {
        enablePersistence: true,
        autoCleanup: false,
        cleanupInterval: 5000,
      };
      const manager = new VaultStorageManager(config);
      expect(manager).toBeInstanceOf(VaultStorageManager);
    });
  });

  describe('Database Lifecycle', () => {
    beforeEach(() => {
      storageManager = new VaultStorageManager({
        autoCleanup: false, // Disable for predictable tests
      });
    });

    afterEach(async () => {
      await storageManager.closeAll();
    });

    it('should create a new database for an account', async () => {
      const db = await storageManager.getDatabase(testAccount1);

      expect(db).toBeInstanceOf(AccountVaultDatabase);
      expect(db.name).toBe(`vault_${testAccount1}`);
      expect(db.isOpen()).toBe(true);
    });

    it('should reuse existing database for same account', async () => {
      const db1 = await storageManager.getDatabase(testAccount1);
      const db2 = await storageManager.getDatabase(testAccount1);

      expect(db1).toBe(db2);
      expect(db1.isOpen()).toBe(true);
    });

    it('should create separate databases for different accounts', async () => {
      const db1 = await storageManager.getDatabase(testAccount1);
      const db2 = await storageManager.getDatabase(testAccount2);

      expect(db1).not.toBe(db2);
      expect(db1.name).toBe(`vault_${testAccount1}`);
      expect(db2.name).toBe(`vault_${testAccount2}`);
    });

    it('should list all persisted accounts', async () => {
      await storageManager.getDatabase(testAccount1);
      await storageManager.getDatabase(testAccount2);

      const accounts = await storageManager.listPersistedAccounts();

      expect(accounts).toContain(testAccount1);
      expect(accounts).toContain(testAccount2);
    });

    it('should remove database for an account', async () => {
      await storageManager.getDatabase(testAccount1);

      let accounts = await storageManager.listPersistedAccounts();
      expect(accounts).toContain(testAccount1);

      await storageManager.removeDatabase(testAccount1);

      accounts = await storageManager.listPersistedAccounts();
      expect(accounts).not.toContain(testAccount1);
    });
  });

  describe('VaultClient Integration', () => {
    beforeEach(async () => {
      // Clean up any existing databases before each test
      const dbs = await Dexie.getDatabaseNames();
      for (const dbName of dbs) {
        if (dbName.startsWith('vault_')) {
          await Dexie.delete(dbName);
        }
      }

      vaultClient = createVaultClient({
        enablePersistence: true,
        autoCleanup: false,
      });
    });

    afterEach(async () => {
      await vaultClient.cleanup();

      // Clean up databases after each test
      const dbs = await Dexie.getDatabaseNames();
      for (const dbName of dbs) {
        if (dbName.startsWith('vault_')) {
          await Dexie.delete(dbName);
        }
      }
    });

    it('should initialize vault with persistence enabled', async () => {
      // Initialize without WASM (will fail but storage should work)
      try {
        await vaultClient.initialize('/fake/path.wasm', testAccount1);
      } catch (error) {
        // Expected to fail due to missing WASM
      }

      // Storage should still be initialized
      const accounts = await vaultClient.listPersistedAccounts();
      expect(accounts).toContain(testAccount1);
    });

    it('should persist and load vault state', async () => {
      try {
        await vaultClient.initialize('/fake/path.wasm', testAccount1);
      } catch (error) {
        // Expected
      }

      // Persist state
      await vaultClient.persistState();

      // Load state
      const state = await vaultClient.loadPersistedState();
      expect(state).toBeDefined();
      expect(state?.accountAddress).toBe(testAccount1);
      expect(state?.isInitialized).toBe(false); // WASM not loaded
    });

    it('should save and retrieve tokens', async () => {
      try {
        await vaultClient.initialize('/fake/path.wasm', testAccount1);
      } catch (error) {
        // Expected
      }

      // Save a mock token
      const mockToken = {
        token: 'eyJ0eXAiOiJKV1QiLCJhbGciOiJFZERTQSJ9...',
        issuer: 'did:sonr:123',
        address: testAccount1,
      };

      await vaultClient.saveToken(mockToken);

      // Retrieve tokens
      const tokens = await vaultClient.getPersistedTokens();
      expect(tokens).toHaveLength(1);
      expect(tokens[0].token).toBe(mockToken.token);
      expect(tokens[0].issuer).toBe(mockToken.issuer);
    });

    it('should clear persisted state', async () => {
      try {
        await vaultClient.initialize('/fake/path.wasm', testAccount1);
      } catch (error) {
        // Expected
      }

      // Add some data
      await vaultClient.persistState();
      await vaultClient.saveToken({
        token: 'test-token',
        issuer: 'did:test',
        address: testAccount1,
      });

      // Verify data exists
      let state = await vaultClient.loadPersistedState();
      let tokens = await vaultClient.getPersistedTokens();
      expect(state).toBeDefined();
      expect(tokens).toHaveLength(1);

      // Clear all data
      await vaultClient.clearPersistedState();

      // Verify data is cleared
      state = await vaultClient.loadPersistedState();
      tokens = await vaultClient.getPersistedTokens();
      expect(state).toBeNull();
      expect(tokens).toHaveLength(0);
    });
  });

  describe('Multi-Account Support', () => {
    beforeEach(async () => {
      // Clean up any existing databases before each test
      const dbs = await Dexie.getDatabaseNames();
      for (const dbName of dbs) {
        if (dbName.startsWith('vault_')) {
          await Dexie.delete(dbName);
        }
      }

      vaultClient = createVaultClient({
        enablePersistence: true,
        autoCleanup: false,
      });
    });

    afterEach(async () => {
      await vaultClient.cleanup();

      // Clean up databases after each test
      const dbs = await Dexie.getDatabaseNames();
      for (const dbName of dbs) {
        if (dbName.startsWith('vault_')) {
          await Dexie.delete(dbName);
        }
      }
    });

    it('should switch between accounts', async () => {
      // Initialize with account 1
      try {
        await vaultClient.initialize('/fake/path.wasm', testAccount1);
      } catch (error) {
        // Expected
      }

      // Save token for account 1
      await vaultClient.saveToken({
        token: 'token-account1',
        issuer: 'did:account1',
        address: testAccount1,
      });

      // Switch to account 2
      await vaultClient.switchAccount(testAccount2);

      // Save token for account 2
      await vaultClient.saveToken({
        token: 'token-account2',
        issuer: 'did:account2',
        address: testAccount2,
      });

      // Verify account 2 has its own token
      let tokens = await vaultClient.getPersistedTokens();
      expect(tokens).toHaveLength(1);
      expect(tokens[0].token).toBe('token-account2');

      // Switch back to account 1
      await vaultClient.switchAccount(testAccount1);

      // Verify account 1 still has its token
      tokens = await vaultClient.getPersistedTokens();
      expect(tokens).toHaveLength(1);
      expect(tokens[0].token).toBe('token-account1');
    });

    it('should maintain separate databases for each account', async () => {
      // Initialize with account 1
      try {
        await vaultClient.initialize('/fake/path.wasm', testAccount1);
      } catch (error) {
        // Expected
      }

      await vaultClient.persistState();
      await vaultClient.saveToken({
        token: 'account1-token',
        issuer: 'did:1',
        address: testAccount1,
      });

      // Switch to account 2
      await vaultClient.switchAccount(testAccount2);
      await vaultClient.persistState();
      await vaultClient.saveToken({
        token: 'account2-token',
        issuer: 'did:2',
        address: testAccount2,
      });

      // List all accounts
      const accounts = await vaultClient.listPersistedAccounts();
      expect(accounts).toContain(testAccount1);
      expect(accounts).toContain(testAccount2);

      // Remove account 1
      await vaultClient.removeAccount(testAccount1);

      // Verify account 1 is removed but account 2 remains
      const remainingAccounts = await vaultClient.listPersistedAccounts();
      expect(remainingAccounts).not.toContain(testAccount1);
      expect(remainingAccounts).toContain(testAccount2);

      // Account 2 data should still be accessible
      const tokens = await vaultClient.getPersistedTokens();
      expect(tokens).toHaveLength(1);
      expect(tokens[0].token).toBe('account2-token');
    });
  });

  describe('Token Expiration and Cleanup', () => {
    let db: AccountVaultDatabase;

    beforeEach(async () => {
      storageManager = new VaultStorageManager({
        autoCleanup: false,
      });
      db = await storageManager.getDatabase(testAccount1);
    });

    afterEach(async () => {
      await storageManager.closeAll();
    });

    it('should store tokens with expiration', async () => {
      const now = Date.now();
      const expiredToken: StoredUCANToken = {
        id: 'token1',
        token: 'expired-token',
        type: 'origin',
        issuer: 'did:expired',
        audience: testAccount1,
        expiresAt: now - 1000, // Expired 1 second ago
        createdAt: now - 10000,
      };

      const validToken: StoredUCANToken = {
        id: 'token2',
        token: 'valid-token',
        type: 'origin',
        issuer: 'did:valid',
        audience: testAccount1,
        expiresAt: now + 10000, // Expires in 10 seconds
        createdAt: now,
      };

      await db.tokens.bulkAdd([expiredToken, validToken]);

      // Verify both tokens are stored
      let tokens = await db.tokens.toArray();
      expect(tokens).toHaveLength(2);

      // Clean up expired tokens
      await storageManager.cleanupExpiredData();

      // Verify only valid token remains
      tokens = await db.tokens.toArray();
      expect(tokens).toHaveLength(1);
      expect(tokens[0].id).toBe('token2');
    });

    it('should clean up expired sessions', async () => {
      const now = Date.now();

      await db.sessions.add({
        id: 'session1',
        accountAddress: testAccount1,
        sessionData: 'expired-session',
        expiresAt: now - 1000,
        createdAt: now - 10000,
      });

      await db.sessions.add({
        id: 'session2',
        accountAddress: testAccount1,
        sessionData: 'valid-session',
        expiresAt: now + 10000,
        createdAt: now,
      });

      // Verify both sessions are stored
      let sessions = await db.sessions.toArray();
      expect(sessions).toHaveLength(2);

      // Clean up expired sessions
      await storageManager.cleanupExpiredData();

      // Verify only valid session remains
      sessions = await db.sessions.toArray();
      expect(sessions).toHaveLength(1);
      expect(sessions[0].id).toBe('session2');
    });
  });

  describe('Storage Persistence API', () => {
    beforeEach(() => {
      storageManager = new VaultStorageManager({
        enablePersistence: true,
      });
    });

    afterEach(async () => {
      await storageManager.closeAll();
    });

    it('should handle storage persistence status', async () => {
      // In test environment, these will return false/never
      const isPersisted = await storageManager.isStoragePersisted();
      expect(typeof isPersisted).toBe('boolean');

      const status = await storageManager.tryPersistWithoutPromptingUser();
      expect(['persisted', 'prompt', 'never']).toContain(status);
    });

    it('should handle storage estimate', async () => {
      const estimate = await storageManager.getStorageEstimate();
      // In test environment, this might return null
      if (estimate) {
        expect(estimate).toHaveProperty('usage');
        expect(estimate).toHaveProperty('quota');
      }
    });
  });

  describe('Database Schema and Tables', () => {
    let db: AccountVaultDatabase;

    beforeEach(async () => {
      db = new AccountVaultDatabase(testAccount1);
      await db.open();
    });

    afterEach(async () => {
      await db.close();
      await db.delete();
    });

    it('should have correct table structure', () => {
      expect(db.state).toBeDefined();
      expect(db.tokens).toBeDefined();
      expect(db.sessions).toBeDefined();
      expect(db.metadata).toBeDefined();
    });

    it('should store and retrieve state', async () => {
      const state: StoredVaultState = {
        id: 'current',
        accountAddress: testAccount1,
        isInitialized: true,
        enclave: JSON.stringify({ test: 'data' }),
        lastAccessed: Date.now(),
        createdAt: Date.now(),
      };

      await db.state.put(state);
      const retrieved = await db.state.get('current');

      expect(retrieved).toBeDefined();
      expect(retrieved?.accountAddress).toBe(testAccount1);
      expect(retrieved?.isInitialized).toBe(true);
    });

    it('should store and retrieve metadata', async () => {
      await db.metadata.add({
        id: 'meta1',
        accountAddress: testAccount1,
        key: 'theme',
        value: 'dark',
        updatedAt: Date.now(),
      });

      const metadata = await db.metadata.toArray();
      expect(metadata).toHaveLength(1);
      expect(metadata[0].key).toBe('theme');
      expect(metadata[0].value).toBe('dark');
    });
  });

  describe('Error Handling', () => {
    beforeEach(() => {
      storageManager = new VaultStorageManager();
    });

    afterEach(async () => {
      await storageManager.closeAll();
    });

    it('should handle invalid account address', async () => {
      await expect(storageManager.getDatabase('')).rejects.toThrow('Account address is required');
      await expect(storageManager.getDatabase(null as any)).rejects.toThrow('Account address is required');
    });

    it('should handle database errors gracefully', async () => {
      const db = await storageManager.getDatabase(testAccount1);

      // Close the database
      await db.close();

      // Try to perform operations on closed database
      try {
        await db.state.toArray();
      } catch (error: any) {
        expect(error).toBeDefined();
        expect(error.message).toContain('closed');
      }
    });
  });

  describe('Backward Compatibility', () => {
    it('should work without persistence enabled', () => {
      const client = createVaultClient({
        enablePersistence: false,
      });

      expect(client).toBeInstanceOf(VaultClient);
      expect(client.isReady()).toBe(false);
    });

    it('should work with default configuration', () => {
      const client = createVaultClient();

      expect(client).toBeInstanceOf(VaultClient);
      expect(client.isReady()).toBe(false);
    });

    it('should not persist when disabled', async () => {
      const client = createVaultClient({
        enablePersistence: false,
      });

      try {
        await client.initialize('/fake/path.wasm');
      } catch (error) {
        // Expected
      }

      // These methods should return empty/null when persistence is disabled
      const state = await client.loadPersistedState();
      const tokens = await client.getPersistedTokens();
      const accounts = await client.listPersistedAccounts();

      expect(state).toBeNull();
      expect(tokens).toHaveLength(0);
      expect(accounts).toHaveLength(0);

      await client.cleanup();
    });
  });
});