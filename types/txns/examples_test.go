package txns

import (
	"context"
	"fmt"
	"math/big"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/std"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/sonr-io/sonr/crypto/mpc"
	"github.com/sonr-io/sonr/types/coins"
)

// Example demonstrates basic transaction building workflow
func ExampleTransactionBuilder_basic() {
	// Setup
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	std.RegisterInterfaces(interfaceRegistry)
	banktypes.RegisterInterfaces(interfaceRegistry)

	marshaler := codec.NewProtoCodec(interfaceRegistry)
	txConfig := authtx.NewTxConfig(marshaler, authtx.DefaultSignModes)

	clientCtx := client.Context{}.
		WithCodec(marshaler).
		WithTxConfig(txConfig).
		WithInterfaceRegistry(interfaceRegistry)

	// Create coins manager and transaction builder
	coinsManager := coins.NewManager("snr", "sonr-1", big.NewInt(1))
	txBuilder := NewTransactionBuilder(coinsManager, "sonr-1")

	// Build a Cosmos transaction
	cosmosBuilder := txBuilder.Cosmos(clientCtx)
	cosmosBuilder.SetGas(200000, sdk.NewDecCoin("usnr", math.NewInt(1000)))
	cosmosBuilder.SetMemo("Example transaction")

	// Create a send message
	sendMsg := &banktypes.MsgSend{
		FromAddress: "snr1sender123",
		ToAddress:   "snr1receiver456",
		Amount:      sdk.NewCoins(sdk.NewCoin("usnr", math.NewInt(1000000))),
	}

	params := &CosmosTransactionParams{
		Messages: []sdk.Msg{sendMsg},
		GasLimit: 200000,
		GasPrice: sdk.NewDecCoin("usnr", math.NewInt(1000)),
		Memo:     "Example transaction",
	}

	unsignedTx, err := cosmosBuilder.BuildUnsigned(params)
	if err != nil {
		fmt.Printf("Error building transaction: %v\n", err)
		return
	}

	fmt.Printf("Transaction type: %s\n", unsignedTx.GetType())
	fmt.Printf("Encoding: %s\n", unsignedTx.GetEncoding())

	// Output:
	// Transaction type: cosmos
	// Encoding: protobuf
}

// Example demonstrates EVM transaction building
func ExampleTransactionBuilder_evm() {
	// Setup
	chainID := big.NewInt(1) // Ethereum mainnet
	coinsManager := coins.NewManager("snr", "sonr-1", chainID)
	txBuilder := NewTransactionBuilder(coinsManager, "sonr-1")

	// Build an EVM transaction
	evmBuilder := txBuilder.EVM(chainID)
	evmBuilder.SetGas(21000, big.NewInt(20000000000)) // 20 Gwei
	evmBuilder.SetNonce(42)

	// Create transfer parameters
	toAddr := common.HexToAddress("0x742d35Cc6634C0532925a3b8D80C6634C0532925")
	params := &EVMTransactionParams{
		To:       &toAddr,
		Value:    big.NewInt(1000000000000000000), // 1 ETH
		GasLimit: 21000,
		GasPrice: big.NewInt(20000000000),
		Nonce:    42,
		ChainID:  chainID,
	}

	unsignedTx, err := evmBuilder.BuildUnsigned(params)
	if err != nil {
		fmt.Printf("Error building transaction: %v\n", err)
		return
	}

	fmt.Printf("Transaction type: %s\n", unsignedTx.GetType())
	fmt.Printf("Encoding: %s\n", unsignedTx.GetEncoding())

	// Output:
	// Transaction type: evm
	// Encoding: rlp
}

// Example demonstrates fee estimation
func ExampleFeeManager_estimate() {
	// Setup client context
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	std.RegisterInterfaces(interfaceRegistry)
	banktypes.RegisterInterfaces(interfaceRegistry)

	marshaler := codec.NewProtoCodec(interfaceRegistry)
	txConfig := authtx.NewTxConfig(marshaler, authtx.DefaultSignModes)

	clientCtx := client.Context{}.
		WithCodec(marshaler).
		WithTxConfig(txConfig).
		WithInterfaceRegistry(interfaceRegistry)

	// Create fee manager
	feeManager := CreateDefaultFeeManager(clientCtx, nil, big.NewInt(1))

	// Estimate Cosmos transaction fee
	cosmosParams := &CosmosTransactionParams{
		Messages: []sdk.Msg{
			&banktypes.MsgSend{
				FromAddress: "snr1sender",
				ToAddress:   "snr1receiver",
				Amount:      sdk.NewCoins(sdk.NewCoin("usnr", math.NewInt(1000))),
			},
		},
	}

	cosmosFee, err := feeManager.EstimateFee(
		context.Background(),
		TransactionTypeCosmos,
		cosmosParams,
	)
	if err != nil {
		fmt.Printf("Error estimating Cosmos fee: %v\n", err)
		return
	}

	fmt.Printf("Cosmos gas limit: %d\n", cosmosFee.GasLimit)
	fmt.Printf("Cosmos fee total: %s\n", cosmosFee.Total)

	// Estimate EVM transaction fee
	toAddr := common.HexToAddress("0x742d35Cc6634C0532925a3b8D80C")
	evmParams := &EVMTransactionParams{
		To:       &toAddr,
		Value:    big.NewInt(1000000000000000000),
		GasLimit: 21000,
		GasPrice: big.NewInt(20000000000),
	}

	evmFee, err := feeManager.EstimateFee(context.Background(), TransactionTypeEVM, evmParams)
	if err != nil {
		fmt.Printf("Error estimating EVM fee: %v\n", err)
		return
	}

	fmt.Printf("EVM gas limit: %d\n", evmFee.GasLimit)
	fmt.Printf("EVM fee total: %s\n", evmFee.Total)
}

// Example demonstrates address derivation
func ExampleAddressManager_derive() {
	// Setup
	coinsManager := coins.NewManager("snr", "sonr-1", big.NewInt(1))
	addressManager := NewAddressManager("snr", coinsManager)

	// Derive addresses from entropy (DID + salt)
	derivation, err := addressManager.DeriveFromEntropy("did:sonr:test", "salt123", "snr")
	if err != nil {
		fmt.Printf("Error deriving addresses: %v\n", err)
		return
	}

	fmt.Printf("Cosmos address: %s\n", derivation.CosmosAddress)
	fmt.Printf("EVM address: %s\n", derivation.EVMAddress)
	fmt.Printf("Derivation path: %s\n", derivation.DerivationPath)
	fmt.Printf("Chain type: %s\n", derivation.ChainType)
}

// Example demonstrates MPC enclave integration
func ExampleMPCSigner_usage() {
	// This example shows how to integrate with MPC enclave
	// Note: In real usage, enclave data would come from actual MPC operations

	// Create mock enclave data for demonstration
	// In real usage, this would come from mpc.NewEnclave() or similar
	enclaveData := &mpc.EnclaveData{
		// Mock data - real implementation would have actual enclave data
	}

	// Create MPC signer
	signer, err := NewMPCSigner(enclaveData, "sonr-1")
	if err != nil {
		fmt.Printf("Error creating MPC signer: %v\n", err)
		return
	}

	// Get public key
	pubKey := signer.GetPublicKey()
	fmt.Printf("Public key length: %d bytes\n", len(pubKey))

	// Get addresses for different chain types
	cosmosAddr, err := signer.GetAddress(TransactionTypeCosmos)
	if err != nil {
		fmt.Printf("Error getting Cosmos address: %v\n", err)
		return
	}

	evmAddr, err := signer.GetAddress(TransactionTypeEVM)
	if err != nil {
		fmt.Printf("Error getting EVM address: %v\n", err)
		return
	}

	fmt.Printf("Cosmos address: %s\n", cosmosAddr)
	fmt.Printf("EVM address: %s\n", evmAddr)
}

// Example demonstrates cross-chain transaction workflow
func ExampleTransactionBuilder_crossChain() {
	// Setup
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	std.RegisterInterfaces(interfaceRegistry)
	banktypes.RegisterInterfaces(interfaceRegistry)

	marshaler := codec.NewProtoCodec(interfaceRegistry)
	txConfig := authtx.NewTxConfig(marshaler, authtx.DefaultSignModes)

	_ = client.Context{}.
		WithCodec(marshaler).
		WithTxConfig(txConfig).
		WithInterfaceRegistry(interfaceRegistry)

	// Create transaction builder and wallet
	coinsManager := coins.NewManager("snr", "sonr-1", big.NewInt(1))
	txBuilder := NewTransactionBuilder(coinsManager, "sonr-1")

	// Create wallet from entropy
	wallet, err := coinsManager.CreateWalletFromEntropy("did:sonr:user", "mysalt")
	if err != nil {
		fmt.Printf("Error creating wallet: %v\n", err)
		return
	}

	// Create signers for both chains
	_, err = txBuilder.CreateSigner(wallet, TransactionTypeCosmos)
	if err != nil {
		fmt.Printf("Error creating Cosmos signer: %v\n", err)
		return
	}

	_, err = txBuilder.CreateSigner(wallet, TransactionTypeEVM)
	if err != nil {
		fmt.Printf("Error creating EVM signer: %v\n", err)
		return
	}

	// Both signers can now be used to sign transactions for their respective chains
	fmt.Printf("Cross-chain wallet ready\n")

	// Output:
	// Cross-chain wallet ready
}

// Example demonstrates transaction encoding and decoding
func ExampleEncoderRegistry_usage() {
	// Setup
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	std.RegisterInterfaces(interfaceRegistry)
	banktypes.RegisterInterfaces(interfaceRegistry)

	marshaler := codec.NewProtoCodec(interfaceRegistry)
	txConfig := authtx.NewTxConfig(marshaler, authtx.DefaultSignModes)

	clientCtx := client.Context{}.
		WithCodec(marshaler).
		WithTxConfig(txConfig).
		WithInterfaceRegistry(interfaceRegistry)

	// Create encoder registry
	registry := DefaultEncoderRegistry(clientCtx)

	// Get Cosmos Protobuf encoder
	cosmosEncoder, err := registry.GetEncoderByType(EncodingTypeProtobuf, TransactionTypeCosmos)
	if err != nil {
		fmt.Printf("Error getting Cosmos encoder: %v\n", err)
		return
	}

	fmt.Printf("Cosmos encoder type: %s\n", cosmosEncoder.GetEncodingType())

	// Get EVM RLP encoder
	evmEncoder, err := registry.GetEncoderByType(EncodingTypeRLP, TransactionTypeEVM)
	if err != nil {
		fmt.Printf("Error getting EVM encoder: %v\n", err)
		return
	}

	fmt.Printf("EVM encoder type: %s\n", evmEncoder.GetEncodingType())

	// Output:
	// Cosmos encoder type: protobuf
	// EVM encoder type: rlp
}

// Example demonstrates batch address derivation
func ExampleAddressManager_batch() {
	// Setup
	coinsManager := coins.NewManager("snr", "sonr-1", big.NewInt(1))
	addressManager := NewAddressManager("snr", coinsManager)

	// Create batch requests
	requests := []AddressRequest{
		{
			Type:   "entropy",
			DID:    "did:sonr:user1",
			Salt:   "salt1",
			Prefix: "snr",
		},
		{
			Type:   "entropy",
			DID:    "did:sonr:user2",
			Salt:   "salt2",
			Prefix: "snr",
		},
	}

	// Derive batch of addresses
	batch, err := addressManager.DeriveAddressBatch(requests)
	if err != nil {
		fmt.Printf("Error deriving address batch: %v\n", err)
		return
	}

	fmt.Printf("Derived %d addresses\n", len(batch.Addresses))
	fmt.Printf("Cosmos prefix: %s\n", batch.Metadata["cosmos_prefix"])

	for i, addr := range batch.Addresses {
		fmt.Printf("Address %d: %s (Cosmos), %s (EVM)\n",
			i+1, addr.CosmosAddress, addr.EVMAddress)
	}
}

// Example demonstrates complete transaction workflow
func ExampleTransactionBuilder_complete() {
	// This example shows a complete transaction workflow from creation to signing

	// Setup
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	std.RegisterInterfaces(interfaceRegistry)
	banktypes.RegisterInterfaces(interfaceRegistry)

	marshaler := codec.NewProtoCodec(interfaceRegistry)
	txConfig := authtx.NewTxConfig(marshaler, authtx.DefaultSignModes)

	clientCtx := client.Context{}.
		WithCodec(marshaler).
		WithTxConfig(txConfig).
		WithInterfaceRegistry(interfaceRegistry)

	// 1. Create wallet and derive addresses
	coinsManager := coins.NewManager("snr", "sonr-1", big.NewInt(1))
	txBuilder := NewTransactionBuilder(coinsManager, "sonr-1")

	// Derive addresses
	derivation, err := txBuilder.DeriveAddresses("did:sonr:example", "examplesalt")
	if err != nil {
		fmt.Printf("Error deriving addresses: %v\n", err)
		return
	}

	// 2. Create wallet and signer
	wallet, err := coinsManager.CreateWalletFromEntropy("did:sonr:example", "examplesalt")
	if err != nil {
		fmt.Printf("Error creating wallet: %v\n", err)
		return
	}

	signer, err := txBuilder.CreateSigner(wallet, TransactionTypeCosmos)
	if err != nil {
		fmt.Printf("Error creating signer: %v\n", err)
		return
	}

	// 3. Build transaction
	cosmosBuilder := txBuilder.Cosmos(clientCtx)

	sendMsg := &banktypes.MsgSend{
		FromAddress: derivation.CosmosAddress,
		ToAddress:   "snr1receiver123",
		Amount:      sdk.NewCoins(sdk.NewCoin("usnr", math.NewInt(1000000))),
	}

	params := &CosmosTransactionParams{
		Messages: []sdk.Msg{sendMsg},
		GasLimit: 200000,
		GasPrice: sdk.NewDecCoin("usnr", math.NewInt(1000)),
		Memo:     "Complete example transaction",
	}

	// 4. Estimate fee
	feeManager := CreateDefaultFeeManager(clientCtx, nil, big.NewInt(1))
	_, err = feeManager.EstimateFee(context.Background(), TransactionTypeCosmos, params)
	if err != nil {
		fmt.Printf("Error estimating fee: %v\n", err)
		return
	}

	// 5. Build unsigned transaction
	unsignedTx, err := cosmosBuilder.BuildUnsigned(params)
	if err != nil {
		fmt.Printf("Error building transaction: %v\n", err)
		return
	}

	// 6. Sign transaction
	signBytes, err := unsignedTx.GetSignBytes()
	if err != nil {
		fmt.Printf("Error getting sign bytes: %v\n", err)
		return
	}

	signature, err := signer.Sign(signBytes)
	if err != nil {
		fmt.Printf("Error signing transaction: %v\n", err)
		return
	}

	pubKey := signer.GetPublicKey()
	signedTx, err := unsignedTx.Sign(signature, pubKey)
	if err != nil {
		fmt.Printf("Error creating signed transaction: %v\n", err)
		return
	}

	// 7. Get final transaction bytes
	_, err = signedTx.GetBytes()
	if err != nil {
		fmt.Printf("Error getting transaction bytes: %v\n", err)
		return
	}

	fmt.Printf("Transaction workflow completed\n")

	// Output:
	// Transaction workflow completed
}
