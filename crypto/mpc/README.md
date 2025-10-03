# MPC (Multi-Party Computation) Cryptographic Library

![Go](https://img.shields.io/badge/Go-1.24+-green)
![MPC](https://img.shields.io/badge/MPC-Threshold_Signing-blue)
![Encryption](https://img.shields.io/badge/Encryption-AES--GCM-red)
![ECDSA](https://img.shields.io/badge/Curve-secp256k1-yellow)

A comprehensive Go implementation of Multi-Party Computation (MPC) primitives for secure distributed cryptography. This package provides threshold signing, encrypted key management, and secure keyshare operations for decentralized applications.

## Features

- ✅ **Threshold Cryptography** - 2-of-2 MPC key generation and signing
- ✅ **Secure Enclaves** - Encrypted keyshare storage and management
- ✅ **Multiple Curves** - Support for secp256k1, P-256, Ed25519, BLS12-381, and more
- ✅ **Key Refresh** - Proactive security through keyshare rotation
- ✅ **ECDSA Signing** - Distributed signature generation with SHA3-256
- ✅ **Encrypted Export/Import** - Secure enclave serialization with AES-GCM
- ✅ **UCAN Integration** - MPC-based JWT signing for User-Controlled Authorization Networks

## Architecture

The package is built around the concept of secure **Enclaves** that manage distributed keyshares:

```
┌─────────────────────────────────────────────────────────┐
│                    MPC Enclave                          │
├─────────────────────────────────────────────────────────┤
│  ┌─────────────┐    ┌─────────────┐                    │
│  │ Alice Share │    │  Bob Share  │  ←── Threshold 2/2  │
│  │ (Validator) │    │   (User)    │                     │
│  └─────────────┘    └─────────────┘                     │
├─────────────────────────────────────────────────────────┤
│  • Distributed Key Generation (DKG)                    │
│  • Threshold Signing (2-of-2)                          │
│  • Key Refresh (Proactive Security)                    │
│  • Encrypted Storage (AES-GCM)                         │
└─────────────────────────────────────────────────────────┘
```

## Quick Start

### Installation

```bash
go get github.com/sonr-io/sonr/crypto/mpc
```

### Basic Usage

#### Creating a New MPC Enclave

```go
package main

import (
    "fmt"
    "github.com/sonr-io/sonr/crypto/mpc"
)

func main() {
    // Generate a new MPC enclave with distributed keyshares
    enclave, err := mpc.NewEnclave()
    if err != nil {
        panic(err)
    }

    // Get the public key
    pubKeyHex := enclave.PubKeyHex()
    fmt.Printf("Public Key: %s\n", pubKeyHex)

    // Verify the enclave is valid
    if enclave.IsValid() {
        fmt.Println("✅ Enclave successfully created!")
    }
}
```

#### Signing and Verification

```go
// Sign data using distributed MPC protocol
message := []byte("Hello, distributed world!")
signature, err := enclave.Sign(message)
if err != nil {
    panic(err)
}

// Verify the signature
isValid, err := enclave.Verify(message, signature)
if err != nil {
    panic(err)
}

fmt.Printf("Signature valid: %t\n", isValid)
```

#### Secure Export and Import

```go
// Export enclave with encryption
secretKey := []byte("my-super-secret-key-32-bytes-long")
encryptedData, err := enclave.Encrypt(secretKey)
if err != nil {
    panic(err)
}

// Import from encrypted data
restoredEnclave, err := mpc.ImportEnclave(
    mpc.WithEncryptedData(encryptedData, secretKey),
)
if err != nil {
    panic(err)
}

fmt.Printf("Restored public key: %s\n", restoredEnclave.PubKeyHex())
```

## Core Concepts

### Enclaves

An **Enclave** represents a secure MPC keyshare environment that manages distributed cryptographic operations:

```go
type Enclave interface {
    // Key Management
    PubKeyHex() string         // Get public key as hex string
    PubKeyBytes() []byte       // Get public key as bytes
    IsValid() bool             // Check if enclave has valid keyshares
    
    // Cryptographic Operations
    Sign(data []byte) ([]byte, error)        // Threshold signing
    Verify(data []byte, sig []byte) (bool, error) // Signature verification
    Refresh() (Enclave, error)               // Proactive key refresh
    
    // Secure Storage
    Encrypt(key []byte) ([]byte, error)      // Export encrypted
    Decrypt(key []byte, data []byte) ([]byte, error) // Import encrypted
    
    // Serialization
    Marshal() ([]byte, error)                // JSON serialization
    Unmarshal(data []byte) error            // JSON deserialization
    
    // Data Access
    GetData() *EnclaveData                   // Access enclave internals
    GetEnclave() Enclave                     // Self-reference
}
```

### Multi-Party Computation Protocol

The package implements a 2-of-2 threshold scheme:

1. **Alice (Validator)** - Server-side keyshare
2. **Bob (User)** - Client-side keyshare

Both parties must participate in:
- **Distributed Key Generation (DKG)** - Creates shared public key
- **Threshold Signing** - Generates valid signatures cooperatively
- **Key Refresh** - Rotates keyshares while preserving public key

### Supported Curves

The package supports multiple elliptic curves:

```go
type CurveName string

const (
    K256Name       CurveName = "secp256k1"  // Bitcoin/Ethereum
    P256Name       CurveName = "P-256"      // NIST P-256
    ED25519Name    CurveName = "ed25519"    // EdDSA
    BLS12381G1Name CurveName = "BLS12381G1" // BLS12-381 G1
    BLS12381G2Name CurveName = "BLS12381G2" // BLS12-381 G2
    // ... more curves supported
)
```

## Advanced Usage

### Custom Import Options

The package provides flexible import mechanisms:

```go
// Import from initial keyshares (after DKG)
enclave, err := mpc.ImportEnclave(
    mpc.WithInitialShares(validatorShare, userShare, mpc.K256Name),
)

// Import from existing enclave data
enclave, err := mpc.ImportEnclave(
    mpc.WithEnclaveData(enclaveData),
)

// Import from encrypted backup
enclave, err := mpc.ImportEnclave(
    mpc.WithEncryptedData(encryptedBytes, secretKey),
)
```

### Key Refresh for Proactive Security

```go
// Refresh keyshares while keeping the same public key
refreshedEnclave, err := enclave.Refresh()
if err != nil {
    panic(err)
}

// Public key remains the same
fmt.Printf("Original:  %s\n", enclave.PubKeyHex())
fmt.Printf("Refreshed: %s\n", refreshedEnclave.PubKeyHex())
// Both should be identical!

// But the enclave now has fresh keyshares
// This provides forward secrecy against key compromise
```

### Standalone Verification

```go
// Verify signatures without the full enclave
pubKeyBytes := enclave.PubKeyBytes()
isValid, err := mpc.VerifyWithPubKey(pubKeyBytes, message, signature)
if err != nil {
    panic(err)
}
```

### UCAN Integration

The package includes MPC-based JWT signing for UCAN tokens:

```go
import "github.com/sonr-io/sonr/crypto/mpc/spec"

// Create MPC-backed UCAN token source
// (Implementation details in spec package)
keyshareSource := spec.KeyshareSource{
    // ... MPC enclave integration
}

// Use with UCAN token creation
token, err := keyshareSource.NewOriginToken(
    "did:key:audience",
    attenuations,
    facts,
    notBefore,
    expires,
)
```

## Security Features

### Encryption

All encrypted operations use **AES-GCM** with **SHA3-256** key derivation:

```go
// Secure key derivation
func GetHashKey(key []byte) []byte {
    hash := sha3.New256()
    hash.Write(key)
    return hash.Sum(nil)[:32] // 256-bit key
}
```

### Threshold Security

- **2-of-2 threshold** - Both parties required for operations
- **No single point of failure** - Neither party alone can sign
- **Proactive refresh** - Regular keyshare rotation without changing public key
- **Forward secrecy** - Old keyshares cannot be used after refresh

### Cryptographic Primitives

- **ECDSA Signing** with **SHA3-256** message hashing
- **AES-GCM** encryption with 12-byte nonces
- **Secure random nonce generation**
- **Multiple curve support** for different use cases

## Public API Reference

### Core Functions

```go
// Generate new MPC enclave
func NewEnclave() (Enclave, error)

// Import enclave from various sources
func ImportEnclave(options ...ImportOption) (Enclave, error)

// Execute distributed signing protocol
func ExecuteSigning(signFuncVal SignFunc, signFuncUser SignFunc) ([]byte, error)

// Execute keyshare refresh protocol
func ExecuteRefresh(refreshFuncVal RefreshFunc, refreshFuncUser RefreshFunc, 
                   curve CurveName) (Enclave, error)

// Standalone signature verification
func VerifyWithPubKey(pubKeyCompressed []byte, data []byte, sig []byte) (bool, error)
```

### Import Options

```go
type ImportOption func(Options) Options

// Create from initial DKG results
func WithInitialShares(valKeyshare Message, userKeyshare Message, 
                      curve CurveName) ImportOption

// Create from encrypted backup
func WithEncryptedData(data []byte, key []byte) ImportOption

// Create from existing data structure
func WithEnclaveData(data *EnclaveData) ImportOption
```

### EnclaveData Structure

```go
type EnclaveData struct {
    PubHex    string    `json:"pub_hex"`   // Compressed public key (hex)
    PubBytes  []byte    `json:"pub_bytes"` // Uncompressed public key
    ValShare  Message   `json:"val_share"` // Alice (validator) keyshare
    UserShare Message   `json:"user_share"`// Bob (user) keyshare
    Nonce     []byte    `json:"nonce"`     // Encryption nonce
    Curve     CurveName `json:"curve"`     // Elliptic curve name
}
```

### Protocol Types

```go
type Message *protocol.Message              // MPC protocol message
type Signature *curves.EcdsaSignature       // ECDSA signature
type RefreshFunc interface{ protocol.Iterator } // Key refresh protocol
type SignFunc interface{ protocol.Iterator }    // Signing protocol
type Point curves.Point                      // Elliptic curve point
```

### Utility Functions

```go
// Cryptographic utilities
func GetHashKey(key []byte) []byte
func SerializeSignature(sig *curves.EcdsaSignature) ([]byte, error)
func DeserializeSignature(sigBytes []byte) (*curves.EcdsaSignature, error)

// Key conversion utilities
func GetECDSAPoint(pubKey []byte) (*curves.EcPoint, error)

// Protocol error handling
func CheckIteratedErrors(aErr, bErr error) error
```

## Error Handling

The package provides comprehensive error handling:

```go
// Common error patterns
enclave, err := mpc.NewEnclave()
if err != nil {
    // Handle DKG failure
    log.Fatalf("Failed to generate enclave: %v", err)
}

signature, err := enclave.Sign(data)
if err != nil {
    // Handle signing protocol failure
    log.Fatalf("Failed to sign: %v", err)
}

// Validation errors
if !enclave.IsValid() {
    log.Fatal("Enclave has invalid keyshares")
}
```

## Performance Considerations

### Memory Usage

- **Minimal footprint** - Only active keyshares kept in memory
- **Efficient serialization** - JSON-based with compression
- **Secure cleanup** - Sensitive data cleared after use

### Network Communication

- **Minimal rounds** - Optimized protocol with few message exchanges
- **Small messages** - Compact protocol message format
- **Stateless operations** - No persistent connections required

### Cryptographic Performance

- **Hardware acceleration** - Leverages optimized curve implementations
- **Efficient hashing** - SHA3-256 with minimal overhead
- **Fast verification** - Public key operations optimized

## Testing

The package includes comprehensive tests:

```bash
# Run all tests
go test -v ./crypto/mpc

# Run specific test suites
go test -v ./crypto/mpc -run TestEnclaveData
go test -v ./crypto/mpc -run TestKeyShareGeneration
go test -v ./crypto/mpc -run TestEnclaveOperations

# Run with race detection
go test -race ./crypto/mpc

# Generate coverage report
go test -cover ./crypto/mpc
```

## Use Cases

### Decentralized Identity

- **DID key management** - Secure distributed identity keys
- **Threshold signing** - Multi-party authorization for identity operations
- **Key recovery** - Distributed backup and restore mechanisms

### Cryptocurrency Wallets

- **Multi-signature wallets** - True threshold custody solutions
- **Exchange security** - Hot wallet protection with distributed keys
- **Institutional custody** - Compliance-friendly key management

### Blockchain Infrastructure

- **Validator signing** - Secure consensus participation
- **Cross-chain bridges** - Multi-party custody of bridged assets
- **DAO governance** - Distributed decision-making mechanisms

### Enterprise Applications

- **Document signing** - Distributed digital signatures
- **API authentication** - Threshold-based service authentication
- **Secure communication** - End-to-end encrypted messaging

## Dependencies

- **Core Cryptography**: `github.com/sonr-io/sonr/crypto/core/curves`
- **Protocol Framework**: `github.com/sonr-io/sonr/crypto/core/protocol`
- **Threshold ECDSA**: `github.com/sonr-io/sonr/crypto/tecdsa/dklsv1`
- **UCAN Integration**: `github.com/sonr-io/sonr/crypto/ucan`
- **Standard Crypto**: `golang.org/x/crypto/sha3`
- **JWT Support**: `github.com/golang-jwt/jwt`

## Security Considerations

### Threat Model

The package is designed to protect against:

- **Key compromise** - Distributed keyshares prevent single points of failure
- **Insider threats** - No single party can perform operations alone
- **Network attacks** - Protocol messages are cryptographically protected
- **Side-channel attacks** - Secure implementations of cryptographic primitives

### Best Practices

1. **Regular key refresh** - Rotate keyshares periodically
2. **Secure communication** - Use TLS for protocol message exchange
3. **Access controls** - Implement proper authentication for MPC operations
4. **Audit logging** - Log all cryptographic operations
5. **Backup strategies** - Securely store encrypted enclave exports

### Limitations

- **2-of-2 threshold only** - Currently supports only 2-party protocols
- **Network dependency** - Requires communication between parties
- **No byzantine fault tolerance** - Assumes honest-but-curious adversaries

## Contributing

We welcome contributions! Please ensure:

1. **Security first** - All cryptographic code must be carefully reviewed
2. **Comprehensive testing** - Include unit tests and integration tests
3. **Documentation** - Document all public APIs and security assumptions
4. **Performance** - Benchmark critical cryptographic operations
5. **Compatibility** - Maintain backward compatibility with existing enclaves

## License

This project follows the same license as the main Sonr project.

---

**⚠️ Security Notice**: This is cryptographic software. While extensively tested, it should be used with appropriate security measures and understanding of the underlying protocols. For production deployments, consider additional security audits and operational security measures.