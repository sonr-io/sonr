import { writeFileSync } from "fs";
import { compile } from "json-schema-to-typescript";
import { dirname, join } from "path";
import { fileURLToPath } from "url";

const __dirname = dirname(fileURLToPath(import.meta.url));

async function genChainRegistryChainInfo() {
  const tsName = "ChainRegistryChainInfo";
  const tsFile = tsName + ".ts";

  console.log("Retrieving JSON schema...");
  const res = await fetch(
    "https://raw.githubusercontent.com/cosmos/chain-registry/master/chain.schema.json"
  );
  const schema = await res.json();
  schema.title = tsName;

  console.log("Compiling JSON schema to TypeScript...");
  const types = await compile(schema, tsName, {
    // See: https://github.com/bcherny/json-schema-to-typescript?tab=readme-ov-file#options
    strictIndexSignatures: true,
  });

  const target = join(__dirname, "..", "src", "registry", "types", tsFile);
  writeFileSync(target, types);
  console.log("Wrote types to", target);
}

async function genChainRegistryAssetList() {
  const tsName = "ChainRegistryAssetList";
  const tsFile = tsName + ".ts";

  console.log("Retrieving JSON schema...");
  const res = await fetch(
    "https://raw.githubusercontent.com/cosmos/chain-registry/master/assetlist.schema.json"
  );
  const schema = await res.json();
  schema.title = tsName;

  console.log("Compiling JSON schema to TypeScript...");
  const types = await compile(schema, tsName, {
    // See: https://github.com/bcherny/json-schema-to-typescript?tab=readme-ov-file#options
    strictIndexSignatures: true,
  });

  const target = join(__dirname, "..", "src", "registry", "types", tsFile);
  writeFileSync(target, types);
  console.log("Wrote types to", target);
}

await genChainRegistryChainInfo();
await genChainRegistryAssetList();
