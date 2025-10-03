/**
 * Example usage of the Motor WASM service worker integration.
 * This demonstrates how to use the Motor plugin for both DWN and Wallet operations.
 */

import {
  createMotorPlugin,
  createMotorPluginForNode,
  createMotorPluginForBrowser,
  isMotorSupported,
  getMotorEnvironment,
  type MotorPlugin,
  type NewOriginTokenRequest,
  type CreateRecordRequest,
} from '@sonr.io/es/client/motor';

// ╭─────────────────────────────────────────────────────────╮
// │                Environment Detection                   │
// ╰─────────────────────────────────────────────────────────╯

async function detectEnvironment(): Promise<void> {
  console.log('🔍 Detecting environment capabilities...');
  
  const env = getMotorEnvironment();
  console.log('Environment info:', {
    browser: env.is_browser,
    node: env.is_node,
    serviceWorker: env.supports_service_worker,
    wasm: env.supports_wasm,
  });
  
  const supported = isMotorSupported();
  console.log('Motor supported:', supported);
}

// ╭─────────────────────────────────────────────────────────╮
// │                 Auto Plugin Creation                   │
// ╰─────────────────────────────────────────────────────────╯

async function createPlugin(): Promise<MotorPlugin> {
  console.log('🚀 Creating Motor plugin...');
  
  // Auto-detects environment and creates appropriate plugin
  const plugin = await createMotorPlugin({
    debug: true,
    timeout: 30000,
    max_retries: 3,
  });
  
  console.log('✅ Plugin created successfully');
  return plugin;
}

// ╭─────────────────────────────────────────────────────────╮
// │              Environment-Specific Creation              │
// ╰─────────────────────────────────────────────────────────╯

async function createBrowserPlugin(): Promise<MotorPlugin> {
  console.log('🌐 Creating browser-specific Motor plugin...');
  
  const plugin = await createMotorPluginForBrowser('/motor-worker', {
    auto_register_worker: true,
    prefer_service_worker: true,
    debug: true,
  });
  
  console.log('✅ Browser plugin created with service worker support');
  return plugin;
}

async function createNodePlugin(): Promise<MotorPlugin> {
  console.log('🖥️ Creating Node.js-specific Motor plugin...');
  
  const plugin = await createMotorPluginForNode('http://localhost:8080', {
    timeout: 15000,
    max_retries: 2,
    debug: true,
  });
  
  console.log('✅ Node.js plugin created with HTTP fallback');
  return plugin;
}

// ╭─────────────────────────────────────────────────────────╮
// │                   Wallet Operations                    │
// ╰─────────────────────────────────────────────────────────╯

async function demonstrateWalletOperations(plugin: MotorPlugin): Promise<void> {
  console.log('💼 Demonstrating wallet operations...');
  
  try {
    // Get issuer DID
    console.log('📋 Getting issuer DID...');
    const issuerResponse = await plugin.getIssuerDID();
    console.log('Issuer DID:', issuerResponse.issuer_did);
    console.log('Address:', issuerResponse.address);
    
    // Create origin token
    console.log('🎫 Creating UCAN origin token...');
    const tokenRequest: NewOriginTokenRequest = {
      audience_did: 'did:sonr:example-audience',
      attenuations: [
        {
          can: ['sign', 'encrypt'],
          with: 'vault://example-vault',
        },
      ],
      facts: ['motor-wasm-demo'],
      expires_at: Date.now() + (24 * 60 * 60 * 1000), // 24 hours
    };
    
    const tokenResponse = await plugin.newOriginToken(tokenRequest);
    console.log('✅ Origin token created:', tokenResponse.token.substring(0, 50) + '...');
    
    // Create attenuated token
    console.log('🔗 Creating attenuated token...');
    const attenuatedResponse = await plugin.newAttenuatedToken({
      parent_token: tokenResponse.token,
      audience_did: 'did:sonr:delegated-audience',
      attenuations: [
        {
          can: ['sign'],
          with: 'vault://limited-access',
        },
      ],
      expires_at: Date.now() + (2 * 60 * 60 * 1000), // 2 hours
    });
    console.log('✅ Attenuated token created:', attenuatedResponse.token.substring(0, 50) + '...');
    
    // Sign data
    console.log('✍️ Signing data...');
    const dataToSign = new TextEncoder().encode('Hello, Motor WASM!');
    const signResponse = await plugin.signData({ data: dataToSign });
    console.log('✅ Data signed, signature length:', signResponse.signature.length);
    
    // Verify signature
    console.log('🔍 Verifying signature...');
    const verifyResponse = await plugin.verifyData({
      data: dataToSign,
      signature: signResponse.signature,
    });
    console.log('✅ Signature valid:', verifyResponse.valid);
    
  } catch (error) {
    console.error('❌ Wallet operation failed:', error);
  }
}

// ╭─────────────────────────────────────────────────────────╮
// │                    DWN Operations                      │
// ╰─────────────────────────────────────────────────────────╯

async function demonstrateDWNOperations(plugin: MotorPlugin): Promise<void> {
  console.log('🌐 Demonstrating DWN operations...');
  
  try {
    // Create a record
    console.log('📝 Creating DWN record...');
    const recordData = new TextEncoder().encode(JSON.stringify({
      message: 'Hello from Motor DWN!',
      timestamp: new Date().toISOString(),
      version: '1.0.0',
    }));
    
    const createRequest: CreateRecordRequest = {
      target: 'did:sonr:alice',
      data: recordData,
      schema: 'https://schema.org/Message',
      protocol: 'https://protocol.example.com/messaging',
      published: true,
      encrypt: true, // Encrypt the data
    };
    
    const createResponse = await plugin.createRecord?.(createRequest);
    if (!createResponse) {
      console.log('ℹ️ DWN operations not available in this plugin instance');
      return;
    }
    
    console.log('✅ Record created:', createResponse.record_id);
    console.log('📅 Created at:', new Date(createResponse.created_at * 1000).toISOString());
    console.log('🔒 Encrypted:', createResponse.is_encrypted);
    
    // Read the record
    console.log('📖 Reading DWN record...');
    const readResponse = await plugin.readRecord?.(createResponse.record_id, createRequest.target);
    if (readResponse) {
      const decodedData = new TextDecoder().decode(readResponse.data);
      console.log('✅ Record data:', JSON.parse(decodedData));
      console.log('🔓 Decrypted successfully:', !readResponse.is_encrypted || readResponse.data.length > 0);
    }
    
    // Update the record
    console.log('✏️ Updating DWN record...');
    const updatedData = new TextEncoder().encode(JSON.stringify({
      message: 'Updated message from Motor DWN!',
      timestamp: new Date().toISOString(),
      version: '1.1.0',
      updated: true,
    }));
    
    const updateResponse = await plugin.updateRecord?.({
      record_id: createResponse.record_id,
      target: createRequest.target,
      data: updatedData,
      published: true,
    });
    
    if (updateResponse) {
      console.log('✅ Record updated at:', new Date(updateResponse.updated_at * 1000).toISOString());
    }
    
    // Delete the record
    console.log('🗑️ Deleting DWN record...');
    const deleteResponse = await plugin.deleteRecord?.(createResponse.record_id, createRequest.target);
    if (deleteResponse) {
      console.log('✅ Record deleted:', deleteResponse.status);
    }
    
  } catch (error) {
    console.error('❌ DWN operation failed:', error);
  }
}

// ╭─────────────────────────────────────────────────────────╮
// │                   Health Monitoring                    │
// ╰─────────────────────────────────────────────────────────╯

async function checkServiceHealth(plugin: MotorPlugin): Promise<void> {
  console.log('🏥 Checking service health...');
  
  try {
    // Test connection
    const connected = await plugin.testConnection();
    console.log('🔗 Connected:', connected);
    
    if (connected) {
      // Get health status
      const health = await plugin.getHealth();
      console.log('💓 Health status:', health.status);
      console.log('🏷️ Service:', health.service);
      console.log('📦 Version:', health.version);
      
      // Get service info
      const info = await plugin.getServiceInfo();
      console.log('📋 Service info:', {
        description: info.description,
        endpoints: Object.keys(info.endpoints),
      });
    }
  } catch (error) {
    console.error('❌ Health check failed:', error);
  }
}

// ╭─────────────────────────────────────────────────────────╮
// │                    Main Demo                           │
// ╰─────────────────────────────────────────────────────────╯

async function runDemo(): Promise<void> {
  console.log('🎬 Starting Motor WASM Service Worker Demo');
  console.log('==========================================');
  
  try {
    // Detect environment
    await detectEnvironment();
    console.log();
    
    // Create plugin (auto-detection)
    const plugin = await createPlugin();
    console.log();
    
    // Check service health
    await checkServiceHealth(plugin);
    console.log();
    
    // Demonstrate wallet operations
    await demonstrateWalletOperations(plugin);
    console.log();
    
    // Demonstrate DWN operations
    await demonstrateDWNOperations(plugin);
    console.log();
    
    // Cleanup
    console.log('🧹 Cleaning up...');
    await plugin.cleanup?.();
    console.log('✅ Demo completed successfully!');
    
  } catch (error) {
    console.error('❌ Demo failed:', error);
  }
}

// Export for use in other modules
export {
  detectEnvironment,
  createPlugin,
  createBrowserPlugin,
  createNodePlugin,
  demonstrateWalletOperations,
  demonstrateDWNOperations,
  checkServiceHealth,
  runDemo,
};

// Run demo if this file is executed directly
if (typeof require !== 'undefined' && require.main === module) {
  runDemo().catch(console.error);
}