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
	Title        string   // schema title
	Description  string
	Type         string
	SubType      string
	Enum         []string
	Const        string  // single allowed value (highest priority in Value())
	Default      string  // default value used when no other value is generated
	Example      string  // example value from OpenAPI spec
	Min          float64
	Max          float64
	ExclusiveMin bool    // minimum bound is exclusive (OAI 3.0)
	ExclusiveMax bool    // maximum bound is exclusive (OAI 3.0)
	MultipleOf   float64 // value must be a multiple of this (0 = no constraint)
	UniqueItems  bool    // array items must be unique
	MinProps     uint64  // minimum number of object properties
	MaxProps     uint64  // maximum number of object properties (0 = no limit)
	In           string
	Pattern      string
	Format       string
	Required     bool
	Deprecated   bool
	Nullable     bool
	ReadOnly     bool
	WriteOnly    bool
	Style        string // parameter serialization style (form, simple, matrix, etc.)
	Explode      bool   // expand array/object parameters into individual values
	Children     []Property
	matchRequest bool
}

func (prop *Property) GetName() string {
	if strings.HasPrefix(prop.Name, ".") {
		return prop.Name[1:]
	}
	return prop.Name
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
	// Const: single allowed value — highest priority for both mock and validation
	if prop.Const != "" {
		if dataTemplate.IncludeType {
			escaped := strings.ReplaceAll(prop.Const, ".", `\.`)
			return map[string]string{prop.Name: escaped}
		}
		return map[string]string{prop.Name: prop.Const}
	}
	// For mock response generation (not pattern matching), prefer Example then Default
	if !prop.matchRequest && !dataTemplate.IncludeType {
		if prop.Example != "" && (prop.Type == "string" || prop.Type == "integer" || prop.Type == "number" || prop.Type == "boolean") {
			return map[string]string{prop.Name: prop.Example}
		}
		if prop.Default != "" && (prop.Type == "string" || prop.Type == "integer" || prop.Type == "number" || prop.Type == "boolean") {
			return map[string]string{prop.Name: prop.Default}
		}
	}
	if prop.Type == "number" || prop.Type == "integer" {
		if dataTemplate.IncludeType {
			if prop.Pattern == "" {
				if prop.SubType == "number" {
					return map[string]string{
						prop.Name: fuzz.NumberPrefixRegex,
					}
				} else {
					return map[string]string{
						prop.Name: fuzz.IntPrefixRegex,
					}
				}
			}
			return map[string]string{
				prop.Name: fuzz.PrefixTypeNumber + prop.Pattern,
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
			if prop.Format != "" {
				if strings.Contains(prop.Format, "date") || strings.Contains(prop.Format, "time") {
					return map[string]string{prop.Name: `(\d{4}-[01]\d-[0-3]\dT[0-2]\d:[0-5]\d:[0-5]\d\.\d+([+-][0-2]\d:[0-5]\d|Z))|(\d{4}-[01]\d-[0-3]\dT[0-2]\d:[0-5]\d:[0-5]\d([+-][0-2]\d:[0-5]\d|Z))|(\d{4}-[01]\d-[0-3]\dT[0-2]\d:[0-5]\d([+-][0-2]\d:[0-5]\d|Z))`}
				} else if prop.Format == "uri" {
					return map[string]string{prop.Name: `http[s]?://(?:[a-zA-Z]|\d|[$-_@.&+]|[!*\(\),]|(?:%[0-9a-fA-F][0-9a-fA-F]))+`}
				} else if prop.Format == "host" || prop.Format == "hostname" {
					return map[string]string{prop.Name: `(?=.{1,255}$)[0-9A-Za-z](?:(?:[0-9A-Za-z]|-){0,61}[0-9A-Za-z])?(?:\.[0-9A-Za-z](?:(?:[0-9A-Za-z]|-){0,61}[0-9A-Za-z])?)*\.?`}
				} else if prop.Format == "email" {
					return map[string]string{prop.Name: `[a-z]{5,15}@[a-z]{5,15}.com`}
				} else if prop.Format == "phone" {
					return map[string]string{prop.Name: `1-\d{3}-\d{4}-\d{4}`}
				} else if prop.Format == "uuid" {
					return map[string]string{prop.Name: `[a-f\d]{8}-[a-f\d]{4}-[a-f\d]{4}-[a-f\d]{4}-[a-f\d]{12}`}
				} else if prop.Format == "ulid" {
					return map[string]string{prop.Name: `[0-9A-HJKMNP-TV-Z]{26}`}
				} else if prop.Format == "airport" {
					return map[string]string{prop.Name: `^[A-Z]{3}$`}
				} else if prop.Format == "locale" {
					return map[string]string{prop.Name: `^[a-z]{2,3}(-[A-Z]{2,3})?$`}
				} else if prop.Format == "country" {
					return map[string]string{prop.Name: `^[A-Z]{2}$`}
				} else if prop.Format == "zip" {
					return map[string]string{prop.Name: `^\d{5}(-\d{4})?$`}
				} else if prop.Format == "ip" || prop.Format == "ipv4" {
					return map[string]string{prop.Name: `^((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$`}
				} else if prop.Format == "ipv6" {
					return map[string]string{prop.Name: `([0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}`}
				} else if strings.Contains(prop.Format, "credit") {
					return map[string]string{prop.Name: `^(?:4[0-9]{12}(?:[0-9]{3})?|5[1-5][0-9]{14}|3[47][0-9]{13}|6(?:011|5[0-9]{2})[0-9]{12})$`}
				} else if prop.Format == "isbn10" {
					return map[string]string{prop.Name: `^(?:[0-9]{9}X|[0-9]{10})$`}
				} else if prop.Format == "isbn13" {
					return map[string]string{prop.Name: `^(?:[0-9]{13})$`}
				} else if prop.Format == "ssn" {
					return map[string]string{prop.Name: `^(?!000|666|9\d{2})(\d{3})-(?!00)(\d{2})-(?!0000)(\d{4})$`}
				} else if prop.Format == "password" {
					return map[string]string{prop.Name: fuzz.PrefixTypeString + `.{8,32}`}
				} else if prop.Format == "byte" || prop.Format == "binary" {
					return map[string]string{prop.Name: fuzz.PrefixTypeString + `[A-Za-z0-9+/]+=*`}
				} else if prop.Format == "int32" || prop.Format == "int64" {
					return map[string]string{prop.Name: fuzz.IntPrefixRegex}
				} else if prop.Format == "float" || prop.Format == "double" {
					return map[string]string{prop.Name: fuzz.NumberPrefixRegex}
				} else {
					return map[string]string{prop.Name: fuzz.PrefixTypeString + `\w+`}
				}
			} else if prop.Pattern != "" {
				return map[string]string{
					prop.Name: fuzz.PrefixTypeString + prop.Pattern,
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
				prop.Name: fuzz.PrefixTypeString + `\w+`, // TODO default string
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
				prop.Name: fuzz.PrefixTypeString + asciiPattern, // TODO default string
			}
		}
		return map[string]string{
			prop.Name: asciiPattern,
		}
	} else if prop.In == "body" && prop.Type == "object" {
		if dataTemplate.IncludeType {
			return map[string]string{
				prop.Name: fuzz.PrefixTypeObject + prop.Pattern,
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
			if prop.Pattern != "" {
				return fuzz.PrefixTypeStringToRegEx(prop.Pattern, dataTemplate)
			}
			return fuzz.PrefixTypeStringToRegEx(`\w+`, dataTemplate)
		}
		m := val.(map[string]string)
		if len(m[prop.Name]) > 0 {
			return m[prop.Name]
		}
	case map[string]any:
		if dataTemplate.IncludeType {
			if prop.Pattern != "" {
				return fuzz.PrefixTypeStringToRegEx(prop.Pattern, dataTemplate)
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
			if prop.Pattern != "" {
				return fuzz.PrefixTypeStringToRegEx(prop.Pattern, dataTemplate)
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
		if prop.Pattern != "" {
			return prop.Pattern
		}
		return `[\d\.]+`
	}

	min := int(prop.Min)
	max := int(prop.Max)
	// Adjust for exclusive bounds
	if prop.ExclusiveMin && min > 0 {
		min++
	}
	if prop.ExclusiveMax && max > 0 {
		max--
	}
	if prop.Type == "number" || prop.SubType == "number" {
		return fmt.Sprintf("{{RandFloatMinMax %d %d}}", min, max)
	}
	return fmt.Sprintf("{{RandIntMinMax %d %d}}", min, max)
}

func (prop *Property) boolValue() string {
	return "{{RandBool}}"
}

func (prop *Property) stringValue() string {
	if prop.matchRequest || prop.In == "path" || prop.In == "query" {
		if prop.Pattern != "" {
			return prop.Pattern
		}
		return `\w+`
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
		} else if prop.Format == "uuid" {
			return "{{UUID}}"
		} else if prop.Format == "ulid" {
			return "{{ULID}}"
		} else if prop.Format == "airport" {
			return "{{RandAirport}}"
		} else if prop.Format == "locale" {
			return "{{RandLocale}}"
		} else if prop.Format == "country" {
			return "{{RandCountryCode}}"
		} else if prop.Format == "zip" {
			return "{{RandZip}}"
		} else if prop.Format == "ip" {
			return "{{RandIP}}"
		} else if strings.Contains(prop.Format, "credit") {
			return "{{RandCreditCard}}"
		} else if prop.Format == "ssn" {
			return "{{RandSSN}}"
		} else if prop.Format == "isbn10" {
			return "{{RandRegex `^(?:[0-9]{9}X|[0-9]{10})$`}}"
		} else if prop.Format == "isbn13" {
			return "{{RandRegex `^(?:[0-9]{13})$`}}"
		} else if prop.Format == "password" {
			min := int(prop.Min)
			if min < 8 {
				min = 8
			}
			max := int(prop.Max)
			if max < min {
				max = 32
			}
			return fmt.Sprintf("{{RandStringMinMax %d %d}}", min, max)
		} else if prop.Format == "byte" || prop.Format == "binary" {
			return fmt.Sprintf("{{RandStringMinMax %d %d}}", int(prop.Min), int(prop.Max))
		} else if prop.Format == "ipv4" {
			return "{{RandIP}}"
		} else if prop.Format == "ipv6" {
			return "{{RandRegex `([0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}`}}"
		} else if prop.Format == "hostname" {
			return "{{RandHost}}"
		} else if prop.Format == "int32" || prop.Format == "int64" {
			return fmt.Sprintf("{{RandIntMinMax %d %d}}", int(prop.Min), int(prop.Max))
		} else if prop.Format == "float" || prop.Format == "double" {
			return fmt.Sprintf("{{RandFloatMinMax %d %d}}", int(prop.Min), int(prop.Max))
		} else {
			return "{{RandString 20}}"
		}
	} else if prop.Pattern != "" {
		return fmt.Sprintf("{{RandRegex `%s`}}", prop.Pattern)
	} else {
		return fmt.Sprintf("{{RandStringMinMax %d %d}}", int(prop.Min), int(prop.Max))
	}
}

func (prop *Property) arrayValue(dataTemplate fuzz.DataTemplateRequest) any {
	// TODO check if prop.matchRequest needs early exit here
	if prop.In == "path" || prop.In == "query" {
		if prop.Pattern != "" {
			return prop.Pattern
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
					if prop.SubType == "number" {
						childArr[i] = fuzz.NumberPrefixRegex
					} else {
						childArr[i] = fuzz.IntPrefixRegex
					}
				} else {
					childArr[i] = "{{RandIntMinMax 0 0}}"
				}
			} else if prop.SubType == "boolean" {
				if dataTemplate.IncludeType {
					childArr[i] = fuzz.BooleanPrefixRegex
				} else {
					childArr[i] = "{{RandBool}}"
				}
			} else if prop.SubType == "string" {
				if dataTemplate.IncludeType {
					childArr[i] = fuzz.PrefixTypeString + `\w+` // TODO default string
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
	if prop.Min == 0 {
		prop.Min = 1
	}
	if prop.Max == 0 {
		prop.Max = prop.Min + float64(fuzz.RandIntMinMax(1, 10))
	}
	return make([]any, fuzz.RandIntMinMax(int(prop.Min), int(prop.Max)))
}

func propsToMap(props []Property, defVal string, dataTemplate fuzz.DataTemplateRequest) (res map[string]string) {
	res = make(map[string]string)
	for _, prop := range props {
		if dataTemplate.IncludeType && !prop.Required {
			continue
		}
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
