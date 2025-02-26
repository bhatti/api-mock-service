package fuzz

import (
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

// extractJsonPath extracts a value using a JSON path expression
func extractJsonPath(path string, data any) any {
	if data == nil {
		return nil
	}

	segments := strings.Split(path, ".")
	current := data

	for _, segment := range segments {
		// Handle array indexing with [n] syntax
		arrayIndexMatch := regexp.MustCompile(`^(.*)\[(\d+)\]$`).FindStringSubmatch(segment)
		if len(arrayIndexMatch) == 3 {
			// We have an array index
			fieldName := arrayIndexMatch[1]
			indexStr := arrayIndexMatch[2]
			index, _ := strconv.Atoi(indexStr)

			if fieldName != "" {
				// Get the array field first
				current = getField(current, fieldName)
				if current == nil {
					return nil
				}
			}

			// Access the array element
			arr, ok := toArray(current)
			if !ok || index >= len(arr) {
				return nil
			}
			current = arr[index]
		} else {
			// Regular field access
			current = getField(current, segment)
			if current == nil {
				return nil
			}
		}
	}

	return current
}

// getField gets a field from a map or struct
func getField(data any, field string) any {
	switch v := data.(type) {
	case map[string]any:
		return v[field]
	case map[string]string:
		val, ok := v[field]
		if !ok {
			return nil
		}
		return val
	default:
		// Try using reflection for structs
		val := reflect.ValueOf(data)
		if val.Kind() == reflect.Ptr {
			val = val.Elem()
		}
		if val.Kind() != reflect.Struct {
			return nil
		}

		f := val.FieldByName(field)
		if !f.IsValid() {
			return nil
		}
		return f.Interface()
	}
}

// toArray converts a value to an array if possible
func toArray(data any) ([]any, bool) {
	switch v := data.(type) {
	case []any:
		return v, true
	default:
		// Try using reflection
		val := reflect.ValueOf(data)
		if val.Kind() == reflect.Slice || val.Kind() == reflect.Array {
			result := make([]any, val.Len())
			for i := 0; i < val.Len(); i++ {
				result[i] = val.Index(i).Interface()
			}
			return result, true
		}
		return nil, false
	}
}
