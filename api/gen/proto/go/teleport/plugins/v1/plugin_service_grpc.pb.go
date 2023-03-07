// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             (unknown)
// source: teleport/plugins/v1/plugin_service.proto

package v1

import (
	context "context"
	types "github.com/gravitational/teleport/api/types"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// PluginServiceClient is the client API for PluginService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type PluginServiceClient interface {
	// CreatePlugin creates a new plugin instance.
	CreatePlugin(ctx context.Context, in *CreatePluginRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	// GetPlugin returns a plugin instance by name.
	GetPlugin(ctx context.Context, in *GetPluginRequest, opts ...grpc.CallOption) (*types.PluginV1, error)
	// DeletePlugin removes the specified plugin instance.
	DeletePlugin(ctx context.Context, in *DeletePluginRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	// ListPlugins returns a paginated view of plugin instances.
	ListPlugins(ctx context.Context, in *ListPluginsRequest, opts ...grpc.CallOption) (*ListPluginsResponse, error)
	// SetPluginCredentials sets the credentials for the given plugin.
	SetPluginCredentials(ctx context.Context, in *SetPluginCredentialsRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	// SetPluginCredentials sets the status for the given plugin.
	SetPluginStatus(ctx context.Context, in *SetPluginStatusRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	// GetAvailablePluginTypes returns the types of plugins
	// that the auth server supports onboarding.
	GetAvailablePluginTypes(ctx context.Context, in *GetAvailablePluginTypesRequest, opts ...grpc.CallOption) (*GetAvailablePluginTypesResponse, error)
}

type pluginServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewPluginServiceClient(cc grpc.ClientConnInterface) PluginServiceClient {
	return &pluginServiceClient{cc}
}

func (c *pluginServiceClient) CreatePlugin(ctx context.Context, in *CreatePluginRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/teleport.plugins.v1.PluginService/CreatePlugin", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *pluginServiceClient) GetPlugin(ctx context.Context, in *GetPluginRequest, opts ...grpc.CallOption) (*types.PluginV1, error) {
	out := new(types.PluginV1)
	err := c.cc.Invoke(ctx, "/teleport.plugins.v1.PluginService/GetPlugin", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *pluginServiceClient) DeletePlugin(ctx context.Context, in *DeletePluginRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/teleport.plugins.v1.PluginService/DeletePlugin", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *pluginServiceClient) ListPlugins(ctx context.Context, in *ListPluginsRequest, opts ...grpc.CallOption) (*ListPluginsResponse, error) {
	out := new(ListPluginsResponse)
	err := c.cc.Invoke(ctx, "/teleport.plugins.v1.PluginService/ListPlugins", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *pluginServiceClient) SetPluginCredentials(ctx context.Context, in *SetPluginCredentialsRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/teleport.plugins.v1.PluginService/SetPluginCredentials", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *pluginServiceClient) SetPluginStatus(ctx context.Context, in *SetPluginStatusRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/teleport.plugins.v1.PluginService/SetPluginStatus", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *pluginServiceClient) GetAvailablePluginTypes(ctx context.Context, in *GetAvailablePluginTypesRequest, opts ...grpc.CallOption) (*GetAvailablePluginTypesResponse, error) {
	out := new(GetAvailablePluginTypesResponse)
	err := c.cc.Invoke(ctx, "/teleport.plugins.v1.PluginService/GetAvailablePluginTypes", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// PluginServiceServer is the server API for PluginService service.
// All implementations must embed UnimplementedPluginServiceServer
// for forward compatibility
type PluginServiceServer interface {
	// CreatePlugin creates a new plugin instance.
	CreatePlugin(context.Context, *CreatePluginRequest) (*emptypb.Empty, error)
	// GetPlugin returns a plugin instance by name.
	GetPlugin(context.Context, *GetPluginRequest) (*types.PluginV1, error)
	// DeletePlugin removes the specified plugin instance.
	DeletePlugin(context.Context, *DeletePluginRequest) (*emptypb.Empty, error)
	// ListPlugins returns a paginated view of plugin instances.
	ListPlugins(context.Context, *ListPluginsRequest) (*ListPluginsResponse, error)
	// SetPluginCredentials sets the credentials for the given plugin.
	SetPluginCredentials(context.Context, *SetPluginCredentialsRequest) (*emptypb.Empty, error)
	// SetPluginCredentials sets the status for the given plugin.
	SetPluginStatus(context.Context, *SetPluginStatusRequest) (*emptypb.Empty, error)
	// GetAvailablePluginTypes returns the types of plugins
	// that the auth server supports onboarding.
	GetAvailablePluginTypes(context.Context, *GetAvailablePluginTypesRequest) (*GetAvailablePluginTypesResponse, error)
	mustEmbedUnimplementedPluginServiceServer()
}

// UnimplementedPluginServiceServer must be embedded to have forward compatible implementations.
type UnimplementedPluginServiceServer struct {
}

func (UnimplementedPluginServiceServer) CreatePlugin(context.Context, *CreatePluginRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreatePlugin not implemented")
}
func (UnimplementedPluginServiceServer) GetPlugin(context.Context, *GetPluginRequest) (*types.PluginV1, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetPlugin not implemented")
}
func (UnimplementedPluginServiceServer) DeletePlugin(context.Context, *DeletePluginRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeletePlugin not implemented")
}
func (UnimplementedPluginServiceServer) ListPlugins(context.Context, *ListPluginsRequest) (*ListPluginsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListPlugins not implemented")
}
func (UnimplementedPluginServiceServer) SetPluginCredentials(context.Context, *SetPluginCredentialsRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SetPluginCredentials not implemented")
}
func (UnimplementedPluginServiceServer) SetPluginStatus(context.Context, *SetPluginStatusRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SetPluginStatus not implemented")
}
func (UnimplementedPluginServiceServer) GetAvailablePluginTypes(context.Context, *GetAvailablePluginTypesRequest) (*GetAvailablePluginTypesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetAvailablePluginTypes not implemented")
}
func (UnimplementedPluginServiceServer) mustEmbedUnimplementedPluginServiceServer() {}

// UnsafePluginServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to PluginServiceServer will
// result in compilation errors.
type UnsafePluginServiceServer interface {
	mustEmbedUnimplementedPluginServiceServer()
}

func RegisterPluginServiceServer(s grpc.ServiceRegistrar, srv PluginServiceServer) {
	s.RegisterService(&PluginService_ServiceDesc, srv)
}

func _PluginService_CreatePlugin_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreatePluginRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PluginServiceServer).CreatePlugin(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/teleport.plugins.v1.PluginService/CreatePlugin",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PluginServiceServer).CreatePlugin(ctx, req.(*CreatePluginRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _PluginService_GetPlugin_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetPluginRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PluginServiceServer).GetPlugin(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/teleport.plugins.v1.PluginService/GetPlugin",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PluginServiceServer).GetPlugin(ctx, req.(*GetPluginRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _PluginService_DeletePlugin_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeletePluginRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PluginServiceServer).DeletePlugin(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/teleport.plugins.v1.PluginService/DeletePlugin",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PluginServiceServer).DeletePlugin(ctx, req.(*DeletePluginRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _PluginService_ListPlugins_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListPluginsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PluginServiceServer).ListPlugins(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/teleport.plugins.v1.PluginService/ListPlugins",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PluginServiceServer).ListPlugins(ctx, req.(*ListPluginsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _PluginService_SetPluginCredentials_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SetPluginCredentialsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PluginServiceServer).SetPluginCredentials(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/teleport.plugins.v1.PluginService/SetPluginCredentials",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PluginServiceServer).SetPluginCredentials(ctx, req.(*SetPluginCredentialsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _PluginService_SetPluginStatus_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SetPluginStatusRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PluginServiceServer).SetPluginStatus(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/teleport.plugins.v1.PluginService/SetPluginStatus",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PluginServiceServer).SetPluginStatus(ctx, req.(*SetPluginStatusRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _PluginService_GetAvailablePluginTypes_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetAvailablePluginTypesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PluginServiceServer).GetAvailablePluginTypes(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/teleport.plugins.v1.PluginService/GetAvailablePluginTypes",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PluginServiceServer).GetAvailablePluginTypes(ctx, req.(*GetAvailablePluginTypesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// PluginService_ServiceDesc is the grpc.ServiceDesc for PluginService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var PluginService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "teleport.plugins.v1.PluginService",
	HandlerType: (*PluginServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CreatePlugin",
			Handler:    _PluginService_CreatePlugin_Handler,
		},
		{
			MethodName: "GetPlugin",
			Handler:    _PluginService_GetPlugin_Handler,
		},
		{
			MethodName: "DeletePlugin",
			Handler:    _PluginService_DeletePlugin_Handler,
		},
		{
			MethodName: "ListPlugins",
			Handler:    _PluginService_ListPlugins_Handler,
		},
		{
			MethodName: "SetPluginCredentials",
			Handler:    _PluginService_SetPluginCredentials_Handler,
		},
		{
			MethodName: "SetPluginStatus",
			Handler:    _PluginService_SetPluginStatus_Handler,
		},
		{
			MethodName: "GetAvailablePluginTypes",
			Handler:    _PluginService_GetAvailablePluginTypes_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "teleport/plugins/v1/plugin_service.proto",
}
