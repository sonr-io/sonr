package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

var (
	amino    = codec.NewLegacyAmino()
	AminoCdc = codec.NewAminoCodec(amino)
)

func init() {
	RegisterLegacyAminoCodec(amino)
	cryptocodec.RegisterCrypto(amino)
	sdk.RegisterLegacyAminoCodec(amino)
}

// RegisterLegacyAminoCodec registers concrete types on the LegacyAmino codec
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgRegisterDEXAccount{}, ModuleName+"/MsgRegisterDEXAccount", nil)
	cdc.RegisterConcrete(&MsgExecuteSwap{}, ModuleName+"/MsgExecuteSwap", nil)
	cdc.RegisterConcrete(&MsgProvideLiquidity{}, ModuleName+"/MsgProvideLiquidity", nil)
	cdc.RegisterConcrete(&MsgRemoveLiquidity{}, ModuleName+"/MsgRemoveLiquidity", nil)
	cdc.RegisterConcrete(&MsgCreateLimitOrder{}, ModuleName+"/MsgCreateLimitOrder", nil)
	cdc.RegisterConcrete(&MsgCancelOrder{}, ModuleName+"/MsgCancelOrder", nil)
}

// RegisterInterfaces registers the x/dex interfaces types with a given
// interface registry
func RegisterInterfaces(registry types.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgRegisterDEXAccount{},
		&MsgExecuteSwap{},
		&MsgProvideLiquidity{},
		&MsgRemoveLiquidity{},
		&MsgCreateLimitOrder{},
		&MsgCancelOrder{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
