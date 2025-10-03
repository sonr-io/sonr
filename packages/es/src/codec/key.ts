import { HDKey } from '@scure/bip32';
import { mnemonicToSeedSync } from '@scure/bip39';

/**
 * Resolves the given `mnemonic` (aka 12-24 words seed phrase) to its public and
 * private key pair. Derivation path uses the default for Cosmos chains - provide
 * the optional `opts` to override.
 */
export function resolveKeyPair(
  mnemonic: string,
  opts?: { coinType?: number | undefined; index?: number | undefined } | undefined
): {
  publicKey: Uint8Array;
  privateKey: Uint8Array;
} {
  const seed = mnemonicToSeedSync(mnemonic);
  const { publicKey, privateKey } = HDKey.fromMasterSeed(seed).derive(
    `m/44'/${opts?.coinType ?? 118}'/0'/0/${opts?.index ?? 0}`
  );
  if (!publicKey || !privateKey) {
    throw new Error('invalid mnemonic');
  }
  return {
    publicKey,
    privateKey,
  };
}
