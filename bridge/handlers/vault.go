package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/hibiken/asynq"
	"github.com/labstack/echo/v4"
	"github.com/sonr-io/sonr/bridge/tasks"
	"github.com/sonr-io/sonr/crypto/mpc"
	"github.com/sonr-io/sonr/types/ipfs"
)

// VaultHandlers holds all vault-related handlers and their dependencies
type VaultHandlers struct {
	IPFSClient        ipfs.IPFSClient
	ConnectionManager *ConnectionManager
	SSEManager        *SSEManager
}

// NewVaultHandlers creates a new VaultHandlers instance
func NewVaultHandlers(
	ipfsClient ipfs.IPFSClient,
	connManager *ConnectionManager,
	sseManager *SSEManager,
) *VaultHandlers {
	return &VaultHandlers{
		IPFSClient:        ipfsClient,
		ConnectionManager: connManager,
		SSEManager:        sseManager,
	}
}

// GetQueueFromPriority determines the appropriate queue based on priority
func GetQueueFromPriority(priority string) string {
	switch priority {
	case "critical", "high":
		return "critical"
	case "low":
		return "low"
	default:
		return "default"
	}
}

// GenerateHandler handles vault generation requests
func (vh *VaultHandlers) GenerateHandler(client *asynq.Client) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Extract user info from JWT token
		user := c.Get("user").(*jwt.Token)
		claims := user.Claims.(jwt.MapClaims)
		userID := claims["user_id"].(string)

		var req struct {
			UserID   int    `json:"user_id"`
			Priority string `json:"priority,omitempty"`
		}

		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid JSON payload"})
		}

		task, err := tasks.NewUCANDIDTask(req.UserID)
		if err != nil {
			return c.JSON(
				http.StatusInternalServerError,
				map[string]string{"error": "Failed to create task"},
			)
		}

		queue := GetQueueFromPriority(req.Priority)
		info, err := client.Enqueue(task, asynq.Queue(queue))
		if err != nil {
			return c.JSON(
				http.StatusInternalServerError,
				map[string]string{"error": "Failed to enqueue task"},
			)
		}

		// Broadcast task status to WebSocket and SSE connections
		go func() {
			message := TaskStatusMessage{
				TaskID: info.ID,
				Status: "enqueued",
				Time:   time.Now(),
			}
			vh.ConnectionManager.BroadcastToTask(info.ID, message)
			vh.SSEManager.BroadcastToSSE(info.ID, message)

			// Simulate task progression for demonstration
			// In a real implementation, this would be triggered by the actual task processors
			time.Sleep(1 * time.Second)
			processingMessage := TaskStatusMessage{
				TaskID:   info.ID,
				Status:   "processing",
				Progress: 50,
				Time:     time.Now(),
			}
			vh.ConnectionManager.BroadcastToTask(info.ID, processingMessage)
			vh.SSEManager.BroadcastToSSE(info.ID, processingMessage)

			// Simulate completion
			time.Sleep(2 * time.Second)
			completedMessage := TaskStatusMessage{
				TaskID:   info.ID,
				Status:   "completed",
				Progress: 100,
				Time:     time.Now(),
			}
			vh.ConnectionManager.BroadcastToTask(info.ID, completedMessage)
			vh.SSEManager.BroadcastToSSE(info.ID, completedMessage)
		}()

		return c.JSON(http.StatusOK, map[string]any{
			"task_id": info.ID,
			"queue":   queue,
			"status":  "enqueued",
			"user_id": userID,
		})
	}
}

// SignHandler handles vault signing requests
func (vh *VaultHandlers) SignHandler(client *asynq.Client) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req struct {
			Message    []byte           `json:"message"`
			Enclave    *mpc.EnclaveData `json:"enclave,omitempty"`     // Direct enclave data (fallback)
			EnclaveCID string           `json:"enclave_cid,omitempty"` // IPFS CID reference
			Password   []byte           `json:"password,omitempty"`    // Decryption password for IPFS stored data
			UCANToken  string           `json:"ucan_token,omitempty"`  // UCAN token for authorization
			Priority   string           `json:"priority,omitempty"`
		}

		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid JSON payload"})
		}

		if len(req.Message) == 0 {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Message is required"})
		}

		// Handle enclave data - either direct or via IPFS CID
		var enclave *mpc.EnclaveData
		if req.EnclaveCID != "" {
			// Retrieve enclave from IPFS
			if vh.IPFSClient == nil {
				return c.JSON(
					http.StatusServiceUnavailable,
					map[string]string{"error": "IPFS client not available"},
				)
			}

			encryptedData, err := vh.IPFSClient.Get(req.EnclaveCID)
			if err != nil {
				return c.JSON(
					http.StatusNotFound,
					map[string]string{"error": "Failed to retrieve enclave from IPFS"},
				)
			}

			// Create temporary enclave for decryption
			tempEnclave := &mpc.EnclaveData{}
			decryptedData, err := tempEnclave.Decrypt(req.Password, encryptedData)
			if err != nil {
				return c.JSON(
					http.StatusBadRequest,
					map[string]string{"error": "Failed to decrypt enclave data"},
				)
			}

			// Unmarshal decrypted data
			enclave = &mpc.EnclaveData{}
			if err := enclave.Unmarshal(decryptedData); err != nil {
				return c.JSON(
					http.StatusBadRequest,
					map[string]string{"error": "Failed to parse enclave data"},
				)
			}
		} else if req.Enclave != nil {
			// Use directly provided enclave data
			enclave = req.Enclave
		} else {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Either enclave or enclave_cid is required"})
		}

		task, err := tasks.NewUCANSignTask(
			0,
			req.Message,
		) // TODO: Extract userID from JWT or context
		if err != nil {
			return c.JSON(
				http.StatusInternalServerError,
				map[string]string{"error": "Failed to create task"},
			)
		}

		queue := GetQueueFromPriority(req.Priority)
		info, err := client.Enqueue(task, asynq.Queue(queue))
		if err != nil {
			return c.JSON(
				http.StatusInternalServerError,
				map[string]string{"error": "Failed to enqueue task"},
			)
		}

		return c.JSON(http.StatusOK, map[string]any{
			"task_id": info.ID,
			"queue":   queue,
			"status":  "enqueued",
		})
	}
}

// VerifyHandler handles vault verification requests
func (vh *VaultHandlers) VerifyHandler(client *asynq.Client) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req struct {
			PublicKey []byte `json:"public_key"`
			Message   []byte `json:"message"`
			Signature []byte `json:"signature"`
			Priority  string `json:"priority,omitempty"`
		}

		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid JSON payload"})
		}

		if len(req.PublicKey) == 0 || len(req.Message) == 0 || len(req.Signature) == 0 {
			return c.JSON(
				http.StatusBadRequest,
				map[string]string{"error": "PublicKey, message, and signature are required"},
			)
		}

		task, err := tasks.NewUCANVerifyTask(
			0,
			req.Message,
			req.Signature,
		) // TODO: Extract userID from JWT or context
		if err != nil {
			return c.JSON(
				http.StatusInternalServerError,
				map[string]string{"error": "Failed to create task"},
			)
		}

		queue := GetQueueFromPriority(req.Priority)
		info, err := client.Enqueue(task, asynq.Queue(queue))
		if err != nil {
			return c.JSON(
				http.StatusInternalServerError,
				map[string]string{"error": "Failed to enqueue task"},
			)
		}

		return c.JSON(http.StatusOK, map[string]any{
			"task_id": info.ID,
			"queue":   queue,
			"status":  "enqueued",
		})
	}
}

// ExportHandler handles vault export requests
func (vh *VaultHandlers) ExportHandler(client *asynq.Client) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req struct {
			Enclave    *mpc.EnclaveData `json:"enclave,omitempty"`     // Direct enclave data
			EnclaveCID string           `json:"enclave_cid,omitempty"` // IPFS CID reference
			Password   []byte           `json:"password"`              // For encryption/decryption
			StoreIPFS  bool             `json:"store_ipfs,omitempty"`  // Whether to store result in IPFS
			Priority   string           `json:"priority,omitempty"`
		}

		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid JSON payload"})
		}

		if len(req.Password) == 0 {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Password is required"})
		}

		// Handle enclave data - either direct or via IPFS CID
		var enclave *mpc.EnclaveData
		if req.EnclaveCID != "" {
			// Retrieve enclave from IPFS
			if vh.IPFSClient == nil {
				return c.JSON(
					http.StatusServiceUnavailable,
					map[string]string{"error": "IPFS client not available"},
				)
			}

			encryptedData, err := vh.IPFSClient.Get(req.EnclaveCID)
			if err != nil {
				return c.JSON(
					http.StatusNotFound,
					map[string]string{"error": "Failed to retrieve enclave from IPFS"},
				)
			}

			// Create temporary enclave for decryption
			tempEnclave := &mpc.EnclaveData{}
			decryptedData, err := tempEnclave.Decrypt(req.Password, encryptedData)
			if err != nil {
				return c.JSON(
					http.StatusBadRequest,
					map[string]string{"error": "Failed to decrypt enclave data"},
				)
			}

			// Unmarshal decrypted data
			enclave = &mpc.EnclaveData{}
			if err := enclave.Unmarshal(decryptedData); err != nil {
				return c.JSON(
					http.StatusBadRequest,
					map[string]string{"error": "Failed to parse enclave data"},
				)
			}
		} else if req.Enclave != nil {
			// Use directly provided enclave data
			enclave = req.Enclave
		} else {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Either enclave or enclave_cid is required"})
		}

		// If store_ipfs is true, encrypt and store the enclave in IPFS first
		if req.StoreIPFS && vh.IPFSClient != nil {
			encryptedData, err := enclave.Encrypt(req.Password)
			if err != nil {
				return c.JSON(
					http.StatusInternalServerError,
					map[string]string{"error": "Failed to encrypt enclave data"},
				)
			}

			cid, err := vh.IPFSClient.Add(encryptedData)
			if err != nil {
				return c.JSON(
					http.StatusInternalServerError,
					map[string]string{"error": "Failed to store enclave in IPFS"},
				)
			}

			// Return the CID for future reference
			return c.JSON(http.StatusOK, map[string]any{
				"cid":     cid,
				"status":  "stored",
				"message": "Enclave data encrypted and stored in IPFS",
			})
		}

		// Export functionality replaced with UCAN token creation for data access
		// Convert export operation to UCAN token generation with export permissions
		attenuations := []map[string]any{
			{
				"can":  []string{"export", "read"},
				"with": "vault://exported-data",
			},
		}
		task, err := tasks.NewUCANTokenTask(
			0,
			"did:sonr:export-recipient",
			attenuations,
			time.Now().Add(24*time.Hour).Unix(),
		)
		if err != nil {
			return c.JSON(
				http.StatusInternalServerError,
				map[string]string{"error": "Failed to create task"},
			)
		}

		queue := GetQueueFromPriority(req.Priority)
		info, err := client.Enqueue(task, asynq.Queue(queue))
		if err != nil {
			return c.JSON(
				http.StatusInternalServerError,
				map[string]string{"error": "Failed to enqueue task"},
			)
		}

		return c.JSON(http.StatusOK, map[string]any{
			"task_id": info.ID,
			"queue":   queue,
			"status":  "enqueued",
		})
	}
}

// ImportHandler handles vault import requests
func (vh *VaultHandlers) ImportHandler(client *asynq.Client) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req struct {
			CID      string `json:"cid"`
			Password []byte `json:"password"`
			Priority string `json:"priority,omitempty"`
		}

		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid JSON payload"})
		}

		if req.CID == "" || len(req.Password) == 0 {
			return c.JSON(
				http.StatusBadRequest,
				map[string]string{"error": "CID and password are required"},
			)
		}

		// Import functionality replaced with UCAN token creation for data import
		// Convert import operation to UCAN token generation with import permissions
		attenuations := []map[string]any{
			{
				"can":  []string{"import", "write"},
				"with": fmt.Sprintf("ipfs://%s", req.CID),
			},
		}
		task, err := tasks.NewUCANTokenTask(
			0,
			"did:sonr:import-recipient",
			attenuations,
			time.Now().Add(1*time.Hour).Unix(),
		)
		if err != nil {
			return c.JSON(
				http.StatusInternalServerError,
				map[string]string{"error": "Failed to create task"},
			)
		}

		queue := GetQueueFromPriority(req.Priority)
		info, err := client.Enqueue(task, asynq.Queue(queue))
		if err != nil {
			return c.JSON(
				http.StatusInternalServerError,
				map[string]string{"error": "Failed to enqueue task"},
			)
		}

		return c.JSON(http.StatusOK, map[string]any{
			"task_id": info.ID,
			"queue":   queue,
			"status":  "enqueued",
		})
	}
}

// RefreshHandler handles vault refresh requests
func (vh *VaultHandlers) RefreshHandler(client *asynq.Client) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req struct {
			Enclave    *mpc.EnclaveData `json:"enclave,omitempty"`     // Direct enclave data
			EnclaveCID string           `json:"enclave_cid,omitempty"` // IPFS CID reference
			Password   []byte           `json:"password,omitempty"`    // Decryption password for IPFS stored data
			Priority   string           `json:"priority,omitempty"`
		}

		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid JSON payload"})
		}

		// Handle enclave data - either direct or via IPFS CID
		var enclave *mpc.EnclaveData
		if req.EnclaveCID != "" {
			// Retrieve enclave from IPFS
			if vh.IPFSClient == nil {
				return c.JSON(
					http.StatusServiceUnavailable,
					map[string]string{"error": "IPFS client not available"},
				)
			}

			encryptedData, err := vh.IPFSClient.Get(req.EnclaveCID)
			if err != nil {
				return c.JSON(
					http.StatusNotFound,
					map[string]string{"error": "Failed to retrieve enclave from IPFS"},
				)
			}

			// Create temporary enclave for decryption
			tempEnclave := &mpc.EnclaveData{}
			decryptedData, err := tempEnclave.Decrypt(req.Password, encryptedData)
			if err != nil {
				return c.JSON(
					http.StatusBadRequest,
					map[string]string{"error": "Failed to decrypt enclave data"},
				)
			}

			// Unmarshal decrypted data
			enclave = &mpc.EnclaveData{}
			if err := enclave.Unmarshal(decryptedData); err != nil {
				return c.JSON(
					http.StatusBadRequest,
					map[string]string{"error": "Failed to parse enclave data"},
				)
			}
		} else if req.Enclave != nil {
			// Use directly provided enclave data
			enclave = req.Enclave
		} else {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Either enclave or enclave_cid is required"})
		}

		// Refresh functionality replaced with UCAN DID generation for new identity
		// Convert refresh operation to DID generation which includes key refresh
		task, err := tasks.NewUCANDIDTask(0) // TODO: Extract userID from JWT or context
		if err != nil {
			return c.JSON(
				http.StatusInternalServerError,
				map[string]string{"error": "Failed to create task"},
			)
		}

		queue := GetQueueFromPriority(req.Priority)
		info, err := client.Enqueue(task, asynq.Queue(queue))
		if err != nil {
			return c.JSON(
				http.StatusInternalServerError,
				map[string]string{"error": "Failed to enqueue task"},
			)
		}

		return c.JSON(http.StatusOK, map[string]any{
			"task_id": info.ID,
			"queue":   queue,
			"status":  "enqueued",
		})
	}
}
