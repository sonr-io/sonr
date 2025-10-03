# Sonr Go Client SDK

The official Go client SDK for the Sonr blockchain, providing idiomatic Go interfaces for interacting with Sonr's decentralized identity, data storage, and service management features.

## Features

- **Blockchain Interaction**: Query chain state and broadcast transactions
- **Module Support**: Comprehensive support for DID, DWN, SVC, and UCAN modules
- **WebAuthn Integration**: Passwordless authentication with hardware-backed keys
- **Transaction Building**: Simplified transaction construction and broadcasting
- **Key Management**: Secure keyring integration with multiple backends
- **Network Configuration**: Pre-configured endpoints for testnet and mainnet

## Installation

```bash
go get github.com/sonr-io/sonr/client
```

## Quick Start

### Basic SDK Setup

```go
package main

import (
    "context"
    "log"

    client "github.com/sonr-io/sonr/client"
)

func main() {
    // Initialize SDK with testnet configuration
    sdk, err := client.NewWithNetwork("testnet")
    if err != nil {
        log.Fatal(err)
    }
    defer sdk.Close()

    // Or create with custom configuration
    cfg := client.DefaultConfig()
    cfg.Network.GRPCEndpoint = "localhost:9090"
    sdk, err = client.New(cfg)
    if err != nil {
        log.Fatal(err)
    }

    // Check connection
    if !sdk.IsConnected() {
        log.Fatal("Failed to connect to network")
    }

    // Access the underlying Sonr client
    sonrClient := sdk.Client()

    // Query chain info
    ctx := context.Background()
    info, err := sonrClient.Query().GetNodeInfo(ctx)
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("Connected to chain: %s", info.Network)
    log.Printf("SDK Version: %s", client.Version())
}
```

### Transaction Broadcasting

```go
// Get the client from SDK
sonrClient := sdk.Client()

// Create and broadcast a transaction
tx := sonrClient.Transaction().
    WithChainID("sonrtest_1-1").
    WithGasPrice(0.001, "usnr").
    WithMemo("Hello Sonr!")

// Add messages to transaction
tx.AddMessage(/* your message */)

// Sign and broadcast
result, err := tx.SignAndBroadcast(ctx, keyring)
if err != nil {
    log.Fatal(err)
}

log.Printf("Transaction hash: %s", result.TxHash)
```

### Module-Specific Operations

```go
// Get the client from SDK
sonrClient := sdk.Client()

// DID operations
didClient := sonrClient.DID()
did, err := didClient.CreateDID(ctx, didOpts)

// DWN operations
dwnClient := sonrClient.DWN()
record, err := dwnClient.CreateRecord(ctx, recordData)

// Service operations
svcClient := sonrClient.SVC()
service, err := svcClient.RegisterService(ctx, serviceConfig)

// UCAN operations
ucanClient := sonrClient.UCAN()
capability, err := ucanClient.CreateCapability(ctx, capabilitySpec)
```

## API Overview

### Core Client (`sonr.Client`)

The main client provides access to:

- **Query operations**: Read blockchain state
- **Transaction building**: Create and broadcast transactions
- **Module clients**: Access to specialized functionality
- **Connection management**: Handle multiple endpoint types

### Module Clients

- **DID Client**: W3C DID operations with WebAuthn support
- **DWN Client**: Decentralized Web Node data management
- **SVC Client**: Service registration and management
- **UCAN Client**: User-Controlled Authorization Networks

### Key Management

- **Keyring integration**: Support for multiple keyring backends
- **Hardware keys**: WebAuthn and hardware wallet support
- **Multi-signature**: Threshold signature schemes

### Network Configuration

Pre-configured networks:

- **Testnet**: `sonrtest_1-1` with development endpoints
- **Mainnet**: Production network configuration (coming soon)

## Examples

See the `examples/` directory for complete working examples:

- **CLI Tool**: Example command-line application
- **Web Integration**: HTTP API server
- **Key Management**: Keyring and signing examples
- **Module Operations**: Comprehensive module usage

## Documentation

- [API Reference](https://pkg.go.dev/github.com/sonr-io/sonr/client)
- [Sonr Documentation](https://sonr.dev)
- [Blockchain Guide](https://sonr.dev/blockchain)

## Contributing

This SDK is part of the main Sonr repository. Please see the main project's contributing guidelines.

## License

This project is licensed under the Apache 2.0 License.
