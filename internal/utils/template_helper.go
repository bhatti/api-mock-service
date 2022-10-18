package utils

import (
	"bytes"
	"fmt"
	"github.com/bhatti/api-mock-service/internal/types"
	"html/template"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// UnescapeHTML flag
const UnescapeHTML = "UnescapeHTML"

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
	t, err := template.New("").Funcs(TemplateFuncs(dir)).Parse(body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template due to %s", err)
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
func TemplateFuncs(dir string) template.FuncMap {
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
			return RandomMinMax(toInt(min), toInt(max))
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
		"RandName": func() string {
			return RandName()
		},
		"SeededName": func(seed interface{}) string {
			return SeededName(toInt64(seed))
		},
		"RandString": func(max interface{}) string {
			return RandomString(toInt(max))
		},
		"Int": func(num interface{}) int64 {
			return toInt64(num)
		},
		"Float": func(num interface{}) float64 {
			return toFloat64(num)
		},
		"LT": func(a interface{}, b interface{}) bool {
			return toInt64(a) < toInt64(b)
		},
		"LE": func(a interface{}, b interface{}) bool {
			return toInt64(a) <= toInt64(b)
		},
		"EQ": func(a interface{}, b interface{}) bool {
			return toInt64(a) == toInt64(b)
		},
		"GT": func(a interface{}, b interface{}) bool {
			return toInt64(a) > toInt64(b)
		},
		"GE": func(a interface{}, b interface{}) bool {
			return toInt64(a) >= toInt64(b)
		},
		"RandFileLine": func(fileName string) template.HTML {
			return randFileLine(dir, fileName+types.MockDataExt, 0)
		},
		"SeededFileLine": func(fileName string, seed interface{}) template.HTML {
			return randFileLine(dir, fileName+types.MockDataExt, toInt64(seed))
		},
		"Time": func() string {
			return time.Now().Format(time.RFC3339)
		},
		"TimeFormat": func(format string) string {
			return time.Now().Format(format)
		},
		"AnySubString": func(str string) string {
			return AnySubString(str)
		},
	}
}

// PRIVATE FUNCTIONS

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

func toFloat64(input interface{}) (res float64) {
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
	default:
		res, _ = strconv.ParseFloat(fmt.Sprintf("%v", input), 64)
	}
	return
}

func toInt64(input interface{}) (res int64) {
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
