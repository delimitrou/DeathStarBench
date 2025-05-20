// Copyright 2017-2017 Amazon.com, Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may not use this file except in compliance with the License. A copy of the License is located at
//
//     http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.

package utils

import (
	"sync/atomic"
	"time"
)

// MockClock is a struct to record current time.
type MockClock struct {
	NowTime int64
	NowNanos int64
}

// Now function returns NowTime value.
func (c *MockClock) Now() time.Time {
	return time.Unix(c.NowTime, c.NowNanos)
}

// Increment is a method to increase current time.
func (c *MockClock) Increment(s int64, ns int64) time.Time {
	sec := atomic.AddInt64(&c.NowTime, s)
	nSec := atomic.AddInt64(&c.NowNanos, ns)

	return time.Unix(sec, nSec)
}
