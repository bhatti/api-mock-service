package types

import "net/http"

// ProducerContractRequest for generating fuzz-data requests to an API implementation for producer based contract testing
type ProducerContractRequest struct {
	// BaseURL of remote server
	BaseURL string `yaml:"base_url" json:"base_url"`
	// ExecutionTimes for contract testing
	ExecutionTimes int `yaml:"execution_times" json:"execution_times"`
	// Track coverage
	TrackCoverage bool `yaml:"track_coverage" json:"track_coverage"`
	// RecordResults determines if contract validation results should be stored
	RecordResults bool `yaml:"record_results" json:"record_results"`
	// Verbose setting
	Verbose bool `yaml:"verbose" json:"verbose"`
	// Headers overrides
	Headers http.Header `yaml:"-" json:"-"`
	// Params local overrides
	Params map[string]any `yaml:"-" json:"-"`
}

// NewProducerContractRequest constructor
func NewProducerContractRequest(baseURL string, execTimes int) *ProducerContractRequest {
	return &ProducerContractRequest{
		BaseURL:        baseURL,
		ExecutionTimes: execTimes,
		Headers:        make(map[string][]string),
		Params:         make(map[string]any),
		RecordResults:  false,
		TrackCoverage:  false,
	}
}

// Overrides helper methods to aggregate headers and params
func (req *ProducerContractRequest) Overrides() map[string]any {
	res := make(map[string]any)
	for k, v := range req.Headers {
		res[k] = v[0]
	}
	for k, v := range req.Params {
		res[k] = v
	}
	return res
}

func (req *ProducerContractRequest) String() string {
	return "ProducerContractRequest(" + req.BaseURL + ")"
}
