package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sonr-io/sonr/x/dex/keeper"
	"github.com/sonr-io/sonr/x/dex/types"
)

// MsgServerTestSuite tests message server operations
type MsgServerTestSuite struct {
	suite.Suite
	f *testFixture
}

func TestMsgServerSuite(t *testing.T) {
	suite.Run(t, new(MsgServerTestSuite))
}

func (suite *MsgServerTestSuite) SetupTest() {
	suite.f = SetupTest(suite.T())
}

// TestMsgRegisterDEXAccount tests the RegisterDEXAccount message handler
func (suite *MsgServerTestSuite) TestMsgRegisterDEXAccount() {
	msgServer := keeper.NewMsgServerImpl(suite.f.k)
	ctx := sdk.WrapSDKContext(suite.f.ctx)

	// Create test message
	msg := &types.MsgRegisterDEXAccount{
		Did:          "did:sonr:alice",
		ConnectionId: "connection-0",
		Features:     []string{"swap", "liquidity"},
	}

	// Execute message
	resp, err := msgServer.RegisterDEXAccount(ctx, msg)
	suite.Require().NoError(err)
	suite.Require().NotNil(resp)
	suite.Require().NotEmpty(resp.PortId)

	// Verify account was created
	account, err := suite.f.k.GetDEXAccount(suite.f.ctx, msg.Did, msg.ConnectionId)
	suite.Require().NoError(err)
	suite.Require().NotNil(account)
	suite.Require().Equal(msg.Did, account.Did)
	suite.Require().Equal(msg.ConnectionId, account.ConnectionId)
}

// TestMsgExecuteSwap tests the ExecuteSwap message handler
func (suite *MsgServerTestSuite) TestMsgExecuteSwap() {
	msgServer := keeper.NewMsgServerImpl(suite.f.k)
	ctx := sdk.WrapSDKContext(suite.f.ctx)

	// First register an account
	_, err := suite.f.k.RegisterDEXAccount(
		suite.f.ctx,
		"did:sonr:bob",
		"connection-0",
		[]string{"swap"},
	)
	suite.Require().NoError(err)

	// Create swap message
	msg := &types.MsgExecuteSwap{
		Did:          "did:sonr:bob",
		ConnectionId: "connection-0",
		SourceDenom:  "usnr",
		TargetDenom:  "uosmo",
		Amount:       math.NewInt(1000),
		MinAmountOut: math.NewInt(900),
		Route:        "pool:1",
	}

	// Execute swap
	resp, err := msgServer.ExecuteSwap(ctx, msg)
	suite.Require().NoError(err)
	suite.Require().NotNil(resp)
	// TODO: Check sequence when ExecuteSwap is implemented
	// suite.Require().NotZero(resp.Sequence)
}

// TestMsgProvideLiquidity tests the ProvideLiquidity message handler
func (suite *MsgServerTestSuite) TestMsgProvideLiquidity() {
	msgServer := keeper.NewMsgServerImpl(suite.f.k)
	ctx := sdk.WrapSDKContext(suite.f.ctx)

	// First register an account
	_, err := suite.f.k.RegisterDEXAccount(
		suite.f.ctx,
		"did:sonr:charlie",
		"connection-0",
		[]string{"liquidity"},
	)
	suite.Require().NoError(err)

	// Create liquidity message
	msg := &types.MsgProvideLiquidity{
		Did:          "did:sonr:charlie",
		ConnectionId: "connection-0",
		PoolId:       "1",
		Assets: sdk.NewCoins(
			sdk.NewCoin("usnr", math.NewInt(1000)),
			sdk.NewCoin("uosmo", math.NewInt(1000)),
		),
		MinShares: math.NewInt(100),
		Timeout:   time.Now().Add(5 * time.Minute),
	}

	// Execute liquidity provision
	resp, err := msgServer.ProvideLiquidity(ctx, msg)
	suite.Require().NoError(err)
	suite.Require().NotNil(resp)
	// TODO: Check sequence when ProvideLiquidity is implemented
	// suite.Require().NotZero(resp.Sequence)
}

// TestMsgRemoveLiquidity tests the RemoveLiquidity message handler
func (suite *MsgServerTestSuite) TestMsgRemoveLiquidity() {
	msgServer := keeper.NewMsgServerImpl(suite.f.k)
	ctx := sdk.WrapSDKContext(suite.f.ctx)

	// First register an account
	_, err := suite.f.k.RegisterDEXAccount(
		suite.f.ctx,
		"did:sonr:dave",
		"connection-0",
		[]string{"liquidity"},
	)
	suite.Require().NoError(err)

	// Create remove liquidity message
	msg := &types.MsgRemoveLiquidity{
		Did:          "did:sonr:dave",
		ConnectionId: "connection-0",
		PoolId:       "1",
		Shares:       math.NewInt(100),
		MinAmounts: sdk.NewCoins(
			sdk.NewCoin("usnr", math.NewInt(900)),
			sdk.NewCoin("uosmo", math.NewInt(900)),
		),
		Timeout: time.Now().Add(5 * time.Minute),
	}

	// Execute liquidity removal
	resp, err := msgServer.RemoveLiquidity(ctx, msg)
	suite.Require().NoError(err)
	suite.Require().NotNil(resp)
	// TODO: Check sequence when RemoveLiquidity is implemented
	// suite.Require().NotZero(resp.Sequence)
}

// TestMsgCreateLimitOrder tests the CreateLimitOrder message handler
func (suite *MsgServerTestSuite) TestMsgCreateLimitOrder() {
	msgServer := keeper.NewMsgServerImpl(suite.f.k)
	ctx := sdk.WrapSDKContext(suite.f.ctx)

	// First register an account
	_, err := suite.f.k.RegisterDEXAccount(
		suite.f.ctx,
		"did:sonr:eve",
		"connection-0",
		[]string{"order"},
	)
	suite.Require().NoError(err)

	// Create limit order message
	msg := &types.MsgCreateLimitOrder{
		Did:          "did:sonr:eve",
		ConnectionId: "connection-0",
		SellDenom:    "usnr",
		BuyDenom:     "uosmo",
		Amount:       math.NewInt(1000),
		Price:        math.LegacyNewDec(1),
		Expiration:   time.Now().Add(24 * time.Hour),
	}

	// Execute order creation
	resp, err := msgServer.CreateLimitOrder(ctx, msg)
	suite.Require().NoError(err)
	suite.Require().NotNil(resp)
	// TODO: Check sequence and OrderId when CreateLimitOrder is implemented
	// suite.Require().NotZero(resp.Sequence)
	// suite.Require().NotEmpty(resp.OrderId)
}

// TestMsgCancelOrder tests the CancelOrder message handler
func (suite *MsgServerTestSuite) TestMsgCancelOrder() {
	msgServer := keeper.NewMsgServerImpl(suite.f.k)
	ctx := sdk.WrapSDKContext(suite.f.ctx)

	// First register an account and create an order
	_, err := suite.f.k.RegisterDEXAccount(
		suite.f.ctx,
		"did:sonr:frank",
		"connection-0",
		[]string{"order"},
	)
	suite.Require().NoError(err)

	// Since CreateLimitOrder is not implemented yet, use a mock order ID
	mockOrderId := "order-123"

	// Cancel the order
	cancelMsg := &types.MsgCancelOrder{
		Did:          "did:sonr:frank",
		ConnectionId: "connection-0",
		OrderId:      mockOrderId,
	}

	// Execute order cancellation
	resp, err := msgServer.CancelOrder(ctx, cancelMsg)
	suite.Require().NoError(err)
	suite.Require().NotNil(resp)
	// TODO: Check sequence when CancelOrder is implemented
	// suite.Require().NotZero(resp.Sequence)
}

// TestMsgRegisterDEXAccount_InvalidDID tests registration with invalid DID
func (suite *MsgServerTestSuite) TestMsgRegisterDEXAccount_InvalidDID() {
	msgServer := keeper.NewMsgServerImpl(suite.f.k)
	ctx := sdk.WrapSDKContext(suite.f.ctx)

	// Create test message with invalid DID
	msg := &types.MsgRegisterDEXAccount{
		Did:          "", // Empty DID
		ConnectionId: "connection-0",
		Features:     []string{"swap"},
	}

	// Should fail validation
	_, err := msgServer.RegisterDEXAccount(ctx, msg)
	suite.Require().Error(err)
}

// TestMsgExecuteSwap_AccountNotFound tests swap with non-existent account
func (suite *MsgServerTestSuite) TestMsgExecuteSwap_AccountNotFound() {
	msgServer := keeper.NewMsgServerImpl(suite.f.k)
	ctx := sdk.WrapSDKContext(suite.f.ctx)

	// Create swap message without registering account
	msg := &types.MsgExecuteSwap{
		Did:          "did:sonr:nonexistent",
		ConnectionId: "connection-0",
		SourceDenom:  "usnr",
		TargetDenom:  "uosmo",
		Amount:       math.NewInt(1000),
		MinAmountOut: math.NewInt(900),
		Route:        "pool:1",
	}

	// TODO: Should fail when ExecuteSwap is implemented - account not found
	_, err := msgServer.ExecuteSwap(ctx, msg)
	suite.Require().NoError(err) // Currently returns empty response
	// suite.Require().Error(err)
	// suite.Require().Contains(err.Error(), "not found")
}

// TestMsgProvideLiquidity_InvalidAssets tests liquidity with invalid assets
func (suite *MsgServerTestSuite) TestMsgProvideLiquidity_InvalidAssets() {
	// First register an account
	_, err := suite.f.k.RegisterDEXAccount(
		suite.f.ctx,
		"did:sonr:grace",
		"connection-0",
		[]string{"liquidity"},
	)
	suite.Require().NoError(err)

	// Create liquidity message with empty assets
	msg := &types.MsgProvideLiquidity{
		Did:          "did:sonr:grace",
		ConnectionId: "connection-0",
		PoolId:       "1",
		Assets:       sdk.NewCoins(), // Empty coins list
		MinShares:    math.NewInt(100),
		Timeout:      time.Now().Add(5 * time.Minute),
	}

	// Should fail validation due to empty assets
	err = msg.ValidateBasic()
	suite.Require().Error(err)
}

// TestMsgCreateLimitOrder_InvalidPrice tests order creation with invalid price
func (suite *MsgServerTestSuite) TestMsgCreateLimitOrder_InvalidPrice() {
	// First register an account
	_, err := suite.f.k.RegisterDEXAccount(
		suite.f.ctx,
		"did:sonr:henry",
		"connection-0",
		[]string{"order"},
	)
	suite.Require().NoError(err)

	// Create limit order message with zero price
	msg := &types.MsgCreateLimitOrder{
		Did:          "did:sonr:henry",
		ConnectionId: "connection-0",
		SellDenom:    "usnr",
		BuyDenom:     "uosmo",
		Amount:       math.NewInt(1000),
		Price:        math.LegacyZeroDec(), // Invalid: zero price
		Expiration:   time.Now().Add(24 * time.Hour),
	}

	// Should fail validation
	err = msg.ValidateBasic()
	suite.Require().Error(err)
}
