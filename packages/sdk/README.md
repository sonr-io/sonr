# @sonr.io/sdk

This is the TypeScript SDK for Highway WebAuthn authentication gateway. It provides a comprehensive client library for interacting with Highway APIs, including WebAuthn authentication, session management, and blockchain operations.

## Features

- **WebAuthn Client**: Complete WebAuthn registration and authentication flows
- **Session Management**: Automatic session handling and token management
- **Type Safety**: Full TypeScript support with comprehensive type definitions
- **HTTP Client**: Configured HTTP client with automatic retries and error handling
- **Blockchain Integration**: Client methods for DID and blockchain operations
- **Environment Support**: Configurable for development, staging, and production

## Installation

```bash
npm install @sonr.io/sdk
# or
yarn add @sonr.io/sdk
# or
pnpm add @sonr.io/sdk
```

## Usage

### Basic Setup

```typescript
import { HighwaySDK } from "@sonr.io/sdk";

const sdk = new HighwaySDK({
  apiUrl: "https://api.yourdomain.com",
  environment: "production",
});
```

### WebAuthn Authentication

```typescript
// Register a new user
try {
  const registrationResult = await sdk.auth.register({
    username: "user@example.com",
    displayName: "John Doe",
  });

  console.log("Registration successful:", registrationResult);
} catch (error) {
  console.error("Registration failed:", error);
}

// Authenticate existing user
try {
  const authResult = await sdk.auth.authenticate({
    username: "user@example.com",
  });

  console.log("Authentication successful:", authResult);
} catch (error) {
  console.error("Authentication failed:", error);
}
```

### Session Management

```typescript
// Get current session
const session = await sdk.session.getCurrent();

// Validate session
const isValid = await sdk.session.validate();

// Logout
await sdk.session.logout();
```

### Blockchain Operations

```typescript
// Create DID
const didResult = await sdk.blockchain.createDID({
  username: "user@example.com",
  credentialId: "webauthn-credential-id",
  publicKey: "public-key-data",
});

// Resolve DID
const didDocument = await sdk.blockchain.resolveDID("did:sonr:123456789");
```

## API Reference

### HighwaySDK

The main SDK class that provides access to all Highway services.

#### Constructor Options

```typescript
interface HighwaySDKOptions {
  apiUrl: string;
  environment?: "development" | "staging" | "production";
  timeout?: number;
  retries?: number;
}
```

#### Methods

- `auth`: WebAuthn authentication methods
- `session`: Session management methods
- `blockchain`: Blockchain and DID operations
- `profile`: User profile management

### Authentication (`sdk.auth`)

- `register(options)`: Register new WebAuthn credential
- `authenticate(options)`: Authenticate with WebAuthn
- `getProfile()`: Get current user profile

### Session (`sdk.session`)

- `getCurrent()`: Get current session information
- `validate()`: Validate current session
- `logout()`: End current session

### Blockchain (`sdk.blockchain`)

- `createDID(options)`: Create new DID document
- `resolveDID(id)`: Resolve DID document by ID
- `getHealth()`: Get blockchain service health

## Development Roadmap

### 🔮 **Future Implementation** (Planned)

- [ ] **SDK Core**: TypeScript SDK with HTTP client configuration
- [ ] **Authentication Module**: WebAuthn registration and authentication methods
- [ ] **Session Management**: Automatic session handling and token management
- [ ] **Blockchain Module**: DID creation and resolution methods
- [ ] **Profile Management**: User profile operations and management
- [ ] **Type Definitions**: Comprehensive TypeScript type definitions
- [ ] **Error Handling**: Robust error handling and retry logic

### 🚧 **Production Readiness** (Next)

- [ ] **Testing Suite**: Unit tests for all SDK methods
- [ ] **Documentation**: Complete API documentation and examples
- [ ] **Performance Optimization**: Request caching and optimization
- [ ] **Bundle Optimization**: Tree-shaking and minimal bundle size
- [ ] **Browser Support**: Cross-browser compatibility testing
- [ ] **Node.js Support**: Server-side usage support

### 🔮 **Future Enhancements**

- [ ] **React Hooks**: React hooks for easy integration
- [ ] **Vue Composables**: Vue.js composables for Vue applications
- [ ] **Offline Support**: Offline capability with background sync
- [ ] **Real-time Features**: WebSocket support for real-time updates
- [ ] **Advanced Caching**: Intelligent caching strategies
- [ ] **Plugin System**: Extensible plugin architecture

## Error Handling

The SDK provides comprehensive error handling:

```typescript
import { HighwayError, WebAuthnError, SessionError } from "@sonr.io/sdk";

try {
  await sdk.auth.register(options);
} catch (error) {
  if (error instanceof WebAuthnError) {
    console.error("WebAuthn error:", error.message);
  } else if (error instanceof SessionError) {
    console.error("Session error:", error.message);
  } else if (error instanceof HighwayError) {
    console.error("Highway error:", error.message);
  }
}
```

## Configuration

### Environment Variables

```env
# Default API URL
HIGHWAY_API_URL=https://api.yourdomain.com

# Development
HIGHWAY_API_URL=http://localhost:8080
```

### SDK Configuration

```typescript
const sdk = new HighwaySDK({
  apiUrl: process.env.HIGHWAY_API_URL || "https://api.yourdomain.com",
  environment: process.env.NODE_ENV || "production",
  timeout: 10000,
  retries: 3,
});
```

## Development

```bash
# Install dependencies
pnpm install

# Build the SDK
pnpm build

# Run tests
pnpm test

# Run tests in watch mode
pnpm test:watch
```

## Integration Examples

### React Application

```typescript
import { HighwaySDK } from "@sonr.io/sdk";
import { useEffect, useState } from "react";

function useHighway() {
  const [sdk] = useState(
    () =>
      new HighwaySDK({
        apiUrl: process.env.NEXT_PUBLIC_API_URL!,
      })
  );

  return sdk;
}

function AuthComponent() {
  const sdk = useHighway();

  const handleRegister = async () => {
    try {
      await sdk.auth.register({
        username: "user@example.com",
        displayName: "John Doe",
      });
    } catch (error) {
      console.error("Registration failed:", error);
    }
  };

  return <button onClick={handleRegister}>Register with WebAuthn</button>;
}
```

### Node.js Application

```typescript
import { HighwaySDK } from "@sonr.io/sdk";

const sdk = new HighwaySDK({
  apiUrl: "https://api.yourdomain.com",
  environment: "production",
});

async function createUser() {
  try {
    const result = await sdk.auth.register({
      username: "user@example.com",
      displayName: "John Doe",
    });

    console.log("User created:", result);
  } catch (error) {
    console.error("Failed to create user:", error);
  }
}
```

## Dependencies

- **TypeScript**: ^5.3.0
- **Axios**: ^1.6.0 (HTTP client)
- **Zod**: ^3.22.0 (Schema validation)
- **@simplewebauthn/browser**: ^9.0.0 (WebAuthn client)

## Contributing

When contributing to the SDK:

1. Follow TypeScript best practices
2. Add comprehensive type definitions
3. Include unit tests for new features
4. Update documentation and examples
5. Ensure backward compatibility
