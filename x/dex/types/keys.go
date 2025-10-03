package types

const (
	// ModuleName defines the name of module.
	ModuleName = "dex"

	// PortID defines the port ID that module module binds to.
	PortID = ModuleName

	// Version defines the current version the IBC module supports
	Version = ModuleName + "-1"

	// StoreKey is the store key string for the module.
	StoreKey = ModuleName

	// RouterKey is the message route for the module.
	RouterKey = ModuleName

	// QuerierRoute is the querier route for the module.
	QuerierRoute = ModuleName
)

// Event types
const (
	EventTypeICAPacketAcknowledged = "ica_packet_acknowledged"
	EventTypeICAPacketTimeout      = "ica_packet_timeout"
	EventTypeDEXAccountRegistered  = "dex_account_registered"
	EventTypeSwapExecuted          = "swap_executed"
	EventTypeLiquidityProvided     = "liquidity_provided"
	EventTypeLiquidityRemoved      = "liquidity_removed"
	EventTypeOrderCreated          = "order_created"
	EventTypeOrderCancelled        = "order_cancelled"
	EventTypeDIDActivity           = "did_activity"
)
