package proxy

import (
	"context"
	"fmt"
	"github.com/bhatti/api-mock-service/internal/fuzz"
	"github.com/bhatti/api-mock-service/internal/utils"
	log "github.com/sirupsen/logrus"
	"net/http"
	"net/url"
	"strings"
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

	assertions := []string{
		`ResponseTimeMillisLE 5000`,
		fmt.Sprintf(`ResponseStatusMatches %d`, status),
	}
	reqHeaderAssertions := make(map[string]string)
	if req.Header.Get(types.ContentTypeHeader) != "" {
		reqHeaderAssertions[types.ContentTypeHeader] = req.Header.Get(types.ContentTypeHeader)
	}
	respHeaderAssertions := make(map[string]string)
	if len(resHeaders[types.ContentTypeHeader]) > 0 {
		assertions = append(assertions, fmt.Sprintf(`VariableMatches headers.Content-Type %s`,
			resHeaders[types.ContentTypeHeader][0]))
		respHeaderAssertions[types.ContentTypeHeader] = resHeaders[types.ContentTypeHeader][0]
	}
	scenario := &types.MockScenario{
		Method:         types.MethodType(req.Method),
		Name:           req.Header.Get(types.MockScenarioName),
		Path:           u.Path,
		BaseURL:        u.Scheme + "://" + u.Host,
		Group:          utils.NormalizeGroup("", u.Path),
		Authentication: make(map[string]types.MockAuthorization),
		Request: types.MockHTTPRequest{
			AssertQueryParamsPattern: make(map[string]string),
			AssertHeadersPattern:     reqHeaderAssertions,
			AssertContentsPattern:    matchReqContents,
			QueryParams:              make(map[string]string),
			Headers:                  make(map[string]string),
			Contents:                 string(reqBody),
			ExampleContents:          string(reqBody),
		},
		Response: types.MockHTTPResponse{
			Headers:               resHeaders,
			Contents:              string(resBody),
			ExampleContents:       string(resBody),
			StatusCode:            status,
			AssertHeadersPattern:  respHeaderAssertions,
			AssertContentsPattern: matchResContents,
			Assertions:            assertions,
		},
	}
	scenario.Tags = []string{scenario.Group}

	for k, v := range req.URL.Query() {
		if len(v) > 0 {
			scenario.Request.QueryParams[k] = fuzz.PrefixTypeExample + v[0]
			if config.AssertQueryParams(k) {
				scenario.Request.AssertQueryParamsPattern[k] = v[0]
			}
		}
	}
	for k, v := range req.Header {
		if len(v) > 0 {
			scenario.Request.Headers[k] = fuzz.PrefixTypeExample + v[0]
			if strings.Contains(strings.ToUpper(k), "TARGET") {
				scenario.Request.AssertHeadersPattern[k] = v[0]
				parts := strings.Split(v[0], ".")
				if u.Path == "/" {
					if len(parts) >= 2 {
						scenario.Group = parts[len(parts)-2] + "_" + parts[len(parts)-1]
						scenario.Tags = []string{scenario.Group}
					}
				}
			} else if config.AssertHeader(k) {
				scenario.Request.AssertHeadersPattern[k] = v[0]
			}
		}
	}
	authHeader := scenario.Request.AuthHeader()
	if strings.Contains(authHeader, "AWS") {
		scenario.Authentication["aws.auth.sigv4"] = types.MockAuthorization{
			Type:   "apiKey",
			Name:   web.Authorization,
			In:     "header",
			Format: "awsSigv4",
			Scheme: "x-amazon-apigateway-authtype",
		}
		scenario.Authentication["smithy.api.httpApiKeyAuth"] = types.MockAuthorization{
			Type: "apiKey",
			Name: "x-api-key",
			In:   "header",
		}
		scenario.Authentication["bearerAuth"] = types.MockAuthorization{
			Type:   "http",
			Name:   web.Authorization,
			In:     "header",
			Scheme: "bearer",
			Format: "JWT",
		}
	} else if authHeader != "" {
		scenario.Authentication["basicAuth"] = types.MockAuthorization{
			Type:   "http",
			Name:   web.Authorization,
			In:     "header",
			Scheme: "basic",
		}
		scenario.Authentication["bearerAuth"] = types.MockAuthorization{
			Type:   "http",
			Name:   web.Authorization,
			In:     "header",
			Scheme: "bearer",
			Format: "auth-scheme",
		}
	}

	if scenario.Name == "" {
		scenario.SetName("recorded-" + scenario.Group + "-")
	}

	scenario.Description = fmt.Sprintf("recorded at %v for %s", time.Now().UTC(), u)
	if err = mockScenarioRepository.Save(scenario); err != nil {
		return "", err
	}
	return
}
