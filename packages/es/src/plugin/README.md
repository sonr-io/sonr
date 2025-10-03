# Vault Plugin with Dexie.js Persistence

The Vault plugin now supports persistent storage using Dexie.js (IndexedDB wrapper), allowing vault state and UCAN tokens to persist across browser sessions.

## Features

- 🔐 **Account-based database separation** - Each account has its own isolated database
- 💾 **Automatic persistence** - Tokens are automatically saved when created
- 🔄 **Cross-browser support** - Works with all modern browsers supporting IndexedDB
- ⚡ **Backward compatible** - Storage is opt-in, existing code continues to work
- 🧹 **Automatic cleanup** - Expired tokens and sessions are cleaned up periodically

## Basic Usage

### Without Persistence (Default Behavior)

```typescript
import { createVaultClient } from '@sonr.io/es/plugins/vault';

// Create vault client without persistence (backward compatible)
const vault = createVaultClient();

// Initialize the vault
await vault.initialize();

// Use vault as before
const token = await vault.newOriginToken({
  audience_did: 'did:example:123',
});
```

### With Persistence Enabled

```typescript
import { createVaultClient } from '@sonr.io/es/plugins/vault';

// Create vault client with persistence enabled
const vault = createVaultClient({
  enablePersistence: true,
  autoCleanup: true,
  cleanupInterval: 3600000, // 1 hour
});

// Initialize with account address for database separation
const accountAddress = 'sonr1abc123...';
await vault.initialize('/plugin.wasm', accountAddress);

// Vault state and tokens are now automatically persisted
const token = await vault.newOriginToken({
  audience_did: 'did:example:123',
});
// Token is automatically saved to IndexedDB

// Retrieve all persisted tokens
const tokens = await vault.getPersistedTokens();
console.log(`Found ${tokens.length} saved tokens`);
```

## Advanced Usage

### Multi-Account Support

```typescript
// Start with account 1
await vault.initialize('/plugin.wasm', 'sonr1account1');

// Do some work...
const token1 = await vault.newOriginToken({ audience_did: 'did:1' });

// Switch to account 2
await vault.switchAccount('sonr1account2');

// Work with account 2's isolated database
const token2 = await vault.newOriginToken({ audience_did: 'did:2' });

// List all accounts with persisted data
const accounts = await vault.listPersistedAccounts();
console.log('Accounts with saved data:', accounts);

// Remove an account's data
await vault.removeAccount('sonr1account1');
```

### State Management

```typescript
// Manually save current state
await vault.persistState();

// Load persisted state
const state = await vault.loadPersistedState();
if (state) {
  console.log('Vault initialized:', state.isInitialized);
  console.log('Last accessed:', new Date(state.lastAccessed));
}

// Clear all persisted data for current account
await vault.clearPersistedState();
```

### Token Management

```typescript
// Manually save a token
await vault.saveToken({
  token: 'eyJ...',
  issuer: 'did:sonr:123',
  address: 'sonr1abc...',
});

// Get all saved tokens
const tokens = await vault.getPersistedTokens();

// Remove expired tokens
await vault.removeExpiredTokens();
```

### Storage Persistence

```typescript
import { VaultStorageManager } from '@sonr.io/es/plugins/vault';

const storageManager = new VaultStorageManager({
  enablePersistence: true,
});

// Request persistent storage (prompts user in some browsers)
const isPersisted = await storageManager.requestPersistentStorage();
console.log('Storage persisted:', isPersisted);

// Check persistence status
const status = await storageManager.tryPersistWithoutPromptingUser();
// Returns: 'persisted' | 'prompt' | 'never'

// Get storage estimate
const estimate = await storageManager.getStorageEstimate();
if (estimate) {
  console.log(`Using ${estimate.usage} of ${estimate.quota} bytes`);
}
```

## Configuration Options

```typescript
interface VaultStorageConfig {
  enablePersistence?: boolean;  // Enable IndexedDB storage (default: false)
  storageQuotaRequest?: number; // Storage quota to request in bytes
  autoCleanup?: boolean;        // Enable automatic cleanup (default: true)
  cleanupInterval?: number;     // Cleanup interval in ms (default: 3600000)
}
```

## Browser Compatibility

- ✅ Chrome/Edge 23+
- ✅ Firefox 16+
- ✅ Safari 10+
- ✅ Opera 15+
- ✅ iOS Safari 10+
- ✅ Chrome for Android

## Storage Limits

- **Chrome/Edge**: 60% of total disk space
- **Firefox**: 50% of free disk space
- **Safari**: 1GB initially, can request more
- **Mobile browsers**: Varies by device

## Migration Guide

### From Non-Persistent to Persistent

```typescript
// Before (non-persistent)
const vault = createVaultClient();
await vault.initialize();

// After (with persistence)
const vault = createVaultClient({
  enablePersistence: true,
});
await vault.initialize('/plugin.wasm', accountAddress);
```

No other code changes required - all existing methods work the same way.

## Security Considerations

- Databases are named by account address for isolation
- No private keys or sensitive cryptographic material is stored
- Only UCAN tokens and metadata are persisted
- Use HTTPS in production for better storage persistence
- Consider encrypting sensitive data before storage

## Troubleshooting

### Storage Not Persisting

1. Check if running on HTTPS (required for persistence in some browsers)
2. Verify IndexedDB is not disabled in browser settings
3. Check available storage quota
4. Try requesting persistent storage explicitly

### Database Errors

```typescript
try {
  await vault.initialize('/plugin.wasm', accountAddress);
} catch (error) {
  if (error.code === 'VAULT_NOT_INITIALIZED') {
    // Handle initialization error
  }
}
```

### Cleanup Issues

If automatic cleanup is not working:

```typescript
// Manually trigger cleanup
const storageManager = new VaultStorageManager();
await storageManager.cleanupExpiredData();
```