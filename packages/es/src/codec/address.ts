import { ripemd160 } from '@noble/hashes/ripemd160';
import { keccak_256 } from '@noble/hashes/sha3';
import { sha256 } from '@noble/hashes/sha256';
import { ProjectivePoint } from '@noble/secp256k1';
import { base64, bech32 } from '@scure/base';

import { ethhex } from './ethhex';

/**
 * Returns the bech32 address from the given `publicKey` and `prefix`. If needed,
 * the `type` of the key should be appropriately set.
 *
 * @param publicKey Must be either a base64 encoded string or a `Uint8Array`.
 */
export function resolveBech32Address(
  publicKey: string | Uint8Array,
  prefix: string,
  type: 'secp256k1' | 'ed25519' | 'ethsecp256k1' = 'secp256k1'
): string {
  const pubKey = typeof publicKey === 'string' ? base64.decode(publicKey) : publicKey;
  const address =
    type === 'secp256k1'
      ? // For cosmos: take the ripemd160 of the sha256 of the public key
        ripemd160(sha256(pubKey))
      : type === 'ed25519'
        ? // For cosmos: take the first 20 bytes of the sha256 of the public key
          sha256(pubKey).slice(0, 20)
        : // For eth: take the last 20 bytes of the keccak of the uncompressed public key without the first byte
          keccak_256(ProjectivePoint.fromHex(pubKey).toRawBytes(false).slice(1)).slice(-20);
  return bech32.encode(prefix, bech32.toWords(address));
}

/**
 * Translates the given ethereum address to a bech32 address.
 * @param ethAddress Must be a valid ethereum address (eg. `0x123...DeF`).
 */
export function translateEthToBech32Address(ethAddress: string, prefix: string) {
  const bytes = ethhex.decode(ethAddress);
  return bech32.encode(prefix, bech32.toWords(bytes));
}
