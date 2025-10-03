import { hmac } from '@noble/hashes/hmac';
import { keccak_256 } from '@noble/hashes/sha3';
import { sha256 } from '@noble/hashes/sha256';
import * as secp256k1 from '@noble/secp256k1';
import { utf8 } from '@scure/base';
import type { CosmosTxV1beta1SignDoc as SignDoc } from '@sonr.io/es/protobufs';
import type { StdSignDoc } from '@sonr.io/es/registry';

import { serialiseSignDoc } from './serialise';

function sign(
  bytes: Uint8Array,
  privateKey: Uint8Array,
  type: 'secp256k1' | 'ethsecp256k1'
): Uint8Array {
  // Required polyfills for secp256k1 that must be called before any sign ops.
  // See: https://github.com/paulmillr/noble-secp256k1?tab=readme-ov-file#usage
  secp256k1.etc.hmacSha256Sync = (k, ...m) => hmac(sha256, k, secp256k1.etc.concatBytes(...m));
  const hash = type === 'secp256k1' ? sha256(bytes) : keccak_256(bytes);
  return secp256k1.sign(hash, privateKey).toCompactRawBytes();
}

/**
 * Signs the given amino-encoded `stdSignDoc` with the given `privateKey` using
 * secp256k1, and returns the signature bytes. For Injective, the `type` param
 * must be set to `ethsecp256k1`.
 */
export function signAmino(
  stdSignDoc: StdSignDoc,
  privateKey: Uint8Array,
  type: 'secp256k1' | 'ethsecp256k1' = 'secp256k1'
): Uint8Array {
  return sign(serialiseSignDoc(stdSignDoc), privateKey, type);
}

/**
 * Signs the given proto-encoded `signDoc` with the given `privateKey` using
 * secp256k1, and returns the signature bytes. For Injective, the `type` param
 * must be set to `ethsecp256k1`.
 */
export function signDirect(
  signDoc: SignDoc,
  privateKey: Uint8Array,
  type: 'secp256k1' | 'ethsecp256k1' = 'secp256k1'
): Uint8Array {
  return sign(signDoc.toBinary(), privateKey, type);
}

/**
 * Hashes and returns the digest of the given EIP191 `message` bytes.
 */
export function hashEthArbitraryMessage(message: Uint8Array): Uint8Array {
  return keccak_256(
    Uint8Array.from([
      ...utf8.decode('\x19Ethereum Signed Message:\n'),
      ...utf8.decode(message.length.toString()),
      ...message,
    ])
  );
}

/**
 * Recovers and returns the secp256k1 public key of the signer given the arbitrary
 * `message` and `signature` that was signed using EIP191.
 */
export function recoverPubKeyFromEthSignature(
  message: Uint8Array,
  signature: Uint8Array
): Uint8Array {
  if (signature.length !== 65) {
    throw new Error('Invalid signature');
  }
  const r = signature.slice(0, 32);
  const s = signature.slice(32, 64);
  const v = signature[64];
  // Adapted from https://github.com/ethers-io/ethers.js/blob/6017d3d39a4d428793bddae33d82fd814cacd878/src.ts/crypto/signature.ts#L255-L265
  const yParity = v <= 1 ? v : (v + 1) % 2;
  const secpSignature = secp256k1.Signature.fromCompact(
    Uint8Array.from([...r, ...s])
  ).addRecoveryBit(yParity);
  const digest = hashEthArbitraryMessage(message);
  return secpSignature.recoverPublicKey(digest).toRawBytes(true);
}
