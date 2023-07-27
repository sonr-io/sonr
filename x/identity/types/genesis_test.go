package types_test

import (
	"testing"

	"github.com/sonrhq/core/x/identity/types"
	"github.com/stretchr/testify/require"
)

func TestGenesisState_Validate(t *testing.T) {
	tests := []struct {
		desc     string
		genState *types.GenesisState
		valid    bool
	}{
		{
			desc:     "default is valid",
			genState: types.DefaultGenesis(),
			valid:    true,
		},
		{
			desc: "valid genesis state",
			genState: &types.GenesisState{
				PortId: types.PortID,
				ControllerAccountList: []types.ControllerAccount{
					{
						Id: 0,
					},
					{
						Id: 1,
					},
				},
				ControllerAccountCount: 2,
				EscrowAccountList: []types.EscrowAccount{
					{
						Id: 0,
					},
					{
						Id: 1,
					},
				},
				EscrowAccountCount: 2,
				// this line is used by starport scaffolding # types/genesis/validField
			},
			valid: true,
		},
		{
			desc: "duplicated controllerAccount",
			genState: &types.GenesisState{
				ControllerAccountList: []types.ControllerAccount{
					{
						Id: 0,
					},
					{
						Id: 0,
					},
				},
			},
			valid: false,
		},
		{
			desc: "invalid controllerAccount count",
			genState: &types.GenesisState{
				ControllerAccountList: []types.ControllerAccount{
					{
						Id: 1,
					},
				},
				ControllerAccountCount: 0,
			},
			valid: false,
		},
		{
			desc: "duplicated escrowAccount",
			genState: &types.GenesisState{
				EscrowAccountList: []types.EscrowAccount{
					{
						Id: 0,
					},
					{
						Id: 0,
					},
				},
			},
			valid: false,
		},
		{
			desc: "invalid escrowAccount count",
			genState: &types.GenesisState{
				EscrowAccountList: []types.EscrowAccount{
					{
						Id: 1,
					},
				},
				EscrowAccountCount: 0,
			},
			valid: false,
		},
		// this line is used by starport scaffolding # types/genesis/testcase
	}
	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			err := tc.genState.Validate()
			if tc.valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}