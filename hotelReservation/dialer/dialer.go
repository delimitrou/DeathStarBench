package dialer

import (
	"fmt"
	"time"

	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	"github.com/harlow/go-micro-services/tls"
	opentracing "github.com/opentracing/opentracing-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/resolver"

	"github.com/simplesurance/grpcconsulresolver/consul"
)

func init() {
	resolver.Register(consul.NewBuilder())
}

// DialOption allows optional config for dialer
type DialOption func(name string) (grpc.DialOption, error)

// WithTracer traces rpc calls
func WithTracer(tracer opentracing.Tracer) DialOption {
	return func(name string) (grpc.DialOption, error) {
		return grpc.WithUnaryInterceptor(otgrpc.OpenTracingClientInterceptor(tracer)), nil
	}
}

// WithBalancer enables client side load balancing
func WithBalancer() DialOption {
	return func(name string) (grpc.DialOption, error) {
		return grpc.WithDefaultServiceConfig(`{"loadBalancingConfig": [{"round_robin":{}}]}`), nil
	}
}

// Dial returns a load balanced grpc client conn with tracing interceptor
func Dial(name string, opts ...DialOption) (*grpc.ClientConn, error) {

	dialopts := []grpc.DialOption{
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Timeout:             120 * time.Second,
			PermitWithoutStream: true,
		}),
	}
	if tlsopt := tls.GetDialOpt(); tlsopt != nil {
		dialopts = append(dialopts, tlsopt)
	} else {
		dialopts = append(dialopts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	for _, fn := range opts {
		opt, err := fn(name)
		if err != nil {
			return nil, fmt.Errorf("config error: %v", err)
		}
		dialopts = append(dialopts, opt)
	}

	conn, err := grpc.Dial(name, dialopts...)
	if err != nil {
		return nil, fmt.Errorf("failed to dial %s: %v", name, err)
	}

	return conn, nil
}
