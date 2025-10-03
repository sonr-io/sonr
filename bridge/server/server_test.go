package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/sonr-io/sonr/bridge/handlers"
	"github.com/sonr-io/sonr/types/ipfs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockIPFSClient provides a test implementation of IPFSClient for server tests
type MockIPFSClient struct{}

func (m *MockIPFSClient) Add(data []byte) (string, error) { return "mock-cid", nil }

func (m *MockIPFSClient) AddFile(
	file ipfs.File,
) (string, error) {
	return "mock-file-cid", nil
}

func (m *MockIPFSClient) AddFolder(
	folder ipfs.Folder,
) (string, error) {
	return "mock-folder-cid", nil
}

func (m *MockIPFSClient) Get(
	cid string,
) ([]byte, error) {
	return []byte("mock-ipfs-data"), nil
}
func (m *MockIPFSClient) GetFile(cid string) (ipfs.File, error)     { return nil, nil }
func (m *MockIPFSClient) GetFolder(cid string) (ipfs.Folder, error) { return nil, nil }
func (m *MockIPFSClient) Pin(cid string, name string) error         { return nil }
func (m *MockIPFSClient) Unpin(cid string) error                    { return nil }
func (m *MockIPFSClient) Exists(cid string) (bool, error)           { return true, nil }
func (m *MockIPFSClient) IsPinned(ipns string) (bool, error)        { return true, nil }
func (m *MockIPFSClient) Ls(cid string) ([]string, error) {
	return []string{"mock-file1", "mock-file2"}, nil
}

func (m *MockIPFSClient) NodeStatus() (*ipfs.NodeStatus, error) {
	return &ipfs.NodeStatus{
		PeerID:         "mock-peer-id",
		Version:        "mock-version",
		PeerType:       "kubo",
		ConnectedPeers: 3,
	}, nil
}

func setupTestServer() *Server {
	config := &Config{
		JWTSecret:  []byte("test-secret"),
		IPFSClient: &MockIPFSClient{},
	}
	s := NewServer(config)

	// Setup routes manually for testing since we can't call Start() which would block
	e := s.Echo()

	// Setup middleware
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			return next(c)
		}
	})

	// Setup routes manually (mimicking server.setupRoutes)
	e.GET("/health", handlers.HealthCheckHandler)
	e.POST("/auth/login", handlers.LoginHandler(config.JWTSecret))

	return s
}

func TestHealthCheckHandler(t *testing.T) {
	s := setupTestServer()
	e := s.Echo()

	req, err := http.NewRequest("GET", "/health", nil)
	require.NoError(t, err)

	rr := httptest.NewRecorder()
	e.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

	var response map[string]string
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)
	// When health checker is not initialized, it returns "starting" status
	// This is expected behavior in test environment
	assert.Equal(t, "starting", response["status"])
}

func TestVaultHandlersCreation(t *testing.T) {
	// Test the vault handlers creation
	vaultHandlers := handlers.NewVaultHandlers(
		&MockIPFSClient{},
		handlers.NewConnectionManager(),
		handlers.NewSSEManager(),
	)
	assert.NotNil(t, vaultHandlers)
}
