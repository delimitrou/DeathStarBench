// Copyright 2017-2017 Amazon.com, Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may not use this file except in compliance with the License. A copy of the License is located at
//
//     http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.

package sampling

import (
	"sync"

	"github.com/aws/aws-xray-sdk-go/v2/internal/logger"
	"github.com/aws/aws-xray-sdk-go/v2/pattern"
	"github.com/aws/aws-xray-sdk-go/v2/utils"
)

// Properties is the base set of properties that define a sampling rule.
type Properties struct {
	ServiceName string  `json:"service_name"`
	Host        string  `json:"host"`
	HTTPMethod  string  `json:"http_method"`
	URLPath     string  `json:"url_path"`
	FixedTarget int64   `json:"fixed_target"`
	Rate        float64 `json:"rate"`
}

// AppliesTo returns true if the sampling rule matches against given parameters. False Otherwise.
// Assumes lock is already held, if required.
func (p *Properties) AppliesTo(host, path, method string) bool {
	return (host == "" || pattern.WildcardMatchCaseInsensitive(p.Host, host)) &&
		(path == "" || pattern.WildcardMatchCaseInsensitive(p.URLPath, path)) &&
		(method == "" || pattern.WildcardMatchCaseInsensitive(p.HTTPMethod, method))
}

// AppliesTo returns true if the sampling rule matches against given sampling request. False Otherwise.
// Assumes lock is already held, if required.
func (r *CentralizedRule) AppliesTo(request *Request) bool {
	return (request.Host == "" || pattern.WildcardMatchCaseInsensitive(r.Host, request.Host)) &&
		(request.URL == "" || pattern.WildcardMatchCaseInsensitive(r.URLPath, request.URL)) &&
		(request.Method == "" || pattern.WildcardMatchCaseInsensitive(r.HTTPMethod, request.Method)) &&
		(request.ServiceName == "" || pattern.WildcardMatchCaseInsensitive(r.ServiceName, request.ServiceName)) &&
		(request.ServiceType == "" || pattern.WildcardMatchCaseInsensitive(r.serviceType, request.ServiceType))
}

// CentralizedRule represents a centralized sampling rule
type CentralizedRule struct {
	// Centralized reservoir for keeping track of reservoir usage
	reservoir *CentralizedReservoir

	// Rule name identifying this rule
	ruleName string

	// Priority of matching against rule
	priority int64

	// Number of requests matched against this rule
	requests int64

	// Number of requests sampled using this rule
	sampled int64

	// Number of requests burrowed
	borrows int64

	// Timestamp for last match against this rule
	usedAt int64

	// Common sampling rule properties
	*Properties

	// ServiceType for the sampling rule
	serviceType string

	// ResourceARN for the sampling rule
	resourceARN string

	// Attributes for the sampling rule
	attributes map[string]*string

	// Provides system time
	clock utils.Clock

	// Provides random numbers
	rand utils.Rand

	mu sync.RWMutex
}

// stale returns true if the quota is due for a refresh. False otherwise.
func (r *CentralizedRule) stale(now int64) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.requests != 0 && now >= r.reservoir.refreshedAt+r.reservoir.interval
}

// Sample returns true if the request should be sampled. False otherwise.
func (r *CentralizedRule) Sample() *Decision {
	now := r.clock.Now().Unix()
	sd := &Decision{
		Rule: &r.ruleName,
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.requests++

	// Fallback to bernoulli sampling if quota has expired
	if r.reservoir.expired(now) {
		if r.reservoir.borrow(now) {
			logger.Debugf(
				"Sampling target has expired for rule %s. Borrowing a request.",
				r.ruleName,
			)
			sd.Sample = true
			r.borrows++

			return sd
		}

		logger.Debugf(
			"Sampling target has expired for rule %s. Using fixed rate.",
			r.ruleName,
		)
		sd.Sample = r.bernoulliSample()

		return sd
	}

	// Take from reservoir quota, if possible
	if r.reservoir.Take(now) {
		r.sampled++
		sd.Sample = true

		return sd
	}

	logger.Debugf(
		"Sampling target has been exhausted for rule %s. Using fixed rate.",
		r.ruleName,
	)

	// Use bernoulli sampling if quota expended
	sd.Sample = r.bernoulliSample()

	return sd
}

// bernoulliSample uses bernoulli sampling rate to make a sampling decision
func (r *CentralizedRule) bernoulliSample() bool {
	if r.rand.Float64() < r.Rate {
		r.sampled++

		return true
	}

	return false
}

// snapshot takes a snapshot of the sampling statistics counters, returning
// SamplingStatistics. It also resets statistics counters.
func (r *CentralizedRule) snapshot() *SamplingStatisticsDocument {
	r.mu.Lock()

	name := &r.ruleName

	// Copy statistics counters since SamplingStatistics expects
	// pointers to counters, and ours are mutable.
	requests, sampled, borrows := r.requests, r.sampled, r.borrows

	// Reset counters
	r.requests, r.sampled, r.borrows = 0, 0, 0

	r.mu.Unlock()

	now := r.clock.Now().Unix()
	s := &SamplingStatisticsDocument{
		RequestCount: &requests,
		SampledCount: &sampled,
		BorrowCount:  &borrows,
		RuleName:     name,
		Timestamp:    &now,
	}

	return s
}

// Rule is local sampling rule.
type Rule struct {
	reservoir *Reservoir

	// Provides random numbers
	rand utils.Rand

	// Common sampling rule properties
	*Properties

	mu sync.RWMutex
}

// Sample is used to provide sampling decision.
func (r *Rule) Sample() *Decision {
	var sd Decision
	r.mu.Lock()
	if r.reservoir.Take() {
		sd.Sample = true
	} else {
		sd.Sample = r.rand.Float64() < r.Rate
	}
	r.mu.Unlock()
	return &sd
}
