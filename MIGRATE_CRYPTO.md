# Sonr Cryptography Library Migration Context

> **Repository Migration**: `sonr-io/sonr/crypto/` → `sonr-io/crypto`
> **Package Name**: `github.com/sonr-io/crypto`
> **Current Version**: `v1.0.1`

## Overview

The Sonr Cryptography Library is a comprehensive collection of cryptographic primitives designed for secure decentralized applications. It provides enterprise-grade implementations of elliptic curve cryptography, multi-party computation, threshold cryptography, zero-knowledge proofs, and advanced signature schemes.

**Key Features:**
- **Multi-Party Computation (MPC)**: Secure distributed key generation and signing
- **Threshold Cryptography**: TECDSA and TED25519 with FROST protocol
- **Advanced Signatures**: BLS aggregation, BBS+ selective disclosure, Schnorr variants
- **Secret Sharing**: Shamir, Feldman VSS, Pedersen VSS implementations
- **Zero-Knowledge Proofs**: Bulletproofs for range proofs
- **Multiple Elliptic Curves**: Ed25519, Secp256k1, P-256, BLS12-381, Pallas/Vesta
- **UCAN Integration**: Capability-based authorization tokens
- **DID Key Management**: Multi-chain wallet address derivation

## Repository Structure

```
sonr-io/crypto/
├── core/                    # Core cryptographic primitives
│   ├── curves/             # Elliptic curve implementations
│   │   ├── native/         # Native curve arithmetic
│   │   │   ├── bls12381/   # BLS12-381 pairing-friendly curve
│   │   │   ├── k256/       # Secp256k1 (Bitcoin/Ethereum)
│   │   │   ├── p256/       # NIST P-256
│   │   │   └── pasta/      # Pallas/Vesta for ZK proofs
│   │   ├── bls12377_curve.go
│   │   ├── bls12381_curve.go
│   │   ├── ed25519_curve.go
│   │   ├── k256_curve.go
│   │   ├── p256_curve.go
│   │   └── pallas_curve.go
│   ├── protocol/           # MPC protocol framework
│   ├── commit.go           # Pedersen commitments
│   ├── hash.go             # Cryptographic hash utilities
│   └── mod.go              # Modular arithmetic
│
├── mpc/                     # Multi-Party Computation
│   ├── enclave.go          # MPC enclave management
│   ├── protocol.go         # DKG and signing protocols
│   ├── codec.go            # Serialization/deserialization
│   ├── import.go           # Enclave import/export
│   ├── verify.go           # Signature verification
│   └── spec/               # UCAN/JWT specifications
│
├── tecdsa/                  # Threshold ECDSA
│   └── dklsv1/             # 2-party ECDSA (DKLS v1)
│       ├── dkg/            # Distributed key generation
│       ├── sign/           # Threshold signing
│       ├── refresh/        # Key refresh protocol
│       └── dealer/         # Trusted dealer mode
│
├── ted25519/                # Threshold Ed25519
│   ├── frost/              # FROST protocol (DKG + signing)
│   └── ted25519/           # Core threshold Ed25519
│
├── signatures/              # Digital signature schemes
│   ├── bls/                # BLS signatures
│   │   └── bls_sig/        # Aggregatable BLS
│   ├── bbs/                # BBS+ selective disclosure
│   ├── schnorr/            # Schnorr variants
│   │   ├── mina/           # Mina protocol integration
│   │   └── nem/            # NEM blockchain support
│   └── common/             # Shared signature utilities
│
├── sharing/                 # Secret sharing schemes
│   ├── shamir.go           # Shamir's Secret Sharing
│   ├── feldman.go          # Feldman VSS
│   ├── pedersen.go         # Pedersen VSS
│   └── v1/                 # Version 1 implementations
│
├── dkg/                     # Distributed Key Generation
│   ├── frost/              # FROST DKG for Ed25519
│   ├── gennaro/            # Gennaro DKG protocol
│   └── gennaro2p/          # 2-party simplified DKG
│
├── bulletproof/             # Bulletproofs (range proofs)
│   ├── range_prover.go     # Range proof generation
│   ├── range_verifier.go   # Range proof verification
│   ├── ipp_prover.go       # Inner product argument
│   └── generators.go       # Generator points
│
├── accumulator/             # Cryptographic accumulators
│   ├── accumulator.go      # RSA accumulator
│   ├── witness.go          # Membership witnesses
│   └── proof.go            # Inclusion/exclusion proofs
│
├── paillier/                # Paillier homomorphic encryption
│   ├── paillier.go         # Public/private key operations
│   └── psf.go              # Proof of safe factorization
│
├── ot/                      # Oblivious Transfer
│   ├── base/simplest/      # Simplest OT protocol
│   └── extension/kos/      # KOS OT extension
│
├── zkp/                     # Zero-Knowledge Proofs
│   └── schnorr/            # Schnorr proofs of knowledge
│
├── ucan/                    # User-Controlled Authorization Networks
│   ├── capability.go       # Capability management
│   ├── crypto.go           # UCAN cryptographic operations
│   ├── jwt.go              # JWT-based UCAN tokens
│   ├── verifier.go         # Delegation chain verification
│   └── vault.go            # Vault-specific capabilities
│
├── keys/                    # Key management utilities
│   ├── didkey.go           # DID key format support
│   ├── pubkey.go           # Public key operations
│   └── parsers/            # Multi-chain key parsers
│       ├── btc_parser.go   # Bitcoin key parsing
│       ├── eth_parser.go   # Ethereum key parsing
│       ├── cosmos_parser.go # Cosmos SDK parsing
│       ├── sol_parser.go   # Solana key parsing
│       └── ...             # Other blockchain parsers
│
├── aead/                    # Authenticated encryption
│   └── aes_gcm.go          # AES-GCM AEAD
│
├── daed/                    # Deterministic AEAD
│   └── aes_siv.go          # AES-SIV encryption
│
├── ecies/                   # Elliptic Curve IES
│   ├── encrypt.go          # ECIES encryption
│   └── keys.go             # Key generation
│
├── argon2/                  # Password hashing
│   └── kdf.go              # Argon2 key derivation
│
├── vrf/                     # Verifiable Random Functions
│   └── vrf.go              # Curve25519 VRF
│
├── ecdsa/                   # ECDSA utilities
│   ├── canonical.go        # Canonical signature encoding
│   └── deterministic.go    # RFC 6979 deterministic signing
│
├── subtle/                  # Low-level crypto utilities
│   ├── hkdf.go             # HKDF key derivation
│   ├── random/             # Secure randomness
│   └── x25519.go           # X25519 key exchange
│
└── internal/                # Internal utilities
    ├── ed25519/            # Extended Ed25519 operations
    ├── hash.go             # Hash utilities
    └── point.go            # Point operations
```

## Core Modules

### 1. Multi-Party Computation (`mpc/`)

**Purpose**: Secure distributed key generation and threshold signing without trusted dealers.

#### MPC Enclave Structure

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

#### Key Functions

```go
// Generate new MPC enclave (2-of-2 threshold)
func NewEnclave() (Enclave, error)

// Import enclave from various sources
func ImportEnclave(options ...ImportOption) (Enclave, error)

// Execute distributed signing protocol
func ExecuteSigning(signFuncVal SignFunc, signFuncUser SignFunc) ([]byte, error)

// Execute keyshare refresh protocol
func ExecuteRefresh(refreshFuncVal RefreshFunc, refreshFuncUser RefreshFunc,
                   curve CurveName) (Enclave, error)

// Verify signature with public key
func VerifyWithPubKey(pubKeyCompressed []byte, data []byte, sig []byte) (bool, error)
```

#### Security Features

- **2-of-2 Threshold**: Both parties required for signing
- **No Single Point of Failure**: Neither party can sign alone
- **Proactive Refresh**: Key rotation without changing public key
- **AES-GCM Encryption**: Secure enclave data encryption
- **SHA3-256 Hashing**: Cryptographic hash operations

#### Supported Curves

- `K256` - Secp256k1 (Bitcoin, Ethereum)
- `P256` - NIST P-256
- `ED25519` - Twisted Edwards curve
- `BLS12381` - Pairing-friendly curve

### 2. Elliptic Curves (`core/curves/`)

**Purpose**: Comprehensive elliptic curve implementations with unified interfaces.

#### Supported Curves

**Ed25519**
- Twisted Edwards curve for EdDSA signatures
- High-performance, constant-time operations
- Used in: Cosmos SDK, Solana, many modern systems

**Secp256k1 (K256)**
- Bitcoin and Ethereum standard curve
- ECDSA signature support
- Native field arithmetic implementations

**P-256 (Secp256r1)**
- NIST standard curve
- FIPS 186-4 compliant
- Wide hardware acceleration support

**BLS12-381**
- Pairing-friendly curve for BLS signatures
- Optimal ate pairing support
- Signature aggregation capabilities
- G1, G2, and GT group operations

**BLS12-377**
- Alternative pairing curve
- Used in certain ZK-SNARK constructions

**Pallas/Vesta**
- Pasta curves for recursive ZK proofs
- Cycle of curves for composition

#### Curve Interface

```go
type Curve interface {
    Scalar
    Point
    Name() string
    NewIdentityPoint() Point
    NewGeneratorPoint() Point
    Hash(input []byte) Point
    // ... additional methods
}
```

### 3. Signature Schemes (`signatures/`)

#### BLS Signatures (`signatures/bls/`)

**Features:**
- Signature aggregation (combine multiple signatures)
- Threshold signatures (t-of-n)
- Multi-signatures with proof of possession
- Both G1 and G2 variants (tiny_bls and usual_bls)

**Key Operations:**
```go
// Sign message
func (sk *SecretKey) Sign(msg []byte) *Signature

// Aggregate multiple signatures
func AggregateSignatures(sigs ...*Signature) (*MultiSignature, error)

// Verify aggregated signature
func (sig *Signature) AggregateVerify(pks []*PublicKey, msgs [][]byte) (bool, error)

// Threshold key generation
func ThresholdGenerateKeys(threshold, total int) (*PublicKey, []*SecretKeyShare, error)
```

#### BBS+ Signatures (`signatures/bbs/`)

**Purpose**: Privacy-preserving signatures with selective disclosure

**Features:**
- Blind signatures for credential issuance
- Selective disclosure of attributes
- Zero-knowledge proofs of possession
- Unlinkable presentations

**Use Cases:**
- Verifiable credentials
- Anonymous authentication
- Privacy-preserving identity systems

#### Schnorr Signatures (`signatures/schnorr/`)

**Mina Protocol** (`mina/`):
- Poseidon hash function
- Schnorr signatures for Mina blockchain
- Challenge derivation

**NEM/Symbol** (`nem/`):
- Ed25519-Keccak variant
- NEM blockchain compatibility

### 4. Secret Sharing (`sharing/`)

#### Shamir's Secret Sharing

```go
type Shamir struct {
    Threshold int
    Limit     int
    Curve     Curve
}

// Split secret into shares
func (s *Shamir) Split(secret []byte) ([]*ShamirShare, error)

// Reconstruct secret from shares
func (s *Shamir) Combine(shares []*ShamirShare) ([]byte, error)
```

#### Feldman Verifiable Secret Sharing

**Added Security**: Public commitments for share verification

```go
type FeldmanVerifier struct {
    Commitments []curves.Point
}

// Verify share validity
func (v *FeldmanVerifier) Verify(share *ShamirShare) error
```

#### Pedersen Verifiable Secret Sharing

**Enhanced Privacy**: Computationally binding commitments

```go
// Split with verifiable commitments
func (p *Pedersen) Split(secret []byte) (*PedersenResult, error)
```

### 5. Threshold Cryptography

#### TECDSA (`tecdsa/dklsv1/`)

**Protocol**: Two-party threshold ECDSA (DKLS v1)

**Components:**
- **DKG**: Distributed key generation without trusted dealer
- **Signing**: Threshold signature generation
- **Refresh**: Proactive keyshare rotation
- **Dealer**: Optional trusted dealer mode

**Key Features:**
- No trusted third party required
- Active security against malicious adversaries
- Compatible with standard ECDSA verification

#### TED25519 (`ted25519/`)

**FROST Protocol** (`frost/`):
- Flexible Round-Optimized Schnorr Threshold signatures
- Efficient threshold Ed25519 signatures
- Three-round signing protocol

**Core Operations** (`ted25519/`):
```go
// Threshold key generation
func KeyGen(threshold, total int) ([]*SecretKeyShare, *PublicKey, error)

// Partial signature generation
func ThresholdSign(expandedSecretKeyShare []byte, publicKey []byte,
                   rShare []byte, R []byte, message []byte) []byte

// Signature aggregation
func AggregateSignatures(partialSigs [][]byte, R []byte) ([]byte, error)
```

### 6. Distributed Key Generation (`dkg/`)

#### Gennaro DKG (`dkg/gennaro/`)

**Standard DKG protocol** with:
- Four-round protocol
- Pedersen commitments
- Complaint handling
- Byzantine fault tolerance

#### FROST DKG (`dkg/frost/`)

**Optimized for Ed25519**:
- Two-round DKG
- Simplified complaint phase
- Integration with FROST signing

#### 2-Party DKG (`dkg/gennaro2p/`)

**Simplified protocol** for two parties:
- Reduced communication overhead
- Faster execution
- Suitable for client-server architectures

### 7. Zero-Knowledge Proofs

#### Bulletproofs (`bulletproof/`)

**Range Proofs without Trusted Setup**:

```go
// Prove value in range [0, 2^n]
func (p *RangeProver) Prove(v *big.Int, n int) (*RangeProof, error)

// Verify range proof
func (v *RangeVerifier) Verify(proof *RangeProof, commitment Point, n int) (bool, error)

// Batched range proofs (aggregate multiple proofs)
func BatchProve(values []*big.Int, n int) (*RangeProof, error)
```

**Features:**
- Logarithmic proof size: O(log n)
- Inner product arguments
- Batch verification support
- No trusted setup required

**Applications:**
- Confidential transactions
- Private smart contracts
- Privacy-preserving audits

#### Schnorr Proofs (`zkp/schnorr/`)

**Proof of Knowledge**:
- Discrete logarithm proofs
- Commitment proofs
- Non-interactive via Fiat-Shamir

### 8. Advanced Cryptography

#### Cryptographic Accumulators (`accumulator/`)

**RSA Accumulator** for set membership:

```go
// Add element to accumulator
func (acc *Accumulator) Add(element []byte) (*Witness, error)

// Generate membership proof
func (w *Witness) GenerateProof() (*Proof, error)

// Verify membership
func (acc *Accumulator) Verify(element []byte, proof *Proof) bool
```

**Use Cases:**
- Revocation lists
- Anonymous credentials
- Blockchain state commitments

#### Paillier Encryption (`paillier/`)

**Homomorphic Properties**:
- Additive homomorphism: E(m1) * E(m2) = E(m1 + m2)
- Scalar multiplication: E(m)^k = E(k * m)
- Threshold decryption support

**Applications:**
- Private computation
- Secure multi-party computation
- E-voting systems

#### Oblivious Transfer (`ot/`)

**Simplest OT** (`base/simplest/`):
- 1-out-of-2 OT protocol
- Based on Curve25519

**KOS Extension** (`extension/kos/`):
- Extend base OT to many OTs
- Efficient batch operations
- Correlated randomness generation

**Applications:**
- Private set intersection
- Secure two-party computation
- Password-authenticated key exchange

### 9. UCAN (User-Controlled Authorization Networks) (`ucan/`)

**Capability-Based Authorization**:

```go
// Create UCAN token
func CreateUCAN(issuer DID, audience DID, capabilities []Capability) (string, error)

// Attenuate capabilities (reduce permissions)
func AttenuateUCAN(parentToken string, newCapabilities []Capability) (string, error)

// Verify delegation chain
func VerifyDelegationChain(tokenString string, rootDID string) error
```

**Capability Types**:
- DID capabilities (read, write, update)
- DWN capabilities (records, protocols)
- Vault capabilities (sign, decrypt)
- DEX capabilities (swap, provide liquidity)

**Features:**
- JWT-based tokens
- Delegation chains
- Capability attenuation
- Proof-of-possession
- Expiration and not-before timestamps

### 10. Key Management (`keys/`)

#### DID Key Support (`didkey.go`)

```go
// Create DID from public key
func NewDID(publicKey []byte, keyType crypto.KeyType) (*DID, error)

// Derive blockchain address from DID
func (did *DID) Address() (string, error)

// Get raw public key bytes
func (did *DID) Raw() ([]byte, error)
```

#### Multi-Chain Parsers (`parsers/`)

**Supported Blockchains**:
- Bitcoin (BTC) - BIP32/BIP44 derivation
- Ethereum (ETH) - Keccak addresses
- Cosmos SDK - Bech32 encoding
- Solana (SOL) - Ed25519 keys
- Filecoin (FIL) - Secp256k1 keys
- TON - Ed25519 keys

### 11. Encryption Utilities

#### AEAD (`aead/`)

**AES-GCM Authenticated Encryption**:

```go
const (
    KeySize   = 32  // 256-bit key
    NonceSize = 12  // 96-bit nonce
    TagSize   = 16  // 128-bit auth tag
)

// Encrypt with automatic nonce generation
func (c *AESGCMCipher) Encrypt(plaintext, aad []byte) ([]byte, error)

// Decrypt and verify
func (c *AESGCMCipher) Decrypt(ciphertext, aad []byte) ([]byte, error)
```

#### DAED (`daed/`)

**Deterministic AES-SIV**:
- Same plaintext → same ciphertext
- Useful for encrypted indices
- Misuse-resistant

#### ECIES (`ecies/`)

**Elliptic Curve Integrated Encryption Scheme**:

```go
// Generate ECIES keypair
func GenerateKey(curve Curve) (*PrivateKey, error)

// Encrypt message to public key
func Encrypt(recipientPubKey *PublicKey, message []byte) ([]byte, error)

// Decrypt with private key
func (sk *PrivateKey) Decrypt(ciphertext []byte) ([]byte, error)
```

### 12. Utility Modules

#### VRF (`vrf/`)

**Verifiable Random Function (Curve25519)**:

```go
// Generate VRF output and proof
func (sk *PrivateKey) Prove(message []byte) (vrf []byte, proof []byte)

// Verify VRF proof
func (pk *PublicKey) Verify(message, vrf, proof []byte) bool
```

**Applications:**
- Leader election
- Lottery systems
- Randomness beacons
- Sortition algorithms

#### Argon2 (`argon2/`)

**Password-Based Key Derivation**:

```go
// Derive key from password
func DeriveKey(password, salt []byte, keyLen uint32) []byte
```

**Parameters**:
- Time cost: 1 iteration (configurable)
- Memory cost: 64 MB (configurable)
- Parallelism: 4 threads (configurable)

#### ECDSA Utilities (`ecdsa/`)

**Canonical Encoding**:
- BIP 66 / RFC 6979 compliance
- Deterministic signature generation
- Low-S normalization

## Integration Patterns

### Usage in Sonr Blockchain

The crypto library is heavily integrated throughout the Sonr ecosystem:

#### DID Module
```go
import "github.com/sonr-io/crypto/keys"
import "github.com/sonr-io/crypto/mpc"

// DID creation from MPC enclave
enclave, _ := mpc.NewEnclave()
pubKey := enclave.GetPubPoint()
did := keys.NewDID(pubKey.Bytes(), crypto.Secp256k1)
```

#### DWN Module
```go
import "github.com/sonr-io/crypto/mpc"
import "github.com/sonr-io/crypto/aead"

// Vault operations
enclave := keeper.LoadEnclave(ctx, vaultID)
signature, _ := enclave.Sign(message)

// Encrypted data storage
cipher := aead.NewAESGCMCipher(key)
encrypted, _ := cipher.Encrypt(data, nil)
```

#### Service Module
```go
import "github.com/sonr-io/crypto/ucan"

// UCAN capability verification
verifier := ucan.NewVerifier(didResolver)
err := verifier.VerifyDelegationChain(ctx, tokenString)
```

### Motor Worker Integration

The WASM worker uses the crypto library extensively:

```go
import (
    "github.com/sonr-io/crypto/mpc"
    "github.com/sonr-io/crypto/core/curves"
)

//go:wasmexport sign
func sign() int32 {
    // Load enclave from WASM memory
    enclave := loadEnclave()

    // Sign message
    signature, _ := enclave.Sign(message)

    return writeOutput(signature)
}
```

## Security Considerations

### Threat Model

The library is designed to protect against:

**Key Compromise**:
- MPC threshold schemes prevent single points of failure
- Proactive refresh rotates keyshares

**Insider Threats**:
- Multi-party protocols require cooperation
- No single party can perform operations alone

**Network Attacks**:
- Protocol messages are cryptographically protected
- Authentication prevents man-in-the-middle attacks

**Side-Channel Attacks**:
- Constant-time implementations where critical
- Secure memory handling
- Zeroization of sensitive data

### Best Practices

1. **Key Management**:
   - Use hardware security modules when available
   - Implement secure key backup and recovery
   - Regular keyshare rotation via refresh protocols

2. **MPC Operations**:
   - Secure communication channels (TLS)
   - Proper authentication of parties
   - Audit logging of all operations

3. **Random Number Generation**:
   - Use `crypto/rand` for all random values
   - Never reuse nonces in AEAD
   - Verify randomness quality in production

4. **Error Handling**:
   - Don't leak sensitive information in errors
   - Validate all inputs
   - Use constant-time comparisons for secrets

## Testing

The library includes comprehensive tests:

```bash
# Run all tests
go test ./...

# Run with race detection
go test -race ./...

# Generate coverage report
go test -cover ./...

# Run specific module tests
go test ./mpc/...
go test ./signatures/bls/...
go test ./bulletproof/...

# Benchmark performance
go test -bench=. ./core/curves/...
```

### Test Coverage

- **MPC**: Enclave operations, protocol execution, refresh
- **Signatures**: BLS aggregation, BBS+ proofs, Schnorr
- **Secret Sharing**: Shamir, Feldman, Pedersen
- **Threshold Crypto**: TECDSA, TED25519, DKG protocols
- **ZK Proofs**: Bulletproofs range proofs, Schnorr proofs
- **Encryption**: AEAD, ECIES, Paillier
- **Curves**: All curve operations, point arithmetic

## Dependencies

### External Libraries

```go
require (
    github.com/btcsuite/btcd/btcec/v2 v2.3.2      // Bitcoin crypto
    github.com/consensys/gnark-crypto v0.19.0     // BLS12-377/381
    golang.org/x/crypto v0.42.0                   // Standard crypto
    github.com/golang-jwt/jwt/v5 v5.3.0           // JWT tokens
)
```

### Internal Dependencies

The crypto library is **self-contained** and has no dependencies on other Sonr modules, making it suitable for independent use.

## Performance Characteristics

### Benchmarks (on AMD64, 2.5 GHz)

**Elliptic Curve Operations**:
- K256 scalar multiplication: ~50 µs
- Ed25519 signing: ~25 µs
- BLS12-381 pairing: ~1.2 ms

**MPC Operations**:
- DKG (2-party): ~15 ms
- Threshold signing: ~10 ms
- Key refresh: ~12 ms

**Signature Schemes**:
- BLS aggregation (100 sigs): ~150 ms
- BBS+ proof generation: ~80 ms
- Schnorr signing: ~30 µs

**Zero-Knowledge Proofs**:
- Bulletproof (64-bit range): ~40 ms
- Verification: ~25 ms

## Migration Checklist

When using the crypto library in a new project:

- [ ] Add dependency: `go get github.com/sonr-io/crypto@v1.0.1`
- [ ] Import required modules
- [ ] Initialize curve instances as needed
- [ ] Set up secure random number generation
- [ ] Implement proper error handling
- [ ] Add comprehensive tests
- [ ] Review security best practices
- [ ] Benchmark critical operations
- [ ] Set up monitoring/logging
- [ ] Document cryptographic assumptions

## Version Compatibility

**Go Version**: 1.24.4+

**Cosmos SDK**: Compatible with v0.50.x (if using Cosmos integration)

**Semantic Versioning**: The library follows semver
- Major: Breaking API changes
- Minor: New features, backwards compatible
- Patch: Bug fixes

## Related Documentation

- [Cosmos SDK Cryptography](https://docs.cosmos.network/main/learn/advanced/crypto)
- [BLS Signatures Spec](https://datatracker.ietf.org/doc/html/draft-irtf-cfrg-bls-signature)
- [FROST Paper](https://eprint.iacr.org/2020/852)
- [Bulletproofs Paper](https://eprint.iacr.org/2017/1066)
- [UCAN Spec](https://github.com/ucan-wg/spec)
- [W3C DID Core](https://www.w3.org/TR/did-core/)

## Support

**Repository**: https://github.com/sonr-io/crypto
**Issues**: https://github.com/sonr-io/crypto/issues
**License**: Apache 2.0
