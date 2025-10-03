import { wait } from '../utils/wait';
import { getTx } from './getTx';

export type PollTxParams = {
  hash: string;
  intervalSeconds?: number;
  maxAttempts?: number;
};

/**
 * Polls for the tx matching the given `hash`, with a minimum interval of
 * `intervalSeconds`. Throws if the tx is not found after the given number
 * of `maxAttempts`.
 */
export async function pollTx(
  endpoint: string,
  { intervalSeconds = 2, maxAttempts = 64, ...getTxParams }: PollTxParams
) {
  const intervalMillis = intervalSeconds * 1000;
  for (let i = 0; i < maxAttempts; i++) {
    try {
      return await getTx(endpoint, getTxParams);
    } catch (_err) {
      await wait(intervalMillis);
    }
  }
  throw new Error('Tx not found');
}
