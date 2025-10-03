// Package main provides the Highway service, a UCAN-based task processing service
// that acts as a bridge proxy for MPC operations and decentralized identity management.
//
// The service processes asynchronous UCAN (User-Controlled Authorization Networks) tasks
// including token creation, delegation, signing, verification, and DID generation.
// It serves as a proxy between the bridge handlers and the underlying blockchain operations.
package main

import (
	"github.com/sonr-io/sonr/bridge"
)

func main() {
	// Create and configure the Highway service
	service := bridge.NewHighwayService()
	defer service.Shutdown()

	// Start the service and block until shutdown
	service.Start()
}
