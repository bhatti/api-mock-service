package oapi

import (
	"fmt"
	"github.com/bhatti/api-mock-service/internal/fuzz"
	"reflect"
	"strings"

	log "github.com/sirupsen/logrus"
)

var asciiPattern = `[\x20-\x7F]{1,128}`

// Property structure
type Property struct {
	Name         string
	Description  string
	Type         string
	SubType      string
	Enum         []string
	Min          float64
	Max          float64
	In           string
	Regex        string
	Format       string
	Children     []Property
	matchRequest bool
}

func (prop *Property) String() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s/%s/%s/%d->[",
		prop.Name, prop.Type, prop.SubType,
		len(prop.Children)))
	for _, child := range prop.Children {
		sb.WriteString(child.String())
	}
	sb.WriteString("];")
	return sb.String()
}

// Value of the property
func (prop *Property) Value(dataTemplate fuzz.DataTemplateRequest) any {
	if prop.Type == "number" || prop.Type == "integer" {
		if dataTemplate.IncludeType {
			if prop.Regex == "" {
				return map[string]string{
					prop.Name: fuzz.NumberPrefixRegex,
				}
			}
			return map[string]string{
				prop.Name: fuzz.PrefixTypeNumber + prop.Regex,
			}
		}
		return map[string]string{
			prop.Name: prop.numericValue(),
		}
	} else if prop.Type == "boolean" {
		if dataTemplate.IncludeType {
			return map[string]string{
				prop.Name: fuzz.BooleanPrefixRegex,
			}
		}
		return map[string]string{
			prop.Name: prop.boolValue(),
		}
	} else if prop.Type == "string" {
		if dataTemplate.IncludeType {
			if prop.Regex != "" {
				return map[string]string{
					prop.Name: fuzz.PrefixTypeString + prop.Regex,
				}
			} else if prop.Format != "" {
				if strings.Contains(prop.Format, "date") || strings.Contains(prop.Format, "time") {
					return `(\d{4}-[01]\d-[0-3]\dT[0-2]\d:[0-5]\d:[0-5]\d\.\d+([+-][0-2]\d:[0-5]\d|Z))|(\d{4}-[01]\d-[0-3]\dT[0-2]\d:[0-5]\d:[0-5]\d([+-][0-2]\d:[0-5]\d|Z))|(\d{4}-[01]\d-[0-3]\dT[0-2]\d:[0-5]\d([+-][0-2]\d:[0-5]\d|Z))`
				} else if prop.Format == "uri" {
					return `http[s]?://(?:[a-zA-Z]|[0-9]|[$-_@.&+]|[!*\(\),]|(?:%[0-9a-fA-F][0-9a-fA-F]))+`
				} else if prop.Format == "host" {
					return `(?=.{1,255}$)[0-9A-Za-z](?:(?:[0-9A-Za-z]|-){0,61}[0-9A-Za-z])?(?:\.[0-9A-Za-z](?:(?:[0-9A-Za-z]|-){0,61}[0-9A-Za-z])?)*\.?`
				} else {
					return `\w+`
				}
			} else if len(prop.Enum) > 0 {
				var sb strings.Builder
				sb.WriteString(fuzz.PrefixTypeString)
				sb.WriteString("(")
				for i, e := range prop.Enum {
					if i > 0 {
						sb.WriteString("|")
					}
					sb.WriteString(e)
				}
				sb.WriteString(")")
				return map[string]string{
					prop.Name: sb.String(),
				}
			}
			return map[string]string{
				prop.Name: fuzz.PrefixTypeString + `\w+`,
			}
		}
		return map[string]string{
			prop.Name: prop.stringValue(),
		}
	} else if len(prop.Children) > 0 || prop.Type == "array" {
		return prop.arrayValue(dataTemplate)
	} else if prop.In == "header" {
		if dataTemplate.IncludeType {
			return map[string]string{
				prop.Name: fuzz.PrefixTypeString + asciiPattern,
			}
		}
		return map[string]string{
			prop.Name: asciiPattern,
		}
	} else if prop.In == "body" && prop.Type == "object" {
		if dataTemplate.IncludeType {
			return map[string]string{
				prop.Name: fuzz.PrefixTypeObject + prop.Regex,
			}
		}
		return map[string]string{
			prop.Name: "{{RandDict}}",
		}
	} else {
		log.WithFields(log.Fields{
			"component": "Property",
			"Name":      prop.Name,
			"In":        prop.In,
			"Children":  len(prop.Children),
			"Type":      prop.Type}).Debugf("unknown type")
		if dataTemplate.IncludeType {
			return map[string]string{}
		}
		return map[string]string{
			prop.Name: "{{RandDict}}",
		}
	}
}

func (prop *Property) mapValue(dataTemplate fuzz.DataTemplateRequest) string {
	val := prop.Value(dataTemplate)
	if val == nil {
		return ""
	}
	switch val.(type) {
	case map[string]string:
		if dataTemplate.IncludeType {
			if prop.Regex != "" {
				return fuzz.PrefixTypeStringToRegEx(prop.Regex, dataTemplate)
			}
			return fuzz.PrefixTypeStringToRegEx(`\w+`, dataTemplate)
		}
		m := val.(map[string]string)
		if len(m[prop.Name]) > 0 {
			return m[prop.Name]
		}
	case map[string]any:
		if dataTemplate.IncludeType {
			if prop.Regex != "" {
				return fuzz.PrefixTypeStringToRegEx(prop.Regex, dataTemplate)
			}
			return fuzz.PrefixTypeStringToRegEx(`\w+`, dataTemplate)
		}
		m := val.(map[string]any)
		val := m[prop.Name]
		if val != nil {
			return fmt.Sprintf("%v", val)
		}
	case string:
		if dataTemplate.IncludeType {
			if prop.Regex != "" {
				return fuzz.PrefixTypeStringToRegEx(prop.Regex, dataTemplate)
			}
			return fuzz.PrefixTypeStringToRegEx(`\w+`, dataTemplate)
		}
		return val.(string)
	default:
		log.WithFields(log.Fields{
			"name":    prop.Name,
			"type":    prop.Type,
			"subtype": prop.SubType,
			"val":     val,
			"valType": reflect.TypeOf(val),
		}).Debug("unknown value type")
	}
	return ""
}

func (prop *Property) numericValue() string {
	if prop.matchRequest || prop.In == "path" || prop.In == "query" {
		if prop.Regex != "" {
			return prop.Regex
		}
		return `\d+`
	}

	return fmt.Sprintf("{{RandNumMinMax %d %d}}", int(prop.Max), int(prop.Max))
}

func (prop *Property) boolValue() string {
	return "{{RandBool}}"
}

func (prop *Property) stringValue() string {
	if prop.matchRequest || prop.In == "path" || prop.In == "query" {
		if prop.Regex != "" {
			return prop.Regex
		}
		return ""
	}
	if len(prop.Enum) > 0 {
		choices := strings.Join(prop.Enum, " ")
		return fmt.Sprintf("{{EnumString `%s`}}", choices)
	} else if prop.Format != "" {
		if prop.Format == "date" {
			return "{{Date}}"
		} else if strings.Contains(prop.Format, "time") {
			return "{{Time}}"
		} else if prop.Format == "host" {
			return "{{RandHost}}"
		} else if prop.Format == "email" {
			return "{{RandEmail}}"
		} else if prop.Format == "phone" {
			return "{{RandPhone}}"
		} else if prop.Format == "uri" {
			return "{{RandURL}}"
		} else {
			return "{{RandString 20}}"
		}
	} else if prop.Regex != "" {
		return fmt.Sprintf("{{RandRegex `%s`}}", prop.Regex)
	} else {
		return fmt.Sprintf("{{RandStringMinMax %d %d}}", int(prop.Min), int(prop.Max))
	}
}

func (prop *Property) arrayValue(dataTemplate fuzz.DataTemplateRequest) any {
	if prop.matchRequest || prop.In == "path" || prop.In == "query" {
		if prop.Regex != "" {
			return prop.Regex
		}
		return nil
	}

	childArr := make([]any, 0)
	for _, child := range prop.Children {
		val := child.Value(dataTemplate)
		if val != nil {
			childArr = append(childArr, val)
		}
	}
	if prop.Type == "array" && prop.SubType != "object" && prop.SubType != "" {
		if len(childArr) > 0 {
			return childArr
		}
		childArr = prop.buildValueArray()
		for i := 0; i < len(childArr); i++ {
			if prop.SubType == "number" || prop.SubType == "integer" {
				if dataTemplate.IncludeType {
					childArr[i] = fuzz.NumberPrefixRegex
				} else {
					childArr[i] = "{{RandNumMinMax 0 0}}"
				}
			} else if prop.SubType == "boolean" {
				if dataTemplate.IncludeType {
					childArr[i] = fuzz.BooleanPrefixRegex
				} else {
					childArr[i] = "{{RandBool}}"
				}
			} else if prop.SubType == "string" {
				if dataTemplate.IncludeType {
					childArr[i] = fuzz.PrefixTypeString + `\w+`
				} else {
					childArr[i] = "{{RandStringMinMax 0 0}}"
				}
			}
		}
		if prop.Name != "" {
			return map[string]any{prop.Name: childArr}
		}
		return childArr
	}

	// if property has name or is object (e.g. jobs openapi)
	if prop.Name == "" || prop.Type == "object" || (prop.Type == "array" && len(childArr) > 1) {
		res := make(map[string]any)
		for _, child := range childArr {
			switch child.(type) {
			case map[string]string:
				subProperty := child.(map[string]string)
				for k, v := range subProperty {
					res[k] = v
				}
			case map[string]any:
				subProperty := child.(map[string]any)
				for k, v := range subProperty {
					res[k] = v
				}
			}
		}

		if prop.Type == "array" {
			arr := prop.buildValueArray()
			for i := 0; i < len(arr); i++ {
				arr[i] = res
			}
			// see jobs-openapi for examples
			if prop.Name != "" {
				return map[string]any{prop.Name: arr}
			}
			return arr
		}
		// see jobs-openapi for examples
		if prop.Name != "" {
			return map[string]any{prop.Name: res}
		}
		return res
	}
	log.WithFields(log.Fields{
		"name":    prop.Name,
		"type":    prop.Type,
		"subtype": prop.SubType,
		"Res":     len(childArr),
	}).Debug("default else for parsing property")
	return map[string]any{
		prop.Name: childArr,
	}
}

func (prop *Property) buildValueArray() []any {
	if prop.Max == 0 {
		prop.Max = prop.Min + float64(fuzz.RandNumMinMax(1, 5))
	}
	if prop.Min == 0 {
		prop.Min = prop.Max
	}
	return make([]any, fuzz.RandNumMinMax(int(prop.Min), int(prop.Max)))
}

func propsToMap(props []Property, defVal string, dataTemplate fuzz.DataTemplateRequest) (res map[string]string) {
	res = make(map[string]string)
	for _, prop := range props {
		val := prop.mapValue(dataTemplate)
		if val == "" {
			val = defVal
		}
		if val != "" {
			res[prop.Name] = val
		}
	}
	return
}

func propsToMapArray(props []Property, dataTemplate fuzz.DataTemplateRequest) (res map[string][]string) {
	res = make(map[string][]string)
	for _, prop := range props {
		val := prop.mapValue(dataTemplate)
		if val != "" {
			res[prop.Name] = []string{val}
		}
	}
	return res
}
