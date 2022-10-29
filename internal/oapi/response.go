package oapi

import "github.com/bhatti/api-mock-service/internal/types"

// Response Body
type Response struct {
	ContentType string
	StatusCode  int
	Headers     map[string][]string
	Body        []Property
}

func (res *Response) buildMockHTTPResponse() (_ types.MockHTTPResponse, err error) {
	var contents []byte
	contents, err = marshalPropertyValue(res.Body)
	if err != nil {
		return
	}
	return types.MockHTTPResponse{
		StatusCode:  res.StatusCode,
		Headers:     res.Headers,
		ContentType: res.ContentType,
		Contents:    string(contents),
	}, nil
}
