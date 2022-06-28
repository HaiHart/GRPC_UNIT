// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.21.1
// source: gRPCTalking.proto

package gRPCTalking

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

// GateWayClient is the client API for GateWay service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type GateWayClient interface {
	//send single block to add to chain
	SendBlock(ctx context.Context, in *Block, opts ...grpc.CallOption) (*Response, error)
	//propergate chain to connecting gateways (prefferably after updating chain)
	SendChain(ctx context.Context, opts ...grpc.CallOption) (GateWay_SendChainClient, error)
}

type gateWayClient struct {
	cc grpc.ClientConnInterface
}

func NewGateWayClient(cc grpc.ClientConnInterface) GateWayClient {
	return &gateWayClient{cc}
}

func (c *gateWayClient) SendBlock(ctx context.Context, in *Block, opts ...grpc.CallOption) (*Response, error) {
	out := new(Response)
	err := c.cc.Invoke(ctx, "/gRPCTalking.gateWay/sendBlock", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *gateWayClient) SendChain(ctx context.Context, opts ...grpc.CallOption) (GateWay_SendChainClient, error) {
	stream, err := c.cc.NewStream(ctx, &GateWay_ServiceDesc.Streams[0], "/gRPCTalking.gateWay/sendChain", opts...)
	if err != nil {
		return nil, err
	}
	x := &gateWaySendChainClient{stream}
	return x, nil
}

type GateWay_SendChainClient interface {
	Send(*Block) error
	CloseAndRecv() (*Response, error)
	grpc.ClientStream
}

type gateWaySendChainClient struct {
	grpc.ClientStream
}

func (x *gateWaySendChainClient) Send(m *Block) error {
	return x.ClientStream.SendMsg(m)
}

func (x *gateWaySendChainClient) CloseAndRecv() (*Response, error) {
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	m := new(Response)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// GateWayServer is the server API for GateWay service.
// All implementations must embed UnimplementedGateWayServer
// for forward compatibility
type GateWayServer interface {
	//send single block to add to chain
	SendBlock(context.Context, *Block) (*Response, error)
	//propergate chain to connecting gateways (prefferably after updating chain)
	SendChain(GateWay_SendChainServer) error
	mustEmbedUnimplementedGateWayServer()
}

// UnimplementedGateWayServer must be embedded to have forward compatible implementations.
type UnimplementedGateWayServer struct {
}

func (UnimplementedGateWayServer) SendBlock(context.Context, *Block) (*Response, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SendBlock not implemented")
}
func (UnimplementedGateWayServer) SendChain(GateWay_SendChainServer) error {
	return status.Errorf(codes.Unimplemented, "method SendChain not implemented")
}
func (UnimplementedGateWayServer) mustEmbedUnimplementedGateWayServer() {}

// UnsafeGateWayServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to GateWayServer will
// result in compilation errors.
type UnsafeGateWayServer interface {
	mustEmbedUnimplementedGateWayServer()
}

func RegisterGateWayServer(s grpc.ServiceRegistrar, srv GateWayServer) {
	s.RegisterService(&GateWay_ServiceDesc, srv)
}

func _GateWay_SendBlock_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Block)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GateWayServer).SendBlock(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/gRPCTalking.gateWay/sendBlock",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GateWayServer).SendBlock(ctx, req.(*Block))
	}
	return interceptor(ctx, in, info, handler)
}

func _GateWay_SendChain_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(GateWayServer).SendChain(&gateWaySendChainServer{stream})
}

type GateWay_SendChainServer interface {
	SendAndClose(*Response) error
	Recv() (*Block, error)
	grpc.ServerStream
}

type gateWaySendChainServer struct {
	grpc.ServerStream
}

func (x *gateWaySendChainServer) SendAndClose(m *Response) error {
	return x.ServerStream.SendMsg(m)
}

func (x *gateWaySendChainServer) Recv() (*Block, error) {
	m := new(Block)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// GateWay_ServiceDesc is the grpc.ServiceDesc for GateWay service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var GateWay_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "gRPCTalking.gateWay",
	HandlerType: (*GateWayServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "sendBlock",
			Handler:    _GateWay_SendBlock_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "sendChain",
			Handler:       _GateWay_SendChain_Handler,
			ClientStreams: true,
		},
	},
	Metadata: "gRPCTalking.proto",
}