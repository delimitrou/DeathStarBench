// Copyright 2017-2017 Amazon.com, Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may not use this file except in compliance with the License. A copy of the License is located at
//
//     http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.

package utils

import (
	"time"
)

// Clock provides an interface to implement method for getting current time.
type Clock interface {
	Now() time.Time
	Increment(int64, int64) time.Time
}

// DefaultClock is an implementation of Clock interface.
type DefaultClock struct{}

// Now returns current time.
func (t *DefaultClock) Now() time.Time {
	return time.Now()
}

// This method returns the current time but can be used to provide different implementation
func (t *DefaultClock) Increment(_, _ int64) time.Time {
	return time.Now()
}
