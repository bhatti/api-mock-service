package oapi

import (
	"github.com/bhatti/api-mock-service/internal/types"
)

// Request Body
type Request struct {
	ContentType string
	PathParams  []Property
	QueryParams []Property
	Headers     []Property
	Body        []Property
}

func (req *Request) buildMockHTTPRequest() (_ types.MockHTTPRequest, err error) {
	var contents []byte
	contents, err = marshalPropertyValue(req.Body)
	if err != nil {
		return
	}
	return types.MockHTTPRequest{
		MatchContentType:   req.ContentType,
		ExampleHeaders:     propsToMap(req.Headers),
		ExamplePathParams:  propsToMap(req.PathParams),
		ExampleQueryParams: propsToMap(req.QueryParams),
		ExampleContents:    string(contents),
	}, nil
}
