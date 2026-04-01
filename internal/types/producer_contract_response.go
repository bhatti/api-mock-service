package types

// ProducerContractResponse for returning summary of producer based test results
type ProducerContractResponse struct {
	Results      map[string]any                       `yaml:"results" json:"results"`
	Errors       map[string]string                    `yaml:"errors" json:"errors"`
	ErrorDetails map[string]*ContractValidationDetail `json:"error_details,omitempty"`
	Metrics      map[string]float64                   `yaml:"metrics" json:"metrics"`
	URLs         map[string]int                       `yaml:"urls" json:"urls"`
	Succeeded    int                                  `yaml:"succeeded" json:"succeeded"`
	Mismatched   int                                  `yaml:"mismatched" json:"mismatched"`
	Failed       int                                  `yaml:"failed" json:"failed"`
	Coverage     *CoverageSummary                     `json:"coverage,omitempty"`
}

// NewProducerContractResponse constructor
func NewProducerContractResponse() *ProducerContractResponse {
	return &ProducerContractResponse{
		Errors:    make(map[string]string),
		Results:   make(map[string]any),
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
			cr.Results[key] = res
		}
		cr.Succeeded++
	}
}

// SetErrorDetail attaches field-level diagnostics for a failed scenario.
func (cr *ProducerContractResponse) SetErrorDetail(key string, detail *ContractValidationDetail) {
	if cr.ErrorDetails == nil {
		cr.ErrorDetails = make(map[string]*ContractValidationDetail)
	}
	cr.ErrorDetails[key] = detail
}
