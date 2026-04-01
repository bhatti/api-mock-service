package types

// ContractValidationDetail gives field-level diagnostics for a failed contract test.
// Returned in ProducerContractResponse.ErrorDetails alongside the existing flat Errors map.
type ContractValidationDetail struct {
	Summary            string                   `json:"summary"`
	Scenario           string                   `json:"scenario"`
	URL                string                   `json:"url"`
	StatusCode         int                      `json:"statusCode,omitempty"`
	ExpectedStatusCode int                      `json:"expectedStatusCode,omitempty"`
	MissingFields      []string                 `json:"missingFields,omitempty"`
	ExtraFields        []string                 `json:"extraFields,omitempty"`
	TypeMismatches     map[string]string        `json:"typeMismatches,omitempty"`
	ValueMismatches    map[string]ValueMismatch `json:"valueMismatches,omitempty"`
	HeaderMismatches   map[string]ValueMismatch `json:"headerMismatches,omitempty"`
	SchemaViolations   []SchemaViolation        `json:"schemaViolations,omitempty"`
}

// ValueMismatch holds expected vs actual values for a single field.
type ValueMismatch struct {
	Expected interface{} `json:"expected"`
	Actual   interface{} `json:"actual"`
}

// SchemaViolation represents a single OpenAPI schema violation.
type SchemaViolation struct {
	Field   string      `json:"field"`
	Message string      `json:"message"`
	Value   interface{} `json:"value,omitempty"`
}
