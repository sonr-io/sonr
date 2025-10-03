#!/usr/bin/env node
// @ts-check

/**
 * This is a custom plugin for `buf` that generates TS files from the services
 * defined in the proto files, and is referred to by the root `buf.gen.yaml`.
 * Files generated using this plugin contains the `_@sonr.io/es` suffix.
 *
 * Do not convert this to a TS file as it runs 4x slower!
 */

import { createEcmaScriptPlugin, runNodeJs } from "@bufbuild/protoplugin";
import {
  literalString,
  localName,
  makeJsDoc,
} from "@bufbuild/protoplugin/ecmascript";

export function generateTs(schema) {
  for (const protoFile of schema.files) {
    const file = schema.generateFile(protoFile.name + "_cosmes.ts");
    file.preamble(protoFile);
    for (const service of protoFile.services) {
      generateService(schema, file, service);
    }
  }
}

function generateService(schema, f, service) {
  f.print("const TYPE_NAME = ", literalString(service.typeName), ";");
  f.print("");
  for (const method of service.methods) {
    f.print(makeJsDoc(method));
    f.print("export const ", localName(service), method.name, "Service = {");
    f.print("  typeName: TYPE_NAME,");
    f.print("  method: ", literalString(method.name), ",");
    f.print("  Request: ", method.input, ",");
    f.print("  Response: ", method.output, ",");
    f.print("} as const;");
    f.print("");
  }
}

runNodeJs(
  createEcmaScriptPlugin({
    name: "protoc-gen-cosmes",
    version: "v0.0.1",
    generateTs,
  })
);
