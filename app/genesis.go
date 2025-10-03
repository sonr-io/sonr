// Package app provides genesis state management for the Sonr blockchain application.
package app

import (
	"encoding/json"
)

// GenesisState represents the initial state of the blockchain as a map of raw JSON
// messages keyed by module identifier strings. Each module's genesis state is stored
// as raw JSON to allow flexible initialization during chain setup.
//
// The identifier is used to route genesis information to the appropriate module
// during the init chain process. Default genesis information is populated by
// the ModuleBasicManager from each registered BasicModule.
type GenesisState map[string]json.RawMessage
