// Copyright 2017-2017 Amazon.com, Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may not use this file except in compliance with the License. A copy of the License is located at
//
//     http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.

package sampling

import (
	crypto "crypto/rand"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-xray-sdk-go/v2/daemoncfg"
	"github.com/aws/aws-xray-sdk-go/v2/internal/logger"
	"github.com/aws/aws-xray-sdk-go/v2/internal/plugins"

	"github.com/aws/aws-xray-sdk-go/v2/utils"
)

// CentralizedStrategy is an implementation of SamplingStrategy. It
// performs quota-based sampling with X-Ray acting as arbitrator for clients.
// It will fall back to LocalizedStrategy if sampling rules are not available from X-Ray APIS.
type CentralizedStrategy struct {
	// List of known centralized sampling rules
	manifest *CentralizedManifest

	// Sampling strategy used if centralized manifest is expired
	fallback *LocalizedStrategy

	// XRay service proxy used for getting quotas and sampling rules
	proxy svcProxy

	// Unique ID used by XRay service to identify this client
	clientID string

	// Provides system time
	clock utils.Clock

	// Provides random numbers
	rand utils.Rand

	// pollerStart, if true represents rule and target pollers are started
	pollerStart bool

	// represents daemon endpoints
	daemonEndpoints *daemoncfg.DaemonEndpoints

	mu sync.RWMutex
}

// svcProxy is the interface for API calls to X-Ray service.
type svcProxy interface {
	GetSamplingTargets(s []*SamplingStatisticsDocument) (*GetSamplingTargetsOutput, error)
	GetSamplingRules() ([]*SamplingRuleRecord, error)
}

// NewCentralizedStrategy creates a centralized sampling strategy with a fallback on
// local default rule.
func NewCentralizedStrategy() (*CentralizedStrategy, error) {
	fb, err := NewLocalizedStrategy()
	if err != nil {
		return nil, err
	}

	return newCentralizedStrategy(fb)
}

// NewCentralizedStrategyWithJSONBytes creates a centralized sampling strategy with a fallback on
// local rules specified in the given byte slice.
func NewCentralizedStrategyWithJSONBytes(b []byte) (*CentralizedStrategy, error) {
	fb, err := NewLocalizedStrategyFromJSONBytes(b)
	if err != nil {
		return nil, err
	}

	return newCentralizedStrategy(fb)
}

// NewCentralizedStrategyWithFilePath creates a centralized sampling strategy with a fallback on
// local rules located at the given file path.
func NewCentralizedStrategyWithFilePath(fp string) (*CentralizedStrategy, error) {
	fb, err := NewLocalizedStrategyFromFilePath(fp)
	if err != nil {
		return nil, err
	}

	return newCentralizedStrategy(fb)
}

func newCentralizedStrategy(fb *LocalizedStrategy) (*CentralizedStrategy, error) {
	// Generate clientID
	var r [12]byte

	_, err := crypto.Read(r[:])
	if err != nil {
		return nil, err
	}

	id := fmt.Sprintf("%02x", r)

	clock := &utils.DefaultClock{}
	rand := &utils.DefaultRand{}

	m := &CentralizedManifest{
		Rules: []*CentralizedRule{},
		Index: map[string]*CentralizedRule{},
		clock: clock,
	}

	ss := &CentralizedStrategy{
		manifest: m,
		fallback: fb,
		clientID: id,
		clock:    clock,
		rand:     rand,
	}

	return ss, nil
}

// ShouldTrace determines whether a request should be sampled. It matches the given parameters against
// a list of known rules and uses the matched rule's values to make a decision.
func (ss *CentralizedStrategy) ShouldTrace(request *Request) *Decision {
	ss.mu.Lock()
	if !ss.pollerStart {
		ss.start()
	}
	ss.mu.Unlock()
	if request.ServiceType == "" {
		request.ServiceType = plugins.InstancePluginMetadata.Origin
	}
	logger.Debugf(
		"Determining ShouldTrace decision for:\n\thost: %s\n\tpath: %s\n\tmethod: %s\n\tservicename: %s\n\tservicetype: %s",
		request.Host,
		request.URL,
		request.Method,
		request.ServiceName,
		request.ServiceType,
	)

	// Use fallback if manifest is expired
	if ss.manifest.expired() {
		logger.Debug("Centralized sampling data expired. Using fallback sampling strategy")

		return ss.fallback.ShouldTrace(request)
	}

	ss.manifest.mu.RLock()
	defer ss.manifest.mu.RUnlock()

	// Match against known rules
	for _, r := range ss.manifest.Rules {

		r.mu.RLock()
		applicable := r.AppliesTo(request)
		r.mu.RUnlock()

		if !applicable {
			continue
		}

		logger.Debugf("Applicable rule: %s", r.ruleName)

		return r.Sample()
	}

	// Match against default rule
	if r := ss.manifest.Default; r != nil {
		logger.Debugf("Applicable rule: %s", r.ruleName)

		return r.Sample()
	}

	// Use fallback if default rule is unavailable
	logger.Debug("Centralized default sampling rule unavailable. Using fallback sampling strategy")

	return ss.fallback.ShouldTrace(request)
}

// start initiates rule and target pollers.
func (ss *CentralizedStrategy) start() {
	if !ss.pollerStart {
		var er error
		ss.proxy, er = newProxy(ss.daemonEndpoints)
		if er != nil {
			panic(er)
		}
		ss.startRulePoller()
		ss.startTargetPoller()
	}

	ss.pollerStart = true
}

// startRulePoller starts rule poller.
func (ss *CentralizedStrategy) startRulePoller() {
	// Initial refresh
	go func() {
		if err := ss.refreshManifest(); err != nil {
			logger.Debugf("Error occurred during initial refresh of sampling rules. %v", err)
		} else {
			logger.Info("Successfully fetched sampling rules")
		}
	}()

	// Periodic manifest refresh
	go func() {
		// Period = 300s, Jitter = 5s
		t := utils.NewTimer(300*time.Second, 5*time.Second)

		for range t.C() {
			t.Reset()
			if err := ss.refreshManifest(); err != nil {
				logger.Debugf("Error occurred while refreshing sampling rules. %v", err)
			} else {
				logger.Debug("Successfully fetched sampling rules")
			}
		}
	}()
}

// startTargetPoller starts target poller.
func (ss *CentralizedStrategy) startTargetPoller() {
	// Periodic quota refresh
	go func() {
		// Period = 10.1s, Jitter = 100ms
		t := utils.NewTimer(10*time.Second+100*time.Millisecond, 100*time.Millisecond)

		for range t.C() {
			t.Reset()
			if err := ss.refreshTargets(); err != nil {
				logger.Debugf("Error occurred while refreshing targets for sampling rules. %v", err)
			}
		}
	}()
}

// refreshManifest refreshes the manifest by calling the XRay service proxy for sampling rules
func (ss *CentralizedStrategy) refreshManifest() (err error) {
	// Explicitly recover from panics since this is the entry point for a long-running goroutine
	// and we can not allow a panic to propagate to the application code.
	defer func() {
		if r := recover(); r != nil {
			// Resort to bring rules array into consistent state.
			ss.manifest.sort()

			err = fmt.Errorf("%v", r)
		}
	}()

	// Compute 'now' before calling GetSamplingRules to avoid marking manifest as
	// fresher than it actually is.
	now := ss.clock.Now().Unix()

	// Get sampling rules from proxy
	records, err := ss.proxy.GetSamplingRules()
	if err != nil {
		return
	}

	// Set of rules to exclude from pruning
	actives := map[*CentralizedRule]bool{}

	// Create missing rules. Update existing ones.
	failed := false
	for _, record := range records {
		// Extract rule from record
		svcRule := record.SamplingRule
		if svcRule == nil {
			logger.Debug("Sampling rule missing from sampling rule record.")
			failed = true
			continue
		}

		if svcRule.RuleName == nil {
			logger.Debug("Sampling rule without rule name is not supported")
			failed = true
			continue
		}

		// Only sampling rule with version 1 is valid
		if svcRule.Version == nil {
			logger.Debug("Sampling rule without version number is not supported: ", *svcRule.RuleName)
			failed = true
			continue
		}
		version := *svcRule.Version
		if version != int64(1) {
			logger.Debug("Sampling rule without version 1 is not supported: ", *svcRule.RuleName)
			failed = true
			continue
		}

		if len(svcRule.Attributes) != 0 {
			logger.Debug("Sampling rule with non nil Attributes is not applicable: ", *svcRule.RuleName)
			continue
		}

		if svcRule.ResourceARN == nil {
			logger.Debug("Sampling rule without ResourceARN is not applicable: ", *svcRule.RuleName)
			continue
		}

		resourceARN := *svcRule.ResourceARN
		if resourceARN != "*" {
			logger.Debug("Sampling rule with ResourceARN not equal to * is not applicable: ", *svcRule.RuleName)
			continue
		}

		// Create/update rule
		r, putErr := ss.manifest.putRule(svcRule)
		if putErr != nil {
			failed = true
			logger.Debugf("Error occurred creating/updating rule. %v", putErr)
		} else if r != nil {
			actives[r] = true
		}
	}

	// Set err if updates failed
	if failed {
		err = errors.New("error occurred creating/updating rules")
	}

	// Prune inactive rules
	ss.manifest.prune(actives)

	// Re-sort to fix matching priorities
	ss.manifest.sort()

	// Update refreshedAt timestamp
	ss.manifest.mu.Lock()
	ss.manifest.refreshedAt = now
	ss.manifest.mu.Unlock()

	return
}

// refreshTargets refreshes targets for sampling rules. It calls the XRay service proxy with sampling
// statistics for the previous interval and receives targets for the next interval.
func (ss *CentralizedStrategy) refreshTargets() (err error) {
	// Explicitly recover from panics since this is the entry point for a long-running goroutine
	// and we can not allow a panic to propagate to customer code.
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()

	// Flag indicating batch failure
	failed := false

	// Flag indicating whether or not manifest should be refreshed
	refresh := false

	// Generate sampling statistics
	statistics := ss.snapshots()

	// Do not refresh targets if no statistics to report
	if len(statistics) == 0 {
		logger.Debugf("No statistics to report. Not refreshing sampling targets.")
		return nil
	}

	// Get sampling targets
	output, err := ss.proxy.GetSamplingTargets(statistics)
	if err != nil {
		return
	}

	// Update sampling targets
	for _, t := range output.SamplingTargetDocuments {
		if err = ss.updateTarget(t); err != nil {
			failed = true
			logger.Debugf("Error occurred updating target for rule. %v", err)
		}
	}

	// Consume unprocessed statistics messages
	for _, s := range output.UnprocessedStatistics {
		logger.Debugf(
			"Error occurred updating sampling target for rule: %s, code: %s, message: %s",
			s.RuleName,
			s.ErrorCode,
			s.Message,
		)

		// Do not set any flags if error is unknown
		if s.ErrorCode == nil || s.RuleName == nil {
			continue
		}

		// Set batch failure if any sampling statistics return 5xx
		if strings.HasPrefix(*s.ErrorCode, "5") {
			failed = true
		}

		// Set refresh flag if any sampling statistics return 4xx
		if strings.HasPrefix(*s.ErrorCode, "4") {
			refresh = true
		}
	}

	// Set err if updates failed
	if failed {
		err = errors.New("error occurred updating sampling targets")
	} else {
		logger.Debug("Successfully refreshed sampling targets")
	}

	// Set refresh flag if modifiedAt timestamp from remote is greater than ours.
	if remote := output.LastRuleModification; remote != nil {
		ss.manifest.mu.RLock()
		local := ss.manifest.refreshedAt
		ss.manifest.mu.RUnlock()

		if *remote >= float64(local) {
			refresh = true
		}
	}
	// Perform out-of-band async manifest refresh if flag is set
	if refresh {
		logger.Infof("Refreshing sampling rules out-of-band.")

		go func() {
			if err := ss.refreshManifest(); err != nil {
				logger.Debugf("Error occurred refreshing sampling rules out-of-band. %v", err)
			}
		}()
	}
	return
}

// samplingStatistics takes a snapshot of sampling statistics from all rules, resetting
// statistics counters in the process.
func (ss *CentralizedStrategy) snapshots() []*SamplingStatisticsDocument {
	now := ss.clock.Now().Unix()

	ss.manifest.mu.RLock()
	defer ss.manifest.mu.RUnlock()

	statistics := make([]*SamplingStatisticsDocument, 0, len(ss.manifest.Rules)+1)

	// Generate sampling statistics for user-defined rules
	for _, r := range ss.manifest.Rules {
		if !r.stale(now) {
			continue
		}

		s := r.snapshot()
		s.ClientID = &ss.clientID

		statistics = append(statistics, s)
	}

	// Generate sampling statistics for default rule
	if r := ss.manifest.Default; r != nil && r.stale(now) {
		s := r.snapshot()
		s.ClientID = &ss.clientID

		statistics = append(statistics, s)
	}

	return statistics
}

// updateTarget updates sampling targets for the rule specified in the target struct.
func (ss *CentralizedStrategy) updateTarget(t *SamplingTargetDocument) (err error) {
	// Pre-emptively dereference SamplingTarget fields and return early on nil values
	// A panic in the middle of an update may leave the rule in an inconsistent state.
	if t.RuleName == nil {
		return errors.New("invalid sampling target. Missing rule name")
	}

	if t.FixedRate == nil {
		return fmt.Errorf("invalid sampling target for rule %s. Missing fixed rate", *t.RuleName)
	}

	// Rule for given target
	ss.manifest.mu.RLock()
	r, ok := ss.manifest.Index[*t.RuleName]
	ss.manifest.mu.RUnlock()

	if !ok {
		return fmt.Errorf("rule %s not found", *t.RuleName)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.reservoir.refreshedAt = ss.clock.Now().Unix()

	// Update non-optional attributes from response
	r.Rate = *t.FixedRate

	// Update optional attributes from response
	if t.ReservoirQuota != nil {
		r.reservoir.quota = *t.ReservoirQuota
	}
	if t.ReservoirQuotaTTL != nil {
		r.reservoir.expiresAt = int64(*t.ReservoirQuotaTTL)
	}
	if t.Interval != nil {
		r.reservoir.interval = *t.Interval
	}

	return nil
}

// LoadDaemonEndpoints configures proxy with the provided endpoint.
func (ss *CentralizedStrategy) LoadDaemonEndpoints(endpoints *daemoncfg.DaemonEndpoints) {
	ss.daemonEndpoints = endpoints
}
