import type { Prettify } from '../../typeutils/prettify';
import { RpcClient } from '../clients/RpcClient';
import type { ToSignedProtoParams, Tx } from '../models/Tx';

export type BroadcastTxParams = Prettify<
  ToSignedProtoParams & {
    tx: Tx;
  }
>;

/**
 * Broadcasts a tx to the network and returns the tx hash if successful.
 */
export async function broadcastTx(endpoint: string, { tx, ...params }: BroadcastTxParams) {
  return RpcClient.broadcastTx(endpoint, tx.toSignedProto(params));
}
