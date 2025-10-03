import type { StdSignDoc } from '@keplr-wallet/types';
import { base16, base64, utf8 } from '@scure/base';
import { describe, expect, it } from 'vitest';

import { ethhex } from './ethhex';
import { hashEthArbitraryMessage, recoverPubKeyFromEthSignature, signAmino } from './sign';

describe('signAmino', () => {
  it('should sign Injective txs correctly', () => {
    const stdSignDoc: StdSignDoc = {
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
            signer: 'inj1l8w4vvmhcku28ryntpeazm37umshetzzl2gc33',
            data: base64.encode(
              utf8.decode(
                'Hi from CosmeES! This is a test message just to prove that the wallet is working.'
              )
            ),
          },
        },
      ],
      memo: '',
    };
    const privKey = base64.decode('o5di+2p2NdLgRYLtBIhJl9gsB9FWll8wKBaep3CmbI0=');
    const expected = // Signature taken from keplr signArbitrary
      'qrkZpuo1jpfXgbF3TtBtdR7DynE1nV3xd//bsGXm2FkS08waXeiJJ+FAvdtt9hvStyP/wGae07hxnyYPHEw+Uw==';
    const actual = base64.encode(signAmino(stdSignDoc, privKey, 'ethsecp256k1'));
    expect(actual).toStrictEqual(expected);
  });
});

describe('hashEthArbitraryMessage', () => {
  it('should hash correctly', () => {
    const msg = utf8.decode('Hello World!');
    const expected = hashEthArbitraryMessage(msg);
    const actual = ethhex.decode(
      '0xec3608877ecbf8084c29896b7eab2a368b2b3c8d003288584d145613dfa4706c'
    );
    expect(actual).toStrictEqual(expected);
  });
});

describe('recoverPubKeyFromEthSignature', () => {
  it('should recover public key correctly from a personal_sign signature', () => {
    const message = utf8.decode('Hello World');
    const signature = ethhex.decode(
      '0x63da4222cbcc36f43b22cbe417aa78963c29d088f7db3c9c6d06417dc34cf2df2dc6ffe9a5c9072a12a16a71c93bebf42bf388357aff81190d7dce166e4fa7ad1c'
    );
    const expected = base16.decode(
      '03f73842e6959e5b79f7979f81016e1e4f4d9481a7351a492ddb0807d98bb31f19'.toUpperCase()
    );
    const actual = recoverPubKeyFromEthSignature(message, signature);
    expect(expected).toStrictEqual(actual);
  });
});
