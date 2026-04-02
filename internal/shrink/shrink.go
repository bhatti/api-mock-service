// SPDX-License-Identifier: MIT

// Package shrink reduces a failing mutation test case to its minimal reproducing payload.
// When a mutation test fails, the failing input is the full mutated payload — shrinking
// bisects it to the smallest set of changes that still triggers the failure.
package shrink

import (
	"context"
	"encoding/json"
	"math"
	"time"

	"github.com/bhatti/api-mock-service/internal/types"
	log "github.com/sirupsen/logrus"
)

// FailureDetector can test a single scenario and report whether it fails.
// Return nil for "this scenario passes", non-nil for "this scenario fails".
type FailureDetector interface {
	TestScenario(ctx context.Context, scenario *types.APIScenario) error
}

// ShrinkOptions controls how hard the shrinker works.
type ShrinkOptions struct {
	// MaxAttempts is the total number of candidate scenarios to try. Default 100.
	MaxAttempts int
	// Timeout caps the total wall-clock time. Default 30s.
	Timeout time.Duration
}

func (o *ShrinkOptions) defaults() {
	if o.MaxAttempts <= 0 {
		o.MaxAttempts = 100
	}
	if o.Timeout <= 0 {
		o.Timeout = 30 * time.Second
	}
}

// Result holds the minimal scenario found by Shrink.
type Result struct {
	// Minimal is the smallest scenario that still triggers the failure.
	Minimal *types.APIScenario
	// Attempts is the number of candidate scenarios evaluated.
	Attempts int
	// Reduced reports whether any reduction was achieved.
	Reduced bool
}

// Shrink takes a known-failing scenario and returns the minimal scenario that
// still fails according to detector. It tries four reduction strategies in order:
//  1. Field removal (delta debugging)
//  2. String shortening (binary search on length)
//  3. Array element removal
//  4. Numeric reduction (exponential backoff from boundary values)
func Shrink(
	ctx context.Context,
	detector FailureDetector,
	failing *types.APIScenario,
	opts ShrinkOptions,
) (*Result, error) {
	opts.defaults()

	deadline := time.Now().Add(opts.Timeout)
	attempts := 0

	current := cloneScenario(failing)
	reduced := false

	// Confirm the original actually fails
	if err := detector.TestScenario(ctx, current); err == nil {
		log.WithFields(log.Fields{"Component": "Shrink"}).
			Warn("original scenario does not fail — nothing to shrink")
		return &Result{Minimal: current, Attempts: 0, Reduced: false}, nil
	}

	strategies := []func(context.Context, *types.APIScenario, FailureDetector, *int, int, time.Time) (*types.APIScenario, bool){
		shrinkFields,
		shrinkStrings,
		shrinkArrays,
		shrinkNumerics,
	}

	for _, strategy := range strategies {
		if time.Now().After(deadline) || attempts >= opts.MaxAttempts {
			break
		}
		candidate, didReduce := strategy(ctx, current, detector, &attempts, opts.MaxAttempts, deadline)
		if didReduce {
			current = candidate
			reduced = true
		}
	}

	return &Result{Minimal: current, Attempts: attempts, Reduced: reduced}, nil
}

// shrinkFields removes request body fields one at a time (delta debugging).
func shrinkFields(ctx context.Context, scenario *types.APIScenario, det FailureDetector,
	attempts *int, maxAttempts int, deadline time.Time) (*types.APIScenario, bool) {

	body := parseBodyMap(scenario.Request.Contents)
	if len(body) == 0 {
		return scenario, false
	}

	current := cloneScenario(scenario)
	reduced := false

	for field := range body {
		if time.Now().After(deadline) || *attempts >= maxAttempts {
			break
		}
		candidate := cloneScenario(current)
		candidateBody := parseBodyMap(candidate.Request.Contents)
		delete(candidateBody, field)
		candidate.Request.Contents = marshalBody(candidateBody)

		*attempts++
		if det.TestScenario(ctx, candidate) != nil {
			// Still fails without this field — keep the removal
			current = candidate
			delete(body, field)
			reduced = true
			log.WithFields(log.Fields{"Component": "Shrink", "RemovedField": field}).
				Debug("field removal kept")
		}
	}
	return current, reduced
}

// shrinkStrings binary-searches string fields for the minimal triggering length.
func shrinkStrings(ctx context.Context, scenario *types.APIScenario, det FailureDetector,
	attempts *int, maxAttempts int, deadline time.Time) (*types.APIScenario, bool) {

	body := parseBodyMap(scenario.Request.Contents)
	if len(body) == 0 {
		return scenario, false
	}

	current := cloneScenario(scenario)
	reduced := false

	for field, val := range body {
		str, ok := val.(string)
		if !ok || len(str) <= 1 {
			continue
		}

		lo, hi := 0, len(str)
		for lo < hi && *attempts < maxAttempts && !time.Now().After(deadline) {
			mid := (lo + hi) / 2
			candidate := cloneScenario(current)
			candidateBody := parseBodyMap(candidate.Request.Contents)
			candidateBody[field] = str[:mid]
			candidate.Request.Contents = marshalBody(candidateBody)

			*attempts++
			if det.TestScenario(ctx, candidate) != nil {
				// Shorter string still fails
				hi = mid
				current = candidate
				body[field] = str[:mid]
				str = str[:mid]
				reduced = true
			} else {
				lo = mid + 1
			}
		}
	}
	return current, reduced
}

// shrinkArrays removes array elements one at a time.
func shrinkArrays(ctx context.Context, scenario *types.APIScenario, det FailureDetector,
	attempts *int, maxAttempts int, deadline time.Time) (*types.APIScenario, bool) {

	body := parseBodyMap(scenario.Request.Contents)
	if len(body) == 0 {
		return scenario, false
	}

	current := cloneScenario(scenario)
	reduced := false

	for field, val := range body {
		arr, ok := val.([]any)
		if !ok || len(arr) <= 1 {
			continue
		}

		for i := len(arr) - 1; i >= 0; i-- {
			if time.Now().After(deadline) || *attempts >= maxAttempts {
				break
			}
			candidate := cloneScenario(current)
			candidateBody := parseBodyMap(candidate.Request.Contents)
			shortened := make([]any, 0, len(arr)-1)
			shortened = append(shortened, arr[:i]...)
			shortened = append(shortened, arr[i+1:]...)
			candidateBody[field] = shortened
			candidate.Request.Contents = marshalBody(candidateBody)

			*attempts++
			if det.TestScenario(ctx, candidate) != nil {
				current = candidate
				arr = shortened
				body[field] = shortened
				reduced = true
			}
		}
	}
	return current, reduced
}

// shrinkNumerics exponentially reduces large boundary values toward zero.
func shrinkNumerics(ctx context.Context, scenario *types.APIScenario, det FailureDetector,
	attempts *int, maxAttempts int, deadline time.Time) (*types.APIScenario, bool) {

	body := parseBodyMap(scenario.Request.Contents)
	if len(body) == 0 {
		return scenario, false
	}

	current := cloneScenario(scenario)
	reduced := false

	for field, val := range body {
		var num float64
		switch v := val.(type) {
		case float64:
			num = v
		case int:
			num = float64(v)
		default:
			continue
		}

		if math.Abs(num) < 1000 {
			continue // Not a boundary value; skip
		}

		// Exponential backoff toward zero
		for math.Abs(num) > 1 && *attempts < maxAttempts && !time.Now().After(deadline) {
			num /= 2
			candidate := cloneScenario(current)
			candidateBody := parseBodyMap(candidate.Request.Contents)
			candidateBody[field] = num
			candidate.Request.Contents = marshalBody(candidateBody)

			*attempts++
			if det.TestScenario(ctx, candidate) != nil {
				current = candidate
				body[field] = num
				reduced = true
			} else {
				break // Can't reduce further — stop
			}
		}
	}
	return current, reduced
}

// --- helpers ---

func cloneScenario(s *types.APIScenario) *types.APIScenario {
	clone := *s
	// Deep-copy APIRequest map fields so mutations don't affect the original
	req := s.Request
	if s.Request.PathParams != nil {
		req.PathParams = make(map[string]string, len(s.Request.PathParams))
		for k, v := range s.Request.PathParams {
			req.PathParams[k] = v
		}
	}
	if s.Request.QueryParams != nil {
		req.QueryParams = make(map[string]string, len(s.Request.QueryParams))
		for k, v := range s.Request.QueryParams {
			req.QueryParams[k] = v
		}
	}
	if s.Request.PostParams != nil {
		req.PostParams = make(map[string]string, len(s.Request.PostParams))
		for k, v := range s.Request.PostParams {
			req.PostParams[k] = v
		}
	}
	if s.Request.Headers != nil {
		req.Headers = make(map[string]string, len(s.Request.Headers))
		for k, v := range s.Request.Headers {
			req.Headers[k] = v
		}
	}
	clone.Request = req
	// Deep-copy APIResponse header map
	resp := s.Response
	if s.Response.Headers != nil {
		resp.Headers = s.Response.Headers.Clone()
	}
	clone.Response = resp
	return &clone
}

func parseBodyMap(contents string) map[string]any {
	if contents == "" {
		return nil
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(contents), &m); err != nil {
		return nil
	}
	return m
}

func marshalBody(m map[string]any) string {
	if m == nil {
		return ""
	}
	b, err := json.Marshal(m)
	if err != nil {
		return ""
	}
	return string(b)
}
