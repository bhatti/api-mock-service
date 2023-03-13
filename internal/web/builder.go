package web

import (
	"github.com/bhatti/api-mock-service/internal/types"
	"github.com/bhatti/api-mock-service/internal/utils"
	"net/http"
)

// BuildMockScenarioKeyData creates api-scenario key from HTTP request
func BuildMockScenarioKeyData(req *http.Request) (keyData *types.APIKeyData, err error) {
	var reqBytes []byte
	reqBytes, req.Body, err = utils.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}

	method, err := types.ToMethod(req.Method)
	if err != nil {
		return nil, err
	}

	keyData = &types.APIKeyData{
		Method:                   method,
		Name:                     req.Header.Get(types.MockScenarioName),
		Path:                     req.URL.Path,
		AssertQueryParamsPattern: make(map[string]string),
		AssertHeadersPattern:     map[string]string{types.ContentTypeHeader: req.Header.Get(types.ContentTypeHeader)},
		AssertContentsPattern:    string(reqBytes),
	}
	for k, v := range req.URL.Query() {
		if len(v) > 0 {
			keyData.AssertQueryParamsPattern[k] = v[0]
		}
	}
	for k, v := range req.Header {
		if len(v) > 0 {
			keyData.AssertHeadersPattern[k] = v[0]
		}
	}
	return keyData, nil
}
