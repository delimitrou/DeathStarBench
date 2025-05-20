// Types in this file are modified from:
// https://github.com/aws/aws-sdk-go/blob/v1.55.6/service/xray/api.go

package sampling

type SamplingStatisticsDocument struct {
	// The number of requests recorded with borrowed reservoir quota.
	BorrowCount *int64 `type:"integer"`

	// A unique identifier for the service in hexadecimal.
	//
	// ClientID is a required field
	ClientID *string `min:"24" type:"string" required:"true"`

	// The number of requests that matched the rule.
	//
	// RequestCount is a required field
	RequestCount *int64 `type:"integer" required:"true"`

	// The name of the sampling rule.
	//
	// RuleName is a required field
	RuleName *string `min:"1" type:"string" required:"true"`

	// The number of requests recorded.
	//
	// SampledCount is a required field
	SampledCount *int64 `type:"integer" required:"true"`

	// The current time.
	//
	// Timestamp is a required field
	Timestamp *int64 `type:"integer" required:"true"`
}

// A sampling rule that services use to decide whether to instrument a request.
// Rule fields can match properties of the service, or properties of a request.
// The service can ignore rules that don't match its properties.
type SamplingRule struct {
	// Matches attributes derived from the request.
	Attributes map[string]*string `json:"Attributes"`

	// The percentage of matching requests to instrument, after the reservoir is
	// exhausted.
	//
	// FixedRate is a required field
	FixedRate *float64 `json:"FixedRate" required:"true"`

	// Matches the HTTP method of a request.
	//
	// HTTPMethod is a required field
	HTTPMethod *string `json:"HTTPMethod" required:"true"`

	// Matches the hostname from a request URL.
	//
	// Host is a required field
	Host *string `json:"Host" required:"true"`

	// The priority of the sampling rule.
	//
	// Priority is a required field
	Priority *int64 `min:"1" json:"Priority" required:"true"`

	// A fixed number of matching requests to instrument per second, prior to applying
	// the fixed rate. The reservoir is not used directly by services, but applies
	// to all services using the rule collectively.
	//
	// ReservoirSize is a required field
	ReservoirSize *int64 `json:"ReservoirSize" required:"true"`

	// Matches the ARN of the Amazon Web Services resource on which the service
	// runs.
	//
	// ResourceARN is a required field
	ResourceARN *string `json:"ResourceARN" required:"true"`

	// The ARN of the sampling rule. Specify a rule by either name or ARN, but not
	// both.
	RuleARN *string `json:"RuleARN"`

	// The name of the sampling rule. Specify a rule by either name or ARN, but
	// not both.
	RuleName *string `min:"1" json:"RuleName"`

	// Matches the name that the service uses to identify itself in segments.
	//
	// ServiceName is a required field
	ServiceName *string `json:"ServiceName" required:"true"`

	// Matches the origin that the service uses to identify its type in segments.
	//
	// ServiceType is a required field
	ServiceType *string `json:"ServiceType" required:"true"`

	// Matches the path from a request URL.
	//
	// URLPath is a required field
	URLPath *string `json:"URLPath" required:"true"`

	// The version of the sampling rule format (1).
	//
	// Version is a required field
	Version *int64 `min:"1" json:"Version" required:"true"`
}

// A SamplingRule (https://docs.aws.amazon.com/xray/latest/api/API_SamplingRule.html)
// and its metadata.
type SamplingRuleRecord struct {
	// When the rule was created.
	CreatedAt *float64 `json:"CreatedAt"`

	// When the rule was last modified.
	ModifiedAt *float64 `json:"ModifiedAt"`

	// The sampling rule.
	SamplingRule *SamplingRule `json:"SamplingRule"`
}

type GetSamplingTargetsOutput struct {
	// The last time a user changed the sampling rule configuration. If the sampling
	// rule configuration changed since the service last retrieved it, the service
	// should call GetSamplingRules (https://docs.aws.amazon.com/xray/latest/api/API_GetSamplingRules.html)
	// to get the latest version.
	LastRuleModification *float64 `json:"LastRuleModification"`

	// Updated rules that the service should use to sample requests.
	SamplingTargetDocuments []*SamplingTargetDocument `json:"SamplingTargetDocuments"`

	// Information about SamplingStatisticsDocument (https://docs.aws.amazon.com/xray/latest/api/API_SamplingStatisticsDocument.html)
	// that X-Ray could not process.
	UnprocessedStatistics []*UnprocessedStatistics `json:"UnprocessedStatistics"`
}

// Temporary changes to a sampling rule configuration. To meet the global sampling
// target for a rule, X-Ray calculates a new reservoir for each service based
// on the recent sampling results of all services that called GetSamplingTargets
// (https://docs.aws.amazon.com/xray/latest/api/API_GetSamplingTargets.html).
type SamplingTargetDocument struct {
	// The percentage of matching requests to instrument, after the reservoir is
	// exhausted.
	FixedRate *float64 `json:"FixedRate"`

	// The number of seconds for the service to wait before getting sampling targets
	// again.
	Interval *int64 `json:"Interval"`

	// The number of requests per second that X-Ray allocated for this service.
	ReservoirQuota *int64 `json:"ReservoirQuota"`

	// When the reservoir quota expires.
	ReservoirQuotaTTL *float64 `json:"ReservoirQuotaTTL"`

	// The name of the sampling rule.
	RuleName *string `json:"RuleName"`
}

// Sampling statistics from a call to GetSamplingTargets (https://docs.aws.amazon.com/xray/latest/api/API_GetSamplingTargets.html)
// that X-Ray could not process.
type UnprocessedStatistics struct {
	// The error code.
	ErrorCode *string `json:"ErrorCode"`

	// The error message.
	Message *string `json:"Message"`

	// The name of the sampling rule.
	RuleName *string `json:"RuleName"`
}

type GetSamplingTargetsInput struct {
	// Information about rules that the service is using to sample requests.
	//
	// SamplingStatisticsDocuments is a required field
	SamplingStatisticsDocuments []*SamplingStatisticsDocument `type:"list" required:"true"`
}

type GetSamplingRulesInput struct {
	// Pagination token.
	NextToken *string `type:"string"`
}

type GetSamplingRulesOutput struct {
	// Pagination token.
	NextToken *string `json:"NextToken"`

	// Rule definitions and metadata.
	SamplingRuleRecords []*SamplingRuleRecord `json:"SamplingRuleRecords"`
}
