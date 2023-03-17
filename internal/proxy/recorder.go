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
	config             *types.Configuration
	client             web.HTTPClient
	scenarioRepository repository.APIScenarioRepository
}

// NewRecorder instantiates controller for updating api -scenarios
func NewRecorder(
	config *types.Configuration,
	client web.HTTPClient,
	scenarioRepository repository.APIScenarioRepository) *Recorder {
	return &Recorder{
		config:             config,
		client:             client,
		scenarioRepository: scenarioRepository,
	}
}

// Handle records request
func (r *Recorder) Handle(c web.APIContext) (err error) {
	started := time.Now()
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

	status, httpVersion, resBody, resHeaders, err := r.client.Handle(
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
		r.config,
		u,
		c.Request(),
		reqBody,
		resBytes,
		resHeaders,
		status,
		httpVersion,
		started,
		time.Now(),
		r.scenarioRepository)
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
	resStatus int,
	resHttpVersion string,
	started time.Time,
	ended time.Time,
	scenarioRepository repository.APIScenarioRepository) (resContentType string, err error) {

	scenario := types.BuildScenarioFromHTTP(
		config,
		"recorded-",
		u,
		req.Method,
		"",
		req.Proto,
		resHttpVersion,
		reqBody,
		resBody,
		req.URL.Query(),
		req.PostForm,
		req.Header,
		"",
		resHeaders,
		"",
		resStatus)

	if err = scenarioRepository.Save(scenario); err != nil {
		return "", err
	}
	if err = scenarioRepository.SaveHistory(scenario, u, started, ended); err != nil {
		return "", err
	}
	return
}
