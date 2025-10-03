// Package server provides the HTTP server for the highway server
package server

import (
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/hibiken/asynq"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sonr-io/sonr/bridge/handlers"
	"github.com/sonr-io/sonr/types/ipfs"
)

const (
	DefaultHTTPAddr = ":8080"
)

// Config holds server configuration
type Config struct {
	HTTPAddr   string
	JWTSecret  []byte
	IPFSClient ipfs.IPFSClient
}

// Server represents the HTTP server
type Server struct {
	config            *Config
	echo              *echo.Echo
	upgrader          *websocket.Upgrader
	connectionManager *handlers.ConnectionManager
	sseManager        *handlers.SSEManager
	vaultHandlers     *handlers.VaultHandlers
}

// NewServer creates a new server instance
func NewServer(config *Config) *Server {
	// WebSocket upgrader with CORS settings
	upgrader := &websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow all origins in development
		},
	}

	// Create connection managers
	connectionManager := handlers.NewConnectionManager()
	sseManager := handlers.NewSSEManager()

	// Create vault handlers
	vaultHandlers := handlers.NewVaultHandlers(config.IPFSClient, connectionManager, sseManager)

	return &Server{
		config:            config,
		echo:              echo.New(),
		upgrader:          upgrader,
		connectionManager: connectionManager,
		sseManager:        sseManager,
		vaultHandlers:     vaultHandlers,
	}
}

// Echo returns the underlying Echo instance for testing
func (s *Server) Echo() *echo.Echo {
	return s.echo
}

// Start starts the HTTP server
func (s *Server) Start(client *asynq.Client) error {
	s.setupMiddleware()
	s.setupRoutes(client)

	addr := s.config.HTTPAddr
	if addr == "" {
		addr = DefaultHTTPAddr
	}
	return s.echo.Start(addr)
}

// setupMiddleware configures Echo middleware
func (s *Server) setupMiddleware() {
	s.echo.Use(middleware.Logger())
	s.echo.Use(middleware.Recover())
	s.echo.Use(middleware.CORS())
}

// setupRoutes configures all routes
func (s *Server) setupRoutes(client *asynq.Client) {
	// Initialize health checker
	handlers.InitHealthChecker(client, s.config.IPFSClient)

	// Public endpoints (no authentication required)
	s.echo.GET("/health", handlers.HealthCheckHandler) // Liveness probe
	s.echo.GET("/ready", handlers.ReadinessHandler)    // Readiness probe
	s.echo.POST("/auth/login", handlers.LoginHandler(s.config.JWTSecret))

	// JWT middleware configuration
	jwtConfig := echojwt.Config{
		SigningKey:    s.config.JWTSecret,
		SigningMethod: "HS256",
	}

	// Protected vault endpoints group with JWT middleware
	vault := s.echo.Group("/vault")
	vault.Use(echojwt.WithConfig(jwtConfig))
	vault.POST("/generate", s.vaultHandlers.GenerateHandler(client))
	vault.POST("/sign", s.vaultHandlers.SignHandler(client))
	vault.POST("/verify", s.vaultHandlers.VerifyHandler(client))
	vault.POST("/export", s.vaultHandlers.ExportHandler(client))
	vault.POST("/import", s.vaultHandlers.ImportHandler(client))
	vault.POST("/refresh", s.vaultHandlers.RefreshHandler(client))

	// WebSocket endpoint for real-time task status updates
	vault.GET("/ws/:task_id", handlers.WebSocketHandler(s.upgrader, s.connectionManager))

	// Server-Sent Events endpoint for task progress streaming
	vault.GET("/events/:task_id", handlers.SSEHandler(s.sseManager))
}
