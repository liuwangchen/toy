// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.26.0
// 	protoc        v3.20.1
// source: natsrpc.proto

package natsrpc

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	descriptorpb "google.golang.org/protobuf/types/descriptorpb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// Request 请求
type Request struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Payload []byte            `protobuf:"bytes,1,opt,name=payload,proto3" json:"payload,omitempty"`                                                                                       // 包体
	Header  map[string]string `protobuf:"bytes,2,rep,name=header,proto3" json:"header,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"` // 包头
}

func (x *Request) Reset() {
	*x = Request{}
	if protoimpl.UnsafeEnabled {
		mi := &file_natsrpc_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Request) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Request) ProtoMessage() {}

func (x *Request) ProtoReflect() protoreflect.Message {
	mi := &file_natsrpc_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Request.ProtoReflect.Descriptor instead.
func (*Request) Descriptor() ([]byte, []int) {
	return file_natsrpc_proto_rawDescGZIP(), []int{0}
}

func (x *Request) GetPayload() []byte {
	if x != nil {
		return x.Payload
	}
	return nil
}

func (x *Request) GetHeader() map[string]string {
	if x != nil {
		return x.Header
	}
	return nil
}

// Reply 返回
type Reply struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Payload []byte `protobuf:"bytes,1,opt,name=payload,proto3" json:"payload,omitempty"` // 包体
	Error   string `protobuf:"bytes,2,opt,name=error,proto3" json:"error,omitempty"`     // 错误
}

func (x *Reply) Reset() {
	*x = Reply{}
	if protoimpl.UnsafeEnabled {
		mi := &file_natsrpc_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Reply) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Reply) ProtoMessage() {}

func (x *Reply) ProtoReflect() protoreflect.Message {
	mi := &file_natsrpc_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Reply.ProtoReflect.Descriptor instead.
func (*Reply) Descriptor() ([]byte, []int) {
	return file_natsrpc_proto_rawDescGZIP(), []int{1}
}

func (x *Reply) GetPayload() []byte {
	if x != nil {
		return x.Payload
	}
	return nil
}

func (x *Reply) GetError() string {
	if x != nil {
		return x.Error
	}
	return ""
}

var file_natsrpc_proto_extTypes = []protoimpl.ExtensionInfo{
	{
		ExtendedType:  (*descriptorpb.ServiceOptions)(nil),
		ExtensionType: (*string)(nil),
		Field:         43231,
		Name:          "natsrpc.serviceQueue",
		Tag:           "bytes,43231,opt,name=serviceQueue",
		Filename:      "natsrpc.proto",
	},
	{
		ExtendedType:  (*descriptorpb.ServiceOptions)(nil),
		ExtensionType: (*string)(nil),
		Field:         43232,
		Name:          "natsrpc.topic",
		Tag:           "bytes,43232,opt,name=topic",
		Filename:      "natsrpc.proto",
	},
	{
		ExtendedType:  (*descriptorpb.MethodOptions)(nil),
		ExtensionType: (*string)(nil),
		Field:         2362,
		Name:          "natsrpc.methodQueue",
		Tag:           "bytes,2362,opt,name=methodQueue",
		Filename:      "natsrpc.proto",
	},
	{
		ExtendedType:  (*descriptorpb.MethodOptions)(nil),
		ExtensionType: (*int32)(nil),
		Field:         2363,
		Name:          "natsrpc.reqId",
		Tag:           "varint,2363,opt,name=reqId",
		Filename:      "natsrpc.proto",
	},
	{
		ExtendedType:  (*descriptorpb.MethodOptions)(nil),
		ExtensionType: (*int32)(nil),
		Field:         2364,
		Name:          "natsrpc.respId",
		Tag:           "varint,2364,opt,name=respId",
		Filename:      "natsrpc.proto",
	},
	{
		ExtendedType:  (*descriptorpb.MethodOptions)(nil),
		ExtensionType: (*bool)(nil),
		Field:         2365,
		Name:          "natsrpc.sequence",
		Tag:           "varint,2365,opt,name=sequence",
		Filename:      "natsrpc.proto",
	},
}

// Extension fields to descriptorpb.ServiceOptions.
var (
	// optional string serviceQueue = 43231;
	E_ServiceQueue = &file_natsrpc_proto_extTypes[0] // service级别queue
	// optional string topic = 43232;
	E_Topic = &file_natsrpc_proto_extTypes[1] // topic
)

// Extension fields to descriptorpb.MethodOptions.
var (
	// optional string methodQueue = 2362;
	E_MethodQueue = &file_natsrpc_proto_extTypes[2] // 方法级别的queue
	// optional int32 reqId = 2363;
	E_ReqId = &file_natsrpc_proto_extTypes[3] // reqId
	// optional int32 respId = 2364;
	E_RespId = &file_natsrpc_proto_extTypes[4] // respId
	// optional bool sequence = 2365;
	E_Sequence = &file_natsrpc_proto_extTypes[5] // sequence
)

var File_natsrpc_proto protoreflect.FileDescriptor

var file_natsrpc_proto_rawDesc = []byte{
	0x0a, 0x0d, 0x6e, 0x61, 0x74, 0x73, 0x72, 0x70, 0x63, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12,
	0x07, 0x6e, 0x61, 0x74, 0x73, 0x72, 0x70, 0x63, 0x1a, 0x20, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65,
	0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69,
	0x70, 0x74, 0x6f, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x94, 0x01, 0x0a, 0x07, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x18, 0x0a, 0x07, 0x70, 0x61, 0x79, 0x6c, 0x6f, 0x61,
	0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x07, 0x70, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64,
	0x12, 0x34, 0x0a, 0x06, 0x68, 0x65, 0x61, 0x64, 0x65, 0x72, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b,
	0x32, 0x1c, 0x2e, 0x6e, 0x61, 0x74, 0x73, 0x72, 0x70, 0x63, 0x2e, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x2e, 0x48, 0x65, 0x61, 0x64, 0x65, 0x72, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x06,
	0x68, 0x65, 0x61, 0x64, 0x65, 0x72, 0x1a, 0x39, 0x0a, 0x0b, 0x48, 0x65, 0x61, 0x64, 0x65, 0x72,
	0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38,
	0x01, 0x22, 0x37, 0x0a, 0x05, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x12, 0x18, 0x0a, 0x07, 0x70, 0x61,
	0x79, 0x6c, 0x6f, 0x61, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x07, 0x70, 0x61, 0x79,
	0x6c, 0x6f, 0x61, 0x64, 0x12, 0x14, 0x0a, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x3a, 0x45, 0x0a, 0x0c, 0x73, 0x65,
	0x72, 0x76, 0x69, 0x63, 0x65, 0x51, 0x75, 0x65, 0x75, 0x65, 0x12, 0x1f, 0x2e, 0x67, 0x6f, 0x6f,
	0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x53, 0x65, 0x72,
	0x76, 0x69, 0x63, 0x65, 0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0xdf, 0xd1, 0x02, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x0c, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x51, 0x75, 0x65, 0x75,
	0x65, 0x3a, 0x37, 0x0a, 0x05, 0x74, 0x6f, 0x70, 0x69, 0x63, 0x12, 0x1f, 0x2e, 0x67, 0x6f, 0x6f,
	0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x53, 0x65, 0x72,
	0x76, 0x69, 0x63, 0x65, 0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0xe0, 0xd1, 0x02, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x05, 0x74, 0x6f, 0x70, 0x69, 0x63, 0x3a, 0x41, 0x0a, 0x0b, 0x6d, 0x65,
	0x74, 0x68, 0x6f, 0x64, 0x51, 0x75, 0x65, 0x75, 0x65, 0x12, 0x1e, 0x2e, 0x67, 0x6f, 0x6f, 0x67,
	0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x4d, 0x65, 0x74, 0x68,
	0x6f, 0x64, 0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0xba, 0x12, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x0b, 0x6d, 0x65, 0x74, 0x68, 0x6f, 0x64, 0x51, 0x75, 0x65, 0x75, 0x65, 0x3a, 0x35, 0x0a,
	0x05, 0x72, 0x65, 0x71, 0x49, 0x64, 0x12, 0x1e, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x4d, 0x65, 0x74, 0x68, 0x6f, 0x64, 0x4f,
	0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0xbb, 0x12, 0x20, 0x01, 0x28, 0x05, 0x52, 0x05, 0x72,
	0x65, 0x71, 0x49, 0x64, 0x3a, 0x37, 0x0a, 0x06, 0x72, 0x65, 0x73, 0x70, 0x49, 0x64, 0x12, 0x1e,
	0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66,
	0x2e, 0x4d, 0x65, 0x74, 0x68, 0x6f, 0x64, 0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0xbc,
	0x12, 0x20, 0x01, 0x28, 0x05, 0x52, 0x06, 0x72, 0x65, 0x73, 0x70, 0x49, 0x64, 0x3a, 0x3b, 0x0a,
	0x08, 0x73, 0x65, 0x71, 0x75, 0x65, 0x6e, 0x63, 0x65, 0x12, 0x1e, 0x2e, 0x67, 0x6f, 0x6f, 0x67,
	0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x4d, 0x65, 0x74, 0x68,
	0x6f, 0x64, 0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0xbd, 0x12, 0x20, 0x01, 0x28, 0x08,
	0x52, 0x08, 0x73, 0x65, 0x71, 0x75, 0x65, 0x6e, 0x63, 0x65, 0x42, 0x32, 0x5a, 0x30, 0x67, 0x69,
	0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6c, 0x69, 0x75, 0x77, 0x61, 0x6e, 0x67,
	0x63, 0x68, 0x65, 0x6e, 0x2f, 0x74, 0x6f, 0x79, 0x2f, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x70, 0x6f,
	0x72, 0x74, 0x2f, 0x72, 0x70, 0x63, 0x2f, 0x6e, 0x61, 0x74, 0x73, 0x72, 0x70, 0x63, 0x62, 0x06,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_natsrpc_proto_rawDescOnce sync.Once
	file_natsrpc_proto_rawDescData = file_natsrpc_proto_rawDesc
)

func file_natsrpc_proto_rawDescGZIP() []byte {
	file_natsrpc_proto_rawDescOnce.Do(func() {
		file_natsrpc_proto_rawDescData = protoimpl.X.CompressGZIP(file_natsrpc_proto_rawDescData)
	})
	return file_natsrpc_proto_rawDescData
}

var file_natsrpc_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_natsrpc_proto_goTypes = []interface{}{
	(*Request)(nil),                     // 0: natsrpc.Request
	(*Reply)(nil),                       // 1: natsrpc.Reply
	nil,                                 // 2: natsrpc.Request.HeaderEntry
	(*descriptorpb.ServiceOptions)(nil), // 3: google.protobuf.ServiceOptions
	(*descriptorpb.MethodOptions)(nil),  // 4: google.protobuf.MethodOptions
}
var file_natsrpc_proto_depIdxs = []int32{
	2, // 0: natsrpc.Request.header:type_name -> natsrpc.Request.HeaderEntry
	3, // 1: natsrpc.serviceQueue:extendee -> google.protobuf.ServiceOptions
	3, // 2: natsrpc.topic:extendee -> google.protobuf.ServiceOptions
	4, // 3: natsrpc.methodQueue:extendee -> google.protobuf.MethodOptions
	4, // 4: natsrpc.reqId:extendee -> google.protobuf.MethodOptions
	4, // 5: natsrpc.respId:extendee -> google.protobuf.MethodOptions
	4, // 6: natsrpc.sequence:extendee -> google.protobuf.MethodOptions
	7, // [7:7] is the sub-list for method output_type
	7, // [7:7] is the sub-list for method input_type
	7, // [7:7] is the sub-list for extension type_name
	1, // [1:7] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_natsrpc_proto_init() }
func file_natsrpc_proto_init() {
	if File_natsrpc_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_natsrpc_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Request); i {
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
		file_natsrpc_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Reply); i {
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
			RawDescriptor: file_natsrpc_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 6,
			NumServices:   0,
		},
		GoTypes:           file_natsrpc_proto_goTypes,
		DependencyIndexes: file_natsrpc_proto_depIdxs,
		MessageInfos:      file_natsrpc_proto_msgTypes,
		ExtensionInfos:    file_natsrpc_proto_extTypes,
	}.Build()
	File_natsrpc_proto = out.File
	file_natsrpc_proto_rawDesc = nil
	file_natsrpc_proto_goTypes = nil
	file_natsrpc_proto_depIdxs = nil
}