import type { ChainRegistryAssetList } from '../types/ChainRegistryAssetList';

export async function getChainRegistryAssetList(chain: string): Promise<ChainRegistryAssetList> {
  const res = await fetch(
    `https://raw.githubusercontent.com/cosmos/chain-registry/master/${chain}/assetlist.json`
  );
  return res.json();
}
