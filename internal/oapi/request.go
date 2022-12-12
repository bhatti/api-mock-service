package oapi

import (
	"github.com/bhatti/api-mock-service/internal/types"
)

// Request Body
type Request struct {
	PathParams  []Property
	QueryParams []Property
	Headers     []Property
	Body        []Property
}

func (req *Request) buildMockHTTPRequest(dataTemplate types.DataTemplateRequest) (res types.MockHTTPRequest, err error) {
	contents, err := marshalPropertyValueWithTypes(req.Body, dataTemplate.WithInclude(true))
	if err != nil {
		return res, err
	}
	return types.MockHTTPRequest{
		MatchHeaders:      propsToMap(req.Headers, asciiPattern, dataTemplate.WithInclude(true)),
		MatchQueryParams:  propsToMap(req.QueryParams, asciiPattern, dataTemplate.WithInclude(true)),
		MatchContents:     contents,
		ExamplePathParams: propsToMap(req.PathParams, asciiPattern, dataTemplate.WithInclude(false)),
	}, nil
}
