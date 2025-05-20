// Copyright 2017-2017 Amazon.com, Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may not use this file except in compliance with the License. A copy of the License is located at
//
//     http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.

package exception

import (
	"bytes"
	"crypto/rand"
	goerrors "errors"
	"fmt"
	"runtime"
	"strings"

	"github.com/pkg/errors"
)

// Copied from:
// https://github.com/aws/aws-sdk-go/blob/v1.55.6/aws/awserr/error.go#L31-L43
type Error interface {
	// Satisfy the generic error interface.
	error

	// Returns the short phrase depicting the classification of the error.
	Code() string

	// Returns the error details message.
	Message() string

	// Returns the original error if one was set.  Nil is returned if not set.
	OrigErr() error
}

// Copied from:
// https://github.com/aws/aws-sdk-go/blob/v1.55.6/aws/awserr/error.go#L129-L139
type RequestFailure interface {
	Error

	// The status code of the HTTP response.
	StatusCode() int

	// The request ID returned by the service for a request failure. This will
	// be empty if no request ID is available such as the request failed due
	// to a connection error.
	RequestID() string
}

// StackTracer is an interface for implementing StackTrace method.
type StackTracer interface {
	StackTrace() []uintptr
}

// Exception provides the shape for unmarshalling an exception.
type Exception struct {
	ID      string  `json:"id,omitempty"`
	Type    string  `json:"type,omitempty"`
	Message string  `json:"message,omitempty"`
	Stack   []Stack `json:"stack,omitempty"`
	Remote  bool    `json:"remote,omitempty"`
}

// Stack provides the shape for unmarshalling an stack.
type Stack struct {
	Path  string `json:"path,omitempty"`
	Line  int    `json:"line,omitempty"`
	Label string `json:"label,omitempty"`
}

// MultiError is a type for a slice of error.
type MultiError []error

// Error returns a string format of concatenating multiple errors.
func (e MultiError) Error() string {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "%d errors occurred:\n", len(e))
	for _, err := range e {
		buf.WriteString("* ")
		buf.WriteString(err.Error())
		buf.WriteByte('\n')
	}
	return buf.String()
}

var defaultErrorFrameCount = 32

// DefaultFormattingStrategy is the default implementation of
// the ExceptionFormattingStrategy and has a configurable frame count.
type DefaultFormattingStrategy struct {
	FrameCount int
}

// NewDefaultFormattingStrategy initializes DefaultFormattingStrategy
// with default value of frame count.
func NewDefaultFormattingStrategy() (*DefaultFormattingStrategy, error) {
	return &DefaultFormattingStrategy{FrameCount: defaultErrorFrameCount}, nil
}

// NewDefaultFormattingStrategyWithDefinedErrorFrameCount initializes
// DefaultFormattingStrategy with customer defined frame count.
func NewDefaultFormattingStrategyWithDefinedErrorFrameCount(frameCount int) (*DefaultFormattingStrategy, error) {
	if frameCount > 32 || frameCount < 0 {
		return nil, errors.New("frameCount must be a non-negative integer and less than 32")
	}
	return &DefaultFormattingStrategy{FrameCount: frameCount}, nil
}

// Error returns the value of XRayError by given error message.
func (dEFS *DefaultFormattingStrategy) Error(message string) *XRayError {
	s := make([]uintptr, dEFS.FrameCount)
	n := runtime.Callers(2, s)
	s = s[:n]

	return &XRayError{
		Type:    "error",
		Message: message,
		Stack:   s,
	}
}

// Errorf formats according to a format specifier and returns value of XRayError.
func (dEFS *DefaultFormattingStrategy) Errorf(formatString string, args ...interface{}) *XRayError {
	e := dEFS.Error(fmt.Sprintf(formatString, args...))
	e.Stack = e.Stack[1:]
	return e
}

// Panic records error type as panic in segment and returns value of XRayError.
func (dEFS *DefaultFormattingStrategy) Panic(message string) *XRayError {
	e := dEFS.Error(message)
	e.Type = "panic"
	e.Stack = filterPanicStack(e.Stack)
	return e
}

// Panicf formats according to a format specifier and returns value of XRayError.
func (dEFS *DefaultFormattingStrategy) Panicf(formatString string, args ...interface{}) *XRayError {
	e := dEFS.Panic(fmt.Sprintf(formatString, args...))
	return e
}

// ExceptionFromError takes an error and returns value of Exception
func (dEFS *DefaultFormattingStrategy) ExceptionFromError(err error) Exception {
	var isRemote bool
	var reqErr RequestFailure
	if goerrors.As(err, &reqErr) {
		// A service error occurs
		if reqErr.RequestID() != "" {
			isRemote = true
		}
	}

	// Fetches type from err
	t := fmt.Sprintf("%T", err)
	// normalize the type
	t = strings.Replace(t, "*", "", -1)
	e := Exception{
		ID:      newExceptionID(),
		Type:    t,
		Message: err.Error(),
		Remote:  isRemote,
	}

	xRayErr := &XRayError{}
	if goerrors.As(err, &xRayErr) {
		e.Type = xRayErr.Type
	}

	var s []uintptr

	// This is our publicly supported interface for passing along stack traces
	var st StackTracer
	if goerrors.As(err, &st) {
		s = st.StackTrace()
	}

	// We also accept github.com/pkg/errors style stack traces for ease of use
	var est interface {
		StackTrace() errors.StackTrace
	}
	if goerrors.As(err, &est) {
		for _, frame := range est.StackTrace() {
			s = append(s, uintptr(frame))
		}
	}

	if s == nil {
		s = make([]uintptr, dEFS.FrameCount)
		n := runtime.Callers(5, s)
		s = s[:n]
	}

	e.Stack = convertStack(s)
	return e
}

func newExceptionID() string {
	var r [8]byte
	_, err := rand.Read(r[:])
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("%02x", r)
}

func filterPanicStack(stack []uintptr) []uintptr {
	// filter out frames through the first runtime/panic.go frame
	frames := runtime.CallersFrames(stack)

	loc := 0
	index := 0
	d := true
	for frame, more := frames.Next(); d; frame, more = frames.Next() {
		loc++
		path, _, label := parseFrame(frame)
		if label == "gopanic" && path == "runtime/panic.go" {
			index = loc
			break
		}
		d = more
	}

	return stack[index:]
}

func convertStack(s []uintptr) []Stack {
	var r []Stack
	frames := runtime.CallersFrames(s)

	d := true
	for frame, more := frames.Next(); d; frame, more = frames.Next() {
		f := &Stack{}
		f.Path, f.Line, f.Label = parseFrame(frame)
		r = append(r, *f)
		d = more
	}
	return r
}

func parseFrame(frame runtime.Frame) (string, int, string) {
	path, line, label := frame.File, frame.Line, frame.Function

	// Strip GOPATH from path by counting the number of seperators in label & path
	// For example:
	//   GOPATH = /home/user
	//   path   = /home/user/src/pkg/sub/file.go
	//   label  = pkg/sub.Type.Method
	// We want to set path to:
	//    pkg/sub/file.go
	i := len(path)
	for n, g := 0, strings.Count(label, "/")+2; n < g; n++ {
		i = strings.LastIndex(path[:i], "/")
		if i == -1 {
			// Something went wrong and path has less seperators than we expected
			// Abort and leave i as -1 to counteract the +1 below
			break
		}
	}
	path = path[i+1:] // Trim the initial /

	// Strip the path from the function name as it's already in the path
	label = label[strings.LastIndex(label, "/")+1:]
	// Likewise strip the package name
	label = label[strings.Index(label, ".")+1:]

	return path, line, label
}
