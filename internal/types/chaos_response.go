package types

// ChaosResponse for response of chaos request
type ChaosResponse struct {
	Errors    []error `yaml:"errors" json:"errors"`
	Succeeded int     `yaml:"succeeded" json:"succeeded"`
	Failed    int     `yaml:"failed" json:"failed"`
}

// NewChaosResponse constructor
func NewChaosResponse(errs []error, succeeded int, failed int) *ChaosResponse {
	return &ChaosResponse{
		Errors:    errs,
		Succeeded: succeeded,
		Failed:    failed,
	}
}
