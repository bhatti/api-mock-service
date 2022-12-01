package oapi

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"strings"

	"github.com/bhatti/api-mock-service/internal/utils"
)

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
func (prop *Property) Value() interface{} {
	if prop.Type == "number" || prop.Type == "integer" {
		return map[string]string{
			prop.Name: prop.numericValue(),
		}
	} else if prop.Type == "boolean" {
		return map[string]string{
			prop.Name: prop.boolValue(),
		}
	} else if prop.Type == "string" {
		return map[string]string{
			prop.Name: prop.stringValue(),
		}
	} else if len(prop.Children) > 0 || prop.Type == "array" {
		return prop.arrayValue()
	} else if prop.In == "header" {
		return map[string]string{
			prop.Name: ".+",
		}
	} else if prop.In == "body" && prop.Type == "object" {
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
		return map[string]string{
			prop.Name: fmt.Sprintf("{{RandStringArrayMinMax %d %d}}", int(prop.Max), int(prop.Max)),
		}
	}
}
func (prop *Property) mapValue() string {
	val := prop.Value()
	if val == nil {
		return ""
	}
	switch val.(type) {
	case map[string]string:
		m := val.(map[string]string)
		if len(m[prop.Name]) > 0 {
			return m[prop.Name]
		}
	case map[string]interface{}:
		m := val.(map[string]interface{})
		if m[prop.Name] != nil {
			return fmt.Sprintf("%v", m[prop.Name])
		}
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
		} else if prop.Format == "date-time" {
			return "{{Time}}"
		} else if prop.Format == "uri" {
			return "https://{{RandName}}.com"
		} else {
			return "{{RandString 20}}"
		}
	} else if prop.Regex != "" {
		return fmt.Sprintf("{{RandRegex `%s`}}", prop.Regex)
	} else {
		return fmt.Sprintf("{{RandStringMinMax %d %d}}", int(prop.Max), int(prop.Max))
	}
}

func (prop *Property) arrayValue() interface{} {
	if prop.matchRequest || prop.In == "path" || prop.In == "query" {
		if prop.Regex != "" {
			return prop.Regex
		}
		return nil
	}

	childArr := make([]interface{}, 0)
	for _, child := range prop.Children {
		val := child.Value()
		if val != nil {
			childArr = append(childArr, val)
		}
	}

	// if property has name
	if prop.Name == "" {
		if prop.Type == "array" && prop.SubType != "object" && prop.SubType != "" {
			return childArr
		}
		res := make(map[string]interface{})
		for _, child := range childArr {
			switch child.(type) {
			case map[string]string:
				for k, v := range child.(map[string]string) {
					res[k] = v
				}
			case map[string]interface{}:
				for k, v := range child.(map[string]interface{}) {
					res[k] = v
				}
			}
		}

		if prop.Type == "array" {
			arr := make([]interface{}, utils.RandNumMinMax(5, 20))
			for i := 0; i < len(arr); i++ {
				arr[i] = res
			}
			return arr
		}
		return res
	}
	return map[string]interface{}{
		prop.Name: childArr,
	}
}

func propsToMap(props []Property) (res map[string]string) {
	res = make(map[string]string)
	for _, prop := range props {
		val := prop.mapValue()
		if val != "" {
			res[prop.Name] = prop.mapValue()
		}
	}
	return
}

func propsToMapArray(props []Property) (res map[string][]string) {
	res = make(map[string][]string)
	for _, prop := range props {
		val := prop.mapValue()
		if val != "" {
			res[prop.Name] = []string{prop.mapValue()}
		}
	}
	return res
}
