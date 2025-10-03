import type { Message, PlainMessage } from '@bufbuild/protobuf';
import { base64 } from '@scure/base';
import {
  CosmosTxV1beta1AuthInfo as ProtoAuthInfo,
  CosmosTxV1beta1Fee as ProtoFee,
  CosmosTxV1beta1SignDoc as ProtoSignDoc,
  CosmosTxSigningV1beta1SignMode as ProtoSignMode,
  type CosmosTxV1beta1SignerInfo as ProtoSignerInfo,
  CosmosTxV1beta1TxBody as ProtoTxBody,
  CosmosTxV1beta1TxRaw as ProtoTxRaw,
} from '@sonr.io/es/protobufs';
import type { SignDoc, StdSignDoc } from '@sonr.io/es/registry';

import { toAny } from '../utils/toAny';
import type { Adapter } from './Adapter';
import type { Secp256k1PubKey } from './Secp256k1PubKey';

type Data = {
  chainId: string;
  pubKey: Secp256k1PubKey;
  msgs: Adapter[];
};

export type ToSignedProtoParams = {
  sequence: bigint;
  fee: ProtoFee;
  signMode: ProtoSignMode;
  signature: Uint8Array;
  memo?: string | undefined;
  timeoutHeight?: bigint | undefined;
  extensionOptions?: Message[] | undefined;
};

export type ToUnsignedProtoParams = Pick<
  ToSignedProtoParams,
  'sequence' | 'memo' | 'timeoutHeight'
>;

export type ToSignDocParams = {
  accountNumber: bigint;
  sequence: bigint;
  fee: ProtoFee;
  memo?: string | undefined;
  timeoutHeight?: bigint | undefined;
};

export type ToStdSignDocParams = ToSignDocParams;

export class Tx {
  private readonly data: Data;

  constructor(data: Data) {
    this.data = data;
  }

  /**
   * Returns the signed, proto-encoded tx, ready to be broadcasted. To create an
   * unsigned tx for the purpose of simulating it, use {@link toUnsignedProto}.
   */
  public toSignedProto({
    fee,
    sequence,
    signMode,
    signature,
    memo,
    timeoutHeight,
    extensionOptions,
  }: ToSignedProtoParams): ProtoTxRaw {
    return new ProtoTxRaw({
      authInfoBytes: new ProtoAuthInfo({
        fee: fee,
        signerInfos: [this.getSignerInfo(sequence, signMode)],
      }).toBinary() as any,
      bodyBytes: new ProtoTxBody({
        messages: this.data.msgs.map((m) => toAny(m.toProto())),
        memo: memo,
        timeoutHeight: timeoutHeight,
        extensionOptions: extensionOptions?.map(toAny),
      }).toBinary() as any,
      signatures: [signature],
    });
  }

  /**
   * Returns the proto-encoded tx with the sign mode set to `UNSPECIFIED`, useful
   * for simulating the tx. To create a signed tx, use {@link toSignedProto}.
   */
  public toUnsignedProto(info: ToUnsignedProtoParams): ProtoTxRaw {
    return this.toSignedProto({
      ...info,
      fee: new ProtoFee(),
      signMode: ProtoSignMode.UNSPECIFIED,
      signature: new Uint8Array(),
    });
  }

  /**
   * Combines the given `StdSignDoc` and `signature` and returns the proto-encoded
   * tx with sign mode set to `LEGACY_AMINO_JSON`, ready to be broadcasted.
   *
   * @param signature Must be a base64 encoded string or an `Uint8Array`
   */
  public toSignedAmino(
    { sequence, fee, memo, timeout_height }: StdSignDoc,
    signature: string | Uint8Array
  ): ProtoTxRaw {
    return this.toSignedProto({
      sequence: BigInt(sequence),
      fee: new ProtoFee({
        amount: fee.amount.slice(),
        gasLimit: BigInt(fee.gas),
        payer: fee.payer,
        granter: fee.granter,
      }),
      signMode: ProtoSignMode.LEGACY_AMINO_JSON,
      signature: typeof signature === 'string' ? base64.decode(signature) : signature,
      memo: memo,
      timeoutHeight: timeout_height ? BigInt(timeout_height) : undefined,
    });
  }

  /**
   * Combines the given `SignDoc` and `signature` and returns the proto-encoded tx,
   * ready to be broadcasted.
   *
   * @param signature Must be a base64 encoded string or an `Uint8Array`
   */
  public toSignedDirect(
    { bodyBytes, authInfoBytes }: SignDoc,
    signature: string | Uint8Array
  ): ProtoTxRaw {
    return new ProtoTxRaw({
      authInfoBytes: authInfoBytes as any,
      bodyBytes: bodyBytes as any,
      signatures: [typeof signature === 'string' ? base64.decode(signature) : signature],
    });
  }

  /**
   * Returns the unsigned, proto-encoded tx ready to be signed by a wallet.
   */
  public toSignDoc({
    accountNumber,
    sequence,
    fee,
    memo,
    timeoutHeight,
  }: ToSignDocParams): ProtoSignDoc {
    return new ProtoSignDoc({
      chainId: this.data.chainId,
      accountNumber: accountNumber,
      authInfoBytes: new ProtoAuthInfo({
        fee: fee,
        signerInfos: [this.getSignerInfo(sequence, ProtoSignMode.DIRECT)],
      }).toBinary() as any,
      bodyBytes: new ProtoTxBody({
        messages: this.data.msgs.map((m) => toAny(m.toProto())),
        memo: memo,
        timeoutHeight: timeoutHeight,
      }).toBinary() as any,
    });
  }

  /**
   * Returns the unsigned, amino-encoded tx ready to be signed by a wallet.
   */
  public toStdSignDoc({
    accountNumber,
    sequence,
    fee,
    memo,
    timeoutHeight,
  }: ToStdSignDocParams): StdSignDoc {
    return {
      chain_id: this.data.chainId,
      account_number: accountNumber.toString(),
      sequence: sequence.toString(),
      fee: {
        amount: fee.amount,
        gas: fee.gasLimit.toString(),
      },
      msgs: this.data.msgs.map((m) => m.toAmino()),
      memo: memo ?? '',
      timeout_height: timeoutHeight?.toString(),
    };
  }

  /**
   * Returns the signer info. The chain ID is used to determine if the public key
   * should be encoded using Injective's custom protobuf.
   *
   * **Warning**: Injective's chain ID might change, causing potential issues here.
   */
  private getSignerInfo(sequence: bigint, mode: ProtoSignMode): PlainMessage<ProtoSignerInfo> {
    return {
      publicKey: toAny(this.data.pubKey.toProto()),
      sequence: sequence,
      modeInfo: {
        sum: {
          case: 'single',
          value: {
            mode: mode,
          },
        },
      },
    };
  }
}
