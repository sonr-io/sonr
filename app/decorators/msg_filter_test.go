package decorators_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"

	"github.com/cometbft/cometbft/crypto/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/suite"

	app "github.com/sonr-io/snrd/app"
	"github.com/sonr-io/snrd/app/decorators"
)

type AnteTestSuite struct {
	suite.Suite

	ctx sdk.Context
	app *app.SonrApp
}

func (s *AnteTestSuite) SetupTest() {
	isCheckTx := false
	s.app = app.Setup(s.T())
	s.ctx = s.app.BaseApp.NewContext(isCheckTx)
}

func TestAnteTestSuite(t *testing.T) {
	suite.Run(t, new(AnteTestSuite))
}

// Test the change rate decorator with standard edit msgs,
func (s *AnteTestSuite) TestAnteMsgFilterLogic() {
	acc := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	// test blocking any BankSend Messages
	ante := decorators.FilterDecorator(&banktypes.MsgSend{})
	msg := banktypes.NewMsgSend(
		acc,
		acc,
		sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(1))),
	)
	_, err := ante.AnteHandle(s.ctx, decorators.NewMockTx(msg), false, decorators.EmptyAnte)
	s.Require().Error(err)

	// validate other messages go through still (such as MsgMultiSend)
	msgMultiSend := banktypes.NewMsgMultiSend(
		banktypes.NewInput(acc, sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(1)))),
		[]banktypes.Output{banktypes.NewOutput(acc, sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(1))))},
	)
	_, err = ante.AnteHandle(s.ctx, decorators.NewMockTx(msgMultiSend), false, decorators.EmptyAnte)
	s.Require().NoError(err)
}
