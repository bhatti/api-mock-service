package types

import (
	"encoding/json"
	"fmt"
	"github.com/bhatti/api-mock-service/internal/fuzz"
	log "github.com/sirupsen/logrus"
	"hash/adler32"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"runtime/debug"
	"strconv"
	"strings"
	"time"
)

// MockRecordMode header
const MockRecordMode = "X-Mock-Record"

// MockPlayback header
const MockPlayback = "X-Mock-Playback"

// MockRecordModeDisabled disabled value
const MockRecordModeDisabled = "false"

// MockRecordModeEnabled enabled value
const MockRecordModeEnabled = "true"

// MockURL header
const MockURL = "X-Mock-Url"

// MockScenarioHeader header
const MockScenarioHeader = "X-Mock-Scenario"

// MockScenarioPath header
const MockScenarioPath = "X-Mock-Path"

// ContentTypeHeader header
const ContentTypeHeader = "Content-Type"

// AuthorizationHeader constant
const AuthorizationHeader = "Authorization"

// MockRequestCount header
const MockRequestCount = "X-Mock-Request-Count"

// MockResponseStatus header
const MockResponseStatus = "X-Mock-Response-Status"

// MockWaitBeforeReply header
const MockWaitBeforeReply = "X-Mock-Wait-Before-Reply"

// ScenarioExt extension
const ScenarioExt = ".yaml"

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
	// Headers for mock response
	Headers map[string]string `yaml:"headers" json:"headers"`
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
	inHeaders map[string][]string,
	overrides map[string]any,
) (templateParams map[string]any, queryParams map[string]string, postParams map[string]string, reqHeaders map[string][]string) {
	templateParams = make(map[string]any)
	queryParams = make(map[string]string)
	postParams = make(map[string]string)
	reqHeaders = make(map[string][]string)
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
		reqHeaders[k] = []string{SanitizeRegexValue(v)}
	}
	for k, v := range r.Headers {
		templateParams[k] = fuzz.StripTypeTags(v)
		reqHeaders[k] = []string{fuzz.StripTypeTags(v)}
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
		reqHeaders[k] = []string{fuzz.StripTypeTags(v[0])}
	}
	// Find any params for query params and path variables
	for k, v := range pathGroups {
		templateParams[k] = v
	}
	for k, v := range overrides {
		templateParams[k] = v
		queryParams[k] = fmt.Sprintf("%v", v)
	}
	for k, v := range inHeaders {
		reqHeaders[k] = v
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
	reqHeaders map[string][]string,
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
		actual := reqHeaders[k]
		if len(actual) == 0 {
			return fmt.Errorf("failed to find required request header '%s' with regex '%s'", k, v)
		}
		match, err := regexp.MatchString(v, actual[0])
		if err != nil {
			return fmt.Errorf("failed to fuzz required request header '%s' with regex '%s' and actual value '%s' due to '%w'",
				k, v, actual[0], err)
		}
		if !match {
			return fmt.Errorf("didn't match required request header '%s' with regex '%s' and actual value '%s'",
				k, v, actual[0])
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
	Headers map[string][]string `yaml:"headers" json:"headers"`
	// Contents for request
	Contents string `yaml:"contents" json:"contents"`
	// ContentsFile for request
	ContentsFile string `yaml:"contents_file" json:"contents_file"`
	// ExampleContents sample for response optionally
	ExampleContents string `yaml:"example_contents" json:"example_contents"`
	// StatusCode for response
	StatusCode int `yaml:"status_code" json:"status_code"`
	// HTTPVersion version of http
	HTTPVersion string `yaml:"http_version" json:"http_version"`
	// SetVariables to set shared variables from response
	SetVariables []string `yaml:"set_variables" json:"set_variables"`
	// UnsetVariables to reset shared variables
	UnsetVariables []string `yaml:"unset_variables" json:"unset_variables"`
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
	resHeaders map[string][]string,
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

	if r.AssertContentsPattern != "" {
		regex := make(map[string]string)
		err := json.Unmarshal([]byte(r.AssertContentsPattern), &regex)
		if err != nil {
			return fmt.Errorf("failed to unmarshal response '%s' regex due to %w", r.AssertContentsPattern, err)
		}
		err = fuzz.ValidateRegexMap(resContents, regex)
		if err != nil {
			return fmt.Errorf("failed to validate response due to %w", err)
		}
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
	// Order of scenario
	Order int `yaml:"order" json:"order"`
	// Group of scenario
	Group string `yaml:"group" json:"group"`
	// Tags of scenario
	Tags []string `yaml:"tags" json:"tags"`
	// Predicate for the request
	Predicate string `yaml:"predicate" json:"predicate"`
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
	reqHeaders map[string][]string,
	reqContentType string,
	resHeaders map[string][]string,
	resContentType string,
	resStatus int,
	started time.Time,
	ended time.Time,
) (*APIScenario, error) {
	if u == nil {
		return nil, fmt.Errorf("url is not specified for building api scenario")
	}
	if queryParams == nil {
		queryParams = make(map[string][]string)
	}
	if postParams == nil {
		postParams = make(map[string][]string)
	}
	if reqHeaders == nil {
		reqHeaders = make(map[string][]string)
	}
	if resHeaders == nil {
		resHeaders = make(map[string][]string)
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
	resAssertions := []string{
		`ResponseTimeMillisLE 5000`,
		fmt.Sprintf(`ResponseStatusMatches %d`, resStatus),
	}
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
		BaseURL:        u.Scheme + "://" + u.Host,
		Group:          group,
		Authentication: make(map[string]APIAuthorization),
		Request: APIRequest{
			QueryParams:              make(map[string]string),
			PostParams:               make(map[string]string),
			Headers:                  make(map[string]string),
			Contents:                 string(reqBody),
			ExampleContents:          string(reqBody),
			HTTPVersion:              reqHTTPVersion,
			AssertQueryParamsPattern: make(map[string]string),
			AssertHeadersPattern:     reqHeaderAssertions,
			AssertContentsPattern:    matchReqContents,
			Assertions:               reqAssertions,
			Variables:                make(map[string]string),
		},
		Response: APIResponse{
			Headers:               resHeaders,
			Contents:              string(resBody),
			ExampleContents:       string(resBody),
			StatusCode:            resStatus,
			HTTPVersion:           resHTTPVersion,
			AssertHeadersPattern:  respHeaderAssertions,
			AssertContentsPattern: matchResContents,
			Assertions:            resAssertions,
			SetVariables:          fuzz.ExtractTopPrimitiveAttributes(resBody, 5),
		},
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
		scenario.Authentication["aws.auth.sigv4"] = APIAuthorization{
			Type:   "apiKey",
			Name:   AuthorizationHeader,
			In:     "header",
			Scheme: "x-amazon-apigateway-authtype",
			Format: "awsSigv4",
		}
		scenario.Authentication["smithy.scenario.httpApiKeyAuth"] = APIAuthorization{
			Type: "apiKey",
			Name: "x-scenario-key",
			In:   "header",
		}
		scenario.Authentication["bearerAuth"] = APIAuthorization{
			Type:   "http",
			Name:   AuthorizationHeader,
			In:     "header",
			Scheme: "bearer",
			Format: "JWT",
		}
	} else if authHeader != "" {
		scenario.Authentication["basicAuth"] = APIAuthorization{
			Type:   "http",
			Name:   AuthorizationHeader,
			In:     "header",
			Scheme: "basic",
		}
		scenario.Authentication["bearerAuth"] = APIAuthorization{
			Type:   "http",
			Name:   AuthorizationHeader,
			In:     "header",
			Scheme: "bearer",
			Format: "auth-scheme",
		}
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
		return nil, fmt.Errorf("failed to parse base '%s' due to %s", defBase, err)
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
	return string(api.Method) + api.Name + api.Group + api.Path
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
	h.Write([]byte(api.Method))
	h.Write([]byte(api.Group))
	h.Write([]byte(api.Path))
	h.Write([]byte(api.Request.Contents))
	for k, v := range api.Request.AssertQueryParamsPattern {
		h.Write([]byte(k))
		h.Write([]byte(v))
	}
	for k, v := range api.Request.AssertHeadersPattern {
		h.Write([]byte(k))
		h.Write([]byte(v))
	}
	h.Write([]byte(api.Request.AssertContentsPattern))
	h.Write([]byte(api.Response.Contents))
	h.Write([]byte(api.Response.ContentsFile))
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

func toFlatMap(headers map[string][]string) map[string]string {
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

// BuildTestScenario helper method
func BuildTestScenario(method MethodType, name string, path string, n int) *APIScenario {
	return &APIScenario{
		Method:      method,
		Name:        name,
		Path:        path,
		Group:       path,
		Description: name,
		Request: APIRequest{
			HTTPVersion:              "1.1",
			QueryParams:              make(map[string]string),
			PostParams:               make(map[string]string),
			Headers:                  make(map[string]string),
			AssertQueryParamsPattern: map[string]string{"a": `\d+`, "b": "abc"},
			AssertHeadersPattern: map[string]string{
				ContentTypeHeader: "application/json",
				"ETag":            `\d{3}`,
			},
		},
		Response: APIResponse{
			HTTPVersion: "1.1",
			Headers: map[string][]string{
				"ETag":            {strconv.Itoa(n)},
				ContentTypeHeader: {"application/json"},
			},
			Contents:   "test body",
			StatusCode: 200,
		},
		WaitBeforeReply: time.Duration(1) * time.Second,
	}
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

func headerValue(headers map[string][]string, name string, defVal string) string {
	vals := headers[name]
	if len(vals) == 0 {
		return defVal
	}
	return vals[0]
}
