package proxy

import (
	"context"
	"fmt"
	"github.com/bhatti/api-mock-service/internal/utils"
	"net/http"
	"net/url"
	"time"

	"github.com/bhatti/api-mock-service/internal/repository"
	"github.com/bhatti/api-mock-service/internal/types"
	"github.com/bhatti/api-mock-service/internal/web"
)

// Recorder structure
type Recorder struct {
	config                 *types.Configuration
	client                 web.HTTPClient
	mockScenarioRepository repository.MockScenarioRepository
}

// NewRecorder instantiates controller for updating mock-scenarios
func NewRecorder(
	config *types.Configuration,
	client web.HTTPClient,
	mockScenarioRepository repository.MockScenarioRepository) *Recorder {
	return &Recorder{
		config:                 config,
		client:                 client,
		mockScenarioRepository: mockScenarioRepository,
	}
}

// Handle records request
func (r *Recorder) Handle(c web.APIContext) (err error) {
	mockURL := c.Request().Header.Get(types.MockURL)
	if mockURL == "" {
		return fmt.Errorf("header for %s is not defined to connect to remote url", types.MockURL)
	}
	u, err := url.Parse(mockURL)
	if err != nil {
		return err
	}

	var reqBody []byte
	reqBody, c.Request().Body, err = utils.ReadAll(c.Request().Body)
	if err != nil {
		return err
	}

	status, resBody, resHeaders, err := r.client.Handle(
		context.Background(),
		mockURL,
		c.Request().Method,
		c.Request().Header,
		nil,
		c.Request().Body,
	)
	if err != nil {
		return err
	}
	var resBytes []byte
	resBytes, resBody, err = utils.ReadAll(resBody)
	if err != nil {
		return err
	}

	resContentType, err := saveMockResponse(
		r.config, u, c.Request(), reqBody, resBytes, resHeaders, status, r.mockScenarioRepository)
	if err != nil {
		return err
	}

	return c.Blob(status, resContentType, resBytes)
}

func saveMockResponse(
	config *types.Configuration,
	u *url.URL,
	req *http.Request,
	reqBody []byte,
	resBytes []byte,
	resHeaders map[string][]string,
	status int,
	mockScenarioRepository repository.MockScenarioRepository) (resContentType string, err error) {

	if resHeaders != nil {
		val := resHeaders[types.ContentTypeHeader]
		if len(val) > 0 {
			resContentType = val[0]
		}

	}

	scenario := &types.MockScenario{
		Method: types.MethodType(req.Method),
		Name:   req.Header.Get(types.MockScenarioName),
		Path:   u.Path,
		Request: types.MockHTTPRequest{
			MatchQueryParams:   make(map[string]string),
			MatchHeaders:       make(map[string]string),
			MatchContents:      string(reqBody),
			MatchContentType:   req.Header.Get(types.ContentTypeHeader),
			ExampleQueryParams: make(map[string]string),
			ExampleHeaders:     make(map[string]string),
			ExampleContents:    string(reqBody),
		},
		Response: types.MockHTTPResponse{
			Headers:     resHeaders,
			ContentType: resContentType,
			Contents:    string(resBytes),
			StatusCode:  status,
		},
	}
	for k, v := range req.URL.Query() {
		if len(v) > 0 {
			scenario.Request.ExampleQueryParams[k] = v[0]
		}
	}
	for k, v := range req.Header {
		if len(v) > 0 {
			scenario.Request.ExampleHeaders[k] = v[0]
			if config.MatchHeader(k) {
				scenario.Request.MatchHeaders[k] = v[0]
			}
		}
	}
	if scenario.Name == "" {
		scenario.Name = fmt.Sprintf("recorded-%s-%s", scenario.NormalName(), scenario.Digest())
	}
	scenario.Description = fmt.Sprintf("recorded at %v for %s", time.Now().UTC(), u)
	if err = mockScenarioRepository.Save(scenario); err != nil {
		return "", err
	}
	return
}
