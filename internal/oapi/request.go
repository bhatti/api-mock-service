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
		if out, err := json.Marshal(obj); err == nil && obj != nil {
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
			return fmt.Sprintf(`VariableSizeGE headers.%s 5`, name)
		}
		return fmt.Sprintf(`VariableMatches headers.%s %s`, name, pattern)
	}
	return ""
}
