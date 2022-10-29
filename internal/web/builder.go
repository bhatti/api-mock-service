package web

import (
	"github.com/bhatti/api-mock-service/internal/types"
	"io"
)

func BuildMockScenarioKeyData(c APIContext) (keyData *types.MockScenarioKeyData, err error) {
	reqBody := []byte{}

	if c.Request().Body != nil {
		reqBody, err = io.ReadAll(c.Request().Body)
		if err != nil {
			return nil, err
		}
	}

	method, err := types.ToMethod(c.Request().Method)
	if err != nil {
		return nil, err
	}

	keyData = &types.MockScenarioKeyData{
		Method:           method,
		Name:             c.Request().Header.Get(types.MockScenarioName),
		Path:             c.Request().URL.Path,
		MatchQueryParams: make(map[string]string),
		MatchHeaders:     make(map[string]string),
		MatchContentType: c.Request().Header.Get(types.ContentTypeHeader),
		MatchContents:    string(reqBody),
	}
	for k, v := range c.Request().URL.Query() {
		if len(v) > 0 {
			keyData.MatchQueryParams[k] = v[0]
		}
	}
	for k, v := range c.Request().Header {
		if len(v) > 0 {
			keyData.MatchHeaders[k] = v[0]
		}
	}
	return keyData, nil
}
