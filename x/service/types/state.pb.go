// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: service/v1/state.proto

package types

import (
	_ "cosmossdk.io/orm"
	fmt "fmt"
	proto "github.com/cosmos/gogoproto/proto"
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

type URI_Protocol int32

const (
	URI_HTTPS URI_Protocol = 0
	URI_IPFS  URI_Protocol = 1
)

var URI_Protocol_name = map[int32]string{
	0: "HTTPS",
	1: "IPFS",
}

var URI_Protocol_value = map[string]int32{
	"HTTPS": 0,
	"IPFS":  1,
}

func (x URI_Protocol) String() string {
	return proto.EnumName(URI_Protocol_name, int32(x))
}

func (URI_Protocol) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_ab6e3654a2974847, []int{2, 0}
}

type Metadata struct {
	Id          uint64   `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	Origin      string   `protobuf:"bytes,2,opt,name=origin,proto3" json:"origin,omitempty"`
	Name        string   `protobuf:"bytes,3,opt,name=name,proto3" json:"name,omitempty"`
	Description string   `protobuf:"bytes,4,opt,name=description,proto3" json:"description,omitempty"`
	Category    string   `protobuf:"bytes,5,opt,name=category,proto3" json:"category,omitempty"`
	Icon        *URI     `protobuf:"bytes,6,opt,name=icon,proto3" json:"icon,omitempty"`
	Tags        []string `protobuf:"bytes,7,rep,name=tags,proto3" json:"tags,omitempty"`
}

func (m *Metadata) Reset()         { *m = Metadata{} }
func (m *Metadata) String() string { return proto.CompactTextString(m) }
func (*Metadata) ProtoMessage()    {}
func (*Metadata) Descriptor() ([]byte, []int) {
	return fileDescriptor_ab6e3654a2974847, []int{0}
}
func (m *Metadata) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *Metadata) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_Metadata.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *Metadata) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Metadata.Merge(m, src)
}
func (m *Metadata) XXX_Size() int {
	return m.Size()
}
func (m *Metadata) XXX_DiscardUnknown() {
	xxx_messageInfo_Metadata.DiscardUnknown(m)
}

var xxx_messageInfo_Metadata proto.InternalMessageInfo

func (m *Metadata) GetId() uint64 {
	if m != nil {
		return m.Id
	}
	return 0
}

func (m *Metadata) GetOrigin() string {
	if m != nil {
		return m.Origin
	}
	return ""
}

func (m *Metadata) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *Metadata) GetDescription() string {
	if m != nil {
		return m.Description
	}
	return ""
}

func (m *Metadata) GetCategory() string {
	if m != nil {
		return m.Category
	}
	return ""
}

func (m *Metadata) GetIcon() *URI {
	if m != nil {
		return m.Icon
	}
	return nil
}

func (m *Metadata) GetTags() []string {
	if m != nil {
		return m.Tags
	}
	return nil
}

// Profile represents a DID alias
type Profile struct {
	// The unique identifier of the alias
	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	// The alias of the DID
	Subject string `protobuf:"bytes,2,opt,name=subject,proto3" json:"subject,omitempty"`
	// Origin of the alias
	Origin string `protobuf:"bytes,3,opt,name=origin,proto3" json:"origin,omitempty"`
	// Controller of the alias
	Controller string `protobuf:"bytes,4,opt,name=controller,proto3" json:"controller,omitempty"`
}

func (m *Profile) Reset()         { *m = Profile{} }
func (m *Profile) String() string { return proto.CompactTextString(m) }
func (*Profile) ProtoMessage()    {}
func (*Profile) Descriptor() ([]byte, []int) {
	return fileDescriptor_ab6e3654a2974847, []int{1}
}
func (m *Profile) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *Profile) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_Profile.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *Profile) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Profile.Merge(m, src)
}
func (m *Profile) XXX_Size() int {
	return m.Size()
}
func (m *Profile) XXX_DiscardUnknown() {
	xxx_messageInfo_Profile.DiscardUnknown(m)
}

var xxx_messageInfo_Profile proto.InternalMessageInfo

func (m *Profile) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *Profile) GetSubject() string {
	if m != nil {
		return m.Subject
	}
	return ""
}

func (m *Profile) GetOrigin() string {
	if m != nil {
		return m.Origin
	}
	return ""
}

func (m *Profile) GetController() string {
	if m != nil {
		return m.Controller
	}
	return ""
}

type URI struct {
	Protocol URI_Protocol `protobuf:"varint,1,opt,name=protocol,proto3,enum=service.v1.URI_Protocol" json:"protocol,omitempty"`
	Uri      string       `protobuf:"bytes,2,opt,name=uri,proto3" json:"uri,omitempty"`
}

func (m *URI) Reset()         { *m = URI{} }
func (m *URI) String() string { return proto.CompactTextString(m) }
func (*URI) ProtoMessage()    {}
func (*URI) Descriptor() ([]byte, []int) {
	return fileDescriptor_ab6e3654a2974847, []int{2}
}
func (m *URI) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *URI) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_URI.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *URI) XXX_Merge(src proto.Message) {
	xxx_messageInfo_URI.Merge(m, src)
}
func (m *URI) XXX_Size() int {
	return m.Size()
}
func (m *URI) XXX_DiscardUnknown() {
	xxx_messageInfo_URI.DiscardUnknown(m)
}

var xxx_messageInfo_URI proto.InternalMessageInfo

func (m *URI) GetProtocol() URI_Protocol {
	if m != nil {
		return m.Protocol
	}
	return URI_HTTPS
}

func (m *URI) GetUri() string {
	if m != nil {
		return m.Uri
	}
	return ""
}

func init() {
	proto.RegisterEnum("service.v1.URI_Protocol", URI_Protocol_name, URI_Protocol_value)
	proto.RegisterType((*Metadata)(nil), "service.v1.Metadata")
	proto.RegisterType((*Profile)(nil), "service.v1.Profile")
	proto.RegisterType((*URI)(nil), "service.v1.URI")
}

func init() { proto.RegisterFile("service/v1/state.proto", fileDescriptor_ab6e3654a2974847) }

var fileDescriptor_ab6e3654a2974847 = []byte{
	// 422 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x5c, 0x52, 0xc1, 0x8e, 0xd3, 0x30,
	0x14, 0xac, 0x9b, 0x6c, 0x9b, 0xbe, 0x45, 0x25, 0xb2, 0xd0, 0x62, 0xed, 0xc1, 0x44, 0x05, 0xa1,
	0x1e, 0x50, 0xa2, 0x5d, 0x38, 0xed, 0x09, 0x71, 0x40, 0xf4, 0x80, 0x54, 0x65, 0x77, 0x2f, 0xdc,
	0x52, 0xd7, 0x14, 0xa3, 0xc4, 0xaf, 0xb2, 0xdd, 0x8a, 0xfd, 0x09, 0x04, 0x3f, 0xc0, 0xf7, 0x70,
	0x5c, 0x89, 0x0b, 0x17, 0x24, 0xd4, 0xfe, 0x01, 0x5f, 0x80, 0x62, 0x92, 0x12, 0xb8, 0x58, 0xef,
	0xcd, 0x8c, 0x3c, 0x33, 0xb2, 0xe1, 0xc4, 0x4a, 0xb3, 0x55, 0x42, 0x66, 0xdb, 0xb3, 0xcc, 0xba,
	0xc2, 0xc9, 0x74, 0x6d, 0xd0, 0x21, 0x85, 0x06, 0x4f, 0xb7, 0x67, 0xa7, 0xf7, 0x05, 0xda, 0x0a,
	0x6d, 0x86, 0xa6, 0xaa, 0x65, 0x68, 0xaa, 0x3f, 0xa2, 0xc9, 0x0f, 0x02, 0xd1, 0x6b, 0xe9, 0x8a,
	0x65, 0xe1, 0x0a, 0x3a, 0x86, 0xbe, 0x5a, 0x32, 0x92, 0x90, 0x69, 0x98, 0xf7, 0xd5, 0x92, 0x9e,
	0xc0, 0x00, 0x8d, 0x5a, 0x29, 0xcd, 0xfa, 0x09, 0x99, 0x8e, 0xf2, 0x66, 0xa3, 0x14, 0x42, 0x5d,
	0x54, 0x92, 0x05, 0x1e, 0xf5, 0x33, 0x4d, 0xe0, 0x78, 0x29, 0xad, 0x30, 0x6a, 0xed, 0x14, 0x6a,
	0x16, 0x7a, 0xaa, 0x0b, 0xd1, 0x53, 0x88, 0x44, 0xe1, 0xe4, 0x0a, 0xcd, 0x0d, 0x3b, 0xf2, 0xf4,
	0x61, 0xa7, 0x0f, 0x21, 0x54, 0x02, 0x35, 0x1b, 0x24, 0x64, 0x7a, 0x7c, 0x7e, 0x37, 0xfd, 0x1b,
	0x3d, 0xbd, 0xce, 0x67, 0xb9, 0x27, 0x6b, 0x5b, 0x57, 0xac, 0x2c, 0x1b, 0x26, 0x41, 0x6d, 0x5b,
	0xcf, 0x17, 0xfc, 0xd7, 0x97, 0x6f, 0x1f, 0x03, 0x06, 0x83, 0x3a, 0x7a, 0x4c, 0xe8, 0x9d, 0x36,
	0x72, 0x4c, 0x18, 0x61, 0x64, 0xf2, 0x99, 0xc0, 0x70, 0x6e, 0xf0, 0xad, 0x2a, 0x65, 0xa7, 0xde,
	0xc8, 0xd7, 0x63, 0x30, 0xb4, 0x9b, 0xc5, 0x7b, 0x29, 0x5c, 0xd3, 0xaf, 0x5d, 0x3b, 0xc5, 0x83,
	0x7f, 0x8a, 0x73, 0x00, 0x81, 0xda, 0x19, 0x2c, 0x4b, 0x69, 0x9a, 0x8e, 0x1d, 0xe4, 0xe2, 0x91,
	0x4f, 0xc3, 0x21, 0xac, 0x9d, 0xe8, 0x3d, 0x18, 0x37, 0x17, 0x3e, 0xe9, 0x64, 0xea, 0x4f, 0x34,
	0x04, 0xd7, 0xf9, 0x8c, 0x3e, 0x83, 0xc8, 0xbf, 0x81, 0xc0, 0xd2, 0x87, 0x1a, 0x9f, 0xb3, 0xff,
	0x7a, 0xa7, 0xf3, 0x86, 0xcf, 0x0f, 0x4a, 0x1a, 0x43, 0xb0, 0x31, 0xaa, 0x09, 0x5c, 0x8f, 0x93,
	0x07, 0x10, 0xb5, 0x3a, 0x3a, 0x82, 0xa3, 0x57, 0x57, 0x57, 0xf3, 0xcb, 0xb8, 0x47, 0x23, 0x08,
	0x67, 0xf3, 0x97, 0x97, 0x31, 0x79, 0xf1, 0xfc, 0xeb, 0x8e, 0x93, 0xdb, 0x1d, 0x27, 0x3f, 0x77,
	0x9c, 0x7c, 0xda, 0xf3, 0xde, 0xed, 0x9e, 0xf7, 0xbe, 0xef, 0x79, 0xef, 0xcd, 0xe3, 0x95, 0x72,
	0xef, 0x36, 0x8b, 0x54, 0x60, 0x95, 0xa1, 0xb6, 0xa8, 0x4d, 0xe6, 0x8f, 0x0f, 0x59, 0xfb, 0xa7,
	0xdc, 0xcd, 0x5a, 0xda, 0xc5, 0xc0, 0xdb, 0x3f, 0xfd, 0x1d, 0x00, 0x00, 0xff, 0xff, 0xf1, 0xad,
	0x51, 0x19, 0x6b, 0x02, 0x00, 0x00,
}

func (m *Metadata) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Metadata) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Metadata) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Tags) > 0 {
		for iNdEx := len(m.Tags) - 1; iNdEx >= 0; iNdEx-- {
			i -= len(m.Tags[iNdEx])
			copy(dAtA[i:], m.Tags[iNdEx])
			i = encodeVarintState(dAtA, i, uint64(len(m.Tags[iNdEx])))
			i--
			dAtA[i] = 0x3a
		}
	}
	if m.Icon != nil {
		{
			size, err := m.Icon.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintState(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0x32
	}
	if len(m.Category) > 0 {
		i -= len(m.Category)
		copy(dAtA[i:], m.Category)
		i = encodeVarintState(dAtA, i, uint64(len(m.Category)))
		i--
		dAtA[i] = 0x2a
	}
	if len(m.Description) > 0 {
		i -= len(m.Description)
		copy(dAtA[i:], m.Description)
		i = encodeVarintState(dAtA, i, uint64(len(m.Description)))
		i--
		dAtA[i] = 0x22
	}
	if len(m.Name) > 0 {
		i -= len(m.Name)
		copy(dAtA[i:], m.Name)
		i = encodeVarintState(dAtA, i, uint64(len(m.Name)))
		i--
		dAtA[i] = 0x1a
	}
	if len(m.Origin) > 0 {
		i -= len(m.Origin)
		copy(dAtA[i:], m.Origin)
		i = encodeVarintState(dAtA, i, uint64(len(m.Origin)))
		i--
		dAtA[i] = 0x12
	}
	if m.Id != 0 {
		i = encodeVarintState(dAtA, i, uint64(m.Id))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func (m *Profile) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Profile) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Profile) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Controller) > 0 {
		i -= len(m.Controller)
		copy(dAtA[i:], m.Controller)
		i = encodeVarintState(dAtA, i, uint64(len(m.Controller)))
		i--
		dAtA[i] = 0x22
	}
	if len(m.Origin) > 0 {
		i -= len(m.Origin)
		copy(dAtA[i:], m.Origin)
		i = encodeVarintState(dAtA, i, uint64(len(m.Origin)))
		i--
		dAtA[i] = 0x1a
	}
	if len(m.Subject) > 0 {
		i -= len(m.Subject)
		copy(dAtA[i:], m.Subject)
		i = encodeVarintState(dAtA, i, uint64(len(m.Subject)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.Id) > 0 {
		i -= len(m.Id)
		copy(dAtA[i:], m.Id)
		i = encodeVarintState(dAtA, i, uint64(len(m.Id)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *URI) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *URI) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *URI) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Uri) > 0 {
		i -= len(m.Uri)
		copy(dAtA[i:], m.Uri)
		i = encodeVarintState(dAtA, i, uint64(len(m.Uri)))
		i--
		dAtA[i] = 0x12
	}
	if m.Protocol != 0 {
		i = encodeVarintState(dAtA, i, uint64(m.Protocol))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func encodeVarintState(dAtA []byte, offset int, v uint64) int {
	offset -= sovState(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *Metadata) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Id != 0 {
		n += 1 + sovState(uint64(m.Id))
	}
	l = len(m.Origin)
	if l > 0 {
		n += 1 + l + sovState(uint64(l))
	}
	l = len(m.Name)
	if l > 0 {
		n += 1 + l + sovState(uint64(l))
	}
	l = len(m.Description)
	if l > 0 {
		n += 1 + l + sovState(uint64(l))
	}
	l = len(m.Category)
	if l > 0 {
		n += 1 + l + sovState(uint64(l))
	}
	if m.Icon != nil {
		l = m.Icon.Size()
		n += 1 + l + sovState(uint64(l))
	}
	if len(m.Tags) > 0 {
		for _, s := range m.Tags {
			l = len(s)
			n += 1 + l + sovState(uint64(l))
		}
	}
	return n
}

func (m *Profile) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Id)
	if l > 0 {
		n += 1 + l + sovState(uint64(l))
	}
	l = len(m.Subject)
	if l > 0 {
		n += 1 + l + sovState(uint64(l))
	}
	l = len(m.Origin)
	if l > 0 {
		n += 1 + l + sovState(uint64(l))
	}
	l = len(m.Controller)
	if l > 0 {
		n += 1 + l + sovState(uint64(l))
	}
	return n
}

func (m *URI) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Protocol != 0 {
		n += 1 + sovState(uint64(m.Protocol))
	}
	l = len(m.Uri)
	if l > 0 {
		n += 1 + l + sovState(uint64(l))
	}
	return n
}

func sovState(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozState(x uint64) (n int) {
	return sovState(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *Metadata) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowState
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
			return fmt.Errorf("proto: Metadata: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Metadata: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Id", wireType)
			}
			m.Id = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowState
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Id |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Origin", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowState
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
				return ErrInvalidLengthState
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthState
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Origin = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Name", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowState
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
				return ErrInvalidLengthState
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthState
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Name = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Description", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowState
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
				return ErrInvalidLengthState
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthState
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Description = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 5:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Category", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowState
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
				return ErrInvalidLengthState
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthState
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Category = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 6:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Icon", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowState
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthState
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthState
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.Icon == nil {
				m.Icon = &URI{}
			}
			if err := m.Icon.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 7:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Tags", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowState
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
				return ErrInvalidLengthState
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthState
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Tags = append(m.Tags, string(dAtA[iNdEx:postIndex]))
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipState(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthState
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
func (m *Profile) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowState
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
			return fmt.Errorf("proto: Profile: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Profile: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Id", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowState
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
				return ErrInvalidLengthState
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthState
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Id = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Subject", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowState
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
				return ErrInvalidLengthState
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthState
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Subject = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Origin", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowState
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
				return ErrInvalidLengthState
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthState
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Origin = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Controller", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowState
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
				return ErrInvalidLengthState
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthState
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Controller = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipState(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthState
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
func (m *URI) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowState
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
			return fmt.Errorf("proto: URI: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: URI: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Protocol", wireType)
			}
			m.Protocol = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowState
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Protocol |= URI_Protocol(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Uri", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowState
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
				return ErrInvalidLengthState
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthState
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Uri = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipState(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthState
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
func skipState(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowState
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
					return 0, ErrIntOverflowState
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
					return 0, ErrIntOverflowState
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
				return 0, ErrInvalidLengthState
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupState
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthState
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthState        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowState          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupState = fmt.Errorf("proto: unexpected end of group")
)