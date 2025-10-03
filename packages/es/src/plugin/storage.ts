import Dexie, { type Table } from 'dexie';
import type {
  StoredVaultState,
  StoredUCANToken,
  VaultStorageConfig
} from './types';

/**
 * Session data stored in IndexedDB
 */
export interface StoredSession {
  id: string;
  accountAddress: string;
  sessionData: string;
  expiresAt: number;
  createdAt: number;
}

/**
 * Metadata stored in IndexedDB
 */
export interface StoredMetadata {
  id: string;
  accountAddress: string;
  key: string;
  value: string;
  updatedAt: number;
}

/**
 * Account-specific vault database
 */
export class AccountVaultDatabase extends Dexie {
  state!: Table<StoredVaultState>;
  tokens!: Table<StoredUCANToken>;
  sessions!: Table<StoredSession>;
  metadata!: Table<StoredMetadata>;

  constructor(accountAddress: string) {
    super(`vault_${accountAddress}`);

    // Define schema version 1
    this.version(1).stores({
      state: 'id, accountAddress, lastAccessed',
      tokens: 'id, type, issuer, audience, expiresAt, createdAt',
      sessions: 'id, accountAddress, expiresAt, createdAt',
      metadata: 'id, accountAddress, key, updatedAt'
    });
  }
}

/**
 * Manages vault storage for multiple accounts
 */
export class VaultStorageManager {
  private databases: Map<string, AccountVaultDatabase> = new Map();
  private config: VaultStorageConfig;
  private cleanupTimer?: NodeJS.Timeout;

  constructor(config: VaultStorageConfig = {}) {
    this.config = {
      enablePersistence: false,
      autoCleanup: true,
      cleanupInterval: 3600000, // 1 hour
      ...config
    };

    if (this.config.autoCleanup) {
      this.startCleanupTimer();
    }
  }

  /**
   * Get or create database for account
   */
  async getDatabase(accountAddress: string): Promise<AccountVaultDatabase> {
    if (!accountAddress) {
      throw new Error('Account address is required');
    }

    // Return existing database if available
    let db = this.databases.get(accountAddress);
    if (db) {
      return db;
    }

    // Create new database for account
    db = new AccountVaultDatabase(accountAddress);
    await db.open();

    this.databases.set(accountAddress, db);

    // Request persistent storage if configured
    if (this.config.enablePersistence) {
      await this.requestPersistentStorage();
    }

    return db;
  }

  /**
   * Remove database for account
   */
  async removeDatabase(accountAddress: string): Promise<void> {
    const db = this.databases.get(accountAddress);
    if (db) {
      await db.close();
      await db.delete();
      this.databases.delete(accountAddress);
    }
  }

  /**
   * List all persisted accounts
   */
  async listPersistedAccounts(): Promise<string[]> {
    const databases = await Dexie.getDatabaseNames();
    return databases
      .filter(name => name.startsWith('vault_'))
      .map(name => name.replace('vault_', ''));
  }

  /**
   * Request persistent storage from browser
   */
  async requestPersistentStorage(): Promise<boolean> {
    if ('storage' in navigator && 'persist' in navigator.storage) {
      try {
        return await navigator.storage.persist();
      } catch (error) {
        console.warn('Failed to request persistent storage:', error);
        return false;
      }
    }
    return false;
  }

  /**
   * Check if storage is persisted
   */
  async isStoragePersisted(): Promise<boolean> {
    if ('storage' in navigator && 'persisted' in navigator.storage) {
      try {
        return await navigator.storage.persisted();
      } catch (error) {
        console.warn('Failed to check storage persistence:', error);
        return false;
      }
    }
    return false;
  }

  /**
   * Get storage estimate
   */
  async getStorageEstimate(): Promise<StorageEstimate | null> {
    if ('storage' in navigator && 'estimate' in navigator.storage) {
      try {
        return await navigator.storage.estimate();
      } catch (error) {
        console.warn('Failed to get storage estimate:', error);
        return null;
      }
    }
    return null;
  }

  /**
   * Clean up expired tokens and sessions
   */
  async cleanupExpiredData(): Promise<void> {
    const now = Date.now();

    for (const [accountAddress, db] of this.databases.entries()) {
      try {
        // Remove expired tokens
        await db.tokens
          .where('expiresAt')
          .below(now)
          .delete();

        // Remove expired sessions
        await db.sessions
          .where('expiresAt')
          .below(now)
          .delete();

        // Update last accessed time for state
        await db.state.where('accountAddress').equals(accountAddress).modify({
          lastAccessed: now
        });
      } catch (error) {
        console.error(`Cleanup failed for account ${accountAddress}:`, error);
      }
    }
  }

  /**
   * Start automatic cleanup timer
   */
  private startCleanupTimer(): void {
    if (this.cleanupTimer) {
      clearInterval(this.cleanupTimer);
    }

    this.cleanupTimer = setInterval(async () => {
      await this.cleanupExpiredData();
    }, this.config.cleanupInterval!);
  }

  /**
   * Stop cleanup timer
   */
  stopCleanupTimer(): void {
    if (this.cleanupTimer) {
      clearInterval(this.cleanupTimer);
      this.cleanupTimer = undefined;
    }
  }

  /**
   * Close all databases
   */
  async closeAll(): Promise<void> {
    this.stopCleanupTimer();

    for (const db of this.databases.values()) {
      await db.close();
    }

    this.databases.clear();
  }

  /**
   * Try to persist storage without user prompt
   */
  async tryPersistWithoutPromptingUser(): Promise<string> {
    if (!('storage' in navigator) || !('persist' in navigator.storage)) {
      return 'never';
    }

    // Check if already persisted
    const persisted = await navigator.storage.persisted();
    if (persisted) {
      return 'persisted';
    }

    // Try to persist without prompt
    const result = await navigator.storage.persist();
    if (result) {
      return 'persisted';
    }

    return 'prompt';
  }
}

/**
 * Default storage manager instance
 */
export const defaultStorageManager = new VaultStorageManager({
  enablePersistence: true,
  autoCleanup: true
});