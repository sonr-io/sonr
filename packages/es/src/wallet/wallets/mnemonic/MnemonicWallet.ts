import { RpcClient, Secp256k1PubKey, Tx } from '@sonr.io/es/client';
import {
  base64,
  resolveBech32Address,
  resolveKeyPair,
  signAmino,
  signDirect,
  utf8,
} from '@sonr.io/es/codec';
import type { CosmosTxV1beta1Fee as Fee } from '@sonr.io/es/protobufs';
import type { StdSignDoc } from '@sonr.io/es/registry';

import type { Prettify } from '../../../typeutils/prettify';
import type { WalletName } from '../../constants/WalletName';
import type { WalletType } from '../../constants/WalletType';
import { ConnectedWallet, type SignArbitraryResponse, type UnsignedTx } from '../ConnectedWallet';
import type { ChainInfo } from '../WalletController';

export type ConnectMnemonicWalletOptions = Prettify<
  {
    /**
     * Also known as the 12-24 words seed phrase. **Warning: keep this safe!**
     */
    mnemonic: string;
    /**
     * The address prefix for the chain (eg. "osmo").
     */
    bech32Prefix: string;
    /**
     * Coin type number for HD derivation (default: `118`). For Terra chains, change
     * this to `330`.
     */
    coinType?: number | undefined;
    /**
     * Address index number for HD derivation (default: `0`).
     */
    index?: number | undefined;
  } & ChainInfo<string>
>;

/**
 * This wallet accepts a mnemonic (aka seed phrase) and is able to directly sign
 * and broadcast transactions to the chain without relying on an external wallet
 * like Keplr or Station. Use this if you want to programmatically broadcast
 * transactions. Unlike the other wallets, there is no Controller class and this
 * object must be instantiated directly.
 *
 * ```ts
 * // Example usage for Osmosis chain
 * const wallet = new MnemonicWallet({
 *   mnemonic: "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about", // Example 12-word mnemonic for testing
 *   bech32Prefix: "osmo",
 *   chainId: "osmosis-1",
 *   rpc: "https://rpc.osmosis.zone",
 *   gasPrice: {
 *     amount: "0.0025",
 *     denom: "uosmo",
 *   },
 *   coinType: 118, // optional (default: 118)
 *   index: 0, // optional (default: 0)
 * });
 * console.log("Address:", wallet.address); // prints the bech32 address
 *
 * // Sign an arbitrary message
 * const { signature } = await wallet.signArbitrary("Hello from @sonr.io/es!");
 * console.log("Signature:", signature);
 *
 * // Sign and broadcast a tx
 * const unsignedTx = { msgs: [...], memo: "Hello Cosmos!", timeoutHeight: 0n };
 * const res = await wallet.broadcastTxSync(unsignedTx, 1.4); // Uses 1.4x fee multiplier
 * console.log("Tx result:", res);
 * ```
 */
export class MnemonicWallet extends ConnectedWallet {
  public readonly publicKey: string;
  public readonly privateKey: Uint8Array;
  public readonly keyType: 'secp256k1' | 'ethsecp256k1';

  constructor({
    mnemonic,
    bech32Prefix,
    coinType,
    index,
    chainId,
    gasPrice,
    rpc,
  }: ConnectMnemonicWalletOptions) {
    const { publicKey, privateKey } = resolveKeyPair(mnemonic, {
      coinType,
      index,
    });
    const keyType =
      chainId.startsWith('injective') || chainId.startsWith('dymension')
        ? 'ethsecp256k1'
        : 'secp256k1';
    const address = resolveBech32Address(publicKey, bech32Prefix, keyType);
    super(
      // We typecast here instead of adding "mnemonic" to `WalletName` and
      // `WalletType` as this wallet is considered a special wallet that is
      // unlikely to be used by most consumers of @sonr.io/es.
      'mnemonic' as WalletName,
      'mnemonic' as WalletType,
      undefined,
      chainId,
      new Secp256k1PubKey({
        chainId,
        key: publicKey as any,
      }),
      address,
      rpc,
      gasPrice
    );
    this.publicKey = base64.encode(publicKey);
    this.privateKey = privateKey;
    this.keyType = keyType;
  }

  public async signArbitrary(data: string): Promise<SignArbitraryResponse> {
    // This sign doc follows ADR 036 specs.
    // See: https://github.com/cosmos/cosmos-sdk/blob/main/docs/architecture/adr-036-arbitrary-signature.md
    const doc: StdSignDoc = {
      chain_id: '',
      account_number: '0',
      sequence: '0',
      fee: {
        gas: '0',
        amount: [],
      },
      msgs: [
        {
          type: 'sign/MsgSignData',
          value: {
            signer: this.address,
            data: base64.encode(utf8.decode(data)),
          },
        },
      ],
      memo: '',
    };
    const signature = signAmino(doc, this.privateKey, this.keyType);
    return {
      data,
      pubKey: this.publicKey,
      signature: base64.encode(signature),
    };
  }

  public async signAndBroadcastTx(
    { msgs, memo, timeoutHeight }: UnsignedTx,
    fee: Fee,
    accountNumber: bigint,
    sequence: bigint
  ): Promise<string> {
    const tx = new Tx({
      chainId: this.chainId,
      pubKey: this.pubKey,
      msgs: msgs,
    });
    const doc = tx.toSignDoc({
      accountNumber,
      sequence,
      fee,
      memo,
      timeoutHeight,
    });
    const signature = signDirect(doc, this.privateKey, this.keyType);
    return RpcClient.broadcastTx(this.rpc, tx.toSignedDirect(doc, signature));
  }
}
