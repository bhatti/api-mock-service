package contract

import (
	"fmt"
	"reflect"
	"regexp"
	"runtime/debug"
	"strings"
)

// compareObjects recursively compares objects and records differences
func compareObjects(expected, actual map[string]interface{}, path string, report *ContractDiffReport) {
	// Check for missing fields
	for k, expectedValue := range expected {
		fieldPath := joinPath(path, k)

		actualValue, exists := actual[k]
		if !exists {
			report.MissingFields = append(report.MissingFields, fieldPath)
			continue
		}

		// Check for type mismatches
		expectedType := getTypeName(expectedValue)
		actualType := getTypeName(actualValue)

		if expectedType != actualType {
			report.TypeMismatches[fieldPath] = fmt.Sprintf("expected %s, got %s",
				expectedType, actualType)
			continue
		}

		// Recursive comparison for nested objects
		if expectedType == "object" {
			if expectedObj, ok := expectedValue.(map[string]interface{}); ok {
				if actualObj, ok := actualValue.(map[string]interface{}); ok {
					compareObjects(expectedObj, actualObj, fieldPath, report)
				}
			}
			continue
		}

		// Handle arrays
		if expectedType == "array" {
			if expectedArr, ok := expectedValue.([]interface{}); ok {
				if actualArr, ok := actualValue.([]interface{}); ok {
					compareArrays(expectedArr, actualArr, fieldPath, report)
				}
			}
			continue
		}

		// For primitive types, check exact equality (except for regex patterns)
		if expectedStr, ok := expectedValue.(string); ok {
			// Check if it's a regex pattern
			if strings.HasPrefix(expectedStr, "(__") && strings.HasSuffix(expectedStr, "__)") {
				// It's a type pattern, handled by type comparison above
				continue
			}

			// If it starts with "(" and contains "|", it might be a regex choice pattern
			if strings.HasPrefix(expectedStr, "(") && strings.Contains(expectedStr, "|") {
				// Try to match against the pattern
				re, err := regexp.Compile(expectedStr)
				if err == nil {
					actualStr := fmt.Sprintf("%v", actualValue)
					if !re.MatchString(actualStr) {
						report.ValueMismatches[fieldPath] = ValueMismatch{
							Expected: expectedStr,
							Actual:   actualStr,
						}
					}
					continue
				}
			}
		}

		// Regular value comparison
		if !reflect.DeepEqual(expectedValue, actualValue) {
			report.ValueMismatches[fieldPath] = ValueMismatch{
				Expected: expectedValue,
				Actual:   actualValue,
			}
		}
	}

	// Check for extra fields
	for k := range actual {
		fieldPath := joinPath(path, k)
		if _, exists := expected[k]; !exists {
			report.ExtraFields = append(report.ExtraFields, fieldPath)
		}
	}
}

// compareArrays compares arrays and records differences
func compareArrays(expected, actual []interface{}, path string, report *ContractDiffReport) {
	// If expected array is empty, nothing to compare
	if len(expected) == 0 {
		return
	}

	// If actual array is empty but expected isn't
	if len(actual) == 0 {
		report.ValueMismatches[path] = ValueMismatch{
			Expected: expected,
			Actual:   actual,
		}
		return
	}

	// If expected has only one item, use it as a template for all actual items
	if len(expected) == 1 {
		template := expected[0]

		// Check if the template is an object
		if templateObj, ok := template.(map[string]interface{}); ok {
			for i, actualItem := range actual {
				if actualObj, ok := actualItem.(map[string]interface{}); ok {
					itemPath := fmt.Sprintf("%s[%d]", path, i)
					compareObjects(templateObj, actualObj, itemPath, report)
				} else {
					debug.PrintStack()
					report.TypeMismatches[fmt.Sprintf("%s[%d]", path, i)] =
						fmt.Sprintf("compareArrays expected object %v, got %T", templateObj, actualItem)
				}
			}
			return
		}
	}

	// Otherwise, compare items by index up to the length of the shorter array
	minLen := len(expected)
	if len(actual) < minLen {
		minLen = len(actual)
	}

	for i := 0; i < minLen; i++ {
		itemPath := fmt.Sprintf("%s[%d]", path, i)

		expectedItem := expected[i]
		actualItem := actual[i]

		expectedType := getTypeName(expectedItem)
		actualType := getTypeName(actualItem)

		if expectedType != actualType {
			report.TypeMismatches[itemPath] = fmt.Sprintf("expected %s, got %s",
				expectedType, actualType)
			continue
		}

		if expectedType == "object" {
			if expectedObj, ok := expectedItem.(map[string]interface{}); ok {
				if actualObj, ok := actualItem.(map[string]interface{}); ok {
					compareObjects(expectedObj, actualObj, itemPath, report)
				}
			}
		} else if !reflect.DeepEqual(expectedItem, actualItem) {
			report.ValueMismatches[itemPath] = ValueMismatch{
				Expected: expectedItem,
				Actual:   actualItem,
			}
		}
	}

	// Check for length mismatch
	if len(expected) != len(actual) {
		report.ValueMismatches[path+".length"] = ValueMismatch{
			Expected: len(expected),
			Actual:   len(actual),
		}
	}
}

// Helper functions for field comparison
func joinPath(base, field string) string {
	if base == "" {
		return field
	}
	return base + "." + field
}

func getTypeName(value interface{}) string {
	if value == nil {
		return "null"
	}

	switch v := value.(type) {
	case map[string]interface{}:
		return "object"
	case []interface{}:
		return "array"
	case string:
		// Check if it's a type pattern
		if strings.HasPrefix(v, "(__") && strings.HasSuffix(v, "__)") {
			typeName := v[3 : len(v)-3]
			if typeName == "string" || typeName == "number" || typeName == "boolean" {
				return typeName
			}
		}
		return "string"
	case float64, int, int64:
		return "number"
	case bool:
		return "boolean"
	default:
		return fmt.Sprintf("%T", value)
	}
}
