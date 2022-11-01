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
	// ignore req.PathParams
	return types.MockHTTPRequest{
		MatchContentType: req.ContentType,
		MatchHeaders:     propsToMap(req.Headers),
		MatchQueryParams: propsToMap(req.QueryParams),
		MatchContents:    string(contents),
	}, nil
}
