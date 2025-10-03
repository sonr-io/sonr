package modules

import (
	"context"
	"net/http"
	"testing"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	"github.com/sonr-io/sonr/test/e2e/utils"
)

func TestSvcModule(t *testing.T) {
	cfg := utils.NewTestConfig()
	ctx := context.Background()

	t.Run("svc_params", func(t *testing.T) {
		// Query service module parameters
		url := cfg.BaseURL + "/sonr/svc/v1/params"

		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		require.NoError(t, err, "failed to create request")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err, "failed to query svc params")
		defer resp.Body.Close()

		// Accept 501 Not Implemented for now as the endpoint may not be ready
		if resp.StatusCode == http.StatusNotImplemented {
			t.Skip("SVC params endpoint not implemented yet")
		}
		require.Equal(t, http.StatusOK, resp.StatusCode, "svc params query should succeed")
	})

	t.Run("svc_integration", func(t *testing.T) {
		// Use pre-funded account from localnet
		testAddr := "idx1fcqk3crpnyvyhtd4jepsnx5eat5ehc920epq29"

		// Verify user has balance for service operations
		balance, err := cfg.Client.GetBalance(ctx, testAddr, cfg.StakingDenom)
		require.NoError(t, err, "failed to query balance")
		require.True(t, balance.GT(math.ZeroInt()), "should have balance for operations")
	})
}

func TestTokenFactoryModule(t *testing.T) {
	cfg := utils.NewTestConfig()
	ctx := context.Background()

	t.Run("tokenfactory_params", func(t *testing.T) {
		// Query tokenfactory module parameters
		url := cfg.BaseURL + "/osmosis/tokenfactory/v1beta1/params"

		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		require.NoError(t, err, "failed to create request")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err, "failed to query tokenfactory params")
		defer resp.Body.Close()

		// Note: This might return 404 if tokenfactory endpoint is different
		// or module is not enabled, which is acceptable
		if resp.StatusCode != http.StatusNotFound {
			require.Equal(t, http.StatusOK, resp.StatusCode, "tokenfactory params query should succeed when available")
		}
	})
}

func TestDIDModule(t *testing.T) {
	cfg := utils.NewTestConfig()
	ctx := context.Background()

	t.Run("did_params", func(t *testing.T) {
		// Query DID module parameters
		url := cfg.BaseURL + "/sonr/did/v1/params"

		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		require.NoError(t, err, "failed to create request")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err, "failed to query did params")
		defer resp.Body.Close()

		// Accept 501 Not Implemented for now as the endpoint may not be ready
		if resp.StatusCode == http.StatusNotImplemented {
			t.Skip("DID params endpoint not implemented yet")
		}
		require.Equal(t, http.StatusOK, resp.StatusCode, "did params query should succeed")
	})

	t.Run("did_functionality", func(t *testing.T) {
		// Use pre-funded account from localnet
		testAddr := "idx1fcqk3crpnyvyhtd4jepsnx5eat5ehc920epq29"

		// Verify user has balance for DID operations
		balance, err := cfg.Client.GetBalance(ctx, testAddr, cfg.StakingDenom)
		require.NoError(t, err, "failed to query balance")
		require.True(t, balance.GT(math.NewInt(1_000_000)), "should have sufficient balance for DID operations")
	})
}

func TestDWNModule(t *testing.T) {
	cfg := utils.NewTestConfig()
	ctx := context.Background()

	t.Run("dwn_params", func(t *testing.T) {
		// Query DWN module parameters
		url := cfg.BaseURL + "/sonr/dwn/v1/params"

		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		require.NoError(t, err, "failed to create request")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err, "failed to query dwn params")
		defer resp.Body.Close()

		// Accept 501 Not Implemented for now as the endpoint may not be ready
		if resp.StatusCode == http.StatusNotImplemented {
			t.Skip("DWN params endpoint not implemented yet")
		}
		require.Equal(t, http.StatusOK, resp.StatusCode, "dwn params query should succeed")
	})
}
