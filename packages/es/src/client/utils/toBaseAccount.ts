import type { Any } from '@bufbuild/protobuf';
import {
  EthermintTypesV1EthAccount as EthermintAccount,
  IbcApplicationsInterchainAccountsV1InterchainAccount as InterchainAccount,
} from '@sonr.io/es/protobufs';

const ERR_UNKNOWN_ACCOUNT_TYPE = 'Unknown account type';
const ERR_UNABLE_TO_RESOLVE_BASE_ACCOUNT = 'Unable to resolve base account';

// Type definition for BaseAccount - this should match the cosmos auth BaseAccount structure
type BaseAccount = {
  address: string;
  pubKey?: any;
  accountNumber: bigint;
  sequence: bigint;
};

/**
 * Parses an `Any` protobuf message and returns the `BaseAccount`. Throws if unable
 * to parse correctly.
 *
 * NOTE: This function currently supports only the account types available in the
 * current protobufs. Missing types that should be added when protobufs are updated:
 * - cosmos.auth.v1beta1.BaseAccount
 * - cosmos.vesting.v1beta1.BaseVestingAccount
 * - cosmos.vesting.v1beta1.ContinuousVestingAccount
 * - cosmos.vesting.v1beta1.DelayedVestingAccount
 * - cosmos.vesting.v1beta1.PeriodicVestingAccount
 * - cosmos.auth.v1beta1.ModuleAccount
 * - cosmos.vesting.v1beta1.PermanentLockedAccount
 */
export function toBaseAccount({ typeUrl, value }: Any): BaseAccount {
  switch (typeUrl.slice(1)) {
    case EthermintAccount.typeName: {
      const { baseAccount } = EthermintAccount.fromBinary(value);
      if (!baseAccount) {
        throw new Error(ERR_UNABLE_TO_RESOLVE_BASE_ACCOUNT);
      }
      return baseAccount;
    }
    case InterchainAccount.typeName: {
      const { baseAccount } = InterchainAccount.fromBinary(value);
      if (!baseAccount) {
        throw new Error(ERR_UNABLE_TO_RESOLVE_BASE_ACCOUNT);
      }
      return baseAccount;
    }
    default: {
      throw new Error(`${ERR_UNKNOWN_ACCOUNT_TYPE}: ${typeUrl.slice(1)}`);
    }
  }
}
