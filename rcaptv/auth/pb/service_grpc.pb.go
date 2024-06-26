// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v4.23.4
// source: auth/pb/service.proto

package pb

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
	wrapperspb "google.golang.org/protobuf/types/known/wrapperspb"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// TokenValidatorServiceClient is the client API for TokenValidatorService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type TokenValidatorServiceClient interface {
	AddUser(ctx context.Context, in *User, opts ...grpc.CallOption) (*AddUserReply, error)
	Ping(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*wrapperspb.BoolValue, error)
}

type tokenValidatorServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewTokenValidatorServiceClient(cc grpc.ClientConnInterface) TokenValidatorServiceClient {
	return &tokenValidatorServiceClient{cc}
}

func (c *tokenValidatorServiceClient) AddUser(ctx context.Context, in *User, opts ...grpc.CallOption) (*AddUserReply, error) {
	out := new(AddUserReply)
	err := c.cc.Invoke(ctx, "/pb.TokenValidatorService/AddUser", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *tokenValidatorServiceClient) Ping(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*wrapperspb.BoolValue, error) {
	out := new(wrapperspb.BoolValue)
	err := c.cc.Invoke(ctx, "/pb.TokenValidatorService/Ping", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// TokenValidatorServiceServer is the server API for TokenValidatorService service.
// All implementations must embed UnimplementedTokenValidatorServiceServer
// for forward compatibility
type TokenValidatorServiceServer interface {
	AddUser(context.Context, *User) (*AddUserReply, error)
	Ping(context.Context, *emptypb.Empty) (*wrapperspb.BoolValue, error)
	mustEmbedUnimplementedTokenValidatorServiceServer()
}

// UnimplementedTokenValidatorServiceServer must be embedded to have forward compatible implementations.
type UnimplementedTokenValidatorServiceServer struct {
}

func (UnimplementedTokenValidatorServiceServer) AddUser(context.Context, *User) (*AddUserReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AddUser not implemented")
}
func (UnimplementedTokenValidatorServiceServer) Ping(context.Context, *emptypb.Empty) (*wrapperspb.BoolValue, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Ping not implemented")
}
func (UnimplementedTokenValidatorServiceServer) mustEmbedUnimplementedTokenValidatorServiceServer() {}

// UnsafeTokenValidatorServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to TokenValidatorServiceServer will
// result in compilation errors.
type UnsafeTokenValidatorServiceServer interface {
	mustEmbedUnimplementedTokenValidatorServiceServer()
}

func RegisterTokenValidatorServiceServer(s grpc.ServiceRegistrar, srv TokenValidatorServiceServer) {
	s.RegisterService(&TokenValidatorService_ServiceDesc, srv)
}

func _TokenValidatorService_AddUser_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(User)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TokenValidatorServiceServer).AddUser(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pb.TokenValidatorService/AddUser",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TokenValidatorServiceServer).AddUser(ctx, req.(*User))
	}
	return interceptor(ctx, in, info, handler)
}

func _TokenValidatorService_Ping_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TokenValidatorServiceServer).Ping(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pb.TokenValidatorService/Ping",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TokenValidatorServiceServer).Ping(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

// TokenValidatorService_ServiceDesc is the grpc.ServiceDesc for TokenValidatorService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var TokenValidatorService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "pb.TokenValidatorService",
	HandlerType: (*TokenValidatorServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "AddUser",
			Handler:    _TokenValidatorService_AddUser_Handler,
		},
		{
			MethodName: "Ping",
			Handler:    _TokenValidatorService_Ping_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "auth/pb/service.proto",
}
