// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.15.6
// source: staff.proto

package proto

import (
	proto "github.com/golang/protobuf/proto"
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

// This is a compile-time assertion that a sufficiently up-to-date version
// of the legacy proto package is being used.
const _ = proto.ProtoPackageIsVersion4

type RegisterRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id   string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Addr string `protobuf:"bytes,2,opt,name=addr,proto3" json:"addr,omitempty"`
}

func (x *RegisterRequest) Reset() {
	*x = RegisterRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_staff_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RegisterRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RegisterRequest) ProtoMessage() {}

func (x *RegisterRequest) ProtoReflect() protoreflect.Message {
	mi := &file_staff_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RegisterRequest.ProtoReflect.Descriptor instead.
func (*RegisterRequest) Descriptor() ([]byte, []int) {
	return file_staff_proto_rawDescGZIP(), []int{0}
}

func (x *RegisterRequest) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *RegisterRequest) GetAddr() string {
	if x != nil {
		return x.Addr
	}
	return ""
}

type RegisterReply struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Addr    string `protobuf:"bytes,1,opt,name=addr,proto3" json:"addr,omitempty"`
	Subject string `protobuf:"bytes,2,opt,name=subject,proto3" json:"subject,omitempty"`
}

func (x *RegisterReply) Reset() {
	*x = RegisterReply{}
	if protoimpl.UnsafeEnabled {
		mi := &file_staff_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RegisterReply) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RegisterReply) ProtoMessage() {}

func (x *RegisterReply) ProtoReflect() protoreflect.Message {
	mi := &file_staff_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RegisterReply.ProtoReflect.Descriptor instead.
func (*RegisterReply) Descriptor() ([]byte, []int) {
	return file_staff_proto_rawDescGZIP(), []int{1}
}

func (x *RegisterReply) GetAddr() string {
	if x != nil {
		return x.Addr
	}
	return ""
}

func (x *RegisterReply) GetSubject() string {
	if x != nil {
		return x.Subject
	}
	return ""
}

type ElectRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	StaffId string `protobuf:"bytes,1,opt,name=staff_id,json=staffId,proto3" json:"staff_id,omitempty"`
	TaskId  string `protobuf:"bytes,2,opt,name=task_id,json=taskId,proto3" json:"task_id,omitempty"`
}

func (x *ElectRequest) Reset() {
	*x = ElectRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_staff_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ElectRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ElectRequest) ProtoMessage() {}

func (x *ElectRequest) ProtoReflect() protoreflect.Message {
	mi := &file_staff_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ElectRequest.ProtoReflect.Descriptor instead.
func (*ElectRequest) Descriptor() ([]byte, []int) {
	return file_staff_proto_rawDescGZIP(), []int{2}
}

func (x *ElectRequest) GetStaffId() string {
	if x != nil {
		return x.StaffId
	}
	return ""
}

func (x *ElectRequest) GetTaskId() string {
	if x != nil {
		return x.TaskId
	}
	return ""
}

type ElectReply struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Addr      string   `protobuf:"bytes,1,opt,name=addr,proto3" json:"addr,omitempty"`
	Subject   string   `protobuf:"bytes,2,opt,name=subject,proto3" json:"subject,omitempty"`
	LeaderId  string   `protobuf:"bytes,3,opt,name=leader_id,json=leaderId,proto3" json:"leader_id,omitempty"`
	WorkerIds []string `protobuf:"bytes,4,rep,name=worker_ids,json=workerIds,proto3" json:"worker_ids,omitempty"`
}

func (x *ElectReply) Reset() {
	*x = ElectReply{}
	if protoimpl.UnsafeEnabled {
		mi := &file_staff_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ElectReply) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ElectReply) ProtoMessage() {}

func (x *ElectReply) ProtoReflect() protoreflect.Message {
	mi := &file_staff_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ElectReply.ProtoReflect.Descriptor instead.
func (*ElectReply) Descriptor() ([]byte, []int) {
	return file_staff_proto_rawDescGZIP(), []int{3}
}

func (x *ElectReply) GetAddr() string {
	if x != nil {
		return x.Addr
	}
	return ""
}

func (x *ElectReply) GetSubject() string {
	if x != nil {
		return x.Subject
	}
	return ""
}

func (x *ElectReply) GetLeaderId() string {
	if x != nil {
		return x.LeaderId
	}
	return ""
}

func (x *ElectReply) GetWorkerIds() []string {
	if x != nil {
		return x.WorkerIds
	}
	return nil
}

type ClockinRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	StaffId string `protobuf:"bytes,1,opt,name=staff_id,json=staffId,proto3" json:"staff_id,omitempty"`
	TaskId  string `protobuf:"bytes,2,opt,name=task_id,json=taskId,proto3" json:"task_id,omitempty"`
}

func (x *ClockinRequest) Reset() {
	*x = ClockinRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_staff_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ClockinRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ClockinRequest) ProtoMessage() {}

func (x *ClockinRequest) ProtoReflect() protoreflect.Message {
	mi := &file_staff_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ClockinRequest.ProtoReflect.Descriptor instead.
func (*ClockinRequest) Descriptor() ([]byte, []int) {
	return file_staff_proto_rawDescGZIP(), []int{4}
}

func (x *ClockinRequest) GetStaffId() string {
	if x != nil {
		return x.StaffId
	}
	return ""
}

func (x *ClockinRequest) GetTaskId() string {
	if x != nil {
		return x.TaskId
	}
	return ""
}

type ClockinReply struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *ClockinReply) Reset() {
	*x = ClockinReply{}
	if protoimpl.UnsafeEnabled {
		mi := &file_staff_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ClockinReply) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ClockinReply) ProtoMessage() {}

func (x *ClockinReply) ProtoReflect() protoreflect.Message {
	mi := &file_staff_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ClockinReply.ProtoReflect.Descriptor instead.
func (*ClockinReply) Descriptor() ([]byte, []int) {
	return file_staff_proto_rawDescGZIP(), []int{5}
}

type ClockoutRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	StaffId string `protobuf:"bytes,1,opt,name=staff_id,json=staffId,proto3" json:"staff_id,omitempty"`
	TaskId  string `protobuf:"bytes,2,opt,name=task_id,json=taskId,proto3" json:"task_id,omitempty"`
}

func (x *ClockoutRequest) Reset() {
	*x = ClockoutRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_staff_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ClockoutRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ClockoutRequest) ProtoMessage() {}

func (x *ClockoutRequest) ProtoReflect() protoreflect.Message {
	mi := &file_staff_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ClockoutRequest.ProtoReflect.Descriptor instead.
func (*ClockoutRequest) Descriptor() ([]byte, []int) {
	return file_staff_proto_rawDescGZIP(), []int{6}
}

func (x *ClockoutRequest) GetStaffId() string {
	if x != nil {
		return x.StaffId
	}
	return ""
}

func (x *ClockoutRequest) GetTaskId() string {
	if x != nil {
		return x.TaskId
	}
	return ""
}

type ClockoutReply struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *ClockoutReply) Reset() {
	*x = ClockoutReply{}
	if protoimpl.UnsafeEnabled {
		mi := &file_staff_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ClockoutReply) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ClockoutReply) ProtoMessage() {}

func (x *ClockoutReply) ProtoReflect() protoreflect.Message {
	mi := &file_staff_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ClockoutReply.ProtoReflect.Descriptor instead.
func (*ClockoutReply) Descriptor() ([]byte, []int) {
	return file_staff_proto_rawDescGZIP(), []int{7}
}

type Acknowledgement struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *Acknowledgement) Reset() {
	*x = Acknowledgement{}
	if protoimpl.UnsafeEnabled {
		mi := &file_staff_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Acknowledgement) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Acknowledgement) ProtoMessage() {}

func (x *Acknowledgement) ProtoReflect() protoreflect.Message {
	mi := &file_staff_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Acknowledgement.ProtoReflect.Descriptor instead.
func (*Acknowledgement) Descriptor() ([]byte, []int) {
	return file_staff_proto_rawDescGZIP(), []int{8}
}

var File_staff_proto protoreflect.FileDescriptor

var file_staff_proto_rawDesc = []byte{
	0x0a, 0x0b, 0x73, 0x74, 0x61, 0x66, 0x66, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x05, 0x73,
	0x74, 0x61, 0x66, 0x66, 0x22, 0x35, 0x0a, 0x0f, 0x52, 0x65, 0x67, 0x69, 0x73, 0x74, 0x65, 0x72,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x61, 0x64, 0x64, 0x72, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x61, 0x64, 0x64, 0x72, 0x22, 0x3d, 0x0a, 0x0d, 0x52,
	0x65, 0x67, 0x69, 0x73, 0x74, 0x65, 0x72, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x12, 0x12, 0x0a, 0x04,
	0x61, 0x64, 0x64, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x61, 0x64, 0x64, 0x72,
	0x12, 0x18, 0x0a, 0x07, 0x73, 0x75, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x07, 0x73, 0x75, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x22, 0x42, 0x0a, 0x0c, 0x45, 0x6c,
	0x65, 0x63, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x19, 0x0a, 0x08, 0x73, 0x74,
	0x61, 0x66, 0x66, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x73, 0x74,
	0x61, 0x66, 0x66, 0x49, 0x64, 0x12, 0x17, 0x0a, 0x07, 0x74, 0x61, 0x73, 0x6b, 0x5f, 0x69, 0x64,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x74, 0x61, 0x73, 0x6b, 0x49, 0x64, 0x22, 0x76,
	0x0a, 0x0a, 0x45, 0x6c, 0x65, 0x63, 0x74, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x12, 0x12, 0x0a, 0x04,
	0x61, 0x64, 0x64, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x61, 0x64, 0x64, 0x72,
	0x12, 0x18, 0x0a, 0x07, 0x73, 0x75, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x07, 0x73, 0x75, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x12, 0x1b, 0x0a, 0x09, 0x6c, 0x65,
	0x61, 0x64, 0x65, 0x72, 0x5f, 0x69, 0x64, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x6c,
	0x65, 0x61, 0x64, 0x65, 0x72, 0x49, 0x64, 0x12, 0x1d, 0x0a, 0x0a, 0x77, 0x6f, 0x72, 0x6b, 0x65,
	0x72, 0x5f, 0x69, 0x64, 0x73, 0x18, 0x04, 0x20, 0x03, 0x28, 0x09, 0x52, 0x09, 0x77, 0x6f, 0x72,
	0x6b, 0x65, 0x72, 0x49, 0x64, 0x73, 0x22, 0x44, 0x0a, 0x0e, 0x43, 0x6c, 0x6f, 0x63, 0x6b, 0x69,
	0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x19, 0x0a, 0x08, 0x73, 0x74, 0x61, 0x66,
	0x66, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x73, 0x74, 0x61, 0x66,
	0x66, 0x49, 0x64, 0x12, 0x17, 0x0a, 0x07, 0x74, 0x61, 0x73, 0x6b, 0x5f, 0x69, 0x64, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x74, 0x61, 0x73, 0x6b, 0x49, 0x64, 0x22, 0x0e, 0x0a, 0x0c,
	0x43, 0x6c, 0x6f, 0x63, 0x6b, 0x69, 0x6e, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x22, 0x45, 0x0a, 0x0f,
	0x43, 0x6c, 0x6f, 0x63, 0x6b, 0x6f, 0x75, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12,
	0x19, 0x0a, 0x08, 0x73, 0x74, 0x61, 0x66, 0x66, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x07, 0x73, 0x74, 0x61, 0x66, 0x66, 0x49, 0x64, 0x12, 0x17, 0x0a, 0x07, 0x74, 0x61,
	0x73, 0x6b, 0x5f, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x74, 0x61, 0x73,
	0x6b, 0x49, 0x64, 0x22, 0x0f, 0x0a, 0x0d, 0x43, 0x6c, 0x6f, 0x63, 0x6b, 0x6f, 0x75, 0x74, 0x52,
	0x65, 0x70, 0x6c, 0x79, 0x22, 0x11, 0x0a, 0x0f, 0x41, 0x63, 0x6b, 0x6e, 0x6f, 0x77, 0x6c, 0x65,
	0x64, 0x67, 0x65, 0x6d, 0x65, 0x6e, 0x74, 0x32, 0x76, 0x0a, 0x05, 0x53, 0x74, 0x61, 0x66, 0x66,
	0x12, 0x3a, 0x0a, 0x08, 0x52, 0x65, 0x67, 0x69, 0x73, 0x74, 0x65, 0x72, 0x12, 0x16, 0x2e, 0x73,
	0x74, 0x61, 0x66, 0x66, 0x2e, 0x52, 0x65, 0x67, 0x69, 0x73, 0x74, 0x65, 0x72, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x1a, 0x14, 0x2e, 0x73, 0x74, 0x61, 0x66, 0x66, 0x2e, 0x52, 0x65, 0x67,
	0x69, 0x73, 0x74, 0x65, 0x72, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x22, 0x00, 0x12, 0x31, 0x0a, 0x05,
	0x45, 0x6c, 0x65, 0x63, 0x74, 0x12, 0x13, 0x2e, 0x73, 0x74, 0x61, 0x66, 0x66, 0x2e, 0x45, 0x6c,
	0x65, 0x63, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x11, 0x2e, 0x73, 0x74, 0x61,
	0x66, 0x66, 0x2e, 0x45, 0x6c, 0x65, 0x63, 0x74, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x22, 0x00, 0x42,
	0x1f, 0x5a, 0x1d, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x61, 0x6f,
	0x73, 0x2d, 0x64, 0x65, 0x76, 0x2f, 0x6e, 0x6f, 0x61, 0x68, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_staff_proto_rawDescOnce sync.Once
	file_staff_proto_rawDescData = file_staff_proto_rawDesc
)

func file_staff_proto_rawDescGZIP() []byte {
	file_staff_proto_rawDescOnce.Do(func() {
		file_staff_proto_rawDescData = protoimpl.X.CompressGZIP(file_staff_proto_rawDescData)
	})
	return file_staff_proto_rawDescData
}

var file_staff_proto_msgTypes = make([]protoimpl.MessageInfo, 9)
var file_staff_proto_goTypes = []interface{}{
	(*RegisterRequest)(nil), // 0: staff.RegisterRequest
	(*RegisterReply)(nil),   // 1: staff.RegisterReply
	(*ElectRequest)(nil),    // 2: staff.ElectRequest
	(*ElectReply)(nil),      // 3: staff.ElectReply
	(*ClockinRequest)(nil),  // 4: staff.ClockinRequest
	(*ClockinReply)(nil),    // 5: staff.ClockinReply
	(*ClockoutRequest)(nil), // 6: staff.ClockoutRequest
	(*ClockoutReply)(nil),   // 7: staff.ClockoutReply
	(*Acknowledgement)(nil), // 8: staff.Acknowledgement
}
var file_staff_proto_depIdxs = []int32{
	0, // 0: staff.Staff.Register:input_type -> staff.RegisterRequest
	2, // 1: staff.Staff.Elect:input_type -> staff.ElectRequest
	1, // 2: staff.Staff.Register:output_type -> staff.RegisterReply
	3, // 3: staff.Staff.Elect:output_type -> staff.ElectReply
	2, // [2:4] is the sub-list for method output_type
	0, // [0:2] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_staff_proto_init() }
func file_staff_proto_init() {
	if File_staff_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_staff_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RegisterRequest); i {
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
		file_staff_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RegisterReply); i {
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
		file_staff_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ElectRequest); i {
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
		file_staff_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ElectReply); i {
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
		file_staff_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ClockinRequest); i {
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
		file_staff_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ClockinReply); i {
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
		file_staff_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ClockoutRequest); i {
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
		file_staff_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ClockoutReply); i {
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
		file_staff_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Acknowledgement); i {
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
			RawDescriptor: file_staff_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   9,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_staff_proto_goTypes,
		DependencyIndexes: file_staff_proto_depIdxs,
		MessageInfos:      file_staff_proto_msgTypes,
	}.Build()
	File_staff_proto = out.File
	file_staff_proto_rawDesc = nil
	file_staff_proto_goTypes = nil
	file_staff_proto_depIdxs = nil
}
