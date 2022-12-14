package types

// ChaosResponse for response of chaos request
type ChaosResponse struct {
	Errors    []string `yaml:"errors" json:"errors"`
	Succeeded int      `yaml:"succeeded" json:"succeeded"`
	Failed    int      `yaml:"failed" json:"failed"`
}

// NewChaosResponse constructor
func NewChaosResponse(errs []string, succeeded int, failed int) *ChaosResponse {
	return &ChaosResponse{
		Errors:    errs,
		Succeeded: succeeded,
		Failed:    failed,
	}
}

func (cr *ChaosResponse) Add(err error) {
	if err != nil {
		cr.Errors = append(cr.Errors, err.Error())
		cr.Failed++
	} else {
		cr.Succeeded++
	}
}
