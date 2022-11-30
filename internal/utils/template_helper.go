package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"html/template"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/bhatti/api-mock-service/internal/types"
)

// UnescapeHTML flag
const UnescapeHTML = "UnescapeHTML"

// MatchScenarioPredicate checks if predicate match
func MatchScenarioPredicate(matched *types.MockScenarioKeyData, target *types.MockScenarioKeyData, requestCount uint64) bool {
	if matched.Predicate == "" {
		return true
	}
	// Find any params for query params and path variables
	params := matched.MatchGroups(target.Path)
	for k, v := range matched.MatchQueryParams {
		params[k] = v
	}
	for k, v := range target.MatchQueryParams {
		params[k] = v
	}
	params[types.RequestCount] = fmt.Sprintf("%d", requestCount)
	out, err := ParseTemplate("", []byte(matched.Predicate), params)
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

// ParseTemplate parses GO template with dynamic parameters
func ParseTemplate(dir string, byteBody []byte, data interface{}) ([]byte, error) {
	body := string(byteBody)
	if !strings.Contains(body, "{{") {
		return byteBody, nil
	}
	emptyLineRegex, err := regexp.Compile(`(?m)^\s*$[\r\n]*|[\r\n]+\s+\z`)
	if err != nil {
		return nil, err
	}
	t, err := template.New("").Funcs(TemplateFuncs(dir, data)).Parse(body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template due to %w", err)
	}
	var out bytes.Buffer
	err = t.Execute(&out, data)
	if err != nil {
		return nil, fmt.Errorf("failed to execute template due to '%s', data=%v",
			err, data)
	}
	strResponse := emptyLineRegex.ReplaceAllString(out.String(), "")
	switch data.(type) {
	case map[string]interface{}:
		m := data.(map[string]interface{})
		if m[UnescapeHTML] == true {
			strResponse = strings.ReplaceAll(strResponse, "&lt;", "<")
		}
	}
	return []byte(strResponse), nil
}

// TemplateFuncs returns template functions
func TemplateFuncs(dir string, data interface{}) template.FuncMap {
	return template.FuncMap{
		"Dict": func(values ...interface{}) (map[string]interface{}, error) {
			if len(values)%2 != 0 {
				return nil, fmt.Errorf("invalid dict call")
			}
			dict := make(map[string]interface{}, len(values)/2)
			for i := 0; i < len(values); i += 2 {
				key, ok := values[i].(string)
				if !ok {
					return nil, fmt.Errorf("dict keys must be strings")
				}
				dict[key] = values[i+1]
			}
			return dict, nil
		},
		"Iterate": func(input interface{}) []int {
			count := toInt(input)
			var i int
			var Items []int
			for i = 0; i < count; i++ {
				Items = append(Items, i)
			}
			return Items
		},
		"LastIter": func(val interface{}, max interface{}) bool {
			return toInt(val) == toInt(max)-1
		},
		"Add": func(n interface{}, plus interface{}) int {
			return toInt(n) + toInt(plus)
		},
		"Unescape": func(s string) template.HTML {
			return template.HTML(s)
		},
		"RandNumMinMax": func(min interface{}, max interface{}) int {
			return RandNumMinMax(toInt(min), toInt(max))
		},
		"RandNumMax": func(max interface{}) int {
			return Random(toInt(max))
		},
		"Udid": func() string {
			return Udid()
		},
		"SeededUdid": func(seed interface{}) string {
			return SeededUdid(toInt64(seed))
		},
		"RandCity": func() string {
			return RandCity()
		},
		"SeededCity": func(seed interface{}) string {
			return SeededCity(toInt64(seed))
		},
		"RandBool": func() bool {
			return RandBool()
		},
		"SeededBool": func(seed interface{}) bool {
			return SeededBool(toInt64(seed))
		},
		"RandCountry": func() string {
			return RandCountry()
		},
		"SeededCountry": func(seed interface{}) string {
			return SeededCountry(toInt64(seed))
		},
		"RandCountryCode": func() string {
			return RandCountryCode()
		},
		"SeededCountryCode": func(seed interface{}) string {
			return SeededCountryCode(toInt64(seed))
		},
		"RandName": func() string {
			return RandName()
		},
		"SeededName": func(seed interface{}) string {
			return SeededName(toInt64(seed))
		},
		"RandString": func(max interface{}) string {
			return RandString(toInt(max))
		},
		"RandStringMinMax": func(min interface{}, max interface{}) string {
			return RandStringMinMax(toInt(min), toInt(max))
		},
		"RandStringArrayMinMax": func(min interface{}, max interface{}) template.HTML {
			arr := RandStringArrayMinMax(toInt(min), toInt(max))
			for i := range arr {
				arr[i] = fmt.Sprintf(`"%s"`, arr[i])
			}
			return template.HTML("[" + strings.Join(arr, ",") + "]")
		},
		"RandIntArrayMinMax": func(min interface{}, max interface{}) []int {
			return RandIntArrayMinMax(toInt(min), toInt(max))
		},
		"RandRegex": func(re string) string {
			return RandRegex(re)
		},
		"RandPhone": func() string {
			return RandPhone()
		},
		"RandEmail": func() string {
			return RandEmail()
		},
		"Int": func(num interface{}) int64 {
			return toInt64(num)
		},
		"Float": func(num interface{}) float64 {
			return ToFloat64(num)
		},
		"LT": func(a interface{}, b interface{}) bool {
			return ToFloat64(a) < ToFloat64(b)
		},
		"LE": func(a interface{}, b interface{}) bool {
			return ToFloat64(a) <= ToFloat64(b)
		},
		"EQ": func(a interface{}, b interface{}) bool {
			return ToFloat64(a) == ToFloat64(b)
		},
		"GT": func(a interface{}, b interface{}) bool {
			return ToFloat64(a) > ToFloat64(b)
		},
		"GE": func(a interface{}, b interface{}) bool {
			return ToFloat64(a) >= ToFloat64(b)
		},
		"Nth": func(a interface{}, b interface{}) bool {
			return toInt(a)%toInt(b) == 0
		},
		"LTRequest": func(n interface{}) bool {
			reqCount := parseRequestCount(data)
			return reqCount >= 0 && reqCount < toInt(n)
		},
		"GERequest": func(n interface{}) bool {
			reqCount := parseRequestCount(data)
			return reqCount >= 0 && reqCount >= toInt(n)
		},
		"NthRequest": func(n interface{}) bool {
			reqCount := parseRequestCount(data)
			return reqCount >= 0 && reqCount%toInt(n) == 0
		},
		"JSONFileProperty": func(fileName string, name string) template.HTML {
			return toJSON(fileProperty(dir, fileName+types.MockDataExt, name))
		},
		"YAMLFileProperty": func(fileName string, name string) template.HTML {
			return toYAML(fileProperty(dir, fileName+types.MockDataExt, name))
		},
		"FileProperty": func(fileName string, name string) interface{} {
			return fileProperty(dir, fileName+types.MockDataExt, name)
		},
		"RandFileLine": func(fileName string) template.HTML {
			return randFileLine(dir, fileName+types.MockDataExt, 0)
		},
		"SeededFileLine": func(fileName string, seed interface{}) template.HTML {
			return randFileLine(dir, fileName+types.MockDataExt, toInt64(seed))
		},
		"Date": func() string {
			return time.Now().Format("2006-01-02")
		},
		"Time": func() string {
			return time.Now().Format(time.RFC3339)
		},
		"TimeFormat": func(format string) string {
			return time.Now().Format(format)
		},
		"EnumString": func(str ...interface{}) string {
			return EnumString(str...)
		},
		"EnumInt": func(vals ...interface{}) int64 {
			return EnumInt(vals...)
		},
	}
}

// PRIVATE FUNCTIONS

func parseRequestCount(data interface{}) int {
	switch data.(type) {
	case map[string]string:
		params := data.(map[string]string)
		count := params[types.RequestCount]
		if count == "" {
			return -1
		}
		return toInt(count)
	case map[string]interface{}:
		params := data.(map[string]interface{})
		count := params[types.RequestCount]
		if count == nil {
			return -1
		}
		return toInt(count)
	default:
		return -1
	}
}

func toJSON(val interface{}) template.HTML {
	str, err := json.Marshal(val)
	if err != nil {
		str = []byte(err.Error())
	}
	return template.HTML(strings.TrimSpace(string(str)))
}

func toYAML(val interface{}) template.HTML {
	str, err := yaml.Marshal(val)
	if err != nil {
		str = []byte(err.Error())
	}
	return template.HTML(strings.TrimSpace(string(str)))
}

func fileProperty(dir string, fileName string, name string) interface{} {
	if validFileName(fileName) {
		return FileProperty(filepath.Join(dir, fileName), name)
	}
	return fmt.Sprintf("invalid file-name '%s'", fileName)
}

func randFileLine(dir string, fileName string, seed int64) template.HTML {
	var line string
	if validFileName(fileName) {
		line = SeededFileLine(filepath.Join(dir, fileName), seed)
	} else {
		line = fmt.Sprintf("invalid file-name '%s'", fileName)
	}
	return template.HTML(line)
}

func toInt(input interface{}) (res int) {
	if input == nil {
		return 0
	}
	switch input.(type) {
	case int:
		res = input.(int)
	case uint:
		res = int(input.(uint))
	case int32:
		res = int(input.(int32))
	case int64:
		res = int(input.(int64))
	default:
		res, _ = strconv.Atoi(fmt.Sprintf("%v", input))
	}
	return
}

// ToFloat64 converter
func ToFloat64(input interface{}) (res float64) {
	if input == nil {
		return 0
	}
	switch input.(type) {
	case int:
		res = float64(input.(int))
	case uint:
		res = float64(input.(uint))
	case int32:
		res = float64(input.(int32))
	case int64:
		res = float64(input.(int64))
	case float32:
		res = float64(input.(float32))
	case float64:
		res = input.(float64)
	case *float64:
		f := input.(*float64)
		if f != nil {
			res = *f
		}
	case *float32:
		f := input.(*float32)
		if f != nil {
			res = float64(*f)
		}
	case *uint64:
		f := input.(*uint64)
		if f != nil {
			res = float64(*f)
		}
	case *int64:
		f := input.(*int64)
		if f != nil {
			res = float64(*f)
		}
	default:
		res, _ = strconv.ParseFloat(fmt.Sprintf("%v", input), 64)
	}
	return
}

func toInt64(input interface{}) (res int64) {
	if input == nil {
		return 0
	}
	switch input.(type) {
	case int:
		res = int64(input.(int))
	case uint:
		res = int64(input.(uint))
	case int32:
		res = int64(input.(int32))
	case int64:
		res = input.(int64)
	default:
		res, _ = strconv.ParseInt(fmt.Sprintf("%v", input), 10, 64)
	}
	return
}

func validFileName(name string) bool {
	if matched, err := regexp.Match(`^[\w\d-_]+(\.?[\w\d-_]+)+$`, []byte(name)); err == nil {
		return matched
	}
	return false
}
