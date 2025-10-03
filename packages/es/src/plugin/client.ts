import { createPlugin, Plugin } from '@extism/extism';
import {
  VaultError,
  VaultErrorCode,
} from './types';
import type {
  VaultConfig,
  VaultPlugin,
  NewOriginTokenRequest,
  NewAttenuatedTokenRequest,
  SignDataRequest,
  VerifyDataRequest,
  UCANTokenResponse,
  SignDataResponse,
  VerifyDataResponse,
  GetIssuerDIDResponse,
  VaultConfigWithStorage,
  StoredVaultState,
  StoredUCANToken,
} from './types';
import { VaultStorageManager } from './storage';
import type { AccountVaultDatabase } from './storage';

/**
 * Vault client for interacting with the WASM vault module
 */
export class VaultClient implements VaultPlugin {
  private plugin: Plugin | null = null;
  private config: VaultConfigWithStorage;
  private wasmModule: ArrayBuffer | null = null;
  private storageManager: VaultStorageManager | null = null;
  private database: any | null = null;
  private accountAddress: string | null = null;

  constructor(config: VaultConfigWithStorage = {}) {
    this.config = config;

    // Initialize storage manager if persistence is enabled
    if (config.enablePersistence) {
      this.storageManager = new VaultStorageManager(config);
    }
  }

  /**
   * Initialize the vault with WASM module
   */
  async initialize(wasmPath?: string, accountAddress?: string): Promise<void> {
    // Initialize storage first if account address is provided and persistence is enabled
    // This ensures storage works even if WASM loading fails
    if (accountAddress && this.config.enablePersistence && this.storageManager) {
      this.accountAddress = accountAddress;
      this.database = await this.storageManager.getDatabase(accountAddress);
      await this.loadPersistedState();
    }

    try {
      // Load WASM module
      if (wasmPath) {
        // Load from provided path
        const response = await fetch(wasmPath);
        this.wasmModule = await response.arrayBuffer();
      } else {
        // Load from default location
        const response = await fetch('/plugin.wasm');
        this.wasmModule = await response.arrayBuffer();
      }

      // Create Extism plugin with configuration
      const pluginConfig = {
        wasm: [{ data: new Uint8Array(this.wasmModule) }],
        config: this.prepareConfig(),
      };

      this.plugin = await createPlugin(pluginConfig, {
        useWasi: true,
      });

    } catch (error) {
      throw new VaultError(
        VaultErrorCode.WASM_NOT_LOADED,
        `Failed to initialize vault: ${error}`,
        error
      );
    }
  }

  /**
   * Prepare configuration for the plugin
   */
  private prepareConfig(): Record<string, string> {
    const config: Record<string, string> = {};
    
    if (this.config.chainId) {
      config.chain_id = this.config.chainId;
    }

    return config;
  }

  /**
   * Ensure plugin is initialized
   */
  private ensureInitialized(): void {
    if (!this.plugin) {
      throw new VaultError(
        VaultErrorCode.NOT_INITIALIZED,
        'Vault client not initialized. Call initialize() first.'
      );
    }
  }

  /**
   * Convert JavaScript object to JSON for plugin input
   */
  private toPluginInput(data: any): Uint8Array {
    const json = JSON.stringify(data);
    return new TextEncoder().encode(json);
  }

  /**
   * Parse plugin output as JSON
   */
  private parsePluginOutput<T>(output: any): T {
    if (!output) {
      throw new VaultError(
        VaultErrorCode.OPERATION_FAILED,
        'No output from plugin'
      );
    }

    // Handle both Uint8Array and PluginOutput types
    let text: string;
    if (output instanceof Uint8Array) {
      text = new TextDecoder().decode(output);
    } else if (output.bytes) {
      // PluginOutput type from Extism
      text = new TextDecoder().decode(output.bytes());
    } else if (output.text) {
      text = output.text();
    } else {
      text = output.toString();
    }
    
    return JSON.parse(text) as T;
  }

  /**
   * Create a new origin UCAN token
   */
  async newOriginToken(request: NewOriginTokenRequest): Promise<UCANTokenResponse> {
    this.ensureInitialized();

    try {
      const input = this.toPluginInput(request);
      const output = await this.plugin!.call('new_origin_token', input);
      const response = this.parsePluginOutput<UCANTokenResponse>(output);

      if (response.error) {
        throw new VaultError(
          VaultErrorCode.OPERATION_FAILED,
          response.error
        );
      }

      // Save token if persistence is enabled
      if (this.config.enablePersistence && this.database) {
        await this.saveToken(response);
      }

      return response;
    } catch (error) {
      if (error instanceof VaultError) {
        throw error;
      }
      throw new VaultError(
        VaultErrorCode.OPERATION_FAILED,
        `Failed to create origin token: ${error}`,
        error
      );
    }
  }

  /**
   * Create a new attenuated UCAN token
   */
  async newAttenuatedToken(request: NewAttenuatedTokenRequest): Promise<UCANTokenResponse> {
    this.ensureInitialized();

    try {
      const input = this.toPluginInput(request);
      const output = await this.plugin!.call('new_attenuated_token', input);
      const response = this.parsePluginOutput<UCANTokenResponse>(output);

      if (response.error) {
        throw new VaultError(
          VaultErrorCode.OPERATION_FAILED,
          response.error
        );
      }

      // Save token if persistence is enabled
      if (this.config.enablePersistence && this.database) {
        await this.saveToken(response);
      }

      return response;
    } catch (error) {
      if (error instanceof VaultError) {
        throw error;
      }
      throw new VaultError(
        VaultErrorCode.OPERATION_FAILED,
        `Failed to create attenuated token: ${error}`,
        error
      );
    }
  }

  /**
   * Sign data with the vault's MPC enclave
   */
  async signData(request: SignDataRequest): Promise<SignDataResponse> {
    this.ensureInitialized();

    try {
      const input = this.toPluginInput({
        data: Array.from(request.data),
      });
      const output = await this.plugin!.call('sign_data', input);
      const response = this.parsePluginOutput<any>(output);

      if (response.error) {
        throw new VaultError(
          VaultErrorCode.OPERATION_FAILED,
          response.error
        );
      }

      return {
        signature: new Uint8Array(response.signature),
        error: response.error,
      };
    } catch (error) {
      if (error instanceof VaultError) {
        throw error;
      }
      throw new VaultError(
        VaultErrorCode.OPERATION_FAILED,
        `Failed to sign data: ${error}`,
        error
      );
    }
  }

  /**
   * Verify a signature with the vault's MPC enclave
   */
  async verifyData(request: VerifyDataRequest): Promise<VerifyDataResponse> {
    this.ensureInitialized();

    try {
      const input = this.toPluginInput({
        data: Array.from(request.data),
        signature: Array.from(request.signature),
      });
      const output = await this.plugin!.call('verify_data', input);
      const response = this.parsePluginOutput<VerifyDataResponse>(output);

      if (response.error) {
        throw new VaultError(
          VaultErrorCode.OPERATION_FAILED,
          response.error
        );
      }

      return response;
    } catch (error) {
      if (error instanceof VaultError) {
        throw error;
      }
      throw new VaultError(
        VaultErrorCode.OPERATION_FAILED,
        `Failed to verify data: ${error}`,
        error
      );
    }
  }

  /**
   * Get the issuer DID and address from the vault
   */
  async getIssuerDID(): Promise<GetIssuerDIDResponse> {
    this.ensureInitialized();

    try {
      const output = await this.plugin!.call('get_issuer_did', new Uint8Array());
      const response = this.parsePluginOutput<GetIssuerDIDResponse>(output);

      if (response.error) {
        throw new VaultError(
          VaultErrorCode.OPERATION_FAILED,
          response.error
        );
      }

      return response;
    } catch (error) {
      if (error instanceof VaultError) {
        throw error;
      }
      throw new VaultError(
        VaultErrorCode.OPERATION_FAILED,
        `Failed to get issuer DID: ${error}`,
        error
      );
    }
  }

  /**
   * Check if the vault is ready
   */
  isReady(): boolean {
    return this.plugin !== null;
  }

  // ============= Storage Management Methods =============

  /**
   * Persist current vault state
   */
  async persistState(): Promise<void> {
    if (!this.database || !this.accountAddress) return;

    const state: StoredVaultState = {
      id: 'current',
      accountAddress: this.accountAddress,
      isInitialized: this.isReady(),
      enclave: this.config.enclave ? JSON.stringify(this.config.enclave) : undefined,
      lastAccessed: Date.now(),
      createdAt: Date.now(),
    };

    await this.database.state.put(state);
  }

  /**
   * Load persisted vault state
   */
  async loadPersistedState(): Promise<StoredVaultState | null> {
    if (!this.database) return null;

    const state = await this.database.state.get('current');
    if (state && state.enclave) {
      // Restore enclave configuration if present
      this.config.enclave = JSON.parse(state.enclave);
    }
    return state || null;
  }

  /**
   * Clear persisted vault state
   */
  async clearPersistedState(): Promise<void> {
    if (!this.database) return;

    await this.database.state.clear();
    await this.database.tokens.clear();
    await this.database.sessions.clear();
    await this.database.metadata.clear();
  }

  // ============= Token Management Methods =============

  /**
   * Save UCAN token to storage
   */
  async saveToken(token: UCANTokenResponse): Promise<void> {
    if (!this.database) return;

    const storedToken: StoredUCANToken = {
      id: `${Date.now()}_${Math.random()}`,
      token: token.token,
      type: 'origin', // Default to origin, can be enhanced
      issuer: token.issuer,
      audience: token.address,
      createdAt: Date.now(),
    };

    await this.database.tokens.put(storedToken);
  }

  /**
   * Get all persisted tokens
   */
  async getPersistedTokens(): Promise<StoredUCANToken[]> {
    if (!this.database) return [];

    const tokens = await this.database.tokens.toArray();
    return tokens || [];
  }

  /**
   * Remove expired tokens
   */
  async removeExpiredTokens(): Promise<void> {
    if (!this.database) return;

    const now = Date.now();
    await this.database.tokens
      .where('expiresAt')
      .below(now)
      .delete();
  }

  // ============= Account Management Methods =============

  /**
   * Switch to a different account
   */
  async switchAccount(newAccountAddress: string): Promise<void> {
    if (!this.storageManager) {
      throw new VaultError(
        VaultErrorCode.NOT_INITIALIZED,
        'Storage manager not initialized'
      );
    }

    // Save current state before switching
    if (this.accountAddress && this.database) {
      await this.persistState();
    }

    // Switch to new account database
    this.accountAddress = newAccountAddress;
    this.database = await this.storageManager.getDatabase(newAccountAddress);

    // Load new account state
    await this.loadPersistedState();
  }

  /**
   * List all persisted accounts
   */
  async listPersistedAccounts(): Promise<string[]> {
    if (!this.storageManager) return [];

    return await this.storageManager.listPersistedAccounts();
  }

  /**
   * Remove an account and its data
   */
  async removeAccount(accountAddress: string): Promise<void> {
    if (!this.storageManager) return;

    // If removing current account, clear local references
    if (accountAddress === this.accountAddress) {
      this.accountAddress = null;
      this.database = null;
    }

    await this.storageManager.removeDatabase(accountAddress);
  }

  /**
   * Cleanup and release resources
   */
  async cleanup(): Promise<void> {
    // Save current state before cleanup
    if (this.database && this.accountAddress) {
      await this.persistState();
    }

    if (this.plugin) {
      await this.plugin.close();
      this.plugin = null;
    }

    if (this.storageManager) {
      await this.storageManager.closeAll();
    }

    this.wasmModule = null;
    this.database = null;
    this.accountAddress = null;
  }
}

/**
 * Create a new vault client instance
 */
export function createVaultClient(config?: VaultConfigWithStorage): VaultClient {
  return new VaultClient(config);
}

/**
 * Default vault client instance
 */
let defaultClient: VaultClient | null = null;

/**
 * Get or create the default vault client
 */
export async function getDefaultVaultClient(config?: VaultConfigWithStorage): Promise<VaultClient> {
  if (!defaultClient) {
    defaultClient = createVaultClient(config);
    await defaultClient.initialize();
  }
  return defaultClient;
}

/**
 * Export error class for convenience
 */
export { VaultError, VaultErrorCode } from './types';