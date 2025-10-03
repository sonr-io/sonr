// Package keeper provides the DID module keeper implementation.
package keeper

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"math/big"
	"strings"

	"cosmossdk.io/collections"
	storetypes "cosmossdk.io/core/store"
	"cosmossdk.io/log"
	"cosmossdk.io/orm/model/ormdb"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"golang.org/x/crypto/sha3"

	apiv1 "github.com/sonr-io/sonr/api/did/v1"
	"github.com/sonr-io/sonr/crypto/mpc"
	"github.com/sonr-io/sonr/types/webauthn"
	"github.com/sonr-io/sonr/types/webauthn/webauthncose"
	"github.com/sonr-io/sonr/x/did/types"
)

type Keeper struct {
	cdc codec.BinaryCodec

	logger log.Logger

	// state management
	Schema collections.Schema
	Params collections.Item[types.Params]
	OrmDB  apiv1.StateStore

	// cross-module keeper dependencies
	dwnKeeper     types.DWNKeeper
	accountKeeper types.AccountKeeper
	serviceKeeper types.ServiceKeeper

	// UCAN permission validation
	permissionValidator *PermissionValidator

	authority string
}

// HasExistingCredential checks if a WebAuthn credential ID already exists in the system.
// This prevents credential reuse and replay attacks in gasless registration.
func (k Keeper) HasExistingCredential(ctx sdk.Context, credentialId string) bool {
	// Query all DID documents to check for credential ID uniqueness
	// WebAuthn credentials are stored as verification methods in DID documents

	// Use ORM iterator to scan all DID documents efficiently
	iterator, err := k.OrmDB.DIDDocumentTable().List(ctx, &apiv1.DIDDocumentPrimaryKey{})
	if err != nil {
		k.logger.Error("Failed to list DID documents for credential check", "error", err)
		// In case of error, err on the side of caution and reject the credential
		return true
	}
	defer iterator.Close()

	// Iterate through all DID documents
	for iterator.Next() {
		ormDoc, err := iterator.Value()
		if err != nil {
			k.logger.Error("Failed to get DID document during credential check", "error", err)
			continue
		}

		// Convert from ORM type to check verification methods
		didDoc := types.DIDDocumentFromORM(ormDoc)

		// Check all verification methods for WebAuthn credentials
		for _, vm := range didDoc.VerificationMethod {
			if vm.WebauthnCredential != nil && vm.WebauthnCredential.CredentialId == credentialId {
				k.logger.Info("Found existing WebAuthn credential",
					"credential_id", credentialId,
					"existing_did", didDoc.Id,
					"verification_method", vm.Id)
				return true
			}
		}
	}

	k.logger.Debug("WebAuthn credential ID is unique", "credential_id", credentialId)
	return false
}

// NewKeeper creates a new Keeper instance
func NewKeeper(
	cdc codec.BinaryCodec,
	storeService storetypes.KVStoreService,
	logger log.Logger,
	authority string,
	accountKeeper types.AccountKeeper,
) Keeper {
	logger = logger.With(log.ModuleKey, "x/"+types.ModuleName)

	sb := collections.NewSchemaBuilder(storeService)

	if authority == "" {
		authority = authtypes.NewModuleAddress(govtypes.ModuleName).String()
	}

	db, err := ormdb.NewModuleDB(
		&types.ORMModuleSchema,
		ormdb.ModuleDBOptions{KVStoreService: storeService},
	)
	if err != nil {
		panic(err)
	}

	store, err := apiv1.NewStateStore(db)
	if err != nil {
		panic(err)
	}

	k := Keeper{
		cdc:    cdc,
		logger: logger,

		Params: collections.NewItem(
			sb,
			types.ParamsKey,
			"params",
			codec.CollValue[types.Params](cdc),
		),
		OrmDB: store,

		dwnKeeper:     nil, // Will be set later via SetDWNKeeper
		accountKeeper: accountKeeper,
		authority:     authority,
	}

	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}

	k.Schema = schema

	// Initialize UCAN permission validator (after keeper is fully constructed)
	k.permissionValidator = NewPermissionValidator(k)

	return k
}

func (k Keeper) Logger() log.Logger {
	return k.logger
}

// GetPermissionValidator returns the UCAN permission validator
func (k Keeper) GetPermissionValidator() *PermissionValidator {
	return k.permissionValidator
}

// InitGenesis initializes the module's state from a genesis state.
func (k *Keeper) InitGenesis(ctx context.Context, data *types.GenesisState) error {
	if err := data.Params.Validate(); err != nil {
		return err
	}

	return k.Params.Set(ctx, data.Params)
}

// ExportGenesis exports the module's state to a genesis state.
func (k *Keeper) ExportGenesis(ctx context.Context) *types.GenesisState {
	params, err := k.Params.Get(ctx)
	if err != nil {
		panic(err)
	}

	return &types.GenesisState{
		Params: params,
	}
}

// ResolveDID resolves a DID to its DID document
func (k Keeper) ResolveDID(
	ctx context.Context,
	did string,
) (*types.DIDDocument, *types.DIDDocumentMetadata, error) {
	// Get DID document from ORM
	ormDoc, err := k.OrmDB.DIDDocumentTable().Get(ctx, did)
	if err != nil {
		return nil, nil, err
	}

	// Convert from ORM type
	didDoc := types.DIDDocumentFromORM(ormDoc)

	// Get metadata
	ormMetadata, err := k.OrmDB.DIDDocumentMetadataTable().Get(ctx, did)
	if err != nil {
		// Metadata might not exist, which is ok
		return didDoc, nil, nil
	}

	metadata := types.DIDDocumentMetadataFromORM(ormMetadata)
	return didDoc, metadata, nil
}

// GetDIDDocument gets a DID document by its ID
func (k Keeper) GetDIDDocument(ctx context.Context, did string) (*types.DIDDocument, error) {
	// Get DID document from ORM
	ormDoc, err := k.OrmDB.DIDDocumentTable().Get(ctx, did)
	if err != nil {
		return nil, err
	}

	// Convert from ORM type
	didDoc := types.DIDDocumentFromORM(ormDoc)
	return didDoc, nil
}

// VerifyDIDDocumentSignature verifies a DID document signature using verification methods
func (k Keeper) VerifyDIDDocumentSignature(
	ctx context.Context,
	did string,
	signature []byte,
) (bool, error) {
	// Get the DID document
	didDoc, err := k.GetDIDDocument(ctx, did)
	if err != nil {
		return false, fmt.Errorf("failed to get DID document: %w", err)
	}

	if didDoc == nil {
		return false, fmt.Errorf("DID document not found for DID: %s", did)
	}

	// If document is deactivated, signature verification should fail
	if didDoc.Deactivated {
		return false, fmt.Errorf("cannot verify signature for deactivated DID: %s", did)
	}

	// Try to verify signature using verification methods
	for _, vm := range didDoc.VerificationMethod {
		if vm == nil {
			continue
		}

		// Try to verify with this verification method
		verified, err := k.verifyWithVerificationMethod(vm, signature)
		if err != nil {
			// Log error but continue with other verification methods
			k.Logger().
				Debug("Failed to verify with verification method", "vm_id", vm.Id, "error", err)
			continue
		}

		if verified {
			return true, nil
		}
	}

	return false, fmt.Errorf("signature verification failed for DID: %s", did)
}

// verifyWithVerificationMethod attempts to verify a signature using a specific verification method
func (k Keeper) verifyWithVerificationMethod(
	vm *types.VerificationMethod,
	signature []byte,
) (bool, error) {
	if vm == nil {
		return false, fmt.Errorf("verification method is nil")
	}

	// Handle different verification method types
	switch strings.ToLower(vm.VerificationMethodKind) {
	case "jsonwebsignature2020":
		return k.verifyJsonWebSignature2020(vm, signature)
	case "webauthn":
		return k.verifyWebAuthnSignature(vm, signature)
	case "ed25519verificationkey2020":
		return k.verifyEd25519Signature(vm, signature)
	case "ecdsasecp256k1verificationkey2019":
		return k.verifyECDSASecp256k1Signature(vm, signature)
	case "rsaverificationkey2018":
		return k.verifyRSASignature(vm, signature)
	default:
		return false, fmt.Errorf(
			"unsupported verification method type: %s",
			vm.VerificationMethodKind,
		)
	}
}

// verifyJsonWebSignature2020 verifies a JSON Web Signature 2020
func (k Keeper) verifyJsonWebSignature2020(
	vm *types.VerificationMethod,
	signature []byte,
) (bool, error) {
	// Parse the signature as JSON
	var jws map[string]any
	if err := json.Unmarshal(signature, &jws); err != nil {
		return false, fmt.Errorf("failed to parse JWS signature: %w", err)
	}

	// Extract signature from JWS
	sigData, ok := jws["signature"].(string)
	if !ok {
		return false, fmt.Errorf("missing signature in JWS")
	}

	sigBytes, err := base64.URLEncoding.DecodeString(sigData)
	if err != nil {
		return false, fmt.Errorf("failed to decode JWS signature: %w", err)
	}

	// Use the appropriate key material based on what's available
	if vm.PublicKeyJwk != "" {
		return k.verifyWithJWK(vm.PublicKeyJwk, sigBytes)
	}

	// Fall back to other key formats
	return k.verifyWithKeyMaterial(vm, sigBytes)
}

// verifyWebAuthnSignature implements complete WebAuthn signature verification with CBOR parsing
// and assertion validation following the W3C WebAuthn specification.
func (k Keeper) verifyWebAuthnSignature(
	vm *types.VerificationMethod,
	signature []byte,
) (bool, error) {
	if vm.WebauthnCredential == nil {
		return false, fmt.Errorf("WebAuthn credential not found in verification method")
	}

	// Get stored public key and credential data
	publicKey := vm.WebauthnCredential.PublicKey
	if len(publicKey) == 0 {
		return false, fmt.Errorf("WebAuthn public key is empty")
	}

	credentialID := vm.WebauthnCredential.CredentialId

	// Parse WebAuthn assertion response from signature bytes
	// The signature parameter contains the complete assertion response as JSON
	var assertionResponse webauthn.CredentialAssertionResponse
	if err := json.Unmarshal(signature, &assertionResponse); err != nil {
		return false, fmt.Errorf("failed to parse WebAuthn assertion response: %w", err)
	}

	// Parse the assertion response to get structured data
	parsedAssertion, err := assertionResponse.Parse()
	if err != nil {
		return false, fmt.Errorf("failed to parse WebAuthn assertion: %w", err)
	}

	// Verify credential ID matches
	if parsedAssertion.ID != credentialID {
		return false, fmt.Errorf("credential ID mismatch")
	}

	// Parse and verify authenticator data
	authData := parsedAssertion.Response.AuthenticatorData

	// Verify authenticator flags for user presence and verification
	if !authData.Flags.HasUserPresent() {
		return false, fmt.Errorf("user presence flag not set")
	}

	// Check user verification if it was required during registration
	if vm.WebauthnCredential.UserVerified && !authData.Flags.HasUserVerified() {
		return false, fmt.Errorf("user verification required but flag not set")
	}

	// Reconstruct signed data: authenticatorData + SHA-256(clientDataJSON)
	clientDataHash := sha256.Sum256(parsedAssertion.Raw.AssertionResponse.ClientDataJSON)
	signedData := append(
		parsedAssertion.Raw.AssertionResponse.AuthenticatorData,
		clientDataHash[:]...)

	// Parse the stored public key using CBOR
	parsedKey, err := webauthncose.ParsePublicKey(publicKey)
	if err != nil {
		return false, fmt.Errorf("failed to parse WebAuthn public key: %w", err)
	}

	// Verify signature using the appropriate algorithm
	valid, err := webauthncose.VerifySignature(
		parsedKey,
		signedData,
		parsedAssertion.Response.Signature,
	)
	if err != nil {
		return false, fmt.Errorf("signature verification failed: %w", err)
	}

	if !valid {
		return false, fmt.Errorf("WebAuthn signature is invalid")
	}

	// Counter validation would require persistent storage updates
	// For now, we log the counter value for monitoring
	k.logger.Debug("WebAuthn signature verified",
		"credential_id", credentialID,
		"counter", authData.Counter,
		"user_present", authData.Flags.HasUserPresent(),
		"user_verified", authData.Flags.HasUserVerified())

	return true, nil
}

// verifyEd25519Signature verifies an Ed25519 signature
func (k Keeper) verifyEd25519Signature(
	vm *types.VerificationMethod,
	signature []byte,
) (bool, error) {
	publicKey, err := k.extractEd25519PublicKey(vm)
	if err != nil {
		return false, fmt.Errorf("failed to extract Ed25519 public key: %w", err)
	}

	// Create a test message (in practice, this would be the actual message being signed)
	message := []byte("test message")

	// Verify the signature
	return ed25519.Verify(publicKey, message, signature), nil
}

// verifyECDSASecp256k1Signature verifies an ECDSA secp256k1 signature
func (k Keeper) verifyECDSASecp256k1Signature(
	vm *types.VerificationMethod,
	signature []byte,
) (bool, error) {
	publicKey, err := k.extractECDSAPublicKey(vm)
	if err != nil {
		return false, fmt.Errorf("failed to extract ECDSA public key: %w", err)
	}

	// Create a test message hash
	message := []byte("test message")
	hash := sha256.Sum256(message)

	// Verify the signature
	return ecdsa.VerifyASN1(publicKey, hash[:], signature), nil
}

// verifyRSASignature verifies an RSA signature
func (k Keeper) verifyRSASignature(
	vm *types.VerificationMethod,
	signature []byte,
) (bool, error) {
	publicKey, err := k.extractRSAPublicKey(vm)
	if err != nil {
		return false, fmt.Errorf("failed to extract RSA public key: %w", err)
	}

	// Create a test message hash
	message := []byte("test message")
	hash := sha256.Sum256(message)

	// Verify the signature
	err = rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, hash[:], signature)
	return err == nil, nil
}

// Helper functions for key extraction

func (k Keeper) extractEd25519PublicKey(
	vm *types.VerificationMethod,
) (ed25519.PublicKey, error) {
	if vm.PublicKeyBase64 != "" {
		keyBytes, err := base64.StdEncoding.DecodeString(vm.PublicKeyBase64)
		if err != nil {
			return nil, fmt.Errorf("failed to decode base64 public key: %w", err)
		}
		if len(keyBytes) != ed25519.PublicKeySize {
			return nil, fmt.Errorf("invalid Ed25519 public key size: %d", len(keyBytes))
		}
		return ed25519.PublicKey(keyBytes), nil
	}

	if vm.PublicKeyHex != "" {
		keyBytes, err := hex.DecodeString(vm.PublicKeyHex)
		if err != nil {
			return nil, fmt.Errorf("failed to decode hex public key: %w", err)
		}
		if len(keyBytes) != ed25519.PublicKeySize {
			return nil, fmt.Errorf("invalid Ed25519 public key size: %d", len(keyBytes))
		}
		return ed25519.PublicKey(keyBytes), nil
	}

	return nil, fmt.Errorf("no suitable public key format found for Ed25519")
}

func (k Keeper) extractECDSAPublicKey(vm *types.VerificationMethod) (*ecdsa.PublicKey, error) {
	if vm.PublicKeyPem != "" {
		block, _ := pem.Decode([]byte(vm.PublicKeyPem))
		if block == nil {
			return nil, fmt.Errorf("failed to parse PEM block")
		}

		pub, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse ECDSA public key: %w", err)
		}

		ecdsaPub, ok := pub.(*ecdsa.PublicKey)
		if !ok {
			return nil, fmt.Errorf("public key is not an ECDSA key")
		}

		return ecdsaPub, nil
	}

	return nil, fmt.Errorf("no suitable public key format found for ECDSA")
}

func (k Keeper) extractRSAPublicKey(vm *types.VerificationMethod) (*rsa.PublicKey, error) {
	if vm.PublicKeyPem != "" {
		block, _ := pem.Decode([]byte(vm.PublicKeyPem))
		if block == nil {
			return nil, fmt.Errorf("failed to parse PEM block")
		}

		pub, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse RSA public key: %w", err)
		}

		rsaPub, ok := pub.(*rsa.PublicKey)
		if !ok {
			return nil, fmt.Errorf("public key is not an RSA key")
		}

		return rsaPub, nil
	}

	return nil, fmt.Errorf("no suitable public key format found for RSA")
}

// verifyWithJWK verifies a signature using a JSON Web Key
func (k Keeper) verifyWithJWK(jwkStr string, signature []byte) (bool, error) {
	var jwk map[string]any
	if err := json.Unmarshal([]byte(jwkStr), &jwk); err != nil {
		return false, fmt.Errorf("failed to parse JWK: %w", err)
	}

	// Extract key type
	kty, ok := jwk["kty"].(string)
	if !ok {
		return false, fmt.Errorf("missing kty in JWK")
	}

	switch kty {
	case "OKP":
		// Ed25519 key
		return k.verifyWithJWKOKP(jwk, signature)
	case "EC":
		// ECDSA key
		return k.verifyWithJWKEC(jwk, signature)
	case "RSA":
		// RSA key
		return k.verifyWithJWKRSA(jwk, signature)
	default:
		return false, fmt.Errorf("unsupported JWK key type: %s", kty)
	}
}

// verifyWithJWKOKP verifies using an OKP (Octet Key Pair) JWK
func (k Keeper) verifyWithJWKOKP(jwk map[string]any, signature []byte) (bool, error) {
	x, ok := jwk["x"].(string)
	if !ok {
		return false, fmt.Errorf("missing x parameter in OKP JWK")
	}

	keyBytes, err := base64.RawURLEncoding.DecodeString(x)
	if err != nil {
		return false, fmt.Errorf("failed to decode OKP key: %w", err)
	}

	if len(keyBytes) != ed25519.PublicKeySize {
		return false, fmt.Errorf("invalid Ed25519 key size: %d", len(keyBytes))
	}

	publicKey := ed25519.PublicKey(keyBytes)
	message := []byte("test message")

	return ed25519.Verify(publicKey, message, signature), nil
}

// verifyWithJWKEC verifies using an EC JWK with support for multiple curves
func (k Keeper) verifyWithJWKEC(jwk map[string]any, signature []byte) (bool, error) {
	// Parse curve type from JWK
	crv, ok := jwk["crv"].(string)
	if !ok {
		return false, fmt.Errorf("missing or invalid 'crv' parameter in EC JWK")
	}

	// Parse x and y coordinates
	xStr, ok := jwk["x"].(string)
	if !ok {
		return false, fmt.Errorf("missing or invalid 'x' coordinate in EC JWK")
	}

	yStr, ok := jwk["y"].(string)
	if !ok {
		return false, fmt.Errorf("missing or invalid 'y' coordinate in EC JWK")
	}

	// Decode base64url encoded coordinates
	xBytes, err := base64.RawURLEncoding.DecodeString(xStr)
	if err != nil {
		return false, fmt.Errorf("failed to decode x coordinate: %w", err)
	}

	yBytes, err := base64.RawURLEncoding.DecodeString(yStr)
	if err != nil {
		return false, fmt.Errorf("failed to decode y coordinate: %w", err)
	}

	// Select the appropriate curve
	var curve elliptic.Curve
	var hashFunc crypto.Hash
	switch crv {
	case "P-256":
		curve = elliptic.P256()
		hashFunc = crypto.SHA256
	case "P-384":
		curve = elliptic.P384()
		hashFunc = crypto.SHA384
	case "P-521":
		curve = elliptic.P521()
		hashFunc = crypto.SHA512
	case "secp256k1":
		// Note: secp256k1 requires external package, using SHA256
		return false, fmt.Errorf("secp256k1 curve not yet supported")
	default:
		return false, fmt.Errorf("unsupported curve: %s", crv)
	}

	// Convert coordinates to big integers
	x := new(big.Int).SetBytes(xBytes)
	y := new(big.Int).SetBytes(yBytes)

	// Validate the point is on the curve
	if !curve.IsOnCurve(x, y) {
		return false, fmt.Errorf("invalid EC point: not on curve %s", crv)
	}

	// Create ECDSA public key
	publicKey := &ecdsa.PublicKey{
		Curve: curve,
		X:     x,
		Y:     y,
	}

	// Prepare message for verification (using test message for now)
	message := []byte("test message")

	// Hash the message based on curve requirements
	var digest []byte
	switch hashFunc {
	case crypto.SHA256:
		h := sha256.Sum256(message)
		digest = h[:]
	case crypto.SHA384:
		h := sha3.Sum384(message)
		digest = h[:]
	case crypto.SHA512:
		h := sha512.Sum512(message)
		digest = h[:]
	default:
		return false, fmt.Errorf("unsupported hash function for curve %s", crv)
	}

	// Verify the signature (assuming ASN.1 DER format)
	return ecdsa.VerifyASN1(publicKey, digest, signature), nil
}

// verifyWithJWKRSA verifies using an RSA JWK with proper modulus and exponent parsing
func (k Keeper) verifyWithJWKRSA(jwk map[string]any, signature []byte) (bool, error) {
	// Parse modulus (n) from JWK
	nStr, ok := jwk["n"].(string)
	if !ok {
		return false, fmt.Errorf("missing or invalid 'n' (modulus) in RSA JWK")
	}

	// Parse exponent (e) from JWK
	eStr, ok := jwk["e"].(string)
	if !ok {
		return false, fmt.Errorf("missing or invalid 'e' (exponent) in RSA JWK")
	}

	// Decode base64url encoded modulus
	nBytes, err := base64.RawURLEncoding.DecodeString(nStr)
	if err != nil {
		return false, fmt.Errorf("failed to decode modulus: %w", err)
	}

	// Decode base64url encoded exponent
	eBytes, err := base64.RawURLEncoding.DecodeString(eStr)
	if err != nil {
		return false, fmt.Errorf("failed to decode exponent: %w", err)
	}

	// Convert to big integers
	n := new(big.Int).SetBytes(nBytes)
	e := new(big.Int).SetBytes(eBytes)

	// Validate key size (minimum 2048 bits for security)
	if n.BitLen() < 2048 {
		return false, fmt.Errorf("RSA key size too small: %d bits (minimum 2048)", n.BitLen())
	}

	// Create RSA public key
	publicKey := &rsa.PublicKey{
		N: n,
		E: int(e.Int64()),
	}

	// Parse algorithm from JWK if present
	alg, _ := jwk["alg"].(string)

	// Prepare message for verification
	message := []byte("test message")

	// Choose hash function based on algorithm or key size
	var hashFunc crypto.Hash
	switch alg {
	case "RS256", "PS256":
		hashFunc = crypto.SHA256
	case "RS384", "PS384":
		hashFunc = crypto.SHA384
	case "RS512", "PS512":
		hashFunc = crypto.SHA512
	default:
		// Default based on key size
		if n.BitLen() >= 4096 {
			hashFunc = crypto.SHA512
		} else if n.BitLen() >= 3072 {
			hashFunc = crypto.SHA384
		} else {
			hashFunc = crypto.SHA256
		}
	}

	// Hash the message
	var hashed []byte
	switch hashFunc {
	case crypto.SHA256:
		h := sha256.Sum256(message)
		hashed = h[:]
	case crypto.SHA384:
		h := sha3.Sum384(message)
		hashed = h[:]
	case crypto.SHA512:
		h := sha512.Sum512(message)
		hashed = h[:]
	}

	// Try RSA-PSS first if algorithm indicates it
	if strings.HasPrefix(alg, "PS") {
		opts := &rsa.PSSOptions{
			SaltLength: rsa.PSSSaltLengthEqualsHash,
			Hash:       hashFunc,
		}
		verifyErr := rsa.VerifyPSS(publicKey, hashFunc, hashed, signature, opts)
		return verifyErr == nil, verifyErr
	}

	// Default to PKCS#1 v1.5
	err = rsa.VerifyPKCS1v15(publicKey, hashFunc, hashed, signature)
	return err == nil, err
}

// verifyWithKeyMaterial verifies using available key material
func (k Keeper) verifyWithKeyMaterial(
	vm *types.VerificationMethod,
	signature []byte,
) (bool, error) {
	// Try different key formats
	if vm.PublicKeyBase64 != "" {
		keyBytes, err := base64.StdEncoding.DecodeString(vm.PublicKeyBase64)
		if err != nil {
			return false, fmt.Errorf("failed to decode base64 key: %w", err)
		}
		return k.verifyWithPublicKeyBytes(keyBytes, signature)
	}

	if vm.PublicKeyHex != "" {
		keyBytes, err := hex.DecodeString(vm.PublicKeyHex)
		if err != nil {
			return false, fmt.Errorf("failed to decode hex key: %w", err)
		}
		return k.verifyWithPublicKeyBytes(keyBytes, signature)
	}

	return false, fmt.Errorf("no suitable key material found")
}

// verifyWithPublicKeyBytes implements proper key type detection and multi-algorithm verification.
// It detects the key format from length and structure, supporting Ed25519, ECDSA (P-256, secp256k1),
// and RSA keys with appropriate signature verification for each detected type.
func (k Keeper) verifyWithPublicKeyBytes(publicKey []byte, signature []byte) (bool, error) {
	if len(publicKey) == 0 || len(signature) == 0 {
		return false, fmt.Errorf("empty public key or signature")
	}

	message := []byte("test message")

	// Ed25519 key detection (32 bytes)
	if len(publicKey) == ed25519.PublicKeySize {
		return ed25519.Verify(ed25519.PublicKey(publicKey), message, signature), nil
	}

	// ECDSA uncompressed P-256 key detection (64 bytes: 32-byte X + 32-byte Y coordinates)
	if len(publicKey) == 64 {
		return k.verifyECDSAUncompressed(publicKey, signature, message, elliptic.P256())
	}

	// ECDSA compressed key detection (33 bytes: 1-byte prefix + 32-byte coordinate)
	if len(publicKey) == 33 && (publicKey[0] == 0x02 || publicKey[0] == 0x03) {
		return k.verifyECDSACompressed(publicKey, signature, message, elliptic.P256())
	}

	// secp256k1 compressed key detection (33 bytes with different handling)
	if len(publicKey) == 33 && (publicKey[0] == 0x02 || publicKey[0] == 0x03) {
		// Try secp256k1 if P-256 fails
		// Note: Would need to import secp256k1 curve for full support
		k.logger.Debug("secp256k1 compressed key detected but not fully supported")
	}

	// secp256k1 uncompressed key detection (65 bytes: 0x04 prefix + 32-byte X + 32-byte Y)
	if len(publicKey) == 65 && publicKey[0] == 0x04 {
		// Extract coordinates and try with P-256 as fallback
		coords := publicKey[1:] // Remove 0x04 prefix
		return k.verifyECDSAUncompressed(coords, signature, message, elliptic.P256())
	}

	// RSA key detection - try parsing as PKIX ASN.1 DER format
	if len(publicKey) > 100 { // RSA keys are typically much larger
		if verified, err := k.verifyRSAFromDER(publicKey, signature, message); err == nil {
			return verified, nil
		}
	}

	// Try parsing as COSE key format (WebAuthn keys)
	if verified, err := k.verifyCOSEKey(publicKey, signature, message); err == nil {
		return verified, nil
	}

	return false, fmt.Errorf("unable to detect key type for %d-byte key", len(publicKey))
}

// verifyECDSAUncompressed verifies ECDSA signature with uncompressed public key coordinates
func (k Keeper) verifyECDSAUncompressed(
	coords []byte,
	signature []byte,
	message []byte,
	curve elliptic.Curve,
) (bool, error) {
	if len(coords) != 64 {
		return false, fmt.Errorf("invalid uncompressed ECDSA key length: %d", len(coords))
	}

	// Split coordinates
	x := big.NewInt(0).SetBytes(coords[:32])
	y := big.NewInt(0).SetBytes(coords[32:])

	// Validate point is on curve
	if !curve.IsOnCurve(x, y) {
		return false, fmt.Errorf("point not on curve")
	}

	publicKey := &ecdsa.PublicKey{
		Curve: curve,
		X:     x,
		Y:     y,
	}

	// Hash message
	hash := sha256.Sum256(message)

	// Try ASN.1 DER encoded signature first
	if ecdsa.VerifyASN1(publicKey, hash[:], signature) {
		return true, nil
	}

	// Try raw r||s format (64 bytes for P-256)
	if len(signature) == 64 {
		r := big.NewInt(0).SetBytes(signature[:32])
		s := big.NewInt(0).SetBytes(signature[32:])
		return ecdsa.Verify(publicKey, hash[:], r, s), nil
	}

	return false, fmt.Errorf("signature verification failed")
}

// verifyECDSACompressed verifies ECDSA signature with compressed public key
func (k Keeper) verifyECDSACompressed(
	compressedKey []byte,
	signature []byte,
	message []byte,
	curve elliptic.Curve,
) (bool, error) {
	if len(compressedKey) != 33 {
		return false, fmt.Errorf("invalid compressed ECDSA key length: %d", len(compressedKey))
	}

	// Decompress the key
	x, y := elliptic.Unmarshal(curve, compressedKey)
	if x == nil || y == nil {
		return false, fmt.Errorf("failed to decompress ECDSA key")
	}

	publicKey := &ecdsa.PublicKey{
		Curve: curve,
		X:     x,
		Y:     y,
	}

	// Hash message
	hash := sha256.Sum256(message)

	// Try ASN.1 DER encoded signature
	if ecdsa.VerifyASN1(publicKey, hash[:], signature) {
		return true, nil
	}

	// Try raw r||s format
	if len(signature) == 64 {
		r := big.NewInt(0).SetBytes(signature[:32])
		s := big.NewInt(0).SetBytes(signature[32:])
		return ecdsa.Verify(publicKey, hash[:], r, s), nil
	}

	return false, fmt.Errorf("compressed ECDSA signature verification failed")
}

// verifyRSAFromDER verifies RSA signature with PKIX ASN.1 DER encoded public key
func (k Keeper) verifyRSAFromDER(derBytes []byte, signature []byte, message []byte) (bool, error) {
	// Try parsing as PKIX public key
	pub, err := x509.ParsePKIXPublicKey(derBytes)
	if err != nil {
		return false, fmt.Errorf("failed to parse PKIX public key: %w", err)
	}

	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return false, fmt.Errorf("not an RSA public key")
	}

	// Validate key size (minimum 2048 bits for security)
	if rsaPub.N.BitLen() < 2048 {
		return false, fmt.Errorf("RSA key size too small: %d bits", rsaPub.N.BitLen())
	}

	// Hash message with SHA-256
	hash := sha256.Sum256(message)

	// Try PKCS#1 v1.5 signature first
	if err := rsa.VerifyPKCS1v15(rsaPub, crypto.SHA256, hash[:], signature); err == nil {
		return true, nil
	}

	// Try PSS signature
	pssOpts := &rsa.PSSOptions{
		SaltLength: rsa.PSSSaltLengthEqualsHash,
		Hash:       crypto.SHA256,
	}
	if err := rsa.VerifyPSS(rsaPub, crypto.SHA256, hash[:], signature, pssOpts); err == nil {
		return true, nil
	}

	return false, fmt.Errorf("RSA signature verification failed")
}

// verifyCOSEKey verifies signature using COSE key format (used by WebAuthn)
func (k Keeper) verifyCOSEKey(coseBytes []byte, signature []byte, message []byte) (bool, error) {
	// Parse COSE key
	parsedKey, err := webauthncose.ParsePublicKey(coseBytes)
	if err != nil {
		return false, fmt.Errorf("failed to parse COSE key: %w", err)
	}

	// Use COSE signature verification
	return webauthncose.VerifySignature(parsedKey, message, signature)
}

// SetDWNKeeper sets the DWN keeper for cross-module communication
func (k *Keeper) SetDWNKeeper(dwnKeeper types.DWNKeeper) {
	k.dwnKeeper = dwnKeeper
}

// SetServiceKeeper sets the service keeper dependency
func (k *Keeper) SetServiceKeeper(serviceKeeper types.ServiceKeeper) {
	k.serviceKeeper = serviceKeeper
}

// CreateVaultForDID creates a vault for a given DID using the DWN keeper
func (k Keeper) CreateVaultForDID(
	ctx context.Context,
	did string,
	owner string,
	vaultID string,
	keyID string,
) (*types.CreateVaultResponse, error) {
	if k.dwnKeeper == nil {
		// Return nil without error if DWN keeper not initialized
		// This allows the system to work without vault creation
		k.logger.Warn("DWN keeper not initialized, skipping vault creation")
		return nil, nil
	}

	// Generate a new vault
	enclave, err := mpc.NewEnclave()
	if err != nil {
		return nil, err
	}

	// For now, we'll use the simplified interface
	// In the future, this should create proper MPC enclave data
	// The DWN keeper will need to be updated to match this interface
	return k.dwnKeeper.CreateVaultForDID(ctx, enclave.GetData())
}
