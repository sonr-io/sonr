export { type BroadcastTxParams, broadcastTx } from './apis/broadcastTx';
export { type GetAccountParams, getAccount } from './apis/getAccount';
// Commented out - CosmWasm files have been removed
// export {
//   type GetCw20BalanceParams,
//   getCw20Balance,
// } from "./apis/getCw20Balance";
export {
  type GetNativeBalancesParams,
  getNativeBalances,
} from './apis/getNativeBalances';
export { type GetTxParams, getTx } from './apis/getTx';
export { type PollTxParams, pollTx } from './apis/pollTx';
// Commented out - CosmWasm files have been removed
// export { type QueryContractParams, queryContract } from "./apis/queryContract";
// Commented out - CosmWasm files have been removed
// export {
//   type SimulateAstroportSinglePoolSwapParams,
//   simulateAstroportSinglePoolSwap,
// } from "./apis/simulateAstroportSinglePoolSwap";
// export {
//   type SimulateKujiraSinglePoolSwapParams,
//   simulateKujiraSinglePoolSwap,
// } from "./apis/simulateKujiraSinglePoolSwap";
export { type SimulateTxParams, simulateTx } from './apis/simulateTx';
export { RpcClient } from './clients/RpcClient';
export type { Adapter } from './models/Adapter';
export { MsgBeginRedelegate } from './models/MsgBeginRedelegate';
export { MsgDelegate } from './models/MsgDelegate';
export { MsgIbcTransfer } from './models/MsgIbcTransfer';
export { MsgSend } from './models/MsgSend';
export { MsgStoreCode } from './models/MsgStoreCode';
export { MsgUndelegate } from './models/MsgUndelegate';
export { MsgWithdrawDelegatorRewards } from './models/MsgWithdrawDelegatorRewards';
export { MsgWithdrawValidatorCommission } from './models/MsgWithdrawValidatorCommission';

export { Secp256k1PubKey } from './models/Secp256k1PubKey';
export {
  type ToSignDocParams,
  type ToSignedProtoParams,
  type ToStdSignDocParams,
  type ToUnsignedProtoParams,
  Tx,
} from './models/Tx';
export { calculateFee } from './utils/calculateFee';
export { toAny } from './utils/toAny';
export { toBaseAccount } from './utils/toBaseAccount';

// Export passkey authentication functions
export {
  registerWithPasskey,
  loginWithPasskey,
  // Utility functions
  bufferToBase64url,
  base64urlToBuffer,
  isWebAuthnSupported,
  isWebAuthnAvailable,
  isConditionalMediationAvailable,
} from './auth';

// Export passkey types
export type {
  PasskeyRegistrationOptions,
  PasskeyLoginOptions,
  PasskeyRegistrationResult,
  PasskeyLoginResult,
} from './auth';
