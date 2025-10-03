package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

// BroadcastService handles blockchain broadcasting operations
type BroadcastService struct {
	// TODO: Add cosmos SDK client for actual broadcasting
}

var broadcastService = &BroadcastService{}

// HandleBroadcast handles generic message broadcasting to blockchain
func HandleBroadcast(c echo.Context) error {
	var req BroadcastRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid broadcast request",
		})
	}

	// Validate request
	if req.Message == nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Message is required",
		})
	}

	// Process based on gasless flag
	if req.Gasless {
		return handleGaslessBroadcast(c, &req)
	}

	return handleStandardBroadcast(c, &req)
}

// handleGaslessBroadcast handles gasless transaction broadcasting
func handleGaslessBroadcast(c echo.Context, req *BroadcastRequest) error {
	// Validate gasless eligibility
	if !isGaslessEligible(req.Message) {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Message not eligible for gasless broadcast",
		})
	}

	// TODO: Implement actual blockchain broadcast
	// For now, simulate successful broadcast
	response := &BroadcastResponse{
		TxHash:  generateTxHash(),
		Height:  12345,
		Code:    0,
		RawLog:  "Transaction broadcast successfully",
		Success: true,
	}

	return c.JSON(http.StatusOK, response)
}

// handleStandardBroadcast handles regular transaction broadcasting
func handleStandardBroadcast(c echo.Context, req *BroadcastRequest) error {
	// Get sender address
	fromAddress := req.FromAddress
	if fromAddress == "" {
		// Try to get from context
		userDID := c.Get("user_did")
		if userDID != nil {
			fromAddress = deriveAddressFromDID(userDID.(string))
		}
	}

	if fromAddress == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "From address is required for standard broadcast",
		})
	}

	// TODO: Implement actual blockchain broadcast
	// For now, simulate successful broadcast
	response := &BroadcastResponse{
		TxHash:  generateTxHash(),
		Height:  12346,
		Code:    0,
		RawLog:  "Transaction broadcast successfully",
		Success: true,
	}

	return c.JSON(http.StatusOK, response)
}

// BroadcastWebAuthnRegistration broadcasts WebAuthn registration to blockchain
func BroadcastWebAuthnRegistration(
	credential *WebAuthnCredential,
	gasless bool,
) (*BroadcastResponse, error) {
	// Create MsgRegisterWebAuthnCredential
	msg := map[string]any{
		"@type":              "/sonr.did.v1.MsgRegisterWebAuthnCredential",
		"username":           credential.Username,
		"credential_id":      credential.CredentialID,
		"public_key":         credential.PublicKey,
		"attestation_object": credential.AttestationObject,
		"client_data_json":   credential.ClientDataJSON,
		"origin":             credential.Origin,
		"algorithm":          credential.Algorithm,
	}

	// Create broadcast request (for future use with actual broadcast)
	_ = &BroadcastRequest{
		Message:  msg,
		Gasless:  gasless,
		AutoSign: true,
	}

	// Simulate broadcast (TODO: Implement actual broadcast)
	return &BroadcastResponse{
		TxHash:  generateTxHash(),
		Height:  12347,
		Code:    0,
		RawLog:  "WebAuthn credential registered successfully",
		Success: true,
	}, nil
}

// BroadcastVaultCreation broadcasts vault creation to blockchain
func BroadcastVaultCreation(
	userDID string,
	vaultConfig map[string]any,
) (*BroadcastResponse, error) {
	// Create MsgCreateVault
	msg := map[string]any{
		"@type":   "/sonr.vault.v1.MsgCreateVault",
		"creator": userDID,
		"config":  vaultConfig,
	}

	// Create broadcast request (for future use with actual broadcast)
	_ = &BroadcastRequest{
		Message:     msg,
		Gasless:     true, // Vault creation is gasless for new users
		AutoSign:    true,
		FromAddress: deriveAddressFromDID(userDID),
	}

	// Simulate broadcast (TODO: Implement actual broadcast)
	return &BroadcastResponse{
		TxHash:  generateTxHash(),
		Height:  12348,
		Code:    0,
		RawLog:  "Vault created successfully",
		Success: true,
	}, nil
}

// BroadcastDIDDocument broadcasts DID document to blockchain
func BroadcastDIDDocument(didDoc map[string]any, gasless bool) (*BroadcastResponse, error) {
	// Create MsgCreateDIDDocument or MsgUpdateDIDDocument
	msg := map[string]any{
		"@type":        "/sonr.did.v1.MsgCreateDIDDocument",
		"did_document": didDoc,
	}

	// Create broadcast request (for future use with actual broadcast)
	_ = &BroadcastRequest{
		Message:  msg,
		Gasless:  gasless,
		AutoSign: true,
	}

	// Simulate broadcast (TODO: Implement actual broadcast)
	return &BroadcastResponse{
		TxHash:  generateTxHash(),
		Height:  12349,
		Code:    0,
		RawLog:  "DID document created successfully",
		Success: true,
	}, nil
}

// HandleTransactionStatus checks transaction status
func HandleTransactionStatus(c echo.Context) error {
	txHash := c.Param("hash")
	if txHash == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Transaction hash is required",
		})
	}

	// TODO: Query actual blockchain for transaction status
	// For now, return simulated status
	status := map[string]any{
		"tx_hash":    txHash,
		"height":     12350,
		"status":     "confirmed",
		"code":       0,
		"gas_used":   50000,
		"gas_wanted": 100000,
		"timestamp":  time.Now().Unix(),
	}

	return c.JSON(http.StatusOK, status)
}

// HandleEstimateGas estimates gas for a transaction
func HandleEstimateGas(c echo.Context) error {
	var req BroadcastRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request",
		})
	}

	// Check if gasless eligible
	if isGaslessEligible(req.Message) {
		return c.JSON(http.StatusOK, map[string]any{
			"gas_estimate": 0,
			"gasless":      true,
			"fee":          "0usnr",
		})
	}

	// TODO: Implement actual gas estimation
	// For now, return default estimate
	return c.JSON(http.StatusOK, map[string]any{
		"gas_estimate": 100000,
		"gasless":      false,
		"fee":          "100usnr",
	})
}

// Helper functions

// isGaslessEligible checks if a message is eligible for gasless broadcasting
func isGaslessEligible(message any) bool {
	// Check message type
	msgMap, ok := message.(map[string]any)
	if !ok {
		return false
	}

	msgType, ok := msgMap["@type"].(string)
	if !ok {
		return false
	}

	// WebAuthn registration and vault creation are gasless
	gaslessTypes := []string{
		"/sonr.did.v1.MsgRegisterWebAuthnCredential",
		"/sonr.vault.v1.MsgCreateVault",
		"/sonr.did.v1.MsgCreateDIDDocument",
	}

	for _, t := range gaslessTypes {
		if msgType == t {
			return true
		}
	}

	return false
}

// deriveAddressFromDID derives a blockchain address from a DID
func deriveAddressFromDID(did string) string {
	// TODO: Implement actual address derivation
	// For now, return a placeholder address
	return "sonr1placeholder" + did[len(did)-10:]
}

// generateTxHash generates a mock transaction hash
func generateTxHash() string {
	// TODO: Replace with actual tx hash from broadcast
	return fmt.Sprintf("%X", time.Now().UnixNano())
}

// CreateWebAuthnBroadcastHandler creates a handler that broadcasts WebAuthn credentials
func CreateWebAuthnBroadcastHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		// Get WebAuthn credential from context
		credential, ok := c.Get("webauthn_credential").(*WebAuthnCredential)
		if !ok {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "No WebAuthn credential found",
			})
		}

		// Broadcast to blockchain
		response, err := BroadcastWebAuthnRegistration(credential, true)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": err.Error(),
			})
		}

		return c.JSON(http.StatusOK, response)
	}
}

// CreateVaultBroadcastHandler creates a handler that broadcasts vault creation
func CreateVaultBroadcastHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		// Get user DID from context
		userDID, ok := c.Get("user_did").(string)
		if !ok {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "User DID not found",
			})
		}

		// Get vault config from request
		var vaultConfig map[string]any
		if err := c.Bind(&vaultConfig); err != nil {
			// Use default config
			vaultConfig = map[string]any{
				"type":       "standard",
				"encryption": "AES256",
			}
		}

		// Broadcast vault creation
		response, err := BroadcastVaultCreation(userDID, vaultConfig)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": err.Error(),
			})
		}

		return c.JSON(http.StatusOK, response)
	}
}
