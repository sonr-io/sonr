// Type definitions for sonrctl

export interface NodeConfig {
  moniker: string;
  chainId: string;
  home: string;
  nodeType: 'validator' | 'sentry' | 'full';
  ports?: {
    rpc?: number;
    rest?: number;
    grpc?: number;
    grpcWeb?: number;
    jsonRpc?: number;
    jsonRpcWs?: number;
  };
}

export interface NetworkConfig {
  name: string;
  chainId: string;
  validators: ValidatorConfig[];
  sentries: SentryConfig[];
}

export interface ValidatorConfig {
  name: string;
  moniker: string;
  network: string;
  home: string;
}

export interface SentryConfig {
  name: string;
  moniker: string;
  network: string;
  home: string;
  exposedPorts: boolean;
}

export interface DockerNode {
  name: string;
  image: string;
  command: string;
  environment: Record<string, string>;
  volumes: string[];
  networks: string[];
  ports?: string[];
}

export interface CommandContext {
  args: string[];
  options: Record<string, any>;
}
