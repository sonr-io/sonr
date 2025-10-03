import {
  type CosmosTxV1beta1GetTxResponse as GetTxResponse,
  CosmosTxV1beta1ServiceGetTxService as GetTxService,
} from '@sonr.io/es/protobufs';

import { RpcClient } from '../clients/RpcClient';

export type GetTxParams = {
  hash: string;
};

/**
 * Returns the tx matching the given hash. Throws if the tx is not found.
 */
export async function getTx(endpoint: string, params: GetTxParams) {
  const res = await RpcClient.query(endpoint, GetTxService, params);
  if (!res.tx || !res.txResponse) {
    throw new Error('Tx not found');
  }
  return res as Required<GetTxResponse>;
}
