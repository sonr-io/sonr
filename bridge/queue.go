package bridge

import (
	"log"

	"github.com/hibiken/asynq"
	"github.com/sonr-io/sonr/bridge/tasks"
)

// QueueManager handles Asynq server setup and task registration
type QueueManager struct {
	server *asynq.Server
	mux    *asynq.ServeMux
	config *Config
}

// NewQueueManager creates a new queue manager with the given configuration
func NewQueueManager(config *Config) *QueueManager {
	// Initialize UCAN task processing server with optimized queue configuration
	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: config.RedisAddr},
		config.AsynqConfig,
	)

	// Register UCAN-based task handlers
	mux := asynq.NewServeMux()
	registerTaskHandlers(mux)

	return &QueueManager{
		server: srv,
		mux:    mux,
		config: config,
	}
}

// registerTaskHandlers registers all UCAN-based task handlers
func registerTaskHandlers(mux *asynq.ServeMux) {
	// Core UCAN token operations
	mux.Handle(tasks.TypeUCANToken, tasks.NewUCANProcessor())
	mux.Handle(tasks.TypeUCANAttenuation, tasks.NewUCANAttenuationProcessor())

	// MPC-based cryptographic operations
	mux.Handle(tasks.TypeUCANSign, tasks.NewUCANSignProcessor())
	mux.Handle(tasks.TypeUCANVerify, tasks.NewUCANVerifyProcessor())

	// DID operations (replaces both TypeVaultGenerate and TypeVaultRefresh)
	// Serves as proxy for future x/did module integration
	mux.Handle(tasks.TypeUCANDIDGeneration, tasks.NewUCANDIDProcessor())

	log.Printf("UCAN task handlers registered successfully")
}

// Run starts the Asynq server with the registered task handlers
func (qm *QueueManager) Run() error {
	log.Printf("Starting Asynq task server with Redis at %s", qm.config.RedisAddr)
	return qm.server.Run(qm.mux)
}

// Shutdown gracefully shuts down the Asynq server
func (qm *QueueManager) Shutdown() {
	qm.server.Shutdown()
}
