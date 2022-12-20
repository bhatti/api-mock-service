package oapi

import (
	"github.com/bhatti/api-mock-service/internal/fuzz"
	"github.com/bhatti/api-mock-service/internal/types"
	"gopkg.in/yaml.v3"
)

// Request Body
type Request struct {
	PathParams  []Property
	QueryParams []Property
	Headers     []Property
	Body        []Property
}

func (req *Request) buildMockHTTPRequest(dataTemplate fuzz.DataTemplateRequest) (res types.MockHTTPRequest, err error) {
	contents, err := marshalPropertyValue(req.Body, dataTemplate.WithInclude(true))
	if err != nil {
		return
	}
	if obj, err := fuzz.UnmarshalArrayOrObject([]byte(contents)); err == nil {
		obj = fuzz.GenerateFuzzData(obj)
		contents, _ = yaml.Marshal(obj)
	}
	matchContents, err := marshalPropertyValueWithTypes(req.Body, dataTemplate.WithInclude(true))
	if err != nil {
		return res, err
	}
	return types.MockHTTPRequest{
		MatchHeaders:       propsToMap(req.Headers, asciiPattern, dataTemplate.WithInclude(true)),
		ExampleHeaders:     propsToMap(req.Headers, asciiPattern, dataTemplate.WithInclude(false)),
		MatchQueryParams:   propsToMap(req.QueryParams, asciiPattern, dataTemplate.WithInclude(true)),
		ExampleQueryParams: propsToMap(req.QueryParams, asciiPattern, dataTemplate.WithInclude(false)),
		ExampleContents:    string(contents),
		MatchContents:      matchContents,
		ExamplePathParams:  propsToMap(req.PathParams, asciiPattern, dataTemplate.WithInclude(false)),
	}, nil
}
