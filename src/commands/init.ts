// Init command - Initialize node configurations
import { Logger } from '../lib/logger';
import { ConfigManager } from '../lib/config';
import { NodeManager } from '../lib/node';
import { DEFAULT_CHAIN_ID, DEFAULT_HOME_DIR, DEFAULT_PORTS } from '../lib/constants';
import type { NodeConfig } from '../types';

export async function init(args: string[]) {
  const subcommand = args[0];

  if (!subcommand || !['validator', 'sentry', 'full'].includes(subcommand)) {
    Logger.error('Usage: sonrctl init <validator|sentry|full> <name> [options]');
    Logger.info('');
    Logger.info('Examples:');
    Logger.info('  sonrctl init validator val-naruto');
    Logger.info('  sonrctl init sentry sentry-naruto --home ~/.sonr/sentry');
    Logger.info('  sonrctl init full my-node --chain-id sonrtest_1-1');
    process.exit(1);
  }

  const name = args[1];
  if (!name) {
    Logger.error('Node name is required');
    Logger.info('Usage: sonrctl init <validator|sentry|full> <name>');
    process.exit(1);
  }

  // Parse options
  const options: Record<string, any> = {};
  for (let i = 2; i < args.length; i += 2) {
    const key = args[i].replace(/^--/, '');
    const value = args[i + 1];
    options[key] = value;
  }

  const config = new ConfigManager();

  // Build node configuration
  const nodeConfig: NodeConfig = {
    moniker: name,
    chainId: options['chain-id'] || config.get('chain_id') || DEFAULT_CHAIN_ID,
    home: options.home || `${DEFAULT_HOME_DIR}/${name}`,
    nodeType: subcommand as 'validator' | 'sentry' | 'full',
    ports: {
      rpc: parseInt(options['rpc-port']) || DEFAULT_PORTS.rpc,
      rest: parseInt(options['rest-port']) || DEFAULT_PORTS.rest,
      grpc: parseInt(options['grpc-port']) || DEFAULT_PORTS.grpc,
      grpcWeb: parseInt(options['grpc-web-port']) || DEFAULT_PORTS.grpcWeb,
      jsonRpc: parseInt(options['json-rpc-port']) || DEFAULT_PORTS.jsonRpc,
      jsonRpcWs: parseInt(options['json-rpc-ws-port']) || DEFAULT_PORTS.jsonRpcWs,
    },
  };

  Logger.header(`Initialize ${nodeConfig.nodeType.toUpperCase()} Node`);

  try {
    // Check if binary is available
    const binary = config.get('binary') || 'snrd';
    const available = await NodeManager.isBinaryAvailable(binary);

    if (!available) {
      Logger.error(`Binary '${binary}' not found`);
      Logger.info('Install it using: sonrctl install');
      process.exit(1);
    }

    // Check if home directory already exists
    const homeExists = await Bun.file(nodeConfig.home).exists();
    if (homeExists) {
      Logger.warn(`Home directory already exists: ${nodeConfig.home}`);
      const response = await confirmPrompt('Overwrite existing configuration?');

      if (!response) {
        Logger.info('Initialization cancelled');
        process.exit(0);
      }

      // Remove existing directory
      await Bun.$`rm -rf ${nodeConfig.home}`;
    }

    // Initialize node
    await NodeManager.init(nodeConfig, binary);

    // Register node in database
    config.addNode(name, nodeConfig.nodeType, nodeConfig.chainId, nodeConfig.home, nodeConfig.moniker);

    Logger.success('Node initialized successfully!');
    Logger.table({
      Name: name,
      Type: nodeConfig.nodeType,
      'Chain ID': nodeConfig.chainId,
      Home: nodeConfig.home,
      Moniker: nodeConfig.moniker,
    });

    Logger.info('');
    Logger.info('Next steps:');
    Logger.info(`  1. Configure genesis: ${nodeConfig.home}/config/genesis.json`);
    Logger.info(`  2. Start the node: sonrctl start ${name}`);
    Logger.info(`  3. Check status: sonrctl status ${name}`);
  } catch (error) {
    Logger.error(`Initialization failed: ${error}`);
    process.exit(1);
  } finally {
    config.close();
  }
}

// Simple confirmation prompt
async function confirmPrompt(message: string): Promise<boolean> {
  process.stdout.write(`${message} (y/N): `);

  for await (const line of console) {
    const answer = line.trim().toLowerCase();
    return answer === 'y' || answer === 'yes';
  }

  return false;
}
