import type { PlainMessage } from '@bufbuild/protobuf';
// TODO: CosmosStakingV1beta1MsgDelegate not available in protobufs
// // TODO: Missing from protobufs
// import { CosmosStakingV1beta1MsgDelegate as ProtoMsgDelegate } from "../../protobufs";
const _ProtoMsgDelegate: any = {};
type ProtoMsgDelegate = any;

import type { DeepPrettify } from '../../typeutils/prettify';
import type { Adapter } from './Adapter';

type Data = DeepPrettify<PlainMessage<ProtoMsgDelegate>>;

export class MsgDelegate implements Adapter {
  private readonly data: Data;

  constructor(data: Data) {
    this.data = data;
  }

  // TODO: Implement toProto() method when CosmosStakingV1beta1MsgDelegate protobuf is available
  // This method should create and return a proper ProtoMsgDelegate instance with this.data
  // Required implementation:
  // 1. Import the correct protobuf type from @sonr.io/es/protobufs
  // 2. Create new ProtoMsgDelegate instance with validated data
  // 3. Set delegatorAddress, validatorAddress, and amount fields
  // 4. Validate validator address format and amount positivity
  // 5. Handle coin conversion for amount field (denom and amount)
  // Currently returns empty object due to missing protobuf definition
  public toProto(): any {
    // TODO: Implement when ProtoMsgDelegate is available
    // throw new Error("MsgDelegate not implemented - missing protobuf definition");
    // return new ProtoMsgDelegate(this.data);
    return {} as any;
  }

  // TODO: Verify toAmino() implementation against latest Cosmos SDK amino encoding standards
  // This method converts the staking delegation message to amino JSON format
  public toAmino() {
    return {
      type: 'cosmos-sdk/MsgDelegate',
      value: {
        delegator_address: this.data.delegatorAddress,
        validator_address: this.data.validatorAddress,
        amount: this.data.amount,
      },
    };
  }
}
