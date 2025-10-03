import type { PlainMessage } from '@bufbuild/protobuf';
import {
  type Adapter,
  type PollTxParams,
  type Secp256k1PubKey,
  Tx,
  calculateFee,
  getAccount,
  pollTx,
  simulateTx,
  toBaseAccount,
} from '@sonr.io/es/client';
import {
  CosmosBaseV1beta1Coin as Coin,
  type CosmosTxV1beta1Fee as Fee,
  type CosmosTxV1beta1GetTxResponse as GetTxResponse,
} from '@sonr.io/es/protobufs';

import type { WalletName } from '../constants/WalletName';
import type { WalletType } from '../constants/WalletType';
import { extractExpectedAccountSequence } from '../utils/sequence';

export type UnsignedTx = {
  msgs: Adapter[];
  memo?: string | undefined;
  timeoutHeight?: bigint | undefined;
};

export type PollTxOptions = Pick<PollTxParams, 'intervalSeconds' | 'maxAttempts'>;

export type SignArbitraryResponse = {
  data: string;
  pubKey: string;
  signature: string;
};

/**
 * Represents a connected wallet that is ready to sign transactions.
 * Use `WalletController` to create an instance of this class.
 */
export abstract class ConnectedWallet {
  /** The identifier of this wallet. */
  public readonly id: WalletName;
  /** The type of connection to the wallet. */
  public readonly type: WalletType;
  /** The user-defined label for this wallet, if any. */
  public readonly label: string | undefined;
  /** The chain ID this wallet is connected to. */
  public readonly chainId: string;
  /** The public key. */
  public readonly pubKey: Secp256k1PubKey;
  /** The bech32 address. */
  public readonly address: string;
  /** The RPC endpoint to use for interacting with the chain. */
  public readonly rpc: string;
  /** The gas price to use for transactions. */
  public readonly gasPrice: Coin;
  private accountNumber: bigint | undefined;
  private sequence: bigint | undefined;

  constructor(
    id: WalletName,
    type: WalletType,
    label: string | undefined,
    chainId: string,
    pubKey: Secp256k1PubKey,
    address: string,
    rpc: string,
    gasPrice: PlainMessage<Coin>
  ) {
    this.id = id;
    this.type = type;
    this.label = label;
    this.chainId = chainId;
    this.pubKey = pubKey;
    this.address = address;
    this.rpc = rpc;
    this.gasPrice = new Coin(gasPrice);
  }

  /**
   * Returns the account number and sequence for the connected address. If `fromCache`
   * is true, the cached values (if they are available) will be returned instead of
   * querying the auth module.
   *
   * @throws if the account does not exist in the auth module.
   */
  public async getAuthInfo(fromCache = false): Promise<{
    accountNumber: bigint;
    sequence: bigint;
  }> {
    if (!this.accountNumber || !this.sequence || !fromCache) {
      const account = await getAccount(this.rpc, { address: this.address });
      const { accountNumber, sequence } = toBaseAccount(account);
      this.accountNumber = accountNumber;
      this.sequence = sequence;
    }
    return {
      accountNumber: this.accountNumber,
      sequence: this.sequence,
    };
  }

  /**
   * Simulates the tx and returns an estimate of the gas fees required.
   *
   * @throws if the tx fails to simulate.
   */
  public async estimateFee({ msgs, memo }: UnsignedTx, feeMultiplier = 1.4): Promise<Fee> {
    const estimate = async () => {
      const { sequence } = await this.getAuthInfo(true);
      const { gasInfo } = await simulateTx(this.rpc, {
        sequence,
        memo,
        tx: new Tx({ chainId: this.chainId, pubKey: this.pubKey, msgs: msgs }),
      });
      if (!gasInfo) {
        throw new Error('Unable to estimate fee');
      }
      return calculateFee(gasInfo, this.gasPrice, feeMultiplier);
    };
    // If we encounter an account sequence mismatch error, we retry exactly once
    // by parsing the error for the correct sequence to use
    try {
      return await estimate();
    } catch (err) {
      if (!(err instanceof Error)) {
        // Rethrow non-errors
        throw err;
      }
      const expectedSequence = extractExpectedAccountSequence(err);
      if (!expectedSequence) {
        // Rethrow errors not related to account sequence mismatch
        throw err;
      }
      // Set the cached sequence to the one from the error message
      this.sequence = expectedSequence;
      return estimate();
    }
  }

  /**
   * Signs and broadcasts the given `unsignedTx`, returning the tx hash if successful.
   * The `fee` parameter can (and should) be obtained by running `estimateFee` on
   * the `unsignedTx` prior to calling this method.
   *
   * **Important**: successful execution of this method does not guarantee that the
   * tx was successfully included in a block. Use `pollTx` to poll for the result of
   * the tx.
   *
   * @throws if the user denies the signing of the tx.
   * @throws if the tx fails to broadcast.
   */
  public async broadcastTx(unsignedTx: UnsignedTx, fee: Fee): Promise<string> {
    const { accountNumber, sequence } = await this.getAuthInfo(true);
    const hash = await this.signAndBroadcastTx(unsignedTx, fee, accountNumber, sequence);
    // Greedily increment the sequence for the next tx. This may result in the wrong
    // sequence, but if `estimateFee` was called prior to this, it will be corrected
    this.sequence = sequence + 1n;
    return hash;
  }

  /**
   * Polls for the tx matching the given `txHash` every `intervalSeconds` until it is
   * included in a block or when `maxAttempts` is reached (default: 2s, 64 attempts).
   *
   * @throws if the tx is not included in a block after the given `maxAttempts`.
   */
  public async pollTx(
    txHash: string,
    { maxAttempts, intervalSeconds }: PollTxOptions = {}
  ): Promise<Required<PlainMessage<GetTxResponse>>> {
    return pollTx(this.rpc, {
      hash: txHash,
      maxAttempts,
      intervalSeconds,
    });
  }

  /**
   * Executes `broadcastTx` and `pollTx` sequentially, returning the result of the
   * tx. If `feeOrFeeMultiplier` is `undefined` or a number, an additional call to
   * `estimateFee` will be made. Use this if there is no need to independently
   * execute the three methods.
   */
  public async broadcastTxSync(
    unsignedTx: UnsignedTx,
    feeOrFeeMultiplier: Fee | number = 1.4,
    pollOpts: PollTxOptions = {}
  ): Promise<Required<PlainMessage<GetTxResponse>>> {
    const fee =
      typeof feeOrFeeMultiplier === 'number'
        ? await this.estimateFee(unsignedTx, feeOrFeeMultiplier)
        : feeOrFeeMultiplier;
    const txHash = await this.broadcastTx(unsignedTx, fee);
    return this.pollTx(txHash, pollOpts);
  }

  /**
   * Signs the UTF-8 encoded `data` string. Note that some mobile wallets do not
   * support this method.
   *
   * @throws if the user denies the signing of the data.
   * @throws if the wallet does not support signing arbitrary data.
   */
  public abstract signArbitrary(data: string): Promise<SignArbitraryResponse>;

  /**
   * Signs the given `unsignedTx` and broadcasts the resulting signed tx, returning
   * the hex encoded tx hash if successful. This abstract method should be implemented
   * by the concrete child classes.
   */
  protected abstract signAndBroadcastTx(
    unsignedTx: UnsignedTx,
    fee: Fee,
    accountNumber: bigint,
    sequence: bigint
  ): Promise<string>;
}
