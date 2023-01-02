package oapi

import (
	"github.com/bhatti/api-mock-service/internal/fuzz"
	"github.com/bhatti/api-mock-service/internal/types"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

// Response Body
type Response struct {
	ContentType string
	StatusCode  int
	Headers     []Property
	Body        []Property
}

func (res *Response) buildMockHTTPResponse(dataTemplate fuzz.DataTemplateRequest) (_ types.MockHTTPResponse, err error) {
	strippedContents, err := marshalPropertyValue(res.Body, dataTemplate.WithInclude(false), true)
	if err != nil {
		return
	}
	quotedContents, err := marshalPropertyValue(res.Body, dataTemplate.WithInclude(false), false)
	if err != nil {
		return
	}
	var exampleContents []byte
	if obj, err := fuzz.UnmarshalArrayOrObject(quotedContents); err == nil {
		obj = fuzz.GenerateFuzzData(obj)
		if out, err := yaml.Marshal(obj); err == nil && obj != nil {
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
	return types.MockHTTPResponse{
		StatusCode:      res.StatusCode,
		Headers:         propsToMapArray(res.Headers, dataTemplate.WithInclude(false)),
		Contents:        string(strippedContents),
		ExampleContents: string(exampleContents),
		MatchContents:   matchContents,
	}, nil
}
