// Package app provides ERC20 token pair configuration for the Sonr blockchain.
package app

import erc20types "github.com/cosmos/evm/x/erc20/types"

// WSonrTokenContractMainnet is the WrappedToken contract address for mainnet.
// This address represents the ERC20 wrapper for the native token.
const WSonrTokenContractMainnet = "0xD4949664cD82660AaE99bEdc034a0deA8A0bd517"

// WSonrTokenContractTestnet is the WrappedToken contract address for testnet.
// This address represents the ERC20 wrapper for the native token.
const WSonrTokenContractTestnet = "0xD4949664cD82660AaE99bEdc034a0deA8A0bd517"

// SonrETHTokenPairs creates a slice of token pairs that define the mapping between
// native Cosmos SDK coins and their ERC20 representations. This allows for seamless
// conversion between the two token standards within the EVM module.
var SonrETHTokenPairs = []erc20types.TokenPair{
	{
		Erc20Address:  WSonrTokenContractTestnet,
		Denom:         BaseDenom,
		Enabled:       true,
		ContractOwner: erc20types.OWNER_MODULE,
	},
}
