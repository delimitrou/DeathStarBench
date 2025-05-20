// Copyright 2017-2017 Amazon.com, Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may not use this file except in compliance with the License. A copy of the License is located at
//
//     http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.

package sampling

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/aws/aws-xray-sdk-go/v2/daemoncfg"
	"github.com/aws/aws-xray-sdk-go/v2/internal/logger"
)

// proxy is an implementation of svcProxy that forwards requests to the XRay daemon
type proxy struct {
	// XRay client for sending unsigned proxied requests to the daemon
	xray *xrayClient
}

// NewProxy returns a Proxy
func newProxy(d *daemoncfg.DaemonEndpoints) (svcProxy, error) {
	if d == nil {
		d = daemoncfg.GetDaemonEndpoints()
	}
	logger.Infof("X-Ray proxy using address : %v", d.TCPAddr.String())
	url := "http://" + d.TCPAddr.String()

	// Construct resolved URLs for getSamplingRules and getSamplingTargets API calls.
	samplingRulesURL := url + "/GetSamplingRules"
	samplingTargetsURL := url + "/SamplingTargets"

	p := &proxy{
		xray: &xrayClient{
			httpClient:         &http.Client{},
			samplingRulesURL:   samplingRulesURL,
			samplingTargetsURL: samplingTargetsURL,
		},
	}

	return p, nil
}

type xrayClient struct {
	// HTTP client for sending sampling requests to the collector.
	httpClient *http.Client

	// Resolved URL to call getSamplingRules API.
	samplingRulesURL string

	// Resolved URL to call getSamplingTargets API.
	samplingTargetsURL string
}

// getSamplingRules calls the collector(aws proxy enabled) for sampling rules.
func (c *xrayClient) getSamplingRules() (*GetSamplingRulesOutput, error) {
	emptySamplingRulesInputJSON := []byte(`{"NextToken": null}`)

	body := bytes.NewReader(emptySamplingRulesInputJSON)

	req, err := http.NewRequest(http.MethodPost, c.samplingRulesURL, body)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve sampling rules, error on http request: %w", err)
	}

	output, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("xray client: unable to retrieve sampling rules, error on http request:  %w", err)
	}
	defer output.Body.Close()

	if output.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("xray client: unable to retrieve sampling rules, expected response status code 200, got: %d", output.StatusCode)
	}

	var samplingRulesOutput *GetSamplingRulesOutput
	if err := json.NewDecoder(output.Body).Decode(&samplingRulesOutput); err != nil {
		return nil, fmt.Errorf("xray client: unable to retrieve sampling rules, unable to unmarshal the response body: %w", err)
	}

	return samplingRulesOutput, nil
}

// getSamplingTargets calls the Daemon (aws proxy enabled) for sampling targets.
func (c *xrayClient) getSamplingTargets(s []*SamplingStatisticsDocument) (*GetSamplingTargetsOutput, error) {
	statistics := GetSamplingTargetsInput{
		SamplingStatisticsDocuments: s,
	}

	statisticsByte, err := json.Marshal(statistics)
	if err != nil {
		return nil, err
	}
	body := bytes.NewReader(statisticsByte)

	req, err := http.NewRequest(http.MethodPost, c.samplingTargetsURL, body)
	if err != nil {
		return nil, fmt.Errorf("xray client: failed to create http request: %w", err)
	}

	output, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("xray client: unable to retrieve sampling targets, error on http request: %w", err)
	}
	defer output.Body.Close()

	if output.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("xray client: unable to retrieve sampling targets, expected response status code 200, got: %d", output.StatusCode)
	}

	var samplingTargetsOutput *GetSamplingTargetsOutput
	if err := json.NewDecoder(output.Body).Decode(&samplingTargetsOutput); err != nil {
		return nil, fmt.Errorf("xray client: unable to retrieve sampling targets, unable to unmarshal the response body: %w", err)
	}

	return samplingTargetsOutput, nil
}

// GetSamplingTargets calls the XRay daemon for sampling targets
func (p *proxy) GetSamplingTargets(s []*SamplingStatisticsDocument) (*GetSamplingTargetsOutput, error) {
	output, err := p.xray.getSamplingTargets(s)
	if err != nil {
		return nil, err
	}

	return output, nil
}

// GetSamplingRules calls the XRay daemon for sampling rules
func (p *proxy) GetSamplingRules() ([]*SamplingRuleRecord, error) {
	output, err := p.xray.getSamplingRules()
	if err != nil {
		return nil, err
	}

	rules := output.SamplingRuleRecords

	return rules, nil
}
