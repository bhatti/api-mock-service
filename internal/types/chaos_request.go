package types

// ChaosRequest for generating random requests to an API
type ChaosRequest struct {
	BaseURL        string         `yaml:"base_url" json:"base_url"`
	ExecutionTimes int            `yaml:"execution_times" json:"execution_times"`
	Verbose        bool           `yaml:"verbose" json:"verbose"`
	Overrides      map[string]any `yaml:"-" json:"-"`
}

// NewChaosRequest constructor
func NewChaosRequest(baseURL string, execTimes int) ChaosRequest {
	return ChaosRequest{
		BaseURL:        baseURL,
		ExecutionTimes: execTimes,
	}
}
