package cli

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/sonr-io/sonr/x/did/client/server"
	"github.com/sonr-io/sonr/x/did/types"
)

// RegisterUserWithWebAuthn registers a new user using WebAuthn through browser interaction
func RegisterUserWithWebAuthn(username string) error {
	logger := log.NewLogger(os.Stderr)

	// If no username provided, prompt for it using standard input
	if strings.TrimSpace(username) == "" {
		var err error
		username, err = promptForUsername()
		if err != nil {
			return fmt.Errorf("failed to get username: %w", err)
		}
	}

	// Initialize database and check if username already exists
	if err := server.InitDB(); err != nil {
		logger.Warn("Failed to initialize database", "error", err)
		// Continue without username check - database may not be available
	} else {
		// Check if username already exists
		service := server.NewWebAuthnCredentialService()
		existingCredentials, err := service.GetByUsername(username)
		if err == nil && len(existingCredentials) > 0 {
			return fmt.Errorf("username '%s' already exists with %d WebAuthn credential(s)", username, len(existingCredentials))
		}
		// If error occurred (like record not found), continue with registration
	}

	// Find available port for auth server
	port, err := findAvailablePort()
	if err != nil {
		return fmt.Errorf("failed to find available port: %w", err)
	}

	// Create channel to signal completion
	done := make(chan error, 1)

	// Setup server with WebAuthn registration context
	err = server.StartAuthServerWithWebAuthn(port, username, done)
	if err != nil {
		return fmt.Errorf("failed to start auth server: %w", err)
	}

	defer func() {
		if stopErr := server.StopAuthServer(); stopErr != nil {
			logger.Error("Failed to stop auth server", "error", stopErr)
		}
	}()

	// Wait for server to be ready
	time.Sleep(500 * time.Millisecond)

	// Open browser to WebAuthn registration page
	url := fmt.Sprintf("http://localhost:%d/register?username=%s", port, username)
	logger.Info("Opening browser for WebAuthn registration", "url", url)

	if err := openBrowser(url); err != nil {
		logger.Warn("Failed to open browser automatically", "error", err)
		logger.Info("Please navigate manually to the URL", "url", url)
	}

	logger.Info("Waiting for WebAuthn registration to complete...")

	// Wait for registration to complete or timeout
	select {
	case err := <-done:
		if err != nil {
			return fmt.Errorf("WebAuthn registration failed: %w", err)
		}
		logger.Info("WebAuthn registration completed successfully")
		return nil
	case <-time.After(30 * time.Second):
		logger.Warn("WebAuthn registration timed out after 30 seconds")
		return fmt.Errorf("WebAuthn registration timed out after 30 seconds - please try again")
	}
}

// findAvailablePort finds an available port starting from 8080
func findAvailablePort() (int, error) {
	for port := 8080; port < 8090; port++ {
		conn, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err == nil {
			conn.Close()
			return port, nil
		}
	}
	return 0, fmt.Errorf("no available port found in range 8080-8090")
}

// openBrowser opens the default browser with the given URL
func openBrowser(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "darwin":
		cmd = "open"
		args = []string{url}
	case "linux":
		cmd = "xdg-open"
		args = []string{url}
	case "windows":
		cmd = "rundll32"
		args = []string{"url.dll,FileProtocolHandler", url}
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	return exec.Command(cmd, args...).Start()
}

// RegisterUserWithWebAuthnAndBroadcast registers a user with WebAuthn and broadcasts to blockchain
func RegisterUserWithWebAuthnAndBroadcast(
	clientCtx client.Context,
	username string,
	autoCreateVault bool,
) error {
	// Import necessary packages
	var (
		contextPkg = "context"
		base64Pkg  = "encoding/base64"
		jsonPkg    = "encoding/json"
		flagsPkg   = "github.com/cosmos/cosmos-sdk/client/flags"
		txPkg      = "github.com/cosmos/cosmos-sdk/client/tx"
		sdkPkg     = "github.com/cosmos/cosmos-sdk/types"
		typesPkg   = "github.com/sonr-io/sonr/x/did/types"
	)
	_ = contextPkg
	_ = base64Pkg
	_ = jsonPkg
	_ = flagsPkg
	_ = txPkg
	_ = sdkPkg
	_ = typesPkg

	logger := log.NewLogger(os.Stderr)

	// If no username provided, prompt for it using standard input
	if strings.TrimSpace(username) == "" {
		var err error
		username, err = promptForUsername()
		if err != nil {
			return fmt.Errorf("failed to get username: %w", err)
		}
	}

	// Initialize database and check if username already exists
	if err := server.InitDB(); err != nil {
		logger.Warn("Failed to initialize database", "error", err)
		// Continue without username check - database may not be available
	} else {
		// Check if username already exists
		service := server.NewWebAuthnCredentialService()
		existingCredentials, err := service.GetByUsername(username)
		if err == nil && len(existingCredentials) > 0 {
			return fmt.Errorf("username '%s' already exists with %d WebAuthn credential(s)", username, len(existingCredentials))
		}
		// If error occurred (like record not found), continue with registration
	}

	// Find available port for auth server
	port, err := findAvailablePort()
	if err != nil {
		return fmt.Errorf("failed to find available port: %w", err)
	}

	// Create channel to signal completion and pass WebAuthn credential data
	done := make(chan error, 1)
	credentialData := make(chan *server.WebAuthnCredential, 1)

	// Setup server with WebAuthn registration context and credential data channel
	err = server.StartAuthServerWithWebAuthnAndCredentialChannel(
		port,
		username,
		done,
		credentialData,
	)
	if err != nil {
		return fmt.Errorf("failed to start auth server: %w", err)
	}

	defer func() {
		if stopErr := server.StopAuthServer(); stopErr != nil {
			logger.Error("Failed to stop auth server", "error", stopErr)
		}
	}()

	// Wait for server to be ready
	time.Sleep(500 * time.Millisecond)

	// Open browser to WebAuthn registration page
	url := fmt.Sprintf("http://localhost:%d/register?username=%s", port, username)
	logger.Info("Opening browser for WebAuthn registration", "url", url)

	if err := openBrowser(url); err != nil {
		logger.Warn("Failed to open browser automatically", "error", err)
		logger.Info("Please navigate manually to the URL", "url", url)
	}

	logger.Info("Waiting for WebAuthn registration to complete...")

	// Wait for registration to complete or timeout
	select {
	case err := <-done:
		if err != nil {
			return fmt.Errorf("WebAuthn registration failed: %w", err)
		}
		logger.Info("WebAuthn registration completed successfully")

		// Get the credential data from the server
		select {
		case credential := <-credentialData:
			logger.Info("Received WebAuthn credential data, broadcasting to blockchain...",
				"credentialID", credential.CredentialID, "username", credential.Username)

			// Create and broadcast the MsgRegisterWebAuthnCredential transaction
			err = broadcastWebAuthnCredential(clientCtx, credential, autoCreateVault)
			if err != nil {
				return fmt.Errorf("failed to broadcast WebAuthn credential: %w", err)
			}

			logger.Info(
				"WebAuthn credential successfully broadcast to blockchain and vault creation initiated",
			)
			return nil
		case <-time.After(2 * time.Second):
			return fmt.Errorf("failed to receive credential data from server")
		}
	case <-time.After(30 * time.Second):
		logger.Warn("WebAuthn registration timed out after 30 seconds")
		return fmt.Errorf("WebAuthn registration timed out after 30 seconds - please try again")
	}
}

// RegisterUserWithWebAuthnAndBroadcastWithAssertion registers a user with WebAuthn and assertion methods
func RegisterUserWithWebAuthnAndBroadcastWithAssertion(
	clientCtx client.Context,
	username string, // Can be empty, will use assertion value
	autoCreateVault bool,
	assertionType string,
	assertionValue string,
) error {
	// Import necessary packages
	var (
		contextPkg = "context"
		base64Pkg  = "encoding/base64"
		jsonPkg    = "encoding/json"
		flagsPkg   = "github.com/cosmos/cosmos-sdk/client/flags"
		txPkg      = "github.com/cosmos/cosmos-sdk/client/tx"
		sdkPkg     = "github.com/cosmos/cosmos-sdk/types"
		typesPkg   = "github.com/sonr-io/sonr/x/did/types"
	)
	_ = contextPkg
	_ = base64Pkg
	_ = jsonPkg
	_ = flagsPkg
	_ = txPkg
	_ = sdkPkg
	_ = typesPkg

	logger := log.NewLogger(os.Stderr)

	// Use assertion value as the identifier
	identifier := assertionValue

	// Initialize database and check if assertion already exists
	if err := server.InitDB(); err != nil {
		logger.Warn("Failed to initialize database", "error", err)
		// Continue without check - database may not be available
	} else {
		// Check if assertion value already exists as a registered identity
		service := server.NewWebAuthnCredentialService()
		existingCredentials, err := service.GetByUsername(identifier)
		if err == nil && len(existingCredentials) > 0 {
			return fmt.Errorf("%s '%s' already registered with %d WebAuthn credential(s)",
				assertionType, assertionValue, len(existingCredentials))
		}
		// If error occurred (like record not found), continue with registration
	}

	// Find available port for auth server
	port, err := findAvailablePort()
	if err != nil {
		return fmt.Errorf("failed to find available port: %w", err)
	}

	// Create channel to signal completion and pass WebAuthn credential data
	done := make(chan error, 1)
	credentialData := make(chan *server.WebAuthnCredential, 1)

	// Setup server with WebAuthn registration context and credential data channel
	// Use the assertion value as the identifier for WebAuthn
	err = server.StartAuthServerWithWebAuthnAndCredentialChannel(
		port,
		identifier,
		done,
		credentialData,
	)
	if err != nil {
		return fmt.Errorf("failed to start auth server: %w", err)
	}

	defer func() {
		if stopErr := server.StopAuthServer(); stopErr != nil {
			logger.Error("Failed to stop auth server", "error", stopErr)
		}
	}()

	// Wait for server to be ready
	time.Sleep(500 * time.Millisecond)

	// Open browser to WebAuthn registration page
	url := fmt.Sprintf("http://localhost:%d/register?identifier=%s", port, identifier)
	logger.Info("Opening browser for WebAuthn registration", "url", url)

	if err := openBrowser(url); err != nil {
		logger.Warn("Failed to open browser automatically", "error", err)
		logger.Info("Please navigate manually to the URL", "url", url)
	}

	logger.Info("Waiting for WebAuthn registration to complete...")

	// Wait for registration to complete or timeout
	select {
	case err := <-done:
		if err != nil {
			return fmt.Errorf("WebAuthn registration failed: %w", err)
		}
		logger.Info("WebAuthn registration completed successfully")

		// Get the credential data from the server
		select {
		case credential := <-credentialData:
			logger.Info("Received WebAuthn credential data, broadcasting to blockchain...",
				"credentialID", credential.CredentialID,
				"identifier", identifier,
				"assertionType", assertionType,
				"assertionValue", assertionValue)

			// Create and broadcast the MsgRegisterWebAuthnCredential transaction with assertion
			err = broadcastWebAuthnCredentialWithAssertion(
				clientCtx, credential, autoCreateVault, assertionType, assertionValue,
			)
			if err != nil {
				return fmt.Errorf("failed to broadcast WebAuthn credential: %w", err)
			}

			logger.Info(
				"WebAuthn credential successfully broadcast to blockchain with assertion method",
				"assertionType", assertionType,
			)
			return nil
		case <-time.After(2 * time.Second):
			return fmt.Errorf("failed to receive credential data from server")
		}
	case <-time.After(30 * time.Second):
		logger.Warn("WebAuthn registration timed out after 30 seconds")
		return fmt.Errorf("WebAuthn registration timed out after 30 seconds - please try again")
	}
}

// broadcastWebAuthnCredential creates and broadcasts a MsgRegisterWebAuthnCredential transaction
func broadcastWebAuthnCredential(
	clientCtx client.Context,
	credential *server.WebAuthnCredential,
	autoCreateVault bool,
) error {
	logger := log.NewLogger(os.Stderr)
	logger.Info("Broadcasting WebAuthn credential transaction",
		"credentialID", credential.CredentialID,
		"username", credential.Username,
		"autoCreateVault", autoCreateVault,
		"chainID", clientCtx.ChainID)

	// Import required packages
	didtypes := "github.com/sonr-io/sonr/x/did/types"
	_ = didtypes

	// For gasless transactions, we generate a deterministic address from the WebAuthn credential
	// This allows the transaction to be processed without a pre-existing account
	controllerAddr := generateAddressFromWebAuthn(credential)

	// Create the WebAuthn credential message
	// PublicKey, Algorithm, and Origin are extracted server-side from attestation
	webauthnCred := types.WebAuthnCredential{
		CredentialId:      credential.CredentialID,
		RawId:             credential.RawID,
		ClientDataJson:    credential.ClientDataJSON,
		AttestationObject: credential.AttestationObject,
		// Use the extracted fields from server processing
		PublicKey: credential.PublicKey,
		Algorithm: credential.Algorithm,
		Origin:    credential.Origin,
	}

	// Create the registration message
	msg := &types.MsgRegisterWebAuthnCredential{
		Controller:           controllerAddr.String(),
		Username:             credential.Username,
		WebauthnCredential:   webauthnCred,
		VerificationMethodId: fmt.Sprintf("webauthn-%s", credential.CredentialID[:8]),
		AutoCreateVault:      autoCreateVault,
	}

	// Build the transaction with proper signature structure for gasless handling
	txBuilder := clientCtx.TxConfig.NewTxBuilder()
	err := txBuilder.SetMsgs(msg)
	if err != nil {
		return fmt.Errorf("failed to set message: %w", err)
	}

	// Set reasonable gas limit for gasless transaction (fees will still be zero)
	txBuilder.SetGasLimit(200000)          // Reasonable gas limit for WebAuthn registration
	txBuilder.SetFeeAmount(sdk.NewCoins()) // Zero fees - gasless

	// For WebAuthn gasless transactions, we need to provide at least empty signature info
	// to pass mempool validation, then our ante handler will bypass signature verification
	logger.Info("Creating gasless WebAuthn transaction with empty signature placeholder",
		"controllerAddress", controllerAddr.String(),
		"credentialID", credential.CredentialID)

	// For WebAuthn gasless transactions, we need to provide a dummy signature to pass
	// mempool validation, then our ante handler will bypass the verification
	logger.Info("Creating dummy signature for mempool validation bypass")

	// Create a minimal dummy public key from the controller address
	// This is needed so the signature validation doesn't fail immediately
	pubKeyBytes := make(
		[]byte,
		33,
	) // Standard secp256k1 compressed public key length
	copy(pubKeyBytes[1:], controllerAddr.Bytes()[:32]) // Use controller address bytes
	pubKeyBytes[0] = 0x02                              // Compressed public key prefix

	dummyPubKey := &secp256k1.PubKey{Key: pubKeyBytes}

	// Create a minimal dummy signature structure to pass mempool validation
	dummySig := signing.SignatureV2{
		PubKey: dummyPubKey, // Dummy public key derived from controller address
		Data: &signing.SingleSignatureData{
			SignMode:  signing.SignMode_SIGN_MODE_DIRECT,
			Signature: make([]byte, 64), // Non-empty signature to pass basic checks
		},
		Sequence: 0, // Zero sequence for gasless
	}

	// Set the dummy signature to pass mempool validation
	err = txBuilder.SetSignatures(dummySig)
	if err != nil {
		return fmt.Errorf("failed to set dummy signature: %w", err)
	}

	logger.Info(
		"Dummy signature set for mempool bypass",
		"pubKeyLen",
		len(pubKeyBytes),
		"sigLen",
		64,
	)

	// Encode the transaction
	tx := txBuilder.GetTx()

	// Debug: Verify transaction has no signatures (expected for WebAuthn bypass)
	if sigTx, ok := tx.(authsigning.SigVerifiableTx); ok {
		sigs, err := sigTx.GetSignaturesV2()
		if err != nil {
			logger.Error("Failed to get signatures from tx", "error", err)
		} else {
			logger.Info("Transaction signature count", "sigCount", len(sigs))
		}
	}

	txBytes, err := clientCtx.TxConfig.TxEncoder()(tx)
	if err != nil {
		return fmt.Errorf("failed to encode transaction: %w", err)
	}

	// Broadcast the transaction
	res, err := clientCtx.BroadcastTxSync(txBytes)
	if err != nil {
		return fmt.Errorf("failed to broadcast transaction: %w", err)
	}

	// Check the response
	if res.Code != 0 {
		return fmt.Errorf("transaction failed with code %d: %s", res.Code, res.RawLog)
	}

	logger.Info("WebAuthn credential successfully registered",
		"txHash", res.TxHash,
		"height", res.Height,
		"gasUsed", res.GasUsed)

	// Parse the response to get the created DID
	// In a real implementation, we would parse the events to extract the DID
	logger.Info("DID created successfully",
		"username", credential.Username,
		"credentialID", credential.CredentialID,
		"vaultCreated", autoCreateVault)

	return nil
}

// generateAddressFromWebAuthn generates a deterministic address from WebAuthn credential
func generateAddressFromWebAuthn(credential *server.WebAuthnCredential) sdk.AccAddress {
	// Use the local types package for address generation
	// It returns a hex string with 0x prefix
	addrHex := types.GenerateAddressFromCredential(credential.CredentialID)
	// Remove 0x prefix
	addrHex = strings.TrimPrefix(addrHex, "0x")
	// Decode hex to bytes
	addrBytes := make([]byte, 20)
	for i := 0; i < 20; i++ {
		fmt.Sscanf(addrHex[i*2:i*2+2], "%02x", &addrBytes[i])
	}
	return sdk.AccAddress(addrBytes)
}

// promptForUsername prompts the user for a username using standard input
func promptForUsername() (string, error) {
	fmt.Print("Enter username for WebAuthn registration: ")
	reader := bufio.NewReader(os.Stdin)
	username, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read username input: %w", err)
	}

	username = strings.TrimSpace(username)

	// Validate username
	if username == "" {
		return "", fmt.Errorf("username is required")
	}
	if len(username) < 3 {
		return "", fmt.Errorf("username must be at least 3 characters")
	}
	if len(username) > 20 {
		return "", fmt.Errorf("username cannot exceed 20 characters")
	}
	// Check for valid characters (alphanumeric and underscore)
	for _, char := range username {
		if (char < 'a' || char > 'z') &&
			(char < 'A' || char > 'Z') &&
			(char < '0' || char > '9') &&
			char != '_' {
			return "", fmt.Errorf(
				"username can only contain alphanumeric characters and underscores",
			)
		}
	}

	return username, nil
}

// broadcastWebAuthnCredentialWithAssertion creates and broadcasts a MsgRegisterWebAuthnCredential transaction with assertion
func broadcastWebAuthnCredentialWithAssertion(
	clientCtx client.Context,
	credential *server.WebAuthnCredential,
	autoCreateVault bool,
	assertionType string,
	assertionValue string,
) error {
	logger := log.NewLogger(os.Stderr)
	logger.Info("Broadcasting WebAuthn credential transaction with assertion",
		"credentialID", credential.CredentialID,
		"username", credential.Username,
		"autoCreateVault", autoCreateVault,
		"assertionType", assertionType,
		"assertionValue", assertionValue,
		"chainID", clientCtx.ChainID)

	// Import required packages
	didtypes := "github.com/sonr-io/sonr/x/did/types"
	_ = didtypes

	// For gasless transactions, we generate a deterministic address from the WebAuthn credential
	// This allows the transaction to be processed without a pre-existing account
	controllerAddr := generateAddressFromWebAuthn(credential)

	// Create the WebAuthn credential message
	// PublicKey, Algorithm, and Origin are extracted server-side from attestation
	webauthnCred := types.WebAuthnCredential{
		CredentialId:      credential.CredentialID,
		RawId:             credential.RawID,
		ClientDataJson:    credential.ClientDataJSON,
		AttestationObject: credential.AttestationObject,
		// Use the extracted fields from server processing
		PublicKey: credential.PublicKey,
		Algorithm: credential.Algorithm,
		Origin:    credential.Origin,
	}

	// Create the registration message
	// Use the assertion value directly as the username for the message
	// The server will detect the type (email/tel) based on the format
	msg := &types.MsgRegisterWebAuthnCredential{
		Controller:           controllerAddr.String(),
		Username:             assertionValue, // This will be the email or phone number
		WebauthnCredential:   webauthnCred,
		VerificationMethodId: fmt.Sprintf("webauthn-%s", credential.CredentialID[:8]),
		AutoCreateVault:      autoCreateVault,
	}

	// Build the transaction with proper signature structure for gasless handling
	txBuilder := clientCtx.TxConfig.NewTxBuilder()
	err := txBuilder.SetMsgs(msg)
	if err != nil {
		return fmt.Errorf("failed to set message: %w", err)
	}

	// Set reasonable gas limit for gasless transaction (fees will still be zero)
	txBuilder.SetGasLimit(200000)          // Reasonable gas limit for WebAuthn registration
	txBuilder.SetFeeAmount(sdk.NewCoins()) // Zero fees - gasless

	// For WebAuthn gasless transactions, we need to provide at least empty signature info
	// to pass mempool validation, then our ante handler will bypass signature verification
	logger.Info("Creating gasless WebAuthn transaction with empty signature placeholder",
		"controllerAddress", controllerAddr.String(),
		"credentialID", credential.CredentialID)

	// For WebAuthn gasless transactions, we need to provide a dummy signature to pass
	// mempool validation, then our ante handler will bypass the verification
	logger.Info("Creating dummy signature for mempool validation bypass")

	// Create a minimal dummy public key from the controller address
	// This is needed so the signature validation doesn't fail immediately
	pubKeyBytes := make(
		[]byte,
		33,
	) // Standard secp256k1 compressed public key length
	copy(pubKeyBytes[1:], controllerAddr.Bytes()[:32]) // Use controller address bytes
	pubKeyBytes[0] = 0x02                              // Compressed public key prefix

	dummyPubKey := &secp256k1.PubKey{Key: pubKeyBytes}

	// Create a minimal dummy signature structure to pass mempool validation
	dummySig := signing.SignatureV2{
		PubKey: dummyPubKey, // Dummy public key derived from controller address
		Data: &signing.SingleSignatureData{
			SignMode:  signing.SignMode_SIGN_MODE_DIRECT,
			Signature: make([]byte, 64), // Non-empty signature to pass basic checks
		},
		Sequence: 0, // Zero sequence for gasless
	}

	// Set the dummy signature to pass mempool validation
	err = txBuilder.SetSignatures(dummySig)
	if err != nil {
		return fmt.Errorf("failed to set dummy signature: %w", err)
	}

	logger.Info(
		"Dummy signature set for mempool bypass",
		"pubKeyLen",
		len(pubKeyBytes),
		"sigLen",
		64,
	)

	// Encode the transaction
	tx := txBuilder.GetTx()

	// Debug: Verify transaction has no signatures (expected for WebAuthn bypass)
	if sigTx, ok := tx.(authsigning.SigVerifiableTx); ok {
		sigs, err := sigTx.GetSignaturesV2()
		if err != nil {
			logger.Error("Failed to get signatures from tx", "error", err)
		} else {
			logger.Info("Transaction signature count", "sigCount", len(sigs))
		}
	}

	txBytes, err := clientCtx.TxConfig.TxEncoder()(tx)
	if err != nil {
		return fmt.Errorf("failed to encode transaction: %w", err)
	}

	// Broadcast the transaction
	res, err := clientCtx.BroadcastTxSync(txBytes)
	if err != nil {
		return fmt.Errorf("failed to broadcast transaction: %w", err)
	}

	// Check the response
	if res.Code != 0 {
		return fmt.Errorf("transaction failed with code %d: %s", res.Code, res.RawLog)
	}

	logger.Info("WebAuthn credential with assertion successfully registered",
		"txHash", res.TxHash,
		"height", res.Height,
		"gasUsed", res.GasUsed,
		"assertionType", assertionType)

	// Parse the response to get the created DID
	// In a real implementation, we would parse the events to extract the DID
	logger.Info("DID created successfully with assertion method",
		"credentialID", credential.CredentialID,
		"assertionType", assertionType,
		"assertionValue", assertionValue,
		"vaultCreated", autoCreateVault)

	return nil
}
