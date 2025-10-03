// Package tasks provides UCAN-based task processing for MPC signing operations.
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

// UCANSignProcessor implements asynq.Handler interface for UCAN signing operations.
type UCANSignProcessor struct {
	asynq.Handler
	pid *actor.PID
}

// NewUCANSignProcessor creates a new UCANSignProcessor for the specified actor system.
func NewUCANSignProcessor() *UCANSignProcessor {
	pid := system.Root.SpawnPrefix(plugin.Props(), TypeUCANSign)
	return &UCANSignProcessor{
		pid: pid,
	}
}

// ╭─────────────────────────────────────────────────────────╮
// │                      Payload 					                 │
// ╰─────────────────────────────────────────────────────────╯

// UCANSignPayload contains parameters for UCAN signing task.
type UCANSignPayload struct {
	UserID int    `json:"user_id"`
	Data   []byte `json:"data"`
}

// NewUCANSignTask creates a new UCAN signing task.
func NewUCANSignTask(userID int, data []byte) (*asynq.Task, error) {
	payload, err := json.Marshal(UCANSignPayload{
		UserID: userID,
		Data:   data,
	})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeUCANSign, payload), nil
}

// ╭───────────────────────────────────────────────────────╮
// │                      Handler 					               │
// ╰───────────────────────────────────────────────────────╯

// ProcessTask processes the UCAN signing task.
func (processor *UCANSignProcessor) ProcessTask(ctx context.Context, t *asynq.Task) error {
	var p UCANSignPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

	// Create signing request for the actor
	request := &plugin.SignDataRequest{
		Data: p.Data,
	}

	resp, err := system.Root.RequestFuture(processor.pid, request, KRequestTimeout).Result()
	if err != nil {
		return err
	}
	switch resp := resp.(type) {
	case *plugin.SignDataResponse:
		if resp.Error != "" {
			return fmt.Errorf("UCAN signing failed: %s", resp.Error)
		}
		return nil
	default:
		return fmt.Errorf("invalid response type: %T", resp)
	}
}

// ╭─────────────────────────────────────────────────────────╮
// │                   Verification Processor                │
// ╰─────────────────────────────────────────────────────────╯

// UCANVerifyProcessor implements asynq.Handler interface for UCAN verification operations.
type UCANVerifyProcessor struct {
	asynq.Handler
	pid *actor.PID
}

// NewUCANVerifyProcessor creates a new UCANVerifyProcessor for the specified actor system.
func NewUCANVerifyProcessor() *UCANVerifyProcessor {
	pid := system.Root.SpawnPrefix(plugin.Props(), TypeUCANVerify)
	return &UCANVerifyProcessor{
		pid: pid,
	}
}

// UCANVerifyPayload contains parameters for UCAN verification task.
type UCANVerifyPayload struct {
	UserID    int    `json:"user_id"`
	Data      []byte `json:"data"`
	Signature []byte `json:"signature"`
}

// NewUCANVerifyTask creates a new UCAN verification task.
func NewUCANVerifyTask(userID int, data, signature []byte) (*asynq.Task, error) {
	payload, err := json.Marshal(UCANVerifyPayload{
		UserID:    userID,
		Data:      data,
		Signature: signature,
	})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeUCANVerify, payload), nil
}

// ProcessTask processes the UCAN verification task.
func (processor *UCANVerifyProcessor) ProcessTask(ctx context.Context, t *asynq.Task) error {
	var p UCANVerifyPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

	// Create verification request for the actor
	request := &plugin.VerifyDataRequest{
		Data:      p.Data,
		Signature: p.Signature,
	}

	resp, err := system.Root.RequestFuture(processor.pid, request, KRequestTimeout).Result()
	if err != nil {
		return err
	}
	switch resp := resp.(type) {
	case *plugin.VerifyDataResponse:
		if resp.Error != "" {
			return fmt.Errorf("UCAN verification failed: %s", resp.Error)
		}
		return nil
	default:
		return fmt.Errorf("invalid response type: %T", resp)
	}
}
