package ante

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sonr-io/sonr/crypto/ucan"
)

// TestUCANDecorator tests the UCAN decorator functionality
func TestUCANDecorator(t *testing.T) {
	// Test decorator creation
	decorator := NewUCANDecorator()
	assert.NotNil(t, decorator)
}

// TestUCANGaslessDecorator tests the gasless decorator
func TestUCANGaslessDecorator(t *testing.T) {
	// Create mock fee decorator
	mockFeeDecorator := &mockAnteDecorator{}

	// Test gasless decorator creation
	gaslessDecorator := NewUCANGaslessDecorator(mockFeeDecorator)
	assert.NotNil(t, gaslessDecorator)
}

// TestConditionalUCANDecorator tests conditional UCAN decorator
func TestConditionalUCANDecorator(t *testing.T) {
	// Create mock UCAN decorator
	mockUCANDecorator := &mockAnteDecorator{}

	// Test conditional decorator creation
	conditionalDecorator := NewConditionalUCANDecorator(mockUCANDecorator)
	assert.NotNil(t, conditionalDecorator)
}

// TestTokenExpiration tests UCAN token expiration validation
func TestTokenExpiration(t *testing.T) {
	decorator := NewUCANDecorator()
	ctx := sdk.Context{}.WithBlockTime(time.Now())

	// Test expired token
	expiredToken := &ucan.Token{
		ExpiresAt: time.Now().Unix() - 3600, // 1 hour ago
	}

	err := decorator.CheckTokenExpiration(ctx, expiredToken)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "UCAN token has expired")

	// Test valid token
	validToken := &ucan.Token{
		ExpiresAt: time.Now().Unix() + 3600, // 1 hour from now
	}

	err = decorator.CheckTokenExpiration(ctx, validToken)
	require.NoError(t, err)

	// Test token not yet valid
	futureToken := &ucan.Token{
		NotBefore: time.Now().Unix() + 3600, // 1 hour from now
	}

	err = decorator.CheckTokenExpiration(ctx, futureToken)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "UCAN token is not yet valid")
}

// TestValidateCapabilities tests capability validation
func TestValidateCapabilities(t *testing.T) {
	decorator := NewUCANDecorator()

	// Create token with single capability
	token := &ucan.Token{
		Attenuations: []ucan.Attenuation{
			{
				Capability: &ucan.SimpleCapability{
					Action: "did/update",
				},
			},
		},
	}

	// Test with matching capability
	err := decorator.ValidateCapabilities(token, []string{"did/update"})
	require.NoError(t, err)

	// Test with non-matching capability
	err = decorator.ValidateCapabilities(token, []string{"did/delete"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "UCAN token does not grant required capabilities")

	// Test with multiple capabilities
	multiToken := &ucan.Token{
		Attenuations: []ucan.Attenuation{
			{
				Capability: &ucan.MultiCapability{
					Actions: []string{"did/update", "did/create"},
				},
			},
		},
	}

	err = decorator.ValidateCapabilities(multiToken, []string{"did/update"})
	require.NoError(t, err)

	err = decorator.ValidateCapabilities(multiToken, []string{"did/create"})
	require.NoError(t, err)

	err = decorator.ValidateCapabilities(multiToken, []string{"did/delete"})
	require.Error(t, err)
}

// mockAnteDecorator is a helper for testing
type mockAnteDecorator struct{}

func (m *mockAnteDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	return next(ctx, tx, simulate)
}
