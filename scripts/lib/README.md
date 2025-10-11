# Sonr Script Library

This directory contains modular helper scripts for Sonr blockchain operations. These helpers provide reusable functions for common tasks like environment setup, configuration management, key handling, and transaction operations.

## Library Structure

### `env.sh` - Environment Management
Centralizes default environment variables and provides utility functions for logging, validation, and system checks.

**Key Functions:**
- `init_env()` - Initialize environment with required tools and cleanup
- `log_info()`, `log_warn()`, `log_error()`, `log_success()` - Colored logging
- `require_cmd()`, `ensure_file()`, `ensure_binary()` - Validation helpers
- `is_docker()` - Check if running in Docker container

**Environment Variables:**
- `DENOM=usnr` - Default denomination
- `CHAIN_BIN=snrd` - Binary name
- `CHAIN_DIR=~/.sonr` - Chain data directory
- `CHAIN_ID=sonrtest_1-1` - Default chain ID

### `jq_patch.sh` - JSON Operations
Provides safe JSON patching operations using `jq` with temporary files and validation.

**Key Functions:**
- `patch_json(file, jq_expr)` - Apply jq expression to file
- `set_json_string()`, `set_json_number()`, `set_json_bool()` - Type-safe setters
- `json_path_exists()`, `get_json_value()` - Query helpers
- `validate_json()` - JSON validation

### `config.sh` - Configuration Management
Handles TOML configuration files for Cosmos SDK nodes using `crudini` or `sed` fallbacks.

**Key Functions:**
- `configure_node()` - Complete node configuration with ports and settings
- `enable_rpc()`, `enable_rest()`, `enable_grpc()` - Enable services
- `set_consensus_timeouts()`, `set_min_gas_prices()` - Parameter setters
- `set_toml_value()` - Generic TOML value setter

### `keys.sh` - Key Management
Provides functions for importing, managing, and using cryptographic keys.

**Key Functions:**
- `import_mnemonic()` - Import key from mnemonic phrase
- `ensure_key()`, `get_key_address()` - Key validation and retrieval
- `fund_key()`, `delegate_to_validator()` - Account operations
- `wait_for_sync()` - Node synchronization waiting

### `tx.sh` - Transaction Operations
Handles blockchain transactions with error handling and retry logic.

**Key Functions:**
- `submit_tx()` - Submit transaction with gas and error handling
- `query_chain()` - Query blockchain state
- `submit_proposal()`, `vote_proposal()` - Governance operations
- `stake_tokens()`, `get_balance()` - Staking operations

### `genesis.sh` - Genesis File Operations
Specialized functions for genesis file creation and modification.

**Key Functions:**
- `generate_vrf_key()` - Generate VRF keypair for network
- `update_genesis_params()` - Apply Sonr-specific genesis parameters
- `add_constitution()` - Add constitution to governance
- `validate_genesis()` - Genesis file validation

## Usage Examples

### Basic Environment Setup
```bash
#!/bin/bash
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/lib/env.sh"
source "${SCRIPT_DIR}/lib/config.sh"

init_env
ensure_chain_dir
configure_node "$CHAIN_DIR" --rpc-port 26657 --rest-port 1317
```

### Key Management
```bash
#!/bin/bash
source "scripts/lib/keys.sh"
import_mnemonic "mykey" "word1 word2 word3..." "eth_secp256k1"
fund_key "mykey" "1000000usnr"
```

### Genesis Creation
```bash
#!/bin/bash
source "scripts/lib/genesis.sh"
update_genesis_params
add_constitution
generate_vrf_key "$CHAIN_DIR"
```

## Integration Guidelines

1. **Always source required libraries** at the top of scripts
2. **Call `init_env()`** early to set up logging and validation
3. **Use consistent error handling** with the provided logging functions
4. **Validate inputs** using `ensure_*` functions before operations
5. **Handle Docker vs local execution** in wrapper functions
6. **Use descriptive log messages** for user feedback

## Error Handling

All functions use consistent error handling:
- Functions return 0 on success, 1 on failure
- Use `log_error()` for user-visible errors
- Use `log_warn()` for non-fatal issues
- Cleanup temporary files automatically via `trap`

## Docker Compatibility

All functions support both local and Docker execution modes:
- Use `is_docker()` to detect environment
- Use `run_binary()` wrapper for command execution
- Mount volumes correctly for Docker containers
- Handle TTY and interactive mode appropriately