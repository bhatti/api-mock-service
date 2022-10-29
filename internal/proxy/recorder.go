package proxy

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/url"
	"time"

	"github.com/bhatti/api-mock-service/internal/repository"
	"github.com/bhatti/api-mock-service/internal/types"
	"github.com/bhatti/api-mock-service/internal/web"
)

// MockURL header
const MockURL = "Mock-Url"

// Recorder structure
type Recorder struct {
	client                 web.HTTPClient
	mockScenarioRepository repository.MockScenarioRepository
}

// NewRecorder instantiates controller for updating mock-scenarios
func NewRecorder(
	client web.HTTPClient,
	mockScenarioRepository repository.MockScenarioRepository) *Recorder {
	return &Recorder{
		client:                 client,
		mockScenarioRepository: mockScenarioRepository,
	}
}

// Handle records request
func (r *Recorder) Handle(c web.APIContext) (err error) {
	mockURL := c.Request().Header.Get(MockURL)
	if mockURL == "" {
		return fmt.Errorf("header for %s is not defined to connect to remote url", MockURL)
	}
	u, err := url.Parse(mockURL)
	if err != nil {
		return err
	}

	reqBody := []byte{}

	if c.Request().Body != nil {
		reqBody, err = io.ReadAll(c.Request().Body)
		if err != nil {
			return err
		}
	}

	status, resBody, resHeaders, err := r.client.Handle(
		context.Background(),
		mockURL,
		c.Request().Method,
		c.Request().Header,
		nil,
		io.NopCloser(bytes.NewReader(reqBody)),
	)
	if err != nil {
		return err
	}
	defer func() {
		_ = resBody.Close()
	}()

	var resBytes []byte
	if resBody != nil {
		resBytes, err = io.ReadAll(resBody)
		if err != nil {
			return err
		}
	}

	var resContentType string
	if resHeaders != nil {
		val := resHeaders[types.ContentTypeHeader]
		if len(val) > 0 {
			resContentType = val[0]
		}

	}

	scenario := &types.MockScenario{
		Method: types.MethodType(c.Request().Method),
		Name:   c.Request().Header.Get(types.MockScenarioName),
		Path:   u.Path,
		Request: types.MockHTTPRequest{
			MatchQueryParams: make(map[string]string),
			MatchHeaders:     make(map[string]string),
			MatchContentType: c.Request().Header.Get(types.ContentTypeHeader),
			MatchContents:    string(reqBody),
		},
		Response: types.MockHTTPResponse{
			Headers:     resHeaders,
			ContentType: resContentType,
			Contents:    string(resBytes),
			StatusCode:  status,
		},
	}
	for k, v := range c.Request().URL.Query() {
		if len(v) > 0 {
			scenario.Request.MatchQueryParams[k] = v[0]
		}
	}
	for k, v := range c.Request().Header {
		if len(v) > 0 {
			scenario.Request.MatchHeaders[k] = v[0]
		}
	}
	if scenario.Name == "" {
		scenario.Name = fmt.Sprintf("recorded-%s-%s", scenario.NormalName(), scenario.Digest())
	}
	scenario.Description = fmt.Sprintf("recorded at %v", time.Now().UTC())
	if err = r.mockScenarioRepository.Save(scenario); err != nil {
		return err
	}

	return c.Blob(status, resContentType, resBytes)
}
