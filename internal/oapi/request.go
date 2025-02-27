package oapi

import (
	"encoding/json"
	"fmt"
	"github.com/bhatti/api-mock-service/internal/fuzz"
	"github.com/bhatti/api-mock-service/internal/types"
)

// Request Body
type Request struct {
	PathParams  []Property
	QueryParams []Property
	Headers     []Property
	Body        []Property
}

func (req *Request) buildMockHTTPRequest(dataTemplate fuzz.DataTemplateRequest) (res types.APIRequest, err error) {
	strippedContents, err := marshalPropertyValue(req.Body, dataTemplate.WithInclude(true), true)
	if err != nil {
		return
	}
	quotedContents, err := marshalPropertyValue(req.Body, dataTemplate.WithInclude(true), false)
	if err != nil {
		return
	}
	var exampleContents []byte
	if obj, err := fuzz.UnmarshalArrayOrObject(quotedContents); err == nil {
		obj = fuzz.GenerateFuzzData(obj)
		if out, err := json.MarshalIndent(obj, "", "  "); err == nil && obj != nil {
			exampleContents = out
		}
	}
	matchContents, err := marshalPropertyValueWithTypes(req.Body, dataTemplate.WithInclude(true), true)
	if err != nil {
		return res, err
	}

	assertions := make([]string, 0)

	for _, header := range req.Headers {
		if val := checkRequestHeader(header.Name, header.Pattern); val != "" {
			assertions = types.AddAssertion(assertions, val)
		}
	}
	for _, bp := range req.Body {
		assertionsMap := make(map[string]bool)
		addPropertyAssertions("contents", bp, assertionsMap, 0)
		for k := range assertionsMap {
			assertions = types.AddAssertion(assertions, k)
		}
	}

	return types.APIRequest{
		Headers:                  propsToMap(req.Headers, asciiPattern, dataTemplate.WithInclude(false)),
		QueryParams:              propsToMap(req.QueryParams, asciiPattern, dataTemplate.WithInclude(false)),
		Contents:                 string(strippedContents),
		ExampleContents:          string(exampleContents),
		PathParams:               propsToMap(req.PathParams, asciiPattern, dataTemplate.WithInclude(false)),
		AssertContentsPattern:    matchContents,
		AssertHeadersPattern:     propsToMap(req.Headers, asciiPattern, dataTemplate.WithInclude(true)),
		AssertQueryParamsPattern: propsToMap(req.QueryParams, asciiPattern, dataTemplate.WithInclude(true)),
		Assertions:               assertions,
		HTTPVersion:              "",
		Variables:                make(map[string]string),
	}, nil
}

func checkRequestHeader(name string, pattern string) string {
	validHeaders := map[string]bool{types.AuthorizationHeader: true, types.ContentTypeHeader: true}
	if validHeaders[name] {
		if pattern == "" {
			return fmt.Sprintf(`PropertyLenGE headers.%s 5`, name)
		}
		return fmt.Sprintf(`PropertyMatches headers.%s %s`, name, pattern)
	}
	return ""
}

func addPropertyAssertions(parent string, property Property, assertions map[string]bool, depth int) {
	if depth > 1 {
		return
	}
	if property.GetName() != "" && parent != "" && property.Required {
		assertions[fmt.Sprintf(`HasProperty %s.%s`, parent, property.GetName())] = true
	}
	for _, child := range property.Children {
		if property.GetName() != "" {
			parent = parent + "." + property.GetName()
		}
		addPropertyAssertions(parent, child, assertions, depth+1)
	}
}
