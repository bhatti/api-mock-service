package web

import (
	"github.com/bhatti/api-mock-service/internal/types"
	"io"
	"net/http"
)

func BuildMockScenarioKeyData(req *http.Request) (keyData *types.MockScenarioKeyData, err error) {
	reqBody := []byte{}

	if req.Body != nil {
		reqBody, err = io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
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
		MatchContents:    string(reqBody),
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
