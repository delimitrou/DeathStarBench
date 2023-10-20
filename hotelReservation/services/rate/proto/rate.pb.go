// Code generated by protoc-gen-go. DO NOT EDIT.
// source: services/rate/proto/rate.proto

/*
Package rate is a generated protocol buffer package.

It is generated from these files:

	services/rate/proto/rate.proto

It has these top-level messages:

	Request
	Result
	RatePlan
	RoomType
*/
package rate

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

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

type Request struct {
	HotelIds []string `protobuf:"bytes,1,rep,name=hotelIds" json:"hotelIds,omitempty"`
	InDate   string   `protobuf:"bytes,2,opt,name=inDate" json:"inDate,omitempty"`
	OutDate  string   `protobuf:"bytes,3,opt,name=outDate" json:"outDate,omitempty"`
}

func (m *Request) Reset()                    { *m = Request{} }
func (m *Request) String() string            { return proto.CompactTextString(m) }
func (*Request) ProtoMessage()               {}
func (*Request) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *Request) GetHotelIds() []string {
	if m != nil {
		return m.HotelIds
	}
	return nil
}

func (m *Request) GetInDate() string {
	if m != nil {
		return m.InDate
	}
	return ""
}

func (m *Request) GetOutDate() string {
	if m != nil {
		return m.OutDate
	}
	return ""
}

type Result struct {
	RatePlans []*RatePlan `protobuf:"bytes,1,rep,name=ratePlans" bson:"ratePlans,omitempty"`
}

func (m *Result) Reset()                    { *m = Result{} }
func (m *Result) String() string            { return proto.CompactTextString(m) }
func (*Result) ProtoMessage()               {}
func (*Result) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *Result) GetRatePlans() []*RatePlan {
	if m != nil {
		return m.RatePlans
	}
	return nil
}

type RatePlan struct {
	HotelId  string    `protobuf:"bytes,1,opt,name=hotelId" bson:"hotelId,omitempty"`
	Code     string    `protobuf:"bytes,2,opt,name=code" bson:"code,omitempty"`
	InDate   string    `protobuf:"bytes,3,opt,name=inDate" bson:"inDate,omitempty"`
	OutDate  string    `protobuf:"bytes,4,opt,name=outDate" bson:"outDate,omitempty"`
	RoomType *RoomType `protobuf:"bytes,5,opt,name=roomType" bson:"roomType,omitempty"`
}

func (m *RatePlan) Reset()                    { *m = RatePlan{} }
func (m *RatePlan) String() string            { return proto.CompactTextString(m) }
func (*RatePlan) ProtoMessage()               {}
func (*RatePlan) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *RatePlan) GetHotelId() string {
	if m != nil {
		return m.HotelId
	}
	return ""
}

func (m *RatePlan) GetCode() string {
	if m != nil {
		return m.Code
	}
	return ""
}

func (m *RatePlan) GetInDate() string {
	if m != nil {
		return m.InDate
	}
	return ""
}

func (m *RatePlan) GetOutDate() string {
	if m != nil {
		return m.OutDate
	}
	return ""
}

func (m *RatePlan) GetRoomType() *RoomType {
	if m != nil {
		return m.RoomType
	}
	return nil
}

type RoomType struct {
	BookableRate       float64 `protobuf:"fixed64,1,opt,name=bookableRate" bson:"bookableRate,omitempty"`
	TotalRate          float64 `protobuf:"fixed64,2,opt,name=totalRate" bson:"totalRate,omitempty"`
	TotalRateInclusive float64 `protobuf:"fixed64,3,opt,name=totalRateInclusive" bson:"totalRateInclusive,omitempty"`
	Code               string  `protobuf:"bytes,4,opt,name=code" bson:"code,omitempty"`
	Currency           string  `protobuf:"bytes,5,opt,name=currency" bson:"currency,omitempty"`
	RoomDescription    string  `protobuf:"bytes,6,opt,name=roomDescription" bson:"roomDescription,omitempty"`
}

func (m *RoomType) Reset()                    { *m = RoomType{} }
func (m *RoomType) String() string            { return proto.CompactTextString(m) }
func (*RoomType) ProtoMessage()               {}
func (*RoomType) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

func (m *RoomType) GetBookableRate() float64 {
	if m != nil {
		return m.BookableRate
	}
	return 0
}

func (m *RoomType) GetTotalRate() float64 {
	if m != nil {
		return m.TotalRate
	}
	return 0
}

func (m *RoomType) GetTotalRateInclusive() float64 {
	if m != nil {
		return m.TotalRateInclusive
	}
	return 0
}

func (m *RoomType) GetCode() string {
	if m != nil {
		return m.Code
	}
	return ""
}

func (m *RoomType) GetCurrency() string {
	if m != nil {
		return m.Currency
	}
	return ""
}

func (m *RoomType) GetRoomDescription() string {
	if m != nil {
		return m.RoomDescription
	}
	return ""
}

func init() {
	proto.RegisterType((*Request)(nil), "rate.Request")
	proto.RegisterType((*Result)(nil), "rate.Result")
	proto.RegisterType((*RatePlan)(nil), "rate.RatePlan")
	proto.RegisterType((*RoomType)(nil), "rate.RoomType")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for Rate service

type RateClient interface {
	// GetRates returns rate codes for hotels for a given date range
	GetRates(ctx context.Context, in *Request, opts ...grpc.CallOption) (*Result, error)
}

type rateClient struct {
	cc *grpc.ClientConn
}

func NewRateClient(cc *grpc.ClientConn) RateClient {
	return &rateClient{cc}
}

func (c *rateClient) GetRates(ctx context.Context, in *Request, opts ...grpc.CallOption) (*Result, error) {
	out := new(Result)
	err := grpc.Invoke(ctx, "/rate.Rate/GetRates", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for Rate service

type RateServer interface {
	// GetRates returns rate codes for hotels for a given date range
	GetRates(context.Context, *Request) (*Result, error)
}

func RegisterRateServer(s *grpc.Server, srv RateServer) {
	s.RegisterService(&_Rate_serviceDesc, srv)
}

func _Rate_GetRates_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Request)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RateServer).GetRates(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rate.Rate/GetRates",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RateServer).GetRates(ctx, req.(*Request))
	}
	return interceptor(ctx, in, info, handler)
}

var _Rate_serviceDesc = grpc.ServiceDesc{
	ServiceName: "rate.Rate",
	HandlerType: (*RateServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetRates",
			Handler:    _Rate_GetRates_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "services/rate/proto/rate.proto",
}

func init() { proto.RegisterFile("services/rate/proto/rate.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 327 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x74, 0x92, 0xdd, 0x4a, 0xc3, 0x30,
	0x14, 0xc7, 0x89, 0xab, 0x5d, 0x7b, 0x9c, 0x0a, 0xe7, 0x42, 0xca, 0x10, 0x19, 0xbd, 0x71, 0x88,
	0x6c, 0x30, 0xc1, 0x27, 0x18, 0xc8, 0xee, 0xe4, 0x20, 0x78, 0xdd, 0x75, 0x07, 0x2c, 0xd6, 0x66,
	0x26, 0xe9, 0x60, 0x2f, 0xe2, 0xa3, 0xf9, 0x3c, 0x92, 0x34, 0xed, 0x3a, 0xd1, 0xbb, 0xff, 0x47,
	0x38, 0xfd, 0x25, 0x3d, 0x70, 0xa3, 0x59, 0xed, 0x8a, 0x9c, 0xf5, 0x5c, 0x65, 0x86, 0xe7, 0x5b,
	0x25, 0x8d, 0x74, 0x72, 0xe6, 0x24, 0x06, 0x56, 0xa7, 0xaf, 0x30, 0x24, 0xfe, 0xac, 0x59, 0x1b,
	0x1c, 0x43, 0xf4, 0x26, 0x0d, 0x97, 0xab, 0x8d, 0x4e, 0xc4, 0x64, 0x30, 0x8d, 0xa9, 0xf3, 0x78,
	0x05, 0x61, 0x51, 0x2d, 0x33, 0xc3, 0xc9, 0xc9, 0x44, 0x4c, 0x63, 0xf2, 0x0e, 0x13, 0x18, 0xca,
	0xda, 0xb8, 0x62, 0xe0, 0x8a, 0xd6, 0xa6, 0x8f, 0x10, 0x12, 0xeb, 0xba, 0x34, 0x78, 0x0f, 0xb1,
	0xfd, 0xd4, 0x73, 0x99, 0x55, 0xcd, 0xe0, 0xb3, 0xc5, 0xc5, 0xcc, 0x81, 0x90, 0x8f, 0xe9, 0x70,
	0x20, 0xfd, 0x12, 0x10, 0xb5, 0xb9, 0x1d, 0xef, 0x11, 0x12, 0xd1, 0x8c, 0xf7, 0x16, 0x11, 0x82,
	0x5c, 0x6e, 0x5a, 0x1c, 0xa7, 0x7b, 0x90, 0x83, 0xff, 0x20, 0x83, 0x23, 0x48, 0xbc, 0x83, 0x48,
	0x49, 0xf9, 0xf1, 0xb2, 0xdf, 0x72, 0x72, 0x3a, 0x11, 0x3d, 0x32, 0x9f, 0x52, 0xd7, 0xa7, 0xdf,
	0x16, 0xcc, 0x1b, 0x4c, 0x61, 0xb4, 0x96, 0xf2, 0x3d, 0x5b, 0x97, 0x6c, 0x61, 0x1d, 0x9d, 0xa0,
	0xa3, 0x0c, 0xaf, 0x21, 0x36, 0xd2, 0x64, 0x25, 0xb5, 0xcf, 0x26, 0xe8, 0x10, 0xe0, 0x0c, 0xb0,
	0x33, 0xab, 0x2a, 0x2f, 0x6b, 0x5d, 0xec, 0x1a, 0x70, 0x41, 0x7f, 0x34, 0xdd, 0x85, 0x83, 0xde,
	0x85, 0xc7, 0x10, 0xe5, 0xb5, 0x52, 0x5c, 0xe5, 0x7b, 0x87, 0x1f, 0x53, 0xe7, 0x71, 0x0a, 0x97,
	0x16, 0x7d, 0xc9, 0x3a, 0x57, 0xc5, 0xd6, 0x14, 0xb2, 0x4a, 0x42, 0x77, 0xe4, 0x77, 0xbc, 0x98,
	0x43, 0xe0, 0x88, 0x6e, 0x21, 0x7a, 0x62, 0x63, 0xa5, 0xc6, 0x73, 0xff, 0x0c, 0xcd, 0x6a, 0x8c,
	0x47, 0xad, 0xb5, 0x3f, 0x74, 0x1d, 0xba, 0x05, 0x7a, 0xf8, 0x09, 0x00, 0x00, 0xff, 0xff, 0x71,
	0x7f, 0xc7, 0x74, 0x62, 0x02, 0x00, 0x00,
}
