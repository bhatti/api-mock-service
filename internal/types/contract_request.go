package types

// ContractRequest for generating random requests to an API
type ContractRequest struct {
	BaseURL        string         `yaml:"base_url" json:"base_url"`
	ExecutionTimes int            `yaml:"execution_times" json:"execution_times"`
	Verbose        bool           `yaml:"verbose" json:"verbose"`
	Overrides      map[string]any `yaml:"-" json:"-"`
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
