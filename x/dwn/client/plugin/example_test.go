package plugin_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/sonr-io/crypto/mpc"
	"github.com/sonr-io/sonr/x/dwn/client/plugin"
)

// ExampleLoadPluginWithEnclave demonstrates how to load and use the refactored
// Motor plugin as an MPC-based UCAN KeyshareSource.
func ExampleLoadPluginWithEnclave() {
	ctx := context.Background()

	// Example MPC enclave data (in practice, this would be real enclave data)
	enclaveData := &mpc.EnclaveData{
		// This would contain actual MPC enclave configuration
		// For example purposes, we'll assume this is properly initialized
	}

	// Serialize enclave data for the plugin
	enclaveJSON, err := json.Marshal(enclaveData)
	if err != nil {
		fmt.Printf("Failed to marshal enclave data: %v\n", err)
		return
	}

	// Optional vault configuration
	vaultConfig := map[string]any{
		"auto_lock_timeout":     300,   // 5 minutes
		"key_rotation_interval": 86400, // 24 hours
		"supported_chains":      []string{"sonr", "cosmos", "ethereum"},
	}

	// Load the plugin with enclave configuration
	p, err := plugin.LoadPluginWithEnclave(ctx, "sonr-testnet-1", enclaveJSON, vaultConfig)
	if err != nil {
		fmt.Printf("Failed to load plugin: %v\n", err)
		return
	}

	// Get issuer DID and address
	issuerResp, err := p.GetIssuerDID()
	if err != nil {
		fmt.Printf("Failed to get issuer DID: %v\n", err)
		return
	}

	fmt.Printf("Issuer DID: %s\n", issuerResp.IssuerDID)
	fmt.Printf("Address: %s\n", issuerResp.Address)
	fmt.Printf("Chain Code: %s\n", issuerResp.ChainCode)

	// Create a UCAN origin token
	originReq := &plugin.NewOriginTokenRequest{
		AudienceDID: "did:sonr:example-audience",
		Attenuations: []map[string]any{
			{
				"can":  []string{"sign", "verify"},
				"with": "vault://example-vault",
			},
		},
		ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
	}

	originResp, err := p.NewOriginToken(originReq)
	if err != nil {
		fmt.Printf("Failed to create origin token: %v\n", err)
		return
	}

	fmt.Printf("Origin Token: %s\n", originResp.Token)

	// Create an attenuated token from the origin token
	attenuatedReq := &plugin.NewAttenuatedTokenRequest{
		ParentToken: originResp.Token,
		AudienceDID: "did:sonr:delegated-user",
		Attenuations: []map[string]any{
			{
				"can":  []string{"sign"}, // More restrictive than parent
				"with": "vault://example-vault",
			},
		},
		ExpiresAt: time.Now().Add(1 * time.Hour).Unix(), // Shorter than parent
	}

	attenuatedResp, err := p.NewAttenuatedToken(attenuatedReq)
	if err != nil {
		fmt.Printf("Failed to create attenuated token: %v\n", err)
		return
	}

	fmt.Printf("Attenuated Token: %s\n", attenuatedResp.Token)

	// Sign some data
	signReq := &plugin.SignDataRequest{
		Data: []byte("Hello, UCAN world!"),
	}

	signResp, err := p.SignData(signReq)
	if err != nil {
		fmt.Printf("Failed to sign data: %v\n", err)
		return
	}

	fmt.Printf("Signature: %x\n", signResp.Signature)

	// Verify the signature
	verifyReq := &plugin.VerifyDataRequest{
		Data:      []byte("Hello, UCAN world!"),
		Signature: signResp.Signature,
	}

	verifyResp, err := p.VerifyData(verifyReq)
	if err != nil {
		fmt.Printf("Failed to verify signature: %v\n", err)
		return
	}

	fmt.Printf("Signature valid: %t\n", verifyResp.Valid)

	fmt.Println("Example completed successfully!")
}

// TestAdvancedOperationsExample demonstrates advanced UCAN operations
// including comprehensive signing and verification workflows.
func TestAdvancedOperationsExample(t *testing.T) {
	t.Skip("Example test - skipped during normal test runs")
	ctx := context.Background()

	// Example MPC enclave data
	enclaveData, _ := json.Marshal(&mpc.EnclaveData{})

	// Load plugin with enhanced configuration
	p, err := plugin.LoadPluginWithEnclave(ctx, "sonr-testnet-1", enclaveData, nil)
	if err != nil {
		fmt.Printf("Failed to load plugin: %v\n", err)
		return
	}

	// UCAN-based signing approach
	signReq := &plugin.SignDataRequest{
		Data: []byte("UCAN secure message"),
	}

	signResp, err := p.SignData(signReq)
	if err != nil {
		fmt.Printf("Signing failed: %v\n", err)
		return
	}

	fmt.Printf("Signature: %x\n", signResp.Signature)

	// Verify the signature
	verifyReq := &plugin.VerifyDataRequest{
		Data:      []byte("UCAN secure message"),
		Signature: signResp.Signature,
	}

	verifyResp, err := p.VerifyData(verifyReq)
	if err != nil {
		fmt.Printf("Verification failed: %v\n", err)
		return
	}

	fmt.Printf("Signature valid: %t\n", verifyResp.Valid)

	fmt.Println("Enhanced UCAN operations completed!")
}

// TestTokenWorkflowExample demonstrates a complete UCAN token workflow
// including token creation, delegation, and validation.
func TestTokenWorkflowExample(t *testing.T) {
	t.Skip("Example test - skipped during normal test runs")
	ctx := context.Background()

	// Load plugin with enclave configuration
	enclaveData, _ := json.Marshal(&mpc.EnclaveData{})
	p, err := plugin.LoadPluginWithEnclave(ctx, "sonr-testnet-1", enclaveData, nil)
	if err != nil {
		fmt.Printf("Failed to load plugin: %v\n", err)
		return
	}

	// Step 1: Get issuer information
	issuer, err := p.GetIssuerDID()
	if err != nil {
		fmt.Printf("Failed to get issuer: %v\n", err)
		return
	}

	fmt.Printf("Vault Issuer: %s\n", issuer.IssuerDID)

	// Step 2: Create admin token with broad permissions
	adminReq := &plugin.NewOriginTokenRequest{
		AudienceDID: "did:sonr:admin",
		Attenuations: []map[string]any{
			{
				"can":  []string{"admin", "sign", "verify", "delegate"},
				"with": fmt.Sprintf("vault://%s", issuer.Address),
			},
		},
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour).Unix(), // 7 days
	}

	adminToken, err := p.NewOriginToken(adminReq)
	if err != nil {
		fmt.Printf("Failed to create admin token: %v\n", err)
		return
	}

	fmt.Printf("Admin Token: %s...\n", adminToken.Token[:50])

	// Step 3: Delegate signing permission to a user
	userReq := &plugin.NewAttenuatedTokenRequest{
		ParentToken: adminToken.Token,
		AudienceDID: "did:sonr:user123",
		Attenuations: []map[string]any{
			{
				"can":  []string{"sign"}, // Only signing permission
				"with": fmt.Sprintf("vault://%s", issuer.Address),
			},
		},
		ExpiresAt: time.Now().Add(24 * time.Hour).Unix(), // 24 hours
	}

	userToken, err := p.NewAttenuatedToken(userReq)
	if err != nil {
		fmt.Printf("Failed to create user token: %v\n", err)
		return
	}

	fmt.Printf("User Token: %s...\n", userToken.Token[:50])

	// Step 4: Further delegate read-only permission
	readOnlyReq := &plugin.NewAttenuatedTokenRequest{
		ParentToken: userToken.Token,
		AudienceDID: "did:sonr:readonly",
		Attenuations: []map[string]any{
			{
				"can":  []string{"verify"}, // Only verification permission
				"with": fmt.Sprintf("vault://%s", issuer.Address),
			},
		},
		ExpiresAt: time.Now().Add(1 * time.Hour).Unix(), // 1 hour
	}

	readOnlyToken, err := p.NewAttenuatedToken(readOnlyReq)
	if err != nil {
		fmt.Printf("Failed to create read-only token: %v\n", err)
		return
	}

	fmt.Printf("Read-only Token: %s...\n", readOnlyToken.Token[:50])

	fmt.Println("UCAN token delegation workflow completed successfully!")
}
