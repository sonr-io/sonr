// Package bridge provides the HTTP bridge server for the Highway service.
package bridge

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/hibiken/asynq"
	"github.com/sonr-io/sonr/bridge/server"
)

// HighwayService encapsulates the entire Highway service setup and lifecycle
type HighwayService struct {
	config       *Config
	client       *asynq.Client
	httpServer   *server.Server
	queueManager *QueueManager

	// Internal channels for coordination
	ctx    context.Context
	cancel context.CancelFunc
	sigCh  chan os.Signal
}

// NewHighwayService creates a new Highway service with all components initialized
func NewHighwayService() *HighwayService {
	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())

	// Handle shutdown signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	// Initialize configuration
	config := NewConfig()

	// Initialize Redis-based task queue client
	client := asynq.NewClient(asynq.RedisClientOpt{Addr: config.RedisAddr})
	log.Printf("Asynq client connected to Redis at %s", config.RedisAddr)

	// Create server configuration for bridge proxy
	serverConfig := &server.Config{
		HTTPAddr:   fmt.Sprintf(":%d", config.HTTPPort),
		JWTSecret:  config.JWTSecret,
		IPFSClient: config.IPFSClient,
	}

	// Create HTTP bridge server
	httpServer := server.NewServer(serverConfig)

	// Initialize UCAN task processing server with queue manager
	queueManager := NewQueueManager(config)

	return &HighwayService{
		config:       config,
		client:       client,
		httpServer:   httpServer,
		queueManager: queueManager,
		ctx:          ctx,
		cancel:       cancel,
		sigCh:        sigCh,
	}
}

// Start begins the Highway service with HTTP server and queue processing
func (hs *HighwayService) Start() error {
	log.Println("Starting Highway Service - UCAN-based MPC Task Processor")

	// Create and start HTTP bridge server in a goroutine
	go func() {
		log.Printf("Starting HTTP bridge server on port %d", hs.config.HTTPPort)
		if err := hs.httpServer.Start(hs.client); err != nil {
			log.Fatalf("HTTP bridge server failed: %v", err)
		}
	}()

	// Start queue server with graceful shutdown support
	go func() {
		if err := hs.queueManager.Run(); err != nil {
			log.Printf("Asynq server error: %v", err)
			hs.cancel()
		}
	}()

	// Wait for shutdown signal
	select {
	case <-hs.sigCh:
		log.Println("Received shutdown signal, initiating graceful shutdown...")
	case <-hs.ctx.Done():
		log.Println("Context cancelled, shutting down...")
	}

	return nil
}

// Shutdown gracefully stops all service components
func (hs *HighwayService) Shutdown() {
	log.Println("Shutting down Highway service...")

	// Close Asynq client
	if hs.client != nil {
		hs.client.Close()
	}

	// Shutdown queue manager
	if hs.queueManager != nil {
		hs.queueManager.Shutdown()
	}

	// Cancel context to signal shutdown to all components
	hs.cancel()

	log.Println("Highway service stopped")
}
