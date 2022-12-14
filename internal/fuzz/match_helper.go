package fuzz

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"reflect"
	"regexp"
	"strings"
	"unicode"

	log "github.com/sirupsen/logrus"
)

// PrefixTypeNumber type
const PrefixTypeNumber = "__number__"

// PrefixTypeBoolean type
const PrefixTypeBoolean = "__boolean__"

// PrefixTypeString type
const PrefixTypeString = "__string__"

// PrefixTypeObject type
const PrefixTypeObject = "__object__"

// UintPrefixRegex constant
const UintPrefixRegex = PrefixTypeNumber + `[0-9]{1,10}`

// IntPrefixRegex constant
const IntPrefixRegex = PrefixTypeNumber + `[+-]?[0-9]{1,10}`

// NumberPrefixRegex constant
const NumberPrefixRegex = PrefixTypeNumber + `[+-]?(([0-9]{1,10}(\.[0-9]{1,5})?)|(\.[0-9]{1,10}))`

// BooleanPrefixRegex constant
const BooleanPrefixRegex = PrefixTypeBoolean + `(false|true)`

// EmailRegex constant
const EmailRegex = `\w+@\w+.?\w+`

// AnyWordRegex  constant
const AnyWordRegex = `\w+`

// WildRegex  constant
const WildRegex = `.+`

// StripTypeTags removes type prefixes
func StripTypeTags(re string) string {
	return regexp.MustCompile(
		fmt.Sprintf("(%s|%s|%s|%s)", PrefixTypeNumber, PrefixTypeBoolean, PrefixTypeString, PrefixTypeObject)).
		ReplaceAllString(re, "")
}

// PopulateRandomData using regex or data
func PopulateRandomData(val any) any {
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
		} else if strings.Contains(strVal, WildRegex) {
			return RandSentence(1, 3)
		} else if strings.HasPrefix(strVal, "__") || strings.HasPrefix(strVal, "(") {
			return RandRegex(strVal)
		} else {
			return strVal
		}
	case map[string]string:
		hm := val.(map[string]string)
		res := make(map[string]any)
		for k, v := range hm {
			res[k] = PopulateRandomData(v)
		}
		return res
	case map[string]any:
		hm := val.(map[string]any)
		res := make(map[string]any)
		for k, v := range hm {
			res[k] = PopulateRandomData(v)
		}
		return res
	case []any:
		arr := val.([]any)
		res := make([]any, len(arr))
		for i, v := range arr {
			res[i] = PopulateRandomData(v)
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

// FlatRegexMap to add all regex in same map
func FlatRegexMap(val any) map[string]string {
	regex := make(map[string]string)
	flatRegexMap(val, regex, "")
	for k, v := range regex {
		if len(v) > 128 {
			log.WithFields(log.Fields{
				"Key": k,
				"Val": v,
			}).Debugf("simplifying regex")
			regex[k] = WildRegex // simplify really long regex
		}
	}
	return regex
}

// ValidateRegexMap validate data against regex map
func ValidateRegexMap(val any, regex map[string]string) error {
	return validateRegexMap(val, regex, "")
}

// UnmarshalArrayOrObjectAndExtractTypes helper method to unmarshal, add types and marshal again
func UnmarshalArrayOrObjectAndExtractTypes(str string, dataTemplate DataTemplateRequest) (map[string]string, error) {
	res, err := UnmarshalArrayOrObject([]byte(str))
	if err != nil {
		return nil, err
	}
	res = ExtractTypes(res, dataTemplate)
	return FlatRegexMap(res), nil
}

// UnmarshalArrayOrObjectAndExtractTypesAndMarshal helper method to unmarshal, add types and marshal again
func UnmarshalArrayOrObjectAndExtractTypesAndMarshal(str string, dataTemplate DataTemplateRequest) (string, error) {
	res, err := UnmarshalArrayOrObjectAndExtractTypes(str, dataTemplate)
	b, err := json.Marshal(res)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// UnmarshalArrayOrObject helper function to unmarshal bytes based on object/array
func UnmarshalArrayOrObject(b []byte) (res any, err error) {
	if len(b) == 0 {
		return nil, nil
	}
	str := strings.TrimSpace(string(b))

	if strings.HasPrefix(str, "{") {
		res = make(map[string]any)
		if err = json.Unmarshal(b, &res); err != nil {
			return nil, err
		}
	} else if strings.HasPrefix(str, "[") {
		res = make([]any, 0)
		if err = json.Unmarshal(b, &res); err != nil {
			return nil, err
		}
	}
	return
}

// PrefixTypeStringToRegEx helper method for adding prefix and regex
func PrefixTypeStringToRegEx(val string, dataTemplate DataTemplateRequest) string {
	return PrefixTypeString + ValueToRegEx(val, dataTemplate)
}

// ValueToRegEx value to regex
func ValueToRegEx(val string, dataTemplate DataTemplateRequest) string {
	var sb strings.Builder
	negative := false
	digits := 0
	fraction := 0
	alphabets := 0
	decimal := false
	specialChars := map[rune]bool{'@': true, '%': true, '(': true, ')': true,
		'#': true, '$': true, '*': true}
	for _, c := range []rune(val) {
		if unicode.IsDigit(c) {
			addAlphabets(alphabets, &sb, dataTemplate)
			alphabets = 0
			if decimal {
				fraction++
			} else {
				digits++
			}
		} else if c == '.' {
			if digits > 0 {
				decimal = true
			} else {
				addAlphabets(alphabets, &sb, dataTemplate)
				alphabets = 0
				sb.WriteString(`.?`)
			}
		} else if specialChars[c] {
			addAlphabets(alphabets, &sb, dataTemplate)
			addDigits(digits, fraction, negative, decimal, &sb, dataTemplate)
			alphabets = 0
			digits = 0
			fraction = 0
			negative = false
			decimal = false
			sb.WriteRune(c)
		} else if c == '-' {
			if digits == 0 && alphabets == 0 {
				negative = true
			} else {
				if digits > 0 {
					addDigits(digits, fraction, negative, decimal, &sb, dataTemplate)
					digits = 0
					fraction = 0
					negative = false
					decimal = false
				} else if alphabets > 0 {
					addAlphabets(alphabets, &sb, dataTemplate)
					alphabets = 0
				}
				sb.WriteString(`[-]`)
			}
		} else {
			addDigits(digits, fraction, negative, decimal, &sb, dataTemplate)
			digits = 0
			fraction = 0
			negative = false
			decimal = false
			alphabets++
		}
	}
	addAlphabets(alphabets, &sb, dataTemplate)
	addDigits(digits, fraction, negative, decimal, &sb, dataTemplate)
	return sb.String()
}

// ExtractTypes extract types from any structure
func ExtractTypes(val any, dataTemplate DataTemplateRequest) any {
	if val == nil {
		return nil
	}
	switch val.(type) {
	case bool:
		return BooleanPrefixRegex
	case int:
		return IntPrefixRegex
	case int8:
		return IntPrefixRegex
	case int16:
		return IntPrefixRegex
	case int32:
		return IntPrefixRegex
	case int64:
		return IntPrefixRegex
	case uint:
		return UintPrefixRegex
	case uint8:
		return UintPrefixRegex
	case uint16:
		return UintPrefixRegex
	case uint32:
		return UintPrefixRegex
	case uint64:
		return UintPrefixRegex
	case float32:
		return NumberPrefixRegex
	case float64:
		f := val.(float64)
		if f == float64(int64(f)) {
			return IntPrefixRegex
		}
		return NumberPrefixRegex
	case string:
		strVal := val.(string)
		if strVal == "" {
			return PrefixTypeString + fmt.Sprintf(`[a-z]{%d,%d}`, 1, dataTemplate.MinMultiplier*10)
		}
		return PrefixTypeStringToRegEx(strVal, dataTemplate)
	case map[string]string:
		hm := val.(map[string]string)
		res := make(map[string]string)
		for k, v := range hm {
			res[k] = v
		}
		return res
	case map[string]any:
		hm := val.(map[string]any)
		res := make(map[string]any)
		for k, v := range hm {
			res[k] = ExtractTypes(v, dataTemplate)
		}
		return res
	case []any:
		arr := val.([]any)
		res := make([]any, len(arr))
		for i, v := range arr {
			res[i] = ExtractTypes(v, dataTemplate)
		}
		return res
	default:
		log.WithFields(log.Fields{
			"val":     val,
			"valType": reflect.TypeOf(val),
		}).Info("cannot extract unknown value type")
	}
	return nil
}

func validateRegexMap(val any, regex map[string]string, prefix string) error {
	if val == nil {
		return nil
	}
	switch val.(type) {
	case map[string]string:
		hm := val.(map[string]string)
		for k, v := range hm {
			fullKey := buildFlatRegexKey(prefix, k)
			if err := matchRegexMap(v, regex, fullKey); err != nil {
				return err
			}
		}
	case map[string]any:
		hm := val.(map[string]any)
		for k, v := range hm {
			fullKey := buildFlatRegexKey(prefix, k)
			if err := validateRegexMap(v, regex, fullKey); err != nil {
				return err
			}
		}
	case []any:
		arr := val.([]any)
		for _, v := range arr {
			if err := validateRegexMap(v, regex, prefix); err != nil {
				return err
			}
		}
	default:
		strVal := fmt.Sprintf("%v", val)
		if err := matchRegexMap(strVal, regex, prefix); err != nil {
			return err
		}
	}
	return nil
}

func flatRegexMap(val any, regex map[string]string, prefix string) {
	if val == nil {
		return
	}
	switch val.(type) {
	case string:
		strVal := val.(string)
		addFlatRegexMapValue(regex, prefix, "", strVal)
	case map[string]string:
		hm := val.(map[string]string)
		for k, v := range hm {
			addFlatRegexMapValue(regex, prefix, k, v)
		}
	case map[string]any:
		hm := val.(map[string]any)
		for k, v := range hm {
			fullKey := buildFlatRegexKey(prefix, k)
			flatRegexMap(v, regex, fullKey)
		}
	case []any:
		arr := val.([]any)
		for _, v := range arr {
			flatRegexMap(v, regex, prefix)
		}
	default:
		log.WithFields(log.Fields{
			"val":     val,
			"valType": reflect.TypeOf(val),
		}).Info("cannot flat map value type")
	}
}

func buildFlatRegexKey(prefix string, k string) string {
	var fullKey string
	if prefix == "" {
		fullKey = k
	} else if k == "" {
		fullKey = prefix
	} else {
		fullKey = prefix + "." + k
	}
	return fullKey
}

func addFlatRegexMapValue(res map[string]string, prefix string, k string, v string) {
	fullKey := buildFlatRegexKey(prefix, k)
	if strings.Contains(v, `\w`) && strings.Contains(v, `[0-9]`) {
		v = WildRegex // mix regex are not supported
	}
	old := res[fullKey]
	if old == "" {
		res[fullKey] = "(" + v + ")"
	} else {
		old = old[1 : len(old)-1]
		parts := strings.Split(old, "|")
		matched := false
		for _, part := range parts {
			if v == part {
				matched = true
				break
			}
		}
		if !matched {
			res[fullKey] = "(" + old + "|" + v + ")"
		}
	}
}

func matchRegexMap(val any, regex map[string]string, key string) error {
	re := regex[key]
	if re == "" {
		return nil
	}

	pattern := `(` + PrefixTypeNumber + `|` + PrefixTypeBoolean + `|` +
		PrefixTypeString + `|` + PrefixTypeObject + `)`
	re = regexp.MustCompile(pattern).ReplaceAllString(re, "")

	match, err := regexp.Match(re, []byte(fmt.Sprintf("%s", val)))
	if err != nil {
		return err
	}
	if !match {
		return fmt.Errorf("key '%s' - value '%v' didn't match regex '%s'", key, val, re)
	}
	return nil
}

func addAlphabets(alphabets int, sb *strings.Builder, _ DataTemplateRequest) {
	if alphabets > 0 {
		//sb.WriteString(fmt.Sprintf(`[0-9a-zA-Z]{%d,%d}`, dataTemplate.MinMultiplier*alphabets, dataTemplate.MaxMultiplier*alphabets))
		sb.WriteString(AnyWordRegex)
	}
}

func addDigits(digits int, fraction int, negative bool, decimal bool,
	sb *strings.Builder, dataTemplate DataTemplateRequest) {
	if digits > 0 {
		if negative {
			sb.WriteString(`[+-]?`)
		}
		sb.WriteString(fmt.Sprintf(`[0-9]{%d,%d}`, dataTemplate.MinMultiplier*digits, dataTemplate.MaxMultiplier*digits))
		//sb.WriteString(`[0-9]{5,10}`)
		if decimal {
			sb.WriteString(fmt.Sprintf(`\.[0-9]{%d,%d}`, dataTemplate.MinMultiplier*fraction, dataTemplate.MaxMultiplier*fraction))
			//sb.WriteString(`\.[0-9]{3,5}`)
		}
	}
}
