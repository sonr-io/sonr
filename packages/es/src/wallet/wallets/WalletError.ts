/**
 * Custom error class which wraps around an error thrown by a wallet.
 */
export class WalletError extends Error {
  /**
   * Holds the original error and type thrown by the wallet.
   */
  public raw: unknown;

  constructor(message: string, raw: unknown) {
    super(message);
    this.name = 'WalletError';
    this.raw = raw;
  }

  /**
   * Returns the result of the `promise` if it resolves successfully, normalising
   * any errors thrown into a `WalletError` instance.
   *
   * It is best to wrap all wallet API calls with this function as some wallets
   * throw other data types other than actual `Error` instances.
   */
  public static async wrap<T>(promise: Promise<T>): Promise<T> {
    try {
      return await promise;
    } catch (err) {
      if (typeof err === 'string') {
        throw new WalletError(err, err);
      }
      if (WalletError.isRecord(err)) {
        // Takes into account normal error instances and objects with the 'error' key
        throw new WalletError(err.message ?? err.error ?? 'unknown error', err);
      }
      throw new WalletError('unknown error', err);
    }
  }

  private static isRecord(value: unknown): value is Record<string, string> {
    return typeof value === 'object' && value != null;
  }
}
