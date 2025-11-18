package types

import host "github.com/cosmos/ibc-go/v8/modules/core/24-host"

// DefaultGenesisState returns the default module GenesisState.
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params: DefaultParams(),
		PortId: PortID,
	}
}

// NewGenesisState initializes and returns a new GenesisState.
func NewGenesisState() *GenesisState {
	return &GenesisState{
		Params: DefaultParams(),
		PortId: PortID,
	}
}

// Validate performs basic validation of the GenesisState.
func (gs *GenesisState) Validate() error {
	if err := host.PortIdentifierValidator(gs.PortId); err != nil {
		return err
	}

	// Validate params
	if err := gs.Params.Validate(); err != nil {
		return err
	}

	return nil
}
