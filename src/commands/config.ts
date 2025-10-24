// Config command - Manage sonrctl configuration
import { Logger } from '../lib/logger';
import { ConfigManager } from '../lib/config';

export async function config(args: string[]) {
  const subcommand = args[0];

  if (!subcommand || !['get', 'set', 'list'].includes(subcommand)) {
    Logger.error('Usage: sonrctl config <get|set|list> [key] [value]');
    Logger.info('');
    Logger.info('Examples:');
    Logger.info('  sonrctl config list              # List all configuration');
    Logger.info('  sonrctl config get chain_id      # Get a specific value');
    Logger.info('  sonrctl config set home ~/.sonr  # Set a configuration value');
    process.exit(1);
  }

  const manager = new ConfigManager();

  try {
    if (subcommand === 'list') {
      await listConfig(manager);
    } else if (subcommand === 'get') {
      await getConfig(manager, args[1]);
    } else if (subcommand === 'set') {
      await setConfig(manager, args[1], args[2]);
    }
  } catch (error) {
    Logger.error(`Config operation failed: ${error}`);
    process.exit(1);
  } finally {
    manager.close();
  }
}

async function listConfig(manager: ConfigManager) {
  Logger.header('Sonrctl Configuration');

  const allConfig = manager.getAll();

  if (Object.keys(allConfig).length === 0) {
    Logger.info('No configuration set');
    return;
  }

  Logger.table(allConfig);
}

async function getConfig(manager: ConfigManager, key?: string) {
  if (!key) {
    Logger.error('Key is required');
    Logger.info('Usage: sonrctl config get <key>');
    process.exit(1);
  }

  const value = manager.get(key);

  if (value === null) {
    Logger.warn(`Configuration key '${key}' not found`);
    process.exit(1);
  }

  console.log(value);
}

async function setConfig(manager: ConfigManager, key?: string, value?: string) {
  if (!key || !value) {
    Logger.error('Key and value are required');
    Logger.info('Usage: sonrctl config set <key> <value>');
    process.exit(1);
  }

  manager.set(key, value);
  Logger.success(`Set ${key} = ${value}`);
}
