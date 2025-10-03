# pkljar

> Interactive CLI for generating type-safe Sonr Network configurations using Apple's Pkl language

[![npm version](https://img.shields.io/npm/v/@sonr.io/pkljar.svg)](https://www.npmjs.com/package/@sonr.io/pkljar)
[![License: ISC](https://img.shields.io/badge/License-ISC-blue.svg)](https://opensource.org/licenses/ISC)

## Overview

`pkljar` is a comprehensive configuration management system for the Sonr Network that provides an interactive CLI to generate, validate, and distribute type-safe configuration files. Built on Apple's Pkl configuration language, it ensures compile-time validation and type safety for all service configurations.

## Features

- 🎯 **Interactive CLI** - User-friendly prompts for selecting and configuring Pkl packages
- 📦 **Pre-built Packages** - Ready-to-use configuration packages for Sonr services
- 🔒 **Type Safety** - Compile-time validation using Pkl's strong typing system
- 🌐 **Remote Package Resolution** - Automatic fetching from `mod.pkl.sh` repository
- 🎨 **Multiple Output Formats** - Support for YAML, JSON, TOML, XML, and more
- 🐳 **Docker Support** - Multi-platform Docker images for containerized deployments

## Installation

### Using npm

```bash
npm install -g @sonr.io/pkljar
```

### Using npx (no installation required)

```bash
npx @sonr.io/pkljar
```

### Using Docker

```bash
docker run onsonr/pkljar eval_beam
```

## Quick Start

1. Run the CLI:
   ```bash
   pkljar
   ```

2. Select a package from the interactive menu:
   - `sonr.beam` - Matrix/Element communication bridge
   - `sonr.core` - Core Sonr Network configuration
   - `sonr.hway` - Highway node configuration
   - `sonr.testnet` - Testnet deployment configuration

3. Choose a module to evaluate (e.g., Config, Docker, Starship)

4. Specify output directory and format

5. Configuration files are generated in your specified directory

## Available Packages

### sonr.beam
Matrix bridge configuration for Sonr Network communication.

**Modules:**
- `Config.pkl` - Main configuration
- `Element.pkl` - Element client configuration
- `Hookshot.pkl` - GitHub/GitLab bridge configuration
- `Synapse.pkl` - Matrix server configuration

### sonr.core
Core Sonr Network blockchain configuration.

**Modules:**
- `Config.pkl` - Core network configuration
- `Keys.pkl` - Key management configuration
- `UCAN.pkl` - UCAN authorization configuration
- `Wallet.pkl` - Wallet service configuration

### sonr.hway
Highway node configuration for network routing.

**Modules:**
- `Config.pkl` - Highway node configuration

### sonr.testnet
Testnet deployment configurations.

**Modules:**
- `Docker.pkl` - Docker Compose configuration
- `Starship.pkl` - Kubernetes deployment via Starship

## CLI Usage

### Interactive Mode (Default)

```bash
pkljar
```

Launches an interactive prompt that guides you through:
1. Package selection
2. Module selection
3. Output directory configuration
4. Format selection (YAML, JSON, TOML, etc.)

### Output Formats

The CLI supports multiple output formats:
- `auto` - Use module's default format
- `yaml` - YAML format
- `json` - JSON format
- `jsonnet` - Jsonnet format
- `pcf` - Pkl Configuration Format
- `plist` - Property List format
- `properties` - Java Properties format
- `textproto` - Text Protocol Buffer format
- `xml` - XML format

## Development

### Project Structure

```
pkljar/
├── src/
│   └── index.js        # CLI entry point
├── packages/           # Pkl package definitions
│   ├── sonr.beam/     # Matrix bridge packages
│   ├── sonr.core/     # Core network packages
│   ├── sonr.hway/     # Highway node packages
│   └── sonr.testnet/  # Testnet packages
├── docker/            # Docker configurations
└── Makefile          # Build automation
```

### Building from Source

```bash
# Clone the repository
git clone https://github.com/sonr-io/pkljar.git
cd pkljar

# Install dependencies
npm install

# Run locally
npm start

# Build packages
make release

# Build Docker image
make docker
```

### Creating Custom Packages

1. Create a new directory in `packages/` following the naming convention:
   ```bash
   mkdir packages/myorg.myservice
   ```

2. Create a `PklProject` file:
   ```pkl
   amends "../../basePklProject.pkl"
   
   dependencies {
     ["base.web"] = import("../base.web/PklProject")
   }
   ```

3. Define your configuration modules

4. Build and publish:
   ```bash
   make release
   ```

## Docker Usage

### Pre-built Commands

```bash
# Generate Beam configuration
docker run onsonr/pkljar eval_beam

# Generate Synapse configuration
docker run onsonr/pkljar eval_beam_synapse

# Generate Hookshot configuration
docker run onsonr/pkljar eval_beam_hookshot
```

### Custom Evaluation

```bash
docker run -v $(pwd)/output:/output onsonr/pkljar \
  pkl eval -m /output https://mod.pkl.sh/sonr.core/Config.pkl
```

## API Reference

### Package Resolution

All packages are resolved from the `mod.pkl.sh` repository:
```
https://mod.pkl.sh/{package}/{module}
```

Example:
```
https://mod.pkl.sh/sonr.core/Config.pkl
```

### Module Structure

Each Pkl module follows this structure:
```pkl
@ModuleInfo { minPklVersion = "0.27.0" }
module package.name.ModuleName

// Type definitions
class ConfigClass {
  property: Type
}

// Configuration instance
config: ConfigClass = new ConfigClass {
  property = value
}

// Output configuration
output {
  renderer = new YamlRenderer {}
}
```

## Advanced Configuration

### Environment-Specific Overrides

Use Pkl's conditional logic for environment-specific configurations:

```pkl
config = new ServiceConfig {
  host = if (env == "production") 
    "prod.example.com" 
  else 
    "localhost"
}
```

### Multi-File Output

Generate multiple configuration files from a single module:

```pkl
output {
  files {
    ["config/app.toml"] = appConfig.output
    ["config/client.toml"] = clientConfig.output
    ["docker-compose.yml"] = dockerConfig.output
  }
}
```

## Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

### Development Workflow

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests: `pkl test packages/*/`
5. Submit a pull request

## Support

- **Issues**: [GitHub Issues](https://github.com/sonr-io/pkljar/issues)
- **Documentation**: [Pkl Language Documentation](https://pkl-lang.org)
- **Community**: [Sonr Discord](https://discord.gg/sonr)

## License

ISC License - see [LICENSE](LICENSE) file for details.

## Acknowledgments

- Built with [Pkl](https://pkl-lang.org) by Apple
- CLI powered by [@clack/prompts](https://github.com/natemoo-re/clack)
- Pkl runtime by [@pkl-community/pkl](https://github.com/pkl-community/pkl-npm)