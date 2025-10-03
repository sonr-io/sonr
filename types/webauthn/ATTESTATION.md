# WebAuthn Attestation Formats Documentation

## Overview

WebAuthn attestation is a mechanism that allows authenticators to provide cryptographic proof about their properties during registration. This document describes the different attestation formats supported by this implementation and their validation requirements.

## Supported Attestation Formats

### 1. Packed Attestation (`packed`)

The packed attestation format is a WebAuthn-specific format that supports both self-attestation and full attestation modes.

#### Characteristics
- **Format identifier**: `packed`
- **Specification**: WebAuthn Level 2, Section 8.2
- **Use cases**: General-purpose attestation format, commonly used by platform authenticators

#### Validation Algorithm

##### Full Attestation (with x5c certificate chain)
1. Verify signature over concatenation of authenticatorData and clientDataHash
2. Validate the certificate chain
3. Verify certificate requirements per §8.2.1:
   - Version MUST be 3
   - Subject-C: Valid ISO 3166 country code
   - Subject-O: Legal name of authenticator vendor
   - Subject-OU: Must be "Authenticator Attestation"
   - Subject-CN: Vendor's chosen string
   - AAGUID extension validation if present
   - Basic Constraints: CA flag must be false

##### Self Attestation (without x5c)
1. Parse credential public key from authenticator data
2. Verify signature using the credential public key itself
3. Algorithm must match between signature and key

### 2. TPM Attestation (`tpm`)

TPM attestation uses a Trusted Platform Module to provide hardware-backed attestation.

#### Characteristics
- **Format identifier**: `tpm`
- **Specification**: WebAuthn Level 2, Section 8.3
- **Use cases**: Enterprise environments requiring hardware security

#### Validation Algorithm
1. Verify TPM attestation statement structure
2. Decode and validate pubArea (TPM public area)
3. Decode and validate certInfo (TPM attestation data):
   - Verify magic value is TPM_GENERATED_VALUE
   - Verify type is TPM_ST_ATTEST_CERTIFY
   - Verify extraData matches hash of attToBeSigned
   - Verify attested name matches pubArea
4. Validate AIK certificate chain and extensions
5. Verify signature using AIK certificate

### 3. Apple Anonymous Attestation (`apple`)

Apple's proprietary attestation format for iOS/macOS devices using the Secure Enclave.

#### Characteristics
- **Format identifier**: `apple`
- **Specification**: Apple-specific implementation
- **Use cases**: Apple devices with Touch ID/Face ID

#### Validation Algorithm
1. Extract and validate certificate chain
2. Decode Apple-specific attestation extension
3. Verify nonce matches SHA256(authenticatorData || clientDataHash)
4. Verify credential public key matches certificate public key
5. Return anonymization CA attestation type

### 4. Android Key Attestation (`android-key`)

Android's hardware-backed key attestation using the Android Keystore.

#### Characteristics
- **Format identifier**: `android-key`
- **Specification**: WebAuthn Level 2, Section 8.4
- **Use cases**: Android devices with hardware-backed keystore

#### Validation Algorithm
1. Validate certificate chain structure
2. Extract and verify Android Key attestation extension
3. Verify challenge matches clientDataHash
4. Validate key properties and security level
5. Verify certificate chain to trusted root

### 5. Android SafetyNet Attestation (`android-safetynet`)

Software-based attestation using Google's SafetyNet API.

#### Characteristics
- **Format identifier**: `android-safetynet`
- **Specification**: WebAuthn Level 2, Section 8.5
- **Use cases**: Android devices without hardware attestation support
- **Deprecated**: Being replaced by Play Integrity API

#### Validation Algorithm
1. Parse and verify JWT response from SafetyNet
2. Verify signature against Google's public keys
3. Check nonce matches hash of authenticatorData || clientDataHash
4. Validate CTS profile match and basic integrity

### 6. FIDO U2F Attestation (`fido-u2f`)

Legacy format for backward compatibility with FIDO U2F authenticators.

#### Characteristics
- **Format identifier**: `fido-u2f`
- **Specification**: FIDO U2F to WebAuthn migration
- **Use cases**: Legacy U2F security keys

#### Validation Algorithm
1. Verify signature over registration data
2. Validate attestation certificate
3. Check certificate OID for FIDO compliance
4. Verify EC P-256 key parameters

## Attestation Types

### Basic Attestation
- Authenticator provides its attestation key pair
- Uniquely identifies the authenticator model

### Self Attestation
- Credential private key signs its own attestation
- No authenticator identification possible
- Used when privacy is prioritized

### Attestation CA (Privacy CA)
- Uses anonymization CA certificates
- Provides model attestation without unique identification
- Balance between privacy and attestation

### None Attestation
- No attestation provided
- Relying party accepts authenticator without verification
- Suitable for consumer scenarios

## Security Considerations

1. **Certificate Validation**
   - Always validate complete certificate chains
   - Check certificate validity periods
   - Verify against known root certificates

2. **Signature Verification**
   - Use constant-time comparison for signatures
   - Validate algorithm consistency
   - Check for signature malleability

3. **Extension Validation**
   - Strictly parse and validate all extensions
   - Reject unknown critical extensions
   - Validate extension data formats

4. **Privacy Considerations**
   - Consider attestation type based on use case
   - Balance security needs with user privacy
   - Implement attestation conveyance preferences

## Error Handling

All attestation validation errors are wrapped with contextual information:
- `ErrAttestationFormat`: Format-specific parsing errors
- `ErrInvalidAttestation`: Validation failures
- `ErrAttestationCertificate`: Certificate validation errors

## Testing

Each attestation format includes comprehensive test coverage:
- Valid attestation verification
- Invalid signature detection
- Certificate chain validation
- Extension parsing
- Edge cases and malformed data

## References

- [WebAuthn Level 2 Specification](https://www.w3.org/TR/webauthn-2/)
- [FIDO Alliance Metadata Service](https://fidoalliance.org/metadata/)
- [TPM 2.0 Specification](https://trustedcomputinggroup.org/resource/tpm-library-specification/)
- [Android Key Attestation](https://developer.android.com/training/articles/security-key-attestation)
- [Apple Platform Authenticator](https://developer.apple.com/documentation/authenticationservices)