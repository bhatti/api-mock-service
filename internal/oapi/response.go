package oapi

import (
	"encoding/json"
	"fmt"
	"github.com/bhatti/api-mock-service/internal/fuzz"
	"github.com/bhatti/api-mock-service/internal/types"
	log "github.com/sirupsen/logrus"
)

// Response Body
type Response struct {
	ContentType string
	StatusCode  int
	Headers     []Property
	Body        []Property
}

func (res *Response) buildMockHTTPResponse(dataTemplate fuzz.DataTemplateRequest) (_ types.APIResponse, err error) {
	strippedContents, err := marshalPropertyValue(res.Body, dataTemplate.WithInclude(false), true)
	if err != nil {
		return
	}
	quotedContents, err := marshalPropertyValue(res.Body, dataTemplate.WithInclude(false), false)
	if err != nil {
		return
	}
	res.Headers = append(res.Headers, Property{
		Name:    types.ContentTypeHeader,
		Type:    "string",
		Pattern: res.ContentType,
	})
	var exampleContents []byte
	if obj, err := fuzz.UnmarshalArrayOrObject(quotedContents); err == nil {
		obj = fuzz.GenerateFuzzData(obj)
		if out, err := json.MarshalIndent(obj, "", "  "); err == nil && obj != nil {
			exampleContents = out
		}
	} else {
		log.WithFields(log.Fields{
			"Error": err,
			"Body":  string(strippedContents),
			"Data":  obj,
		}).Warnf("failed to unmarshal response")
	}
	matchContents, err := marshalPropertyValueWithTypes(res.Body, dataTemplate.WithInclude(true), true)
	if err != nil {
		return
	}
	assertions := make([]string, 0)
	respHeaderAssertions := make(map[string]string)
	if res.ContentType != "" {
		assertions = types.AddAssertion(assertions, fmt.Sprintf(`PropertyMatches headers.Content-Type %s`, res.ContentType))
		respHeaderAssertions[types.ContentTypeHeader] = res.ContentType
	}
	for _, bp := range res.Body {
		assertionsMap := make(map[string]bool)
		addPropertyAssertions("contents", bp, assertionsMap, 0)
		for k := range assertionsMap {
			assertions = types.AddAssertion(assertions, k)
		}
	}
	return types.APIResponse{
		StatusCode:            res.StatusCode,
		Headers:               propsToMapArray(res.Headers, dataTemplate.WithInclude(false)),
		Contents:              string(strippedContents),
		ExampleContents:       string(exampleContents),
		AssertHeadersPattern:  respHeaderAssertions,
		AssertContentsPattern: matchContents,
		Assertions:            assertions,
		HTTPVersion:           "",
		AddSharedVariables:    fuzz.ExtractTopPrimitiveAttributes(exampleContents, 5),
	}, nil
}
