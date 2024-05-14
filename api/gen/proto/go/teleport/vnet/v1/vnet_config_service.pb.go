// Copyright 2024 Gravitational, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.0
// 	protoc        (unknown)
// source: teleport/vnet/v1/vnet_config_service.proto

package vnet

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// Request for GetVnetConfig.
type GetVnetConfigRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *GetVnetConfigRequest) Reset() {
	*x = GetVnetConfigRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_teleport_vnet_v1_vnet_config_service_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetVnetConfigRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetVnetConfigRequest) ProtoMessage() {}

func (x *GetVnetConfigRequest) ProtoReflect() protoreflect.Message {
	mi := &file_teleport_vnet_v1_vnet_config_service_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetVnetConfigRequest.ProtoReflect.Descriptor instead.
func (*GetVnetConfigRequest) Descriptor() ([]byte, []int) {
	return file_teleport_vnet_v1_vnet_config_service_proto_rawDescGZIP(), []int{0}
}

// Request for CreateVnetConfig.
type CreateVnetConfigRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The VnetConfig resource to create.
	VnetConfig *VnetConfig `protobuf:"bytes,1,opt,name=vnet_config,json=vnetConfig,proto3" json:"vnet_config,omitempty"`
}

func (x *CreateVnetConfigRequest) Reset() {
	*x = CreateVnetConfigRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_teleport_vnet_v1_vnet_config_service_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CreateVnetConfigRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CreateVnetConfigRequest) ProtoMessage() {}

func (x *CreateVnetConfigRequest) ProtoReflect() protoreflect.Message {
	mi := &file_teleport_vnet_v1_vnet_config_service_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CreateVnetConfigRequest.ProtoReflect.Descriptor instead.
func (*CreateVnetConfigRequest) Descriptor() ([]byte, []int) {
	return file_teleport_vnet_v1_vnet_config_service_proto_rawDescGZIP(), []int{1}
}

func (x *CreateVnetConfigRequest) GetVnetConfig() *VnetConfig {
	if x != nil {
		return x.VnetConfig
	}
	return nil
}

// Request for UpdateVnetConfig.
type UpdateVnetConfigRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The VnetConfig resource to create.
	VnetConfig *VnetConfig `protobuf:"bytes,1,opt,name=vnet_config,json=vnetConfig,proto3" json:"vnet_config,omitempty"`
}

func (x *UpdateVnetConfigRequest) Reset() {
	*x = UpdateVnetConfigRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_teleport_vnet_v1_vnet_config_service_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UpdateVnetConfigRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UpdateVnetConfigRequest) ProtoMessage() {}

func (x *UpdateVnetConfigRequest) ProtoReflect() protoreflect.Message {
	mi := &file_teleport_vnet_v1_vnet_config_service_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UpdateVnetConfigRequest.ProtoReflect.Descriptor instead.
func (*UpdateVnetConfigRequest) Descriptor() ([]byte, []int) {
	return file_teleport_vnet_v1_vnet_config_service_proto_rawDescGZIP(), []int{2}
}

func (x *UpdateVnetConfigRequest) GetVnetConfig() *VnetConfig {
	if x != nil {
		return x.VnetConfig
	}
	return nil
}

// Request for UpsertVnetConfig.
type UpsertVnetConfigRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The VnetConfig resource to create.
	VnetConfig *VnetConfig `protobuf:"bytes,1,opt,name=vnet_config,json=vnetConfig,proto3" json:"vnet_config,omitempty"`
}

func (x *UpsertVnetConfigRequest) Reset() {
	*x = UpsertVnetConfigRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_teleport_vnet_v1_vnet_config_service_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UpsertVnetConfigRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UpsertVnetConfigRequest) ProtoMessage() {}

func (x *UpsertVnetConfigRequest) ProtoReflect() protoreflect.Message {
	mi := &file_teleport_vnet_v1_vnet_config_service_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UpsertVnetConfigRequest.ProtoReflect.Descriptor instead.
func (*UpsertVnetConfigRequest) Descriptor() ([]byte, []int) {
	return file_teleport_vnet_v1_vnet_config_service_proto_rawDescGZIP(), []int{3}
}

func (x *UpsertVnetConfigRequest) GetVnetConfig() *VnetConfig {
	if x != nil {
		return x.VnetConfig
	}
	return nil
}

// Request for DeleteVnetConfig.
type DeleteVnetConfigRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *DeleteVnetConfigRequest) Reset() {
	*x = DeleteVnetConfigRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_teleport_vnet_v1_vnet_config_service_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DeleteVnetConfigRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DeleteVnetConfigRequest) ProtoMessage() {}

func (x *DeleteVnetConfigRequest) ProtoReflect() protoreflect.Message {
	mi := &file_teleport_vnet_v1_vnet_config_service_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DeleteVnetConfigRequest.ProtoReflect.Descriptor instead.
func (*DeleteVnetConfigRequest) Descriptor() ([]byte, []int) {
	return file_teleport_vnet_v1_vnet_config_service_proto_rawDescGZIP(), []int{4}
}

var File_teleport_vnet_v1_vnet_config_service_proto protoreflect.FileDescriptor

var file_teleport_vnet_v1_vnet_config_service_proto_rawDesc = []byte{
	0x0a, 0x2a, 0x74, 0x65, 0x6c, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x2f, 0x76, 0x6e, 0x65, 0x74, 0x2f,
	0x76, 0x31, 0x2f, 0x76, 0x6e, 0x65, 0x74, 0x5f, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x5f, 0x73,
	0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x10, 0x74, 0x65,
	0x6c, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x2e, 0x76, 0x6e, 0x65, 0x74, 0x2e, 0x76, 0x31, 0x1a, 0x1b,
	0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f,
	0x65, 0x6d, 0x70, 0x74, 0x79, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x22, 0x74, 0x65, 0x6c,
	0x65, 0x70, 0x6f, 0x72, 0x74, 0x2f, 0x76, 0x6e, 0x65, 0x74, 0x2f, 0x76, 0x31, 0x2f, 0x76, 0x6e,
	0x65, 0x74, 0x5f, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22,
	0x16, 0x0a, 0x14, 0x47, 0x65, 0x74, 0x56, 0x6e, 0x65, 0x74, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x22, 0x58, 0x0a, 0x17, 0x43, 0x72, 0x65, 0x61, 0x74,
	0x65, 0x56, 0x6e, 0x65, 0x74, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x12, 0x3d, 0x0a, 0x0b, 0x76, 0x6e, 0x65, 0x74, 0x5f, 0x63, 0x6f, 0x6e, 0x66, 0x69,
	0x67, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1c, 0x2e, 0x74, 0x65, 0x6c, 0x65, 0x70, 0x6f,
	0x72, 0x74, 0x2e, 0x76, 0x6e, 0x65, 0x74, 0x2e, 0x76, 0x31, 0x2e, 0x56, 0x6e, 0x65, 0x74, 0x43,
	0x6f, 0x6e, 0x66, 0x69, 0x67, 0x52, 0x0a, 0x76, 0x6e, 0x65, 0x74, 0x43, 0x6f, 0x6e, 0x66, 0x69,
	0x67, 0x22, 0x58, 0x0a, 0x17, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x56, 0x6e, 0x65, 0x74, 0x43,
	0x6f, 0x6e, 0x66, 0x69, 0x67, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x3d, 0x0a, 0x0b,
	0x76, 0x6e, 0x65, 0x74, 0x5f, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x1c, 0x2e, 0x74, 0x65, 0x6c, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x2e, 0x76, 0x6e, 0x65,
	0x74, 0x2e, 0x76, 0x31, 0x2e, 0x56, 0x6e, 0x65, 0x74, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x52,
	0x0a, 0x76, 0x6e, 0x65, 0x74, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x22, 0x58, 0x0a, 0x17, 0x55,
	0x70, 0x73, 0x65, 0x72, 0x74, 0x56, 0x6e, 0x65, 0x74, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x3d, 0x0a, 0x0b, 0x76, 0x6e, 0x65, 0x74, 0x5f, 0x63,
	0x6f, 0x6e, 0x66, 0x69, 0x67, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1c, 0x2e, 0x74, 0x65,
	0x6c, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x2e, 0x76, 0x6e, 0x65, 0x74, 0x2e, 0x76, 0x31, 0x2e, 0x56,
	0x6e, 0x65, 0x74, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x52, 0x0a, 0x76, 0x6e, 0x65, 0x74, 0x43,
	0x6f, 0x6e, 0x66, 0x69, 0x67, 0x22, 0x19, 0x0a, 0x17, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x56,
	0x6e, 0x65, 0x74, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x32, 0xd8, 0x03, 0x0a, 0x11, 0x56, 0x6e, 0x65, 0x74, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x53,
	0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x55, 0x0a, 0x0d, 0x47, 0x65, 0x74, 0x56, 0x6e, 0x65,
	0x74, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x12, 0x26, 0x2e, 0x74, 0x65, 0x6c, 0x65, 0x70, 0x6f,
	0x72, 0x74, 0x2e, 0x76, 0x6e, 0x65, 0x74, 0x2e, 0x76, 0x31, 0x2e, 0x47, 0x65, 0x74, 0x56, 0x6e,
	0x65, 0x74, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a,
	0x1c, 0x2e, 0x74, 0x65, 0x6c, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x2e, 0x76, 0x6e, 0x65, 0x74, 0x2e,
	0x76, 0x31, 0x2e, 0x56, 0x6e, 0x65, 0x74, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x12, 0x5b, 0x0a,
	0x10, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x56, 0x6e, 0x65, 0x74, 0x43, 0x6f, 0x6e, 0x66, 0x69,
	0x67, 0x12, 0x29, 0x2e, 0x74, 0x65, 0x6c, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x2e, 0x76, 0x6e, 0x65,
	0x74, 0x2e, 0x76, 0x31, 0x2e, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x56, 0x6e, 0x65, 0x74, 0x43,
	0x6f, 0x6e, 0x66, 0x69, 0x67, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1c, 0x2e, 0x74,
	0x65, 0x6c, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x2e, 0x76, 0x6e, 0x65, 0x74, 0x2e, 0x76, 0x31, 0x2e,
	0x56, 0x6e, 0x65, 0x74, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x12, 0x5b, 0x0a, 0x10, 0x55, 0x70,
	0x64, 0x61, 0x74, 0x65, 0x56, 0x6e, 0x65, 0x74, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x12, 0x29,
	0x2e, 0x74, 0x65, 0x6c, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x2e, 0x76, 0x6e, 0x65, 0x74, 0x2e, 0x76,
	0x31, 0x2e, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x56, 0x6e, 0x65, 0x74, 0x43, 0x6f, 0x6e, 0x66,
	0x69, 0x67, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1c, 0x2e, 0x74, 0x65, 0x6c, 0x65,
	0x70, 0x6f, 0x72, 0x74, 0x2e, 0x76, 0x6e, 0x65, 0x74, 0x2e, 0x76, 0x31, 0x2e, 0x56, 0x6e, 0x65,
	0x74, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x12, 0x5b, 0x0a, 0x10, 0x55, 0x70, 0x73, 0x65, 0x72,
	0x74, 0x56, 0x6e, 0x65, 0x74, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x12, 0x29, 0x2e, 0x74, 0x65,
	0x6c, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x2e, 0x76, 0x6e, 0x65, 0x74, 0x2e, 0x76, 0x31, 0x2e, 0x55,
	0x70, 0x73, 0x65, 0x72, 0x74, 0x56, 0x6e, 0x65, 0x74, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1c, 0x2e, 0x74, 0x65, 0x6c, 0x65, 0x70, 0x6f, 0x72,
	0x74, 0x2e, 0x76, 0x6e, 0x65, 0x74, 0x2e, 0x76, 0x31, 0x2e, 0x56, 0x6e, 0x65, 0x74, 0x43, 0x6f,
	0x6e, 0x66, 0x69, 0x67, 0x12, 0x55, 0x0a, 0x10, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x56, 0x6e,
	0x65, 0x74, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x12, 0x29, 0x2e, 0x74, 0x65, 0x6c, 0x65, 0x70,
	0x6f, 0x72, 0x74, 0x2e, 0x76, 0x6e, 0x65, 0x74, 0x2e, 0x76, 0x31, 0x2e, 0x44, 0x65, 0x6c, 0x65,
	0x74, 0x65, 0x56, 0x6e, 0x65, 0x74, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x1a, 0x16, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x42, 0x4a, 0x5a, 0x48, 0x67,
	0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x67, 0x72, 0x61, 0x76, 0x69, 0x74,
	0x61, 0x74, 0x69, 0x6f, 0x6e, 0x61, 0x6c, 0x2f, 0x74, 0x65, 0x6c, 0x65, 0x70, 0x6f, 0x72, 0x74,
	0x2f, 0x61, 0x70, 0x69, 0x2f, 0x67, 0x65, 0x6e, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x67,
	0x6f, 0x2f, 0x74, 0x65, 0x6c, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x2f, 0x76, 0x6e, 0x65, 0x74, 0x2f,
	0x76, 0x31, 0x3b, 0x76, 0x6e, 0x65, 0x74, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_teleport_vnet_v1_vnet_config_service_proto_rawDescOnce sync.Once
	file_teleport_vnet_v1_vnet_config_service_proto_rawDescData = file_teleport_vnet_v1_vnet_config_service_proto_rawDesc
)

func file_teleport_vnet_v1_vnet_config_service_proto_rawDescGZIP() []byte {
	file_teleport_vnet_v1_vnet_config_service_proto_rawDescOnce.Do(func() {
		file_teleport_vnet_v1_vnet_config_service_proto_rawDescData = protoimpl.X.CompressGZIP(file_teleport_vnet_v1_vnet_config_service_proto_rawDescData)
	})
	return file_teleport_vnet_v1_vnet_config_service_proto_rawDescData
}

var file_teleport_vnet_v1_vnet_config_service_proto_msgTypes = make([]protoimpl.MessageInfo, 5)
var file_teleport_vnet_v1_vnet_config_service_proto_goTypes = []interface{}{
	(*GetVnetConfigRequest)(nil),    // 0: teleport.vnet.v1.GetVnetConfigRequest
	(*CreateVnetConfigRequest)(nil), // 1: teleport.vnet.v1.CreateVnetConfigRequest
	(*UpdateVnetConfigRequest)(nil), // 2: teleport.vnet.v1.UpdateVnetConfigRequest
	(*UpsertVnetConfigRequest)(nil), // 3: teleport.vnet.v1.UpsertVnetConfigRequest
	(*DeleteVnetConfigRequest)(nil), // 4: teleport.vnet.v1.DeleteVnetConfigRequest
	(*VnetConfig)(nil),              // 5: teleport.vnet.v1.VnetConfig
	(*emptypb.Empty)(nil),           // 6: google.protobuf.Empty
}
var file_teleport_vnet_v1_vnet_config_service_proto_depIdxs = []int32{
	5, // 0: teleport.vnet.v1.CreateVnetConfigRequest.vnet_config:type_name -> teleport.vnet.v1.VnetConfig
	5, // 1: teleport.vnet.v1.UpdateVnetConfigRequest.vnet_config:type_name -> teleport.vnet.v1.VnetConfig
	5, // 2: teleport.vnet.v1.UpsertVnetConfigRequest.vnet_config:type_name -> teleport.vnet.v1.VnetConfig
	0, // 3: teleport.vnet.v1.VnetConfigService.GetVnetConfig:input_type -> teleport.vnet.v1.GetVnetConfigRequest
	1, // 4: teleport.vnet.v1.VnetConfigService.CreateVnetConfig:input_type -> teleport.vnet.v1.CreateVnetConfigRequest
	2, // 5: teleport.vnet.v1.VnetConfigService.UpdateVnetConfig:input_type -> teleport.vnet.v1.UpdateVnetConfigRequest
	3, // 6: teleport.vnet.v1.VnetConfigService.UpsertVnetConfig:input_type -> teleport.vnet.v1.UpsertVnetConfigRequest
	4, // 7: teleport.vnet.v1.VnetConfigService.DeleteVnetConfig:input_type -> teleport.vnet.v1.DeleteVnetConfigRequest
	5, // 8: teleport.vnet.v1.VnetConfigService.GetVnetConfig:output_type -> teleport.vnet.v1.VnetConfig
	5, // 9: teleport.vnet.v1.VnetConfigService.CreateVnetConfig:output_type -> teleport.vnet.v1.VnetConfig
	5, // 10: teleport.vnet.v1.VnetConfigService.UpdateVnetConfig:output_type -> teleport.vnet.v1.VnetConfig
	5, // 11: teleport.vnet.v1.VnetConfigService.UpsertVnetConfig:output_type -> teleport.vnet.v1.VnetConfig
	6, // 12: teleport.vnet.v1.VnetConfigService.DeleteVnetConfig:output_type -> google.protobuf.Empty
	8, // [8:13] is the sub-list for method output_type
	3, // [3:8] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_teleport_vnet_v1_vnet_config_service_proto_init() }
func file_teleport_vnet_v1_vnet_config_service_proto_init() {
	if File_teleport_vnet_v1_vnet_config_service_proto != nil {
		return
	}
	file_teleport_vnet_v1_vnet_config_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_teleport_vnet_v1_vnet_config_service_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetVnetConfigRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_teleport_vnet_v1_vnet_config_service_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CreateVnetConfigRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_teleport_vnet_v1_vnet_config_service_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UpdateVnetConfigRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_teleport_vnet_v1_vnet_config_service_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UpsertVnetConfigRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_teleport_vnet_v1_vnet_config_service_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DeleteVnetConfigRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_teleport_vnet_v1_vnet_config_service_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   5,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_teleport_vnet_v1_vnet_config_service_proto_goTypes,
		DependencyIndexes: file_teleport_vnet_v1_vnet_config_service_proto_depIdxs,
		MessageInfos:      file_teleport_vnet_v1_vnet_config_service_proto_msgTypes,
	}.Build()
	File_teleport_vnet_v1_vnet_config_service_proto = out.File
	file_teleport_vnet_v1_vnet_config_service_proto_rawDesc = nil
	file_teleport_vnet_v1_vnet_config_service_proto_goTypes = nil
	file_teleport_vnet_v1_vnet_config_service_proto_depIdxs = nil
}