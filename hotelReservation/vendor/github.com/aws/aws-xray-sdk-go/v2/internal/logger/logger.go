// Copyright 2017-2017 Amazon.com, Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may not use this file except in compliance with the License. A copy of the License is located at
//
//     http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.

package logger

import (
	"fmt"
	"os"

	"github.com/aws/aws-xray-sdk-go/v2/xraylog"
)

// This internal package hides the actual logging functions from the user.

// Logger instance used by xray to log. Set via xray.SetLogger().
var Logger xraylog.Logger = xraylog.NewDefaultLogger(os.Stdout, xraylog.LogLevelInfo)

func Debugf(format string, args ...interface{}) {
	Logger.Log(xraylog.LogLevelDebug, printfArgs{format, args})
}

func Debug(args ...interface{}) {
	Logger.Log(xraylog.LogLevelDebug, printArgs(args))
}

func DebugDeferred(fn func() string) {
	Logger.Log(xraylog.LogLevelDebug, stringerFunc(fn))
}

func Infof(format string, args ...interface{}) {
	Logger.Log(xraylog.LogLevelInfo, printfArgs{format, args})
}

func Info(args ...interface{}) {
	Logger.Log(xraylog.LogLevelInfo, printArgs(args))
}

func Warnf(format string, args ...interface{}) {
	Logger.Log(xraylog.LogLevelWarn, printfArgs{format, args})
}

func Warn(args ...interface{}) {
	Logger.Log(xraylog.LogLevelWarn, printArgs(args))
}

func Errorf(format string, args ...interface{}) {
	Logger.Log(xraylog.LogLevelError, printfArgs{format, args})
}

func Error(args ...interface{}) {
	Logger.Log(xraylog.LogLevelError, printArgs(args))
}

type printfArgs struct {
	format string
	args   []interface{}
}

func (p printfArgs) String() string {
	return fmt.Sprintf(p.format, p.args...)
}

type printArgs []interface{}

func (p printArgs) String() string {
	return fmt.Sprint([]interface{}(p)...)
}

type stringerFunc func() string

func (sf stringerFunc) String() string {
	return sf()
}
