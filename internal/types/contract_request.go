package types

// ContractRequest for generating random requests to an API
type ContractRequest struct {
	// BaseURL of remote server
	BaseURL string `yaml:"base_url" json:"base_url"`
	// ExecutionTimes for contract testing
	ExecutionTimes int `yaml:"execution_times" json:"execution_times"`
	// Verbose setting
	Verbose bool `yaml:"verbose" json:"verbose"`
	// Overrides local overrides
	Overrides map[string]any `yaml:"-" json:"-"`
}

// NewContractRequest constructor
func NewContractRequest(baseURL string, execTimes int) ContractRequest {
	return ContractRequest{
		BaseURL:        baseURL,
		ExecutionTimes: execTimes,
	}
}

func (req *ContractRequest) String() string {
	return "ContractRequest(" + req.BaseURL + ")"
}
