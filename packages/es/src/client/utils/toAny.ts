import { Any, type Message } from '@bufbuild/protobuf';

export function toAny(msg: Message): Any {
  return new Any({
    typeUrl: `/${msg.getType().typeName}`,
    value: msg.toBinary(),
  });
}
