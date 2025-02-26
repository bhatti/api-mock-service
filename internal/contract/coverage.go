package contract

import (
	"fmt"
	"github.com/bhatti/api-mock-service/internal/types"
	"github.com/getkin/kin-openapi/openapi3"
	"time"
)

type ContractValidationError struct {
	OriginalError error
	DiffReport    *ContractDiffReport
	Scenario      string
	URL           string
}

// Error implements the error interface
func (e *ContractValidationError) Error() string {
	return fmt.Sprintf("contract validation failed for scenario '%s' at '%s': %s",
		e.Scenario, e.URL, e.OriginalError)
}

// Unwrap returns the original error
func (e *ContractValidationError) Unwrap() error {
	return e.OriginalError
}

// ContractValidationResult stores the result of a contract validation test
type ContractValidationResult struct {
	Scenario        string                 `json:"scenario"`
	Path            string                 `json:"path"`
	Method          string                 `json:"method"`
	StatusCode      int                    `json:"statusCode"`
	ResponseTime    int64                  `json:"responseTimeMs"`
	Timestamp       time.Time              `json:"timestamp"`
	Success         bool                   `json:"success"`
	ErrorMessage    string                 `json:"errorMessage,omitempty"`
	ResponseBody    string                 `json:"responseBody,omitempty"`
	SharedVariables map[string]interface{} `json:"sharedVariables,omitempty"`
	DiffReport      *ContractDiffReport    `json:"diffReport,omitempty"`
}

// ContractValidationStats summarizes validation results over time
type ContractValidationStats struct {
	ScenarioName    string
	TotalExecutions int
	SuccessCount    int
	FailureCount    int
	AverageLatency  float64
	Top5Failures    []string
	LastExecuted    time.Time
	SuccessRate     float64
}

// ContractDiffReport diff reporting structure and helper functions
type ContractDiffReport struct {
	ExpectedFields   map[string]interface{}   `json:"expectedFields"`
	ActualFields     map[string]interface{}   `json:"actualFields"`
	MissingFields    []string                 `json:"missingFields"`
	ExtraFields      []string                 `json:"extraFields"`
	TypeMismatches   map[string]string        `json:"typeMismatches"`
	ValueMismatches  map[string]ValueMismatch `json:"valueMismatches"`
	HeaderMismatches map[string]ValueMismatch `json:"headerMismatches"`
}

type ValueMismatch struct {
	Expected interface{} `json:"expected"`
	Actual   interface{} `json:"actual"`
}

// OpenAPIContractCoverage analyzes contract coverage against OpenAPI spec
type OpenAPIContractCoverage struct {
	spec            *openapi3.T
	contracts       []*types.APIScenario
	pathCoverage    map[string]float64
	overallCoverage float64
	uncoveredPaths  []string
}

// NewOpenAPIContractCoverage creates a new coverage analyzer
func NewOpenAPIContractCoverage(spec *openapi3.T, contracts []*types.APIScenario) *OpenAPIContractCoverage {
	return &OpenAPIContractCoverage{
		spec:         spec,
		contracts:    contracts,
		pathCoverage: make(map[string]float64),
	}
}

// CoverageReport structure for API contract coverage
type CoverageReport struct {
	TotalPaths     int                `json:"totalPaths"`
	CoveredPaths   int                `json:"coveredPaths"`
	Coverage       float64            `json:"coveragePercentage"`
	PathCoverage   map[string]float64 `json:"pathCoverage"`
	UncoveredPaths []string           `json:"uncoveredPaths"`
	StatusCodes    map[int]int        `json:"statusCodeCoverage"`
	MethodCoverage map[string]float64 `json:"methodCoverage"`
	TagCoverage    map[string]float64 `json:"tagCoverage"`
}

// Analyze Enhanced Analyze method to include additional coverage metrics
func (c *OpenAPIContractCoverage) Analyze() *CoverageReport {
	// Basic path coverage calculation
	totalPaths := 0
	coveredPaths := 0

	// Map contracts to paths
	contractPaths := make(map[string]bool)
	for _, contract := range c.contracts {
		key := string(contract.Method) + " " + contract.Path
		contractPaths[key] = true
	}

	// Check coverage against OpenAPI spec
	for path, pathItem := range c.spec.Paths {
		for _, method := range []string{"GET", "POST", "PUT", "DELETE", "PATCH"} {
			var op *openapi3.Operation

			switch method {
			case "GET":
				op = pathItem.Get
			case "POST":
				op = pathItem.Post
			case "PUT":
				op = pathItem.Put
			case "DELETE":
				op = pathItem.Delete
			case "PATCH":
				op = pathItem.Patch
			}

			if op != nil {
				totalPaths++
				key := method + " " + path

				if contractPaths[key] {
					coveredPaths++
					c.pathCoverage[key] = 100.0
				} else {
					c.pathCoverage[key] = 0.0
					c.uncoveredPaths = append(c.uncoveredPaths, key)
				}
			}
		}
	}

	if totalPaths > 0 {
		c.overallCoverage = float64(coveredPaths) / float64(totalPaths) * 100.0
	}

	return &CoverageReport{
		TotalPaths:     totalPaths,
		CoveredPaths:   coveredPaths,
		Coverage:       c.overallCoverage,
		PathCoverage:   c.pathCoverage,
		UncoveredPaths: c.uncoveredPaths,
		StatusCodes:    c.calculateStatusCodeCoverage(),
		MethodCoverage: c.calculateMethodCoverage(),
		TagCoverage:    c.calculateTagCoverage(),
	}
}

// Helper to calculate different coverage metrics
func (c *OpenAPIContractCoverage) calculateStatusCodeCoverage() map[int]int {
	statusCodes := make(map[int]int)

	for _, contract := range c.contracts {
		statusCodes[contract.Response.StatusCode]++
	}

	return statusCodes
}

func (c *OpenAPIContractCoverage) calculateMethodCoverage() map[string]float64 {
	methods := make(map[string]int)
	methodCovered := make(map[string]int)
	methodCoverage := make(map[string]float64)

	// Count methods in OpenAPI spec
	for _, pathItem := range c.spec.Paths {
		if pathItem.Get != nil {
			methods["GET"]++
		}
		if pathItem.Post != nil {
			methods["POST"]++
		}
		if pathItem.Put != nil {
			methods["PUT"]++
		}
		if pathItem.Delete != nil {
			methods["DELETE"]++
		}
		if pathItem.Patch != nil {
			methods["PATCH"]++
		}
	}

	// Count covered methods
	for _, contract := range c.contracts {
		methodCovered[string(contract.Method)]++
	}

	// Calculate coverage percentages
	for method, total := range methods {
		if total > 0 {
			covered := methodCovered[method]
			methodCoverage[method] = float64(covered) / float64(total) * 100.0
		} else {
			methodCoverage[method] = 0.0
		}
	}

	return methodCoverage
}

func (c *OpenAPIContractCoverage) calculateTagCoverage() map[string]float64 {
	tags := make(map[string]int)
	tagsCovered := make(map[string]int)
	tagCoverage := make(map[string]float64)

	// Count tags in OpenAPI spec
	for _, pathItem := range c.spec.Paths {
		for _, op := range []*openapi3.Operation{
			pathItem.Get, pathItem.Post, pathItem.Put,
			pathItem.Delete, pathItem.Patch,
		} {
			if op != nil {
				for _, tag := range op.Tags {
					tags[tag]++
				}
			}
		}
	}

	// Count covered tags
	for _, contract := range c.contracts {
		for _, tag := range contract.Tags {
			tagsCovered[tag]++
		}
	}

	// Calculate coverage percentages
	for tag, total := range tags {
		if total > 0 {
			covered := tagsCovered[tag]
			tagCoverage[tag] = float64(covered) / float64(total) * 100.0
		} else {
			tagCoverage[tag] = 0.0
		}
	}

	return tagCoverage
}

// ContractCoverage include tracking
type ContractCoverage struct {
	ScenarioName    string          `json:"scenarioName"`
	Timestamp       time.Time       `json:"timestamp"`
	ResponseStatus  int             `json:"responseStatus"`
	ResponseTime    int64           `json:"responseTimeMs"`
	CoveredFields   []string        `json:"coveredFields"`
	UncoveredFields []string        `json:"uncoveredFields"`
	CoveragePercent float64         `json:"coveragePercent"`
	FieldCoverage   map[string]bool `json:"fieldCoverage"`
}

// CalculateCoverage calculates the coverage percentage
func (c *ContractCoverage) CalculateCoverage() {
	totalFields := len(c.CoveredFields) + len(c.UncoveredFields)
	if totalFields > 0 {
		c.CoveragePercent = float64(len(c.CoveredFields)) / float64(totalFields) * 100
	} else {
		c.CoveragePercent = 0
	}
}

// TrackFieldCoverage tracks which fields were covered in the response
func TrackFieldCoverage(expected, actual interface{}, path string, coverage *ContractCoverage) {
	// Make sure FieldCoverage is initialized
	if coverage.FieldCoverage == nil {
		coverage.FieldCoverage = make(map[string]bool)
	}

	switch exp := expected.(type) {
	case map[string]interface{}:
		// Handle object fields
		actMap, ok := actual.(map[string]interface{})
		if !ok {
			coverage.UncoveredFields = append(coverage.UncoveredFields, path)
			if path != "" {
				coverage.FieldCoverage[path] = false
			}
			return
		}

		for k, v := range exp {
			fieldPath := path
			if fieldPath == "" {
				fieldPath = k
			} else {
				fieldPath = fieldPath + "." + k
			}

			if actVal, exists := actMap[k]; exists {
				// Field exists in response
				coverage.CoveredFields = append(coverage.CoveredFields, fieldPath)
				coverage.FieldCoverage[fieldPath] = true
				// Recursively check nested fields
				TrackFieldCoverage(v, actVal, fieldPath, coverage)
			} else {
				coverage.UncoveredFields = append(coverage.UncoveredFields, fieldPath)
				coverage.FieldCoverage[fieldPath] = false
			}
		}

	case []interface{}:
		// Handle arrays
		actArr, ok := actual.([]interface{})
		if !ok || len(actArr) == 0 {
			coverage.UncoveredFields = append(coverage.UncoveredFields, path+"[]")
			coverage.FieldCoverage[path+"[]"] = false
			return
		}

		// Arrays are covered if they have at least one element
		coverage.CoveredFields = append(coverage.CoveredFields, path+"[]")
		coverage.FieldCoverage[path+"[]"] = true

		// If the expected array has a template item, check it against actual items
		if len(exp) > 0 {
			template := exp[0]
			for i, item := range actArr {
				// Limit to first few items to avoid excessive checks
				if i >= 3 {
					break
				}
				TrackFieldCoverage(template, item, fmt.Sprintf("%s[%d]", path, i), coverage)
			}
		}

	case string, float64, int, bool, nil:
		// For primitive types, just mark the field as covered
		if path != "" {
			coverage.CoveredFields = append(coverage.CoveredFields, path)
			coverage.FieldCoverage[path] = true
		}
	}
}
