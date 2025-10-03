package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sonr-io/sonr/x/dwn/types"
)

func TestParams(t *testing.T) {
	f := SetupTest(t)

	// Test valid case only
	_, err := f.msgServer.UpdateParams(f.ctx, &types.MsgUpdateParams{
		Authority: f.govModAddr,
		Params:    types.DefaultParams(),
	})
	require.NoError(t, err)
}
