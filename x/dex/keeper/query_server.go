package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/sonr-io/sonr/x/dex/types"
)

var _ types.QueryServer = queryServer{}

type queryServer struct {
	Keeper
}

// NewQueryServerImpl returns an implementation of the module QueryServer.
func NewQueryServerImpl(k Keeper) types.QueryServer {
	return queryServer{Keeper: k}
}

// Params queries the module parameters.
func (qs queryServer) Params(ctx context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	params, err := qs.Keeper.Params.Get(sdkCtx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryParamsResponse{Params: params}, nil
}

// Account queries a specific DEX account.
func (qs queryServer) Account(ctx context.Context, req *types.QueryAccountRequest) (*types.QueryAccountResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	if req.Did == "" || req.ConnectionId == "" {
		return nil, status.Error(codes.InvalidArgument, "did and connection_id are required")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	account, err := qs.Keeper.GetDEXAccount(sdkCtx, req.Did, req.ConnectionId)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	return &types.QueryAccountResponse{Account: account}, nil
}

// Accounts queries all DEX accounts for a specific DID.
func (qs queryServer) Accounts(ctx context.Context, req *types.QueryAccountsRequest) (*types.QueryAccountsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	if req.Did == "" {
		return nil, status.Error(codes.InvalidArgument, "did is required")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	accounts, err := qs.Keeper.GetDEXAccountsByDID(sdkCtx, req.Did)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Convert to pointer slice for response
	accountPtrs := make([]*types.InterchainDEXAccount, len(accounts))
	for i := range accounts {
		accountPtrs[i] = &accounts[i]
	}

	return &types.QueryAccountsResponse{Accounts: accountPtrs}, nil
}

// TODO: Balance - Implement cross-chain balance query via IBC
// This method should query token balances on remote chains through IBC queries
// Required implementation steps:
// 1. Validate request parameters (DID, connection ID, denoms)
// 2. Retrieve the ICA account address for this DID and connection
// 3. Construct IBC query packet for bank balance on remote chain
// 4. Send IBC query through the appropriate channel
// 5. Parse the response and convert remote denoms to local representation
// 6. Cache balance data temporarily for performance optimization
// Returns: List of coin balances on the remote chain
// Balance queries remote chain balance.
func (qs queryServer) Balance(ctx context.Context, req *types.QueryBalanceRequest) (*types.QueryBalanceResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	// TODO: Implement balance query via ICA
	// This would require querying the remote chain through IBC
	return &types.QueryBalanceResponse{
		Balances: sdk.NewCoins(),
	}, nil
}

// TODO: Pool - Implement cross-chain liquidity pool query via IBC
// This method should query pool information from remote DEX protocols
// Required implementation steps:
// 1. Validate request parameters (pool ID, connection ID)
// 2. Construct IBC query packet for pool state on remote DEX
// 3. Send IBC query through the appropriate channel
// 4. Parse pool data including reserves, total shares, and fee parameters
// 5. Calculate derived metrics (price, APY, volume) if available
// 6. Cache pool data with appropriate TTL for performance
// Returns: Pool reserves, LP token supply, fee rate, and current price
// Pool queries pool information.
func (qs queryServer) Pool(ctx context.Context, req *types.QueryPoolRequest) (*types.QueryPoolResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	// TODO: Implement pool query via ICA
	// This would require querying the remote chain through IBC
	return &types.QueryPoolResponse{}, nil
}

// TODO: Orders - Implement order book query for user's limit orders
// This method should retrieve all orders for a specific DID across connections
// Required implementation steps:
// 1. Validate request parameters (DID, optional status filter)
// 2. Query local state for stored order records by DID
// 3. Filter orders by status (open, filled, cancelled) if specified
// 4. For open orders, optionally query remote chain for current status
// 5. Sort orders by creation time or specified sort parameter
// 6. Apply pagination if limits are provided
// 7. Include order fills and partial fill information
// Returns: List of orders with status, amounts, prices, and timestamps
// Orders queries orders for a DID.
func (qs queryServer) Orders(ctx context.Context, req *types.QueryOrdersRequest) (*types.QueryOrdersResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	// TODO: Implement orders query
	// This would require storing order information in state or DWN
	return &types.QueryOrdersResponse{
		Orders: []*types.Order{}, // Empty for now
	}, nil
}

// TODO: History - Implement transaction history query from DWN storage
// This method should retrieve complete transaction history for a DID
// Required implementation steps:
// 1. Validate request parameters (DID, time range, transaction type filter)
// 2. Query DWN for stored transaction records using DID as key
// 3. Filter transactions by type (swap, liquidity, order) if specified
// 4. Apply time range filter for date-based queries
// 5. Calculate profit/loss metrics for each transaction
// 6. Include gas costs and fees in transaction details
// 7. Sort by timestamp (newest first by default)
// 8. Apply pagination with cursor-based navigation
// Returns: List of transactions with full details and pagination info
// History queries transaction history.
func (qs queryServer) History(ctx context.Context, req *types.QueryHistoryRequest) (*types.QueryHistoryResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	// TODO: Implement history query
	// This would require storing transaction history in state or DWN
	return &types.QueryHistoryResponse{
		Transactions: []*types.Transaction{}, // Empty for now
	}, nil
}
