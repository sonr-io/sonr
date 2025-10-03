// TODO: This file needs CosmosBankV1beta1QueryAllBalancesService from protobufs
// which is not currently available. The functions are stubbed out until
// the protobuf definitions are added.

export type GetNativeBalancesParams = {
  address: string;
  pagination?: {
    key?: Uint8Array;
    offset?: bigint;
    limit?: bigint;
    countTotal?: boolean;
    reverse?: boolean;
  };
};

/**
 * Gets native balances for an address with pagination support.
 *
 * NOTE: This function currently has a placeholder implementation because
 * CosmosBankV1beta1QueryAllBalancesService is not available in the current
 * protobufs. It should be updated when bank module protobufs are added.
 */
// TODO: Implement getNativeBalances function when CosmosBankV1beta1QueryAllBalancesService protobuf is available
// This function should query all balances for a given address with pagination support
// Parameters: endpoint (RPC/API endpoint), params (address and pagination options)
// Returns: Array of coin balances with pagination metadata
export async function getNativeBalances(_endpoint: string, _params: GetNativeBalancesParams) {
  // TODO: Implement when CosmosBankV1beta1QueryAllBalancesService is available
  // Steps needed:
  // 1. Import CosmosBankV1beta1QueryAllBalancesService from protobufs
  // 2. Create query request with address and pagination from params
  // 3. Send query to endpoint
  // 4. Parse and return balances with pagination info
  throw new Error(
    'Bank module not available - CosmosBankV1beta1QueryAllBalancesService needs to be added to protobufs'
  );
}

// TODO: Implement getAllNativeBalances function when CosmosBankV1beta1QueryAllBalancesService protobuf is available
// This function should query all balances for a given address without pagination limits
// Parameters: endpoint (RPC/API endpoint), address (account address to query)
// Returns: Complete array of all coin balances for the address
export async function getAllNativeBalances(_endpoint: string, _address: string): Promise<any[]> {
  // TODO: Implement when CosmosBankV1beta1QueryAllBalancesService is available
  // Steps needed:
  // 1. Import CosmosBankV1beta1QueryAllBalancesService from protobufs
  // 2. Create query request with address, iterate through all pages
  // 3. Send multiple queries if needed to get all balances
  // 4. Aggregate and return complete balance list
  throw new Error(
    'Bank module not available - CosmosBankV1beta1QueryAllBalancesService needs to be added to protobufs'
  );
}
