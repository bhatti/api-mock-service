package types

// CoverageSummary is a serializable (no openapi3 dep) summary of contract coverage.
// Populated when TrackCoverage=true and an OpenAPI spec is attached.
type CoverageSummary struct {
	TotalPaths              int                `json:"totalPaths"`
	CoveredPaths            int                `json:"coveredPaths"`
	Coverage                float64            `json:"coveragePercentage"`
	PathCoverage            map[string]float64 `json:"pathCoverage,omitempty"`
	UncoveredPaths          []string           `json:"uncoveredPaths,omitempty"`
	StatusCodes             map[int]int        `json:"statusCodeCoverage,omitempty"`
	MethodCoverage          map[string]float64 `json:"methodCoverage,omitempty"`
	TagCoverage             map[string]float64 `json:"tagCoverage,omitempty"`
	FieldCoverageByScenario map[string]float64 `json:"fieldCoverageByScenario,omitempty"`
}
