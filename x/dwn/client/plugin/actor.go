package plugin

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/asynkron/protoactor-go/actor"
)

// Actor represents the UCAN actor that handles MPC-based cryptographic operations.
// It maintains a WebAssembly plugin for secure UCAN token operations and
// uses behavioral state management to handle different actor lifecycle phases.
type Actor struct {
	ctx      context.Context // Context for managing plugin lifecycle
	behavior actor.Behavior  // Behavioral state manager for handling different phases
	enclave  Plugin          // MPC enclave plugin instance
	config   *EnclaveConfig  // Configuration for enclave initialization (optional)
}

// NewActor creates a new UCAN actor instance.
// The actor starts in an uninitialized state and transitions to initialized
// once the MPC enclave plugin is successfully loaded.
func NewActor() actor.Actor {
	return &Actor{
		ctx:      context.Background(),
		behavior: actor.NewBehavior(),
	}
}

// Props returns the actor properties configuration for creating new UCAN actors.
// It uses NewActor as the producer function to create UCAN actor instances.
func Props() *actor.Props {
	return actor.PropsFromProducer(NewActor)
}

// PropsWithConfig returns actor properties configured with specific enclave data.
// This allows creating actors with pre-configured enclave data for testing or specific use cases.
func PropsWithConfig(config *EnclaveConfig) *actor.Props {
	return actor.PropsFromProducer(func() actor.Actor {
		return &Actor{
			ctx:      context.Background(),
			behavior: actor.NewBehavior(),
			config:   config, // Store config for initialization
		}
	})
}

// Receive is the main message handler for the UCAN actor.
// It handles actor lifecycle messages and delegates other messages to the current behavior.
// The actor uses behavioral patterns to handle different states (uninitialized vs initialized).
func (a *Actor) Receive(c actor.Context) {
	switch c.Message().(type) {
	case *actor.Started:
		a.handleStarted(c)
	case *actor.Stopping:
		a.handleStopping(c)
	default:
		a.behavior.Receive(c)
	}
}

// Initialized handles UCAN operation messages once the actor is fully initialized.
// This method is set as the behavior after the MPC enclave plugin is successfully loaded.
// It processes NewOriginToken, NewAttenuatedToken, SignData, VerifyData, and GetIssuerDID requests.
func (a *Actor) Initialized(c actor.Context) {
	switch msg := c.Message().(type) {
	case *NewOriginTokenRequest:
		a.handleNewOriginToken(c, msg)
	case *NewAttenuatedTokenRequest:
		a.handleNewAttenuatedToken(c, msg)
	case *SignDataRequest:
		a.handleSignData(c, msg)
	case *VerifyDataRequest:
		a.handleVerifyData(c, msg)
	case *GetIssuerDIDResponse: // Used as request for DID retrieval
		a.handleGetIssuerDID(c)
	}
}

// handleStarted initializes the MPC enclave plugin when the actor starts.
// It loads the WebAssembly plugin and transitions the actor to the initialized state.
// If plugin loading fails, the actor remains in an uninitialized state.
func (a *Actor) handleStarted(c actor.Context) {
	a.ctx = context.Background()

	// Use provided config if available, otherwise use default
	var config *EnclaveConfig
	if a.config != nil {
		config = a.config
	} else {
		config = DefaultEnclaveConfig()
	}

	// For testing, we need to provide mock enclave data
	// In production, this would be provided by the caller
	if config.EnclaveData == nil {
		c.Logger().Warn("No enclave data provided, actor will not initialize",
			slog.String("config", fmt.Sprintf("%+v", config)),
		)
		return
	}

	c.Logger().Info("Attempting to load MPC enclave plugin",
		slog.String("config", fmt.Sprintf("%+v", config)),
	)

	e, err := LoadPluginWithManager(a.ctx, config)
	if err != nil {
		c.Logger().Error("Failed to create MPC enclave host",
			slog.String("error", err.Error()),
			slog.String("config", fmt.Sprintf("%+v", config)),
		)
		return
	}
	a.enclave = e
	c.Logger().Info("MPC enclave actor started successfully",
		slog.String("config", fmt.Sprintf("%+v", config)),
	)
	a.behavior.Become(a.Initialized)
}

// handleStopping performs cleanup when the actor is stopping.
// It releases the MPC enclave plugin resources and cleans up the context.
func (a *Actor) handleStopping(c actor.Context) {
	a.ctx.Done()
	a.enclave = nil
	c.Logger().Info("MPC enclave plugin done")
}

// handleNewOriginToken processes UCAN origin token creation requests by delegating to the MPC enclave plugin.
// It validates the request and responds with the generated UCAN token or an error.
func (a *Actor) handleNewOriginToken(context actor.Context, msg *NewOriginTokenRequest) {
	resp, err := a.enclave.NewOriginToken(msg)
	if err != nil {
		context.Logger().Error("failed to create origin token", slog.String("error", err.Error()))
		context.Respond(err)
		return
	}
	context.Respond(resp)
}

// handleNewAttenuatedToken processes UCAN attenuated token creation requests by delegating to the MPC enclave plugin.
// It creates a delegated token with reduced permissions and responds with the token or an error.
func (a *Actor) handleNewAttenuatedToken(context actor.Context, msg *NewAttenuatedTokenRequest) {
	resp, err := a.enclave.NewAttenuatedToken(msg)
	if err != nil {
		context.Logger().
			Error("failed to create attenuated token", slog.String("error", err.Error()))
		context.Respond(err)
		return
	}
	context.Respond(resp)
}

// handleSignData processes data signing requests by delegating to the MPC enclave plugin.
// It creates a cryptographic signature using MPC and responds with the signature or an error.
func (a *Actor) handleSignData(context actor.Context, msg *SignDataRequest) {
	resp, err := a.enclave.SignData(msg)
	if err != nil {
		context.Logger().Error("failed to sign data", slog.String("error", err.Error()))
		context.Respond(err)
		return
	}
	context.Respond(resp)
}

// handleVerifyData processes signature verification requests by delegating to the MPC enclave plugin.
// It validates the signature against the data and responds with the verification result or an error.
func (a *Actor) handleVerifyData(context actor.Context, msg *VerifyDataRequest) {
	resp, err := a.enclave.VerifyData(msg)
	if err != nil {
		context.Logger().Error("failed to verify data", slog.String("error", err.Error()))
		context.Respond(err)
		return
	}
	context.Respond(resp)
}

// handleGetIssuerDID processes DID retrieval requests by delegating to the MPC enclave plugin.
// It retrieves the issuer DID, address, and chain code from the enclave.
func (a *Actor) handleGetIssuerDID(context actor.Context) {
	resp, err := a.enclave.GetIssuerDID()
	if err != nil {
		context.Logger().Error("failed to get issuer DID", slog.String("error", err.Error()))
		context.Respond(err)
		return
	}
	context.Respond(resp)
}
