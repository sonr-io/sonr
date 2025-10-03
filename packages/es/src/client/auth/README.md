# Sonr WebAuthn Authentication Module

This module provides WebAuthn helper methods for integrating passwordless authentication with the Sonr blockchain. It wraps the `@simplewebauthn/browser` library and provides convenient methods for registering and authenticating users with WebAuthn credentials.

## Installation

The auth module is part of the `@sonr.io/es` package:

```bash
npm install @sonr.io/es
# or
pnpm add @sonr.io/es
```

## Usage

```typescript
import { 
  registerWebAuthn, 
  loginWebAuthn,
  isWebAuthnSupported 
} from '@sonr.io/es/auth';
```

## Core Functions

### Registration

#### `registerWebAuthn(apiUrl, options)`
Performs a complete WebAuthn registration flow.

```typescript
await registerWebAuthn('http://localhost:8080', {
  username: 'alice',
  rpId: 'localhost',
  rpName: 'Sonr Local',
  timeout: 60000 // optional, defaults to 60 seconds
});
```

#### `beginRegistration(apiUrl, options)`
Initiates the registration process and returns credential creation options.

#### `finishRegistration(apiUrl, username, credential)`
Completes the registration by sending the credential to the server for verification.

### Authentication

#### `loginWebAuthn(apiUrl, options)`
Performs a complete WebAuthn authentication flow.

```typescript
const result = await loginWebAuthn('http://localhost:8080', {
  username: 'alice',
  rpId: 'localhost',
  timeout: 30000 // optional, defaults to 30 seconds
});

if (result.success) {
  console.log('Login successful');
}
```

#### `beginLogin(apiUrl, options)`
Initiates the authentication process and returns credential request options.

#### `finishLogin(apiUrl, username, credential)`
Completes the authentication by verifying the credential with the server.

### Utility Functions

#### `isWebAuthnSupported()`
Checks if the browser supports WebAuthn.

```typescript
if (isWebAuthnSupported()) {
  // WebAuthn is available
}
```

#### `isWebAuthnAvailable()`
Checks if a platform authenticator is available (e.g., Touch ID, Face ID, Windows Hello).

```typescript
const available = await isWebAuthnAvailable();
if (available) {
  // Platform authenticator is available
}
```

#### `isConditionalMediationAvailable()`
Checks if conditional mediation (autofill) is supported for seamless authentication.

```typescript
const autofillSupported = await isConditionalMediationAvailable();
```

#### `bufferToBase64url(buffer)`
Converts an ArrayBuffer to a base64url-encoded string.

#### `base64urlToBuffer(base64url)`
Converts a base64url-encoded string to an ArrayBuffer.

## Advanced Usage

### Custom Registration Flow

```typescript
import { 
  beginRegistration, 
  finishRegistration 
} from '@sonr.io/es/auth';
import { startRegistration } from '@simplewebauthn/browser';

// Step 1: Get registration options
const options = await beginRegistration('http://localhost:8080', {
  username: 'alice',
  rpId: 'localhost',
  rpName: 'Sonr Local'
});

// Step 2: Create credential (with custom UI feedback)
console.log('Touch your security key...');
const credential = await startRegistration(options);

// Step 3: Verify with server
await finishRegistration('http://localhost:8080', 'alice', credential);
```

### Custom Authentication Flow

```typescript
import { 
  beginLogin, 
  finishLogin 
} from '@sonr.io/es/auth';
import { startAuthentication } from '@simplewebauthn/browser';

// Step 1: Get authentication options
const options = await beginLogin('http://localhost:8080', {
  username: 'alice',
  rpId: 'localhost'
});

// Step 2: Get credential from authenticator
const credential = await startAuthentication(options);

// Step 3: Verify with server
const result = await finishLogin('http://localhost:8080', 'alice', credential);
```

## Server Requirements

The WebAuthn helpers expect a server that implements the following endpoints:

- `GET /begin-register?username={username}` - Returns credential creation options
- `POST /finish-register?username={username}` - Verifies and stores the credential
- `GET /begin-login?username={username}` - Returns credential request options
- `POST /finish-login?username={username}` - Verifies the authentication credential

These endpoints are implemented in the Sonr blockchain's x/did module WebAuthn server.

## Browser Compatibility

WebAuthn is supported in modern browsers:
- Chrome/Edge 67+
- Firefox 60+
- Safari 14+

Platform authenticator support varies by device:
- macOS: Touch ID (MacBooks), Face ID (supported iPads)
- Windows: Windows Hello
- Android: Fingerprint, Face unlock
- iOS: Touch ID, Face ID

## Security Considerations

1. Always use HTTPS in production (WebAuthn requires secure contexts)
2. The `rpId` must match the domain where the authentication is performed
3. Store credentials securely on the server
4. Implement proper session management after successful authentication
5. Consider implementing backup authentication methods

## Examples

See `examples.ts` for complete usage examples including:
- Checking WebAuthn support
- Simple registration and authentication
- Advanced flows with custom handling
- Conditional UI authentication (autofill)