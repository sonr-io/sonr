// Package query provides query functionality for reading blockchain state.
package query

import (
	"context"
	"fmt"

	"google.golang.org/grpc"

	"github.com/cosmos/cosmos-sdk/types/query"
	authTypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/sonr-io/sonr/client/config"
	"github.com/sonr-io/sonr/client/errors"
)

// QueryClient provides an interface for querying blockchain state.
type QueryClient interface {
	// Chain information
	ChainInfo(ctx context.Context) (*ChainInfo, error)
	NodeInfo(ctx context.Context) (*NodeInfo, error)

	// Account queries
	Account(ctx context.Context, address string) (*authTypes.QueryAccountResponse, error)
	Accounts(ctx context.Context, pagination *PageRequest) (*authTypes.QueryAccountsResponse, error)

	// Balance queries
	Balance(ctx context.Context, address, denom string) (*banktypes.QueryBalanceResponse, error)
	AllBalances(ctx context.Context, address string, pagination *PageRequest) (*banktypes.QueryAllBalancesResponse, error)
	TotalSupply(ctx context.Context, pagination *PageRequest) (*banktypes.QueryTotalSupplyResponse, error)
	SupplyOf(ctx context.Context, denom string) (*banktypes.QuerySupplyOfResponse, error)

	// Staking queries
	Validators(ctx context.Context, status string, pagination *PageRequest) (*stakingtypes.QueryValidatorsResponse, error)
	Validator(ctx context.Context, validatorAddr string) (*stakingtypes.QueryValidatorResponse, error)
	Delegations(ctx context.Context, delegatorAddr string, pagination *PageRequest) (*stakingtypes.QueryDelegatorDelegationsResponse, error)
	Delegation(ctx context.Context, delegatorAddr, validatorAddr string) (*stakingtypes.QueryDelegationResponse, error)

	// Transaction queries
	Tx(ctx context.Context, hash string) (*TxResponse, error)
	TxsByEvents(ctx context.Context, events []string, pagination *PageRequest) (*TxSearchResponse, error)

	// Module-specific queries will be handled by module clients
}

// ChainInfo contains basic information about the blockchain.
type ChainInfo struct {
	ChainID            string `json:"chain_id"`
	BlockHeight        int64  `json:"block_height"`
	BlockTime          string `json:"block_time"`
	NodeVersion        string `json:"node_version"`
	ApplicationVersion string `json:"application_version"`
}

// NodeInfo contains information about the connected node.
type NodeInfo struct {
	NodeID     string `json:"node_id"`
	Network    string `json:"network"`
	Version    string `json:"version"`
	ListenAddr string `json:"listen_addr"`
	Moniker    string `json:"moniker"`
}

// TxResponse represents a transaction response.
type TxResponse struct {
	Hash      string  `json:"hash"`
	Height    int64   `json:"height"`
	Code      uint32  `json:"code"`
	Log       string  `json:"log"`
	GasWanted int64   `json:"gas_wanted"`
	GasUsed   int64   `json:"gas_used"`
	Events    []Event `json:"events"`
}

// Event represents a transaction event.
type Event struct {
	Type       string      `json:"type"`
	Attributes []Attribute `json:"attributes"`
}

// Attribute represents an event attribute.
type Attribute struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// TxSearchResponse represents a transaction search response.
type TxSearchResponse struct {
	Txs        []*TxResponse `json:"txs"`
	TotalCount int64         `json:"total_count"`
}

// PageRequest represents pagination parameters.
type PageRequest struct {
	Key        []byte `json:"key,omitempty"`
	Offset     uint64 `json:"offset,omitempty"`
	Limit      uint64 `json:"limit,omitempty"`
	CountTotal bool   `json:"count_total,omitempty"`
	Reverse    bool   `json:"reverse,omitempty"`
}

// queryClient implements QueryClient.
type queryClient struct {
	grpcConn *grpc.ClientConn
	config   *config.NetworkConfig

	// Cosmos SDK service clients
	authQueryClient    authTypes.QueryClient
	bankQueryClient    banktypes.QueryClient
	stakingQueryClient stakingtypes.QueryClient
}

// NewQueryClient creates a new query client.
func NewQueryClient(grpcConn *grpc.ClientConn, cfg *config.NetworkConfig) (QueryClient, error) {
	if grpcConn == nil {
		return nil, fmt.Errorf("gRPC connection is required")
	}

	if cfg == nil {
		return nil, fmt.Errorf("network configuration is required")
	}

	return &queryClient{
		grpcConn:           grpcConn,
		config:             cfg,
		authQueryClient:    authTypes.NewQueryClient(grpcConn),
		bankQueryClient:    banktypes.NewQueryClient(grpcConn),
		stakingQueryClient: stakingtypes.NewQueryClient(grpcConn),
	}, nil
}

// ChainInfo retrieves basic chain information.
func (qc *queryClient) ChainInfo(ctx context.Context) (*ChainInfo, error) {
	// TODO: Implement proper chain info query using Tendermint RPC client
	// Should query /status endpoint for current block height and time
	// Get node version and application version from /abci_info
	// Return comprehensive chain information with real-time data
	return &ChainInfo{
		ChainID:            qc.config.ChainID,
		NodeVersion:        "unknown",
		ApplicationVersion: "unknown",
	}, nil
}

// NodeInfo retrieves node information.
func (qc *queryClient) NodeInfo(ctx context.Context) (*NodeInfo, error) {
	// TODO: Implement proper node info query using Tendermint RPC client
	// Should query /status endpoint for node ID and network info
	// Get listen address and moniker from node status
	// Include peer count and sync status information
	return &NodeInfo{
		NodeID:  "unknown",
		Network: qc.config.ChainID,
		Version: "unknown",
	}, nil
}

// Account retrieves account information by address.
func (qc *queryClient) Account(ctx context.Context, address string) (*authTypes.QueryAccountResponse, error) {
	req := &authTypes.QueryAccountRequest{Address: address}

	resp, err := qc.authQueryClient.Account(ctx, req)
	if err != nil {
		return nil, errors.WrapError(err, errors.ErrQueryFailed, "failed to query account %s", address)
	}

	return resp, nil
}

// Accounts retrieves all accounts with pagination.
func (qc *queryClient) Accounts(ctx context.Context, pagination *PageRequest) (*authTypes.QueryAccountsResponse, error) {
	req := &authTypes.QueryAccountsRequest{}

	if pagination != nil {
		req.Pagination = &query.PageRequest{
			Key:        pagination.Key,
			Offset:     pagination.Offset,
			Limit:      pagination.Limit,
			CountTotal: pagination.CountTotal,
			Reverse:    pagination.Reverse,
		}
	}

	resp, err := qc.authQueryClient.Accounts(ctx, req)
	if err != nil {
		return nil, errors.WrapError(err, errors.ErrQueryFailed, "failed to query accounts")
	}

	return resp, nil
}

// Balance retrieves the balance of a specific denomination for an address.
func (qc *queryClient) Balance(ctx context.Context, address, denom string) (*banktypes.QueryBalanceResponse, error) {
	req := &banktypes.QueryBalanceRequest{
		Address: address,
		Denom:   denom,
	}

	resp, err := qc.bankQueryClient.Balance(ctx, req)
	if err != nil {
		return nil, errors.WrapError(err, errors.ErrQueryFailed, "failed to query balance for %s", address)
	}

	return resp, nil
}

// AllBalances retrieves all balances for an address.
func (qc *queryClient) AllBalances(ctx context.Context, address string, pagination *PageRequest) (*banktypes.QueryAllBalancesResponse, error) {
	req := &banktypes.QueryAllBalancesRequest{Address: address}

	if pagination != nil {
		req.Pagination = &query.PageRequest{
			Key:        pagination.Key,
			Offset:     pagination.Offset,
			Limit:      pagination.Limit,
			CountTotal: pagination.CountTotal,
			Reverse:    pagination.Reverse,
		}
	}

	resp, err := qc.bankQueryClient.AllBalances(ctx, req)
	if err != nil {
		return nil, errors.WrapError(err, errors.ErrQueryFailed, "failed to query all balances for %s", address)
	}

	return resp, nil
}

// TotalSupply retrieves the total supply of all denominations.
func (qc *queryClient) TotalSupply(ctx context.Context, pagination *PageRequest) (*banktypes.QueryTotalSupplyResponse, error) {
	req := &banktypes.QueryTotalSupplyRequest{}

	if pagination != nil {
		req.Pagination = &query.PageRequest{
			Key:        pagination.Key,
			Offset:     pagination.Offset,
			Limit:      pagination.Limit,
			CountTotal: pagination.CountTotal,
			Reverse:    pagination.Reverse,
		}
	}

	resp, err := qc.bankQueryClient.TotalSupply(ctx, req)
	if err != nil {
		return nil, errors.WrapError(err, errors.ErrQueryFailed, "failed to query total supply")
	}

	return resp, nil
}

// SupplyOf retrieves the supply of a specific denomination.
func (qc *queryClient) SupplyOf(ctx context.Context, denom string) (*banktypes.QuerySupplyOfResponse, error) {
	req := &banktypes.QuerySupplyOfRequest{Denom: denom}

	resp, err := qc.bankQueryClient.SupplyOf(ctx, req)
	if err != nil {
		return nil, errors.WrapError(err, errors.ErrQueryFailed, "failed to query supply of %s", denom)
	}

	return resp, nil
}

// Validators retrieves validators with optional status filter.
func (qc *queryClient) Validators(ctx context.Context, status string, pagination *PageRequest) (*stakingtypes.QueryValidatorsResponse, error) {
	req := &stakingtypes.QueryValidatorsRequest{Status: status}

	if pagination != nil {
		req.Pagination = &query.PageRequest{
			Key:        pagination.Key,
			Offset:     pagination.Offset,
			Limit:      pagination.Limit,
			CountTotal: pagination.CountTotal,
			Reverse:    pagination.Reverse,
		}
	}

	resp, err := qc.stakingQueryClient.Validators(ctx, req)
	if err != nil {
		return nil, errors.WrapError(err, errors.ErrQueryFailed, "failed to query validators")
	}

	return resp, nil
}

// Validator retrieves a specific validator by address.
func (qc *queryClient) Validator(ctx context.Context, validatorAddr string) (*stakingtypes.QueryValidatorResponse, error) {
	req := &stakingtypes.QueryValidatorRequest{ValidatorAddr: validatorAddr}

	resp, err := qc.stakingQueryClient.Validator(ctx, req)
	if err != nil {
		return nil, errors.WrapError(err, errors.ErrQueryFailed, "failed to query validator %s", validatorAddr)
	}

	return resp, nil
}

// Delegations retrieves all delegations for a delegator.
func (qc *queryClient) Delegations(ctx context.Context, delegatorAddr string, pagination *PageRequest) (*stakingtypes.QueryDelegatorDelegationsResponse, error) {
	req := &stakingtypes.QueryDelegatorDelegationsRequest{DelegatorAddr: delegatorAddr}

	if pagination != nil {
		req.Pagination = &query.PageRequest{
			Key:        pagination.Key,
			Offset:     pagination.Offset,
			Limit:      pagination.Limit,
			CountTotal: pagination.CountTotal,
			Reverse:    pagination.Reverse,
		}
	}

	resp, err := qc.stakingQueryClient.DelegatorDelegations(ctx, req)
	if err != nil {
		return nil, errors.WrapError(err, errors.ErrQueryFailed, "failed to query delegations for %s", delegatorAddr)
	}

	return resp, nil
}

// Delegation retrieves a specific delegation.
func (qc *queryClient) Delegation(ctx context.Context, delegatorAddr, validatorAddr string) (*stakingtypes.QueryDelegationResponse, error) {
	req := &stakingtypes.QueryDelegationRequest{
		DelegatorAddr: delegatorAddr,
		ValidatorAddr: validatorAddr,
	}

	resp, err := qc.stakingQueryClient.Delegation(ctx, req)
	if err != nil {
		return nil, errors.WrapError(err, errors.ErrQueryFailed, "failed to query delegation from %s to %s", delegatorAddr, validatorAddr)
	}

	return resp, nil
}

// Tx retrieves a transaction by hash.
func (qc *queryClient) Tx(ctx context.Context, hash string) (*TxResponse, error) {
	// TODO: Implement transaction query using Tendermint RPC client
	// Should query /tx endpoint with transaction hash
	// Parse transaction result and decode events
	// Return formatted transaction response with gas usage
	// Handle transaction not found errors gracefully
	return nil, fmt.Errorf("transaction queries not yet implemented")
}

// TxsByEvents retrieves transactions by events.
func (qc *queryClient) TxsByEvents(ctx context.Context, events []string, pagination *PageRequest) (*TxSearchResponse, error) {
	// TODO: Implement transaction search using Tendermint RPC client
	// Should query /tx_search endpoint with event filters
	// Support pagination with page and per_page parameters
	// Parse and format transaction results with event data
	// Handle complex event queries with AND/OR logic
	return nil, fmt.Errorf("transaction search not yet implemented")
}
