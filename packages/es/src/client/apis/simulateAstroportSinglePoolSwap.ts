import { queryContract } from './queryContract';

export type SimulateAstroportSinglePoolSwapParams = {
  poolId: string;
  fromAsset: string;
  fromAmount: bigint;
  isCW20?: boolean | undefined;
};

type Response = {
  return_amount: string;
  spread_amount: string;
  commission_amount: string;
};

/**
 * Simulates the amount of assets that would be received by swapping
 * `fromAmount` amount of `fromAsset` assets via the `poolId` pool.
 * If `fromAsset` is a CW20 token, `isCW20` must be set to `true`.
 */
export async function simulateAstroportSinglePoolSwap(
  endpoint: string,
  { poolId, fromAsset, fromAmount, isCW20 }: SimulateAstroportSinglePoolSwapParams
): Promise<bigint> {
  try {
    const { return_amount } = await queryContract<Response>(endpoint, {
      address: poolId,
      query: {
        simulation: {
          offer_asset: {
            info: isCW20
              ? { token: { contract_addr: fromAsset } }
              : { native_token: { denom: fromAsset } },
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
