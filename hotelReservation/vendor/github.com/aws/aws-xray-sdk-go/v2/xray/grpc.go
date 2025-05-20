package xray

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/aws/aws-xray-sdk-go/v2/internal/logger"

	"google.golang.org/grpc/codes"

	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	"github.com/aws/aws-xray-sdk-go/v2/header"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// UnaryClientInterceptor provides gRPC unary client interceptor.
func UnaryClientInterceptor(clientInterceptorOptions ...GrpcOption) grpc.UnaryClientInterceptor {
	var option grpcOption
	for _, interceptorOption := range clientInterceptorOptions {
		interceptorOption.apply(&option)
	}

	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		var segmentName string
		if option.segmentNamer == nil {
			segmentName = inferServiceName(method)
		} else {
			segmentName = option.segmentNamer.Name(cc.Target())
		}
		if option.config != nil {
			ctx = context.WithValue(ctx, RecorderContextKey{}, option.config)
		}
		return Capture(ctx, segmentName, func(ctx context.Context) error {
			seg := GetSegment(ctx)
			if seg == nil {
				return errors.New("failed to record gRPC transaction: segment cannot be found")
			}

			ctx = metadata.AppendToOutgoingContext(ctx, TraceIDHeaderKey, seg.DownstreamHeader().String())

			seg.Lock()
			seg.Namespace = "remote"
			seg.GetHTTP().GetRequest().URL = "grpc://" + cc.Target() + method
			seg.GetHTTP().GetRequest().Method = http.MethodPost
			seg.Unlock()

			err := invoker(ctx, method, req, reply, cc, opts...)

			recordContentLength(seg, reply)
			if err != nil {
				classifyErrorStatus(seg, err)
			}

			return err
		})
	}
}

// UnaryServerInterceptor provides gRPC unary server interceptor.
func UnaryServerInterceptor(serverInterceptorOptions ...GrpcOption) grpc.UnaryServerInterceptor {
	var option grpcOption
	for _, options := range serverInterceptorOptions {
		options.apply(&option)
	}

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		md, ok := metadata.FromIncomingContext(ctx)

		var traceID string
		if ok && len(md.Get(TraceIDHeaderKey)) == 1 {
			traceID = md.Get(TraceIDHeaderKey)[0]
		}
		traceHeader := header.FromString(traceID)

		var host string

		if len(md.Get(":authority")) == 1 {
			host = md.Get(":authority")[0]
		}
		requestURL := url.URL{
			Scheme: "grpc",
			Host:   host,
			Path:   info.FullMethod,
		}

		var name string
		if option.segmentNamer == nil {
			name = inferServiceName(info.FullMethod)
		} else {
			name = option.segmentNamer.Name(host)
		}

		if option.config != nil {
			ctx = context.WithValue(ctx, RecorderContextKey{}, option.config)
		}

		var seg *Segment
		ctx, seg = NewSegmentFromHeader(ctx, name, &http.Request{
			Host:   host,
			URL:    &requestURL,
			Method: http.MethodPost,
		}, traceHeader)
		defer seg.Close(nil)

		seg.Lock()
		seg.GetHTTP().GetRequest().ClientIP, seg.GetHTTP().GetRequest().XForwardedFor = clientIPFromGrpcMetadata(md)
		seg.GetHTTP().GetRequest().URL = requestURL.String()
		seg.GetHTTP().GetRequest().Method = http.MethodPost
		if len(md.Get("user-agent")) == 1 {
			seg.GetHTTP().GetRequest().UserAgent = md.Get("user-agent")[0]
		}
		seg.Unlock()

		resp, err = handler(ctx, req)
		if err != nil {
			classifyErrorStatus(seg, err)
		}
		recordContentLength(seg, resp)
		if headerErr := addResponseTraceHeader(ctx, seg, traceHeader); headerErr != nil {
			logger.Debug("fail to set the grpc trace header")
		}

		return resp, err
	}
}

func classifyErrorStatus(seg *Segment, err error) {
	seg.Lock()
	defer seg.Unlock()
	grpcStatus, ok := status.FromError(err)
	if !ok {
		seg.Fault = true
		return
	}
	switch grpcStatus.Code() {
	case codes.Canceled, codes.InvalidArgument, codes.NotFound, codes.AlreadyExists, codes.PermissionDenied, codes.Unauthenticated, codes.FailedPrecondition, codes.Aborted, codes.OutOfRange:
		seg.Error = true
	case codes.Unknown, codes.DeadlineExceeded, codes.Unimplemented, codes.Internal, codes.Unavailable, codes.DataLoss:
		seg.Fault = true
	case codes.ResourceExhausted:
		seg.Throttle = true
	}
}

func clientIPFromGrpcMetadata(md metadata.MD) (string, bool) {
	if len(md.Get("x-forwarded-for")) != 1 {
		return "", false
	}
	forwardedFor := md.Get("x-forwarded-for")[0]
	if forwardedFor != "" {
		return strings.TrimSpace(strings.Split(forwardedFor, ",")[0]), true
	}
	return "", false
}

func recordContentLength(seg *Segment, reply interface{}) {
	seg.Lock()
	defer seg.Unlock()
	if protoMessage, isProtoMessage := reply.(proto.Message); isProtoMessage {
		seg.GetHTTP().GetResponse().ContentLength = proto.Size(protoMessage)
	}
}

func addResponseTraceHeader(ctx context.Context, seg *Segment, incomingTraceHeader *header.Header) error {
	var respHeader bytes.Buffer
	respHeader.WriteString("Root=")
	respHeader.WriteString(seg.TraceID)
	if incomingTraceHeader.SamplingDecision == header.Requested {
		respHeader.WriteString(";Sampled=")
		respHeader.WriteString(strconv.Itoa(btoi(seg.Sampled)))
	}

	headers := metadata.New(map[string]string{
		TraceIDHeaderKey: respHeader.String(),
	})
	return grpc.SetHeader(ctx, headers)
}

func inferServiceName(fullMethodName string) string {
	fullMethodName = fullMethodName[1:]
	return fullMethodName[:strings.Index(fullMethodName, "/")]
}

type GrpcOption interface {
	apply(option *grpcOption)
}

type grpcOption struct {
	config       *Config
	segmentNamer SegmentNamer
}

func newFuncGrpcOption(f func(option *grpcOption)) GrpcOption {
	return funcGrpcOption{f: f}
}

type funcGrpcOption struct {
	f func(option *grpcOption)
}

func (f funcGrpcOption) apply(option *grpcOption) {
	f.f(option)
}

// WithRecorder configures the instrumentation by given xray.Config.
func WithRecorder(cfg *Config) GrpcOption {
	return newFuncGrpcOption(func(option *grpcOption) {
		option.config = cfg
	})
}

// WithSegmentNamer makes the interceptor use the segment namer to name the segment.
func WithSegmentNamer(sn SegmentNamer) GrpcOption {
	return newFuncGrpcOption(func(option *grpcOption) {
		option.segmentNamer = sn
	})
}
