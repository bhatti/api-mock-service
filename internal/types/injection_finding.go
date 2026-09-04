package types

// InjectionFinding represents a detected vulnerability in an API response.
type InjectionFinding struct {
	Type     string `json:"type" yaml:"type"`
	Severity string `json:"severity" yaml:"severity"`
	Evidence string `json:"evidence" yaml:"evidence"`
	Pattern  string `json:"pattern" yaml:"pattern"`
	Payload  string `json:"payload" yaml:"payload"`
	Field    string `json:"field" yaml:"field"`
	Endpoint string `json:"endpoint" yaml:"endpoint"`
	CWE      string `json:"cwe" yaml:"cwe"`
}

// SecuritySummary aggregates injection findings with OWASP classification.
type SecuritySummary struct {
	TotalFindings int                `json:"totalFindings"`
	BySeverity    map[string]int     `json:"bySeverity"`
	ByCategory    map[string]int     `json:"byCategory"`
	TopFindings   []InjectionFinding `json:"topFindings"`
	OWASPMapping  map[string]string  `json:"owaspMapping"`
	PassedChecks  []string           `json:"passedChecks"`
}

// FailureClusterKey identifies a unique class of failure for deduplication.
type FailureClusterKey struct {
	StatusCode    int    `json:"statusCode"`
	ErrorCategory string `json:"errorCategory"`
	Endpoint      string `json:"endpoint"`
	ResponseHash  string `json:"responseHash"`
}

// FailureCluster groups duplicate failures by root cause.
type FailureCluster struct {
	Key               FailureClusterKey `json:"key"`
	Count             int               `json:"count"`
	Representative    string            `json:"representative"`
	AffectedMutations []string          `json:"affectedMutations"`
	Severity          string            `json:"severity"`
}
