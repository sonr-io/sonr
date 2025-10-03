// Package tasks provides UCAN-based task processing for token attenuation operations.
package tasks

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/asynkron/protoactor-go/actor"
	"github.com/hibiken/asynq"
	"github.com/sonr-io/sonr/x/dwn/client/plugin"
)

// ╭─────────────────────────────────────────────────────────╮
// │                      Processor 					               │
// ╰─────────────────────────────────────────────────────────╯

// UCANAttenuationProcessor implements asynq.Handler interface for UCAN attenuation operations.
type UCANAttenuationProcessor struct {
	asynq.Handler
	pid *actor.PID
}

// NewUCANAttenuationProcessor creates a new UCANAttenuationProcessor for the specified actor system.
func NewUCANAttenuationProcessor() *UCANAttenuationProcessor {
	pid := system.Root.SpawnPrefix(plugin.Props(), TypeUCANAttenuation)
	return &UCANAttenuationProcessor{
		pid: pid,
	}
}

// ╭─────────────────────────────────────────────────────────╮
// │                      Payload 					                 │
// ╰─────────────────────────────────────────────────────────╯

// UCANAttenuationPayload contains parameters for UCAN token attenuation task.
type UCANAttenuationPayload struct {
	UserID       int              `json:"user_id"`
	ParentToken  string           `json:"parent_token"`
	AudienceDID  string           `json:"audience_did"`
	Attenuations []map[string]any `json:"attenuations,omitempty"`
	ExpiresAt    int64            `json:"expires_at,omitempty"`
}

// NewUCANAttenuationTask creates a new UCAN token attenuation task.
func NewUCANAttenuationTask(
	userID int,
	parentToken string,
	audienceDID string,
	attenuations []map[string]any,
	expiresAt int64,
) (*asynq.Task, error) {
	payload, err := json.Marshal(UCANAttenuationPayload{
		UserID:       userID,
		ParentToken:  parentToken,
		AudienceDID:  audienceDID,
		Attenuations: attenuations,
		ExpiresAt:    expiresAt,
	})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeUCANAttenuation, payload), nil
}

// ╭───────────────────────────────────────────────────────╮
// │                      Handler 					               │
// ╰───────────────────────────────────────────────────────╯

// ProcessTask processes the UCAN token attenuation task.
func (processor *UCANAttenuationProcessor) ProcessTask(ctx context.Context, t *asynq.Task) error {
	var p UCANAttenuationPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

	// Create attenuated token request for the actor
	request := &plugin.NewAttenuatedTokenRequest{
		ParentToken:  p.ParentToken,
		AudienceDID:  p.AudienceDID,
		Attenuations: p.Attenuations,
		ExpiresAt:    p.ExpiresAt,
	}

	resp, err := system.Root.RequestFuture(processor.pid, request, KRequestTimeout).Result()
	if err != nil {
		return err
	}
	switch resp := resp.(type) {
	case *plugin.UCANTokenResponse:
		if resp.Error != "" {
			return fmt.Errorf("UCAN token attenuation failed: %s", resp.Error)
		}
		return nil
	default:
		return fmt.Errorf("invalid response type: %T", resp)
	}
}

// ╭─────────────────────────────────────────────────────────╮
// │                   DID Generation Processor              │
// ╰─────────────────────────────────────────────────────────╯

// UCANDIDProcessor implements asynq.Handler interface for DID generation operations.
type UCANDIDProcessor struct {
	asynq.Handler
	pid *actor.PID
}

// NewUCANDIDProcessor creates a new UCANDIDProcessor for the specified actor system.
func NewUCANDIDProcessor() *UCANDIDProcessor {
	pid := system.Root.SpawnPrefix(plugin.Props(), TypeUCANDIDGeneration)
	return &UCANDIDProcessor{
		pid: pid,
	}
}

// UCANDIDPayload contains parameters for DID generation task.
type UCANDIDPayload struct {
	UserID int `json:"user_id"`
}

// NewUCANDIDTask creates a new DID generation task.
func NewUCANDIDTask(userID int) (*asynq.Task, error) {
	payload, err := json.Marshal(UCANDIDPayload{
		UserID: userID,
	})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeUCANDIDGeneration, payload), nil
}

// ProcessTask processes the DID generation task.
func (processor *UCANDIDProcessor) ProcessTask(ctx context.Context, t *asynq.Task) error {
	var p UCANDIDPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

	// Request DID generation from the actor
	resp, err := system.Root.RequestFuture(processor.pid, &plugin.GetIssuerDIDResponse{}, KRequestTimeout).
		Result()
	if err != nil {
		return err
	}
	switch resp := resp.(type) {
	case *plugin.GetIssuerDIDResponse:
		if resp.Error != "" {
			return fmt.Errorf("DID generation failed: %s", resp.Error)
		}
		return nil
	default:
		return fmt.Errorf("invalid response type: %T", resp)
	}
}
