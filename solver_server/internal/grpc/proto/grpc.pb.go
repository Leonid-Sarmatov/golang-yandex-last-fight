// Code generated by protoc-gen-go. DO NOT EDIT.
// source: grpc.proto

package proto

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

type Solver struct {
	SolverName           string   `protobuf:"bytes,1,opt,name=solverName,proto3" json:"solverName,omitempty"`
	SolvingNowExpression string   `protobuf:"bytes,2,opt,name=solvingNowExpression,proto3" json:"solvingNowExpression,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Solver) Reset()         { *m = Solver{} }
func (m *Solver) String() string { return proto.CompactTextString(m) }
func (*Solver) ProtoMessage()    {}
func (*Solver) Descriptor() ([]byte, []int) {
	return fileDescriptor_bedfbfc9b54e5600, []int{0}
}

func (m *Solver) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Solver.Unmarshal(m, b)
}
func (m *Solver) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Solver.Marshal(b, m, deterministic)
}
func (m *Solver) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Solver.Merge(m, src)
}
func (m *Solver) XXX_Size() int {
	return xxx_messageInfo_Solver.Size(m)
}
func (m *Solver) XXX_DiscardUnknown() {
	xxx_messageInfo_Solver.DiscardUnknown(m)
}

var xxx_messageInfo_Solver proto.InternalMessageInfo

func (m *Solver) GetSolverName() string {
	if m != nil {
		return m.SolverName
	}
	return ""
}

func (m *Solver) GetSolvingNowExpression() string {
	if m != nil {
		return m.SolvingNowExpression
	}
	return ""
}

type Empty struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Empty) Reset()         { *m = Empty{} }
func (m *Empty) String() string { return proto.CompactTextString(m) }
func (*Empty) ProtoMessage()    {}
func (*Empty) Descriptor() ([]byte, []int) {
	return fileDescriptor_bedfbfc9b54e5600, []int{1}
}

func (m *Empty) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Empty.Unmarshal(m, b)
}
func (m *Empty) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Empty.Marshal(b, m, deterministic)
}
func (m *Empty) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Empty.Merge(m, src)
}
func (m *Empty) XXX_Size() int {
	return xxx_messageInfo_Empty.Size(m)
}
func (m *Empty) XXX_DiscardUnknown() {
	xxx_messageInfo_Empty.DiscardUnknown(m)
}

var xxx_messageInfo_Empty proto.InternalMessageInfo

func init() {
	proto.RegisterType((*Solver)(nil), "grpc.Solver")
	proto.RegisterType((*Empty)(nil), "grpc.Empty")
}

func init() {
	proto.RegisterFile("grpc.proto", fileDescriptor_bedfbfc9b54e5600)
}

var fileDescriptor_bedfbfc9b54e5600 = []byte{
	// 227 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x6c, 0x90, 0x3f, 0x4b, 0x03, 0x41,
	0x10, 0xc5, 0x89, 0x68, 0xc4, 0x51, 0x10, 0x0e, 0x8b, 0x60, 0x21, 0x92, 0x46, 0x9b, 0xbb, 0x85,
	0xd8, 0x59, 0x0a, 0xc1, 0x46, 0x82, 0x78, 0x88, 0x20, 0x82, 0x6c, 0x2e, 0xe3, 0x66, 0xe1, 0x76,
	0xe6, 0x98, 0x1d, 0xd7, 0xdc, 0xb7, 0x97, 0xdb, 0xb3, 0xb0, 0x48, 0x35, 0xef, 0x37, 0x7f, 0x1f,
	0x03, 0xe0, 0xa4, 0x6b, 0xaa, 0x4e, 0x58, 0xb9, 0x38, 0x1c, 0xf4, 0xfc, 0x03, 0xa6, 0x35, 0xb7,
	0x09, 0xa5, 0xb8, 0x02, 0x88, 0x59, 0xad, 0x6c, 0xc0, 0xd9, 0xe4, 0x7a, 0x72, 0x7b, 0xf2, 0xf2,
	0x2f, 0x53, 0x2c, 0xe0, 0x62, 0x20, 0x4f, 0x6e, 0xc5, 0x3f, 0xcb, 0x5d, 0x27, 0x18, 0xa3, 0x67,
	0x9a, 0x1d, 0xe4, 0xce, 0xbd, 0xb5, 0xf9, 0x31, 0x1c, 0x2d, 0x43, 0xa7, 0xfd, 0xe2, 0x1e, 0xce,
	0x1f, 0x91, 0x03, 0xaa, 0xf4, 0x35, 0x4a, 0xf2, 0x0d, 0x16, 0x37, 0x00, 0xcf, 0x9e, 0xdc, 0xdf,
	0xf5, 0xb3, 0x2a, 0x5b, 0x1b, 0xe9, 0xf2, 0x74, 0xa4, 0x3c, 0xfb, 0xf0, 0xf6, 0xfe, 0xea, 0xbc,
	0x6e, 0xbf, 0xd7, 0x55, 0xc3, 0xc1, 0x3c, 0x21, 0x93, 0xdf, 0x94, 0xb5, 0x95, 0x60, 0x95, 0x93,
	0x71, 0xdc, 0x5a, 0x72, 0x65, 0x6f, 0x69, 0x83, 0xbb, 0xb2, 0xb5, 0x51, 0xcb, 0x2f, 0xef, 0xb6,
	0x6a, 0x46, 0xeb, 0x9f, 0x11, 0x25, 0xa1, 0x18, 0x4f, 0x8a, 0x42, 0xb6, 0x35, 0xc3, 0x62, 0x93,
	0x3f, 0xb0, 0x9e, 0xe6, 0x70, 0xf7, 0x1b, 0x00, 0x00, 0xff, 0xff, 0x81, 0xd1, 0x5a, 0xa8, 0x16,
	0x01, 0x00, 0x00,
}
