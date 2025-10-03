package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sonr-io/sonr/x/dwn/types"
)

func TestQueryParams(t *testing.T) {
	f := SetupTest(t)

	resp, err := f.queryServer.Params(f.ctx, &types.QueryParamsRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, resp.Params)
	require.True(t, resp.Params.VaultCreationEnabled)
}

func TestQueryVaultNotFound(t *testing.T) {
	f := SetupTest(t)

	// Try to query non-existent vault
	_, err := f.queryServer.Vault(f.ctx, &types.QueryVaultRequest{
		VaultId: "non-existent-vault",
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "not found")
}

func TestQueryVaultsEmpty(t *testing.T) {
	f := SetupTest(t)

	// Query vaults for non-existent owner
	resp, err := f.queryServer.Vaults(f.ctx, &types.QueryVaultsRequest{
		Owner: "nonexistent-owner",
	})
	require.NoError(t, err)
	require.Empty(t, resp.Vaults)
}

func TestQueryCIDValidation(t *testing.T) {
	f := SetupTest(t)

	// Test empty CID
	resp, err := f.queryServer.CID(f.ctx, &types.QueryCIDRequest{
		Cid: "",
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, int32(400), resp.StatusCode) // Bad Request
	require.Nil(t, resp.Data)

	// Test invalid CID format
	resp, err = f.queryServer.CID(f.ctx, &types.QueryCIDRequest{
		Cid: "invalid-cid",
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, int32(400), resp.StatusCode) // Bad Request
	require.Nil(t, resp.Data)

	// Test valid CID format but non-existent content
	// This is a valid CIDv1 with SHA256 hash but content won't exist
	validCID := "bafybeigdyrzt5sfp7udm7hu76uh7y26nf3efuylqabf3oclgtqy55fbzdi"
	resp, err = f.queryServer.CID(f.ctx, &types.QueryCIDRequest{
		Cid: validCID,
	})

	// Response should either be 500 (IPFS client unavailable) or 404 (not found)
	// depending on whether IPFS is running in the test environment
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Contains(t, []int32{404, 500}, resp.StatusCode)
	require.Nil(t, resp.Data)
}
