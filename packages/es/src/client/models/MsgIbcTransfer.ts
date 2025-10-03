import type { PlainMessage } from '@bufbuild/protobuf';
import { IbcApplicationsTransferV1MsgTransfer as ProtoMsgIbcTransfer } from '../../protobufs';

import type { DeepPrettify } from '../../typeutils/prettify';
import type { Adapter } from './Adapter';

type Data = DeepPrettify<PlainMessage<ProtoMsgIbcTransfer>>;

export class MsgIbcTransfer implements Adapter {
  private readonly data: Data;

  constructor(data: Data) {
    this.data = data;
  }

  public toProto() {
    return new ProtoMsgIbcTransfer(this.data);
  }

  public toAmino() {
    return {
      type: 'cosmos-sdk/MsgTransfer',
      value: {
        source_port: this.data.sourcePort,
        source_channel: this.data.sourceChannel,
        token: this.data.token,
        sender: this.data.sender,
        receiver: this.data.receiver,
        /**
         * Protobuf type is optional, but Amino type is non-optional.
         *
         * @see https://github.com/cosmos/cosmjs/blob/358260bff71c9d3e7ad6644fcf64dc00325cdfb9/packages/stargate/src/modules/ibc/aminomessages.ts#L16-L42
         */
        timeout_height: this.data.timeoutHeight
          ? {
              revision_number: this.data.timeoutHeight.revisionNumber.toString(),
              revision_height: this.data.timeoutHeight.revisionHeight.toString(),
            }
          : {},
        timeout_timestamp: this.data.timeoutTimestamp.toString(),
        memo: this.data.memo,
      },
    };
  }
}
