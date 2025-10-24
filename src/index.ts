#!/usr/bin/env bun

// Sonrctl - Sonr Network Control CLI
import { Logger } from './lib/logger';
import { install } from './commands/install';
import { init } from './commands/init';
import { start } from './commands/start';
import { stop } from './commands/stop';
import { status } from './commands/status';
import { config } from './commands/config';

const VERSION = '0.1.0';

// Main CLI router
async function main() {
  const args = process.argv.slice(2);

  if (args.length === 0 || args[0] === 'help' || args[0] === '--help' || args[0] === '-h') {
    showHelp();
    process.exit(0);
  }

  if (args[0] === 'version' || args[0] === '--version' || args[0] === '-v') {
    console.log(VERSION);
    process.exit(0);
  }

  const command = args[0];
  const commandArgs = args.slice(1);

  try {
    switch (command) {
      case 'install':
        await install(commandArgs);
        break;

      case 'init':
        await init(commandArgs);
        break;

      case 'start':
        await start(commandArgs);
        break;

      case 'stop':
        await stop(commandArgs);
        break;

      case 'status':
        await status(commandArgs);
        break;

      case 'config':
        await config(commandArgs);
        break;

      default:
        Logger.error(`Unknown command: ${command}`);
        Logger.info('Run "sonrctl help" for usage information');
        process.exit(1);
    }
  } catch (error) {
    Logger.error(`Command failed: ${error}`);
    process.exit(1);
  }
}

function showHelp() {
  const COLORS = {
    reset: '\x1b[0m',
    bright: '\x1b[1m',
    cyan: '\x1b[36m',
    blue: '\x1b[34m',
    gray: '\x1b[90m',
  };

  console.log(`
${COLORS.bright}${COLORS.cyan}sonrctl${COLORS.reset} - Sonr Network Control CLI

${COLORS.bright}USAGE${COLORS.reset}
  ${COLORS.cyan}sonrctl${COLORS.reset} <command> [options]

${COLORS.bright}COMMANDS${COLORS.reset}
  ${COLORS.cyan}install${COLORS.reset} [version]
      Install or update the snrd binary
      ${COLORS.gray}Example: sonrctl install latest${COLORS.reset}

  ${COLORS.cyan}init${COLORS.reset} <validator|sentry|full> <name> [options]
      Initialize a new node configuration
      ${COLORS.gray}Example: sonrctl init validator val-naruto${COLORS.reset}
      ${COLORS.gray}Options:${COLORS.reset}
      ${COLORS.gray}  --chain-id <id>        Chain ID (default: sonrtest_1-1)${COLORS.reset}
      ${COLORS.gray}  --home <path>          Home directory${COLORS.reset}
      ${COLORS.gray}  --rpc-port <port>      RPC port (default: 26657)${COLORS.reset}
      ${COLORS.gray}  --rest-port <port>     REST API port (default: 1317)${COLORS.reset}
      ${COLORS.gray}  --grpc-port <port>     gRPC port (default: 9090)${COLORS.reset}

  ${COLORS.cyan}start${COLORS.reset} <node-name|network> [options]
      Start a node or the entire network
      ${COLORS.gray}Example: sonrctl start val-naruto${COLORS.reset}
      ${COLORS.gray}Example: sonrctl start network${COLORS.reset}
      ${COLORS.gray}Options:${COLORS.reset}
      ${COLORS.gray}  --detach, -d           Run in background${COLORS.reset}

  ${COLORS.cyan}stop${COLORS.reset} <node-name|network|all>
      Stop a node, network, or all containers
      ${COLORS.gray}Example: sonrctl stop val-naruto${COLORS.reset}
      ${COLORS.gray}Example: sonrctl stop network${COLORS.reset}

  ${COLORS.cyan}status${COLORS.reset} [node-name|network]
      Check status of nodes and network
      ${COLORS.gray}Example: sonrctl status${COLORS.reset}
      ${COLORS.gray}Example: sonrctl status val-naruto${COLORS.reset}

  ${COLORS.cyan}config${COLORS.reset} <get|set|list> [key] [value]
      Manage sonrctl configuration
      ${COLORS.gray}Example: sonrctl config list${COLORS.reset}
      ${COLORS.gray}Example: sonrctl config get chain_id${COLORS.reset}
      ${COLORS.gray}Example: sonrctl config set home ~/.sonr${COLORS.reset}

  ${COLORS.cyan}help${COLORS.reset}
      Show this help message

  ${COLORS.cyan}version${COLORS.reset}
      Show version information

${COLORS.bright}GLOBAL OPTIONS${COLORS.reset}
  ${COLORS.gray}-h, --help       Show help${COLORS.reset}
  ${COLORS.gray}-v, --version    Show version${COLORS.reset}

${COLORS.bright}EXAMPLES${COLORS.reset}
  ${COLORS.gray}# Install the snrd binary${COLORS.reset}
  ${COLORS.blue}sonrctl install${COLORS.reset}

  ${COLORS.gray}# Initialize a validator node${COLORS.reset}
  ${COLORS.blue}sonrctl init validator val-naruto --chain-id sonrtest_1-1${COLORS.reset}

  ${COLORS.gray}# Start the entire testnet network${COLORS.reset}
  ${COLORS.blue}sonrctl start network --detach${COLORS.reset}

  ${COLORS.gray}# Check network status${COLORS.reset}
  ${COLORS.blue}sonrctl status network${COLORS.reset}

  ${COLORS.gray}# Stop the network${COLORS.reset}
  ${COLORS.blue}sonrctl stop network${COLORS.reset}

${COLORS.bright}CONFIGURATION${COLORS.reset}
  Config directory: ${COLORS.cyan}~/.config/sonr${COLORS.reset}
  Node home: ${COLORS.cyan}~/.sonr${COLORS.reset}

${COLORS.bright}MORE INFO${COLORS.reset}
  Docs: ${COLORS.blue}https://docs.sonr.io${COLORS.reset}
  Repo: ${COLORS.blue}https://github.com/sonr-io/sonr${COLORS.reset}
`);
}

// Run the CLI
main().catch((error) => {
  Logger.error(`Fatal error: ${error}`);
  process.exit(1);
});