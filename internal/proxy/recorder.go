package proxy

import (
	"context"
	"fmt"
	"github.com/bhatti/api-mock-service/internal/fuzz"
	"github.com/bhatti/api-mock-service/internal/utils"
	log "github.com/sirupsen/logrus"
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
		return fmt.Errorf("header for %s is not defined to connect to remote url '%s'", types.MockURL, c.Request().URL)
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
	resBody []byte,
	resHeaders map[string][]string,
	status int,
	mockScenarioRepository repository.MockScenarioRepository) (resContentType string, err error) {

	if resHeaders != nil {
		val := resHeaders[types.ContentTypeHeader]
		if len(val) > 0 {
			resContentType = val[0]
		}
	}

	dataTemplate := fuzz.NewDataTemplateRequest(true, 1, 1)
	matchReqContents, err := fuzz.UnmarshalArrayOrObjectAndExtractTypesAndMarshal(string(reqBody), dataTemplate)
	if err != nil {
		log.WithFields(log.Fields{
			"Path":   req.URL,
			"Method": req.Method,
			"Error":  err,
		}).Warnf("failed to unmarshal and extrate types for request")
	}
	matchResContents, err := fuzz.UnmarshalArrayOrObjectAndExtractTypesAndMarshal(string(resBody), dataTemplate)
	if err != nil {
		log.WithFields(log.Fields{
			"Path":   req.URL,
			"Method": req.Method,
			"Error":  err,
		}).Warnf("failed to unmarshal and extrate types for response")
	}
	scenario := &types.MockScenario{
		Method:         types.MethodType(req.Method),
		Name:           req.Header.Get(types.MockScenarioName),
		Path:           u.Path,
		Group:          utils.NormalizeGroup("", u.Path),
		Authentication: make(map[string]types.MockAuthorization),
		Request: types.MockHTTPRequest{
			AssertQueryParamsPattern: make(map[string]string),
			AssertHeadersPattern:     map[string]string{types.ContentTypeHeader: req.Header.Get(types.ContentTypeHeader)},
			AssertContentsPattern:    matchReqContents,
			QueryParams:              make(map[string]string),
			Headers:                  make(map[string]string),
			Contents:                 string(reqBody),
		},
		Response: types.MockHTTPResponse{
			Headers:               resHeaders,
			Contents:              string(resBody),
			StatusCode:            status,
			AssertContentsPattern: matchResContents,
		},
	}

	for k, v := range req.URL.Query() {
		if len(v) > 0 {
			scenario.Request.QueryParams[k] = v[0]
			if config.AssertQueryParams(k) {
				scenario.Request.AssertQueryParamsPattern[k] = v[0]
			}
		}
	}
	for k, v := range req.Header {
		if len(v) > 0 {
			scenario.Request.Headers[k] = v[0]
			if config.AssertHeader(k) {
				scenario.Request.AssertHeadersPattern[k] = v[0]
			}
		}
	}

	if scenario.Name == "" {
		scenario.SetName("recorded-")
	}

	scenario.Description = fmt.Sprintf("recorded at %v for %s", time.Now().UTC(), u)
	if err = mockScenarioRepository.Save(scenario); err != nil {
		return "", err
	}
	return
}
