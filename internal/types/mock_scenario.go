package types

import (
	"crypto/sha256"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"
)

// MockRecordMode header
const MockRecordMode = "X-Mock-Record"

// MockRecordModeDisabled disabled value
const MockRecordModeDisabled = "false"

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

// MockDataExt extension
const MockDataExt = ".dat"

// ScenarioExt extension
const ScenarioExt = ".scr"

// RequestCount name
const RequestCount = "_RequestCount"

// PrefixTypeNumber type
const PrefixTypeNumber = "__number__"

// PrefixTypeBoolean type
const PrefixTypeBoolean = "__boolean__"

// PrefixTypeString type
const PrefixTypeString = "__string__"

// PrefixTypeObject type
const PrefixTypeObject = "__object__"

// MockHTTPRequest defines mock request for APIs
type MockHTTPRequest struct {
	// MatchQueryParams for the API
	MatchQueryParams map[string]string `yaml:"match_query_params" json:"match_query_params"`
	// MatchHeaders for mock response
	MatchHeaders map[string]string `yaml:"match_headers" json:"match_headers"`
	// MatchContents for request optionally
	MatchContents string `yaml:"match_contents" json:"match_contents"`
	// ExamplePathParams sample for the API
	ExamplePathParams map[string]string `yaml:"example_path_params" json:"example_path_params"`
	// ExampleQueryParams sample for the API
	ExampleQueryParams map[string]string `yaml:"example_query_params" json:"example_query_params"`
	// ExampleHeaders for mock response
	ExampleHeaders map[string]string `yaml:"example_headers" json:"example_headers"`
	// ExampleContents sample for request optionally
	ExampleContents string `yaml:"example_contents" json:"example_contents"`
}

// MockHTTPResponse defines mock respons for APIs
type MockHTTPResponse struct {
	// Headers for mock response
	Headers map[string][]string `yaml:"headers" json:"headers"`
	// Contents for request
	Contents string `yaml:"contents" json:"contents"`
	// ContentsFile for request
	ContentsFile string `yaml:"contents_file" json:"contents_file"`
	// StatusCode for response
	StatusCode int `yaml:"status_code" json:"status_code"`
	// MatchHeaders for mock response
	MatchHeaders map[string]string `yaml:"match_headers" json:"match_headers"`
	// MatchContents for request optionally
	MatchContents string `yaml:"match_contents" json:"match_contents"`
	// Assertions for validating response
	Assertions []string `yaml:"assertions" json:"assertions"`
}

// ContentType find content-type
func (r MockHTTPResponse) ContentType() string {
	for k, v := range r.Headers {
		if strings.ToUpper(k) == strings.ToUpper(ContentTypeHeader) {
			return v[0]
		}
	}
	return ""
}

// MockScenario defines mock scenario for APIs
type MockScenario struct {
	// Method for HTTP API
	Method MethodType `yaml:"method" json:"method"`
	// Name to uniquely identify the scenario
	Name string `yaml:"name" json:"name"`
	// Path for the API (excluding query params)
	Path string `yaml:"path" json:"path"`
	// Description of scenario
	Description string `yaml:"description" json:"description"`
	// Group of scenario
	Group string `yaml:"group" json:"group"`
	// Predicate for the request
	Predicate string `yaml:"predicate" json:"predicate"`
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
		Method:           ms.Method,
		Name:             ms.Name,
		Path:             rawPath,
		Group:            ms.Group,
		Predicate:        ms.Predicate,
		MatchQueryParams: ms.Request.MatchQueryParams,
		MatchContents:    ms.Request.MatchContents,
		MatchHeaders:     ms.Request.MatchHeaders,
	}
}

// String
func (ms *MockScenario) String() string {
	return string(ms.Method) + ms.Name + ms.Group + ms.Path
}

// Digest of scenario
func (ms *MockScenario) Digest() string {
	h := sha256.New()
	h.Write([]byte(ms.Method))
	h.Write([]byte(ms.Group))
	h.Write([]byte(ms.Path))
	for k, v := range ms.Request.MatchQueryParams {
		h.Write([]byte(k))
		h.Write([]byte(v))
	}
	h.Write([]byte(ms.Request.MatchContents))
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

// StripTypeTags removes type prefixes
func StripTypeTags(re string) string {
	return regexp.MustCompile(
		fmt.Sprintf("(%s|%s|%s|%s)", PrefixTypeNumber, PrefixTypeBoolean, PrefixTypeString, PrefixTypeObject)).
		ReplaceAllString(re, "")
}

func reMatch(re string, str string) bool {
	re = StripTypeTags(re)
	match, err := regexp.MatchString(re, str)
	if err != nil {
		return false
	}
	return match
}
