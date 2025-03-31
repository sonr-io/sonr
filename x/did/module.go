package module

import (
	"context"
	"encoding/json"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"

	abci "github.com/cometbft/cometbft/abci/types"

	"cosmossdk.io/client/v2/autocli"
	errorsmod "cosmossdk.io/errors"
	nftkeeper "cosmossdk.io/x/nft/keeper"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/sonr-io/snrd/x/did/keeper"
	"github.com/sonr-io/snrd/x/did/types"
	// this line is used by starport scaffolding # 1
)

const (
	ConsensusVersion = 1

// this line is used by starport scaffolding # simapp/module/const
)

var (
	_ module.AppModuleBasic    = AppModuleBasic{}
	_ module.AppModuleGenesis  = AppModule{}
	_ module.AppModule         = AppModule{}
	_ autocli.HasAutoCLIConfig = AppModule{}
)

// AppModuleBasic defines the basic application module used by the wasm module.
type AppModuleBasic struct {
	cdc codec.Codec
}

type AppModule struct {
	AppModuleBasic

	keeper    keeper.Keeper
	nftKeeper nftkeeper.Keeper
}

// NewAppModule constructor
func NewAppModule(
	cdc codec.Codec,
	keeper keeper.Keeper,
	nftKeeper nftkeeper.Keeper,
) *AppModule {
	return &AppModule{
		AppModuleBasic: AppModuleBasic{cdc: cdc},
		keeper:         keeper,
		nftKeeper:      nftKeeper,
	}
}

func (a AppModuleBasic) Name() string {
	return types.ModuleName
}

func (a AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(&types.GenesisState{
		Params: types.DefaultParams(),
	})
}

func (a AppModuleBasic) ValidateGenesis(marshaler codec.JSONCodec, _ client.TxEncodingConfig, message json.RawMessage) error {
	var data types.GenesisState
	err := marshaler.UnmarshalJSON(message, &data)
	if err != nil {
		return err
	}
	if err := data.Params.Validate(); err != nil {
		return errorsmod.Wrap(err, "params")
	}
	return nil
}

func (a AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	err := types.RegisterQueryHandlerClient(context.Background(), mux, types.NewQueryClient(clientCtx))
	if err != nil {
		panic(err)
	}
}

func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	types.RegisterLegacyAminoCodec(cdc)
}

func (a AppModuleBasic) RegisterInterfaces(r codectypes.InterfaceRegistry) {
	types.RegisterInterfaces(r)
}

func (a AppModule) InitGenesis(ctx sdk.Context, marshaler codec.JSONCodec, message json.RawMessage) []abci.ValidatorUpdate {
	didGenesisState := types.DefaultGenesis()
	if err := a.keeper.Params.Set(ctx, didGenesisState.Params); err != nil {
		panic(err)
	}
	// nftGenesisState := nft.DefaultGenesisState()
	// if err := types.DefaultNFTClasses(nftGenesisState); err != nil {
	// 	panic(err)
	// }
	// a.nftKeeper.InitGenesis(ctx, nftGenesisState)
	return nil
}

func (a AppModule) ExportGenesis(ctx sdk.Context, marshaler codec.JSONCodec) json.RawMessage {
	genState := a.keeper.ExportGenesis(ctx)
	return marshaler.MustMarshalJSON(genState)
}

func (a AppModule) RegisterInvariants(_ sdk.InvariantRegistry) {
}

func (a AppModule) QuerierRoute() string {
	return types.QuerierRoute
}

func (a AppModule) RegisterServices(cfg module.Configurator) {
	types.RegisterMsgServer(cfg.MsgServer(), keeper.NewMsgServerImpl(a.keeper))
	types.RegisterQueryServer(cfg.QueryServer(), keeper.NewQuerier(a.keeper))
}

// ConsensusVersion is a sequence number for state-breaking change of the
// module. It should be incremented on each consensus-breaking change
// introduced by the module. To avoid wrong/empty versions, the initial version
// should be set to 1.
func (a AppModule) ConsensusVersion() uint64 {
	return ConsensusVersion
}
