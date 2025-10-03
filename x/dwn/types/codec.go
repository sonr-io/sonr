package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

var (
	amino     = codec.NewLegacyAmino()
	AminoCdc  = codec.NewAminoCodec(amino)
	ModuleCdc = codec.NewProtoCodec(types.NewInterfaceRegistry())
)

func init() {
	RegisterLegacyAminoCodec(amino)
	cryptocodec.RegisterCrypto(amino)
	sdk.RegisterLegacyAminoCodec(amino)
}

// RegisterLegacyAminoCodec registers concrete types on the LegacyAmino codec
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgUpdateParams{}, ModuleName+"/MsgUpdateParams", nil)
	cdc.RegisterConcrete(&MsgRecordsWrite{}, ModuleName+"/MsgRecordsWrite", nil)
	cdc.RegisterConcrete(&MsgRecordsDelete{}, ModuleName+"/MsgRecordsDelete", nil)
	cdc.RegisterConcrete(&MsgProtocolsConfigure{}, ModuleName+"/MsgProtocolsConfigure", nil)
	cdc.RegisterConcrete(&MsgPermissionsGrant{}, ModuleName+"/MsgPermissionsGrant", nil)
	cdc.RegisterConcrete(&MsgPermissionsRevoke{}, ModuleName+"/MsgPermissionsRevoke", nil)
}

func RegisterInterfaces(registry types.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgUpdateParams{},
		&MsgRecordsWrite{},
		&MsgRecordsDelete{},
		&MsgProtocolsConfigure{},
		&MsgPermissionsGrant{},
		&MsgPermissionsRevoke{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
