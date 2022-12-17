package types

// ChaosResponse for response of chaos request
type ChaosResponse struct {
	Results   map[string]any    `yaml:"results" json:"results"`
	Errors    map[string]string `yaml:"errors" json:"errors"`
	Succeeded int               `yaml:"succeeded" json:"succeeded"`
	Failed    int               `yaml:"failed" json:"failed"`
}

// NewChaosResponse constructor
func NewChaosResponse() *ChaosResponse {
	return &ChaosResponse{
		Errors:    make(map[string]string),
		Results:   make(map[string]any),
		Succeeded: 0,
		Failed:    0,
	}
}

// Add result or error
func (cr *ChaosResponse) Add(key string, res any, err error) {
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
