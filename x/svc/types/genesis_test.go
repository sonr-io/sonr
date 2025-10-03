package types_test

import (
	"testing"

	"github.com/sonr-io/sonr/x/svc/types"

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
				Params:       types.DefaultParams(),
				Capabilities: []types.ServiceCapability{},
			},
			valid: true,
		},
		{
			desc: "valid genesis state with capabilities",
			genState: &types.GenesisState{
				Params: types.DefaultParams(),
				Capabilities: []types.ServiceCapability{
					{
						CapabilityId: "cap_1",
						ServiceId:    "service_1",
						Domain:       "example.com",
						Owner:        "cosmos1abc",
						Abilities:    []string{"read", "write"},
						CreatedAt:    1234567890,
						ExpiresAt:    1234567900,
						Revoked:      false,
					},
				},
			},
			valid: true,
		},
		{
			desc:     "empty params is invalid",
			genState: &types.GenesisState{},
			valid:    false,
		},
		{
			desc: "duplicate capability IDs is invalid",
			genState: &types.GenesisState{
				Params: types.DefaultParams(),
				Capabilities: []types.ServiceCapability{
					{
						CapabilityId: "cap_1",
						ServiceId:    "service_1",
						Domain:       "example.com",
						Owner:        "cosmos1abc",
						Abilities:    []string{"read"},
						CreatedAt:    1234567890,
						ExpiresAt:    1234567900,
						Revoked:      false,
					},
					{
						CapabilityId: "cap_1", // duplicate
						ServiceId:    "service_2",
						Domain:       "example.org",
						Owner:        "cosmos1xyz",
						Abilities:    []string{"write"},
						CreatedAt:    1234567891,
						ExpiresAt:    1234567901,
						Revoked:      false,
					},
				},
			},
			valid: false,
		},
		{
			desc: "capability with empty ID is invalid",
			genState: &types.GenesisState{
				Params: types.DefaultParams(),
				Capabilities: []types.ServiceCapability{
					{
						CapabilityId: "", // empty
						ServiceId:    "service_1",
						Domain:       "example.com",
						Owner:        "cosmos1abc",
						Abilities:    []string{"read"},
						CreatedAt:    1234567890,
						ExpiresAt:    1234567900,
						Revoked:      false,
					},
				},
			},
			valid: false,
		},
		{
			desc: "capability with empty service ID is invalid",
			genState: &types.GenesisState{
				Params: types.DefaultParams(),
				Capabilities: []types.ServiceCapability{
					{
						CapabilityId: "cap_1",
						ServiceId:    "", // empty
						Domain:       "example.com",
						Owner:        "cosmos1abc",
						Abilities:    []string{"read"},
						CreatedAt:    1234567890,
						ExpiresAt:    1234567900,
						Revoked:      false,
					},
				},
			},
			valid: false,
		},
		{
			desc: "capability with no abilities is invalid",
			genState: &types.GenesisState{
				Params: types.DefaultParams(),
				Capabilities: []types.ServiceCapability{
					{
						CapabilityId: "cap_1",
						ServiceId:    "service_1",
						Domain:       "example.com",
						Owner:        "cosmos1abc",
						Abilities:    []string{}, // empty
						CreatedAt:    1234567890,
						ExpiresAt:    1234567900,
						Revoked:      false,
					},
				},
			},
			valid: false,
		},
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
