// Copyright 2017-2017 Amazon.com, Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may not use this file except in compliance with the License. A copy of the License is located at
//
//     http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.

package utils

// MockRand is an implementation of Rand interface.
type MockRand struct {
	F64   float64
	Int   int
	Int64 int64
}

// Float64 returns value of F64.
func (r *MockRand) Float64() float64 {
	return r.F64
}

// Intn returns value of Int.
func (r *MockRand) Intn(n int) int {
	return r.Int
}

// Int63n returns value of Int64.
func (r *MockRand) Int63n(n int64) int64 {
	return r.Int64
}
