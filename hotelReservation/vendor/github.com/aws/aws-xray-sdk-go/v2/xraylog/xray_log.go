// Copyright 2017-2017 Amazon.com, Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may not use this file except in compliance with the License. A copy of the License is located at
//
//     http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.

package xraylog

import (
	"fmt"
	"io"
	"sync"
	"time"
)

// Logger is the logging interface used by xray. fmt.Stringer is used to
// defer expensive serialization operations until the message is actually
// logged (i.e. don't bother serializing debug messages if they aren't going
// to show up).
type Logger interface {
	// Log can be called concurrently from multiple goroutines so make sure
	// your implementation is goroutine safe.
	Log(level LogLevel, msg fmt.Stringer)
}

// LogLevel represents the severity of a log message, where a higher value
// means more severe. The integer value should not be serialized as it is
// subject to change.
type LogLevel int

const (
	// LogLevelDebug is usually only enabled when debugging.
	LogLevelDebug LogLevel = iota + 1

	// LogLevelInfo is general operational entries about what's going on inside the application.
	LogLevelInfo

	// LogLevelWarn is non-critical entries that deserve eyes.
	LogLevelWarn

	// LogLevelError is used for errors that should definitely be noted.
	LogLevelError
)

func (ll LogLevel) String() string {
	switch ll {
	case LogLevelDebug:
		return "DEBUG"
	case LogLevelInfo:
		return "INFO"
	case LogLevelWarn:
		return "WARN"
	case LogLevelError:
		return "ERROR"
	default:
		return fmt.Sprintf("UNKNOWNLOGLEVEL<%d>", ll)
	}
}

// NewDefaultLogger makes a Logger object that writes newline separated
// messages to w, if the level of the message is at least minLogLevel.
// The default logger synchronizes around Write() calls to the underlying
// io.Writer.
func NewDefaultLogger(w io.Writer, minLogLevel LogLevel) Logger {
	return &defaultLogger{w: w, minLevel: minLogLevel}
}

type defaultLogger struct {
	mu       sync.Mutex
	w        io.Writer
	minLevel LogLevel
}

func (l *defaultLogger) Log(ll LogLevel, msg fmt.Stringer) {
	if ll < l.minLevel {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	fmt.Fprintf(l.w, "%s [%s] %s\n", time.Now().Format(time.RFC3339), ll, msg)
}

// NullLogger can be used to disable logging (pass to xray.SetLogger()).
var NullLogger = nullLogger{}

type nullLogger struct{}

func (nl nullLogger) Log(ll LogLevel, msg fmt.Stringer) {
}
