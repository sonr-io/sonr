type Decr = [never, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10];

/**
 * Helper type to prettify type intersections.
 * @see https://twitter.com/mattpocockuk/status/1622730181453533185
 */
export type Prettify<T> = {
  [P in keyof T]: T[P];
} & unknown;

/**
 * Helper type to prettify deeply nested types.
 */
export type DeepPrettify<T, N extends number = 10> = N extends 0
  ? T
  : T extends Record<string | number | symbol, unknown> | unknown[]
    ? { [P in keyof T]: DeepPrettify<T[P], Decr[N]> }
    : T;
