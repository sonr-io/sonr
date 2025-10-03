import type { PlainMessage } from '@bufbuild/protobuf';
// TODO: CosmosBankV1beta1MsgSend not available in protobufs
// // TODO: Missing from protobufs
// import { CosmosBankV1beta1MsgSend as ProtoMsgSend } from "@sonr.io/es/protobufs";
const _ProtoMsgSend: any = {};
type ProtoMsgSend = any;

import type { DeepPrettify } from '../../typeutils/prettify';
import type { Adapter } from './Adapter';

type Data = DeepPrettify<PlainMessage<ProtoMsgSend>>;

export class MsgSend implements Adapter {
  private readonly data: Data;
  private readonly legacy: boolean;

  constructor(data: Data, legacy = false) {
    this.data = data;
    this.legacy = legacy;
  }

  // TODO: Implement toProto() method when CosmosBankV1beta1MsgSend protobuf is available
  // This method should create and return a proper ProtoMsgSend instance with this.data
  // Required implementation:
  // 1. Import the correct protobuf type from @sonr.io/es/protobufs
  // 2. Create new ProtoMsgSend instance with validated data
  // 3. Set fromAddress, toAddress, and amount fields
  // 4. Handle coin conversion for amount field (denom and amount)
  // Currently returns empty object due to missing protobuf definition
  public toProto(): any {
    // TODO: Implement when ProtoMsgSend is available
    // throw new Error("MsgSend not implemented - missing protobuf definition");
    // return new ProtoMsgSend(this.data);
    return {} as any;
  }

  // TODO: Verify toAmino() implementation against latest Cosmos SDK amino encoding standards
  // This method converts the message to amino JSON format for legacy support
  public toAmino() {
    return {
      type: this.legacy ? 'bank/MsgSend' : 'cosmos-sdk/MsgSend',
      value: {
        from_address: this.data.fromAddress,
        to_address: this.data.toAddress,
        amount: this.data.amount,
      },
    };
  }
}
