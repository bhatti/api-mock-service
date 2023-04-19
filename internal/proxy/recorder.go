package proxy

import (
	"context"
	"fmt"
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
	config                *types.Configuration
	client                web.HTTPClient
	scenarioRepository    repository.APIScenarioRepository
	groupConfigRepository repository.GroupConfigRepository
}

// NewRecorder instantiates controller for updating api -scenarios
func NewRecorder(
	config *types.Configuration,
	client web.HTTPClient,
	scenarioRepository repository.APIScenarioRepository,
	groupConfigRepository repository.GroupConfigRepository,
) *Recorder {
	return &Recorder{
		config:                config,
		client:                client,
		scenarioRepository:    scenarioRepository,
		groupConfigRepository: groupConfigRepository,
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
		return fmt.Errorf("failed to parse mock url due to %w", err)
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

	scenario, resContentType, err := saveMockResponse(
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

	// Embedding this check for chaos settings
	if groupConfig, err := r.groupConfigRepository.Load(scenario.Group); err == nil {
		status := groupConfig.GetHTTPStatus()
		if status >= 300 {
			return c.String(status, "injected fault from recorder")
		}
		delay := groupConfig.GetDelayLatency()
		if delay > 0 {
			log.WithFields(log.Fields{
				"Component":   "Recorder",
				"Group":       scenario.Group,
				"GroupConfig": groupConfig,
				"Delay":       delay,
			}).Infof("artificial sleep wait")
			time.Sleep(delay)
		}
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
	resHTTPVersion string,
	started time.Time,
	ended time.Time,
	scenarioRepository repository.APIScenarioRepository) (scenario *types.APIScenario, resContentType string, err error) {

	scenario, err = types.BuildScenarioFromHTTP(
		config,
		"Recorded",
		u,
		req.Method,
		"",
		req.Proto,
		resHTTPVersion,
		reqBody,
		resBody,
		req.URL.Query(),
		req.PostForm,
		req.Header,
		"",
		resHeaders,
		"",
		resStatus,
		started,
		ended)
	if err != nil {
		return nil, "", err
	}

	if err = scenarioRepository.Save(scenario); err != nil {
		return nil, "", err
	}
	if err = scenarioRepository.SaveHistory(scenario, u.String(), started, ended); err != nil {
		return nil, "", err
	}
	resContentType = scenario.Response.ContentType("")
	return
}
