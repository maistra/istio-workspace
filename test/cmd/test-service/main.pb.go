// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.15.8
// source: main.proto

package main

import (
	context "context"
	proto "github.com/golang/protobuf/proto"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
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

// An empty request object
type Callee struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *Callee) Reset() {
	*x = Callee{}
	if protoimpl.UnsafeEnabled {
		mi := &file_main_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Callee) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Callee) ProtoMessage() {}

func (x *Callee) ProtoReflect() protoreflect.Message {
	mi := &file_main_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Callee.ProtoReflect.Descriptor instead.
func (*Callee) Descriptor() ([]byte, []int) {
	return file_main_proto_rawDescGZIP(), []int{0}
}

// The response message containing information from the called service.
type CallStack struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Caller    string       `protobuf:"bytes,1,opt,name=caller,proto3" json:"caller,omitempty"`
	Protocol  string       `protobuf:"bytes,2,opt,name=protocol,proto3" json:"protocol,omitempty"`
	Path      string       `protobuf:"bytes,3,opt,name=path,proto3" json:"path,omitempty"`
	Color     string       `protobuf:"bytes,4,opt,name=color,proto3" json:"color,omitempty"`
	StartTime int64        `protobuf:"varint,5,opt,name=startTime,proto3" json:"startTime,omitempty"`
	EndTime   int64        `protobuf:"varint,6,opt,name=endTime,proto3" json:"endTime,omitempty"`
	Called    []*CallStack `protobuf:"bytes,7,rep,name=called,proto3" json:"called,omitempty"`
}

func (x *CallStack) Reset() {
	*x = CallStack{}
	if protoimpl.UnsafeEnabled {
		mi := &file_main_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CallStack) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CallStack) ProtoMessage() {}

func (x *CallStack) ProtoReflect() protoreflect.Message {
	mi := &file_main_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CallStack.ProtoReflect.Descriptor instead.
func (*CallStack) Descriptor() ([]byte, []int) {
	return file_main_proto_rawDescGZIP(), []int{1}
}

func (x *CallStack) GetCaller() string {
	if x != nil {
		return x.Caller
	}
	return ""
}

func (x *CallStack) GetProtocol() string {
	if x != nil {
		return x.Protocol
	}
	return ""
}

func (x *CallStack) GetPath() string {
	if x != nil {
		return x.Path
	}
	return ""
}

func (x *CallStack) GetColor() string {
	if x != nil {
		return x.Color
	}
	return ""
}

func (x *CallStack) GetStartTime() int64 {
	if x != nil {
		return x.StartTime
	}
	return 0
}

func (x *CallStack) GetEndTime() int64 {
	if x != nil {
		return x.EndTime
	}
	return 0
}

func (x *CallStack) GetCalled() []*CallStack {
	if x != nil {
		return x.Called
	}
	return nil
}

var File_main_proto protoreflect.FileDescriptor

var file_main_proto_rawDesc = []byte{
	0x0a, 0x0a, 0x6d, 0x61, 0x69, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x08, 0x0a, 0x06,
	0x43, 0x61, 0x6c, 0x6c, 0x65, 0x65, 0x22, 0xc5, 0x01, 0x0a, 0x09, 0x43, 0x61, 0x6c, 0x6c, 0x53,
	0x74, 0x61, 0x63, 0x6b, 0x12, 0x16, 0x0a, 0x06, 0x63, 0x61, 0x6c, 0x6c, 0x65, 0x72, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x63, 0x61, 0x6c, 0x6c, 0x65, 0x72, 0x12, 0x1a, 0x0a, 0x08,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x12, 0x12, 0x0a, 0x04, 0x70, 0x61, 0x74, 0x68,
	0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x70, 0x61, 0x74, 0x68, 0x12, 0x14, 0x0a, 0x05,
	0x63, 0x6f, 0x6c, 0x6f, 0x72, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x63, 0x6f, 0x6c,
	0x6f, 0x72, 0x12, 0x1c, 0x0a, 0x09, 0x73, 0x74, 0x61, 0x72, 0x74, 0x54, 0x69, 0x6d, 0x65, 0x18,
	0x05, 0x20, 0x01, 0x28, 0x03, 0x52, 0x09, 0x73, 0x74, 0x61, 0x72, 0x74, 0x54, 0x69, 0x6d, 0x65,
	0x12, 0x18, 0x0a, 0x07, 0x65, 0x6e, 0x64, 0x54, 0x69, 0x6d, 0x65, 0x18, 0x06, 0x20, 0x01, 0x28,
	0x03, 0x52, 0x07, 0x65, 0x6e, 0x64, 0x54, 0x69, 0x6d, 0x65, 0x12, 0x22, 0x0a, 0x06, 0x63, 0x61,
	0x6c, 0x6c, 0x65, 0x64, 0x18, 0x07, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x0a, 0x2e, 0x43, 0x61, 0x6c,
	0x6c, 0x53, 0x74, 0x61, 0x63, 0x6b, 0x52, 0x06, 0x63, 0x61, 0x6c, 0x6c, 0x65, 0x64, 0x32, 0x27,
	0x0a, 0x06, 0x43, 0x61, 0x6c, 0x6c, 0x65, 0x72, 0x12, 0x1d, 0x0a, 0x04, 0x43, 0x61, 0x6c, 0x6c,
	0x12, 0x07, 0x2e, 0x43, 0x61, 0x6c, 0x6c, 0x65, 0x65, 0x1a, 0x0a, 0x2e, 0x43, 0x61, 0x6c, 0x6c,
	0x53, 0x74, 0x61, 0x63, 0x6b, 0x22, 0x00, 0x42, 0x3f, 0x5a, 0x3d, 0x67, 0x69, 0x74, 0x68, 0x75,
	0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6d, 0x61, 0x69, 0x73, 0x74, 0x72, 0x61, 0x2f, 0x69, 0x73,
	0x74, 0x69, 0x6f, 0x2d, 0x77, 0x6f, 0x72, 0x6b, 0x73, 0x70, 0x61, 0x63, 0x65, 0x2f, 0x74, 0x65,
	0x73, 0x74, 0x2f, 0x63, 0x6d, 0x64, 0x2f, 0x74, 0x65, 0x73, 0x74, 0x2d, 0x73, 0x65, 0x72, 0x76,
	0x69, 0x63, 0x65, 0x3b, 0x6d, 0x61, 0x69, 0x6e, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_main_proto_rawDescOnce sync.Once
	file_main_proto_rawDescData = file_main_proto_rawDesc
)

func file_main_proto_rawDescGZIP() []byte {
	file_main_proto_rawDescOnce.Do(func() {
		file_main_proto_rawDescData = protoimpl.X.CompressGZIP(file_main_proto_rawDescData)
	})
	return file_main_proto_rawDescData
}

var file_main_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_main_proto_goTypes = []interface{}{
	(*Callee)(nil),    // 0: Callee
	(*CallStack)(nil), // 1: CallStack
}
var file_main_proto_depIdxs = []int32{
	1, // 0: CallStack.called:type_name -> CallStack
	0, // 1: Caller.Call:input_type -> Callee
	1, // 2: Caller.Call:output_type -> CallStack
	2, // [2:3] is the sub-list for method output_type
	1, // [1:2] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_main_proto_init() }
func file_main_proto_init() {
	if File_main_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_main_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Callee); i {
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
		file_main_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CallStack); i {
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
			RawDescriptor: file_main_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_main_proto_goTypes,
		DependencyIndexes: file_main_proto_depIdxs,
		MessageInfos:      file_main_proto_msgTypes,
	}.Build()
	File_main_proto = out.File
	file_main_proto_rawDesc = nil
	file_main_proto_goTypes = nil
	file_main_proto_depIdxs = nil
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConnInterface

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion6

// CallerClient is the client API for Caller service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type CallerClient interface {
	// Call
	Call(ctx context.Context, in *Callee, opts ...grpc.CallOption) (*CallStack, error)
}

type callerClient struct {
	cc grpc.ClientConnInterface
}

func NewCallerClient(cc grpc.ClientConnInterface) CallerClient {
	return &callerClient{cc}
}

func (c *callerClient) Call(ctx context.Context, in *Callee, opts ...grpc.CallOption) (*CallStack, error) {
	out := new(CallStack)
	err := c.cc.Invoke(ctx, "/Caller/Call", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// CallerServer is the server API for Caller service.
type CallerServer interface {
	// Call
	Call(context.Context, *Callee) (*CallStack, error)
}

// UnimplementedCallerServer can be embedded to have forward compatible implementations.
type UnimplementedCallerServer struct {
}

func (*UnimplementedCallerServer) Call(context.Context, *Callee) (*CallStack, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Call not implemented")
}

func RegisterCallerServer(s *grpc.Server, srv CallerServer) {
	s.RegisterService(&_Caller_serviceDesc, srv)
}

func _Caller_Call_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Callee)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CallerServer).Call(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Caller/Call",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CallerServer).Call(ctx, req.(*Callee))
	}
	return interceptor(ctx, in, info, handler)
}

var _Caller_serviceDesc = grpc.ServiceDesc{
	ServiceName: "Caller",
	HandlerType: (*CallerServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Call",
			Handler:    _Caller_Call_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "main.proto",
}
