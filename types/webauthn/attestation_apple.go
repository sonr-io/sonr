package webauthn

import (
	"bytes"
	"crypto/sha256"
	"crypto/x509"
	"encoding/asn1"
	"fmt"

	"github.com/sonr-io/sonr/types/webauthn/metadata"
	"github.com/sonr-io/sonr/types/webauthn/webauthncose"
)

func init() {
	RegisterAttestationFormat(AttestationFormatApple, verifyAppleFormat)
}

// The apple attestation statement looks like:
// $$attStmtType //= (
//
//	fmt: "apple",
//	attStmt: appleStmtFormat
//
// )
//
//	appleStmtFormat = {
//			x5c: [ credCert: bytes, * (caCert: bytes) ]
//	  }
//
// Specification: §8.8. Apple Anonymous Attestation Statement Format (https://www.w3.org/TR/webauthn/#sctn-apple-anonymous-attestation)
func verifyAppleFormat(
	att AttestationObject,
	clientDataHash []byte,
	_ metadata.Provider,
) (string, []any, error) {
	// Step 1. Verify that attStmt is valid CBOR conforming to the syntax defined
	// above and perform CBOR decoding on it to extract the contained fields.
	// If x5c is not present, return an error.
	x5c, x509present := att.AttStatement[stmtX5C].([]any)
	if !x509present {
		// Handle Basic Attestation steps for the x509 Certificate
		return "", nil, ErrAttestationFormat.WithDetails("Error retrieving x5c value")
	}

	credCertBytes, valid := x5c[0].([]byte)
	if !valid {
		return "", nil, ErrAttestation.WithDetails("Error getting certificate from x5c cert chain")
	}

	credCert, err := x509.ParseCertificate(credCertBytes)
	if err != nil {
		return "", nil, ErrAttestationFormat.WithDetails(fmt.Sprintf("Error parsing certificate from ASN.1 data: %+v", err)).
			WithError(err)
	}

	// Step 2. Concatenate authenticatorData and clientDataHash to form nonceToHash.
	nonceToHash := append(att.RawAuthData, clientDataHash...)

	// Step 3. Perform SHA-256 hash of nonceToHash to produce nonce.
	nonce := sha256.Sum256(nonceToHash)

	// Step 4. Verify that nonce equals the value of the extension with OID 1.2.840.113635.100.8.2 in credCert.
	var attExtBytes []byte

	for _, ext := range credCert.Extensions {
		if ext.Id.Equal([]int{1, 2, 840, 113635, 100, 8, 2}) {
			attExtBytes = ext.Value
		}
	}

	if len(attExtBytes) == 0 {
		return "", nil, ErrAttestationFormat.WithDetails(
			"Attestation certificate extensions missing 1.2.840.113635.100.8.2",
		)
	}

	decoded := AppleAnonymousAttestation{}

	if _, err = asn1.Unmarshal(attExtBytes, &decoded); err != nil {
		return "", nil, ErrAttestationFormat.WithDetails("Unable to parse apple attestation certificate extensions").
			WithError(err)
	}

	if !bytes.Equal(decoded.Nonce, nonce[:]) {
		return "", nil, ErrInvalidAttestation.WithDetails(
			"Attestation certificate does not contain expected nonce",
		)
	}

	// Step 5. Verify that the credential public key equals the Subject Public Key of credCert.
	pubKey, err := webauthncose.ParsePublicKey(att.AuthData.AttData.CredentialPublicKey)
	if err != nil {
		return "", nil, ErrInvalidAttestation.WithDetails(fmt.Sprintf("Error parsing public key: %+v\n", err)).
			WithError(err)
	}

	// Use the centralized verification function in webauthncose
	if err := webauthncose.VerifyEC2PublicKeyMatch(pubKey, credCert.PublicKey); err != nil {
		return "", nil, ErrInvalidAttestation.WithDetails(
			"Certificate public key does not match public key in authData",
		).WithError(err)
	}

	// Step 6. If successful, return implementation-specific values representing attestation type Anonymization CA and attestation trust path x5c.
	return string(metadata.AnonCA), x5c, nil
}

// AppleAnonymousAttestation represents the attestation format for Apple, who have not yet published a schema for the
// extension (as of JULY 2021.)
type AppleAnonymousAttestation struct {
	Nonce []byte `asn1:"tag:1,explicit"`
}
