package types

import (
	"encoding/json"
	"fmt"
	"github.com/beevik/etree"
	"github.com/bhatti/api-mock-service/internal/fuzz"
	log "github.com/sirupsen/logrus"
	"github.com/xeipuuv/gojsonschema"
	"hash/adler32"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"runtime/debug"
	"strconv"
	"strings"
	"time"
)

// APIAuthorization defines mock auth parameters
type APIAuthorization struct {
	Type   string `json:"type,omitempty" yaml:"type,omitempty"`
	Name   string `json:"name,omitempty" yaml:"name,omitempty"`
	In     string `json:"in,omitempty" yaml:"in,omitempty"`
	Format string `json:"format,omitempty" yaml:"format,omitempty"`
	Scheme string `json:"scheme,omitempty" yaml:"scheme,omitempty"`
	URL    string `json:"url,omitempty" yaml:"url,omitempty"`
}

// APIRequest defines mock request for APIs
type APIRequest struct {
	// PathParams for the API
	PathParams map[string]string `yaml:"path_params" json:"path_params"`
	// QueryParams for the API
	QueryParams map[string]string `yaml:"query_params" json:"query_params"`
	// PostParams for the API
	PostParams map[string]string `yaml:"post_params" json:"post_params"`
	// Headers for mock request
	Headers map[string]string `yaml:"headers" json:"headers"`
	// Description for response optionally
	Description string `yaml:"description" json:"description"`
	// Contents for request optionally
	Contents string `yaml:"contents" json:"contents"`
	// ExampleContents sample for request optionally
	ExampleContents string `yaml:"example_contents" json:"example_contents"`
	// HTTPVersion version of http
	HTTPVersion string `yaml:"http_version" json:"http_version"`
	// AssertQueryParamsPattern for the API
	AssertQueryParamsPattern map[string]string `yaml:"assert_query_params_pattern" json:"assert_query_params_pattern"`
	// AssertPostParamsPattern for the API
	AssertPostParamsPattern map[string]string `yaml:"assert_post_params_pattern" json:"assert_post_params_pattern"`
	// AssertHeadersPattern for mock response
	AssertHeadersPattern map[string]string `yaml:"assert_headers_pattern" json:"assert_headers_pattern"`
	// AssertContentsPattern for request optionally
	AssertContentsPattern string `yaml:"assert_contents_pattern" json:"assert_contents_pattern"`
	// Assertions for validating response
	Assertions []string `yaml:"assertions" json:"assertions"`
	// Variables to set for templates
	Variables map[string]string `yaml:"variables" json:"variables"`
}

// AddAssertion helper method
func AddAssertion(assertions []string, assert string) []string {
	for _, next := range assertions {
		if assert == next {
			return assertions
		}
	}
	return append(assertions, assert)
}

// ContentType find content-type
func (r APIRequest) ContentType(defContentType string) string {
	for k, v := range r.Headers {
		if strings.ToUpper(k) == strings.ToUpper(ContentTypeHeader) {
			return fuzz.StripTypeTags(v)
		}
	}
	return defContentType
}

// AuthHeader finds AuthHeaderType
func (r APIRequest) AuthHeader() string {
	for k, v := range r.Headers {
		if strings.ToUpper(k) == "AUTHORIZATION" {
			return fuzz.StripTypeTags(v)
		}
	}
	return ""
}

// SanitizeRegexValue sanitizes (val string) string {
func SanitizeRegexValue(val string) string {
	if strings.HasPrefix(val, "__") || strings.HasPrefix(val, "(") {
		return fuzz.RandRegex(val)
	}
	return fuzz.StripTypeTags(val)
}

// BuildTemplateParams builds template
func (r APIRequest) BuildTemplateParams(
	req *http.Request,
	pathGroups map[string]string,
	inHeaders http.Header,
	overrides map[string]any,
) (templateParams map[string]any, queryParams map[string]string, postParams map[string]string, reqHeaders http.Header) {
	templateParams = make(map[string]any)
	queryParams = make(map[string]string)
	postParams = make(map[string]string)
	reqHeaders = make(http.Header)
	//for _, env := range os.Environ() {
	//	parts := strings.Split(env, "=")
	//	if len(parts) == 2 {
	//		templateParams[parts[0]] = parts[1]
	//	}
	//}

	for k, v := range r.Variables {
		templateParams[k] = fuzz.StripTypeTags(v)
	}
	for k, v := range r.PathParams {
		templateParams[k] = fuzz.StripTypeTags(v)
		queryParams[k] = fuzz.StripTypeTags(v)
	}
	for k, v := range r.AssertQueryParamsPattern {
		templateParams[k] = SanitizeRegexValue(v)
		queryParams[k] = SanitizeRegexValue(v)
	}
	for k, v := range r.QueryParams {
		templateParams[k] = fuzz.StripTypeTags(v)
		queryParams[k] = fuzz.StripTypeTags(v)
	}
	for k, v := range r.AssertPostParamsPattern {
		templateParams[k] = SanitizeRegexValue(v)
		postParams[k] = SanitizeRegexValue(v)
	}
	for k, v := range r.PostParams {
		templateParams[k] = fuzz.StripTypeTags(v)
		postParams[k] = fuzz.StripTypeTags(v)
	}
	for k, v := range r.AssertHeadersPattern {
		templateParams[k] = SanitizeRegexValue(v)
		reqHeaders.Set(k, SanitizeRegexValue(v))
	}
	for k, v := range r.Headers {
		templateParams[k] = fuzz.StripTypeTags(v)
		reqHeaders.Set(k, fuzz.StripTypeTags(v))
	}
	if req.URL != nil {
		for k, v := range req.URL.Query() {
			templateParams[k] = fuzz.StripTypeTags(v[0])
			queryParams[k] = fuzz.StripTypeTags(v[0])
		}
		for k, v := range req.PostForm {
			templateParams[k] = fuzz.StripTypeTags(v[0])
			postParams[k] = fuzz.StripTypeTags(v[0])
		}
	}
	for k, v := range req.Header {
		templateParams[k] = fuzz.StripTypeTags(v[0])
		reqHeaders.Set(k, fuzz.StripTypeTags(v[0]))
	}
	// Find any params for query params and path variables
	for k, v := range pathGroups {
		templateParams[k] = v
	}
	for k, v := range overrides {
		templateParams[k] = v
		queryParams[k] = fmt.Sprintf("%v", v)
	}
	for k := range inHeaders {
		reqHeaders.Set(k, inHeaders.Get(k))
	}

	return
}

// TargetHeader find header matching target
func (r APIRequest) TargetHeader() string {
	for k, v := range r.Headers {
		if strings.Contains(strings.ToUpper(k), "TARGET") {
			return fuzz.StripTypeTags(v)
		}
	}
	return ""
}

// Assert asserts response
func (r APIRequest) Assert(
	queryParams map[string]string,
	postParams map[string]string,
	reqHeaders http.Header,
	reqContents any,
	templateParams map[string]any) error {
	if reqContents != nil {
		templateParams["contents"] = reqContents
	}
	templateParams["headers"] = toFlatMap(reqHeaders)
	for k, v := range r.AssertQueryParamsPattern {
		actual := queryParams[k]
		if actual == "" {
			return fmt.Errorf("failed to find required query parameter '%s' with regex '%s'", k, v)
		}
		match, err := regexp.MatchString(v, actual)
		if err != nil {
			return fmt.Errorf("failed to fuzz required request query param '%s' with regex '%s' and actual value '%s' due to '%w'",
				k, v, actual, err)
		}
		if !match {
			return fmt.Errorf("didn't match required request query param '%s' with regex '%s' and actual value '%s'",
				k, v, actual)
		}
	}

	for k, v := range r.AssertPostParamsPattern {
		actual := postParams[k]
		if actual == "" {
			return fmt.Errorf("failed to find required post parameter '%s' with regex '%s'", k, v)
		}
		match, err := regexp.MatchString(v, actual)
		if err != nil {
			return fmt.Errorf("failed to fuzz required request post param '%s' with regex '%s' and actual value '%s' due to '%w'",
				k, v, actual, err)
		}
		if !match {
			return fmt.Errorf("didn't match required request post param '%s' with regex '%s' and actual value '%s'",
				k, v, actual)
		}
	}

	for k, v := range r.AssertHeadersPattern {
		actual := reqHeaders.Get(k)
		if actual == v {
			continue
		}
		if actual == "" {
			if log.IsLevelEnabled(log.DebugLevel) {
				//debug.PrintStack()
			}

			return fmt.Errorf("scenario-request %s failed to find required header '%s' with regex '%s' [headers: %v]",
				r.Description, k, v, reqHeaders)
		}
		match, err := regexp.MatchString(v, actual)
		if err != nil {
			return fmt.Errorf("scenario-request %s failed to fuzz required header '%s' with regex '%s' and actual value '%s' due to '%w'",
				r.Description, k, v, actual, err)
		}
		if !match {
			return fmt.Errorf("scenario-request %s didn't match required request header '%s' with regex '%s' and actual value '%s' [headers: %v]",
				r.Description, k, v, actual, reqHeaders)
		}
	}

	if r.AssertContentsPattern != "" {
		regex := make(map[string]string)
		err := json.Unmarshal([]byte(r.AssertContentsPattern), &regex)
		if err != nil {
			return fmt.Errorf("failed to unmarshal request '%s' regex due to %w", r.AssertContentsPattern, err)
		}
		err = fuzz.ValidateRegexMap(reqContents, regex)
		if err != nil {
			return fmt.Errorf("failed to validate request due to %w", err)
		}

		//if reqContents != nil {
		//	// Check if the pattern is JSON Schema
		//	if strings.HasPrefix(r.AssertContentsPattern, `{"$schema":`) {
		//		// Validate using JSON Schema
		//		return validateWithJSONSchema(r.AssertContentsPattern, reqContents)
		//	}
		//
		//	// Check if we need to do partial matching rather than exact structure
		//	if strings.Contains(r.AssertContentsPattern, "__partial__") {
		//		return validatePartialContent(r.AssertContentsPattern, reqContents)
		//	}
		//}
	}

	for _, assertion := range r.Assertions {
		assertion = normalizeAssertion(assertion)
		b, err := fuzz.ParseTemplate("", []byte(assertion), templateParams)
		if err != nil {
			return fmt.Errorf("failed to parse request assertion %s due to %w", assertion, err)
		}

		if string(b) != "true" {
			return fmt.Errorf("failed to assert request '%s' with value '%s', params: %v",
				assertion, b, templateParams)
		}
	}
	return nil
}

// AssertContentsPatternOrContent helper method
func (r APIRequest) AssertContentsPatternOrContent() string {
	if r.ExampleContents != "" {
		return r.ExampleContents
	}
	if r.AssertContentsPattern != "" {
		return r.AssertContentsPattern
	}
	return r.Contents
}

// APIResponse defines mock response for APIs
type APIResponse struct {
	// Headers for mock response
	Headers http.Header `yaml:"headers" json:"headers"`
	// Contents for request
	Contents string `yaml:"contents" json:"contents"`
	// ContentsFile for request
	ContentsFile string `yaml:"contents_file" json:"contents_file"`
	// Description for response optionally
	Description string `yaml:"description" json:"description"`
	// ExampleContents sample for response optionally
	ExampleContents string `yaml:"example_contents" json:"example_contents"`
	// StatusCode for response
	StatusCode int `yaml:"status_code" json:"status_code"`
	// HTTPVersion version of http
	HTTPVersion string `yaml:"http_version" json:"http_version"`
	// AddSharedVariables to set shared variables from response
	AddSharedVariables []string `yaml:"add_shared_variables" json:"add_shared_variables"`
	// DeleteSharedVariables to reset shared variables
	DeleteSharedVariables []string `yaml:"delete_shared_variables" json:"delete_shared_variables"`
	// AssertHeadersPattern for mock response
	AssertHeadersPattern map[string]string `yaml:"assert_headers_pattern" json:"assert_headers_pattern"`
	// AssertContentsPattern for request optionally
	AssertContentsPattern string `yaml:"assert_contents_pattern" json:"assert_contents_pattern"`
	// Assertions for validating response
	Assertions []string `yaml:"assertions" json:"assertions"`
}

// ContentType find content-type
func (r APIResponse) ContentType(defContentType string) string {
	for k, v := range r.Headers {
		if strings.ToUpper(k) == strings.ToUpper(ContentTypeHeader) {
			return fuzz.StripTypeTags(v[0])
		}
	}
	return defContentType
}

// Assert asserts response
func (r APIResponse) Assert(
	resHeaders http.Header,
	resContents any,
	templateParams map[string]any) error {
	if resContents != nil {
		templateParams["contents"] = resContents
	}
	templateParams["headers"] = toFlatMap(resHeaders)
	for k, v := range r.AssertHeadersPattern {
		actualHeader := resHeaders[k]
		if len(actualHeader) == 0 {
			return fmt.Errorf("failed to find required response header %s with regex %s", k, v)
		}
		match, err := regexp.MatchString(v, actualHeader[0])
		if err != nil {
			return fmt.Errorf("failed to fuzz required response header %s with regex %s and actual value %s due to %w",
				k, v, actualHeader[0], err)
		}
		if !match {
			return fmt.Errorf("didn't match required response header %s with regex %s and actual value %s",
				k, v, actualHeader[0])
		}
	}

	if r.AssertContentsPattern != "" && resContents != nil {
		regex := make(map[string]string)
		err := json.Unmarshal([]byte(r.AssertContentsPattern), &regex)
		if err != nil {
			return fmt.Errorf("failed to unmarshal response '%s' regex due to %w", r.AssertContentsPattern, err)
		}
		err = fuzz.ValidateRegexMap(resContents, regex)
		if err != nil {
			return fmt.Errorf("failed to validate response due to %w", err)
		}

		// Determine content type
		//contentType := r.ContentType("application/json")
		//
		//switch {
		//case strings.Contains(contentType, "json"):
		//	// Use JSON-specific validation
		//	return validateJSONResponse(r.AssertContentsPattern, resContents)
		//
		//case strings.Contains(contentType, "xml"):
		//	// Use XML-specific validation
		//	return validateXMLResponse(r.AssertContentsPattern, resContents)
		//
		//case strings.Contains(contentType, "text/plain"):
		//	// Use text validation
		//	return validateTextResponse(r.AssertContentsPattern, resContents)
		//
		//default:
		//	// Fall back to generic validation
		//	regex = make(map[string]string)
		//	err = json.Unmarshal([]byte(r.AssertContentsPattern), &regex)
		//	if err != nil {
		//		return fmt.Errorf("failed to unmarshal response '%s' regex due to %w",
		//			r.AssertContentsPattern, err)
		//	}
		//	return fuzz.ValidateRegexMap(resContents, regex)
		//}
	}

	for _, assertion := range r.Assertions {
		assertion = normalizeAssertion(assertion)
		b, err := fuzz.ParseTemplate("", []byte(assertion), templateParams)
		if err != nil {
			return fmt.Errorf("failed to parse assertion %s due to %w", assertion, err)
		}

		if string(b) != "true" {
			return fmt.Errorf("failed to assert response '%s' with value '%s', params: %v",
				assertion, b, templateParams)
		}
	}
	return nil
}

// AssertContentsPatternOrContent helper method
func (r APIResponse) AssertContentsPatternOrContent() string {
	if r.ExampleContents != "" {
		return r.ExampleContents
	}
	if r.AssertContentsPattern != "" {
		return r.AssertContentsPattern
	}
	return r.Contents
}

// APIVariables defines shared variables for APIs
type APIVariables struct {
	// Name of variable collection
	Name string `yaml:"name" json:"name"`
	// Variables to set for templates
	Variables map[string]string `yaml:"variables" json:"variables"`
}

func (v *APIVariables) Validate() error {
	if v.Name == "" {
		debug.PrintStack()
		return fmt.Errorf("api variables name is not specified")
	}
	return nil
}

// APIScenario defines mock scenario for APIs
type APIScenario struct {
	// Method for HTTP API
	Method MethodType `yaml:"method" json:"method"`
	// Name to uniquely identify the scenario
	Name string `yaml:"name" json:"name"`
	// Path for the API (excluding query params)
	Path string `yaml:"path" json:"path"`
	// BaseURL of remote server
	BaseURL string `yaml:"base_url" json:"base_url"`
	// Description of scenario
	Description string `yaml:"description" json:"description"`
	// Load next request before executing current scenario
	NextRequest string `yaml:"next_request" json:"next_request"`
	// Order of scenario
	Order int `yaml:"order" json:"order"`
	// Group of scenario
	Group string `yaml:"group" json:"group"`
	// Tags of scenario
	Tags []string `yaml:"tags" json:"tags"`
	// Predicate for the  scenario
	Predicate string `yaml:"predicate" json:"predicate"`
	// Variables File for the scenario
	VariablesFile string `yaml:"variables_file" json:"variables_file"`
	// Authentication for the API
	Authentication map[string]APIAuthorization `yaml:"authentication" json:"authentication"`
	// Request for the API
	Request APIRequest `yaml:"request" json:"request"`
	// Response for the API
	Response APIResponse `yaml:"response" json:"response"`
	// WaitMillisBeforeReply for response
	WaitBeforeReply time.Duration `yaml:"wait_before_reply" json:"wait_before_reply"`
	// StartTime of request
	StartTime time.Time `yaml:"start_time" json:"start_time"`
	// EndTime of request
	EndTime time.Time `yaml:"end_time" json:"end_time"`
	// RequestCount of request
	RequestCount uint64 `yaml:"-" json:"-"`
}

// ToKeyData converts scenario to key data
func (api *APIScenario) ToKeyData() *APIKeyData {
	rawPath := NormalizePath(api.Path, '/')
	if !strings.HasPrefix(rawPath, "/") {
		rawPath = "/" + rawPath
	}
	return &APIKeyData{
		Method:                   api.Method,
		Name:                     api.Name,
		Path:                     rawPath,
		Group:                    api.Group,
		Tags:                     api.Tags,
		Response:                 APIResponseKey{StatusCode: api.Response.StatusCode},
		Order:                    api.Order,
		Predicate:                api.Predicate,
		AssertQueryParamsPattern: api.Request.AssertQueryParamsPattern,
		AssertContentsPattern:    api.Request.AssertContentsPattern,
		AssertHeadersPattern:     api.Request.AssertHeadersPattern,
	}
}

// BuildScenarioFromHTTP helper method
func BuildScenarioFromHTTP(
	config *Configuration,
	prefix string,
	u *url.URL,
	method string,
	group string,
	reqHTTPVersion string,
	resHTTPVersion string,
	reqBody []byte,
	resBody []byte,
	queryParams map[string][]string,
	postParams map[string][]string,
	reqHeaders http.Header,
	reqContentType string,
	resHeaders http.Header,
	resContentType string,
	resStatus int,
	started time.Time,
	ended time.Time,
) (*APIScenario, error) {
	if u == nil {
		return nil, fmt.Errorf("url is not specified for building api scenario")
	}
	// Initialize headers if nil
	if reqHeaders == nil {
		reqHeaders = make(http.Header)
	}
	if resHeaders == nil {
		resHeaders = make(http.Header)
	}
	if queryParams == nil {
		queryParams = make(map[string][]string)
	}
	if postParams == nil {
		postParams = make(map[string][]string)
	}

	reqContentType = headerValue(reqHeaders, ContentTypeHeader, reqContentType)
	resContentType = headerValue(resHeaders, ContentTypeHeader, resContentType)

	dataTemplate := fuzz.NewDataTemplateRequest(true, 1, 1)
	matchReqContents, err := fuzz.UnmarshalArrayOrObjectAndExtractTypesAndMarshal(string(reqBody), dataTemplate)
	if err != nil {
		log.WithFields(log.Fields{
			"Path":   u,
			"Method": method,
			"Error":  err,
		}).Warnf("failed to unmarshal and extrate types for request")
	}
	matchResContents, err := fuzz.UnmarshalArrayOrObjectAndExtractTypesAndMarshal(string(resBody), dataTemplate)
	if err != nil {
		log.WithFields(log.Fields{
			"Path":   u,
			"Method": method,
			"Error":  err,
		}).Warnf("failed to unmarshal and extrate types for response")
	}

	reqAssertions := make([]string, 0)
	resAssertions := make([]string, 0)
	reqHeaderAssertions := make(map[string]string)
	if reqContentType != "" {
		reqAssertions = AddAssertion(reqAssertions, fmt.Sprintf(`VariableMatches headers.Content-Type %s`,
			reqContentType))
		reqHeaderAssertions[ContentTypeHeader] = reqContentType
	}
	respHeaderAssertions := make(map[string]string)
	if resContentType != "" {
		resAssertions = AddAssertion(resAssertions, fmt.Sprintf(`VariableMatches headers.Content-Type %s`,
			resContentType))
		respHeaderAssertions[ContentTypeHeader] = resContentType
	}
	path := u.Path
	if path == "" {
		path = "/"
	}
	scenario := &APIScenario{
		Method:         MethodType(method),
		Name:           headerValue(reqHeaders, MockScenarioHeader, ""),
		Path:           path,
		Group:          group,
		Authentication: make(map[string]APIAuthorization),
		Request: APIRequest{
			QueryParams:              make(map[string]string),
			PostParams:               make(map[string]string),
			Headers:                  make(map[string]string),
			Contents:                 fuzz.ReMarshalArrayOrObjectWithIndent(reqBody),
			ExampleContents:          fuzz.ReMarshalArrayOrObjectWithIndent(reqBody),
			HTTPVersion:              reqHTTPVersion,
			AssertQueryParamsPattern: make(map[string]string),
			AssertHeadersPattern:     reqHeaderAssertions,
			AssertContentsPattern:    matchReqContents,
			Assertions:               reqAssertions,
			Variables:                make(map[string]string),
		},
		Response: APIResponse{
			Headers:               resHeaders,
			Contents:              fuzz.ReMarshalArrayOrObjectWithIndent(resBody),
			ExampleContents:       fuzz.ReMarshalArrayOrObjectWithIndent(resBody),
			StatusCode:            resStatus,
			HTTPVersion:           resHTTPVersion,
			AssertHeadersPattern:  respHeaderAssertions,
			AssertContentsPattern: matchResContents,
			Assertions:            resAssertions,
			AddSharedVariables:    fuzz.ExtractTopPrimitiveAttributes(resBody, 5),
		},
	}
	if u.Scheme != "" && u.Host != "" {
		scenario.BaseURL = u.Scheme + "://" + u.Host
	}
	if scenario.Group == "" {
		scenario.Group = NormalizeGroup("", u.Path)
	}

	for k, v := range queryParams {
		if len(v) > 0 {
			scenario.Request.QueryParams[k] = fuzz.PrefixTypeExample + v[0]
			if config.AssertQueryParams(k) {
				scenario.Request.AssertQueryParamsPattern[k] = v[0]
			}
		}
	}
	for k, v := range postParams {
		if len(v) > 0 {
			scenario.Request.PostParams[k] = fuzz.PrefixTypeExample + v[0]
			if config.AssertQueryParams(k) {
				scenario.Request.AssertQueryParamsPattern[k] = v[0]
			}
		}
	}
	for k, v := range reqHeaders {
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
		scenario.addAWSHeaders()
	} else if authHeader != "" {
		scenario.addAuthHeaders()
	}
	if scenario.Name == "" {
		scenario.SetName(prefix + scenario.Group) // Request / Response are added
	}
	scenario.Tags = []string{scenario.Group}
	if scenario.Response.StatusCode >= 300 {
		scenario.Predicate = "{{NthRequest 2}}"
	} else {
		scenario.Predicate = "{{NthRequest 1}}"
	}
	scenario.Description = fmt.Sprintf("%s at %v for %s", time.Now().UTC(), prefix, u)

	scenario.StartTime = started.UTC()
	scenario.EndTime = ended.UTC()
	return scenario, nil
}

func (api *APIScenario) LoadFileVariables(apiVariables *APIVariables) {
	for k, v := range apiVariables.Variables {
		api.Request.Variables[k] = v
	}
}

func (api *APIScenario) addAuthHeaders() {
	api.Authentication["basicAuth"] = APIAuthorization{
		Type:   "http",
		Name:   AuthorizationHeader,
		In:     "header",
		Scheme: "basic",
	}
	api.Authentication["bearerAuth"] = APIAuthorization{
		Type:   "http",
		Name:   AuthorizationHeader,
		In:     "header",
		Scheme: "bearer",
		Format: "auth-scheme",
	}
}

func (api *APIScenario) addAWSHeaders() {
	api.Authentication["aws.auth.sigv4"] = APIAuthorization{
		Type:   "apiKey",
		Name:   AuthorizationHeader,
		In:     "header",
		Scheme: "x-amazon-apigateway-authtype",
		Format: "awsSigv4",
	}
	api.Authentication["smithy.scenario.httpApiKeyAuth"] = APIAuthorization{
		Type: "apiKey",
		Name: "x-scenario-key",
		In:   "header",
	}
	api.Authentication["bearerAuth"] = APIAuthorization{
		Type:   "http",
		Name:   AuthorizationHeader,
		In:     "header",
		Scheme: "bearer",
		Format: "JWT",
	}
}

// GetStartTime helper method
func (api *APIScenario) GetStartTime() time.Time {
	if !api.StartTime.IsZero() {
		return api.StartTime
	}
	return api.StartTime
}

// GetMillisTime helper method
func (api *APIScenario) GetMillisTime() int64 {
	return api.GetEndTime().UnixMilli() - api.GetStartTime().UnixMilli()
}

// GetEndTime helper method
func (api *APIScenario) GetEndTime() time.Time {
	if !api.EndTime.IsZero() {
		return api.EndTime
	}
	return api.EndTime
}

// HasURL helper method
func (api *APIScenario) HasURL() bool {
	return api.BaseURL != ""
}

// GetNetURL helper method
func (api *APIScenario) GetNetURL(u *url.URL) (*url.URL, error) {
	return api.GetURL(u.Scheme + "://" + u.Host)
}

// GetURL helper method
func (api *APIScenario) GetURL(defBase string) (u *url.URL, err error) {
	if api.BaseURL != "" {
		u, err = url.Parse(api.BaseURL)
	} else {
		u, err = url.Parse(defBase)
	}
	if err != nil {
		return nil, fmt.Errorf("scenario %s [%s] failed to parse base '%s' due to %s",
			api.Name, api.Request.Description, defBase, err)
	}
	params := url.Values{}
	for k, v := range api.Request.QueryParams {
		params.Add(k, v)
	}
	u, err = url.Parse(u.Scheme + "://" + u.Host + api.Path)
	if u != nil {
		u.RawQuery = params.Encode()
	}
	return
}

// String
func (api *APIScenario) String() string {
	return fmt.Sprintf("%s-%s-%s-%d", api.Method, api.Path, api.Name, api.Response.StatusCode)
}

// SafeName strips invalid characters
func (api *APIScenario) SafeName() string {
	return SanitizeNonAlphabet(api.Name, "")
}

// MethodPath helper method
func (api *APIScenario) MethodPath() string {
	return strings.ToLower(string(api.Method)) + "_" + SanitizeNonAlphabet(api.Path, "_")
}

// MethodPathTarget helper method
func (api *APIScenario) MethodPathTarget() string {
	return strings.ToLower(string(api.Method)) + "_" + SanitizeNonAlphabet(api.Path, "_") + // replace slashes
		"_" + strings.ToLower(api.Request.TargetHeader())
}

// BuildURL helper method
func (api *APIScenario) BuildURL(overrideBaseURL string) string {
	if overrideBaseURL == "" {
		overrideBaseURL = api.BaseURL
	}
	return overrideBaseURL + api.Path
}

// Digest of scenario
func (api *APIScenario) Digest() string {
	h := adler32.New()
	_, _ = h.Write([]byte(api.Method))
	_, _ = h.Write([]byte(api.Group))
	_, _ = h.Write([]byte(api.Path))
	_, _ = h.Write([]byte(api.Request.Contents))
	for k, v := range api.Request.AssertQueryParamsPattern {
		_, _ = h.Write([]byte(k))
		_, _ = h.Write([]byte(v))
	}
	for k, v := range api.Request.AssertHeadersPattern {
		_, _ = h.Write([]byte(k))
		_, _ = h.Write([]byte(v))
	}
	_, _ = h.Write([]byte(api.Request.AssertContentsPattern))
	_, _ = h.Write([]byte(api.Response.Contents))
	_, _ = h.Write([]byte(api.Response.ContentsFile))
	return fmt.Sprintf("%x", h.Sum(nil))
}

// Validate scenario
func (api *APIScenario) Validate() error {
	if api.Method == "" {
		return fmt.Errorf("method is not specified")
	}
	if api.Path == "" {
		debug.PrintStack()
		return fmt.Errorf("scenario path is not specified %s", api.BaseURL)
	}
	if len(api.Path) > 200 {
		return fmt.Errorf("path is too long %d", len(api.Path))
	}
	if matched, err := regexp.Match(`^[\w\d\.\-_\/\\:{}]+$`, []byte(api.Path)); err == nil && !matched {
		return fmt.Errorf("path is invalid with special characters '%s'", api.Path)
	}
	api.Path = NormalizePath(api.Path, '/')
	if !strings.HasPrefix(api.Path, "/") {
		api.Path = "/" + api.Path
	}
	if api.Name == "" {
		return fmt.Errorf("scenario name is not specified")
	}
	if len(api.Name) > 200 {
		return fmt.Errorf("scenario name is too long %d", len(api.Name))
	}
	api.Name = sanitizeSpecialChars(api.Name, "")
	if len(api.Response.Contents) > 1024*1024*1024 {
		return fmt.Errorf("contents is too long %d", len(api.Response.Contents))
	}
	return nil
}

// NormalPath normalizes path
func (api *APIScenario) NormalPath(sep uint8) string {
	return NormalizePath(api.Path, sep)
}

// SetName sets name
func (api *APIScenario) SetName(prefix string) {
	api.Name = api.BuildName(prefix)
}

// BuildName builds name
func (api *APIScenario) BuildName(prefix string) string {
	return fmt.Sprintf("%s%s-%d-%s", prefix, NormalizeDirPath(api.NormalName()), api.Response.StatusCode, api.Digest())
}

// NormalName normalizes name from path
func (api *APIScenario) NormalName() string {
	return NormalizePath(api.Path, '-')
}

// NormalizeDirPath normalizes dir path
func NormalizeDirPath(path string) string {
	path = NormalizePath(path, os.PathSeparator)
	ndx := strings.Index(path, ":")
	if ndx == -1 {
		ndx = strings.Index(path, "{")
	}
	if ndx > 1 {
		path = path[0 : ndx-1]
	} else if ndx == 0 {
		path = ""
	}
	return path
}

// NormalizePath normalizes path
func NormalizePath(path string, sepChar uint8) string {
	sep := fmt.Sprintf("%c", sepChar)
	if re, err := regexp.Compile(`[\/\\]+`); err == nil {
		path = re.ReplaceAllString(path, sep)
	}
	if len(path) < 2 {
		return path
	}

	from := 0
	to := len(path)

	if strings.HasPrefix(path, sep) {
		from = 1
	}
	if strings.HasSuffix(path, sep) {
		to = len(path) - 1
	}
	return path[from:to]
}

func reMatch(re string, str string) bool {
	re = fuzz.StripTypeTags(re)
	match, err := regexp.MatchString(re, str)
	if err != nil {
		return false
	}
	return match
}

func normalizeAssertion(assertion string) string {
	if !strings.HasPrefix(assertion, "{{") {
		parts := strings.Split(assertion, " ")
		var sb strings.Builder
		sb.WriteString("{{")
		for i, next := range parts {
			if i > 0 {
				if strings.HasPrefix(next, "\"") {
					sb.WriteString(fmt.Sprintf(` %s`, next))
				} else {
					sb.WriteString(fmt.Sprintf(` "%s"`, next))
				}
			} else {
				sb.WriteString(next)
			}
		}
		sb.WriteString("}}")
		assertion = sb.String()
	}
	return assertion
}

func toFlatMap(headers http.Header) map[string]string {
	flatHeaders := make(map[string]string)
	for k, v := range headers {
		flatHeaders[k] = v[0]
	}
	return flatHeaders
}

// sanitizeSpecialChars helper method
func sanitizeSpecialChars(name string, rep string) string {
	re := regexp.MustCompile(`[^\w\d-_\. ]`)
	return strings.TrimSpace(re.ReplaceAllString(name, rep))
}

// SanitizeNonAlphabet helper method
func SanitizeNonAlphabet(name string, rep string) string {
	if re, err := regexp.Compile(`[^a-zA-Z0-9_\-:]`); err == nil {
		name = re.ReplaceAllString(name, rep)
	}
	if re, err := regexp.Compile(rep + `+`); err == nil {
		name = re.ReplaceAllString(name, rep)
	}
	if re, err := regexp.Compile(rep + `$`); err == nil {
		name = re.ReplaceAllString(name, "")
	}
	return name
}

// NormalizeGroup normalizes group name
func NormalizeGroup(title string, path string) string {
	if title != "" {
		return title
	}
	n := strings.Index(path, "{")
	if n > 0 {
		path = path[0 : n-1]
	} else if n == 0 {
		path = ""
	}
	n = strings.Index(path, ":")
	if n > 0 {
		path = path[0 : n-1]
	} else if n == 0 {
		path = ""
	}
	if len(path) > 0 {
		path = path[1:]
	}
	group := strings.ReplaceAll(path, "/", "_")
	if group == "" {
		group = "root"
	}
	return group
}

func headerValue(headers http.Header, name string, defVal string) string {
	vals := headers.Get(name)
	if vals == "" {
		return defVal
	}
	return vals
}

// JSON validation with fuzzy matching
func validateJSONResponse(pattern string, actual any) error {
	// Parse pattern
	var patternObj interface{}
	if err := json.Unmarshal([]byte(pattern), &patternObj); err != nil {
		return fmt.Errorf("invalid JSON pattern: %w", err)
	}

	// Validate structure matches with fuzzy type checking
	return validateJSONStructure(patternObj, actual, "")
}

// Example implementation for validateJSONStructure
func validateJSONStructure(pattern, actual interface{}, path string) error {
	switch p := pattern.(type) {
	case map[string]interface{}:
		// Pattern is an object, actual should also be an object
		actualMap, ok := actual.(map[string]interface{})
		if !ok {
			return fmt.Errorf("validateJSONStructure at %s: expected object %v, got %T", path, pattern, actual)
		}

		// Check each field in the pattern
		for key, patternValue := range p {
			actualValue, exists := actualMap[key]
			if !exists {
				// Field doesn't exist in actual
				return fmt.Errorf("validateJSONStructure at %s: missing field %s", path, key)
			}

			// Recursively validate this field
			fieldPath := path
			if fieldPath == "" {
				fieldPath = key
			} else {
				fieldPath = fieldPath + "." + key
			}

			if err := validateJSONStructure(patternValue, actualValue, fieldPath); err != nil {
				return err
			}
		}

	case []interface{}:
		// Pattern is an array, actual should also be an array
		actualArray, ok := actual.([]interface{})
		if !ok {
			return fmt.Errorf("validateJSONStructure at %s: expected array, got %T", path, actual)
		}

		if len(p) > 0 && len(actualArray) > 0 {
			// Validate first item against pattern
			patternItem := p[0]
			for i, actualItem := range actualArray {
				itemPath := fmt.Sprintf("%s[%d]", path, i)
				if err := validateJSONStructure(patternItem, actualItem, itemPath); err != nil {
					return err
				}
			}
		}

	case string:
		// For strings, check if it's a special pattern like __string__ or __number__
		if strings.HasPrefix(p, "__") && strings.HasSuffix(p, "__") {
			typePattern := p[2 : len(p)-2]
			switch typePattern {
			case "string":
				if _, ok := actual.(string); !ok {
					return fmt.Errorf("validateJSONStructure at %s: expected string, got %T", path, actual)
				}
			case "number":
				switch actual.(type) {
				case float64, int, int64:
					// Valid numeric types
				default:
					return fmt.Errorf("validateJSONStructure at %s: expected number, got %T", path, actual)
				}
			case "boolean":
				if _, ok := actual.(bool); !ok {
					return fmt.Errorf("validateJSONStructure at %s: expected boolean, got %T", path, actual)
				}
			}
		} else {
			// Regular string, exact match
			if fmt.Sprintf("%v", actual) != p {
				return fmt.Errorf("validateJSONStructure at %s: expected %v, got %v", path, p, actual)
			}
		}

	default:
		// For other primitives, check exact equality
		if fmt.Sprintf("%v", actual) != fmt.Sprintf("%v", p) {
			return fmt.Errorf("validateJSONStructure at %s: expected %v, got %v", path, p, actual)
		}
	}

	return nil
}

// validateXMLResponse validates XML response against a pattern
func validateXMLResponse(pattern string, actual any) error {
	actualStr, ok := actual.(string)
	if !ok {
		// Try to convert to string if it's not already
		actualStr = fmt.Sprintf("%v", actual)
	}

	// Parse XML
	actualDoc := etree.NewDocument()
	if err := actualDoc.ReadFromString(actualStr); err != nil {
		return fmt.Errorf("invalid XML in response: %w", err)
	}

	patternDoc := etree.NewDocument()
	if err := patternDoc.ReadFromString(pattern); err != nil {
		return fmt.Errorf("invalid XML in pattern: %w", err)
	}

	// Validate structure and content
	return validateXMLElement(patternDoc.Root(), actualDoc.Root(), "")
}

// validateXMLElement recursively validates XML elements
func validateXMLElement(pattern, actual *etree.Element, path string) error {
	if pattern == nil || actual == nil {
		return fmt.Errorf("at %s: nil element", path)
	}

	// Check element name
	if pattern.Tag != actual.Tag {
		return fmt.Errorf("at %s: expected tag <%s>, got <%s>", path, pattern.Tag, actual.Tag)
	}

	currentPath := path
	if currentPath == "" {
		currentPath = pattern.Tag
	} else {
		currentPath = currentPath + "/" + pattern.Tag
	}

	// Check attributes if the pattern has __attr__ markers
	for _, attr := range pattern.Attr {
		if strings.HasPrefix(attr.Value, "__") && strings.HasSuffix(attr.Value, "__") {
			// This is a type pattern, validate the actual attribute exists
			actualAttr := actual.SelectAttr(attr.Key)
			if actualAttr == nil {
				return fmt.Errorf("at %s: missing attribute %s", currentPath, attr.Key)
			}

			// Validate type
			typePattern := attr.Value[2 : len(attr.Value)-2]
			switch typePattern {
			case "string":
				// Any string is valid
			case "number":
				if _, err := strconv.ParseFloat(actualAttr.Value, 64); err != nil {
					return fmt.Errorf("at %s: attribute %s should be a number, got %s",
						currentPath, attr.Key, actualAttr.Value)
				}
			case "boolean":
				if actualAttr.Value != "true" && actualAttr.Value != "false" {
					return fmt.Errorf("at %s: attribute %s should be a boolean, got %s",
						currentPath, attr.Key, actualAttr.Value)
				}
			}
		} else if attr.Value != "" {
			// Exact match required
			actualAttr := actual.SelectAttr(attr.Key)
			if actualAttr == nil || actualAttr.Value != attr.Value {
				return fmt.Errorf("at %s: attribute %s expected value %s, got %v",
					currentPath, attr.Key, attr.Value, actualAttr)
			}
		}
	}

	// Check text content if pattern has text
	if pattern.Text() != "" {
		if strings.HasPrefix(pattern.Text(), "__") && strings.HasSuffix(pattern.Text(), "__") {
			// Type pattern for content
			typePattern := pattern.Text()[2 : len(pattern.Text())-2]
			switch typePattern {
			case "string":
				// Any string is valid
			case "number":
				if _, err := strconv.ParseFloat(actual.Text(), 64); err != nil {
					return fmt.Errorf("at %s: text content should be a number, got %s",
						currentPath, actual.Text())
				}
			case "boolean":
				if actual.Text() != "true" && actual.Text() != "false" {
					return fmt.Errorf("at %s: text content should be a boolean, got %s",
						currentPath, actual.Text())
				}
			}
		} else if pattern.Text() != actual.Text() {
			// Exact text match required
			return fmt.Errorf("at %s: expected text content %s, got %s",
				currentPath, pattern.Text(), actual.Text())
		}
	}

	// Check child elements recursively
	patternChildren := pattern.ChildElements()
	actualChildren := actual.ChildElements()

	if len(patternChildren) > 0 {
		// If pattern has one child with a name ending with [], it's a repeating element pattern
		if len(patternChildren) == 1 && strings.HasSuffix(patternChildren[0].Tag, "[]") {
			baseTag := strings.TrimSuffix(patternChildren[0].Tag, "[]")
			patternChild := patternChildren[0]
			patternChild.Tag = baseTag // Fix the tag for comparison

			// Validate each matching actual child against the pattern
			for i, actualChild := range actualChildren {
				if actualChild.Tag == baseTag {
					childPath := fmt.Sprintf("%s/%s[%d]", currentPath, baseTag, i)
					if err := validateXMLElement(patternChild, actualChild, childPath); err != nil {
						return err
					}
				}
			}
		} else {
			// Regular child elements - match by tag name
			for _, patternChild := range patternChildren {
				found := false
				for _, actualChild := range actualChildren {
					if patternChild.Tag == actualChild.Tag {
						found = true
						childPath := currentPath + "/" + patternChild.Tag
						if err := validateXMLElement(patternChild, actualChild, childPath); err != nil {
							return err
						}
						break
					}
				}

				if !found {
					return fmt.Errorf("at %s: missing child element <%s>",
						currentPath, patternChild.Tag)
				}
			}
		}
	}

	return nil
}

// validateTextResponse validates text response against a pattern
func validateTextResponse(pattern string, actual any) error {
	actualStr, ok := actual.(string)
	if !ok {
		// Try to convert to string if it's not already
		actualStr = fmt.Sprintf("%v", actual)
	}

	// If pattern is a regex (starts with ^ or contains special chars)
	if strings.HasPrefix(pattern, "^") || strings.Contains(pattern, "(?") {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return fmt.Errorf("invalid regex pattern: %w", err)
		}

		if !re.MatchString(actualStr) {
			return fmt.Errorf("text doesn't match pattern '%s': %s", pattern, actualStr)
		}
		return nil
	}

	// If pattern contains wildcards like * or ?
	if strings.Contains(pattern, "*") || strings.Contains(pattern, "?") {
		matched, err := filepath.Match(pattern, actualStr)
		if err != nil {
			return fmt.Errorf("invalid glob pattern: %w", err)
		}

		if !matched {
			return fmt.Errorf("text doesn't match pattern '%s': %s", pattern, actualStr)
		}
		return nil
	}

	// Otherwise, do exact string matching
	if pattern != actualStr {
		return fmt.Errorf("expected text '%s', got '%s'", pattern, actualStr)
	}

	return nil
}

// validateWithJSONSchema validates content against a JSON Schema
func validateWithJSONSchema(schemaStr string, content any) error {
	// Create schema loader from the string
	schemaLoader := gojsonschema.NewStringLoader(schemaStr)

	// Convert content to JSON and create document loader
	contentJSON, err := json.Marshal(content)
	if err != nil {
		return fmt.Errorf("failed to marshal content for validation: %w", err)
	}
	contentLoader := gojsonschema.NewBytesLoader(contentJSON)

	// Validate
	result, err := gojsonschema.Validate(schemaLoader, contentLoader)
	if err != nil {
		return fmt.Errorf("schema validation error: %w", err)
	}

	if !result.Valid() {
		// Collect validation errors
		var errMsgs []string
		for _, err := range result.Errors() {
			errMsgs = append(errMsgs, err.String())
		}
		return fmt.Errorf("JSON Schema validation failed: %s", strings.Join(errMsgs, ", "))
	}

	return nil
}

// validatePartialContent validates that content contains at least the required fields
func validatePartialContent(patternStr string, content any) error {
	// Remove the __partial__ marker
	patternStr = strings.Replace(patternStr, "__partial__", "", -1)

	var pattern map[string]interface{}
	if err := json.Unmarshal([]byte(patternStr), &pattern); err != nil {
		return fmt.Errorf("invalid partial match pattern: %w", err)
	}

	contentMap, ok := content.(map[string]interface{})
	if !ok {
		return fmt.Errorf("content must be an object for partial matching")
	}

	// Check if all required fields from pattern exist in content
	return validatePartialObject(pattern, contentMap, "")
}

// validatePartialObject recursively validates that all fields in pattern exist in content
func validatePartialObject(pattern, content map[string]interface{}, path string) error {
	for k, v := range pattern {
		actualValue, exists := content[k]
		if !exists {
			return fmt.Errorf("validatePartialObject at %s: missing required field %s", path, k)
		}

		fieldPath := path
		if fieldPath == "" {
			fieldPath = k
		} else {
			fieldPath = fieldPath + "." + k
		}

		switch typedVal := v.(type) {
		case map[string]interface{}:
			actualMap, ok := actualValue.(map[string]interface{})
			if !ok {
				return fmt.Errorf("validatePartialObject at %s: expected object %v, got %T", fieldPath, v, actualValue)
			}

			if err := validatePartialObject(typedVal, actualMap, fieldPath); err != nil {
				return err
			}

		case []interface{}:
			actualArray, ok := actualValue.([]interface{})
			if !ok {
				return fmt.Errorf("validatePartialObject at %s: expected array, got %T", fieldPath, actualValue)
			}

			if len(typedVal) > 0 && len(actualArray) > 0 {
				// First item in pattern array is used as template for array items
				patternItem := typedVal[0]

				switch patternItemTyped := patternItem.(type) {
				case map[string]interface{}:
					for i, actualItem := range actualArray {
						actualItemMap, ok := actualItem.(map[string]interface{})
						if !ok {
							return fmt.Errorf("validatePartialObject at %s[%d]: expected object, got %T",
								fieldPath, i, actualItem)
						}

						if err := validatePartialObject(patternItemTyped, actualItemMap,
							fmt.Sprintf("%s[%d]", fieldPath, i)); err != nil {
							return fmt.Errorf("failed to assert parital obj %s[%d]: %s", fieldPath, i, err)
						}
					}
				}
			}

		case string:
			// Check for type pattern
			if strings.HasPrefix(typedVal, "__") && strings.HasSuffix(typedVal, "__") {
				typePattern := typedVal[2 : len(typedVal)-2]
				if typePattern == "string" {
					if _, ok := actualValue.(string); !ok {
						return fmt.Errorf("validatePartialObject at %s: expected string, got %T", fieldPath, actualValue)
					}
				} else if typePattern == "number" {
					switch actualValue.(type) {
					case float64, int, int64:
						// Valid numeric types
					default:
						return fmt.Errorf("validatePartialObject at %s: expected number, got %T", fieldPath, actualValue)
					}
				} else if typePattern == "boolean" {
					if _, ok := actualValue.(bool); !ok {
						return fmt.Errorf("validatePartialObject at %s: expected boolean, got %T", fieldPath, actualValue)
					}
				}
			}
		}
	}

	return nil
}
