# `@sonr.io/es`

A tree-shakeable, framework agnostic, [pure ESM](https://gist.github.com/sindresorhus/a39789f98801d908bbc7ff3ecc99d99c) alternative of [CosmJS](https://github.com/cosmos/cosmjs) and [Cosmos Kit](https://cosmoskit.com) (**generate bundles up to 10x smaller than Cosmos Kit**).

- [Features](#features)
- [Installing](#installing)
  - [Using with TypeScript](#using-with-typescript)
  - [Using with Vite](#using-with-vite)
  - [Using Station wallet](#using-station-wallet)
- [Examples](#examples)
- [Modules](#modules)
  - [`@sonr.io/es/client`](#@sonr.io/esclient)
  - [`@sonr.io/es/codec`](#@sonr.io/escodec)
  - [`@sonr.io/es/protobufs`](#@sonr.io/esprotobufs)
  - [`@sonr.io/es/registry`](#@sonr.io/esregistry)
  - [`@sonr.io/es/wallet`](#@sonr.io/eswallet)
  - [`@sonr.io/es/ipfs`](#ipfs-integration)
- [Benchmarks](#benchmarks)
  - [Results](#results)
- [See More](#see-more)

## Features

- **Fully tree-shakeable**: import and bundle only the modules you need
- **Framework agnostic**: integrate with any web framework (React, Vue, Svelte, Solid, etc.)
- **Lightweight and minimal**: 153 KB gzipped to connect a React app to Keplr via browser extension or WalletConnect, 10x smaller than Cosmos Kit V2 (see [benchmarks](#benchmarks))
- **Uses modern web APIs**: no dependencies on Node.js and minimal dependencies on third-party libraries where possible
- **Supports modern bundlers**: works with Vite, SWC, Rollup, etc.
- **Fully typed**: written in TypeScript and ships with type definitions

## Installing

For Cosmos SDK v0.47 and below:

```sh
npm install @sonr.io/es

pnpm add @sonr.io/es

yarn add @sonr.io/es
```

For Cosmos SDK v0.50, install using the `sdk50` tag:

```sh
npm install @sonr.io/es@sdk50

pnpm add @sonr.io/es@sdk50

yarn add @sonr.io/es@sdk50
```

> [!IMPORTANT]  
> The bump from v0.47 to v0.50 introduces significant breaking changes and is not recommended to be used unless necessary. To reduce the impact on consumers, the `main` branch and the published package on npm with the `latest` tag will continue to target v0.47 until the majority of live chains have migrated to v0.50.
>
> The [`parallel/sdk50`](https://github.com/coinhall/@sonr.io/es/tree/parallel/sdk50) branch targetting v0.50 will be developed and maintained in parallel with the `main` branch, where the same patch version number should have feature parity (eg. `@sonr.io/es@0.0.69` should have the same features as `@sonr.io/es@0.0.69-sdk50.0`).

### Using with TypeScript

This library only exports ES modules. To ensure imports from this library work correctly, the following configuration is required in `tsconfig.json`:

```ts
{
  "compilerOptions": {
    "moduleResolution": "bundler", // recommended if using modern bundlers
    // or "node16"
    // or "nodenext"
    // but NOT "node"
  }
}
```

### Using with Vite

If you are using Vite, the following configuration is required in `vite.config.ts`:

```ts
export default defineConfig({
  define: {
    global: "window",
  },
});
```

> This can be removed once support for WalletConnect v1 is no longer required.

### Using Station wallet

The Station wallet currently relies on WalletConnect v1. If you want to import and use `StationController`, a polyfill for `Buffer` is required:

```ts
// First, install the buffer package
npm install buffer

// Then, create a new file 'polyfill.ts'
import { Buffer } from "buffer";
(window as any).Buffer = Buffer;

// Finally, import the above file in your entry file
import "./polyfill";
```

See [`examples/solid-vite`](./examples/solid-vite) for a working example.

> This can be removed once support for WalletConnect v1 is no longer required.

## Examples

### Using the ESM Autoloader (Browser/CDN)

The library includes an autoloader that automatically initializes all modules and makes them available globally via `window.Sonr`. This is perfect for quick prototyping or when you want to use the library without a build system.

#### Method 1: Load from CDN

```html
<!DOCTYPE html>
<html>
<head>
    <title>Sonr ES Example</title>
</head>
<body>
    <!-- Load the autoloader from CDN -->
    <script type="module" src="https://unpkg.com/@sonr.io/es@latest/dist/autoloader.js"></script>
    
    <script type="module">
        // Wait for the library to be ready
        window.addEventListener('sonr:ready', async (event) => {
            const Sonr = event.detail;
            console.log('Sonr is ready!', Sonr);
            
            // Check WebAuthn availability
            if (await Sonr.webauthn.isAvailable()) {
                console.log('WebAuthn is available');
                
                // Register with passkey
                const registration = await Sonr.webauthn.register({
                    username: 'alice',
                    displayName: 'Alice',
                    rpId: window.location.hostname,
                    rpName: 'My App'
                });
                console.log('Registration successful:', registration);
            }
            
            // Access other modules
            console.log('Available modules:', {
                auth: Sonr.auth,
                client: Sonr.client,
                codec: Sonr.codec,
                wallet: Sonr.wallet,
                plugins: Sonr.plugins
            });
        });
    </script>
</body>
</html>
```

#### Method 2: Import as ES Module

```html
<script type="module">
    // Import the autoloader
    import Sonr from 'https://unpkg.com/@sonr.io/es@latest/dist/autoloader.js';
    
    // Initialize with custom configuration
    await Sonr.init({
        enableMotor: true,  // Enable Motor WASM plugin
        enableVault: true,  // Enable Vault client
        motor: {
            wasmUrl: '/motor.wasm'  // Custom WASM URL
        },
        vault: {
            endpoint: 'https://vault.example.com'
        }
    });
    
    // Use the library
    console.log('Environment:', Sonr.getEnvironment());
    
    // WebAuthn operations
    if (await Sonr.webauthn.isAvailable()) {
        // Login with passkey
        const login = await Sonr.webauthn.login({
            rpId: window.location.hostname
        });
        console.log('Login successful:', login);
    }
</script>
```

#### Method 3: Using in Node.js/Build Systems

```javascript
// Import specific modules (tree-shakeable)
import { registerWithPasskey, loginWithPasskey } from '@sonr.io/es/client/auth';
import { createMotorPlugin } from '@sonr.io/es/plugins';
import { bech32 } from '@sonr.io/es/codec';

// Or import the entire autoloader
import Sonr from '@sonr.io/es/autoloader';

// Initialize and use
async function main() {
    // Initialize Sonr
    await Sonr.init({
        enableMotor: true,
        enableVault: false
    });
    
    // Use WebAuthn
    if (await Sonr.webauthn.isAvailable()) {
        const result = await Sonr.webauthn.register({
            username: 'bob',
            displayName: 'Bob Smith',
            rpId: 'example.com',
            rpName: 'Example App'
        });
        console.log('Registered:', result);
    }
    
    // Access plugins
    const motor = await Sonr.createMotorPlugin();
    console.log('Motor plugin ready:', motor);
    
    // Use codec utilities
    const address = Sonr.codec.bech32.encode('sonr', [1, 2, 3, 4]);
    console.log('Encoded address:', address);
}

main();
```

#### Autoloader API Reference

The autoloader exposes the following on `window.Sonr`:

```javascript
window.Sonr = {
    // Core modules
    auth: {...},        // Authentication utilities
    client: {...},      // Blockchain client
    codec: {...},       // Encoding/decoding utilities
    wallet: {...},      // Wallet management
    registry: {...},    // Chain registry
    plugins: {...},     // WASM plugins (motor, vault)
    
    // WebAuthn shortcuts
    webauthn: {
        register: registerWithPasskey,
        login: loginWithPasskey,
        isSupported: isWebAuthnSupported,
        isAvailable: isWebAuthnAvailable,
        isConditionalAvailable: isConditionalMediationAvailable,
        bufferToBase64url: bufferToBase64url,
        base64urlToBuffer: base64urlToBuffer
    },
    
    // Plugin shortcuts
    motor: {...},       // Motor plugin namespace
    vault: {...},       // Vault plugin namespace
    
    // Factory functions
    createMotorPlugin: Function,
    createVaultClient: Function,
    
    // Utilities
    init: async (config) => {...},     // Initialize with config
    getEnvironment: () => {...},       // Get environment info
    isBrowser: Boolean,                // Check if running in browser
    isNode: Boolean,                   // Check if running in Node.js
    version: String                    // Library version
}
```

#### Events

The autoloader dispatches the following events:

- `sonr:ready` - Fired when the library is fully loaded and initialized
  ```javascript
  window.addEventListener('sonr:ready', (event) => {
      const Sonr = event.detail;
      console.log('Sonr is ready!', Sonr);
  });
  ```

### Other Examples

See the [`examples`](./examples) folder for more detailed examples:

1. [How do I connect to third party wallets via browser extension or WalletConnect? How do I create, sign, and broadcast transactions?](./examples/solid-vite)
2. [How do I programmatically sign and broadcast transactions without relying on a third party wallet?](./examples/mnemonic-wallet)
3. [How do I verify signatures signed using the `signArbitrary` function?](./examples/verify-signatures)
4. [How do I batch queries to the blockchain?](./examples/batch-query)
5. [How do I use the ESM autoloader in a browser?](./examples/autoloader.html)

## Modules

This package is split into multiple subdirectories, with each subdirectory having their own set of functionalities. The root directory does not contain any exports, and all exports are exported from the subdirectories. Thus, imports must be done by referencing the subdirectories (ie. `import { ... } from  "@sonr.io/es/client"`).

### `@sonr.io/es/client`

This directory contains models and helper functions to interact with Cosmos SDK via the [CometBFT RPC](https://docs.cosmos.network/v0.50/core/grpc_rest#cometbft-rpc).

### `@sonr.io/es/codec`

This directory contains various encoding and decoding functions that relies solely on [Web APIs](https://developer.mozilla.org/en-US/docs/Web/API) and has no dependencies on Node.js. For modern browsers and Node v16+, this should work out of the box.

### `@sonr.io/es/protobufs`

This directory contains the auto-generated code for various Cosmos SDK based protobufs. See `scripts/gen-protobufs.mjs` for the script that generates the code.

### `@sonr.io/es/registry`

This directory contains various APIs, data, and types needed for wallet interactions (ie. Keplr). Some types are auto-generated, see `scripts/gen-registry.mjs` for the script that generates the types.

### `@sonr.io/es/wallet`

This directory is a [Cosmos Kit](https://cosmoskit.com) alternative to interact with wallets across all Cosmos SDK based blockchains. See [`examples/solid-vite`](./examples/solid-vite) for a working example.

**Wallets supported**:

- [Station](https://docs.terra.money/learn/station/)
- [Keplr](https://www.keplr.app/)
- [Leap](https://www.leapwallet.io/)
- [Cosmostation](https://wallet.cosmostation.io/)
- [OWallet](https://owallet.dev/)
- [Compass](https://compasswallet.io/) (for Sei only)
- [MetaMask](https://metamask.io/) (for Injective only)
- [Ninji](https://ninji.xyz/) (for Injective only)

**Features**:

- Supports both browser extension (desktop) and WalletConnect (mobile)
- Unified interface for connecting, signing, broadcasting, and event handling
- Signing of arbitrary messages (for wallets that support it)
- Simultaneous connections to multiple WalletConnect wallets

## Benchmarks

See the [`benchmarks`](./benchmarks) folder, where the bundle size of SonrES is compared against Cosmos Kit. The following are adhered to:

- Apps should only contain the minimal functionality of connecting to Osmosis via Keplr using both the browser extension and WalletConnect wallets
- Apps should be built using React 18 (as Cosmos Kit has a [hard dependency](https://docs.cosmoskit.com/get-started)) and Vite
- Use the total sum of all generated bundles as reported by Vite after running the `vite build` command, including the size of all other dependencies like React/HTML/CSS/etc. (note: this is crude and not 100% accurate, but is the simplest method)

### Results

> Last updated: 4th May 2024

| Package       | Minified | Gzipped |
| ------------- | -------- | ------- |
| SonrES        | 553 KB   | 153 KB  |
| Cosmos Kit v1 | 6010 KB  | 1399 KB |
| Cosmos Kit v2 | 6780 KB  | 1556 KB |

## See More

- [Changelog](./CHANGELOG.md) - for notable changes

## IPFS Integration

The `@sonr.io/es` package now includes comprehensive IPFS/Helia support for distributed MPC enclave data storage. This integration enables secure, decentralized storage of vault encryption keys and sensitive cryptographic material.

### Features

- 🌐 **Modern IPFS with Helia**: Built on the latest Helia implementation for JavaScript/TypeScript
- 🔐 **MPC Enclave Support**: Secure storage and retrieval of Multi-Party Computation enclave data
- ⚡ **Performance Optimized**: LRU caching, connection pooling, and retry logic with exponential backoff
- 🌍 **Browser & Node.js**: Full support for both environments with automatic transport selection
- 🔄 **Gateway Fallbacks**: Automatic fallback to IPFS gateways when direct connections fail
- 📦 **Batch Operations**: Efficient batch storage and retrieval of multiple enclaves
- 🔍 **DWN Integration**: Query service for backend IPFS operations through Decentralized Web Nodes

### Quick Start

```typescript
import { ipfs } from '@sonr.io/es';

// Create IPFS client
const client = await ipfs.createIPFSClient({
  gateways: ['https://gateway.pinata.cloud'],
  enablePersistence: true,
});

// Store enclave data
const enclaveData = {
  publicKey: 'ed25519:...',
  privateKeyShares: ['share1', 'share2', 'share3'],
  threshold: 2,
  parties: 3,
};

const { cid } = await client.addEnclaveData(
  new TextEncoder().encode(JSON.stringify(enclaveData))
);

// Retrieve data
const retrieved = await client.getEnclaveData(cid);
const data = JSON.parse(new TextDecoder().decode(retrieved));

// Clean up
await client.cleanup();
```

### API Reference

#### IPFSClient

The main IPFS client for interacting with the network.

```typescript
interface IPFSClient {
  initialize(): Promise<void>
  addEnclaveData(data: Uint8Array): Promise<EnclaveDataCID>
  getEnclaveData(cid: string): Promise<Uint8Array>
  pin(cid: string): Promise<void>
  unpin(cid: string): Promise<void>
  isPinned(cid: string): Promise<boolean>
  listPins(): Promise<string[]>
  getNodeStatus(): Promise<IPFSNodeStatus>
  cleanup(): Promise<void>
}
```

#### EnclaveIPFSManager

Manages MPC enclave data with encryption and integrity verification.

```typescript
interface EnclaveIPFSManager {
  storeEnclaveData(
    data: EnclaveDataWithCID,
    payload: Uint8Array
  ): Promise<EnclaveStorageResult>
  
  retrieveEnclaveData(cid: string): Promise<Uint8Array>
  
  verifyEnclaveDataIntegrity(
    cid: string,
    expectedData: Uint8Array
  ): Promise<boolean>
  
  batchStoreEnclaves(
    enclaves: Array<{data: EnclaveDataWithCID, payload: Uint8Array}>
  ): Promise<EnclaveStorageResult[]>
}
```

#### VaultClientWithIPFS

Enhanced vault client with integrated IPFS support.

```typescript
interface VaultClientWithIPFS extends VaultClient {
  initializeWithIPFS(
    wasmPath?: string,
    accountAddress?: string,
    ipfsConfig?: any
  ): Promise<void>
  
  storeEnclaveToIPFS(
    data: EnclaveDataWithCID,
    payload: Uint8Array
  ): Promise<string>
  
  retrieveEnclaveFromIPFS(cid: string): Promise<Uint8Array>
  
  listPinnedEnclaves(): Promise<string[]>
  
  syncWithIPFS(): Promise<void>
}
```

#### IPFSCache

High-performance caching layer with LRU eviction and TTL support.

```typescript
interface IPFSCache {
  get(cid: string): Promise<Uint8Array | null>
  set(cid: string, data: Uint8Array, metadata?: any): Promise<void>
  has(cid: string): Promise<boolean>
  remove(cid: string): Promise<boolean>
  clear(): Promise<void>
  preload(
    cids: string[],
    fetchFn: (cid: string) => Promise<Uint8Array>
  ): Promise<void>
  getStats(): CacheStats
}
```

### Configuration

#### IPFSClientConfig

```typescript
interface IPFSClientConfig {
  gatewayUrl?: string           // Primary IPFS gateway URL for content retrieval
  apiUrl?: string              // IPFS API URL for node operations (e.g., pinning)
  gateways?: string[]          // List of fallback IPFS gateway URLs
  enablePersistence?: boolean   // Enable persistent storage
  libp2pConfig?: any           // Custom libp2p configuration
  environment?: 'local' | 'testnet' | 'mainnet' // Auto-selects appropriate endpoints
  timeout?: number             // Request timeout in milliseconds (default: 30000)
  maxRetries?: number          // Max retries for failed requests (default: 3)
}
```

#### Default Configuration

```typescript
const DEFAULT_IPFS_CONFIG = {
  gatewayUrl: 'https://gateway.pinata.cloud',
  apiUrl: 'http://localhost:5001',  // Changes based on environment
  gateways: [
    'https://gateway.pinata.cloud',
    'https://ipfs.io',
    'https://cloudflare-ipfs.com',
    'https://dweb.link'
  ],
  apiEndpoints: {
    local: 'http://localhost:5001',
    testnet: 'https://ipfs.testnet.sonr.io',
    mainnet: 'https://ipfs.sonr.io'
  }
}
```

#### Environment Variables

The IPFS client automatically detects and uses these environment variables:

- `IPFS_GATEWAY_URL` - Primary gateway URL
- `IPFS_API_URL` - API endpoint URL
- `SONR_ENV` - Environment ('local', 'testnet', 'mainnet')

```bash
# Example .env file
IPFS_GATEWAY_URL=https://my-custom-gateway.com
IPFS_API_URL=https://my-ipfs-api.com:5001
SONR_ENV=testnet
```

#### EnclaveStorageConfig

```typescript
interface EnclaveStorageConfig {
  encryptionRequired: boolean   // Require encryption for all data
  pinningEnabled: boolean       // Auto-pin stored data
  redundancy: number           // Number of redundant copies
  maxRetries: number          // Max retry attempts
  operationTimeout?: number   // Operation timeout in ms
}
```

### Usage Examples

#### Basic Configuration

```typescript
import { createIPFSClient, DEFAULT_IPFS_CONFIG } from '@sonr.io/es/ipfs'

// Use defaults
const client = await createIPFSClient()

// Custom configuration
const customClient = await createIPFSClient({
  gatewayUrl: 'https://my.gateway.com',
  apiUrl: 'https://my.api.com:5001',
  environment: 'testnet',
  timeout: 60000,
  maxRetries: 5
})

// Update configuration dynamically
client.updateConfig({
  gatewayUrl: 'https://new.gateway.com',
  timeout: 30000
})

// Get current configuration
const config = client.getConfig()
console.log('Using gateway:', config.gatewayUrl)
```

#### Using the API Methods

```typescript
// Add content via HTTP API (requires apiUrl configuration)
const data = new Uint8Array([1, 2, 3, 4])
const cid = await client.addViaAPI(data)
console.log('Added via API:', cid)

// Pin content to prevent garbage collection
await client.pinViaAPI(cid)

// Get IPFS node information
const nodeInfo = await client.getNodeInfoViaAPI()
console.log('Node ID:', nodeInfo.ID)
console.log('Agent:', nodeInfo.AgentVersion)
```

See the [examples directory](./examples/ipfs-enclave-usage.ts) for comprehensive usage examples including:

- Basic IPFS operations
- MPC enclave storage
- Vault integration
- Caching strategies
- Error handling
- Performance optimization

### Performance Best Practices

1. **Use Caching Aggressively**: Enable the cache layer for frequently accessed data
2. **Batch Operations**: Use batch methods when storing/retrieving multiple items
3. **Preload Critical Data**: Use cache preloading for known CIDs
4. **Configure Gateways**: Provide multiple gateway URLs for redundancy
5. **Set Appropriate Timeouts**: Configure timeouts based on your network conditions
6. **Monitor Cache Stats**: Use cache statistics to optimize hit rates

### Troubleshooting

#### Common Issues

**Connection Failed**
```typescript
// Provide fallback gateways
const client = await createIPFSClient({
  gateways: [
    'http://localhost:5001',      // Local node
    'https://gateway.pinata.cloud', // Public gateway
    'https://ipfs.io'              // Fallback
  ]
});
```

**Slow Retrieval**
```typescript
// Enable caching for better performance
const cache = createIPFSCache({
  maxSize: 200,
  ttl: 300000, // 5 minutes
  enablePersistence: true
});

// Preload frequently used CIDs
await cache.preload(cids, fetchFunction);
```

**Network Timeouts**
```typescript
// Configure retry logic
const manager = new EnclaveIPFSManager(client, {
  maxRetries: 5,
  operationTimeout: 30000 // 30 seconds
});
```

### Security Considerations

1. **Always Encrypt Sensitive Data**: Use consensus-based encryption for enclave data
2. **Verify CID Integrity**: Always verify retrieved data matches expected CID
3. **Use HTTPS Gateways**: Prefer HTTPS gateways over HTTP
4. **Validate Enclave Structure**: Validate threshold and parties before storage
5. **Implement Access Control**: Use UCAN tokens for authorization

### Testing

Run unit tests:
```bash
pnpm test
```

Run integration tests (requires Docker):
```bash
docker-compose up -d ipfs
pnpm test:integration
```

### Dependencies

- `helia`: Core IPFS implementation
- `@helia/unixfs`: UnixFS for file operations
- `@helia/verified-fetch`: Verified content fetching
- `@libp2p/webrtc`: WebRTC transport
- `@libp2p/websockets`: WebSocket transport
- `multiformats`: CID and multiformat support
- `@tanstack/query-core`: Query caching for DWN service

