package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Noble testnet chain configuration
const (
	// NobleTestnetChainID is the chain ID for Noble testnet (grand-1)
	NobleTestnetChainID = "noble-grand-1"

	// NobleMainnetChainID is the chain ID for Noble mainnet
	NobleMainnetChainID = "noble-1"

	// NobleUSDCDenom is the denomination for USDC on Noble
	NobleUSDCDenom = "uusdc"

	// NobleUSDCDecimals is the number of decimals for USDC (6 decimals like native USDC)
	NobleUSDCDecimals = 6
)

// NobleChainConfig holds configuration for Noble blockchain integration
type NobleChainConfig struct {
	ChainID     string
	RPCEndpoint string
	GRPCEndpoint string
	USDCDenom   string
	Decimals    uint8
}

// GetNobleTestnetConfig returns the configuration for Noble testnet
func GetNobleTestnetConfig() NobleChainConfig {
	return NobleChainConfig{
		ChainID:      NobleTestnetChainID,
		RPCEndpoint:  "https://noble-testnet-rpc.polkachu.com:443",
		GRPCEndpoint: "noble-testnet-grpc.polkachu.com:21590",
		USDCDenom:    NobleUSDCDenom,
		Decimals:     NobleUSDCDecimals,
	}
}

// GetNobleMainnetConfig returns the configuration for Noble mainnet
func GetNobleMainnetConfig() NobleChainConfig {
	return NobleChainConfig{
		ChainID:      NobleMainnetChainID,
		RPCEndpoint:  "https://noble-rpc.polkachu.com:443",
		GRPCEndpoint: "noble-grpc.polkachu.com:21590",
		USDCDenom:    NobleUSDCDenom,
		Decimals:     NobleUSDCDecimals,
	}
}

// IsNobleChain checks if a chain ID corresponds to Noble (testnet or mainnet)
func IsNobleChain(chainID string) bool {
	return chainID == NobleTestnetChainID || chainID == NobleMainnetChainID
}

// ValidateNobleConnection validates that a connection ID is properly configured for Noble
func ValidateNobleConnection(connectionID string, allowedConnections []string) error {
	for _, allowed := range allowedConnections {
		if connectionID == allowed {
			return nil
		}
	}
	return fmt.Errorf("connection %s not in allowed connections list", connectionID)
}

// ConvertToUSDC converts an amount from base units to USDC representation
// For example: 1000000 base units = 1.000000 USDC
func ConvertToUSDC(amount sdk.Int) sdk.Dec {
	// Convert to decimal and divide by 10^6
	return sdk.NewDecFromInt(amount).QuoInt64(1000000)
}

// ConvertFromUSDC converts USDC amount to base units
// For example: 1.5 USDC = 1500000 base units
func ConvertFromUSDC(usdcAmount sdk.Dec) sdk.Int {
	// Multiply by 10^6 and truncate to integer
	return usdcAmount.MulInt64(1000000).TruncateInt()
}

// FormatUSDCAmount formats a USDC amount for display
// For example: 1500000 -> "1.500000 USDC"
func FormatUSDCAmount(amount sdk.Int) string {
	usdcDec := ConvertToUSDC(amount)
	return fmt.Sprintf("%s USDC", usdcDec.String())
}

// ParseUSDCAmount parses a string USDC amount to base units
// For example: "1.5" -> 1500000
func ParseUSDCAmount(amountStr string) (sdk.Int, error) {
	usdcDec, err := sdk.NewDecFromStr(amountStr)
	if err != nil {
		return sdk.ZeroInt(), fmt.Errorf("invalid USDC amount: %w", err)
	}

	if usdcDec.IsNegative() {
		return sdk.ZeroInt(), fmt.Errorf("USDC amount cannot be negative")
	}

	return ConvertFromUSDC(usdcDec), nil
}

// NobleSwapParams defines parameters for a swap on Noble or via Noble USDC
type NobleSwapParams struct {
	// InputDenom is the denomination being swapped from
	InputDenom string
	// OutputDenom is the denomination being swapped to
	OutputDenom string
	// Amount is the input amount in base units
	Amount sdk.Int
	// MinOutput is the minimum output amount (slippage protection)
	MinOutput sdk.Int
	// Receiver is the address to receive the output tokens
	Receiver string
}

// Validate performs basic validation on NobleSwapParams
func (p NobleSwapParams) Validate() error {
	if p.InputDenom == "" {
		return fmt.Errorf("input denom cannot be empty")
	}
	if p.OutputDenom == "" {
		return fmt.Errorf("output denom cannot be empty")
	}
	if p.InputDenom == p.OutputDenom {
		return fmt.Errorf("input and output denoms must be different")
	}
	if !p.Amount.IsPositive() {
		return fmt.Errorf("amount must be positive")
	}
	if !p.MinOutput.IsPositive() {
		return fmt.Errorf("min output must be positive")
	}
	if p.Receiver == "" {
		return fmt.Errorf("receiver cannot be empty")
	}

	// Validate receiver is a valid bech32 address
	_, err := sdk.AccAddressFromBech32(p.Receiver)
	if err != nil {
		return fmt.Errorf("invalid receiver address: %w", err)
	}

	return nil
}

// NobleLiquidityParams defines parameters for providing liquidity via Noble
type NobleLiquidityParams struct {
	// PoolID is the target pool identifier
	PoolID string
	// Token0 is the first token denomination
	Token0 string
	// Token1 is the second token denomination
	Token1 string
	// Amount0 is the amount of first token
	Amount0 sdk.Int
	// Amount1 is the amount of second token
	Amount1 sdk.Int
	// MinShares is the minimum LP shares to receive
	MinShares sdk.Int
}

// Validate performs basic validation on NobleLiquidityParams
func (p NobleLiquidityParams) Validate() error {
	if p.PoolID == "" {
		return fmt.Errorf("pool ID cannot be empty")
	}
	if p.Token0 == "" {
		return fmt.Errorf("token0 cannot be empty")
	}
	if p.Token1 == "" {
		return fmt.Errorf("token1 cannot be empty")
	}
	if !p.Amount0.IsPositive() {
		return fmt.Errorf("amount0 must be positive")
	}
	if !p.Amount1.IsPositive() {
		return fmt.Errorf("amount1 must be positive")
	}
	if !p.MinShares.IsPositive() {
		return fmt.Errorf("min shares must be positive")
	}

	return nil
}

// GetNobleUSDCPairs returns common trading pairs that include Noble USDC
// This is useful for routing swaps through USDC as an intermediary
func GetNobleUSDCPairs() []TradingPair {
	return []TradingPair{
		{Base: "uatom", Quote: NobleUSDCDenom, Description: "ATOM/USDC"},
		{Base: "uosmo", Quote: NobleUSDCDenom, Description: "OSMO/USDC"},
		{Base: "uakt", Quote: NobleUSDCDenom, Description: "AKT/USDC"},
		{Base: "ujuno", Quote: NobleUSDCDenom, Description: "JUNO/USDC"},
		{Base: "ustars", Quote: NobleUSDCDenom, Description: "STARS/USDC"},
	}
}

// TradingPair represents a trading pair on a DEX
type TradingPair struct {
	Base        string
	Quote       string
	Description string
}

// String returns a string representation of the trading pair
func (tp TradingPair) String() string {
	if tp.Description != "" {
		return tp.Description
	}
	return fmt.Sprintf("%s/%s", tp.Base, tp.Quote)
}

// Reverse returns the reverse trading pair (swaps base and quote)
func (tp TradingPair) Reverse() TradingPair {
	return TradingPair{
		Base:        tp.Quote,
		Quote:       tp.Base,
		Description: fmt.Sprintf("%s/%s", tp.Quote, tp.Base),
	}
}
