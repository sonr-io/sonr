package types

// DIDAccounts wraps a slice of account IDs for use with collections
type DIDAccounts struct {
	Accounts []string `protobuf:"bytes,1,rep,name=accounts,proto3" json:"accounts,omitempty"`
}

// ProtoMessage implements proto.Message
func (DIDAccounts) ProtoMessage() {}

// Reset implements proto.Message
func (m *DIDAccounts) Reset() {
	*m = DIDAccounts{}
}

// String implements proto.Message
func (m DIDAccounts) String() string {
	return m.Accounts[0] // Simple string representation
}
