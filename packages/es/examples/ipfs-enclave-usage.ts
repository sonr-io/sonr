/**
 * Usage examples for IPFS/Helia integration with MPC enclave data
 *
 * This file demonstrates how to use the IPFS integration for
 * storing and retrieving MPC enclave data in the Sonr ecosystem.
 */

import {
  // IPFS client
  IPFSClient,
  createIPFSClient,
  type IPFSClientConfig,
  // Enclave manager
  EnclaveIPFSManager,
  createEnclaveIPFSManager,
  type EnclaveDataWithCID,
  // Enhanced vault client
  VaultClientWithIPFS,
  createVaultClientWithIPFS,
  // Caching
  IPFSCache,
  createIPFSCache,
  // DWN query service
  DWNIPFSQueryService,
  createDWNIPFSQueryService,
} from '@sonr.io/es';

// ============================================
// Example 1: Basic IPFS Client Usage
// ============================================
async function basicIPFSExample() {
  console.log('🌐 Basic IPFS Client Example');

  // Create and initialize IPFS client
  const ipfsClient = await createIPFSClient({
    gateways: ['https://gateway.pinata.cloud', 'https://ipfs.io'],
    enablePersistence: true,
  });

  try {
    // Store data
    const data = new TextEncoder().encode('Hello from Sonr!');
    const { cid, size, timestamp } = await ipfsClient.addEnclaveData(data);
    console.log(`✅ Stored data with CID: ${cid} (${size} bytes)`);

    // Retrieve data
    const retrieved = await ipfsClient.getEnclaveData(cid);
    const text = new TextDecoder().decode(retrieved);
    console.log(`✅ Retrieved: "${text}"`);

    // Get node status
    const status = await ipfsClient.getNodeStatus();
    console.log(`📊 Node Status:`, status);

    // Pin important data
    await ipfsClient.pin(cid);
    console.log(`📌 Pinned CID: ${cid}`);

    // List all pins
    const pins = await ipfsClient.listPins();
    console.log(`📋 Total pinned items: ${pins.length}`);
  } finally {
    await ipfsClient.cleanup();
  }
}

// ============================================
// Example 2: MPC Enclave Storage
// ============================================
async function mpcEnclaveExample() {
  console.log('🔐 MPC Enclave Storage Example');

  const ipfsClient = await createIPFSClient();
  const enclaveManager = await createEnclaveIPFSManager(ipfsClient, {
    encryptionRequired: true,
    pinningEnabled: true,
    maxRetries: 3,
  });

  try {
    // Create MPC enclave data
    const enclaveData: EnclaveDataWithCID = {
      publicKey: 'ed25519:8FYH...publickey...ZKpQ',
      privateKeyShares: [
        'share1_encrypted_base64...',
        'share2_encrypted_base64...',
        'share3_encrypted_base64...',
      ],
      threshold: 2, // Need 2 out of 3 shares to reconstruct
      parties: 3,
      encryptionMetadata: {
        algorithm: 'AES-256-GCM',
        keyVersion: 1,
        consensusHeight: 12345,
        nonce: crypto.randomUUID(),
      },
    };

    // Encrypt the payload (in production, use consensus keys)
    const payload = JSON.stringify({
      ...enclaveData,
      timestamp: Date.now(),
      chainId: 'sonr-mainnet-1',
    });
    const encryptedPayload = new TextEncoder().encode(payload);

    // Store enclave data
    const result = await enclaveManager.storeEnclaveData(
      enclaveData,
      encryptedPayload
    );
    console.log(`✅ Stored enclave with CID: ${result.cid}`);
    console.log(`   - Size: ${result.size} bytes`);
    console.log(`   - Pinned: ${result.isPinned}`);

    // Retrieve and verify
    const retrieved = await enclaveManager.retrieveEnclaveData(result.cid);
    const isValid = await enclaveManager.verifyEnclaveDataIntegrity(
      result.cid,
      encryptedPayload
    );
    console.log(`✅ Retrieved enclave data (integrity: ${isValid})`);

    // Check status
    const status = await enclaveManager.getEnclaveStatus(result.cid);
    console.log(`📊 Enclave Status:`, status);
  } finally {
    await ipfsClient.cleanup();
  }
}

// ============================================
// Example 3: Vault Client with IPFS
// ============================================
async function vaultWithIPFSExample() {
  console.log('🔒 Vault Client with IPFS Example');

  const vaultClient = createVaultClientWithIPFS({
    chainId: 'sonr-testnet-1',
    enableIPFSPersistence: true,
    ipfsGateways: ['https://gateway.pinata.cloud'],
    enclave: {
      publicKey: 'test-public-key',
      privateKeyShares: ['share1', 'share2', 'share3'],
      threshold: 2,
      parties: 3,
    },
  });

  try {
    // Initialize vault with IPFS
    await vaultClient.initializeWithIPFS(
      '/path/to/vault.wasm',
      'sonr1abc...xyz'
    );

    // Store vault enclave
    const cid = await vaultClient.storeVaultEnclave([
      'encrypted_share1',
      'encrypted_share2',
      'encrypted_share3',
    ]);
    console.log(`✅ Stored vault enclave: ${cid}`);

    // Retrieve vault enclave
    const enclave = await vaultClient.retrieveVaultEnclave(cid);
    console.log(`✅ Retrieved enclave for ${enclave.parties} parties`);

    // List pinned enclaves
    const pinnedEnclaves = await vaultClient.listPinnedEnclaves();
    console.log(`📌 Pinned enclaves: ${pinnedEnclaves.length}`);

    // Sync with IPFS network
    await vaultClient.syncWithIPFS();
    console.log('✅ Synced with IPFS network');

    // Get IPFS status
    const ipfsStatus = await vaultClient.getIPFSStatus();
    console.log(`📊 IPFS Status:`, ipfsStatus);
  } catch (error) {
    // Handle WASM errors gracefully
    if (error.message.includes('WASM')) {
      console.log('⚠️ WASM not available, using IPFS features only');
    } else {
      throw error;
    }
  } finally {
    await vaultClient.cleanup();
  }
}

// ============================================
// Example 4: Caching Layer
// ============================================
async function cachingExample() {
  console.log('⚡ IPFS Caching Example');

  const ipfsClient = await createIPFSClient();
  const cache = createIPFSCache({
    maxSize: 50,
    ttl: 60000, // 1 minute TTL
    enablePersistence: true,
  });

  try {
    // Store some data in IPFS
    const data1 = new TextEncoder().encode('Data item 1');
    const data2 = new TextEncoder().encode('Data item 2');
    const data3 = new TextEncoder().encode('Data item 3');

    const { cid: cid1 } = await ipfsClient.addEnclaveData(data1);
    const { cid: cid2 } = await ipfsClient.addEnclaveData(data2);
    const { cid: cid3 } = await ipfsClient.addEnclaveData(data3);

    // Preload into cache
    console.log('📥 Preloading data into cache...');
    await cache.preload(
      [cid1, cid2, cid3],
      async (cid) => await ipfsClient.getEnclaveData(cid)
    );

    // Measure cache performance
    console.time('Cache retrieval');
    const cached1 = await cache.get(cid1);
    const cached2 = await cache.get(cid2);
    const cached3 = await cache.get(cid3);
    console.timeEnd('Cache retrieval');

    // Get cache statistics
    const stats = cache.getStats();
    console.log('📊 Cache Statistics:');
    console.log(`   - Size: ${stats.size} items`);
    console.log(`   - Total bytes: ${stats.totalBytes}`);
    console.log(`   - Hit rate: ${stats.hitRate.toFixed(2)}%`);
    console.log(`   - Avg access time: ${stats.avgAccessTime.toFixed(2)}ms`);

    // Clean up expired entries
    const removed = await cache.cleanup();
    console.log(`🧹 Cleaned up ${removed} expired entries`);
  } finally {
    await cache.destroy();
    await ipfsClient.cleanup();
  }
}

// ============================================
// Example 5: DWN IPFS Query Service
// ============================================
async function dwnQueryServiceExample() {
  console.log('🔍 DWN IPFS Query Service Example');

  const queryService = createDWNIPFSQueryService({
    rpcEndpoint: 'http://localhost:1317',
    defaultStaleTime: 30000,
  });

  try {
    // Query IPFS status from backend
    const ipfsStatus = await queryService.queryIPFSStatus();
    console.log('📊 Backend IPFS Status:');
    console.log(`   - Enabled: ${ipfsStatus.enabled}`);
    console.log(`   - Peer ID: ${ipfsStatus.peerId}`);
    console.log(`   - Connected peers: ${ipfsStatus.connectedPeers}`);

    // Query specific CID content
    const cidResponse = await queryService.queryCIDContent(
      'QmYwAPJzv5CZsnA625s3Xf2nemtYgPpHdWEz79ojWnPbdG',
      false // Don't decrypt
    );
    if (cidResponse.error) {
      console.log(`⚠️ CID not found: ${cidResponse.error}`);
    } else {
      console.log(`✅ Found CID content (${cidResponse.size} bytes)`);
    }

    // Query enclave data for a vault
    const enclaveResponse = await queryService.queryEnclaveData(
      'did:sonr:vault123',
      false // Don't include private shares
    );
    if (enclaveResponse.enclaveCid) {
      console.log(`✅ Found enclave for vault: ${enclaveResponse.enclaveCid}`);
    }

    // Store new enclave data via backend
    const newEnclaveData = new TextEncoder().encode('New enclave data');
    const storeResult = await queryService.storeEnclaveData(
      'did:sonr:newvault',
      newEnclaveData
    );
    console.log(`✅ Stored via backend: ${storeResult.cid}`);
  } finally {
    await queryService.cleanup();
  }
}

// ============================================
// Example 6: Error Handling and Recovery
// ============================================
async function errorHandlingExample() {
  console.log('🛡️ Error Handling Example');

  const ipfsClient = await createIPFSClient({
    gateways: [
      'https://gateway.pinata.cloud',
      'https://ipfs.io',
      'http://localhost:8080', // Local fallback
    ],
  });

  const enclaveManager = new EnclaveIPFSManager(ipfsClient, {
    encryptionRequired: true,
    pinningEnabled: true,
    maxRetries: 3,
    operationTimeout: 10000,
  });

  try {
    // Handle invalid CID
    try {
      await ipfsClient.getEnclaveData('invalid-cid-format');
    } catch (error) {
      console.log('✅ Caught invalid CID error:', error.message);
    }

    // Handle missing encryption metadata
    try {
      const invalidEnclave: EnclaveDataWithCID = {
        publicKey: 'test',
        privateKeyShares: ['share1'],
        threshold: 1,
        parties: 1,
        // Missing encryptionMetadata when encryptionRequired is true
      };
      await enclaveManager.storeEnclaveData(
        invalidEnclave,
        new Uint8Array()
      );
    } catch (error) {
      console.log('✅ Caught missing encryption metadata:', error.message);
    }

    // Handle network timeout with retry
    console.log('🔄 Testing retry logic...');
    const enclaveData: EnclaveDataWithCID = {
      publicKey: 'retry-test',
      privateKeyShares: ['share1'],
      threshold: 1,
      parties: 1,
      encryptionMetadata: {
        algorithm: 'AES-256-GCM',
        keyVersion: 1,
        consensusHeight: 1,
        nonce: 'test',
      },
    };

    const result = await enclaveManager.storeEnclaveData(
      enclaveData,
      new Uint8Array([1, 2, 3])
    );
    console.log('✅ Succeeded with retry logic');
  } finally {
    await ipfsClient.cleanup();
  }
}

// ============================================
// Example 7: Performance Best Practices
// ============================================
async function performanceExample() {
  console.log('🚀 Performance Best Practices Example');

  // 1. Use connection pooling
  const ipfsClient = await createIPFSClient({
    gateways: ['https://gateway.pinata.cloud'],
    libp2pConfig: {
      connectionManager: {
        maxConnections: 100,
        minConnections: 10,
      },
    },
  });

  // 2. Use caching aggressively
  const cache = createIPFSCache({
    maxSize: 200,
    ttl: 300000, // 5 minutes
    enablePersistence: true,
  });

  // 3. Batch operations
  const enclaveManager = new EnclaveIPFSManager(ipfsClient, {
    encryptionRequired: false,
    pinningEnabled: true,
    redundancy: 3,
    maxRetries: 2,
  });

  try {
    // Batch store multiple items
    console.log('📦 Batch storing enclaves...');
    const enclaves = Array.from({ length: 10 }, (_, i) => ({
      data: {
        publicKey: `key-${i}`,
        privateKeyShares: [`share-${i}`],
        threshold: 1,
        parties: 1,
      } as EnclaveDataWithCID,
      payload: new Uint8Array([i, i, i]),
    }));

    console.time('Batch store');
    const results = await enclaveManager.batchStoreEnclaves(enclaves);
    console.timeEnd('Batch store');
    console.log(`✅ Stored ${results.length} enclaves in batch`);

    // Preload frequently accessed data
    const cids = results.map((r) => r.cid);
    console.time('Preload cache');
    await cache.preload(cids.slice(0, 5), async (cid) =>
      enclaveManager.retrieveEnclaveData(cid)
    );
    console.timeEnd('Preload cache');

    // Use cache for fast retrieval
    console.time('Cached retrieval');
    for (const cid of cids.slice(0, 5)) {
      await cache.get(cid);
    }
    console.timeEnd('Cached retrieval');

    const stats = cache.getStats();
    console.log(`📊 Cache hit rate: ${stats.hitRate.toFixed(2)}%`);
  } finally {
    await cache.destroy();
    await ipfsClient.cleanup();
  }
}

// ============================================
// Main: Run all examples
// ============================================
async function main() {
  console.log('🌟 Sonr IPFS/Helia Integration Examples\n');

  try {
    await basicIPFSExample();
    console.log('\n---\n');

    await mpcEnclaveExample();
    console.log('\n---\n');

    await vaultWithIPFSExample();
    console.log('\n---\n');

    await cachingExample();
    console.log('\n---\n');

    await dwnQueryServiceExample();
    console.log('\n---\n');

    await errorHandlingExample();
    console.log('\n---\n');

    await performanceExample();

    console.log('\n✅ All examples completed successfully!');
  } catch (error) {
    console.error('❌ Example failed:', error);
    process.exit(1);
  }
}

// Run if executed directly
if (require.main === module) {
  main();
}

// Export examples for use in other modules
export {
  basicIPFSExample,
  mpcEnclaveExample,
  vaultWithIPFSExample,
  cachingExample,
  dwnQueryServiceExample,
  errorHandlingExample,
  performanceExample,
};