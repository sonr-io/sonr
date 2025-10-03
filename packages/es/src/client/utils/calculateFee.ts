import {
  type CosmosBaseV1beta1Coin as Coin,
  CosmosTxV1beta1Fee as Fee,
  type CosmosBaseAbciV1beta1GasInfo as GasInfo,
} from '@sonr.io/es/protobufs';

/**
 * Estimates the fee for a transaction. For txs which uses more gas, the
 * `multiplier` can be decreased (default: `1.4`).
 */
export function calculateFee({ gasUsed }: GasInfo, { amount, denom }: Coin, multiplier = 1.4): Fee {
  const gasLimit = Number(gasUsed) * multiplier;
  return new Fee({
    amount: [
      {
        amount: Math.ceil(gasLimit * Number(amount)).toFixed(0),
        denom: denom,
      },
    ],
    gasLimit: BigInt(Math.floor(gasLimit)),
  });
}
