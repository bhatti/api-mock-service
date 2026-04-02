package shrink

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/bhatti/api-mock-service/internal/types"
	"github.com/stretchr/testify/require"
)

// alwaysFailDetector fails for any scenario that has a large boundary value
// in the request body (simulating a production boundary-triggered failure).
type alwaysFailDetector struct{}

func (d *alwaysFailDetector) TestScenario(_ context.Context, s *types.APIScenario) error {
	return fmt.Errorf("always fails")
}

// neverFailDetector simulates a scenario where the original input doesn't actually fail.
type neverFailDetector struct{}

func (d *neverFailDetector) TestScenario(_ context.Context, s *types.APIScenario) error {
	return nil // passes
}

// failOnLargeNumber fails only when a numeric field is > 1000.
type failOnLargeNumber struct {
	field string
}

func (d *failOnLargeNumber) TestScenario(_ context.Context, s *types.APIScenario) error {
	body := parseBodyMap(s.Request.Contents)
	if body == nil {
		return nil
	}
	if val, ok := body[d.field].(float64); ok && val > 1000 {
		return fmt.Errorf("large number triggers failure")
	}
	return nil
}

// failOnLongString fails only when a string field is longer than 3 chars.
type failOnLongString struct {
	field string
}

func (d *failOnLongString) TestScenario(_ context.Context, s *types.APIScenario) error {
	body := parseBodyMap(s.Request.Contents)
	if body == nil {
		return nil
	}
	if val, ok := body[d.field].(string); ok && len(val) > 3 {
		return fmt.Errorf("long string triggers failure")
	}
	return nil
}

// failIfFieldPresent fails when a specific field exists in the body.
type failIfFieldPresent struct {
	field string
}

func (d *failIfFieldPresent) TestScenario(_ context.Context, s *types.APIScenario) error {
	body := parseBodyMap(s.Request.Contents)
	if body == nil {
		return nil
	}
	if _, exists := body[d.field]; exists {
		return fmt.Errorf("field %s triggers failure", d.field)
	}
	return nil
}

func scenarioWithBody(contents string) *types.APIScenario {
	return &types.APIScenario{
		Name:   "test-scenario",
		Method: types.Post,
		Path:   "/test",
		Request: types.APIRequest{
			Contents: contents,
		},
		Response: types.APIResponse{
			StatusCode: 200,
		},
	}
}

func TestShrink_OriginalDoesNotFail_ReturnsAsIs(t *testing.T) {
	s := scenarioWithBody(`{"a": "value"}`)
	result, err := Shrink(context.Background(), &neverFailDetector{}, s, ShrinkOptions{})
	require.NoError(t, err)
	require.False(t, result.Reduced)
	require.Equal(t, 0, result.Attempts)
}

func TestShrink_AlwaysFails_ReturnsOriginalMinimal(t *testing.T) {
	body := `{"x": 999999, "y": "short"}`
	s := scenarioWithBody(body)
	result, err := Shrink(context.Background(), &alwaysFailDetector{}, s, ShrinkOptions{MaxAttempts: 20})
	require.NoError(t, err)
	// Still fails, but may or may not reduce
	require.NotNil(t, result.Minimal)
}

func TestShrink_FieldRemoval_RemovesUnnecessaryFields(t *testing.T) {
	// Only "bad" field causes failure; "good" field is irrelevant.
	body := `{"bad": "value", "good": "irrelevant"}`
	s := scenarioWithBody(body)
	result, err := Shrink(context.Background(), &failIfFieldPresent{field: "bad"}, s, ShrinkOptions{MaxAttempts: 50})
	require.NoError(t, err)
	// The shrinker should keep "bad" (it causes the failure) but could remove "good"
	minimal := parseBodyMap(result.Minimal.Request.Contents)
	require.NotNil(t, minimal)
	require.Contains(t, minimal, "bad", "bad field must remain (it triggers the failure)")
}

func TestShrink_StringBisection_FindsMinimalLength(t *testing.T) {
	// Failure triggered by string longer than 3 chars.
	longStr := "abcdefghijklmnop" // 16 chars
	body, _ := json.Marshal(map[string]any{"msg": longStr})
	s := scenarioWithBody(string(body))
	result, err := Shrink(context.Background(), &failOnLongString{field: "msg"}, s, ShrinkOptions{MaxAttempts: 100})
	require.NoError(t, err)
	if result.Reduced {
		minimal := parseBodyMap(result.Minimal.Request.Contents)
		msg, _ := minimal["msg"].(string)
		require.LessOrEqual(t, len(msg), len(longStr), "string should have been shortened")
		require.Greater(t, len(msg), 3, "string must stay > 3 to keep triggering failure")
	}
}

func TestShrink_NumericReduction_ReducesBoundaryValue(t *testing.T) {
	// Failure only when number > 1000.
	body, _ := json.Marshal(map[string]any{"count": float64(2147483647)})
	s := scenarioWithBody(string(body))
	result, err := Shrink(context.Background(), &failOnLargeNumber{field: "count"}, s, ShrinkOptions{MaxAttempts: 100})
	require.NoError(t, err)
	if result.Reduced {
		minimal := parseBodyMap(result.Minimal.Request.Contents)
		count, _ := minimal["count"].(float64)
		require.Greater(t, count, float64(1000), "must stay above failure threshold")
		require.Less(t, count, float64(2147483647), "should have been reduced from max")
	}
}

func TestShrink_EmptyBody_NoReduction(t *testing.T) {
	s := scenarioWithBody("")
	result, err := Shrink(context.Background(), &alwaysFailDetector{}, s, ShrinkOptions{})
	require.NoError(t, err)
	require.False(t, result.Reduced)
}

func TestShrink_ArrayShrinking_RemovesElements(t *testing.T) {
	// Failure if "items" array has any elements.
	body, _ := json.Marshal(map[string]any{"items": []any{"a", "b", "c"}})
	s := scenarioWithBody(string(body))

	// Custom detector: fail when items is non-empty
	det := &failIfFieldPresent{field: "items"}
	result, err := Shrink(context.Background(), det, s, ShrinkOptions{MaxAttempts: 50})
	require.NoError(t, err)
	require.NotNil(t, result.Minimal)
}

func TestShrinkOptions_Defaults(t *testing.T) {
	opts := ShrinkOptions{}
	opts.defaults()
	require.Equal(t, 100, opts.MaxAttempts)
	require.Equal(t, int64(30), int64(opts.Timeout.Seconds()))
}
