import type { PlainMessage } from '@bufbuild/protobuf';
// TODO: CosmosDistributionV1beta1MsgWithdrawValidatorCommission not available in protobufs
// import { CosmosDistributionV1beta1MsgWithdrawValidatorCommission as ProtoMsgWithdrawValidatorCommission } from "@sonr.io/es/protobufs";
type ProtoMsgWithdrawValidatorCommission = any;

import type { DeepPrettify } from '../../typeutils/prettify';
import type { Adapter } from './Adapter';

type Data = DeepPrettify<PlainMessage<ProtoMsgWithdrawValidatorCommission>>;

export class MsgWithdrawValidatorCommission implements Adapter {
  private readonly data: Data;

  constructor(data: Data) {
    this.data = data;
  }

  // TODO: Implement toProto() method when CosmosDistributionV1beta1MsgWithdrawValidatorCommission protobuf is available
  // This method should create and return a proper ProtoMsgWithdrawValidatorCommission instance with this.data
  // Currently returns empty object due to missing protobuf definition
  public toProto(): any {
    // TODO: Implement when ProtoMsgWithdrawValidatorCommission is available
    // throw new Error("MsgWithdrawValidatorCommission not implemented - missing protobuf definition");
    // return new ProtoMsgWithdrawValidatorCommission(this.data);
    return {} as any;
  }

  // TODO: Verify toAmino() implementation against latest Cosmos SDK amino encoding standards
  // This method converts the withdraw validator commission message to amino JSON format
  public toAmino() {
    return {
      type: 'cosmos-sdk/MsgWithdrawValidatorCommission',
      value: {
        validator_address: this.data.validatorAddress,
      },
    };
  }
}
