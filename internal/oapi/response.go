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
		if out, err := json.Marshal(obj); err == nil && obj != nil {
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
	assertions := []string{
		`ResponseTimeMillisLE 5000`,
		fmt.Sprintf(`ResponseStatusMatches %d`, res.StatusCode),
	}
	respHeaderAssertions := make(map[string]string)
	if res.ContentType != "" {
		assertions = append(assertions, fmt.Sprintf(`VariableMatches headers.Content-Type %s`, res.ContentType))
		respHeaderAssertions[types.ContentTypeHeader] = res.ContentType
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
		SetVariables:          fuzz.ExtractTopPrimitiveAttributes(exampleContents, 5),
	}, nil
}
