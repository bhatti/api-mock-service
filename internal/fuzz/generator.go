package fuzz

import (
	log "github.com/sirupsen/logrus"
	"math/rand"
	"reflect"
	"strings"
)

// GenerateFuzzData using regex, data or types
func GenerateFuzzData(val any) any {
	if val == nil {
		return nil
	}
	switch val.(type) {
	case bool:
		return val //RandBool()
	case int:
		return val //RandNumMinMax(-1000, 10000)
	case int8:
		return val //RandNumMinMax(-1000, 10000)
	case int16:
		return val //RandNumMinMax(-1000, 10000)
	case int32:
		return val //RandNumMinMax(-1000, 10000)
	case int64:
		return val //RandNumMinMax(-1000, 10000)
	case uint:
		return val //RandNumMinMax(0, 10000)
	case uint8:
		return val //RandNumMinMax(0, 10000)
	case uint16:
		return val //RandNumMinMax(0, 10000)
	case uint32:
		return val //RandNumMinMax(0, 10000)
	case uint64:
		return val //RandNumMinMax(0, 10000)
	case float32:
		return val //rand.Float32()
	case float64:
		return val //rand.Float64()
	case string:
		strVal := val.(string)
		if strVal == NumberPrefixRegex {
			return float64(RandNumMinMax(0, 10000)) * rand.ExpFloat64()
		} else if strVal == BooleanPrefixRegex {
			return RandBool()
		} else if strVal == UintPrefixRegex {
			return RandNumMinMax(0, 10000)
		} else if strVal == IntPrefixRegex {
			return RandNumMinMax(-100, 10000)
		} else if strings.HasPrefix(strVal, "__") || strings.HasPrefix(strVal, "(") {
			return RandRegex(strVal)
		} else if strings.HasPrefix(strVal, "{{") {
			if out, err := ParseTemplate("", []byte(strVal), nil); err == nil {
				return string(out)
			}
			return RandSentence(1, 3)
		} else if strings.Contains(strVal, WildRegex) {
			return RandRegex(strVal)
			//return RandSentence(1, 3)
		} else {
			return strVal
		}
	case map[string]string:
		hm := val.(map[string]string)
		res := make(map[string]any)
		for k, v := range hm {
			res[k] = GenerateFuzzData(v)
		}
		return res
	case map[string]any:
		hm := val.(map[string]any)
		res := make(map[string]any)
		for k, v := range hm {
			res[k] = GenerateFuzzData(v)
		}
		return res
	case []string:
		arr := val.([]string)
		res := make([]string, len(arr))
		for i, v := range arr {
			res[i] = GenerateFuzzData(v).(string)
		}
		return res
	case []any:
		arr := val.([]any)
		res := make([]any, len(arr))
		for i, v := range arr {
			res[i] = GenerateFuzzData(v)
		}
		return res
	default:
		log.WithFields(log.Fields{
			"val":     val,
			"valType": reflect.TypeOf(val),
		}).Info("cannot populate unknown value type")
	}
	return val
}
