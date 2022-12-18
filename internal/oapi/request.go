package oapi

import (
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

func (req *Request) buildMockHTTPRequest(dataTemplate fuzz.DataTemplateRequest) (res types.MockHTTPRequest, err error) {
	contents, err := marshalPropertyValue(req.Body, dataTemplate.WithInclude(false))
	if err != nil {
		return
	}
	matchContents, err := marshalPropertyValueWithTypes(req.Body, dataTemplate.WithInclude(true))
	if err != nil {
		return res, err
	}
	return types.MockHTTPRequest{
		MatchHeaders:      propsToMap(req.Headers, asciiPattern, dataTemplate.WithInclude(true)),
		MatchQueryParams:  propsToMap(req.QueryParams, asciiPattern, dataTemplate.WithInclude(true)),
		ExampleContents:   string(contents),
		MatchContents:     matchContents,
		ExamplePathParams: propsToMap(req.PathParams, asciiPattern, dataTemplate.WithInclude(false)),
	}, nil
}
