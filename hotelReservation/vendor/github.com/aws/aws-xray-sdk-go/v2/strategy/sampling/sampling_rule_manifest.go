// Copyright 2017-2017 Amazon.com, Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may not use this file except in compliance with the License. A copy of the License is located at
//
//     http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.

package sampling

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/aws/aws-xray-sdk-go/v2/utils"
)

// RuleManifest represents a full sampling ruleset, with a list of
// custom rules and default values for incoming requests that do
// not match any of the provided rules.
type RuleManifest struct {
	Version int     `json:"version"`
	Default *Rule   `json:"default"`
	Rules   []*Rule `json:"rules"`
}

// ManifestFromFilePath creates a sampling ruleset from a given filepath fp.
func ManifestFromFilePath(fp string) (*RuleManifest, error) {
	b, err := ioutil.ReadFile(fp)
	if err == nil {
		return ManifestFromJSONBytes(b)
	}

	return nil, err
}

// ManifestFromJSONBytes creates a sampling ruleset from given JSON bytes b.
func ManifestFromJSONBytes(b []byte) (*RuleManifest, error) {
	s := &RuleManifest{}
	err := json.Unmarshal(b, s)
	if err != nil {
		return nil, err
	}
	err = processManifest(s)
	if err != nil {
		return nil, err
	}

	initSamplingRules(s)

	return s, nil
}

// Init local reservoir and add random number generator
func initSamplingRules(srm *RuleManifest) {
	// Init user-defined rules
	for _, r := range srm.Rules {
		r.rand = &utils.DefaultRand{}

		r.reservoir = &Reservoir{
			clock: &utils.DefaultClock{},
			reservoir: &reservoir{
				capacity:     r.FixedTarget,
				used:         0,
				currentEpoch: time.Now().Unix(),
			},
		}
	}

	// Init default rule
	srm.Default.rand = &utils.DefaultRand{}

	srm.Default.reservoir = &Reservoir{
		clock: &utils.DefaultClock{},
		reservoir: &reservoir{
			capacity:     srm.Default.FixedTarget,
			used:         0,
			currentEpoch: time.Now().Unix(),
		},
	}
}

// processManifest returns the provided manifest if valid, or an error if the provided manifest is invalid.
func processManifest(srm *RuleManifest) error {
	if srm == nil {
		return errors.New("sampling rule manifest must not be nil")
	}
	if srm.Version != 1 && srm.Version != 2 {
		return fmt.Errorf("sampling rule manifest version %d not supported", srm.Version)
	}
	if srm.Default == nil {
		return errors.New("sampling rule manifest must include a default rule")
	}
	if srm.Default.URLPath != "" || srm.Default.ServiceName != "" || srm.Default.HTTPMethod != "" {
		return errors.New("the default rule must not specify values for url_path, service_name, or http_method")
	}
	if srm.Default.FixedTarget < 0 || srm.Default.Rate < 0 {
		return errors.New("the default rule must specify non-negative values for fixed_target and rate")
	}

	c := &utils.DefaultClock{}

	srm.Default.reservoir = &Reservoir{
		clock: c,
		reservoir: &reservoir{
			capacity: srm.Default.FixedTarget,
		},
	}

	if srm.Rules != nil {
		for _, r := range srm.Rules {

			if srm.Version == 1 {
				if err := validateVersion1(r); err != nil {
					return err
				}
				r.Host = r.ServiceName // V1 sampling rule contains service name and not host
				r.ServiceName = ""
			}

			if srm.Version == 2 {
				if err := validateVersion2(r); err != nil {
					return err
				}
			}

			r.reservoir = &Reservoir{
				clock: c,
				reservoir: &reservoir{
					capacity: r.FixedTarget,
				},
			}
		}
	}
	return nil
}

func validateVersion2(rule *Rule) error {
	if rule.FixedTarget < 0 || rule.Rate < 0 {
		return errors.New("all rules must have non-negative values for fixed_target and rate")
	}
	if rule.ServiceName != "" || rule.Host == "" || rule.HTTPMethod == "" || rule.URLPath == "" {
		return errors.New("all non-default rules must have values for url_path, host, and http_method")
	}
	return nil
}

func validateVersion1(rule *Rule) error {
	if rule.FixedTarget < 0 || rule.Rate < 0 {
		return errors.New("all rules must have non-negative values for fixed_target and rate")
	}
	if rule.Host != "" || rule.ServiceName == "" || rule.HTTPMethod == "" || rule.URLPath == "" {
		return errors.New("all non-default rules must have values for url_path, service_name, and http_method")
	}
	return nil
}
