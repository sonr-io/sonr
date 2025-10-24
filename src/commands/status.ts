// Status command - Check node and network status
import { Logger } from '../lib/logger';
import { ConfigManager } from '../lib/config';
import { DockerManager } from '../lib/docker';
import { NodeManager } from '../lib/node';

export async function status(args: string[]) {
  const target = args[0];

  if (!target) {
    // Show overall status
    await showOverallStatus();
  } else if (target === 'network') {
    await showNetworkStatus();
  } else {
    await showNodeStatus(target);
  }
}

async function showOverallStatus() {
  Logger.header('Sonr Network Status');

  const config = new ConfigManager();

  // Show registered nodes
  Logger.section('Registered Nodes');
  const nodes = config.listNodes();

  if (nodes.length === 0) {
    Logger.info('No nodes registered');
    Logger.info('Initialize a node using: sonrctl init <type> <name>');
  } else {
    console.log();
    console.log('  Name               Type        Chain ID          Home');
    console.log('  ─────────────────  ──────────  ────────────────  ─────────────────────');

    for (const node of nodes) {
      const name = node.name.padEnd(17);
      const type = node.type.padEnd(10);
      const chainId = node.chain_id.padEnd(16);
      console.log(`  ${name}  ${type}  ${chainId}  ${node.home}`);
    }
    console.log();
  }

  // Show docker containers
  const dockerAvailable = await DockerManager.isAvailable();
  if (dockerAvailable) {
    Logger.section('Docker Containers');
    const containers = await DockerManager.listContainers();

    if (containers.length === 0) {
      Logger.info('No containers running');
    } else {
      console.log();
      console.log('  Name                    Status                    Image');
      console.log('  ──────────────────────  ────────────────────────  ────────────────────');

      for (const container of containers) {
        const name = container.name.padEnd(22);
        const status = container.status.substring(0, 24).padEnd(24);
        console.log(`  ${name}  ${status}  ${container.image}`);
      }
      console.log();
    }
  }

  config.close();
}

async function showNetworkStatus() {
  Logger.header('Testnet Network Status');

  const dockerAvailable = await DockerManager.isAvailable();
  if (!dockerAvailable) {
    Logger.error('Docker is not available');
    process.exit(1);
  }

  // Get all testnet containers
  const containers = await DockerManager.listContainers('sonr-testnet');

  if (containers.length === 0) {
    Logger.info('No network containers found');
    Logger.info('Start the network using: sonrctl start network');
    return;
  }

  Logger.section('Network Containers');
  console.log();

  for (const container of containers) {
    const { running, status } = await DockerManager.getContainerStatus(container.name);
    const statusIcon = running ? '✓' : '✗';
    const statusColor = running ? '\x1b[32m' : '\x1b[31m';

    console.log(`  ${statusColor}${statusIcon}\x1b[0m ${container.name}`);
    console.log(`     Status: ${status}`);
    console.log(`     Image:  ${container.image}`);

    // Try to get node status if running
    if (running) {
      try {
        const rpcPort = container.name.includes('sentry') ? '26657' : null;
        if (rpcPort) {
          const nodeStatus = await NodeManager.getStatus(`http://localhost:${rpcPort}`);
          const syncInfo = nodeStatus.result.sync_info;
          const catchingUp = syncInfo.catching_up ? 'Yes' : 'No';

          console.log(`     Height: ${syncInfo.latest_block_height}`);
          console.log(`     Syncing: ${catchingUp}`);
        }
      } catch {
        // Ignore errors getting node status
      }
    }
    console.log();
  }
}

async function showNodeStatus(name: string) {
  Logger.header(`Node Status: ${name}`);

  const config = new ConfigManager();
  const node = config.getNode(name);

  if (!node) {
    Logger.error(`Node '${name}' not found`);
    Logger.info('List registered nodes using: sonrctl status');
    config.close();
    process.exit(1);
  }

  // Show node info
  Logger.section('Node Information');
  Logger.table({
    Name: node.name,
    Type: node.type,
    'Chain ID': node.chain_id,
    Home: node.home,
    Moniker: node.moniker,
  });

  // Check docker container status
  const dockerAvailable = await DockerManager.isAvailable();
  if (dockerAvailable) {
    const { running, status } = await DockerManager.getContainerStatus(name);

    Logger.section('Container Status');
    Logger.info(`Running: ${running ? 'Yes' : 'No'}`);
    Logger.info(`Status: ${status}`);

    if (running) {
      // Try to get node status
      try {
        const nodeStatus = await NodeManager.getStatus('http://localhost:26657');
        const syncInfo = nodeStatus.result.sync_info;

        Logger.section('Blockchain Status');
        Logger.table({
          'Latest Height': syncInfo.latest_block_height,
          'Latest Block Time': new Date(syncInfo.latest_block_time).toLocaleString(),
          'Catching Up': syncInfo.catching_up ? 'Yes' : 'No',
        });
      } catch (error) {
        Logger.warn('Unable to fetch node status');
      }
    }
  }

  config.close();
}
