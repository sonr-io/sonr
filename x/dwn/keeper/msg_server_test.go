package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sonr-io/snrd/x/dwn/types"
)

func TestParams(t *testing.T) {
	f := SetupTest(t)
	require := require.New(t)

	testCases := []struct {
		name    string
		request *types.MsgUpdateParams
		err     bool
	}{
		{
			name: "fail; invalid authority",
			request: &types.MsgUpdateParams{
				Authority: f.addrs[0].String(),
				Params:    types.DefaultParams(),
			},
			err: true,
		},
		{
			name: "success",
			request: &types.MsgUpdateParams{
				Authority: f.govModAddr,
				Params:    types.DefaultParams(),
			},
			err: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			_, err := f.msgServer.UpdateParams(f.ctx, tc.request)

			if tc.err {
				require.Error(err)
			} else {
				require.NoError(err)

				r, err := f.queryServer.Params(f.ctx, &types.QueryParamsRequest{})
				require.NoError(err)

				require.EqualValues(&tc.request.Params, r.Params)
			}

		})
	}
}
