// Copyright 2017-2017 Amazon.com, Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may not use this file except in compliance with the License. A copy of the License is located at
//
//     http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.

package xray

import (
	"context"
	"net"
	"os"
	"sync"

	"github.com/aws/aws-xray-sdk-go/v2/daemoncfg"
	"github.com/aws/aws-xray-sdk-go/v2/internal/logger"
	"github.com/aws/aws-xray-sdk-go/v2/xraylog"

	"github.com/aws/aws-xray-sdk-go/v2/strategy/ctxmissing"
	"github.com/aws/aws-xray-sdk-go/v2/strategy/exception"
	"github.com/aws/aws-xray-sdk-go/v2/strategy/sampling"
)

// SDKVersion records the current X-Ray Go SDK version.
const SDKVersion = "2.0.0"

// SDKType records which X-Ray SDK customer uses.
const SDKType = "X-Ray for Go"

// SDK provides the shape for unmarshalling an SDK struct.
type SDK struct {
	Version  string `json:"sdk_version,omitempty"`
	Type     string `json:"sdk,omitempty"`
	RuleName string `json:"sampling_rule_name,omitempty"`
}

// SetLogger sets the logger instance used by xray.
// Only set from init() functions as SetLogger is not goroutine safe.
func SetLogger(l xraylog.Logger) {
	logger.Logger = l
}

var globalCfg = newGlobalConfig()

func newGlobalConfig() *globalConfig {
	ret := &globalConfig{}

	daemonEndpoint, err := daemoncfg.GetDaemonEndpointsFromEnv()
	if err != nil {
		panic(err)
	}
	if daemonEndpoint == nil {
		daemonEndpoint = daemoncfg.GetDefaultDaemonEndpoints()
	}
	ret.daemonAddr = daemonEndpoint.UDPAddr

	ss, err := sampling.NewCentralizedStrategy()
	if err != nil {
		panic(err)
	}
	ret.samplingStrategy = ss

	efs, err := exception.NewDefaultFormattingStrategy()
	if err != nil {
		panic(err)
	}
	ret.exceptionFormattingStrategy = efs

	sts, err := NewDefaultStreamingStrategy()
	if err != nil {
		panic(err)
	}
	ret.streamingStrategy = sts

	emt, err := NewDefaultEmitter(ret.daemonAddr)
	if err != nil {
		panic(err)
	}
	ret.emitter = emt

	cms := os.Getenv("AWS_XRAY_CONTEXT_MISSING")
	if cms != "" {
		if cms == ctxmissing.RuntimeErrorStrategy {
			cm := ctxmissing.NewDefaultRuntimeErrorStrategy()
			ret.contextMissingStrategy = cm
		} else if cms == ctxmissing.LogErrorStrategy {
			cm := ctxmissing.NewDefaultLogErrorStrategy()
			ret.contextMissingStrategy = cm
		} else if cms == ctxmissing.IgnoreErrorStrategy {
			cm := ctxmissing.NewDefaultIgnoreErrorStrategy()
			ret.contextMissingStrategy = cm
		}
	} else {
		cm := ctxmissing.NewDefaultLogErrorStrategy()
		ret.contextMissingStrategy = cm
	}

	return ret
}

type globalConfig struct {
	sync.RWMutex

	daemonAddr                  *net.UDPAddr
	emitter                     Emitter
	serviceVersion              string
	samplingStrategy            sampling.Strategy
	streamingStrategy           StreamingStrategy
	exceptionFormattingStrategy exception.FormattingStrategy
	contextMissingStrategy      ctxmissing.Strategy
}

// Config is a set of X-Ray configurations.
type Config struct {
	DaemonAddr                  string
	ServiceVersion              string
	Emitter                     Emitter
	SamplingStrategy            sampling.Strategy
	StreamingStrategy           StreamingStrategy
	ExceptionFormattingStrategy exception.FormattingStrategy
	ContextMissingStrategy      ctxmissing.Strategy

	// LogLevel and LogFormat are deprecated and no longer have any effect.
	// See SetLogger() and the associated xraylog.Logger interface to control
	// logging.
	LogLevel  string
	LogFormat string
}

// ContextWithConfig returns context with given configuration settings.
func ContextWithConfig(ctx context.Context, c Config) (context.Context, error) {
	var errors exception.MultiError

	daemonEndpoints, er := daemoncfg.GetDaemonEndpointsFromString(c.DaemonAddr)

	if daemonEndpoints != nil {
		if c.Emitter != nil {
			c.Emitter.RefreshEmitterWithAddress(daemonEndpoints.UDPAddr)
		}
		if c.SamplingStrategy != nil {
			configureStrategy(c.SamplingStrategy, daemonEndpoints)
		}
	} else if er != nil {
		errors = append(errors, er)
	}

	cms := os.Getenv("AWS_XRAY_CONTEXT_MISSING")
	if cms != "" {
		if cms == ctxmissing.RuntimeErrorStrategy {
			cm := ctxmissing.NewDefaultRuntimeErrorStrategy()
			c.ContextMissingStrategy = cm
		} else if cms == ctxmissing.LogErrorStrategy {
			cm := ctxmissing.NewDefaultLogErrorStrategy()
			c.ContextMissingStrategy = cm
		} else if cms == ctxmissing.IgnoreErrorStrategy {
			cm := ctxmissing.NewDefaultIgnoreErrorStrategy()
			c.ContextMissingStrategy = cm
		}
	}

	var err error
	switch len(errors) {
	case 0:
		err = nil
	case 1:
		err = errors[0]
	default:
		err = errors
	}

	return context.WithValue(ctx, RecorderContextKey{}, &c), err
}

func configureStrategy(s sampling.Strategy, daemonEndpoints *daemoncfg.DaemonEndpoints) {
	if s == nil {
		return
	}
	strategy, ok := s.(*sampling.CentralizedStrategy)
	if ok {
		strategy.LoadDaemonEndpoints(daemonEndpoints)
	}
}

// Configure overrides default configuration options with customer-defined values.
func Configure(c Config) error {
	globalCfg.Lock()
	defer globalCfg.Unlock()

	var errors exception.MultiError

	if c.SamplingStrategy != nil {
		globalCfg.samplingStrategy = c.SamplingStrategy
	}

	if c.Emitter != nil {
		globalCfg.emitter = c.Emitter
	}

	daemonEndpoints, er := daemoncfg.GetDaemonEndpointsFromString(c.DaemonAddr)
	if daemonEndpoints != nil {
		globalCfg.daemonAddr = daemonEndpoints.UDPAddr
		globalCfg.emitter.RefreshEmitterWithAddress(globalCfg.daemonAddr)
		configureStrategy(globalCfg.samplingStrategy, daemonEndpoints)
	} else if er != nil {
		errors = append(errors, er)
	}

	if c.ExceptionFormattingStrategy != nil {
		globalCfg.exceptionFormattingStrategy = c.ExceptionFormattingStrategy
	}

	if c.StreamingStrategy != nil {
		globalCfg.streamingStrategy = c.StreamingStrategy
	}

	cms := os.Getenv("AWS_XRAY_CONTEXT_MISSING")
	if cms != "" {
		if cms == ctxmissing.RuntimeErrorStrategy {
			cm := ctxmissing.NewDefaultRuntimeErrorStrategy()
			globalCfg.contextMissingStrategy = cm
		} else if cms == ctxmissing.LogErrorStrategy {
			cm := ctxmissing.NewDefaultLogErrorStrategy()
			globalCfg.contextMissingStrategy = cm
		} else if cms == ctxmissing.IgnoreErrorStrategy {
			cm := ctxmissing.NewDefaultIgnoreErrorStrategy()
			globalCfg.contextMissingStrategy = cm
		}
	} else if c.ContextMissingStrategy != nil {
		globalCfg.contextMissingStrategy = c.ContextMissingStrategy
	}

	if c.ServiceVersion != "" {
		globalCfg.serviceVersion = c.ServiceVersion
	}

	switch len(errors) {
	case 0:
		return nil
	case 1:
		return errors[0]
	default:
		return errors
	}
}

func (c *globalConfig) DaemonAddr() *net.UDPAddr {
	c.RLock()
	defer c.RUnlock()
	return c.daemonAddr
}

func (c *globalConfig) SamplingStrategy() sampling.Strategy {
	c.RLock()
	defer c.RUnlock()
	return c.samplingStrategy
}

func (c *globalConfig) StreamingStrategy() StreamingStrategy {
	c.RLock()
	defer c.RUnlock()
	return c.streamingStrategy
}

func (c *globalConfig) ExceptionFormattingStrategy() exception.FormattingStrategy {
	c.RLock()
	defer c.RUnlock()
	return c.exceptionFormattingStrategy
}

func (c *globalConfig) ContextMissingStrategy() ctxmissing.Strategy {
	c.RLock()
	defer c.RUnlock()
	return c.contextMissingStrategy
}

func (c *globalConfig) ServiceVersion() string {
	c.RLock()
	defer c.RUnlock()
	return c.serviceVersion
}
