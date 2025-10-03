import type { ChainRegistryChainInfo } from '../types/ChainRegistryChainInfo';

export async function getChainRegistryChainInfo(chain: string): Promise<ChainRegistryChainInfo> {
  const res = await fetch(
    `https://raw.githubusercontent.com/cosmos/chain-registry/master/${chain}/chain.json`
  );
  return res.json();
}
