package types

import "fmt"

// DefaultIndex is the default global index
const DefaultIndex uint64 = 1

// DefaultGenesis returns the default genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params:       DefaultParams(),
		Capabilities: []ServiceCapability{},
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	// Validate parameters
	if err := gs.Params.Validate(); err != nil {
		return err
	}

	// Validate capabilities
	capabilityIDs := make(map[string]bool)
	for i, cap := range gs.Capabilities {
		// Check for duplicate capability IDs
		if capabilityIDs[cap.CapabilityId] {
			return fmt.Errorf("duplicate capability ID at index %d: %s", i, cap.CapabilityId)
		}
		capabilityIDs[cap.CapabilityId] = true

		// Validate individual capability fields
		if cap.CapabilityId == "" {
			return fmt.Errorf("capability at index %d has empty ID", i)
		}
		if cap.ServiceId == "" {
			return fmt.Errorf("capability %s has empty service ID", cap.CapabilityId)
		}
		if cap.Domain == "" {
			return fmt.Errorf("capability %s has empty domain", cap.CapabilityId)
		}
		if cap.Owner == "" {
			return fmt.Errorf("capability %s has empty owner", cap.CapabilityId)
		}
		if len(cap.Abilities) == 0 {
			return fmt.Errorf("capability %s has no abilities", cap.CapabilityId)
		}
	}

	return nil
}
