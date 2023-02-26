package types

import (
	"crypto/sha256"
	"fmt"
	"github.com/bhatti/api-mock-service/internal/fuzz"
	"os"
	"regexp"
	"strings"
	"time"
)

// MockRecordMode header
const MockRecordMode = "X-Mock-Record"

// MockRecordModeDisabled disabled value
const MockRecordModeDisabled = "false"

// MockRecordModeEnabled enabled value
const MockRecordModeEnabled = "true"

// MockURL header
const MockURL = "X-Mock-Url"

// MockScenarioName header
const MockScenarioName = "X-Mock-Scenario"

// MockScenarioPath header
const MockScenarioPath = "X-Mock-Path"

// ContentTypeHeader header
const ContentTypeHeader = "Content-Type"

// MockRequestCount header
const MockRequestCount = "X-Mock-Request-Count"

// MockResponseStatus header
const MockResponseStatus = "X-Mock-Response-Status"

// MockWaitBeforeReply header
const MockWaitBeforeReply = "X-Mock-Wait-Before-Reply"

// ScenarioExt extension
const ScenarioExt = ".scr"

// MockAuthorization defines mock auth parameters
type MockAuthorization struct {
	Type   string `json:"type,omitempty" yaml:"type,omitempty"`
	Name   string `json:"name,omitempty" yaml:"name,omitempty"`
	In     string `json:"in,omitempty" yaml:"in,omitempty"`
	Format string `json:"format,omitempty" yaml:"format,omitempty"`
	Scheme string `json:"scheme,omitempty" yaml:"scheme,omitempty"`
	URL    string `json:"url,omitempty" yaml:"url,omitempty"`
}

// MockHTTPRequest defines mock request for APIs
type MockHTTPRequest struct {
	// AssertQueryParamsPattern for the API
	AssertQueryParamsPattern map[string]string `yaml:"assert_query_params_pattern" json:"assert_query_params_pattern"`
	// AssertHeadersPattern for mock response
	AssertHeadersPattern map[string]string `yaml:"assert_headers_pattern" json:"assert_headers_pattern"`
	// AssertContentsPattern for request optionally
	AssertContentsPattern string `yaml:"assert_contents_pattern" json:"assert_contents_pattern"`
	// PathParams sample for the API
	PathParams map[string]string `yaml:"path_params" json:"path_params"`
	// QueryParams sample for the API
	QueryParams map[string]string `yaml:"query_params" json:"query_params"`
	// Headers for mock response
	Headers map[string]string `yaml:"headers" json:"headers"`
	// Contents for request optionally
	Contents string `yaml:"contents" json:"contents"`
	// ExampleContents sample for request optionally
	ExampleContents string `yaml:"example_contents" json:"example_contents"`
}

// ContentType find content-type
func (r MockHTTPRequest) ContentType(defContentType string) string {
	for k, v := range r.Headers {
		if strings.ToUpper(k) == strings.ToUpper(ContentTypeHeader) {
			return fuzz.StripTypeTags(v)
		}
	}
	return defContentType
}

// TargetHeader find header matching target
func (r MockHTTPRequest) TargetHeader() string {
	for k, v := range r.Headers {
		if strings.Contains(strings.ToUpper(k), "TARGET") {
			return fuzz.StripTypeTags(v)
		}
	}
	return ""
}

// AssertContentsPatternOrContent helper method
func (r MockHTTPRequest) AssertContentsPatternOrContent() string {
	if r.ExampleContents != "" {
		return r.ExampleContents
	}
	if r.AssertContentsPattern != "" {
		return r.AssertContentsPattern
	}
	return r.Contents
}

// MockHTTPResponse defines mock response for APIs
type MockHTTPResponse struct {
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
	// AssertHeadersPattern for mock response
	AssertHeadersPattern map[string]string `yaml:"assert_headers_pattern" json:"assert_headers_pattern"`
	// AssertContentsPattern for request optionally
	AssertContentsPattern string `yaml:"assert_contents_pattern" json:"assert_contents_pattern"`
	// PipeProperties to extract properties from response
	PipeProperties []string `yaml:"pipe_properties" json:"pipe_properties"`
	// Assertions for validating response
	Assertions []string `yaml:"assertions" json:"assertions"`
}

// ContentType find content-type
func (r MockHTTPResponse) ContentType(defContentType string) string {
	for k, v := range r.Headers {
		if strings.ToUpper(k) == strings.ToUpper(ContentTypeHeader) {
			return fuzz.StripTypeTags(v[0])
		}
	}
	return defContentType
}

// AssertContentsPatternOrContent helper method
func (r MockHTTPResponse) AssertContentsPatternOrContent() string {
	if r.ExampleContents != "" {
		return r.ExampleContents
	}
	if r.AssertContentsPattern != "" {
		return r.AssertContentsPattern
	}
	return r.Contents
}

// MockScenario defines mock scenario for APIs
type MockScenario struct {
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
	Authentication map[string]MockAuthorization `yaml:"authentication" json:"authentication"`
	// Request for the API
	Request MockHTTPRequest `yaml:"request" json:"request"`
	// Response for the API
	Response MockHTTPResponse `yaml:"response" json:"response"`
	// WaitMillisBeforeReply for response
	WaitBeforeReply time.Duration `yaml:"wait_before_reply" json:"wait_before_reply"`
	// RequestCount of request
	RequestCount uint64 `yaml:"-" json:"-"`
}

// ToKeyData converts scenario to key data
func (ms *MockScenario) ToKeyData() *MockScenarioKeyData {
	rawPath := NormalizePath(ms.Path, '/')
	if !strings.HasPrefix(rawPath, "/") {
		rawPath = "/" + rawPath
	}
	return &MockScenarioKeyData{
		Method:                   ms.Method,
		Name:                     ms.Name,
		Path:                     rawPath,
		Group:                    ms.Group,
		Tags:                     ms.Tags,
		Order:                    ms.Order,
		Predicate:                ms.Predicate,
		AssertQueryParamsPattern: ms.Request.AssertQueryParamsPattern,
		AssertContentsPattern:    ms.Request.AssertContentsPattern,
		AssertHeadersPattern:     ms.Request.AssertHeadersPattern,
	}
}

// String
func (ms *MockScenario) String() string {
	return string(ms.Method) + ms.Name + ms.Group + ms.Path
}

// SafeName strips invalid characters
func (ms *MockScenario) SafeName() string {
	if regexp, err := regexp.Compile(`[^a-zA-Z0-9_:]`); err == nil {
		return regexp.ReplaceAllString(ms.Name, "")
	}
	return ms.Name
}

// MethodPath helper method
func (ms *MockScenario) MethodPath() string {
	return strings.ToLower(string(ms.Method)) + "_" + strings.ReplaceAll(ms.Path, "/", "_")
}

// MethodPathTarget helper method
func (ms *MockScenario) MethodPathTarget() string {
	return strings.ToLower(string(ms.Method)) + "_" + strings.ReplaceAll(ms.Path, "/", "_") +
		"_" + strings.ToLower(ms.Request.TargetHeader())
}

// BuildURL helper method
func (ms *MockScenario) BuildURL(overrideBaseURL string) string {
	if overrideBaseURL == "" {
		overrideBaseURL = ms.BaseURL
	}
	return overrideBaseURL + ms.Path
}

// Digest of scenario
func (ms *MockScenario) Digest() string {
	h := sha256.New()
	h.Write([]byte(ms.Method))
	h.Write([]byte(ms.Group))
	h.Write([]byte(ms.Path))
	for k, v := range ms.Request.AssertQueryParamsPattern {
		h.Write([]byte(k))
		h.Write([]byte(v))
	}
	for k, v := range ms.Request.AssertHeadersPattern {
		h.Write([]byte(k))
		h.Write([]byte(v))
	}
	h.Write([]byte(ms.Request.AssertContentsPattern))
	h.Write([]byte(ms.Response.Contents))
	h.Write([]byte(ms.Response.ContentsFile))
	return fmt.Sprintf("%x", h.Sum(nil))
}

// Validate scenario
func (ms *MockScenario) Validate() error {
	if ms.Method == "" {
		return fmt.Errorf("method is not specified")
	}
	if ms.Path == "" {
		return fmt.Errorf("path is not specified")
	}
	if len(ms.Path) > 200 {
		return fmt.Errorf("path is too long %d", len(ms.Path))
	}
	if matched, err := regexp.Match(`^[\w\d\.\-_\/\\:{}]+$`, []byte(ms.Path)); err == nil && !matched {
		return fmt.Errorf("path is invalid with special characters '%s'", ms.Path)
	}
	ms.Path = NormalizePath(ms.Path, '/')
	if !strings.HasPrefix(ms.Path, "/") {
		ms.Path = "/" + ms.Path
	}
	if ms.Name == "" {
		return fmt.Errorf("scenario name is not specified")
	}
	if len(ms.Name) > 200 {
		return fmt.Errorf("scenario name is too long %d", len(ms.Name))
	}
	if matched, err := regexp.Match(`^[\w\d-_]+\.?[\w\d-_]+$`, []byte(ms.Name)); err == nil && !matched {
		return fmt.Errorf("scenario name is invalid with special characters %s", ms.Name)
	}
	if len(ms.Response.Contents) > 1024*1024*1024 {
		return fmt.Errorf("contents is too long %d", len(ms.Response.Contents))
	}
	return nil
}

// NormalPath normalizes path
func (ms *MockScenario) NormalPath(sep uint8) string {
	return NormalizePath(ms.Path, sep)
}

// SetName sets name
func (ms *MockScenario) SetName(prefix string) {
	ms.Name = fmt.Sprintf("%s%s-%d-%s", prefix, NormalizeDirPath(ms.NormalName()), ms.Response.StatusCode, ms.Digest())
}

// NormalName normalizes name from path
func (ms *MockScenario) NormalName() string {
	return NormalizePath(ms.Path, '-')
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
