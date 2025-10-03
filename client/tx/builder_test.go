package tx

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/sonr-io/sonr/client/config"
)

// TxBuilderTestSuite tests the transaction builder.
type TxBuilderTestSuite struct {
	suite.Suite
	builder TxBuilder
	config  *config.NetworkConfig
}

func (suite *TxBuilderTestSuite) SetupTest() {
	cfg := config.LocalNetwork()
	suite.config = &cfg

	// Create mock gRPC connection for testing
	conn, err := grpc.Dial("localhost:9090", grpc.WithInsecure())
	suite.Require().NoError(err)

	builder, err := NewTxBuilder(suite.config, conn)
	suite.Require().NoError(err)
	suite.builder = builder
}

func (suite *TxBuilderTestSuite) TestAddMessage() {
	// Create a test message
	msg := &banktypes.MsgSend{
		FromAddress: "sonr1xyz...",
		ToAddress:   "sonr1abc...",
		Amount:      sdk.NewCoins(sdk.NewInt64Coin("usnr", 1000)),
	}

	// Add message
	suite.builder.AddMessage(msg)

	// Verify message was added
	unsignedTx, err := suite.builder.Build()
	suite.Require().NoError(err)
	suite.Require().NotNil(unsignedTx)
	suite.Require().Len(unsignedTx.Messages, 1)
}

func (suite *TxBuilderTestSuite) TestWithMemo() {
	memo := "test transaction"
	suite.builder = suite.builder.WithMemo(memo)

	unsignedTx, err := suite.builder.Build()
	suite.Require().NoError(err)
	suite.Require().NotNil(unsignedTx)
	// Memo is set internally in the transaction
}

func (suite *TxBuilderTestSuite) TestWithGasLimit() {
	gasLimit := uint64(200000)
	suite.builder = suite.builder.WithGasLimit(gasLimit)

	unsignedTx, err := suite.builder.Build()
	suite.Require().NoError(err)
	suite.Require().NotNil(unsignedTx)
	// Gas limit is set internally
}

func (suite *TxBuilderTestSuite) TestWithFee() {
	fee := sdk.NewCoins(sdk.NewInt64Coin("usnr", 5000))
	suite.builder = suite.builder.WithFee(fee)

	unsignedTx, err := suite.builder.Build()
	suite.Require().NoError(err)
	suite.Require().NotNil(unsignedTx)
	// Fee is set internally
}

func (suite *TxBuilderTestSuite) TestClearMessages() {
	// Add some data
	msg := &banktypes.MsgSend{
		FromAddress: "sonr1xyz...",
		ToAddress:   "sonr1abc...",
		Amount:      sdk.NewCoins(sdk.NewInt64Coin("usnr", 1000)),
	}
	suite.builder = suite.builder.AddMessage(msg)
	suite.builder = suite.builder.WithMemo("test")

	// Clear messages
	suite.builder = suite.builder.ClearMessages()

	// Build should create transaction with no messages
	unsignedTx, err := suite.builder.Build()
	suite.Require().NoError(err)
	suite.Require().Len(unsignedTx.Messages, 0)
}

func (suite *TxBuilderTestSuite) TestMultipleMessages() {
	// Add multiple messages
	msg1 := &banktypes.MsgSend{
		FromAddress: "sonr1xyz...",
		ToAddress:   "sonr1abc...",
		Amount:      sdk.NewCoins(sdk.NewInt64Coin("usnr", 1000)),
	}

	msg2 := &banktypes.MsgSend{
		FromAddress: "sonr1abc...",
		ToAddress:   "sonr1def...",
		Amount:      sdk.NewCoins(sdk.NewInt64Coin("usnr", 2000)),
	}

	suite.builder.AddMessage(msg1)
	suite.builder.AddMessage(msg2)

	unsignedTx, err := suite.builder.Build()
	suite.Require().NoError(err)
	suite.Require().Len(unsignedTx.Messages, 2)
}

func TestTxBuilderTestSuite(t *testing.T) {
	suite.Run(t, new(TxBuilderTestSuite))
}

// TestTxBuilderValidation tests transaction builder validation.
func TestTxBuilderValidation(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(TxBuilder) TxBuilder
		wantError bool
		errorMsg  string
	}{
		{
			name: "valid transaction",
			setup: func(b TxBuilder) TxBuilder {
				msg := &banktypes.MsgSend{
					FromAddress: "sonr1xyz...",
					ToAddress:   "sonr1abc...",
					Amount:      sdk.NewCoins(sdk.NewInt64Coin("usnr", 1000)),
				}
				return b.AddMessage(msg).WithGasLimit(100000)
			},
			wantError: false,
		},
		{
			name: "no messages",
			setup: func(b TxBuilder) TxBuilder {
				return b.WithGasLimit(100000)
			},
			wantError: false, // Empty transactions are technically valid
		},
		{
			name: "zero gas limit",
			setup: func(b TxBuilder) TxBuilder {
				msg := &banktypes.MsgSend{
					FromAddress: "sonr1xyz...",
					ToAddress:   "sonr1abc...",
					Amount:      sdk.NewCoins(sdk.NewInt64Coin("usnr", 1000)),
				}
				return b.AddMessage(msg).WithGasLimit(0)
			},
			wantError: false, // Zero gas is allowed for simulation
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.LocalNetwork()
			conn, _ := grpc.Dial("localhost:9090", grpc.WithInsecure())

			builder, err := NewTxBuilder(&cfg, conn)
			require.NoError(t, err)

			builder = tt.setup(builder)

			_, err = builder.Build()
			if tt.wantError {
				require.Error(t, err)
				if tt.errorMsg != "" {
					require.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}
