import { describe, expect, it } from 'vitest';

import { sortObjectByKey } from './serialise';

describe('sortObjectByKey', () => {
  it('should sort keys correctly', () => {
    const obj = {
      zzz: 1,
      aaa: 1,
      xxx: null,
      bbb: {
        ttt: {
          ppp: true,
          iii: undefined,
          lll: '1',
        },
        ddd: [4, 8, 3, undefined, 4, 5, 7, 8],
      },
    };
    const expected = {
      aaa: 1,
      bbb: {
        ddd: [4, 8, 3, undefined, 4, 5, 7, 8], // arrays are not sorted
        ttt: {
          iii: undefined,
          lll: '1',
          ppp: true,
        },
      },
      xxx: null,
      zzz: 1,
    };
    // Before sorting, the stringified versions of the objects should NOT be equal
    expect(JSON.stringify(obj)).not.toBe(JSON.stringify(expected));
    // After sorting, the stringified versions of the objects should be equal
    expect(JSON.stringify(sortObjectByKey(obj))).toBe(JSON.stringify(expected));
  });
});
