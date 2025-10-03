import { base64, utf8 } from '@scure/base';
import { describe, expect, it } from 'vitest';

import { verifyADR36, verifyECDSA, verifyEIP191 } from './verify';

const DATA = utf8.decode(
  'Hi from CosmeES! This is a test message just to prove that the wallet is working.'
);
// Generated using coin type "330" and seed phrase "poverty flat amazing draw goose clay sorry nothing erase switch law intact only invest find memory what weasel fan connect tilt detect trap viable"
const VALID_PUBKEY_1 = base64.decode('Ai7ZXTtRWFte/tX7Z6MlKWVd9XA49p3cDNqd61RuKTdT');
// Generated using coin type "118" and seed phrase "poverty flat amazing draw goose clay sorry nothing erase switch law intact only invest find memory what weasel fan connect tilt detect trap viable"
const VALID_PUBKEY_2 = base64.decode('A8i9vMNUGcTtUgpbmiZqcFtsIrPZ6n8ZYN4/PVRlQvGr');
// Generated using coin type "60" and seed phrase "poverty flat amazing draw goose clay sorry nothing erase switch law intact only invest find memory what weasel fan connect tilt detect trap viable"
const VALID_PUBKEY_3 = base64.decode('AmGjuPKUsuIAuGgJ3xH7KGWlSU9cwVnsesrwWwyYLbMg');

describe('verifyECDSA', () => {
  it('should verify correctly', () => {
    // Signed using Station wallet on Terra
    const signature = base64.decode(
      'Od87qNoOyXuDOVdLCGTXB6dFN7U0XF9Oegc8KDa+AWwX3jkrDXG++2nlPfsF4VJzlDHsoikPeZpxrB7v9PINnw=='
    );
    const res1 = verifyECDSA({
      pubKey: VALID_PUBKEY_1,
      data: DATA,
      signature,
    });
    expect(res1).toBe(true);

    // Different pub key
    const res2 = verifyECDSA({
      pubKey: VALID_PUBKEY_2,
      data: DATA,
      signature,
    });
    expect(res2).toBe(false);
  });
});

describe('verifyADR36', () => {
  it('should verify correctly', () => {
    // Signed using Keplr wallet on Osmosis
    const signature = base64.decode(
      'nvlcV0x0Ge8ADXLSAQGtfMw6EJkOfpmkDxgP7UI79uR8MhnAOp9T+e+ofgW9kY4bEIr0yhyBG+vSVAZRv9uCxA=='
    );
    const res1 = verifyADR36({
      bech32Prefix: 'osmo',
      pubKey: VALID_PUBKEY_2,
      data: DATA,
      signature,
    });
    expect(res1).toBe(true);

    // Different bech32 prefix
    const res2 = verifyADR36({
      bech32Prefix: 'terra',
      pubKey: VALID_PUBKEY_2,
      data: DATA,
      signature,
    });
    expect(res2).toBe(false);

    // Different pub key
    const res3 = verifyADR36({
      bech32Prefix: 'osmo',
      pubKey: VALID_PUBKEY_1,
      data: DATA,
      signature,
    });
    expect(res3).toBe(false);
  });

  it('should verify ethsecp256k1 type signatures correctly', () => {
    // Signed using Keplr wallet on Injective
    const signature = base64.decode(
      '+7PNZm4XxKtpvZA+HqxpMKJgZcqA2w3WVSheLGvzrrIBJZGOTdcpBT7wLUhluY46EokTeRRWUaBDSv2vVoEdfw=='
    );
    const res1 = verifyADR36({
      bech32Prefix: 'inj',
      pubKey: VALID_PUBKEY_3,
      data: DATA,
      signature,
      type: 'ethsecp256k1',
    });
    expect(res1).toBe(true);
  });
});

describe('verifyEIP191', () => {
  it('should verify correctly', () => {
    // Signed using MetaMask wallet on Injective
    const signature = base64.decode(
      'MpriWY0Kq7C+/jR3eOfNB5vUQM144tQk7KkzKyYCTFB5QHGLZjzJyeOSr8/ENFES0k+aaEF47Wepk7OHoZuLzxs='
    );
    const res1 = verifyEIP191({
      pubKey: VALID_PUBKEY_3,
      data: DATA,
      signature,
    });
    expect(res1).toBe(true);

    // Different pub key
    const res2 = verifyEIP191({
      pubKey: VALID_PUBKEY_2,
      data: DATA,
      signature,
    });
    expect(res2).toBe(false);
  });
});
