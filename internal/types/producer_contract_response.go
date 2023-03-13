package types

import "fmt"

// ProducerContractResponse for returning summary of producer based test results
type ProducerContractResponse struct {
	Results   map[string]string  `yaml:"results" json:"results"`
	Errors    map[string]string  `yaml:"errors" json:"errors"`
	Metrics   map[string]float64 `yaml:"metrics" json:"metrics"`
	URLs      map[string]int     `yaml:"urls" json:"urls"`
	Succeeded int                `yaml:"succeeded" json:"succeeded"`
	Failed    int                `yaml:"failed" json:"failed"`
}

// NewProducerContractResponse constructor
func NewProducerContractResponse() *ProducerContractResponse {
	return &ProducerContractResponse{
		Errors:    make(map[string]string),
		Results:   make(map[string]string),
		Metrics:   make(map[string]float64),
		URLs:      make(map[string]int),
		Succeeded: 0,
		Failed:    0,
	}
}

// Add result or error
func (cr *ProducerContractResponse) Add(key string, res any, err error) {
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
