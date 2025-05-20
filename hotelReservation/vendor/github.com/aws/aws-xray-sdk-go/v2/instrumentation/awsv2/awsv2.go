// Copyright 2017-2017 Amazon.com, Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may not use this file except in compliance with the License. A copy of the License is located at
//
//     http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.

package awsv2

import (
	"context"

	v2Middleware "github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/smithy-go/middleware"
	smithyhttp "github.com/aws/smithy-go/transport/http"
	"github.com/aws/aws-xray-sdk-go/v2/xray"
)

// RequestIDKey is the key name of the request id.
const RequestIDKey string = "request_id"

// ExtendedRequestIDKey is the key name of the extend request id.
const ExtendedRequestIDKey string = "id_2"

// S3ExtendedRequestIDHeaderKey is the key name of the s3 extend request id.
const S3ExtendedRequestIDHeaderKey string = "x-amz-id-2"

type awsV2SubsegmentKey struct{}

func initializeMiddlewareAfter(stack *middleware.Stack) error {
	return stack.Initialize.Add(middleware.InitializeMiddlewareFunc("XRayInitializeMiddlewareAfter", func(
		ctx context.Context, in middleware.InitializeInput, next middleware.InitializeHandler) (
		out middleware.InitializeOutput, metadata middleware.Metadata, err error) {

		serviceName := v2Middleware.GetServiceID(ctx)
		// Start the subsegment
		ctx, subseg := xray.BeginSubsegment(ctx, serviceName)
		if subseg == nil {
			return next.HandleInitialize(ctx, in)
		}
		subseg.Namespace = "aws"
		subseg.GetAWS()["region"] = v2Middleware.GetRegion(ctx)
		subseg.GetAWS()["operation"] = v2Middleware.GetOperationName(ctx)

		// set the subsegment in the context
		ctx = context.WithValue(ctx, awsV2SubsegmentKey{}, subseg)

		out, metadata, err = next.HandleInitialize(ctx, in)

		// End the subsegment when the response returns from this middleware
		defer subseg.Close(err)

		return out, metadata, err
	}),
		middleware.After)
}

func deserializeMiddleware(stack *middleware.Stack) error {
	return stack.Deserialize.Add(middleware.DeserializeMiddlewareFunc("XRayDeserializeMiddleware", func(
		ctx context.Context, in middleware.DeserializeInput, next middleware.DeserializeHandler) (
		out middleware.DeserializeOutput, metadata middleware.Metadata, err error) {

		subseg, ok := ctx.Value(awsV2SubsegmentKey{}).(*xray.Segment)
		if !ok {
			return next.HandleDeserialize(ctx, in)
		}

		in.Request.(*smithyhttp.Request).Header.Set(xray.TraceIDHeaderKey, subseg.DownstreamHeader().String())

		out, metadata, err = next.HandleDeserialize(ctx, in)

		resp, ok := out.RawResponse.(*smithyhttp.Response)
		if !ok {
			// No raw response to wrap with.
			return out, metadata, err
		}

		// Lock subseg before updating
		subseg.Lock()

		subseg.GetHTTP().GetResponse().ContentLength = int(resp.ContentLength)
		requestID, ok := v2Middleware.GetRequestIDMetadata(metadata)

		if ok {
			subseg.GetAWS()[RequestIDKey] = requestID
		}
		if extendedRequestID := resp.Header.Get(S3ExtendedRequestIDHeaderKey); extendedRequestID != "" {
			subseg.GetAWS()[ExtendedRequestIDKey] = extendedRequestID
		}

		subseg.Unlock()

		xray.HttpCaptureResponse(subseg, resp.StatusCode)
		return out, metadata, err
	}),
		middleware.Before)
}

func AWSV2Instrumentor(apiOptions *[]func(*middleware.Stack) error) {
	*apiOptions = append(*apiOptions, initializeMiddlewareAfter, deserializeMiddleware)
}
