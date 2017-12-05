// Code generated by protoc-gen-go. DO NOT EDIT.
// source: server.proto

/*
Package keystore is a generated protocol buffer package.

It is generated from these files:
	server.proto

It has these top-level messages:
	Empty
	SaveRequest
	ReadRequest
	StoreMeta
	ReadResponse
	GetDirectoryRequest
	GetDirectoryResponse
*/
package keystore

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import google_protobuf "github.com/golang/protobuf/ptypes/any"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type Empty struct {
}

func (m *Empty) Reset()                    { *m = Empty{} }
func (m *Empty) String() string            { return proto.CompactTextString(m) }
func (*Empty) ProtoMessage()               {}
func (*Empty) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

type SaveRequest struct {
	Key          string               `protobuf:"bytes,1,opt,name=key" json:"key,omitempty"`
	Value        *google_protobuf.Any `protobuf:"bytes,2,opt,name=value" json:"value,omitempty"`
	WriteVersion int64                `protobuf:"varint,3,opt,name=write_version,json=writeVersion" json:"write_version,omitempty"`
}

func (m *SaveRequest) Reset()                    { *m = SaveRequest{} }
func (m *SaveRequest) String() string            { return proto.CompactTextString(m) }
func (*SaveRequest) ProtoMessage()               {}
func (*SaveRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *SaveRequest) GetKey() string {
	if m != nil {
		return m.Key
	}
	return ""
}

func (m *SaveRequest) GetValue() *google_protobuf.Any {
	if m != nil {
		return m.Value
	}
	return nil
}

func (m *SaveRequest) GetWriteVersion() int64 {
	if m != nil {
		return m.WriteVersion
	}
	return 0
}

type ReadRequest struct {
	Key string `protobuf:"bytes,1,opt,name=key" json:"key,omitempty"`
}

func (m *ReadRequest) Reset()                    { *m = ReadRequest{} }
func (m *ReadRequest) String() string            { return proto.CompactTextString(m) }
func (*ReadRequest) ProtoMessage()               {}
func (*ReadRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *ReadRequest) GetKey() string {
	if m != nil {
		return m.Key
	}
	return ""
}

type StoreMeta struct {
	Version int64 `protobuf:"varint,1,opt,name=version" json:"version,omitempty"`
}

func (m *StoreMeta) Reset()                    { *m = StoreMeta{} }
func (m *StoreMeta) String() string            { return proto.CompactTextString(m) }
func (*StoreMeta) ProtoMessage()               {}
func (*StoreMeta) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

func (m *StoreMeta) GetVersion() int64 {
	if m != nil {
		return m.Version
	}
	return 0
}

type ReadResponse struct {
	Payload  *google_protobuf.Any `protobuf:"bytes,1,opt,name=payload" json:"payload,omitempty"`
	ReadTime int64                `protobuf:"varint,2,opt,name=read_time,json=readTime" json:"read_time,omitempty"`
}

func (m *ReadResponse) Reset()                    { *m = ReadResponse{} }
func (m *ReadResponse) String() string            { return proto.CompactTextString(m) }
func (*ReadResponse) ProtoMessage()               {}
func (*ReadResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

func (m *ReadResponse) GetPayload() *google_protobuf.Any {
	if m != nil {
		return m.Payload
	}
	return nil
}

func (m *ReadResponse) GetReadTime() int64 {
	if m != nil {
		return m.ReadTime
	}
	return 0
}

type GetDirectoryRequest struct {
}

func (m *GetDirectoryRequest) Reset()                    { *m = GetDirectoryRequest{} }
func (m *GetDirectoryRequest) String() string            { return proto.CompactTextString(m) }
func (*GetDirectoryRequest) ProtoMessage()               {}
func (*GetDirectoryRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{5} }

type GetDirectoryResponse struct {
	Keys    []string `protobuf:"bytes,1,rep,name=keys" json:"keys,omitempty"`
	Version int64    `protobuf:"varint,2,opt,name=version" json:"version,omitempty"`
}

func (m *GetDirectoryResponse) Reset()                    { *m = GetDirectoryResponse{} }
func (m *GetDirectoryResponse) String() string            { return proto.CompactTextString(m) }
func (*GetDirectoryResponse) ProtoMessage()               {}
func (*GetDirectoryResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{6} }

func (m *GetDirectoryResponse) GetKeys() []string {
	if m != nil {
		return m.Keys
	}
	return nil
}

func (m *GetDirectoryResponse) GetVersion() int64 {
	if m != nil {
		return m.Version
	}
	return 0
}

func init() {
	proto.RegisterType((*Empty)(nil), "keystore.Empty")
	proto.RegisterType((*SaveRequest)(nil), "keystore.SaveRequest")
	proto.RegisterType((*ReadRequest)(nil), "keystore.ReadRequest")
	proto.RegisterType((*StoreMeta)(nil), "keystore.StoreMeta")
	proto.RegisterType((*ReadResponse)(nil), "keystore.ReadResponse")
	proto.RegisterType((*GetDirectoryRequest)(nil), "keystore.GetDirectoryRequest")
	proto.RegisterType((*GetDirectoryResponse)(nil), "keystore.GetDirectoryResponse")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for KeyStoreService service

type KeyStoreServiceClient interface {
	Save(ctx context.Context, in *SaveRequest, opts ...grpc.CallOption) (*Empty, error)
	Read(ctx context.Context, in *ReadRequest, opts ...grpc.CallOption) (*ReadResponse, error)
	GetMeta(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*StoreMeta, error)
	GetDirectory(ctx context.Context, in *GetDirectoryRequest, opts ...grpc.CallOption) (*GetDirectoryResponse, error)
}

type keyStoreServiceClient struct {
	cc *grpc.ClientConn
}

func NewKeyStoreServiceClient(cc *grpc.ClientConn) KeyStoreServiceClient {
	return &keyStoreServiceClient{cc}
}

func (c *keyStoreServiceClient) Save(ctx context.Context, in *SaveRequest, opts ...grpc.CallOption) (*Empty, error) {
	out := new(Empty)
	err := grpc.Invoke(ctx, "/keystore.KeyStoreService/Save", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *keyStoreServiceClient) Read(ctx context.Context, in *ReadRequest, opts ...grpc.CallOption) (*ReadResponse, error) {
	out := new(ReadResponse)
	err := grpc.Invoke(ctx, "/keystore.KeyStoreService/Read", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *keyStoreServiceClient) GetMeta(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*StoreMeta, error) {
	out := new(StoreMeta)
	err := grpc.Invoke(ctx, "/keystore.KeyStoreService/GetMeta", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *keyStoreServiceClient) GetDirectory(ctx context.Context, in *GetDirectoryRequest, opts ...grpc.CallOption) (*GetDirectoryResponse, error) {
	out := new(GetDirectoryResponse)
	err := grpc.Invoke(ctx, "/keystore.KeyStoreService/GetDirectory", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for KeyStoreService service

type KeyStoreServiceServer interface {
	Save(context.Context, *SaveRequest) (*Empty, error)
	Read(context.Context, *ReadRequest) (*ReadResponse, error)
	GetMeta(context.Context, *Empty) (*StoreMeta, error)
	GetDirectory(context.Context, *GetDirectoryRequest) (*GetDirectoryResponse, error)
}

func RegisterKeyStoreServiceServer(s *grpc.Server, srv KeyStoreServiceServer) {
	s.RegisterService(&_KeyStoreService_serviceDesc, srv)
}

func _KeyStoreService_Save_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SaveRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(KeyStoreServiceServer).Save(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/keystore.KeyStoreService/Save",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(KeyStoreServiceServer).Save(ctx, req.(*SaveRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _KeyStoreService_Read_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ReadRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(KeyStoreServiceServer).Read(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/keystore.KeyStoreService/Read",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(KeyStoreServiceServer).Read(ctx, req.(*ReadRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _KeyStoreService_GetMeta_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(KeyStoreServiceServer).GetMeta(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/keystore.KeyStoreService/GetMeta",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(KeyStoreServiceServer).GetMeta(ctx, req.(*Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _KeyStoreService_GetDirectory_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetDirectoryRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(KeyStoreServiceServer).GetDirectory(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/keystore.KeyStoreService/GetDirectory",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(KeyStoreServiceServer).GetDirectory(ctx, req.(*GetDirectoryRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _KeyStoreService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "keystore.KeyStoreService",
	HandlerType: (*KeyStoreServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Save",
			Handler:    _KeyStoreService_Save_Handler,
		},
		{
			MethodName: "Read",
			Handler:    _KeyStoreService_Read_Handler,
		},
		{
			MethodName: "GetMeta",
			Handler:    _KeyStoreService_GetMeta_Handler,
		},
		{
			MethodName: "GetDirectory",
			Handler:    _KeyStoreService_GetDirectory_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "server.proto",
}

func init() { proto.RegisterFile("server.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 377 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x74, 0x51, 0x41, 0x8f, 0xd2, 0x40,
	0x18, 0xa5, 0x14, 0x2c, 0xfd, 0xa8, 0xc1, 0x0c, 0x60, 0x6a, 0x8d, 0xda, 0x8c, 0x31, 0x69, 0x3c,
	0x14, 0xc5, 0x83, 0x67, 0x13, 0x0c, 0x07, 0x63, 0x4c, 0x8a, 0xf1, 0xe2, 0x81, 0x0c, 0xf0, 0x2d,
	0x69, 0x80, 0x4e, 0x77, 0x3a, 0x74, 0x33, 0xb7, 0xfd, 0xe9, 0x9b, 0xce, 0xd0, 0x50, 0xc8, 0x72,
	0xfb, 0xfa, 0xfa, 0xe6, 0x7d, 0xef, 0x7b, 0x0f, 0xbc, 0x02, 0x45, 0x89, 0x22, 0xce, 0x05, 0x97,
	0x9c, 0xf4, 0x76, 0xa8, 0x0a, 0xc9, 0x05, 0x06, 0x6f, 0xb6, 0x9c, 0x6f, 0xf7, 0x38, 0xd1, 0xf8,
	0xea, 0x78, 0x37, 0x61, 0x99, 0x32, 0x24, 0xea, 0x40, 0xf7, 0xe7, 0x21, 0x97, 0x8a, 0xe6, 0xd0,
	0x5f, 0xb0, 0x12, 0x13, 0xbc, 0x3f, 0x62, 0x21, 0xc9, 0x2b, 0xb0, 0x77, 0xa8, 0x7c, 0x2b, 0xb4,
	0x22, 0x37, 0xa9, 0x46, 0xf2, 0x19, 0xba, 0x25, 0xdb, 0x1f, 0xd1, 0x6f, 0x87, 0x56, 0xd4, 0x9f,
	0x8e, 0x62, 0x23, 0x1a, 0xd7, 0xa2, 0xf1, 0x8f, 0x4c, 0x25, 0x86, 0x42, 0x3e, 0xc2, 0xcb, 0x07,
	0x91, 0x4a, 0x5c, 0x96, 0x28, 0x8a, 0x94, 0x67, 0xbe, 0x1d, 0x5a, 0x91, 0x9d, 0x78, 0x1a, 0xfc,
	0x67, 0x30, 0xfa, 0x01, 0xfa, 0x09, 0xb2, 0xcd, 0xcd, 0x8d, 0xf4, 0x13, 0xb8, 0x8b, 0xca, 0xff,
	0x6f, 0x94, 0x8c, 0xf8, 0xe0, 0xd4, 0x62, 0x96, 0x16, 0xab, 0x3f, 0xe9, 0x7f, 0xf0, 0x8c, 0x4e,
	0x91, 0xf3, 0xac, 0x40, 0x12, 0x83, 0x93, 0x33, 0xb5, 0xe7, 0x6c, 0xa3, 0x99, 0xb7, 0xac, 0xd6,
	0x24, 0xf2, 0x16, 0x5c, 0x81, 0x6c, 0xb3, 0x94, 0xe9, 0xc1, 0x1c, 0x67, 0x27, 0xbd, 0x0a, 0xf8,
	0x9b, 0x1e, 0x90, 0x8e, 0x61, 0x38, 0x47, 0x39, 0x4b, 0x05, 0xae, 0x25, 0x17, 0xea, 0x64, 0x96,
	0xce, 0x60, 0x74, 0x09, 0x9f, 0x76, 0x13, 0xe8, 0x54, 0xa9, 0xfb, 0x56, 0x68, 0x47, 0x6e, 0xa2,
	0xe7, 0xa6, 0xf3, 0xf6, 0x85, 0xf3, 0xe9, 0x63, 0x1b, 0x06, 0xbf, 0x50, 0xe9, 0x23, 0x17, 0x28,
	0xca, 0x74, 0x8d, 0xe4, 0x0b, 0x74, 0xaa, 0x1e, 0xc8, 0x38, 0xae, 0xeb, 0x8b, 0x1b, 0xbd, 0x04,
	0x83, 0x33, 0x6c, 0x7a, 0x6b, 0x91, 0xef, 0xd0, 0xa9, 0xee, 0x6f, 0xbe, 0x68, 0xe4, 0x1a, 0xbc,
	0xbe, 0x86, 0x8d, 0x55, 0xda, 0x22, 0x5f, 0xc1, 0x99, 0xa3, 0xd4, 0xe9, 0x5e, 0xcb, 0x06, 0xc3,
	0xc6, 0xfa, 0xba, 0x03, 0xda, 0x22, 0x7f, 0xc0, 0x6b, 0xde, 0x4d, 0xde, 0x9d, 0x69, 0xcf, 0xc4,
	0x14, 0xbc, 0xbf, 0xf5, 0xbb, 0xf6, 0xb0, 0x7a, 0xa1, 0x3b, 0xf9, 0xf6, 0x14, 0x00, 0x00, 0xff,
	0xff, 0x66, 0xa5, 0xb2, 0x43, 0xbb, 0x02, 0x00, 0x00,
}
