import type { JsonValue, Message, PartialMessage } from '@bufbuild/protobuf';
import { base16, base64 } from '@sonr.io/es/codec';
import type { CosmosTxV1beta1TxRaw as TxRaw } from '@sonr.io/es/protobufs';

import { FetchClient } from './FetchClient';

type ErrorResponse = {
  id: number;
  jsonrpc: string;
  error: {
    code: number;
    message: string;
    data: string;
  };
  result: never;
};
type SuccessResponse<T> = {
  id: number;
  jsonrpc: string;
  result: T;
  error: never;
};
type Response<T> = SuccessResponse<T> | ErrorResponse;

type QueryResult = {
  response: {
    code: number;
    log: string;
    info: string;
    index: string;
    key: string | null;
    value: string | null;
    proofOps: string[] | null;
    height: string;
    codespace: string;
  };
};
type QueryService<T extends Message<T>, U extends Message<U>> = {
  typeName: string;
  method: string;
  Request: new (msg: PartialMessage<T>) => T;
  Response: { fromBinary: (bytes: Uint8Array) => U };
};

type BroadcastTxResult = {
  code: number;
  codespace: string;
  data: string;
  hash: string;
  log: string;
};

/**
 * Wraps the request message with an optional `height` field.
 */
type RequestMessage<T extends Message<T>> = T extends {
  height: bigint | string | number;
}
  ? PartialMessage<T>
  : PartialMessage<T> & {
      /**
       * The block height at which the query should be executed. Providing a height
       * that is outside the range of the full node will result in an error. Leave
       * this field empty to default to the latest block.
       */
      height?: number | undefined;
    };

export class RpcClient {
  private static async doRequest<T>(endpoint: string, method: string, params: JsonValue) {
    const { result, error } = await FetchClient.post<Response<T>>(endpoint, {
      id: Date.now(),
      jsonrpc: '2.0',
      method,
      params,
    });
    if (error != null) {
      throw new Error(error.data);
    }
    return result;
  }

  /**
   * Posts an ABCI query to the RPC `endpoint`. If successful, returns the response,
   * otherwise throws an error.
   */
  public static async query<T extends Message<T>, U extends Message<U>>(
    endpoint: string,
    { typeName, method, Request, Response }: QueryService<T, U>,
    requestMsg: RequestMessage<T>
  ): Promise<U> {
    const { response } = await RpcClient.doRequest<QueryResult>(endpoint, 'abci_query', {
      path: `/${typeName}/${method}`,
      data: base16.encode(new Request(requestMsg).toBinary()),
      ...(requestMsg.height ? { height: requestMsg.height.toString() } : {}),
    });
    const { log, value } = response;
    if (!value) {
      throw new Error(log);
    }
    return Response.fromBinary(base64.decode(value));
  }

  /**
   * Posts a `broadcast_tx_sync` request to the RPC `endpoint`. If successful,
   * returns the tx hash, otherwise throws an error.
   */
  public static async broadcastTx(endpoint: string, txRaw: TxRaw): Promise<string> {
    const { code, log, hash } = await RpcClient.doRequest<BroadcastTxResult>(
      endpoint,
      'broadcast_tx_sync',
      {
        tx: base64.encode(txRaw.toBinary()),
      }
    );
    if (code !== 0) {
      throw new Error(log);
    }
    return hash;
  }

  /**
   * Creates a new ABCI batch query.
   */
  public static newBatchQuery(endpoint: string): BatchQuery {
    return new BatchQuery(endpoint);
  }
}

class BatchQuery {
  private readonly endpoint: string;
  private readonly queries: {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    queryService: QueryService<any, any>;
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    requestMsg: RequestMessage<any>;
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    callback: (err: Error | null, response: any) => unknown;
  }[] = [];

  constructor(endpoint: string) {
    this.endpoint = endpoint;
  }

  /**
   * Adds an `abci_query` to this query batch.
   *
   * @param callback An error-first callback function for the response of the query.
   * If `err` is not `null`, `response` will be `null` and should not be used.
   */
  public add<T extends Message<T>, U extends Message<U>>(
    queryService: QueryService<T, U>,
    requestMsg: RequestMessage<T>,
    callback: (err: Error | null, response: U) => unknown
  ) {
    this.queries.push({ queryService, requestMsg, callback });
    return this;
  }

  /**
   * Executes the batched query.
   */
  public async send() {
    if (this.queries.length === 0) {
      return;
    }
    const payload = this.queries.map(({ queryService, requestMsg }, idx) => ({
      id: idx,
      jsonrpc: '2.0',
      method: 'abci_query',
      params: {
        path: `/${queryService.typeName}/${queryService.method}`,
        data: base16.encode(new queryService.Request(requestMsg).toBinary()),
        ...(requestMsg.height ? { height: requestMsg.height.toString() } : {}),
      },
    }));
    const res = await FetchClient.post<
      // Array is returned if and only if the payload has more than one query
      Response<QueryResult>[] | Response<QueryResult>
    >(this.endpoint, payload);
    const results = Array.isArray(res) ? res : [res];
    for (const { id, result, error } of results) {
      const query = this.queries[id];
      if (!query) continue;
      const { queryService, callback: handler } = query;
      if (error != null) {
        handler(new Error(error.data), null);
        continue;
      }
      const { log, value } = result.response;
      if (!value) {
        handler(new Error(log), null);
        continue;
      }
      const responseMsg = queryService.Response.fromBinary(base64.decode(value));
      handler(null, responseMsg);
    }
  }
}
