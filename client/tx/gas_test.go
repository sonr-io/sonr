package tx

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/sonr-io/sonr/client/config"
)

// GasEstimatorTestSuite tests the gas estimator.
type GasEstimatorTestSuite struct {
	suite.Suite
	estimator GasEstimator
	config    *config.NetworkConfig
}

func (suite *GasEstimatorTestSuite) SetupTest() {
	cfg := config.LocalNetwork()
	suite.config = &cfg

	// Create mock gRPC connection for testing
	conn, err := grpc.Dial("localhost:9090", grpc.WithInsecure())
	suite.Require().NoError(err)

	suite.estimator = NewGasEstimator(conn, suite.config)
}

func (suite *GasEstimatorTestSuite) TestCalculateFee() {
	gasUsed := uint64(100000)
	gasPrice := 0.025
	denom := "usnr"

	fee := suite.estimator.CalculateFee(gasUsed, gasPrice, denom)

	suite.Require().NotNil(fee)
	suite.Require().Len(fee, 1)
	suite.Require().Equal(denom, fee[0].Denom)
	suite.Require().Equal(int64(2500), fee[0].Amount.Int64())
}

func (suite *GasEstimatorTestSuite) TestCalculateFeeWithAdjustment() {
	gasUsed := uint64(100000)
	gasPrice := 0.025
	adjustment := 1.5
	denom := "usnr"

	fee := suite.estimator.CalculateFeeWithAdjustment(gasUsed, gasPrice, adjustment, denom)

	suite.Require().NotNil(fee)
	suite.Require().Len(fee, 1)
	suite.Require().Equal(denom, fee[0].Denom)
	suite.Require().Equal(int64(3750), fee[0].Amount.Int64())
}

func (suite *GasEstimatorTestSuite) TestWithGasAdjustment() {
	adjustment := 2.0
	updated := suite.estimator.WithGasAdjustment(adjustment)

	suite.Require().NotNil(updated)
	// Verify adjustment was applied
	ge := updated.(*gasEstimator)
	suite.Require().Equal(adjustment, ge.gasConfig.Adjustment)
}

func (suite *GasEstimatorTestSuite) TestWithMinGasPrice() {
	price := 0.05
	updated := suite.estimator.WithMinGasPrice(price)

	suite.Require().NotNil(updated)
	// Verify price was applied
	ge := updated.(*gasEstimator)
	suite.Require().Equal(price, ge.gasConfig.MinGasPrice)
}

func (suite *GasEstimatorTestSuite) TestWithMaxGasLimit() {
	limit := uint64(5000000)
	updated := suite.estimator.WithMaxGasLimit(limit)

	suite.Require().NotNil(updated)
	// Verify limit was applied
	ge := updated.(*gasEstimator)
	suite.Require().Equal(limit, ge.gasConfig.MaxGasLimit)
}

func (suite *GasEstimatorTestSuite) TestGetRecommendedGasPrice() {
	price, err := suite.estimator.GetRecommendedGasPrice(context.Background())

	suite.Require().NoError(err)
	suite.Require().Greater(price, 0.0)
}

func (suite *GasEstimatorTestSuite) TestGetNetworkGasInfo() {
	info, err := suite.estimator.GetNetworkGasInfo(context.Background())

	suite.Require().NoError(err)
	suite.Require().NotNil(info)
	suite.Require().Greater(info.MinGasPrice, 0.0)
	suite.Require().Greater(info.MaxGasLimit, uint64(0))
}

func TestGasEstimatorTestSuite(t *testing.T) {
	suite.Run(t, new(GasEstimatorTestSuite))
}

// TestGasLimitForMessageType tests gas limit recommendations.
func TestGasLimitForMessageType(t *testing.T) {
	tests := []struct {
		msgType     string
		expectedGas uint64
	}{
		{
			msgType:     "/cosmos.bank.v1beta1.MsgSend",
			expectedGas: SendGasLimit,
		},
		{
			msgType:     "/cosmos.staking.v1beta1.MsgDelegate",
			expectedGas: DelegateGasLimit,
		},
		{
			msgType:     "/cosmos.staking.v1beta1.MsgUndelegate",
			expectedGas: DelegateGasLimit,
		},
		{
			msgType:     "/cosmos.staking.v1beta1.MsgRedelegate",
			expectedGas: DelegateGasLimit * 2,
		},
		{
			msgType:     "/unknown.message.type",
			expectedGas: DefaultGasLimitValue,
		},
	}

	for _, tt := range tests {
		t.Run(tt.msgType, func(t *testing.T) {
			gas := GasLimitForMessageType(tt.msgType)
			require.Equal(t, tt.expectedGas, gas)
		})
	}
}

// TestEstimateGasForMessages tests quick gas estimation.
func TestEstimateGasForMessages(t *testing.T) {
	msgs := []sdk.Msg{
		&banktypes.MsgSend{
			FromAddress: "sonr1xyz...",
			ToAddress:   "sonr1abc...",
			Amount:      sdk.NewCoins(sdk.NewInt64Coin("usnr", 1000)),
		},
	}

	gas := EstimateGasForMessages(msgs)

	// Should be SendGasLimit + base overhead
	expected := uint64(SendGasLimit + 50000)
	require.Equal(t, expected, gas)
}

// TestValidateGasPrice tests gas price validation.
func TestValidateGasPrice(t *testing.T) {
	tests := []struct {
		name      string
		gasPrice  float64
		wantError bool
	}{
		{
			name:      "valid gas price",
			gasPrice:  0.025,
			wantError: false,
		},
		{
			name:      "zero gas price",
			gasPrice:  0,
			wantError: true,
		},
		{
			name:      "negative gas price",
			gasPrice:  -0.1,
			wantError: true,
		},
		{
			name:      "excessive gas price",
			gasPrice:  2.0,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateGasPrice(tt.gasPrice)
			if tt.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestValidateGasLimit tests gas limit validation.
func TestValidateGasLimit(t *testing.T) {
	tests := []struct {
		name      string
		gasLimit  uint64
		wantError bool
	}{
		{
			name:      "valid gas limit",
			gasLimit:  200000,
			wantError: false,
		},
		{
			name:      "zero gas limit",
			gasLimit:  0,
			wantError: true,
		},
		{
			name:      "excessive gas limit",
			gasLimit:  100000000,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateGasLimit(tt.gasLimit)
			if tt.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestOptimizeGasConfig tests gas configuration optimization.
func TestOptimizeGasConfig(t *testing.T) {
	config := &GasConfig{
		Adjustment:  0.5, // Too low
		MinGasPrice: 0.01,
		MaxGasLimit: 10000000,
		Denom:       "usnr",
	}

	networkInfo := &NetworkGasInfo{
		MinGasPrice:         0.025,
		MedianGasPrice:      0.03,
		RecommendedGasPrice: 0.035,
		MaxGasLimit:         5000000,
	}

	optimized := OptimizeGasConfig(config, networkInfo)

	// Should use recommended gas price
	require.Equal(t, networkInfo.RecommendedGasPrice, optimized.MinGasPrice)

	// Should adjust to minimum adjustment
	require.Equal(t, MinGasAdjustment, optimized.Adjustment)

	// Should use network max gas limit
	require.Equal(t, networkInfo.MaxGasLimit, optimized.MaxGasLimit)
}
