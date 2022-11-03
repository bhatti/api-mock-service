package web

import (
	"github.com/bhatti/api-mock-service/internal/types"
	"github.com/bhatti/api-mock-service/internal/utils"
	"net/http"
)

// BuildMockScenarioKeyData creates mock-scenario key from HTTP request
func BuildMockScenarioKeyData(req *http.Request) (keyData *types.MockScenarioKeyData, err error) {
	var reqBytes []byte
	reqBytes, req.Body, err = utils.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}

	method, err := types.ToMethod(req.Method)
	if err != nil {
		return nil, err
	}

	keyData = &types.MockScenarioKeyData{
		Method:           method,
		Name:             req.Header.Get(types.MockScenarioName),
		Path:             req.URL.Path,
		MatchQueryParams: make(map[string]string),
		MatchHeaders:     make(map[string]string),
		MatchContentType: req.Header.Get(types.ContentTypeHeader),
		MatchContents:    string(reqBytes),
	}
	for k, v := range req.URL.Query() {
		if len(v) > 0 {
			keyData.MatchQueryParams[k] = v[0]
		}
	}
	for k, v := range req.Header {
		if len(v) > 0 {
			keyData.MatchHeaders[k] = v[0]
		}
	}
	return keyData, nil
}
