package decorators

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	protov2 "google.golang.org/protobuf/proto"
)

// EmptyAnte is a no-op ante handler used for testing purposes.
// It simply returns the context without performing any operations.
var (
	EmptyAnte = func(ctx sdk.Context, tx sdk.Tx, simulate bool) (sdk.Context, error) {
		return ctx, nil
	}
)

// MockTx is a mock transaction implementation used for testing decorators.
// It implements the sdk.Tx interface with minimal functionality.
type MockTx struct {
	msgs []sdk.Msg
}

// NewMockTx creates a new mock transaction with the provided messages.
// This is useful for testing ante handler decorators in isolation.
func NewMockTx(msgs ...sdk.Msg) MockTx {
	return MockTx{
		msgs: msgs,
	}
}

// GetMsgs returns the messages contained in the mock transaction.
func (tx MockTx) GetMsgs() []sdk.Msg {
	return tx.msgs
}

// GetMsgsV2 implements the sdk.Tx interface. Returns nil as this is a mock.
func (tx MockTx) GetMsgsV2() ([]protov2.Message, error) {
	return nil, nil
}

// ValidateBasic implements the sdk.Tx interface. Always returns nil for the mock.
func (tx MockTx) ValidateBasic() error {
	return nil
}
