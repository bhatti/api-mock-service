package types

import "fmt"

// ContractResponse for response of contract request
type ContractResponse struct {
	Results   map[string]string  `yaml:"results" json:"results"`
	Errors    map[string]string  `yaml:"errors" json:"errors"`
	Metrics   map[string]float64 `yaml:"metrics" json:"metrics"`
	Succeeded int                `yaml:"succeeded" json:"succeeded"`
	Failed    int                `yaml:"failed" json:"failed"`
}

// NewContractResponse constructor
func NewContractResponse() *ContractResponse {
	return &ContractResponse{
		Errors:    make(map[string]string),
		Results:   make(map[string]string),
		Metrics:   make(map[string]float64),
		Succeeded: 0,
		Failed:    0,
	}
}

// Add result or error
func (cr *ContractResponse) Add(key string, res any, err error) {
	if err != nil {
		cr.Errors[key] = err.Error()
		cr.Failed++
	} else {
		if res != nil {
			cr.Results[key] = fmt.Sprintf("%v", res)
		}
		cr.Succeeded++
	}
}
