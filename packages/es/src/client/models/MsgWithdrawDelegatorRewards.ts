import type { PlainMessage } from '@bufbuild/protobuf';
// TODO: CosmosDistributionV1beta1MsgWithdrawDelegatorReward not available in protobufs
// import { CosmosDistributionV1beta1MsgWithdrawDelegatorReward as ProtoMsgWithdrawDelegatorRewards } from "@sonr.io/es/protobufs";
type ProtoMsgWithdrawDelegatorRewards = any;

import type { DeepPrettify } from '../../typeutils/prettify';
import type { Adapter } from './Adapter';

type Data = DeepPrettify<PlainMessage<ProtoMsgWithdrawDelegatorRewards>>;

export class MsgWithdrawDelegatorRewards implements Adapter {
  private readonly data: Data;
  private readonly isLegacy: boolean;

  constructor(data: Data, isLegacy = false) {
    this.data = data;
    this.isLegacy = isLegacy;
  }

  // TODO: Implement toProto() method when CosmosDistributionV1beta1MsgWithdrawDelegatorReward protobuf is available
  // This method should create and return a proper ProtoMsgWithdrawDelegatorRewards instance with this.data
  // Currently returns empty object due to missing protobuf definition
  public toProto(): any {
    // TODO: Implement when ProtoMsgWithdrawDelegatorRewards is available
    // throw new Error("MsgWithdrawDelegatorRewards not implemented - missing protobuf definition");
    // return new ProtoMsgWithdrawDelegatorRewards(this.data);
    return {} as any;
  }

  // TODO: Verify toAmino() implementation against latest Cosmos SDK amino encoding standards
  // This method converts the withdraw delegator rewards message to amino JSON format
  public toAmino() {
    return {
      type: this.isLegacy
        ? 'distribution/MsgWithdrawDelegationReward'
        : 'cosmos-sdk/MsgWithdrawDelegationReward',
      value: {
        validator_address: this.data.validatorAddress,
        delegator_address: this.data.delegatorAddress,
      },
    };
  }
}
