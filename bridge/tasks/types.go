// Package tasks provides UCAN-based tasks for the MPC enclave service.
package tasks

import (
	"time"

	"github.com/asynkron/protoactor-go/actor"
)

var system = actor.NewActorSystem()

const KRequestTimeout = 20 * time.Second

// A list of UCAN-based task types.
const (
	TypeUCANToken           = "ucan:token"            // Create UCAN origin tokens
	TypeUCANAttenuation     = "ucan:attenuation"      // Create attenuated UCAN tokens
	TypeUCANSign            = "ucan:sign"             // MPC-based data signing
	TypeUCANVerify          = "ucan:verify"           // MPC-based signature verification
	TypeUCANDIDGeneration   = "ucan:did:generation"   // DID generation from MPC enclave
	TypeUCANTokenValidation = "ucan:token:validation" // UCAN token validation
)
