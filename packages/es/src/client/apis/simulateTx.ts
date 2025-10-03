import { CosmosTxV1beta1ServiceSimulateService as SimulateService } from '@sonr.io/es/protobufs';

import type { Prettify } from '../../typeutils/prettify';
import { RpcClient } from '../clients/RpcClient';
import type { ToUnsignedProtoParams, Tx } from '../models/Tx';

export type SimulateTxParams = Prettify<
  ToUnsignedProtoParams & {
    tx: Tx;
  }
>;

/**
 * Simulates a tx for the purpose of estimating gas fees.
 */
export async function simulateTx(endpoint: string, { tx, ...params }: SimulateTxParams) {
  return RpcClient.query(endpoint, SimulateService, {
    txBytes: tx.toUnsignedProto(params).toBinary() as any,
  });
}
