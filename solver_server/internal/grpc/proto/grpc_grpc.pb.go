// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v5.26.1
// source: grpc.proto

package proto

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

const (
	GeometryService_PingSolver_FullMethodName = "/grpc.GeometryService/PingSolver"
)

// GeometryServiceClient is the client API for GeometryService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type GeometryServiceClient interface {
	PingSolver(ctx context.Context, in *Solver, opts ...grpc.CallOption) (*Empty, error)
}

type geometryServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewGeometryServiceClient(cc grpc.ClientConnInterface) GeometryServiceClient {
	return &geometryServiceClient{cc}
}

func (c *geometryServiceClient) PingSolver(ctx context.Context, in *Solver, opts ...grpc.CallOption) (*Empty, error) {
	out := new(Empty)
	err := c.cc.Invoke(ctx, GeometryService_PingSolver_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// GeometryServiceServer is the server API for GeometryService service.
// All implementations must embed UnimplementedGeometryServiceServer
// for forward compatibility
type GeometryServiceServer interface {
	PingSolver(context.Context, *Solver) (*Empty, error)
	mustEmbedUnimplementedGeometryServiceServer()
}

// UnimplementedGeometryServiceServer must be embedded to have forward compatible implementations.
type UnimplementedGeometryServiceServer struct {
}

func (UnimplementedGeometryServiceServer) PingSolver(context.Context, *Solver) (*Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PingSolver not implemented")
}
func (UnimplementedGeometryServiceServer) mustEmbedUnimplementedGeometryServiceServer() {}

// UnsafeGeometryServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to GeometryServiceServer will
// result in compilation errors.
type UnsafeGeometryServiceServer interface {
	mustEmbedUnimplementedGeometryServiceServer()
}

func RegisterGeometryServiceServer(s grpc.ServiceRegistrar, srv GeometryServiceServer) {
	s.RegisterService(&GeometryService_ServiceDesc, srv)
}

func _GeometryService_PingSolver_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Solver)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GeometryServiceServer).PingSolver(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: GeometryService_PingSolver_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GeometryServiceServer).PingSolver(ctx, req.(*Solver))
	}
	return interceptor(ctx, in, info, handler)
}

// GeometryService_ServiceDesc is the grpc.ServiceDesc for GeometryService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var GeometryService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "grpc.GeometryService",
	HandlerType: (*GeometryServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "PingSolver",
			Handler:    _GeometryService_PingSolver_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "grpc.proto",
}