package types

// ContractRequest for generating random requests to an API
type ContractRequest struct {
	// BaseURL of remote server
	BaseURL string `yaml:"base_url" json:"base_url"`
	// ExecutionTimes for contract testing
	ExecutionTimes int `yaml:"execution_times" json:"execution_times"`
	// Verbose setting
	Verbose bool `yaml:"verbose" json:"verbose"`
	// Headers overrides
	Headers map[string][]string `yaml:"-" json:"-"`
	// Params local overrides
	Params map[string]any `yaml:"-" json:"-"`
}

// NewContractRequest constructor
func NewContractRequest(baseURL string, execTimes int) *ContractRequest {
	return &ContractRequest{
		BaseURL:        baseURL,
		ExecutionTimes: execTimes,
		Headers:        make(map[string][]string),
		Params:         make(map[string]any),
	}
}

// Overrides helper methods to aggregate headers and params
func (req *ContractRequest) Overrides() map[string]any {
	res := make(map[string]any)
	for k, v := range req.Headers {
		res[k] = v[0]
	}
	for k, v := range req.Params {
		res[k] = v
	}
	return res
}

func (req *ContractRequest) String() string {
	return "ContractRequest(" + req.BaseURL + ")"
}
