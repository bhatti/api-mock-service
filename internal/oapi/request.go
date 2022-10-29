package oapi

import (
	"github.com/bhatti/api-mock-service/internal/types"
)

// Request Body
type Request struct {
	ContentType string
	PathParams  []Property
	QueryParams []Property
	Headers     map[string]string
	Body        []Property
}

func (req *Request) queryParamsMap() map[string]string {
	res := make(map[string]string)
	for _, param := range req.QueryParams {
		res[param.Name] = param.ValuesFor(param.Name)
	}
	return res
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
		MatchHeaders:     req.Headers,
		MatchQueryParams: req.queryParamsMap(),
		MatchContents:    string(contents),
	}, nil
}
