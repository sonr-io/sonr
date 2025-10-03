import { queryContract } from './queryContract';

export type SimulateKujiraSinglePoolSwapParams = {
  poolId: string;
  fromAsset: string;
  fromAmount: bigint;
};

type Response = {
  return_amount: string;
  spread_amount: string;
  commission_amount: string;
};

/**
 * Simulates the amount of assets that would be received by swapping
 * `fromAmount` amount of `fromAsset` assets via the `poolId` pool.
 */
export async function simulateKujiraSinglePoolSwap(
  endpoint: string,
  { poolId, fromAsset, fromAmount }: SimulateKujiraSinglePoolSwapParams
): Promise<bigint> {
  try {
    const { return_amount } = await queryContract<Response>(endpoint, {
      address: poolId,
      query: {
        simulation: {
          offer_asset: {
            info: {
              native_token: {
                denom: fromAsset,
              },
            },
            amount: fromAmount.toString(),
          },
        },
      },
    });
    return BigInt(return_amount);
  } catch (err) {
    console.error(err);
    return 0n;
  }
}
