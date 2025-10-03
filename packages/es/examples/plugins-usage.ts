/**
 * Example demonstrating the usage of plugins from @sonr.io/es
 */

// Import plugins module
import { plugins } from '@sonr.io/es';

// Or import specific plugins
import { createVaultClient, createMotorPlugin } from '@sonr.io/es/plugins';

// Or import plugins directly
import { VaultClient } from '@sonr.io/es/plugins/vault';
import { MotorPluginImpl } from '@sonr.io/es/plugins/motor';

async function demonstrateVaultPlugin() {
  console.log('=== Vault Plugin Demo ===');
  
  // Create a vault client
  const vault = createVaultClient({
    chainId: 'sonr-testnet-1',
    // enclave configuration would go here
  });

  // Initialize the vault
  await vault.initialize();

  // Get issuer DID
  const issuerInfo = await vault.getIssuerDID();
  console.log('Issuer DID:', issuerInfo.issuer_did);
  console.log('Address:', issuerInfo.address);

  // Create a UCAN token
  const tokenResponse = await vault.newOriginToken({
    audience_did: 'did:sonr:example123',
    attenuations: [
      { can: ['sign', 'verify'], with: 'vault://keys/*' }
    ],
    expires_at: Date.now() + 3600000, // 1 hour from now
  });
  console.log('UCAN Token created:', tokenResponse.token.substring(0, 50) + '...');

  // Sign some data
  const dataToSign = new TextEncoder().encode('Hello, Sonr!');
  const signature = await vault.signData({ data: dataToSign });
  console.log('Signature created');

  // Verify the signature
  const verification = await vault.verifyData({
    data: dataToSign,
    signature: signature.signature,
  });
  console.log('Signature valid:', verification.valid);

  // Clean up
  await vault.cleanup();
}

async function demonstrateMotorPlugin() {
  console.log('\n=== Motor Plugin Demo ===');
  
  // Create a motor plugin (auto-detects environment)
  const motor = await createMotorPlugin({
    debug: true,
    timeout: 30000,
  });

  // Check if motor is ready
  const isReady = await motor.isReady();
  console.log('Motor ready:', isReady);

  // Get service info
  const serviceInfo = await motor.getServiceInfo();
  console.log('Service version:', serviceInfo.version);
  console.log('Service status:', serviceInfo.status);

  // Create a DWN record
  const record = await motor.createRecord({
    data: { message: 'Hello from Motor!' },
    published: false,
    schema: 'https://schema.org/Message',
    dataFormat: 'application/json',
  });
  console.log('Record created:', record.record_id);

  // Read the record back
  const readResult = await motor.readRecord(record.record_id);
  console.log('Record data:', readResult.data);

  // Create a UCAN token using motor
  const tokenResponse = await motor.newOriginToken({
    audience_did: 'did:sonr:motor123',
    attenuations: [
      { can: ['create', 'read', 'update', 'delete'], with: 'dwn://records/*' }
    ],
  });
  console.log('Motor UCAN Token:', tokenResponse.token.substring(0, 50) + '...');

  // Clean up
  await motor.cleanup();
}

async function demonstratePluginNamespaces() {
  console.log('\n=== Using Plugin Namespaces ===');
  
  // Access plugins through namespace
  const vaultClient = plugins.vault.createVaultClient();
  const motorPlugin = await plugins.motor.createMotorPlugin();

  console.log('Vault client created via namespace');
  console.log('Motor plugin created via namespace');

  // Use the plugins...
  // ...

  // Clean up
  await vaultClient.cleanup();
  await motorPlugin.cleanup();
}

// Main execution
async function main() {
  try {
    await demonstrateVaultPlugin();
    await demonstrateMotorPlugin();
    await demonstratePluginNamespaces();
    
    console.log('\n✅ All plugin demonstrations completed successfully!');
  } catch (error) {
    console.error('❌ Error during plugin demonstration:', error);
  }
}

// Run if executed directly
if (import.meta.url === `file://${process.argv[1]}`) {
  main();
}