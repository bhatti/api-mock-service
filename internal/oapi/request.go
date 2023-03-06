package oapi

import (
	"fmt"
	"github.com/bhatti/api-mock-service/internal/fuzz"
	"github.com/bhatti/api-mock-service/internal/types"
	"gopkg.in/yaml.v3"
)

// Request Body
type Request struct {
	PathParams  []Property
	QueryParams []Property
	Headers     []Property
	Body        []Property
}

func (req *Request) buildMockHTTPRequest(dataTemplate fuzz.DataTemplateRequest) (res types.MockHTTPRequest, err error) {
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
		if out, err := yaml.Marshal(obj); err == nil && obj != nil {
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
			assertions = append(assertions, val)
		}
	}

	return types.MockHTTPRequest{
		AssertHeadersPattern:     propsToMap(req.Headers, asciiPattern, dataTemplate.WithInclude(true)),
		AssertQueryParamsPattern: propsToMap(req.QueryParams, asciiPattern, dataTemplate.WithInclude(true)),
		Assertions:               assertions,
		Headers:                  propsToMap(req.Headers, asciiPattern, dataTemplate.WithInclude(false)),
		QueryParams:              propsToMap(req.QueryParams, asciiPattern, dataTemplate.WithInclude(false)),
		Contents:                 string(strippedContents),
		ExampleContents:          string(exampleContents),
		AssertContentsPattern:    matchContents,
		PathParams:               propsToMap(req.PathParams, asciiPattern, dataTemplate.WithInclude(false)),
	}, nil
}

func checkRequestHeader(name string, pattern string) string {
	validHeaders := map[string]bool{types.Authorization: true, types.ContentTypeHeader: true}
	if validHeaders[name] {
		if pattern == "" {
			return fmt.Sprintf(`VariableSizeGE headers.%s 1`, name)
		} else {
			return fmt.Sprintf(`VariableMatches headers.%s %s`, name, pattern)
		}
	}
	return ""
}
