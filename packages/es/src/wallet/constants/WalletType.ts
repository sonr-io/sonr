/**
 * The type of connection to the wallet.
 */
export const WalletType = {
  EXTENSION: 'extension',
  WALLETCONNECT: 'walletconnect',
} as const;
export type WalletType = (typeof WalletType)[keyof typeof WalletType];
