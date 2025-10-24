# sonrctl 🚀

**Blazingly fast CLI for managing Sonr blockchain nodes, validators, and networks**

Built with [Bun](https://bun.sh) - the all-in-one JavaScript runtime.

## Features

- ⚡ **Blazingly Fast** - Built on Bun for instant startup and execution
- 🎯 **Simple API** - Clean, intuitive commands for node management
- 🐳 **Docker Integration** - Seamless container orchestration for testnets
- 💾 **SQLite Config** - Lightweight, fast configuration management
- 🎨 **Beautiful Output** - Colored, formatted CLI output
- 🔧 **Node Management** - Initialize, start, stop, and monitor nodes
- 🌐 **Network Operations** - Full testnet deployment and management

## Installation

### Install Bun (if not already installed)

```bash
curl -fsSL https://bun.sh/install | bash
```

### Install sonrctl

```bash
# Clone the repository
git clone https://github.com/sonr-io/sonr.git
cd sonr

# Install dependencies
bun install

# Link the CLI globally
bun link
```

### Install snrd binary

```bash
sonrctl install latest
```

## Quick Start

```bash
# Show help
sonrctl help

# Initialize a validator node
sonrctl init validator val-naruto --chain-id sonrtest_1-1

# Start the entire testnet network
sonrctl start network --detach

# Check network status
sonrctl status network

# Stop the network
sonrctl stop network
```

## Commands

### `sonrctl install [version]`

Install or update the snrd binary.

```bash
# Install latest version
sonrctl install latest

# Install specific version (coming soon)
sonrctl install v1.0.0
```

### `sonrctl init <type> <name> [options]`

Initialize a new node configuration.

**Types:** `validator`, `sentry`, `full`

**Options:**
- `--chain-id <id>` - Chain ID (default: sonrtest_1-1)
- `--home <path>` - Home directory (default: ~/.sonr/<name>)
- `--rpc-port <port>` - RPC port (default: 26657)
- `--rest-port <port>` - REST API port (default: 1317)
- `--grpc-port <port>` - gRPC port (default: 9090)
- `--grpc-web-port <port>` - gRPC-Web port (default: 9091)
- `--json-rpc-port <port>` - JSON-RPC port (default: 8545)
- `--json-rpc-ws-port <port>` - JSON-RPC WebSocket port (default: 8546)

**Examples:**

```bash
# Initialize a validator
sonrctl init validator val-naruto

# Initialize a sentry with custom home
sonrctl init sentry sentry-naruto --home ~/.sonr/custom-sentry

# Initialize with custom ports
sonrctl init validator my-val --rpc-port 26658 --rest-port 1318
```

### `sonrctl start <target> [options]`

Start a node or the entire network.

**Targets:**
- `<node-name>` - Start a specific registered node
- `network` - Start the entire testnet using docker-compose

**Options:**
- `--detach`, `-d` - Run in background (network only)

**Examples:**

```bash
# Start a specific node
sonrctl start val-naruto

# Start network in foreground
sonrctl start network

# Start network in background
sonrctl start network --detach
```

### `sonrctl stop <target>`

Stop a node, network, or all containers.

**Targets:**
- `<node-name>` - Stop a specific container
- `network` - Stop the entire testnet
- `all` - Stop all running containers

**Examples:**

```bash
# Stop a specific node
sonrctl stop val-naruto

# Stop the network
sonrctl stop network

# Stop all containers
sonrctl stop all
```

### `sonrctl status [target]`

Check status of nodes and network.

**Targets:**
- (none) - Show overall status
- `<node-name>` - Show specific node status
- `network` - Show network status

**Examples:**

```bash
# Show overall status
sonrctl status

# Show specific node status
sonrctl status val-naruto

# Show network status
sonrctl status network
```

### `sonrctl config <command> [args]`

Manage sonrctl configuration.

**Commands:**
- `list` - List all configuration
- `get <key>` - Get a specific value
- `set <key> <value>` - Set a configuration value

**Examples:**

```bash
# List all configuration
sonrctl config list

# Get chain ID
sonrctl config get chain_id

# Set home directory
sonrctl config set home ~/.sonr-custom
```

## Configuration

sonrctl stores its configuration in:
- **Config Database:** `~/.config/sonr/config.db`
- **Node Home:** `~/.sonr/` (default)

The SQLite database stores:
- Global configuration (chain ID, binary path, etc.)
- Registered nodes
- Node metadata

## Project Structure

```
src/
├── index.ts              # Main CLI entry point
├── types.ts              # TypeScript type definitions
├── commands/             # Command implementations
│   ├── install.ts        # Install snrd binary
│   ├── init.ts           # Initialize nodes
│   ├── start.ts          # Start nodes/network
│   ├── stop.ts           # Stop nodes/network
│   ├── status.ts         # Check status
│   └── config.ts         # Configuration management
└── lib/                  # Utility libraries
    ├── constants.ts      # Constants and defaults
    ├── logger.ts         # Beautiful logging
    ├── config.ts         # SQLite configuration manager
    ├── toml.ts           # TOML file utilities
    ├── docker.ts         # Docker operations
    └── node.ts           # Node operations
```

## Architecture

sonrctl is built with a clean, modular architecture:

1. **Command Layer** - Handles user input and command routing
2. **Library Layer** - Provides utilities for common operations
3. **Storage Layer** - SQLite database for configuration persistence

### Key Design Decisions

- **Bun Runtime** - For blazingly fast execution and built-in tools
- **SQLite Storage** - Lightweight, fast, and portable
- **No External Dependencies** - Minimal deps, maximum performance
- **Clean Separation** - Commands, libraries, and utilities are separate
- **Type Safety** - Full TypeScript support

## Docker Integration

sonrctl integrates seamlessly with Docker for testnet deployment:

- Start/stop docker-compose networks
- Manage individual containers
- Monitor container status
- Execute commands in containers
- View container logs

The CLI can work as a drop-in replacement for manual docker-compose operations.

## Development

### Prerequisites

- [Bun](https://bun.sh) v1.0+
- Docker (for network operations)
- Git

### Running Locally

```bash
# Run directly with Bun
bun run src/index.ts help

# Run a specific command
bun run src/index.ts status
```

### Adding New Commands

1. Create a new file in `src/commands/`
2. Export an async function that takes `args: string[]`
3. Import and add to the switch statement in `src/index.ts`
4. Update the help text

Example:

```typescript
// src/commands/mycommand.ts
export async function mycommand(args: string[]) {
  Logger.header('My Command');
  // Implementation
}

// src/index.ts
import { mycommand } from './commands/mycommand';

// Add to switch statement
case 'mycommand':
  await mycommand(commandArgs);
  break;
```

## Examples

### Setting Up a Testnet

```bash
# 1. Install the binary
sonrctl install

# 2. Start the testnet network
sonrctl start network --detach

# 3. Check status
sonrctl status network

# 4. View logs
docker logs -f val-naruto

# 5. Stop when done
sonrctl stop network
```

### Creating a Custom Node

```bash
# 1. Initialize a custom validator
sonrctl init validator my-validator \
  --chain-id mychain-1 \
  --home ~/.sonr/my-validator \
  --rpc-port 26667

# 2. Configure genesis (manual step)
# Edit ~/.sonr/my-validator/config/genesis.json

# 3. Start the node
sonrctl start my-validator

# 4. Check status
sonrctl status my-validator
```

## Troubleshooting

### Command not found

```bash
# Make sure Bun's bin directory is in your PATH
export PATH="$HOME/.bun/bin:$PATH"

# Re-link the CLI
bun link
```

### Docker permission denied

```bash
# Add your user to the docker group
sudo usermod -aG docker $USER

# Log out and back in, or run:
newgrp docker
```

### Binary not found after install

```bash
# Check if ~/.local/bin is in your PATH
export PATH="$HOME/.local/bin:$PATH"

# Add to your shell profile
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
```

## Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

Apache 2.0 - See LICENSE file for details

## Links

- [Sonr Network](https://sonr.io)
- [Documentation](https://docs.sonr.io)
- [GitHub](https://github.com/sonr-io/sonr)
- [Bun](https://bun.sh)

---

**Built with ❤️ using Bun**
