import { base64 } from '@scure/base';
import { describe, expect, it } from 'vitest';

import { resolveKeyPair } from './key';

// Randomly generated seed phrase
const SEED_PHRASE_1 =
  'witness snack faint milk gesture memory exhibit oak require mountain hammer crawl innocent day library drum youth result mutual remove capable hour front connect';

describe('resolveKeyPair', () => {
  it('should resolve 118 coin type correctly', () => {
    const { publicKey, privateKey } = resolveKeyPair(SEED_PHRASE_1);
    expect(base64.encode(publicKey)).toBe('AijdjMWZdjiXxSj0YCNbJHgnW6EsYwNyB9Yf7Wg5PcmE');
    expect(base64.encode(privateKey)).toBe('SojKJzJhFNruMSceBq3Imw3qZ+kS4p/6+iEpxdPsNg0=');
  });

  it('should resolve 330 coin type correctly', () => {
    const { publicKey, privateKey } = resolveKeyPair(SEED_PHRASE_1, {
      coinType: 330,
    });
    expect(base64.encode(publicKey)).toBe('A5G4nX2MIYCsnEdm40NJx7Bb1Z+oUNbEWWcVMssrgI3n');
    expect(base64.encode(privateKey)).toBe('kpvZKN+f7oWhVLLLk1pmKOazycgfECinugqQZgKRlXg=');
  });

  it('should resolve provided index correctly', () => {
    const { publicKey, privateKey } = resolveKeyPair(SEED_PHRASE_1, {
      index: 69,
    });
    expect(base64.encode(publicKey)).toBe('ArHwuHKnyiuPDbprTpWLVl3ZuomV70yzquzzlunGXlmj');
    expect(base64.encode(privateKey)).toBe('Fdm+CxL/KnM35bjDx/wg3eZc3tZN2q83I+xcY2wSMwk=');
  });
});
