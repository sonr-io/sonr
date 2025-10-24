// Constants for sonrctl
import { homedir } from 'os';
import { join } from 'path';

export const DEFAULT_CHAIN_ID = 'sonrtest_1-1';
export const DEFAULT_DENOM = 'usnr';
export const DEFAULT_HOME_DIR = join(homedir(), '.sonr');
export const DEFAULT_KEYRING_BACKEND = 'test';
export const DEFAULT_BINARY = 'snrd';
export const DEFAULT_DOCKER_IMAGE = 'onsonr/snrd:latest';

export const DEFAULT_PORTS = {
  rpc: 26657,
  rest: 1317,
  grpc: 9090,
  grpcWeb: 9091,
  jsonRpc: 8545,
  jsonRpcWs: 8546,
  p2p: 26656,
};

export const CONSENSUS_TIMEOUTS = {
  propose: '5s',
  prevote: '1s',
  precommit: '1s',
  commit: '1s',
};

export const TESTNET_VALIDATORS = [
  { name: 'val-naruto', network: 'net-naruto' },
  { name: 'val-senku', network: 'net-senku' },
  { name: 'val-yaeger', network: 'net-yaeger' },
];

export const TESTNET_SENTRIES = [
  { name: 'sentry-naruto', network: 'net-naruto', validator: 'val-naruto' },
  { name: 'sentry-senku', network: 'net-senku', validator: 'val-senku' },
  { name: 'sentry-yaeger', network: 'net-yaeger', validator: 'val-yaeger' },
];
