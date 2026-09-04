package contract

import (
	"crypto/sha256"
	"fmt"
	"sort"
	"strings"

	"github.com/bhatti/api-mock-service/internal/types"
)

// DeduplicateFailures groups mutation failures by (statusCode, errorCategory, endpoint, responseHash)
// to reduce noise from repeated root causes. Returns clusters sorted by count descending.
func DeduplicateFailures(
	errors map[string]string,
	details map[string]*types.ContractValidationDetail,
) []types.FailureCluster {
	if len(errors) == 0 {
		return nil
	}

	clusters := make(map[types.FailureClusterKey]*types.FailureCluster)

	for mutationName, errMsg := range errors {
		detail := details[mutationName]

		key := types.FailureClusterKey{
			StatusCode:    extractStatusCode(detail),
			ErrorCategory: classifyError(errMsg, detail),
			Endpoint:      extractEndpoint(detail),
			ResponseHash:  hashPrefix(errMsg, 512),
		}

		cluster, exists := clusters[key]
		if !exists {
			cluster = &types.FailureCluster{
				Key:            key,
				Representative: mutationName,
				Severity:       classifyClusterSeverity(detail),
			}
			clusters[key] = cluster
		}
		cluster.Count++
		cluster.AffectedMutations = append(cluster.AffectedMutations, mutationName)
	}

	result := make([]types.FailureCluster, 0, len(clusters))
	for _, c := range clusters {
		result = append(result, *c)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Count > result[j].Count
	})
	return result
}

func extractStatusCode(detail *types.ContractValidationDetail) int {
	if detail == nil {
		return 0
	}
	return detail.StatusCode
}

func classifyError(errMsg string, detail *types.ContractValidationDetail) string {
	if detail != nil && len(detail.InjectionFindings) > 0 {
		return "injection"
	}
	if detail != nil && len(detail.SchemaViolations) > 0 {
		return "schema-violation"
	}
	lower := strings.ToLower(errMsg)
	if strings.Contains(lower, "status") && strings.Contains(lower, "didn't match") {
		return "status-mismatch"
	}
	if strings.Contains(lower, "assertion") || strings.Contains(lower, "assert") {
		return "assertion"
	}
	return "other"
}

func extractEndpoint(detail *types.ContractValidationDetail) string {
	if detail == nil {
		return ""
	}
	return detail.URL
}

func hashPrefix(s string, maxLen int) string {
	if len(s) > maxLen {
		s = s[:maxLen]
	}
	h := sha256.Sum256([]byte(s))
	return fmt.Sprintf("%x", h[:8])
}

func classifyClusterSeverity(detail *types.ContractValidationDetail) string {
	if detail != nil && len(detail.InjectionFindings) > 0 {
		best := "info"
		order := map[string]int{"critical": 3, "high": 2, "medium": 1, "info": 0}
		for _, f := range detail.InjectionFindings {
			if order[f.Severity] > order[best] {
				best = f.Severity
			}
		}
		return best
	}
	if detail != nil && len(detail.SchemaViolations) > 0 {
		return "medium"
	}
	return "high"
}
