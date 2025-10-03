package context

import (
	"context"
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
)

// BroadcastTx broadcasts a transaction to the blockchain using the stored client context
func (sc *SonrContext) BroadcastTx(txBytes []byte) error {
	clientCtx, err := sc.GetClientContext()
	if err != nil {
		return fmt.Errorf("failed to get client context: %w", err)
	}

	// Create broadcast request
	txReq := &txtypes.BroadcastTxRequest{
		TxBytes: txBytes,
		Mode:    txtypes.BroadcastMode_BROADCAST_MODE_SYNC,
	}

	// Get the gRPC client connection
	grpcConn := clientCtx.GRPCClient

	// Create transaction service client
	txClient := txtypes.NewServiceClient(grpcConn)

	// Broadcast the transaction
	res, err := txClient.BroadcastTx(context.Background(), txReq)
	if err != nil {
		return fmt.Errorf("failed to broadcast transaction: %w", err)
	}

	// Check if transaction was accepted
	if res.TxResponse.Code != 0 {
		return fmt.Errorf(
			"transaction failed with code %d: %s",
			res.TxResponse.Code,
			res.TxResponse.RawLog,
		)
	}

	return nil
}

// BroadcastTxWithResponse broadcasts a transaction and returns the response
func (sc *SonrContext) BroadcastTxWithResponse(
	txBytes []byte,
) (*txtypes.BroadcastTxResponse, error) {
	clientCtx, err := sc.GetClientContext()
	if err != nil {
		return nil, fmt.Errorf("failed to get client context: %w", err)
	}

	// Create broadcast request
	txReq := &txtypes.BroadcastTxRequest{
		TxBytes: txBytes,
		Mode:    txtypes.BroadcastMode_BROADCAST_MODE_SYNC,
	}

	// Get the gRPC client connection
	grpcConn := clientCtx.GRPCClient

	// Create transaction service client
	txClient := txtypes.NewServiceClient(grpcConn)

	// Broadcast the transaction
	res, err := txClient.BroadcastTx(context.Background(), txReq)
	if err != nil {
		return nil, fmt.Errorf("failed to broadcast transaction: %w", err)
	}

	return res, nil
}

// SignAndBroadcastTx signs a transaction and broadcasts it
func (sc *SonrContext) SignAndBroadcastTx(txBuilder client.TxBuilder) error {
	clientCtx, err := sc.GetClientContext()
	if err != nil {
		return fmt.Errorf("failed to get client context: %w", err)
	}

	// Sign the transaction
	txFactory, err := tx.NewFactoryCLI(clientCtx, nil)
	if err != nil {
		return fmt.Errorf("failed to create tx factory: %w", err)
	}
	err = tx.Sign(clientCtx.CmdContext, txFactory, clientCtx.GetFromName(), txBuilder, true)
	if err != nil {
		return fmt.Errorf("failed to sign transaction: %w", err)
	}

	// Encode the transaction
	txBytes, err := clientCtx.TxConfig.TxEncoder()(txBuilder.GetTx())
	if err != nil {
		return fmt.Errorf("failed to encode transaction: %w", err)
	}

	// Broadcast the transaction
	return sc.BroadcastTx(txBytes)
}

// CreateUnsignedTx creates an unsigned transaction from messages
func (sc *SonrContext) CreateUnsignedTx(msgs ...sdk.Msg) (client.TxBuilder, error) {
	clientCtx, err := sc.GetClientContext()
	if err != nil {
		return nil, fmt.Errorf("failed to get client context: %w", err)
	}

	// Create transaction builder
	txBuilder := clientCtx.TxConfig.NewTxBuilder()

	// Set messages
	err = txBuilder.SetMsgs(msgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to set messages: %w", err)
	}

	// Set gas limit and fees (these should be estimated or configured)
	txBuilder.SetGasLimit(200000) // Default gas limit

	// Set fee amount (you may want to make this configurable)
	feeAmount := sdk.NewCoins(sdk.NewInt64Coin("usonr", 1000))
	txBuilder.SetFeeAmount(feeAmount)

	return txBuilder, nil
}

// EstimateGas estimates gas for a transaction
func (sc *SonrContext) EstimateGas(txBuilder client.TxBuilder) (uint64, error) {
	clientCtx, err := sc.GetClientContext()
	if err != nil {
		return 0, fmt.Errorf("failed to get client context: %w", err)
	}

	// Simulate the transaction to estimate gas
	simReq, err := sc.buildSimTx(clientCtx, txBuilder)
	if err != nil {
		return 0, fmt.Errorf("failed to build simulation request: %w", err)
	}

	// Get the gRPC client connection
	grpcConn := clientCtx.GRPCClient

	// Create transaction service client
	txClient := txtypes.NewServiceClient(grpcConn)

	// Simulate the transaction
	simRes, err := txClient.Simulate(context.Background(), simReq)
	if err != nil {
		return 0, fmt.Errorf("failed to simulate transaction: %w", err)
	}

	// Return estimated gas with some buffer
	return simRes.GasInfo.GasUsed + 10000, nil
}

// buildSimTx builds a simulation request from a transaction builder
func (sc *SonrContext) buildSimTx(
	clientCtx client.Context,
	txBuilder client.TxBuilder,
) (*txtypes.SimulateRequest, error) {
	// Create a copy of the transaction builder for simulation
	simBuilder := clientCtx.TxConfig.NewTxBuilder()
	err := simBuilder.SetMsgs(txBuilder.GetTx().GetMsgs()...)
	if err != nil {
		return nil, err
	}

	// Get account info for signature
	fromAddr := clientCtx.GetFromAddress()
	if fromAddr.Empty() {
		return nil, fmt.Errorf("from address is empty")
	}

	// Set dummy signature for simulation - we need a public key from the keyring
	keyInfo, err := clientCtx.Keyring.Key(clientCtx.GetFromName())
	if err != nil {
		return nil, fmt.Errorf("failed to get key info: %w", err)
	}
	pubKey, err := keyInfo.GetPubKey()
	if err != nil {
		return nil, fmt.Errorf("failed to get public key: %w", err)
	}

	sigV2 := signing.SignatureV2{
		PubKey: pubKey,
		Data: &signing.SingleSignatureData{
			SignMode:  signing.SignMode_SIGN_MODE_DIRECT,
			Signature: nil,
		},
		Sequence: 0,
	}

	err = simBuilder.SetSignatures(sigV2)
	if err != nil {
		return nil, err
	}

	// Encode the simulation transaction
	simTxBytes, err := clientCtx.TxConfig.TxEncoder()(simBuilder.GetTx())
	if err != nil {
		return nil, err
	}

	return &txtypes.SimulateRequest{
		TxBytes: simTxBytes,
	}, nil
}
