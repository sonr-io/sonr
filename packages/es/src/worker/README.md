# Motor WASM Service Worker Integration

This module provides TypeScript client integration for the Motor WASM service worker, enabling type-safe interactions with both DWN (Decentralized Web Node) and Wallet operations through a WebAssembly-based service worker or direct HTTP API calls.

## Features

- **🔧 Type-Safe Plugin Interface**: Mirrors the Go Plugin interface from `x/dwn/client/plugin`
- **🔄 Automatic Service Worker Management**: Handles registration, updates, and health monitoring
- **🌍 Cross-Environment Support**: Works in both browser and Node.js environments
- **🔄 Fallback Support**: Automatically falls back to HTTP when service workers are unavailable
- **🛡️ Comprehensive Error Handling**: Includes retries, timeouts, and detailed error reporting
- **🎯 Modern TypeScript**: Uses advanced TypeScript features for optimal developer experience

## Architecture

```
Motor WASM Integration
├── types.ts          # TypeScript type definitions
├── client.ts         # HTTP client for Motor API calls
├── worker.ts         # Service worker lifecycle management
├── plugin.ts         # Main plugin implementation
└── index.ts          # Entry point and exports
```

## Quick Start

### Basic Usage (Auto-detection)

```typescript
import { createMotorPlugin } from '@sonr.io/es/client/motor';

const plugin = await createMotorPlugin({
  debug: true,
  timeout: 30000,
});

// Create a UCAN token
const tokenResponse = await plugin.newOriginToken({
  audience_did: 'did:sonr:example',
  attenuations: [{ can: ['sign'], with: 'vault://example' }],
});

console.log('Token created:', tokenResponse.token);
```

### Browser-Specific Usage

```typescript
import { createMotorPluginForBrowser } from '@sonr.io/es/client/motor';

const plugin = await createMotorPluginForBrowser('/motor-worker', {
  auto_register_worker: true,
  prefer_service_worker: true,
  debug: true,
});
```

### Node.js Usage

```typescript
import { createMotorPluginForNode } from '@sonr.io/es/client/motor';

const plugin = await createMotorPluginForNode('http://localhost:8080', {
  timeout: 10000,
});
```

## Environment Detection

The plugin automatically detects the environment and capabilities:

```typescript
import { 
  isMotorSupported, 
  getMotorEnvironment,
  MotorServiceWorkerManager 
} from '@sonr.io/es/client/motor';

// Check overall support
const supported = isMotorSupported();

// Get detailed environment info
const env = getMotorEnvironment();
console.log('Browser:', env.is_browser);
console.log('Service Worker Support:', env.supports_service_worker);
console.log('WebAssembly Support:', env.supports_wasm);

// Get browser compatibility details
const compatibility = MotorServiceWorkerManager.getBrowserCompatibility();
if (!compatibility.compatible) {
  console.warn('Issues:', compatibility.issues);
  console.log('Recommendations:', compatibility.recommendations);
}
```

## API Reference

### Plugin Interface

The `MotorPlugin` interface provides access to all Motor operations:

#### UCAN Token Operations

```typescript
// Create origin token
const originToken = await plugin.newOriginToken({
  audience_did: 'did:sonr:audience',
  attenuations: [{ can: ['sign', 'encrypt'], with: 'vault://example' }],
  facts: ['example-fact'],
  expires_at: Date.now() + 86400000, // 24 hours
});

// Create attenuated token
const attenuatedToken = await plugin.newAttenuatedToken({
  parent_token: originToken.token,
  audience_did: 'did:sonr:delegated',
  attenuations: [{ can: ['sign'], with: 'vault://limited' }],
});
```

#### Cryptographic Operations

```typescript
// Sign data
const data = new TextEncoder().encode('Hello, World!');
const signature = await plugin.signData({ data });

// Verify signature
const verification = await plugin.verifyData({
  data,
  signature: signature.signature,
});
console.log('Valid:', verification.valid);
```

#### Identity Operations

```typescript
// Get issuer DID
const issuer = await plugin.getIssuerDID();
console.log('DID:', issuer.issuer_did);
console.log('Address:', issuer.address);
```

#### DWN Operations

```typescript
// Create record
const record = await plugin.createRecord({
  target: 'did:sonr:alice',
  data: new TextEncoder().encode('Record data'),
  schema: 'https://schema.org/Message',
  published: true,
  encrypt: true,
});

// Read record
const readResponse = await plugin.readRecord(record.record_id, 'did:sonr:alice');

// Update record
await plugin.updateRecord({
  record_id: record.record_id,
  target: 'did:sonr:alice',
  data: new TextEncoder().encode('Updated data'),
});

// Delete record
await plugin.deleteRecord(record.record_id, 'did:sonr:alice');
```

### Service Worker Management

```typescript
import { MotorServiceWorkerManager } from '@sonr.io/es/client/motor';

const manager = new MotorServiceWorkerManager({
  worker_script: '/motor-worker.js',
  scope: '/motor-worker',
  debug: true,
});

// Register service worker
const status = await manager.register();
console.log('Registered:', status.registered);

// Check status
const currentStatus = manager.getStatus();
console.log('State:', currentStatus.state);

// Update service worker
await manager.update();

// Send message to service worker
await manager.sendMessage({ type: 'custom-message', data: 'hello' });

// Cleanup
await manager.unregister();
```

### Direct HTTP Client

```typescript
import { MotorClient } from '@sonr.io/es/client/motor';

const client = new MotorClient({
  worker_url: 'http://localhost:8080',
  timeout: 15000,
  max_retries: 2,
});

// Test connection
const connected = await client.testConnection();

// Get service info
const info = await client.getServiceInfo();

// Health check
const health = await client.healthCheck();
```

## Configuration

### MotorPluginConfig

```typescript
interface MotorPluginConfig extends MotorServiceWorkerConfig {
  /** Whether to automatically register the service worker */
  auto_register_worker?: boolean;
  /** Whether to use service worker when available */
  prefer_service_worker?: boolean;
  /** Fallback configuration for direct HTTP calls */
  fallback_url?: string;
}
```

### MotorServiceWorkerConfig

```typescript
interface MotorServiceWorkerConfig {
  /** URL where the Motor WASM service worker is available */
  worker_url?: string;
  /** Timeout for HTTP requests in milliseconds */
  timeout?: number;
  /** Maximum number of retry attempts */
  max_retries?: number;
  /** Whether to enable debug logging */
  debug?: boolean;
}
```

## Error Handling

The plugin includes comprehensive error handling:

```typescript
try {
  const plugin = await createMotorPlugin();
  const result = await plugin.signData({ data: new Uint8Array([1, 2, 3]) });
} catch (error) {
  if (error.message.includes('Service worker not supported')) {
    // Handle service worker unavailability
    console.log('Falling back to HTTP client');
  } else if (error.message.includes('timeout')) {
    // Handle timeout
    console.log('Request timed out, retrying with increased timeout');
  } else {
    // Handle other errors
    console.error('Operation failed:', error);
  }
}
```

## Event Handling

Listen for service worker events:

```typescript
// Service worker update available
window.addEventListener('motor-update-available', (event) => {
  console.log('New service worker version available');
  // Prompt user to refresh
});

// Service worker health failure
window.addEventListener('motor-health-failure', (event) => {
  console.warn('Service worker health check failed');
  // Implement fallback strategy
});
```

## Browser Compatibility

### Supported Browsers

- Chrome 40+ (Service Workers)
- Firefox 44+ (Service Workers)
- Safari 11.1+ (Service Workers)
- Edge 17+ (Service Workers)

### Required Features

- **WebAssembly**: For WASM execution
- **Service Workers**: For background operations (browser only)
- **Fetch API**: For HTTP communication
- **Secure Context**: HTTPS required for service workers

### Polyfills

For older browsers, consider including:

```html
<!-- WebAssembly polyfill -->
<script src="https://unpkg.com/@webassemblyjs/wasm-polyfill"></script>

<!-- Fetch polyfill -->
<script src="https://unpkg.com/whatwg-fetch"></script>
```

## Performance Considerations

### Service Worker Benefits

- **Background Processing**: Operations continue even when page is not active
- **Caching**: Reduces network requests for repeated operations
- **Offline Support**: Some operations can work offline
- **Resource Sharing**: Multiple tabs share the same service worker instance

### HTTP Fallback Benefits

- **Lower Latency**: Direct communication without service worker overhead
- **Simpler Debugging**: Easier to trace network requests
- **Better Error Handling**: More predictable error responses

### Optimization Tips

1. **Use appropriate timeouts** based on operation complexity
2. **Enable debug mode** during development
3. **Monitor service worker health** in production
4. **Implement proper retry logic** for transient failures
5. **Use connection pooling** for multiple operations

## Security Considerations

### Service Worker Security

- Service workers must be served over HTTPS in production
- Same-origin policy applies to service worker registration
- Service worker scope determines accessible resources

### Data Protection

- All cryptographic operations occur in the WASM enclave
- Private keys never leave the secure execution environment
- Data encryption uses consensus-based algorithms

### Network Security

- All HTTP communications use secure protocols
- Request/response data is validated and sanitized
- Retry logic includes exponential backoff to prevent DoS

## Troubleshooting

### Common Issues

#### Service Worker Registration Fails

```typescript
// Check if running in secure context
if (!window.isSecureContext) {
  console.error('Service workers require HTTPS or localhost');
}

// Check browser support
if (!('serviceWorker' in navigator)) {
  console.error('Service workers not supported');
}
```

#### HTTP Connection Fails

```typescript
// Verify server is running
const client = new MotorClient({ worker_url: 'http://localhost:8080' });
const connected = await client.testConnection();
if (!connected) {
  console.error('Motor service not available at configured URL');
}
```

#### Type Errors

```typescript
// Ensure proper type imports
import type { 
  MotorPlugin, 
  NewOriginTokenRequest,
  UCANTokenResponse 
} from '@sonr.io/es/client/motor';
```

### Debug Mode

Enable debug logging for detailed operation information:

```typescript
const plugin = await createMotorPlugin({
  debug: true, // Enable debug logging
  timeout: 30000,
  max_retries: 3,
});
```

### Health Monitoring

Implement health monitoring for production:

```typescript
const plugin = await createMotorPlugin();

// Periodic health checks
setInterval(async () => {
  const connected = await plugin.testConnection();
  if (!connected) {
    console.warn('Motor service connection lost');
    // Implement reconnection logic
  }
}, 30000);
```

## Contributing

When contributing to the Motor integration:

1. **Follow TypeScript best practices**
2. **Add comprehensive JSDoc documentation**
3. **Include unit tests for new functionality**
4. **Update type definitions when changing APIs**
5. **Test in both browser and Node.js environments**

## License

This module is part of the Sonr project and follows the same licensing terms.