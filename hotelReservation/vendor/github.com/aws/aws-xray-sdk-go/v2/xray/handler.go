// Copyright 2017-2017 Amazon.com, Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may not use this file except in compliance with the License. A copy of the License is located at
//
//     http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.

package xray

import (
	"bytes"
	"context"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/aws/aws-xray-sdk-go/v2/header"
	"github.com/aws/aws-xray-sdk-go/v2/pattern"
)

// SegmentNamer is the interface for naming service node.
type SegmentNamer interface {
	Name(host string) string
}

// FixedSegmentNamer records the fixed name of service node.
type FixedSegmentNamer struct {
	FixedName string
}

// NewFixedSegmentNamer initializes a FixedSegmentNamer which
// will provide a fixed segment name for every generated segment.
// If the AWS_XRAY_TRACING_NAME environment variable is set,
// its value will override the provided name argument.
func NewFixedSegmentNamer(name string) *FixedSegmentNamer {
	return &FixedSegmentNamer{
		FixedName: name,
	}
}

// Name returns the segment name for the given host header value.
// In this case, FixedName is always returned.
func (fSN *FixedSegmentNamer) Name(host string) string {
	return fSN.FixedName
}

// DynamicSegmentNamer chooses names for segments generated
// for incoming requests by parsing the HOST header of the
// incoming request. If the host header matches a given
// recognized pattern (using the included pattern package),
// it is used as the segment name. Otherwise, the fallback
// name is used.
type DynamicSegmentNamer struct {
	FallbackName    string
	RecognizedHosts string
}

// NewDynamicSegmentNamer creates a new dynamic segment namer.
func NewDynamicSegmentNamer(fallback string, recognized string) *DynamicSegmentNamer {
	return &DynamicSegmentNamer{
		FallbackName:    fallback,
		RecognizedHosts: recognized,
	}
}

// Name returns the segment name for the given host header value.
func (dSN *DynamicSegmentNamer) Name(host string) string {
	if pattern.WildcardMatchCaseInsensitive(dSN.RecognizedHosts, host) {
		return host
	}
	return dSN.FallbackName
}

// HandlerWithContext wraps the provided http handler and context to parse
// the incoming headers, add response headers if needed, and sets HTTP
// specific trace fields. HandlerWithContext names the generated segments
// using the provided SegmentNamer.
func HandlerWithContext(ctx context.Context, sn SegmentNamer, h http.Handler) http.Handler {
	cfg := GetRecorder(ctx)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := sn.Name(r.Host)

		traceHeader := header.FromString(r.Header.Get(TraceIDHeaderKey))
		ctx := context.WithValue(r.Context(), RecorderContextKey{}, cfg)
		c, seg := NewSegmentFromHeader(ctx, name, r, traceHeader)
		defer seg.Close(nil)
		r = r.WithContext(c)

		HttpTrace(seg, h, w, r, traceHeader)
	})
}

// Handler wraps the provided http handler with xray.Capture
// using the request's context, parsing the incoming headers,
// adding response headers if needed, and sets HTTP specific trace fields.
// Handler names the generated segments using the provided SegmentNamer.
func Handler(sn SegmentNamer, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := sn.Name(r.Host)

		traceHeader := header.FromString(r.Header.Get(TraceIDHeaderKey))
		ctx, seg := NewSegmentFromHeader(r.Context(), name, r, traceHeader)
		defer seg.Close(nil)
		r = r.WithContext(ctx)

		HttpTrace(seg, h, w, r, traceHeader)
	})
}

func HttpTrace(seg *Segment, h http.Handler, w http.ResponseWriter, r *http.Request, traceHeader *header.Header) {
	httpCaptureRequest(seg, r)
	traceIDHeaderValue := generateTraceIDHeaderValue(seg, traceHeader)
	w.Header().Set(TraceIDHeaderKey, traceIDHeaderValue)

	capturer := &responseCapturer{w, 200, 0}
	resp := capturer.wrappedResponseWriter()
	h.ServeHTTP(resp, r)

	seg.Lock()
	seg.GetHTTP().GetResponse().ContentLength, _ = strconv.Atoi(capturer.Header().Get("Content-Length"))
	seg.Unlock()
	HttpCaptureResponse(seg, capturer.status)
}

func clientIP(r *http.Request) (string, bool) {
	forwardedFor := r.Header.Get("X-Forwarded-For")
	if forwardedFor != "" {
		return strings.TrimSpace(strings.Split(forwardedFor, ",")[0]), true
	}
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr, false
	}
	return ip, false
}

func btoi(b bool) int {
	if b {
		return 1
	}
	return 0
}

// generateTraceIDHeaderValue generates value for _x_amzn_trace_id header key
func generateTraceIDHeaderValue(seg *Segment, traceHeader *header.Header) string {
	seg.Lock()
	defer seg.Unlock()

	var respHeader bytes.Buffer
	respHeader.WriteString("Root=")
	respHeader.WriteString(seg.TraceID)

	if traceHeader.ParentID != "" {
		respHeader.WriteString(";Parent=")
		respHeader.WriteString(traceHeader.ParentID)
	}

	if traceHeader.SamplingDecision == header.NotSampled {
		respHeader.WriteString(";Sampled=0")
	} else if traceHeader.SamplingDecision == header.Sampled {
		respHeader.WriteString(";Sampled=1")
	} else if traceHeader.SamplingDecision == header.Requested {
		respHeader.WriteString(";Sampled=")
		respHeader.WriteString(strconv.Itoa(btoi(seg.Sampled)))
	}

	return respHeader.String()
}

// HttpCaptureResponse fill response by http status code
func HttpCaptureResponse(seg *Segment, statusCode int) {
	seg.Lock()
	defer seg.Unlock()

	seg.GetHTTP().GetResponse().Status = statusCode

	if statusCode >= 400 && statusCode < 500 {
		seg.Error = true
	}
	if statusCode == 429 {
		seg.Throttle = true
	}
	if statusCode >= 500 && statusCode < 600 {
		seg.Fault = true
	}
}

// httpCaptureRequest fill request data by http.Request
func httpCaptureRequest(seg *Segment, r *http.Request) {
	seg.Lock()
	defer seg.Unlock()

	scheme := "https://"
	if r.TLS == nil {
		scheme = "http://"
	}

	seg.GetHTTP().GetRequest().Method = r.Method
	seg.GetHTTP().GetRequest().URL = scheme + r.Host + r.URL.Path
	seg.GetHTTP().GetRequest().ClientIP, seg.GetHTTP().GetRequest().XForwardedFor = clientIP(r)
	seg.GetHTTP().GetRequest().UserAgent = r.UserAgent()
}
