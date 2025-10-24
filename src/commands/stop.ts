// Stop command - Stop nodes or network
import { Logger } from '../lib/logger';
import { ConfigManager } from '../lib/config';
import { DockerManager } from '../lib/docker';
import { join } from 'path';

export async function stop(args: string[]) {
  const target = args[0];

  if (!target) {
    Logger.error('Usage: sonrctl stop <node-name|network|all>');
    Logger.info('');
    Logger.info('Examples:');
    Logger.info('  sonrctl stop val-naruto    # Stop a specific node container');
    Logger.info('  sonrctl stop network       # Stop entire testnet network');
    Logger.info('  sonrctl stop all           # Stop all running containers');
    process.exit(1);
  }

  try {
    if (target === 'network') {
      await stopNetwork();
    } else if (target === 'all') {
      await stopAll();
    } else {
      await stopNode(target);
    }
  } catch (error) {
    Logger.error(`Failed to stop: ${error}`);
    process.exit(1);
  }
}

async function stopNode(name: string) {
  Logger.header(`Stop Node: ${name}`);

  const dockerAvailable = await DockerManager.isAvailable();
  if (!dockerAvailable) {
    Logger.error('Docker is not available');
    process.exit(1);
  }

  // Check if container is running
  const running = await DockerManager.isContainerRunning(name);
  if (!running) {
    Logger.warn(`Container '${name}' is not running`);
    return;
  }

  await DockerManager.stopContainer(name);
}

async function stopNetwork() {
  Logger.header('Stop Sonr Testnet Network');

  const dockerAvailable = await DockerManager.isAvailable();
  if (!dockerAvailable) {
    Logger.error('Docker is not available');
    process.exit(1);
  }

  // Find docker-compose file
  const cwd = process.cwd();
  const composePath = join(cwd, 'networks', 'testnet', 'docker-compose.yml');

  const composeExists = await Bun.file(composePath).exists();
  if (!composeExists) {
    Logger.error(`docker-compose.yml not found at: ${composePath}`);
    Logger.info('Make sure you are in the sonr repository root directory');
    process.exit(1);
  }

  await DockerManager.composeDown(composePath);
}

async function stopAll() {
  Logger.header('Stop All Containers');

  const dockerAvailable = await DockerManager.isAvailable();
  if (!dockerAvailable) {
    Logger.error('Docker is not available');
    process.exit(1);
  }

  // Get all containers
  const containers = await DockerManager.listContainers();

  if (containers.length === 0) {
    Logger.info('No running containers found');
    return;
  }

  Logger.info(`Found ${containers.length} containers`);

  for (const container of containers) {
    if (container.status.startsWith('Up')) {
      Logger.info(`Stopping ${container.name}...`);
      await DockerManager.stopContainer(container.name);
    }
  }

  Logger.success('All containers stopped');
}
