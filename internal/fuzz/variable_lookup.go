package fuzz

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"regexp"
	"strconv"
	"strings"
)

// FindVariable extracts variable property
func FindVariable(name string, data any) any {
	n := strings.Index(name, ".")
	var nextName string
	if n != -1 {
		nextName = name[n+1:]
		name = name[0:n]
	}

	switch data.(type) {
	case map[string]string:
		params := data.(map[string]string)
		val := params[name]
		if val == "" {
			return nil
		}
		if nextName == "" {
			return val
		}
		return FindVariable(nextName, val)
	case map[string]any:
		params := data.(map[string]any)
		val := params[name]
		if val == nil {
			return nil
		}
		if nextName == "" {
			return val
		}
		return FindVariable(nextName, val)
	case map[string][]any:
		params := data.(map[string][]any)
		val := params[name]
		if val == nil {
			return nil
		}
		if nextName == "" {
			return val
		}
		var all []any
		for _, param := range val {
			res := FindVariable(nextName, param)
			if res != nil {
				all = append(all, res)
			}
		}
		return all
	case []any:
		params := data.([]any)
		var all []any
		for _, param := range params {
			res := FindVariable(name, param)
			if res != nil {
				all = append(all, res)
			}
		}
		return all
	default:
		return nil
	}
}

// VariableSize finds size of the variable
func VariableSize(name string, data any) int {
	val := FindVariable(name, data)
	if val == nil {
		return -1
	}
	switch val.(type) {
	case map[string]string:
		return len(val.(map[string]string))
	case map[string]any:
		return len(val.(map[string]any))
	case []any:
		return len(val.([]any))
	case []int:
		return len(val.([]int))
	case []string:
		return len(val.([]string))
	case []float64:
		return len(val.([]float64))
	default:
		return -1
	}
}

// VariableContains checks if variable contains value
func VariableContains(name string, target any, data any) bool {
	val := FindVariable(name, data)
	if val == nil {
		return false
	}
	valStr := fmt.Sprintf("%v", val)
	reStr := fmt.Sprintf("%v", target)
	re, err := regexp.Compile(reStr)
	if err != nil {
		log.WithFields(log.Fields{
			"Name":   name,
			"Target": target,
			"Regex":  reStr,
			"Error":  err,
		}).Warnf("failed to compile regex")
		return false
	}
	return re.MatchString(valStr)
}

// VariableNumber returns numeric value for variable
func VariableNumber(name string, data any) float64 {
	val := FindVariable(name, data)
	if n, err := strconv.ParseFloat(fmt.Sprintf("%v", val), 64); err == nil {
		return n
	}
	return 0
}

// VariableEquals checks if variable is equal to the value
func VariableEquals(name string, data any, target any) bool {
	val := FindVariable(name, data)
	if val == nil {
		return false
	}
	return fmt.Sprintf("%v", val) == fmt.Sprintf("%v", target)
}
