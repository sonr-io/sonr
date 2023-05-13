// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: sonr/common/info.proto

package common

import (
	fmt "fmt"
	proto "github.com/gogo/protobuf/proto"
	io "io"
	math "math"
	math_bits "math/bits"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

// Account represents a user's account on a blockchain or network. It contains the user's address, name, DID, and other information that is used to identify and manage the account.
type AccountInfo struct {
	// This field represents the unique account address associated with the user. It is typically a hash or an encoded public key, depending on the underlying blockchain or network.
	Address string `protobuf:"bytes,1,opt,name=address,proto3" json:"address,omitempty"`
	// This field contains the human-readable name associated with the account. It is used for easier identification and management of the account by the user.
	Name string `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	// This field stores the Decentralized Identifier (DID) of the user, which is a unique, resolvable, and cryptographically verifiable identifier. DIDs are used to enable secure and decentralized identity management.
	Did string `protobuf:"bytes,3,opt,name=did,proto3" json:"did,omitempty"`
	// This field specifies the type of the cryptocurrency or token associated with the account. It is used to differentiate between various cryptocurrencies or tokens that the user may hold in their wallet.
	CoinType string `protobuf:"bytes,4,opt,name=coin_type,json=coinType,proto3" json:"coin_type,omitempty"`
	// This field represents the identifier of the blockchain or network that the account is associated with. Chain IDs are used to distinguish between different blockchains or networks, such as Ethereum, Cosmos, or Filecoin.
	ChainId string `protobuf:"bytes,5,opt,name=chain_id,json=chainId,proto3" json:"chain_id,omitempty"`
	// This field stores the base64 encoded public key of the account, which is used for cryptographic operations such as signing and verifying transactions. The public key is derived from the user's private key and is an essential part of the account's security.
	PublicKey string `protobuf:"bytes,6,opt,name=public_key,json=publicKey,proto3" json:"public_key,omitempty"`
	// This field stores the type of the public key. It is used to differentiate between various public key types, such as secp256k1, ed25519, and sr25519.
	Type string `protobuf:"bytes,7,opt,name=type,proto3" json:"type,omitempty"`
}

func (m *AccountInfo) Reset()         { *m = AccountInfo{} }
func (m *AccountInfo) String() string { return proto.CompactTextString(m) }
func (*AccountInfo) ProtoMessage()    {}
func (*AccountInfo) Descriptor() ([]byte, []int) {
	return fileDescriptor_117f1cca4f9b8f25, []int{0}
}
func (m *AccountInfo) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *AccountInfo) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_AccountInfo.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *AccountInfo) XXX_Merge(src proto.Message) {
	xxx_messageInfo_AccountInfo.Merge(m, src)
}
func (m *AccountInfo) XXX_Size() int {
	return m.Size()
}
func (m *AccountInfo) XXX_DiscardUnknown() {
	xxx_messageInfo_AccountInfo.DiscardUnknown(m)
}

var xxx_messageInfo_AccountInfo proto.InternalMessageInfo

func (m *AccountInfo) GetAddress() string {
	if m != nil {
		return m.Address
	}
	return ""
}

func (m *AccountInfo) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *AccountInfo) GetDid() string {
	if m != nil {
		return m.Did
	}
	return ""
}

func (m *AccountInfo) GetCoinType() string {
	if m != nil {
		return m.CoinType
	}
	return ""
}

func (m *AccountInfo) GetChainId() string {
	if m != nil {
		return m.ChainId
	}
	return ""
}

func (m *AccountInfo) GetPublicKey() string {
	if m != nil {
		return m.PublicKey
	}
	return ""
}

func (m *AccountInfo) GetType() string {
	if m != nil {
		return m.Type
	}
	return ""
}

// Basic Info Sent to Peers to Establish Connections
type PeerInfo struct {
	Id        string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Name      string `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	PeerId    string `protobuf:"bytes,3,opt,name=peer_id,json=peerId,proto3" json:"peer_id,omitempty"`
	Multiaddr string `protobuf:"bytes,4,opt,name=multiaddr,proto3" json:"multiaddr,omitempty"`
}

func (m *PeerInfo) Reset()         { *m = PeerInfo{} }
func (m *PeerInfo) String() string { return proto.CompactTextString(m) }
func (*PeerInfo) ProtoMessage()    {}
func (*PeerInfo) Descriptor() ([]byte, []int) {
	return fileDescriptor_117f1cca4f9b8f25, []int{1}
}
func (m *PeerInfo) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *PeerInfo) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_PeerInfo.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *PeerInfo) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PeerInfo.Merge(m, src)
}
func (m *PeerInfo) XXX_Size() int {
	return m.Size()
}
func (m *PeerInfo) XXX_DiscardUnknown() {
	xxx_messageInfo_PeerInfo.DiscardUnknown(m)
}

var xxx_messageInfo_PeerInfo proto.InternalMessageInfo

func (m *PeerInfo) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *PeerInfo) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *PeerInfo) GetPeerId() string {
	if m != nil {
		return m.PeerId
	}
	return ""
}

func (m *PeerInfo) GetMultiaddr() string {
	if m != nil {
		return m.Multiaddr
	}
	return ""
}

type WalletInfo struct {
	// Controller is the associated Sonr address.
	Controller string `protobuf:"bytes,1,opt,name=controller,proto3" json:"controller,omitempty"`
	// DiscoverPaths is a list of all known hardened coin type paths.
	DiscoveredPaths []int32 `protobuf:"varint,2,rep,packed,name=discovered_paths,json=discoveredPaths,proto3" json:"discovered_paths,omitempty"`
	// Algorithm is the algorithm of the wallet. CMP is the default.
	Algorithm string `protobuf:"bytes,3,opt,name=algorithm,proto3" json:"algorithm,omitempty"`
	// CreatedAt is the time the wallet was created.
	CreatedAt int64 `protobuf:"varint,4,opt,name=created_at,json=createdAt,proto3" json:"created_at,omitempty"`
	// LastUpdated is the last time the wallet was updated.
	LastUpdated int64 `protobuf:"varint,5,opt,name=last_updated,json=lastUpdated,proto3" json:"last_updated,omitempty"`
}

func (m *WalletInfo) Reset()         { *m = WalletInfo{} }
func (m *WalletInfo) String() string { return proto.CompactTextString(m) }
func (*WalletInfo) ProtoMessage()    {}
func (*WalletInfo) Descriptor() ([]byte, []int) {
	return fileDescriptor_117f1cca4f9b8f25, []int{2}
}
func (m *WalletInfo) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *WalletInfo) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_WalletInfo.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *WalletInfo) XXX_Merge(src proto.Message) {
	xxx_messageInfo_WalletInfo.Merge(m, src)
}
func (m *WalletInfo) XXX_Size() int {
	return m.Size()
}
func (m *WalletInfo) XXX_DiscardUnknown() {
	xxx_messageInfo_WalletInfo.DiscardUnknown(m)
}

var xxx_messageInfo_WalletInfo proto.InternalMessageInfo

func (m *WalletInfo) GetController() string {
	if m != nil {
		return m.Controller
	}
	return ""
}

func (m *WalletInfo) GetDiscoveredPaths() []int32 {
	if m != nil {
		return m.DiscoveredPaths
	}
	return nil
}

func (m *WalletInfo) GetAlgorithm() string {
	if m != nil {
		return m.Algorithm
	}
	return ""
}

func (m *WalletInfo) GetCreatedAt() int64 {
	if m != nil {
		return m.CreatedAt
	}
	return 0
}

func (m *WalletInfo) GetLastUpdated() int64 {
	if m != nil {
		return m.LastUpdated
	}
	return 0
}

func init() {
	proto.RegisterType((*AccountInfo)(nil), "sonrhq.sonr.common.AccountInfo")
	proto.RegisterType((*PeerInfo)(nil), "sonrhq.sonr.common.PeerInfo")
	proto.RegisterType((*WalletInfo)(nil), "sonrhq.sonr.common.WalletInfo")
}

func init() { proto.RegisterFile("sonr/common/info.proto", fileDescriptor_117f1cca4f9b8f25) }

var fileDescriptor_117f1cca4f9b8f25 = []byte{
	// 398 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x6c, 0x92, 0xb1, 0xae, 0xd3, 0x30,
	0x14, 0x86, 0x9b, 0xe4, 0xde, 0xb6, 0x39, 0x17, 0xc1, 0x95, 0x07, 0x30, 0x02, 0xa2, 0x4b, 0x59,
	0xca, 0xd2, 0x0c, 0xcc, 0x0c, 0x65, 0xab, 0x58, 0xaa, 0x08, 0x84, 0xc4, 0x12, 0xb9, 0xf6, 0x69,
	0x63, 0xe1, 0xd8, 0xc1, 0x71, 0x90, 0xf2, 0x16, 0xbc, 0x0d, 0x3c, 0x02, 0x63, 0x47, 0x46, 0xd4,
	0xbe, 0x08, 0xb2, 0x13, 0x28, 0x03, 0x53, 0xce, 0xf9, 0x2c, 0x9d, 0xff, 0xcf, 0xaf, 0x1f, 0x1e,
	0xb6, 0x46, 0xdb, 0x9c, 0x9b, 0xba, 0x36, 0x3a, 0x97, 0x7a, 0x6f, 0x56, 0x8d, 0x35, 0xce, 0x10,
	0xe2, 0x79, 0xf5, 0x79, 0xe5, 0x3f, 0xab, 0xe1, 0x79, 0xf1, 0x3d, 0x82, 0x9b, 0x35, 0xe7, 0xa6,
	0xd3, 0x6e, 0xa3, 0xf7, 0x86, 0x50, 0x98, 0x31, 0x21, 0x2c, 0xb6, 0x2d, 0x8d, 0xee, 0xa2, 0x65,
	0x5a, 0xfc, 0x59, 0x09, 0x81, 0x2b, 0xcd, 0x6a, 0xa4, 0x71, 0xc0, 0x61, 0x26, 0xb7, 0x90, 0x08,
	0x29, 0x68, 0x12, 0x90, 0x1f, 0xc9, 0x13, 0x48, 0xb9, 0x91, 0xba, 0x74, 0x7d, 0x83, 0xf4, 0x2a,
	0xf0, 0xb9, 0x07, 0xef, 0xfa, 0x06, 0xc9, 0x63, 0x98, 0xf3, 0x8a, 0x49, 0x5d, 0x4a, 0x41, 0xaf,
	0x87, 0xeb, 0x61, 0xdf, 0x08, 0xf2, 0x0c, 0xa0, 0xe9, 0x76, 0x4a, 0xf2, 0xf2, 0x13, 0xf6, 0x74,
	0x1a, 0x1e, 0xd3, 0x81, 0xbc, 0xc5, 0xde, 0x8b, 0x87, 0x8b, 0xb3, 0x41, 0xdc, 0xcf, 0x0b, 0x84,
	0xf9, 0x16, 0xd1, 0x06, 0xdb, 0xf7, 0x21, 0x96, 0x62, 0x74, 0x1c, 0x4b, 0xf1, 0x5f, 0xb3, 0x8f,
	0x60, 0xd6, 0x20, 0xda, 0xf2, 0xaf, 0xe1, 0xa9, 0x5f, 0x37, 0x82, 0x3c, 0x85, 0xb4, 0xee, 0x94,
	0x93, 0xfe, 0x4f, 0x47, 0xcf, 0x17, 0xb0, 0xf8, 0x16, 0x01, 0x7c, 0x60, 0x4a, 0xe1, 0x10, 0x50,
	0x06, 0xc0, 0x8d, 0x76, 0xd6, 0x28, 0x85, 0x76, 0x54, 0xfc, 0x87, 0x90, 0x97, 0x70, 0x2b, 0x64,
	0xcb, 0xcd, 0x17, 0xb4, 0x28, 0xca, 0x86, 0xb9, 0xaa, 0xa5, 0xf1, 0x5d, 0xb2, 0xbc, 0x2e, 0x1e,
	0x5c, 0xf8, 0xd6, 0x63, 0xaf, 0xcb, 0xd4, 0xc1, 0x58, 0xe9, 0xaa, 0x7a, 0xb4, 0x74, 0x01, 0x3e,
	0x11, 0x6e, 0x91, 0x39, 0x14, 0x25, 0x73, 0xc1, 0x56, 0x52, 0xa4, 0x23, 0x59, 0x3b, 0xf2, 0x1c,
	0xee, 0x29, 0xd6, 0xba, 0xb2, 0x6b, 0x84, 0x27, 0x21, 0xcf, 0xa4, 0xb8, 0xf1, 0xec, 0xfd, 0x80,
	0xde, 0xbc, 0xfe, 0x71, 0xca, 0xa2, 0xe3, 0x29, 0x8b, 0x7e, 0x9d, 0xb2, 0xe8, 0xeb, 0x39, 0x9b,
	0x1c, 0xcf, 0xd9, 0xe4, 0xe7, 0x39, 0x9b, 0x7c, 0x7c, 0x71, 0x90, 0xae, 0xea, 0x76, 0xbe, 0x04,
	0xf9, 0x50, 0x8a, 0x9c, 0x1b, 0x8b, 0xb9, 0xcf, 0xb5, 0x1d, 0x9b, 0xb3, 0x9b, 0x86, 0xd6, 0xbc,
	0xfa, 0x1d, 0x00, 0x00, 0xff, 0xff, 0x4a, 0x89, 0x8a, 0x22, 0x4f, 0x02, 0x00, 0x00,
}

func (m *AccountInfo) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *AccountInfo) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *AccountInfo) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Type) > 0 {
		i -= len(m.Type)
		copy(dAtA[i:], m.Type)
		i = encodeVarintInfo(dAtA, i, uint64(len(m.Type)))
		i--
		dAtA[i] = 0x3a
	}
	if len(m.PublicKey) > 0 {
		i -= len(m.PublicKey)
		copy(dAtA[i:], m.PublicKey)
		i = encodeVarintInfo(dAtA, i, uint64(len(m.PublicKey)))
		i--
		dAtA[i] = 0x32
	}
	if len(m.ChainId) > 0 {
		i -= len(m.ChainId)
		copy(dAtA[i:], m.ChainId)
		i = encodeVarintInfo(dAtA, i, uint64(len(m.ChainId)))
		i--
		dAtA[i] = 0x2a
	}
	if len(m.CoinType) > 0 {
		i -= len(m.CoinType)
		copy(dAtA[i:], m.CoinType)
		i = encodeVarintInfo(dAtA, i, uint64(len(m.CoinType)))
		i--
		dAtA[i] = 0x22
	}
	if len(m.Did) > 0 {
		i -= len(m.Did)
		copy(dAtA[i:], m.Did)
		i = encodeVarintInfo(dAtA, i, uint64(len(m.Did)))
		i--
		dAtA[i] = 0x1a
	}
	if len(m.Name) > 0 {
		i -= len(m.Name)
		copy(dAtA[i:], m.Name)
		i = encodeVarintInfo(dAtA, i, uint64(len(m.Name)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.Address) > 0 {
		i -= len(m.Address)
		copy(dAtA[i:], m.Address)
		i = encodeVarintInfo(dAtA, i, uint64(len(m.Address)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *PeerInfo) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *PeerInfo) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *PeerInfo) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Multiaddr) > 0 {
		i -= len(m.Multiaddr)
		copy(dAtA[i:], m.Multiaddr)
		i = encodeVarintInfo(dAtA, i, uint64(len(m.Multiaddr)))
		i--
		dAtA[i] = 0x22
	}
	if len(m.PeerId) > 0 {
		i -= len(m.PeerId)
		copy(dAtA[i:], m.PeerId)
		i = encodeVarintInfo(dAtA, i, uint64(len(m.PeerId)))
		i--
		dAtA[i] = 0x1a
	}
	if len(m.Name) > 0 {
		i -= len(m.Name)
		copy(dAtA[i:], m.Name)
		i = encodeVarintInfo(dAtA, i, uint64(len(m.Name)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.Id) > 0 {
		i -= len(m.Id)
		copy(dAtA[i:], m.Id)
		i = encodeVarintInfo(dAtA, i, uint64(len(m.Id)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *WalletInfo) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *WalletInfo) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *WalletInfo) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.LastUpdated != 0 {
		i = encodeVarintInfo(dAtA, i, uint64(m.LastUpdated))
		i--
		dAtA[i] = 0x28
	}
	if m.CreatedAt != 0 {
		i = encodeVarintInfo(dAtA, i, uint64(m.CreatedAt))
		i--
		dAtA[i] = 0x20
	}
	if len(m.Algorithm) > 0 {
		i -= len(m.Algorithm)
		copy(dAtA[i:], m.Algorithm)
		i = encodeVarintInfo(dAtA, i, uint64(len(m.Algorithm)))
		i--
		dAtA[i] = 0x1a
	}
	if len(m.DiscoveredPaths) > 0 {
		dAtA2 := make([]byte, len(m.DiscoveredPaths)*10)
		var j1 int
		for _, num1 := range m.DiscoveredPaths {
			num := uint64(num1)
			for num >= 1<<7 {
				dAtA2[j1] = uint8(uint64(num)&0x7f | 0x80)
				num >>= 7
				j1++
			}
			dAtA2[j1] = uint8(num)
			j1++
		}
		i -= j1
		copy(dAtA[i:], dAtA2[:j1])
		i = encodeVarintInfo(dAtA, i, uint64(j1))
		i--
		dAtA[i] = 0x12
	}
	if len(m.Controller) > 0 {
		i -= len(m.Controller)
		copy(dAtA[i:], m.Controller)
		i = encodeVarintInfo(dAtA, i, uint64(len(m.Controller)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func encodeVarintInfo(dAtA []byte, offset int, v uint64) int {
	offset -= sovInfo(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *AccountInfo) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Address)
	if l > 0 {
		n += 1 + l + sovInfo(uint64(l))
	}
	l = len(m.Name)
	if l > 0 {
		n += 1 + l + sovInfo(uint64(l))
	}
	l = len(m.Did)
	if l > 0 {
		n += 1 + l + sovInfo(uint64(l))
	}
	l = len(m.CoinType)
	if l > 0 {
		n += 1 + l + sovInfo(uint64(l))
	}
	l = len(m.ChainId)
	if l > 0 {
		n += 1 + l + sovInfo(uint64(l))
	}
	l = len(m.PublicKey)
	if l > 0 {
		n += 1 + l + sovInfo(uint64(l))
	}
	l = len(m.Type)
	if l > 0 {
		n += 1 + l + sovInfo(uint64(l))
	}
	return n
}

func (m *PeerInfo) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Id)
	if l > 0 {
		n += 1 + l + sovInfo(uint64(l))
	}
	l = len(m.Name)
	if l > 0 {
		n += 1 + l + sovInfo(uint64(l))
	}
	l = len(m.PeerId)
	if l > 0 {
		n += 1 + l + sovInfo(uint64(l))
	}
	l = len(m.Multiaddr)
	if l > 0 {
		n += 1 + l + sovInfo(uint64(l))
	}
	return n
}

func (m *WalletInfo) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Controller)
	if l > 0 {
		n += 1 + l + sovInfo(uint64(l))
	}
	if len(m.DiscoveredPaths) > 0 {
		l = 0
		for _, e := range m.DiscoveredPaths {
			l += sovInfo(uint64(e))
		}
		n += 1 + sovInfo(uint64(l)) + l
	}
	l = len(m.Algorithm)
	if l > 0 {
		n += 1 + l + sovInfo(uint64(l))
	}
	if m.CreatedAt != 0 {
		n += 1 + sovInfo(uint64(m.CreatedAt))
	}
	if m.LastUpdated != 0 {
		n += 1 + sovInfo(uint64(m.LastUpdated))
	}
	return n
}

func sovInfo(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozInfo(x uint64) (n int) {
	return sovInfo(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *AccountInfo) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowInfo
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: AccountInfo: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: AccountInfo: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Address", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowInfo
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthInfo
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthInfo
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Address = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Name", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowInfo
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthInfo
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthInfo
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Name = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Did", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowInfo
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthInfo
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthInfo
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Did = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field CoinType", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowInfo
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthInfo
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthInfo
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.CoinType = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 5:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ChainId", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowInfo
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthInfo
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthInfo
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.ChainId = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 6:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field PublicKey", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowInfo
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthInfo
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthInfo
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.PublicKey = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 7:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Type", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowInfo
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthInfo
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthInfo
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Type = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipInfo(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthInfo
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *PeerInfo) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowInfo
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: PeerInfo: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: PeerInfo: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Id", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowInfo
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthInfo
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthInfo
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Id = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Name", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowInfo
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthInfo
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthInfo
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Name = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field PeerId", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowInfo
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthInfo
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthInfo
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.PeerId = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Multiaddr", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowInfo
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthInfo
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthInfo
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Multiaddr = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipInfo(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthInfo
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *WalletInfo) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowInfo
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: WalletInfo: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: WalletInfo: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Controller", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowInfo
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthInfo
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthInfo
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Controller = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType == 0 {
				var v int32
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return ErrIntOverflowInfo
					}
					if iNdEx >= l {
						return io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					v |= int32(b&0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				m.DiscoveredPaths = append(m.DiscoveredPaths, v)
			} else if wireType == 2 {
				var packedLen int
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return ErrIntOverflowInfo
					}
					if iNdEx >= l {
						return io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					packedLen |= int(b&0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				if packedLen < 0 {
					return ErrInvalidLengthInfo
				}
				postIndex := iNdEx + packedLen
				if postIndex < 0 {
					return ErrInvalidLengthInfo
				}
				if postIndex > l {
					return io.ErrUnexpectedEOF
				}
				var elementCount int
				var count int
				for _, integer := range dAtA[iNdEx:postIndex] {
					if integer < 128 {
						count++
					}
				}
				elementCount = count
				if elementCount != 0 && len(m.DiscoveredPaths) == 0 {
					m.DiscoveredPaths = make([]int32, 0, elementCount)
				}
				for iNdEx < postIndex {
					var v int32
					for shift := uint(0); ; shift += 7 {
						if shift >= 64 {
							return ErrIntOverflowInfo
						}
						if iNdEx >= l {
							return io.ErrUnexpectedEOF
						}
						b := dAtA[iNdEx]
						iNdEx++
						v |= int32(b&0x7F) << shift
						if b < 0x80 {
							break
						}
					}
					m.DiscoveredPaths = append(m.DiscoveredPaths, v)
				}
			} else {
				return fmt.Errorf("proto: wrong wireType = %d for field DiscoveredPaths", wireType)
			}
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Algorithm", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowInfo
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthInfo
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthInfo
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Algorithm = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 4:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field CreatedAt", wireType)
			}
			m.CreatedAt = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowInfo
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.CreatedAt |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 5:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field LastUpdated", wireType)
			}
			m.LastUpdated = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowInfo
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.LastUpdated |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipInfo(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthInfo
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipInfo(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowInfo
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowInfo
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
		case 1:
			iNdEx += 8
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowInfo
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if length < 0 {
				return 0, ErrInvalidLengthInfo
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupInfo
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthInfo
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthInfo        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowInfo          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupInfo = fmt.Errorf("proto: unexpected end of group")
)