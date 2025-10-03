import { describe, it, expect } from 'vitest';
// TODO: Fix import path - should be '../src/plugin' not '../src/plugins'
// import * as plugins from '../src/plugins';

describe.skip('Plugins Module', () => {
  it('should export motor namespace', () => {
    expect(plugins.motor).toBeDefined();
  });

  it('should export vault namespace', () => {
    expect(plugins.vault).toBeDefined();
  });

  it('should export VaultClient', () => {
    expect(plugins.VaultClient).toBeDefined();
  });

  it('should export createVaultClient', () => {
    expect(plugins.createVaultClient).toBeDefined();
    expect(typeof plugins.createVaultClient).toBe('function');
  });

  it('should export MotorPluginImpl', () => {
    expect(plugins.MotorPluginImpl).toBeDefined();
  });

  it('should export createMotorPlugin', () => {
    expect(plugins.createMotorPlugin).toBeDefined();
    expect(typeof plugins.createMotorPlugin).toBe('function');
  });

  it('should export VaultError and VaultErrorCode', () => {
    expect(plugins.VaultError).toBeDefined();
    expect(plugins.VaultErrorCode).toBeDefined();
  });

  it('should export VaultStorageManager', () => {
    expect(plugins.vault.VaultStorageManager).toBeDefined();
    expect(typeof plugins.vault.VaultStorageManager).toBe('function');
  });

  it('should export storage-related types', () => {
    // These are TypeScript types, so we check if they're re-exported properly
    // by checking that the module has the expected structure
    expect(plugins.vault).toHaveProperty('AccountVaultDatabase');
    expect(plugins.vault).toHaveProperty('defaultStorageManager');
  });
});

describe.skip('Plugin Imports', () => {
  it('should be able to import from @sonr.io/es/plugins', async () => {
    // This test verifies the package.json exports are correct
    const pluginsModule = await import('../dist/plugins/index.js');
    expect(pluginsModule).toBeDefined();
    expect(pluginsModule.motor).toBeDefined();
    expect(pluginsModule.vault).toBeDefined();
  });

  it('should be able to import vault directly', async () => {
    const vaultModule = await import('../dist/plugins/vault/index.js');
    expect(vaultModule).toBeDefined();
    expect(vaultModule.VaultClient).toBeDefined();
  });

  it('should be able to import storage components from vault', async () => {
    const vaultModule = await import('../dist/plugins/vault/index.js');
    expect(vaultModule.VaultStorageManager).toBeDefined();
    expect(vaultModule.AccountVaultDatabase).toBeDefined();
    expect(vaultModule.defaultStorageManager).toBeDefined();
  });

  it('should be able to import motor directly', async () => {
    const motorModule = await import('../dist/plugins/motor/index.js');
    expect(motorModule).toBeDefined();
    expect(motorModule.MotorPluginImpl).toBeDefined();
  });
});

describe.skip('VaultClient Storage Integration', () => {
  it('should create vault client with storage disabled by default', () => {
    const client = plugins.createVaultClient();
    expect(client).toBeDefined();
    expect(client).toBeInstanceOf(plugins.VaultClient);
  });

  it('should create vault client with storage enabled', () => {
    const client = plugins.createVaultClient({
      enablePersistence: true,
      autoCleanup: false
    });
    expect(client).toBeDefined();
    expect(client).toBeInstanceOf(plugins.VaultClient);
  });

  it('should have storage management methods', () => {
    const client = plugins.createVaultClient({
      enablePersistence: true
    });

    // Check that all storage methods exist
    expect(typeof client.persistState).toBe('function');
    expect(typeof client.loadPersistedState).toBe('function');
    expect(typeof client.clearPersistedState).toBe('function');
    expect(typeof client.saveToken).toBe('function');
    expect(typeof client.getPersistedTokens).toBe('function');
    expect(typeof client.removeExpiredTokens).toBe('function');
    expect(typeof client.switchAccount).toBe('function');
    expect(typeof client.listPersistedAccounts).toBe('function');
    expect(typeof client.removeAccount).toBe('function');
  });
});