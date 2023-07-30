// Copyright (c) 2023 - for information on the respective copyright owner
// see the NOTICE file and/or the repository at
// https://github.com/hyperledger-labs/perun-node
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.30.0
// 	protoc        v4.23.3
// source: watching_service.proto

// Package pb contains proto3 definitions for user API and the corresponding
// generated code for grpc server and client.

package pb

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type StartWatchingLedgerChannelReq struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	SessionID string   `protobuf:"bytes,1,opt,name=sessionID,proto3" json:"sessionID,omitempty"`
	Params    *Params  `protobuf:"bytes,2,opt,name=params,proto3" json:"params,omitempty"`
	State     *State   `protobuf:"bytes,3,opt,name=state,proto3" json:"state,omitempty"`
	Sigs      [][]byte `protobuf:"bytes,4,rep,name=sigs,proto3" json:"sigs,omitempty"`
}

func (x *StartWatchingLedgerChannelReq) Reset() {
	*x = StartWatchingLedgerChannelReq{}
	if protoimpl.UnsafeEnabled {
		mi := &file_watching_service_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *StartWatchingLedgerChannelReq) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StartWatchingLedgerChannelReq) ProtoMessage() {}

func (x *StartWatchingLedgerChannelReq) ProtoReflect() protoreflect.Message {
	mi := &file_watching_service_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StartWatchingLedgerChannelReq.ProtoReflect.Descriptor instead.
func (*StartWatchingLedgerChannelReq) Descriptor() ([]byte, []int) {
	return file_watching_service_proto_rawDescGZIP(), []int{0}
}

func (x *StartWatchingLedgerChannelReq) GetSessionID() string {
	if x != nil {
		return x.SessionID
	}
	return ""
}

func (x *StartWatchingLedgerChannelReq) GetParams() *Params {
	if x != nil {
		return x.Params
	}
	return nil
}

func (x *StartWatchingLedgerChannelReq) GetState() *State {
	if x != nil {
		return x.State
	}
	return nil
}

func (x *StartWatchingLedgerChannelReq) GetSigs() [][]byte {
	if x != nil {
		return x.Sigs
	}
	return nil
}

type StartWatchingLedgerChannelResp struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Types that are assignable to Response:
	//
	//	*StartWatchingLedgerChannelResp_RegisteredEvent
	//	*StartWatchingLedgerChannelResp_ProgressedEvent
	//	*StartWatchingLedgerChannelResp_ConcludedEvent
	//	*StartWatchingLedgerChannelResp_Error
	Response isStartWatchingLedgerChannelResp_Response `protobuf_oneof:"response"`
}

func (x *StartWatchingLedgerChannelResp) Reset() {
	*x = StartWatchingLedgerChannelResp{}
	if protoimpl.UnsafeEnabled {
		mi := &file_watching_service_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *StartWatchingLedgerChannelResp) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StartWatchingLedgerChannelResp) ProtoMessage() {}

func (x *StartWatchingLedgerChannelResp) ProtoReflect() protoreflect.Message {
	mi := &file_watching_service_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StartWatchingLedgerChannelResp.ProtoReflect.Descriptor instead.
func (*StartWatchingLedgerChannelResp) Descriptor() ([]byte, []int) {
	return file_watching_service_proto_rawDescGZIP(), []int{1}
}

func (m *StartWatchingLedgerChannelResp) GetResponse() isStartWatchingLedgerChannelResp_Response {
	if m != nil {
		return m.Response
	}
	return nil
}

func (x *StartWatchingLedgerChannelResp) GetRegisteredEvent() *RegisteredEvent {
	if x, ok := x.GetResponse().(*StartWatchingLedgerChannelResp_RegisteredEvent); ok {
		return x.RegisteredEvent
	}
	return nil
}

func (x *StartWatchingLedgerChannelResp) GetProgressedEvent() *ProgressedEvent {
	if x, ok := x.GetResponse().(*StartWatchingLedgerChannelResp_ProgressedEvent); ok {
		return x.ProgressedEvent
	}
	return nil
}

func (x *StartWatchingLedgerChannelResp) GetConcludedEvent() *ConcludedEvent {
	if x, ok := x.GetResponse().(*StartWatchingLedgerChannelResp_ConcludedEvent); ok {
		return x.ConcludedEvent
	}
	return nil
}

func (x *StartWatchingLedgerChannelResp) GetError() *MsgError {
	if x, ok := x.GetResponse().(*StartWatchingLedgerChannelResp_Error); ok {
		return x.Error
	}
	return nil
}

type isStartWatchingLedgerChannelResp_Response interface {
	isStartWatchingLedgerChannelResp_Response()
}

type StartWatchingLedgerChannelResp_RegisteredEvent struct {
	RegisteredEvent *RegisteredEvent `protobuf:"bytes,1,opt,name=registeredEvent,proto3,oneof"`
}

type StartWatchingLedgerChannelResp_ProgressedEvent struct {
	ProgressedEvent *ProgressedEvent `protobuf:"bytes,2,opt,name=progressedEvent,proto3,oneof"`
}

type StartWatchingLedgerChannelResp_ConcludedEvent struct {
	ConcludedEvent *ConcludedEvent `protobuf:"bytes,3,opt,name=concludedEvent,proto3,oneof"`
}

type StartWatchingLedgerChannelResp_Error struct {
	Error *MsgError `protobuf:"bytes,4,opt,name=error,proto3,oneof"`
}

func (*StartWatchingLedgerChannelResp_RegisteredEvent) isStartWatchingLedgerChannelResp_Response() {}

func (*StartWatchingLedgerChannelResp_ProgressedEvent) isStartWatchingLedgerChannelResp_Response() {}

func (*StartWatchingLedgerChannelResp_ConcludedEvent) isStartWatchingLedgerChannelResp_Response() {}

func (*StartWatchingLedgerChannelResp_Error) isStartWatchingLedgerChannelResp_Response() {}

type StopWatchingReq struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	SessionID string `protobuf:"bytes,1,opt,name=sessionID,proto3" json:"sessionID,omitempty"`
	ChID      []byte `protobuf:"bytes,2,opt,name=chID,proto3" json:"chID,omitempty"`
}

func (x *StopWatchingReq) Reset() {
	*x = StopWatchingReq{}
	if protoimpl.UnsafeEnabled {
		mi := &file_watching_service_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *StopWatchingReq) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StopWatchingReq) ProtoMessage() {}

func (x *StopWatchingReq) ProtoReflect() protoreflect.Message {
	mi := &file_watching_service_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StopWatchingReq.ProtoReflect.Descriptor instead.
func (*StopWatchingReq) Descriptor() ([]byte, []int) {
	return file_watching_service_proto_rawDescGZIP(), []int{2}
}

func (x *StopWatchingReq) GetSessionID() string {
	if x != nil {
		return x.SessionID
	}
	return ""
}

func (x *StopWatchingReq) GetChID() []byte {
	if x != nil {
		return x.ChID
	}
	return nil
}

type StopWatchingResp struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Error *MsgError `protobuf:"bytes,1,opt,name=error,proto3" json:"error,omitempty"`
}

func (x *StopWatchingResp) Reset() {
	*x = StopWatchingResp{}
	if protoimpl.UnsafeEnabled {
		mi := &file_watching_service_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *StopWatchingResp) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StopWatchingResp) ProtoMessage() {}

func (x *StopWatchingResp) ProtoReflect() protoreflect.Message {
	mi := &file_watching_service_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StopWatchingResp.ProtoReflect.Descriptor instead.
func (*StopWatchingResp) Descriptor() ([]byte, []int) {
	return file_watching_service_proto_rawDescGZIP(), []int{3}
}

func (x *StopWatchingResp) GetError() *MsgError {
	if x != nil {
		return x.Error
	}
	return nil
}

var File_watching_service_proto protoreflect.FileDescriptor

var file_watching_service_proto_rawDesc = []byte{
	0x0a, 0x16, 0x77, 0x61, 0x74, 0x63, 0x68, 0x69, 0x6e, 0x67, 0x5f, 0x73, 0x65, 0x72, 0x76, 0x69,
	0x63, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x02, 0x70, 0x62, 0x1a, 0x0c, 0x65, 0x72,
	0x72, 0x6f, 0x72, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x0e, 0x73, 0x64, 0x6b, 0x74,
	0x79, 0x70, 0x65, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x96, 0x01, 0x0a, 0x1d, 0x53,
	0x74, 0x61, 0x72, 0x74, 0x57, 0x61, 0x74, 0x63, 0x68, 0x69, 0x6e, 0x67, 0x4c, 0x65, 0x64, 0x67,
	0x65, 0x72, 0x43, 0x68, 0x61, 0x6e, 0x6e, 0x65, 0x6c, 0x52, 0x65, 0x71, 0x12, 0x1c, 0x0a, 0x09,
	0x73, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x49, 0x44, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x09, 0x73, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x49, 0x44, 0x12, 0x22, 0x0a, 0x06, 0x70, 0x61,
	0x72, 0x61, 0x6d, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0a, 0x2e, 0x70, 0x62, 0x2e,
	0x50, 0x61, 0x72, 0x61, 0x6d, 0x73, 0x52, 0x06, 0x70, 0x61, 0x72, 0x61, 0x6d, 0x73, 0x12, 0x1f,
	0x0a, 0x05, 0x73, 0x74, 0x61, 0x74, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x09, 0x2e,
	0x70, 0x62, 0x2e, 0x53, 0x74, 0x61, 0x74, 0x65, 0x52, 0x05, 0x73, 0x74, 0x61, 0x74, 0x65, 0x12,
	0x12, 0x0a, 0x04, 0x73, 0x69, 0x67, 0x73, 0x18, 0x04, 0x20, 0x03, 0x28, 0x0c, 0x52, 0x04, 0x73,
	0x69, 0x67, 0x73, 0x22, 0x92, 0x02, 0x0a, 0x1e, 0x53, 0x74, 0x61, 0x72, 0x74, 0x57, 0x61, 0x74,
	0x63, 0x68, 0x69, 0x6e, 0x67, 0x4c, 0x65, 0x64, 0x67, 0x65, 0x72, 0x43, 0x68, 0x61, 0x6e, 0x6e,
	0x65, 0x6c, 0x52, 0x65, 0x73, 0x70, 0x12, 0x3f, 0x0a, 0x0f, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74,
	0x65, 0x72, 0x65, 0x64, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x13, 0x2e, 0x70, 0x62, 0x2e, 0x52, 0x65, 0x67, 0x69, 0x73, 0x74, 0x65, 0x72, 0x65, 0x64, 0x45,
	0x76, 0x65, 0x6e, 0x74, 0x48, 0x00, 0x52, 0x0f, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x65, 0x72,
	0x65, 0x64, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x12, 0x3f, 0x0a, 0x0f, 0x70, 0x72, 0x6f, 0x67, 0x72,
	0x65, 0x73, 0x73, 0x65, 0x64, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x13, 0x2e, 0x70, 0x62, 0x2e, 0x50, 0x72, 0x6f, 0x67, 0x72, 0x65, 0x73, 0x73, 0x65, 0x64,
	0x45, 0x76, 0x65, 0x6e, 0x74, 0x48, 0x00, 0x52, 0x0f, 0x70, 0x72, 0x6f, 0x67, 0x72, 0x65, 0x73,
	0x73, 0x65, 0x64, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x12, 0x3c, 0x0a, 0x0e, 0x63, 0x6f, 0x6e, 0x63,
	0x6c, 0x75, 0x64, 0x65, 0x64, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x12, 0x2e, 0x70, 0x62, 0x2e, 0x43, 0x6f, 0x6e, 0x63, 0x6c, 0x75, 0x64, 0x65, 0x64, 0x45,
	0x76, 0x65, 0x6e, 0x74, 0x48, 0x00, 0x52, 0x0e, 0x63, 0x6f, 0x6e, 0x63, 0x6c, 0x75, 0x64, 0x65,
	0x64, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x12, 0x24, 0x0a, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x18,
	0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0c, 0x2e, 0x70, 0x62, 0x2e, 0x4d, 0x73, 0x67, 0x45, 0x72,
	0x72, 0x6f, 0x72, 0x48, 0x00, 0x52, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x42, 0x0a, 0x0a, 0x08,
	0x72, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x43, 0x0a, 0x0f, 0x53, 0x74, 0x6f, 0x70,
	0x57, 0x61, 0x74, 0x63, 0x68, 0x69, 0x6e, 0x67, 0x52, 0x65, 0x71, 0x12, 0x1c, 0x0a, 0x09, 0x73,
	0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x49, 0x44, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09,
	0x73, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x49, 0x44, 0x12, 0x12, 0x0a, 0x04, 0x63, 0x68, 0x49,
	0x44, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x04, 0x63, 0x68, 0x49, 0x44, 0x22, 0x36, 0x0a,
	0x10, 0x53, 0x74, 0x6f, 0x70, 0x57, 0x61, 0x74, 0x63, 0x68, 0x69, 0x6e, 0x67, 0x52, 0x65, 0x73,
	0x70, 0x12, 0x22, 0x0a, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x0c, 0x2e, 0x70, 0x62, 0x2e, 0x4d, 0x73, 0x67, 0x45, 0x72, 0x72, 0x6f, 0x72, 0x52, 0x05,
	0x65, 0x72, 0x72, 0x6f, 0x72, 0x32, 0xb6, 0x01, 0x0a, 0x0c, 0x57, 0x61, 0x74, 0x63, 0x68, 0x69,
	0x6e, 0x67, 0x5f, 0x41, 0x50, 0x49, 0x12, 0x69, 0x0a, 0x1a, 0x53, 0x74, 0x61, 0x72, 0x74, 0x57,
	0x61, 0x74, 0x63, 0x68, 0x69, 0x6e, 0x67, 0x4c, 0x65, 0x64, 0x67, 0x65, 0x72, 0x43, 0x68, 0x61,
	0x6e, 0x6e, 0x65, 0x6c, 0x12, 0x21, 0x2e, 0x70, 0x62, 0x2e, 0x53, 0x74, 0x61, 0x72, 0x74, 0x57,
	0x61, 0x74, 0x63, 0x68, 0x69, 0x6e, 0x67, 0x4c, 0x65, 0x64, 0x67, 0x65, 0x72, 0x43, 0x68, 0x61,
	0x6e, 0x6e, 0x65, 0x6c, 0x52, 0x65, 0x71, 0x1a, 0x22, 0x2e, 0x70, 0x62, 0x2e, 0x53, 0x74, 0x61,
	0x72, 0x74, 0x57, 0x61, 0x74, 0x63, 0x68, 0x69, 0x6e, 0x67, 0x4c, 0x65, 0x64, 0x67, 0x65, 0x72,
	0x43, 0x68, 0x61, 0x6e, 0x6e, 0x65, 0x6c, 0x52, 0x65, 0x73, 0x70, 0x22, 0x00, 0x28, 0x01, 0x30,
	0x01, 0x12, 0x3b, 0x0a, 0x0c, 0x53, 0x74, 0x6f, 0x70, 0x57, 0x61, 0x74, 0x63, 0x68, 0x69, 0x6e,
	0x67, 0x12, 0x13, 0x2e, 0x70, 0x62, 0x2e, 0x53, 0x74, 0x6f, 0x70, 0x57, 0x61, 0x74, 0x63, 0x68,
	0x69, 0x6e, 0x67, 0x52, 0x65, 0x71, 0x1a, 0x14, 0x2e, 0x70, 0x62, 0x2e, 0x53, 0x74, 0x6f, 0x70,
	0x57, 0x61, 0x74, 0x63, 0x68, 0x69, 0x6e, 0x67, 0x52, 0x65, 0x73, 0x70, 0x22, 0x00, 0x42, 0x06,
	0x5a, 0x04, 0x2e, 0x3b, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_watching_service_proto_rawDescOnce sync.Once
	file_watching_service_proto_rawDescData = file_watching_service_proto_rawDesc
)

func file_watching_service_proto_rawDescGZIP() []byte {
	file_watching_service_proto_rawDescOnce.Do(func() {
		file_watching_service_proto_rawDescData = protoimpl.X.CompressGZIP(file_watching_service_proto_rawDescData)
	})
	return file_watching_service_proto_rawDescData
}

var file_watching_service_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_watching_service_proto_goTypes = []interface{}{
	(*StartWatchingLedgerChannelReq)(nil),  // 0: pb.StartWatchingLedgerChannelReq
	(*StartWatchingLedgerChannelResp)(nil), // 1: pb.StartWatchingLedgerChannelResp
	(*StopWatchingReq)(nil),                // 2: pb.StopWatchingReq
	(*StopWatchingResp)(nil),               // 3: pb.StopWatchingResp
	(*Params)(nil),                         // 4: pb.Params
	(*State)(nil),                          // 5: pb.State
	(*RegisteredEvent)(nil),                // 6: pb.RegisteredEvent
	(*ProgressedEvent)(nil),                // 7: pb.ProgressedEvent
	(*ConcludedEvent)(nil),                 // 8: pb.ConcludedEvent
	(*MsgError)(nil),                       // 9: pb.MsgError
}
var file_watching_service_proto_depIdxs = []int32{
	4, // 0: pb.StartWatchingLedgerChannelReq.params:type_name -> pb.Params
	5, // 1: pb.StartWatchingLedgerChannelReq.state:type_name -> pb.State
	6, // 2: pb.StartWatchingLedgerChannelResp.registeredEvent:type_name -> pb.RegisteredEvent
	7, // 3: pb.StartWatchingLedgerChannelResp.progressedEvent:type_name -> pb.ProgressedEvent
	8, // 4: pb.StartWatchingLedgerChannelResp.concludedEvent:type_name -> pb.ConcludedEvent
	9, // 5: pb.StartWatchingLedgerChannelResp.error:type_name -> pb.MsgError
	9, // 6: pb.StopWatchingResp.error:type_name -> pb.MsgError
	0, // 7: pb.Watching_API.StartWatchingLedgerChannel:input_type -> pb.StartWatchingLedgerChannelReq
	2, // 8: pb.Watching_API.StopWatching:input_type -> pb.StopWatchingReq
	1, // 9: pb.Watching_API.StartWatchingLedgerChannel:output_type -> pb.StartWatchingLedgerChannelResp
	3, // 10: pb.Watching_API.StopWatching:output_type -> pb.StopWatchingResp
	9, // [9:11] is the sub-list for method output_type
	7, // [7:9] is the sub-list for method input_type
	7, // [7:7] is the sub-list for extension type_name
	7, // [7:7] is the sub-list for extension extendee
	0, // [0:7] is the sub-list for field type_name
}

func init() { file_watching_service_proto_init() }
func file_watching_service_proto_init() {
	if File_watching_service_proto != nil {
		return
	}
	file_errors_proto_init()
	file_sdktypes_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_watching_service_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*StartWatchingLedgerChannelReq); i {
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
		file_watching_service_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*StartWatchingLedgerChannelResp); i {
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
		file_watching_service_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*StopWatchingReq); i {
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
		file_watching_service_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*StopWatchingResp); i {
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
	file_watching_service_proto_msgTypes[1].OneofWrappers = []interface{}{
		(*StartWatchingLedgerChannelResp_RegisteredEvent)(nil),
		(*StartWatchingLedgerChannelResp_ProgressedEvent)(nil),
		(*StartWatchingLedgerChannelResp_ConcludedEvent)(nil),
		(*StartWatchingLedgerChannelResp_Error)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_watching_service_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_watching_service_proto_goTypes,
		DependencyIndexes: file_watching_service_proto_depIdxs,
		MessageInfos:      file_watching_service_proto_msgTypes,
	}.Build()
	File_watching_service_proto = out.File
	file_watching_service_proto_rawDesc = nil
	file_watching_service_proto_goTypes = nil
	file_watching_service_proto_depIdxs = nil
}