// Copyright 2017-2017 Amazon.com, Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may not use this file except in compliance with the License. A copy of the License is located at
//
//     http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.

package sampling

// Decision contains sampling decision and the rule matched for an incoming request
type Decision struct {
	Sample bool
	Rule   *string
}

// Request represents parameters used to make a sampling decision.
type Request struct {
	Host        string
	Method      string
	URL         string
	ServiceName string
	ServiceType string
}
