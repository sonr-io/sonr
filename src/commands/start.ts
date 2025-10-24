// Start command - Start nodes or network
import { Logger } from '../lib/logger';
import { ConfigManager } from '../lib/config';
import { DockerManager } from '../lib/docker';
import { NodeManager } from '../lib/node';
import { join } from 'path';

export async function start(args: string[]) {
  const subcommand = args[0];

  if (!subcommand) {
    Logger.error('Usage: sonrctl start <node-name|network>');
    Logger.info('');
    Logger.info('Examples:');
    Logger.info('  sonrctl start val-naruto              # Start a specific node');
    Logger.info('  sonrctl start network                 # Start entire testnet with docker-compose');
    Logger.info('  sonrctl start network --detach        # Start network in background');
    process.exit(1);
  }

  const config = new ConfigManager();

  try {
    if (subcommand === 'network') {
      await startNetwork(args.slice(1));
    } else {
      await startNode(subcommand, config);
    }
  } catch (error) {
    Logger.error(`Failed to start: ${error}`);
    process.exit(1);
  } finally {
    config.close();
  }
}

async function startNode(name: string, config: ConfigManager) {
  Logger.header(`Start Node: ${name}`);

  // Get node from registry
  const node = config.getNode(name);

  if (!node) {
    Logger.error(`Node '${name}' not found`);
    Logger.info('Initialize it first using: sonrctl init <type> <name>');
    Logger.info('');
    Logger.info('Registered nodes:');
    const nodes = config.listNodes();
    if (nodes.length === 0) {
      Logger.info('  (none)');
    } else {
      nodes.forEach((n: any) => Logger.info(`  - ${n.name} (${n.type})`));
    }
    process.exit(1);
  }

  const binary = config.get('binary') || 'snrd';

  Logger.info(`Starting ${node.type} node: ${node.moniker}`);
  Logger.table({
    Name: node.name,
    Type: node.type,
    'Chain ID': node.chain_id,
    Home: node.home,
  });

  // Check if binary is available
  const available = await NodeManager.isBinaryAvailable(binary);
  if (!available) {
    Logger.error(`Binary '${binary}' not found`);
    Logger.info('Install it using: sonrctl install');
    process.exit(1);
  }

  Logger.info('');
  Logger.info('Starting node...');
  Logger.info('Press Ctrl+C to stop');
  Logger.info('');

  // Start node (this will block)
  await NodeManager.start(node.home, node.chain_id, binary);
}

async function startNetwork(args: string[]) {
  Logger.header('Start Sonr Testnet Network');

  // Check if docker is available
  const dockerAvailable = await DockerManager.isAvailable();
  if (!dockerAvailable) {
    Logger.error('Docker is not available or not running');
    Logger.info('Install Docker: https://docs.docker.com/get-docker/');
    process.exit(1);
  }

  const composeAvailable = await DockerManager.isComposeAvailable();
  if (!composeAvailable) {
    Logger.error('docker-compose is not available');
    Logger.info('Install docker-compose: https://docs.docker.com/compose/install/');
    process.exit(1);
  }

  // Parse options
  const detached = args.includes('--detach') || args.includes('-d');

  // Find docker-compose file
  const cwd = process.cwd();
  const composePath = join(cwd, 'networks', 'testnet', 'docker-compose.yml');

  const composeExists = await Bun.file(composePath).exists();
  if (!composeExists) {
    Logger.error(`docker-compose.yml not found at: ${composePath}`);
    Logger.info('Make sure you are in the sonr repository root directory');
    process.exit(1);
  }

  Logger.info('Starting network with docker-compose...');
  Logger.info(`Compose file: ${composePath}`);
  Logger.info('');

  try {
    await DockerManager.composeUp(composePath, detached);

    if (detached) {
      Logger.success('Network started in background');
      Logger.info('');
      Logger.info('Check status with: docker ps');
      Logger.info('View logs with: docker logs -f <container-name>');
      Logger.info('Stop network with: sonrctl stop network');
    } else {
      Logger.success('Network started');
      Logger.info('Press Ctrl+C to stop');
    }

    // Show running containers
    Logger.info('');
    Logger.section('Running Containers');
    const containers = await DockerManager.listContainers('sonr-testnet');

    if (containers.length > 0) {
      for (const container of containers) {
        Logger.info(`  ${container.name}: ${container.status}`);
      }
    }
  } catch (error) {
    Logger.error(`Failed to start network: ${error}`);
    throw error;
  }
}
