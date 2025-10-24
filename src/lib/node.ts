// Node operations utilities
import { join } from 'path';
import { Logger } from './logger';
import { TomlManager } from './toml';
import { DEFAULT_PORTS, CONSENSUS_TIMEOUTS } from './constants';
import type { NodeConfig } from '../types';

export class NodeManager {
  /**
   * Check if snrd binary is available
   */
  static async isBinaryAvailable(binary: string = 'snrd'): Promise<boolean> {
    try {
      const result = await Bun.$`which ${binary}`.quiet();
      return result.exitCode === 0;
    } catch {
      return false;
    }
  }

  /**
   * Get snrd version
   */
  static async getVersion(binary: string = 'snrd'): Promise<string> {
    try {
      const result = await Bun.$`${binary} version`.quiet();
      return (await result.text()).trim();
    } catch (error) {
      throw new Error(`Failed to get version: ${error}`);
    }
  }

  /**
   * Initialize a node
   */
  static async init(config: NodeConfig, binary: string = 'snrd'): Promise<void> {
    Logger.section(`Initializing ${config.nodeType} node: ${config.moniker}`);

    // Initialize chain
    Logger.step(1, 3, 'Initializing chain directory');
    await Bun.$`${binary} init ${config.moniker} --chain-id ${config.chainId} --home ${config.home}`;

    // Configure node
    Logger.step(2, 3, 'Configuring node settings');
    await this.configureNode(config);

    // Set VRF keys if needed
    Logger.step(3, 3, 'Generating VRF keypair');
    try {
      await this.generateVRFKey(config.home, binary);
    } catch (error) {
      Logger.warn('VRF key generation failed, but continuing...');
    }

    Logger.success(`Node initialized: ${config.moniker}`);
  }

  /**
   * Configure node settings
   */
  static async configureNode(config: NodeConfig): Promise<void> {
    const configToml = join(config.home, 'config', 'config.toml');
    const appToml = join(config.home, 'config', 'app.toml');

    // Configure RPC
    if (config.ports?.rpc) {
      await TomlManager.setValue(configToml, 'rpc.laddr', `tcp://0.0.0.0:${config.ports.rpc}`);
      await TomlManager.setValue(configToml, 'rpc.cors_allowed_origins', ['*']);
    }

    // Configure P2P
    await TomlManager.setValue(configToml, 'p2p.laddr', `tcp://0.0.0.0:${DEFAULT_PORTS.p2p}`);

    // Set consensus timeouts
    await TomlManager.setValue(configToml, 'consensus.timeout_propose', CONSENSUS_TIMEOUTS.propose);
    await TomlManager.setValue(configToml, 'consensus.timeout_prevote', CONSENSUS_TIMEOUTS.prevote);
    await TomlManager.setValue(configToml, 'consensus.timeout_precommit', CONSENSUS_TIMEOUTS.precommit);
    await TomlManager.setValue(configToml, 'consensus.timeout_commit', CONSENSUS_TIMEOUTS.commit);

    // Configure API
    if (config.ports?.rest) {
      await TomlManager.setValue(appToml, 'api.enable', true);
      await TomlManager.setValue(appToml, 'api.address', `tcp://0.0.0.0:${config.ports.rest}`);
      await TomlManager.setValue(appToml, 'api.enabled-unsafe-cors', true);
    }

    // Configure gRPC
    if (config.ports?.grpc) {
      await TomlManager.setValue(appToml, 'grpc.address', `0.0.0.0:${config.ports.grpc}`);
    }

    // Configure gRPC-Web
    if (config.ports?.grpcWeb) {
      await TomlManager.setValue(appToml, 'grpc-web.address', `0.0.0.0:${config.ports.grpcWeb}`);
    }

    // Configure JSON-RPC
    if (config.ports?.jsonRpc && config.ports?.jsonRpcWs) {
      await TomlManager.setValue(appToml, 'json-rpc.enable', true);
      await TomlManager.setValue(appToml, 'json-rpc.address', `0.0.0.0:${config.ports.jsonRpc}`);
      await TomlManager.setValue(appToml, 'json-rpc.ws-address', `0.0.0.0:${config.ports.jsonRpcWs}`);
      await TomlManager.setValue(appToml, 'json-rpc.api', 'eth,txpool,personal,net,debug,web3');
    }

    // Set pruning
    await TomlManager.setValue(appToml, 'pruning', 'nothing');

    // Set minimum gas prices
    await TomlManager.setValue(appToml, 'minimum-gas-prices', '0usnr');
  }

  /**
   * Generate VRF keypair
   */
  static async generateVRFKey(home: string, binary: string = 'snrd'): Promise<void> {
    try {
      await Bun.$`${binary} keys vrf generate --home ${home}`.quiet();
    } catch (error) {
      throw new Error(`Failed to generate VRF key: ${error}`);
    }
  }

  /**
   * Start a node
   */
  static async start(home: string, chainId: string, binary: string = 'snrd'): Promise<void> {
    Logger.info(`Starting node from: ${home}`);
    await Bun.$`${binary} start --home ${home} --chain-id ${chainId} --pruning nothing --minimum-gas-prices 0usnr`;
  }

  /**
   * Get node status
   */
  static async getStatus(rpcUrl: string = 'http://localhost:26657'): Promise<any> {
    try {
      const response = await fetch(`${rpcUrl}/status`);
      return await response.json();
    } catch (error) {
      throw new Error(`Failed to get node status: ${error}`);
    }
  }

  /**
   * Wait for node to be ready
   */
  static async waitForReady(rpcUrl: string = 'http://localhost:26657', maxAttempts: number = 30): Promise<void> {
    Logger.info('Waiting for node to be ready...');

    for (let i = 0; i < maxAttempts; i++) {
      try {
        const response = await fetch(`${rpcUrl}/health`);
        if (response.ok) {
          Logger.success('Node is ready!');
          return;
        }
      } catch {
        // Ignore errors and retry
      }

      await Bun.sleep(1000);
    }

    throw new Error('Node failed to become ready');
  }

  /**
   * Check if node is syncing
   */
  static async isSyncing(rpcUrl: string = 'http://localhost:26657'): Promise<boolean> {
    const status = await this.getStatus(rpcUrl);
    return status.result.sync_info.catching_up;
  }
}
