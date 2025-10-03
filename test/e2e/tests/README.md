# WebAuthn Integration Tests

## Overview

This test suite provides comprehensive end-to-end testing for the WebAuthn implementation in the Sonr blockchain's DID module. The tests validate the complete WebAuthn registration, authentication, and key management workflows.

## Test Coverage

The test suite covers the following scenarios:

1. **Attestation Parsing**
   - Validates parsing of CBOR attestation objects
   - Checks public key extraction
   - Verifies algorithm and authenticator data detection

2. **Registration Flow**
   - Complete WebAuthn credential registration
   - Challenge verification
   - Origin validation
   - DID document creation with WebAuthn credentials
   - Credential uniqueness enforcement

3. **Signature Verification**
   - WebAuthn assertion verification
   - User presence and verification flag checking
   - Multi-algorithm signature support (ES256, RS256, EdDSA)
   - Counter increment validation

4. **Security Scenarios**
   - Challenge replay attack prevention
   - Invalid origin rejection
   - Oversized credential handling
   - Credential ID reuse prevention

## Test Methodology

- Uses Cosmos SDK testing framework
- Employs table-driven tests for multiple scenarios
- Mocks cryptographic keys and challenge responses
- Validates both positive and negative test cases

## Running Tests

```bash
go test github.com/sonr-io/sonr/test/e2e/tests -v
```

## Dependencies

- Cosmos SDK v0.50.14
- Internal WebAuthn libraries
- testify assertion library