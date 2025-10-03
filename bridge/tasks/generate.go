// Package tasks provides UCAN-based task processing for the refactored Motor plugin.
// This package handles asynchronous UCAN token creation, signing operations,
// and DID management tasks using the MPC-based plugin architecture.
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

// UCANProcessor implements asynq.Handler interface for UCAN operations.
type UCANProcessor struct {
	asynq.Handler
	pid *actor.PID
}

// NewUCANProcessor creates a new UCANProcessor for the specified actor system.
func NewUCANProcessor() *UCANProcessor {
	pid := system.Root.SpawnPrefix(plugin.Props(), TypeUCANToken)
	return &UCANProcessor{
		pid: pid,
	}
}

// ╭─────────────────────────────────────────────────────────╮
// │                      Payload 					                 │
// ╰─────────────────────────────────────────────────────────╯

// UCANTokenPayload contains parameters for UCAN token creation task.
type UCANTokenPayload struct {
	UserID       int              `json:"user_id"`
	AudienceDID  string           `json:"audience_did"`
	Attenuations []map[string]any `json:"attenuations,omitempty"`
	ExpiresAt    int64            `json:"expires_at,omitempty"`
}

// NewUCANTokenTask creates a new UCAN token creation task.
func NewUCANTokenTask(
	userID int,
	audienceDID string,
	attenuations []map[string]any,
	expiresAt int64,
) (*asynq.Task, error) {
	payload, err := json.Marshal(UCANTokenPayload{
		UserID:       userID,
		AudienceDID:  audienceDID,
		Attenuations: attenuations,
		ExpiresAt:    expiresAt,
	})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeUCANToken, payload), nil
}

// ╭───────────────────────────────────────────────────────╮
// │                      Handler 					               │
// ╰───────────────────────────────────────────────────────╯

// ProcessTask processes the UCAN token creation task.
func (processor *UCANProcessor) ProcessTask(ctx context.Context, t *asynq.Task) error {
	var p UCANTokenPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

	// Create UCAN token request for the actor
	request := &plugin.NewOriginTokenRequest{
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
			return fmt.Errorf("UCAN token creation failed: %s", resp.Error)
		}
		return nil
	default:
		return fmt.Errorf("invalid response type: %T", resp)
	}
}
