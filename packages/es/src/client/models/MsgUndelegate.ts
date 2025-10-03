import type { PlainMessage } from '@bufbuild/protobuf';
// TODO: CosmosStakingV1beta1MsgUndelegate not available in protobufs
// // TODO: Missing from protobufs
// import { CosmosStakingV1beta1MsgUndelegate as ProtoMsgUndelegate } from "@sonr.io/es/protobufs";
const _ProtoMsgUndelegate: any = {};
type ProtoMsgUndelegate = any;

import type { DeepPrettify } from '../../typeutils/prettify';
import type { Adapter } from './Adapter';

type Data = DeepPrettify<PlainMessage<ProtoMsgUndelegate>>;

export class MsgUndelegate implements Adapter {
  private readonly data: Data;

  constructor(data: Data) {
    this.data = data;
  }

  // TODO: Implement toProto() method when CosmosStakingV1beta1MsgUndelegate protobuf is available
  // This method should create and return a proper ProtoMsgUndelegate instance with this.data
  // Currently returns empty object due to missing protobuf definition
  public toProto(): any {
    // TODO: Implement when ProtoMsgUndelegate is available
    // throw new Error("MsgUndelegate not implemented - missing protobuf definition");
    // return new ProtoMsgUndelegate(this.data);
    return {} as any;
  }

  // TODO: Verify toAmino() implementation against latest Cosmos SDK amino encoding standards
  // This method converts the undelegation message to amino JSON format
  public toAmino() {
    return {
      type: 'cosmos-sdk/MsgUndelegate',
      value: {
        delegator_address: this.data.delegatorAddress,
        validator_address: this.data.validatorAddress,
        amount: this.data.amount,
      },
    };
  }
}
