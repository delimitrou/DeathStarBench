// Copyright 2017-2017 Amazon.com, Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may not use this file except in compliance with the License. A copy of the License is located at
//
//     http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.

package sampling

import "github.com/aws/aws-xray-sdk-go/v2/utils"

// Reservoirs allow a specified (`perSecond`) amount of `Take()`s per second.

// reservoir is a set of properties common to all reservoirs
type reservoir struct {
	// Total size of reservoir
	capacity int64

	// Reservoir consumption for current epoch
	used int64

	// Unix epoch. Reservoir usage is reset every second.
	currentEpoch int64
}

// CentralizedReservoir is a reservoir distributed among all running instances of the SDK
type CentralizedReservoir struct {
	// Quota assigned to client
	quota int64

	// Quota refresh timestamp
	refreshedAt int64

	// Quota expiration timestamp
	expiresAt int64

	// Polling interval for quota
	interval int64

	// True if reservoir has been borrowed from this epoch
	borrowed bool

	// Common reservoir properties
	*reservoir
}

// expired returns true if current time is past expiration timestamp. False otherwise.
func (r *CentralizedReservoir) expired(now int64) bool {
	return now > r.expiresAt
}

// borrow returns true if the reservoir has not been borrowed from this epoch
func (r *CentralizedReservoir) borrow(now int64) bool {
	if now != r.currentEpoch {
		r.reset(now)
	}

	s := r.borrowed
	r.borrowed = true

	return !s && r.reservoir.capacity != 0
}

// Take consumes quota from reservoir, if any remains, and returns true. False otherwise.
func (r *CentralizedReservoir) Take(now int64) bool {
	if now != r.currentEpoch {
		r.reset(now)
	}

	// Consume from quota, if available
	if r.quota > r.used {
		r.used++

		return true
	}

	return false
}

func (r *CentralizedReservoir) reset(now int64) {
	r.currentEpoch, r.used, r.borrowed = now, 0, false
}

// Reservoir is a reservoir local to the running instance of the SDK
type Reservoir struct {
	// Provides system time
	clock utils.Clock

	*reservoir
}

// Take attempts to consume a unit from the local reservoir. Returns true if unit taken, false otherwise.
func (r *Reservoir) Take() bool {
	// Reset counters if new second
	if now := r.clock.Now().Unix(); now != r.currentEpoch {
		r.used = 0
		r.currentEpoch = now
	}

	// Take from reservoir, if available
	if r.used >= r.capacity {
		return false
	}

	// Increment reservoir usage
	r.used++

	return true
}
