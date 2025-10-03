import { describe, expect, it } from 'vitest';

import { extractExpectedAccountSequence } from './sequence';

describe('extractExpectedAccountSequence', () => {
  it('should extract expected account sequence numbers from valid errors', () => {
    // Older cosmos sdk chains
    const err1 = new Error(
      'account sequence mismatch, expected 10, got 11: incorrect account sequence: invalid request'
    );
    expect(extractExpectedAccountSequence(err1)).toBe(10n);

    // Newer cosmos sdk chains (0.45+)
    const err2 = new Error(
      'rpc error: code = Unknown desc = account sequence mismatch, expected 10, got 11: ......'
    );
    expect(extractExpectedAccountSequence(err2)).toBe(10n);

    // Injective
    const err3 = new Error(
      `[reason]:"incorrect account sequence" metadata:{key:"ABCICode" value:"32"} ...... rpc error: code = Unknown desc = account sequence mismatch, expected 10, got 11: ......`
    );
    expect(extractExpectedAccountSequence(err3)).toBe(10n);
  });
});
