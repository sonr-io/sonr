import { type BytesCoder, hex } from '@scure/base';

/**
 * Convenience wrapper around `hex` that deals with hex strings typically
 * seen in Ethereum, where strings start with `0x` and are lower case.
 *
 * - For `encode`, the resulting string will be lower case
 * - For `decode`, the `str` arg can either be lower or upper case
 */
export const ethhex = {
  encode: (bytes) => `0x${hex.encode(bytes)}`,
  decode: (str) => hex.decode(str.replace(/^0x/, '').toLowerCase()),
} satisfies BytesCoder;
