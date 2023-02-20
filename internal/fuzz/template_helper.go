package fuzz

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v3"
	"html/template"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

// UnescapeHTML flag
const UnescapeHTML = "UnescapeHTML"

// RequestCount name
const RequestCount = "_RequestCount"

// MockDataExt extension
const MockDataExt = ".dat"

// ParseTemplate parses GO template with dynamic parameters
func ParseTemplate(dir string, byteBody []byte, data any) ([]byte, error) {
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
		log.WithFields(log.Fields{
			"Error": err,
			"Dir":   dir,
			"Body":  string(byteBody),
			"Data":  data,
		}).Warnf("failed to execute template")
		return nil, fmt.Errorf("failed to execute template due to '%s', data=%v",
			err, data)
	}
	strResponse := emptyLineRegex.ReplaceAllString(out.String(), "")
	switch data.(type) {
	case map[string]any:
		m := data.(map[string]any)
		if m[UnescapeHTML] == true {
			strResponse = strings.ReplaceAll(strResponse, "&lt;", "<")
		}
	}
	return []byte(strResponse), nil
}

// TemplateFuncs returns template functions
func TemplateFuncs(dir string, data any) template.FuncMap {
	return template.FuncMap{
		"Dict": func(values ...any) (map[string]any, error) {
			if len(values)%2 != 0 {
				return nil, fmt.Errorf("invalid dict call")
			}
			dict := make(map[string]any, len(values)/2)
			for i := 0; i < len(values); i += 2 {
				key, ok := values[i].(string)
				if !ok {
					return nil, fmt.Errorf("dict keys must be strings")
				}
				dict[key] = values[i+1]
			}
			return dict, nil
		},
		"Iterate": func(input any) []int {
			count := toInt(input)
			var i int
			var Items []int
			for i = 0; i < count; i++ {
				Items = append(Items, i)
			}
			return Items
		},
		"LastIter": func(val any, max any) bool {
			return toInt(val) == toInt(max)-1
		},
		"Add": func(n any, plus any) int {
			return toInt(n) + toInt(plus)
		},
		"Unescape": func(s string) template.HTML {
			return template.HTML(s)
		},
		"RandIntMinMax": func(min any, max any) int {
			return RandIntMinMax(toInt(min), toInt(max))
		},
		"RandIntMax": func(max any) int {
			return RandIntMax(toInt(max))
		},
		"RandFloatMinMax": func(min any, max any) float64 {
			return RandFloatMinMax(ToFloat64(min), ToFloat64(max))
		},
		"RandFloatMax": func(max any) float64 {
			return RandFloatMax(ToFloat64(max))
		},
		"RandWord": func(min, max any) string {
			return RandWord(toInt(min), toInt(max))
		},
		"RandSentence": func(min, max any) string {
			return RandSentence(toInt(min), toInt(max))
		},
		"RandParagraph": func(min, max any) string {
			return RandParagraph(toInt(min), toInt(max))
		},
		"UUID": func() string {
			return UUID()
		},
		"SeededUUID": func(seed any) string {
			return SeededUUID(toInt64(seed))
		},
		"RandCity": func() string {
			return RandCity()
		},
		"SeededCity": func(seed any) string {
			return SeededCity(toInt64(seed))
		},
		"RandBool": func() bool {
			return RandBool()
		},
		"SeededBool": func(seed any) bool {
			return SeededBool(toInt64(seed))
		},
		"RandCountry": func() string {
			return RandCountry()
		},
		"SeededCountry": func(seed any) string {
			return SeededCountry(toInt64(seed))
		},
		"RandCountryCode": func() string {
			return RandCountryCode()
		},
		"SeededCountryCode": func(seed any) string {
			return SeededCountryCode(toInt64(seed))
		},
		"RandName": func() string {
			return RandName()
		},
		"SeededName": func(seed any) string {
			return SeededName(toInt64(seed))
		},
		"RandString": func(max any) string {
			return RandString(toInt(max))
		},
		"RandStringMinMax": func(min any, max any) string {
			return RandStringMinMax(toInt(min), toInt(max))
		},
		"RandStringArrayMinMax": func(min any, max any) template.HTML {
			arr := RandStringArrayMinMax(toInt(min), toInt(max))
			for i := range arr {
				arr[i] = fmt.Sprintf(`"%s"`, arr[i])
			}
			return template.HTML("[" + strings.Join(arr, ",") + "]")
		},
		"RandIntArrayMinMax": func(min any, max any) []int {
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
		"RandHost": func() string {
			return RandHost()
		},
		"RandURL": func() string {
			return RandURL()
		},
		"RandDict": func() template.HTML {
			dict := make(map[string]any)
			for i := 0; i < RandIntMinMax(3, 6); i += 2 {
				key := RandName()
				if i == 0 {
					dict[key] = RandTriString(".")
				} else if i == 2 {
					dict[key] = RandBool()
				} else {
					dict[key] = RandIntMinMax(100, 5000)
				}
			}
			j, _ := json.Marshal(dict)
			return template.HTML(j)
		},
		"Int": func(num any) int64 {
			return toInt64(num)
		},
		"Float": func(num any) float64 {
			return ToFloat64(num)
		},
		"LT": func(a any, b any) bool {
			return ToFloat64(a) < ToFloat64(b)
		},
		"LE": func(a any, b any) bool {
			return ToFloat64(a) <= ToFloat64(b)
		},
		"EQ": func(a any, b any) bool {
			return ToFloat64(a) == ToFloat64(b)
		},
		"GT": func(a any, b any) bool {
			return ToFloat64(a) > ToFloat64(b)
		},
		"GE": func(a any, b any) bool {
			return ToFloat64(a) >= ToFloat64(b)
		},
		"Nth": func(a any, b any) bool {
			return toInt(a)%toInt(b) == 0
		},
		"LTRequest": func(n any) bool {
			reqCount := parseRequestCount(data)
			return reqCount >= 0 && reqCount < toInt(n)
		},
		"GERequest": func(n any) bool {
			reqCount := parseRequestCount(data)
			return reqCount >= 0 && reqCount >= toInt(n)
		},
		"NthRequest": func(n any) bool {
			reqCount := parseRequestCount(data)
			return reqCount >= 0 && reqCount%toInt(n) == 0
		},
		"JSONFileProperty": func(fileName string, name string) template.HTML {
			return toJSON(fileProperty(dir, fileName+MockDataExt, name))
		},
		"YAMLFileProperty": func(fileName string, name string) template.HTML {
			return toYAML(fileProperty(dir, fileName+MockDataExt, name))
		},
		"FileProperty": func(fileName string, name string) any {
			return fileProperty(dir, fileName+MockDataExt, name)
		},
		"RandFileLine": func(fileName string) template.HTML {
			return randFileLine(dir, fileName+MockDataExt, 0)
		},
		"SeededFileLine": func(fileName string, seed any) template.HTML {
			return randFileLine(dir, fileName+MockDataExt, toInt64(seed))
		},
		"VariableEquals": func(varName string, target any) bool {
			return VariableEquals(varName, data, target)
		},
		"VariableContains": func(varName string, target any) bool {
			return VariableContains(varName, target, data)
		},
		"VariableEQ": func(varName string, size any) bool {
			return VariableNumber(varName, data) == ToFloat64(size)
		},
		"VariableLE": func(varName string, size any) bool {
			return VariableNumber(varName, data) <= ToFloat64(size)
		},
		"VariableGE": func(varName string, size any) bool {
			return VariableNumber(varName, data) >= ToFloat64(size)
		},
		"VariableSizeEQ": func(varName string, size any) bool {
			return VariableSize(varName, data) == toInt(size)
		},
		"VariableSizeLE": func(varName string, size any) bool {
			return VariableSize(varName, data) <= toInt(size)
		},
		"VariableSizeGE": func(varName string, size any) bool {
			return VariableSize(varName, data) >= toInt(size)
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
		"EnumString": func(str ...any) string {
			return EnumString(str...)
		},
		"EnumInt": func(vals ...any) int64 {
			return EnumInt(vals...)
		},
	}
}

// PRIVATE FUNCTIONS

func parseRequestCount(data any) int {
	switch data.(type) {
	case map[string]string:
		params := data.(map[string]string)
		count := params[RequestCount]
		if count == "" {
			return -1
		}
		return toInt(count)
	case map[string]any:
		params := data.(map[string]any)
		count := params[RequestCount]
		if count == nil {
			return -1
		}
		return toInt(count)
	default:
		return -1
	}
}

func toJSON(val any) template.HTML {
	str, err := json.Marshal(val)
	if err != nil {
		str = []byte(err.Error())
	}
	return template.HTML(strings.TrimSpace(string(str)))
}

func toYAML(val any) template.HTML {
	str, err := yaml.Marshal(val)
	if err != nil {
		str = []byte(err.Error())
	}
	return template.HTML(strings.TrimSpace(string(str)))
}

func fileProperty(dir string, fileName string, name string) any {
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

func toInt(input any) (res int) {
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
func ToFloat64(input any) (res float64) {
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

func toInt64(input any) (res int64) {
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
