package tracing

import (
	"time"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go/config"
)

// Init returns a newly configured tracer
func Init(serviceName, host string) (opentracing.Tracer, error) {
	cfg := config.Configuration{
		Sampler: &config.SamplerConfig{
			Type:  "probabilistic",
			Param: 0.3,
		},
		Reporter: &config.ReporterConfig{
			LogSpans:            false,
			BufferFlushInterval: 1 * time.Second,
			LocalAgentHostPort:  host,
		},
	}

	tracer, _, err := cfg.New(serviceName)
	if err != nil {
		return nil, err
	}
	return tracer, nil
}
