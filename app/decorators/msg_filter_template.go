// Package decorators provides custom ante handler decorators for transaction processing
// in the Sonr blockchain application.
package decorators

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	"github.com/cosmos/gogoproto/proto"
)

// MsgFilterDecorator is an ante handler decorator that filters out specific message types
// from transactions. It prevents certain message types from being processed by the chain.
type MsgFilterDecorator struct {
	blockedTypes []sdk.Msg
}

// FilterDecorator returns a new MsgFilterDecorator. This errors if the transaction
// contains any of the blocked message types.
//
// Example:
//   - decorators.FilterDecorator(&banktypes.MsgSend{})
//
// This would block any MsgSend messages from being included in a transaction if set in ante.go
func FilterDecorator(blockedMsgTypes ...sdk.Msg) MsgFilterDecorator {
	return MsgFilterDecorator{
		blockedTypes: blockedMsgTypes,
	}
}

// AnteHandle implements the AnteDecorator interface. It checks if the transaction
// contains any disallowed message types and rejects it if found.
func (mfd MsgFilterDecorator) AnteHandle(
	ctx sdk.Context,
	tx sdk.Tx,
	simulate bool,
	next sdk.AnteHandler,
) (newCtx sdk.Context, err error) {
	if mfd.HasDisallowedMessage(ctx, tx.GetMsgs()) {
		currHeight := ctx.BlockHeight()
		return ctx, fmt.Errorf("tx contains unsupported message types at height %d", currHeight)
	}

	return next(ctx, tx, simulate)
}

// HasDisallowedMessage recursively checks if any of the provided messages or their
// nested messages (in case of authz.MsgExec) match the blocked message types.
// Returns true if a disallowed message is found.
func (mfd MsgFilterDecorator) HasDisallowedMessage(ctx sdk.Context, msgs []sdk.Msg) bool {
	for _, msg := range msgs {
		// check nested messages in a recursive manner
		if execMsg, ok := msg.(*authz.MsgExec); ok {
			msgs, err := execMsg.GetMessages()
			if err != nil {
				return true
			}

			if mfd.HasDisallowedMessage(ctx, msgs) {
				return true
			}
		}

		for _, blockedType := range mfd.blockedTypes {
			if proto.MessageName(msg) == proto.MessageName(blockedType) {
				return true
			}
		}
	}

	return false
}
