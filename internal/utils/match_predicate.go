package utils

import (
	"fmt"
	"github.com/bhatti/api-mock-service/internal/fuzz"
	"github.com/bhatti/api-mock-service/internal/types"
	log "github.com/sirupsen/logrus"
)

// MatchScenarioPredicate checks if predicate match
func MatchScenarioPredicate(matched *types.MockScenarioKeyData, target *types.MockScenarioKeyData, requestCount uint64) bool {
	if matched.Predicate == "" {
		return true
	}
	// Find any params for query params and path variables
	params := matched.MatchGroups(target.Path)
	for k, v := range matched.AssertQueryParamsPattern {
		params[k] = v
	}
	for k, v := range target.AssertQueryParamsPattern {
		params[k] = v
	}
	params[fuzz.RequestCount] = fmt.Sprintf("%d", requestCount)
	out, err := fuzz.ParseTemplate("", []byte(matched.Predicate), params)
	log.WithFields(log.Fields{
		"Path":          matched.Path,
		"Name":          matched.Name,
		"Method":        matched.Method,
		"RequestCount":  requestCount,
		"Timestamp":     matched.LastUsageTime,
		"MatchedOutput": string(out),
		"Error":         err,
	}).Debugf("matching predicate...")

	return err != nil || string(out) == "true"
}
