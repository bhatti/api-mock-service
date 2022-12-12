package oapi

import (
	"github.com/bhatti/api-mock-service/internal/fuzz"
	"github.com/bhatti/api-mock-service/internal/types"
)

// Response Body
type Response struct {
	ContentType string
	StatusCode  int
	Headers     []Property
	Body        []Property
}

func (res *Response) buildMockHTTPResponse(dataTemplate fuzz.DataTemplateRequest) (_ types.MockHTTPResponse, err error) {
	contents, err := marshalPropertyValue(res.Body, dataTemplate.WithInclude(false))
	if err != nil {
		return
	}
	matchContents, err := marshalPropertyValueWithTypes(res.Body, dataTemplate.WithInclude(true))
	if err != nil {
		return
	}
	return types.MockHTTPResponse{
		StatusCode:    res.StatusCode,
		Headers:       propsToMapArray(res.Headers, dataTemplate.WithInclude(false)),
		Contents:      string(contents),
		MatchContents: matchContents,
	}, nil
}
