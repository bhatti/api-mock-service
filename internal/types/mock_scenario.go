package types

import (
	"crypto/sha256"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"
)

// ContentTypeHeader header
const ContentTypeHeader = "Content-Type"

// MockDataExt extension
const MockDataExt = ".dat"

// ScenarioExt extension
const ScenarioExt = ".scr"

// RequestCount name
const RequestCount = "_RequestCount"

// MockHTTPRequest defines mock request for APIs
type MockHTTPRequest struct {
	// QueryParams for the API
	QueryParams string `yaml:"query_params" json:"query_params"`
	// Headers for mock response
	Headers map[string][]string `yaml:"headers" json:"headers"`
	// ContentType for the response
	ContentType string `yaml:"content_type" json:"content_type"`
	// Contents for request optionally
	Contents string `yaml:"contents" json:"contents"`
}

// MockHTTPResponse defines mock respons for APIs
type MockHTTPResponse struct {
	// Headers for mock response
	Headers map[string][]string `yaml:"headers" json:"headers"`
	// ContentType for the response
	ContentType string `yaml:"content_type" json:"content_type"`
	// Contents for request
	Contents string `yaml:"contents" json:"contents"`
	// ContentsFile for request
	ContentsFile string `yaml:"contents_file" json:"contents_file"`
	// StatusCode for response
	StatusCode int `yaml:"status_code" json:"status_code"`
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
	rawPath := "/" + NormalizePath(ms.Path, '/')
	re := regexp.MustCompile(`:[\w\d-_]+`)
	rePath := re.ReplaceAllString(rawPath, `(.*)`)
	return &MockScenarioKeyData{
		Method:      ms.Method,
		Name:        ms.Name,
		rePath:      rePath,
		Path:        rawPath,
		QueryParams: ms.Request.QueryParams,
		ContentType: ms.Request.ContentType,
		Contents:    ms.Request.Contents,
	}
}

// String
func (ms *MockScenario) String() string {
	return string(ms.Method) + ms.Name + ms.Path
}

// Digest of scenario
func (ms *MockScenario) Digest() string {
	h := sha256.New()
	h.Write([]byte(ms.Method))
	h.Write([]byte(ms.Path))
	h.Write([]byte(ms.Request.QueryParams))
	h.Write([]byte(ms.Request.ContentType))
	h.Write([]byte(ms.Request.Contents))
	return fmt.Sprintf("%x", h.Sum(nil))
}

// Equals method
func (ms *MockScenario) Equals(other *MockScenario) error {
	if ms.Method != other.Method {
		return fmt.Errorf("method didn't match")
	}
	if ms.Name != other.Name {
		return fmt.Errorf("scenario name didn't match")
	}
	if ms.Path != other.Path {
		return fmt.Errorf("path didn't match")
	}
	if ms.Request.QueryParams != other.Request.QueryParams {
		return fmt.Errorf("query params didn't match")
	}
	if ms.Response.Contents != other.Response.Contents {
		return fmt.Errorf("response contents didn't match")
	}
	if ms.Request.ContentType != other.Request.ContentType {
		return fmt.Errorf("request content type didn't match")
	}
	if ms.Response.ContentType != other.Response.ContentType {
		return fmt.Errorf("response content type didn't match")
	}
	if len(ms.Request.Headers) != len(other.Request.Headers) {
		return fmt.Errorf("headers didnn't match")
	}
	for k, vals := range ms.Request.Headers {
		if len(other.Request.Headers[k]) != len(vals) {
			return fmt.Errorf("%s request header didn't match", k)
		}
		for i, val := range vals {
			if other.Request.Headers[k][i] != val {
				return fmt.Errorf("%s request header didn't match", k)
			}
		}
	}
	if ms.WaitBeforeReply != other.WaitBeforeReply {
		return fmt.Errorf("wait time didn't match")
	}
	if ms.Response.StatusCode != other.Response.StatusCode {
		return fmt.Errorf("response code didn't match")
	}
	return nil
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
	if matched, err := regexp.Match(`^[\w\d-_\/\\:]+$`, []byte(ms.Path)); err == nil && !matched {
		return fmt.Errorf("path is invalid with special characters %s", ms.Path)
	}
	ms.Path = "/" + NormalizePath(ms.Path, '/')
	if ms.Name == "" {
		return fmt.Errorf("scenario name is not specified")
	}
	if len(ms.Name) > 200 {
		return fmt.Errorf("scenario name is too long %d", len(ms.Name))
	}
	if matched, err := regexp.Match(`^[\w\d-_]+\.?[\w\d-_]+$`, []byte(ms.Name)); err == nil && !matched {
		return fmt.Errorf("scenario name is invalid with special characters %s", ms.Name)
	}
	if ms.Response.Contents == "" {
		return fmt.Errorf("contents is not specified")
	}
	if len(ms.Response.Contents) > 1024*1024*1024 {
		return fmt.Errorf("contents is too long %d", len(ms.Response.Contents))
	}
	if ms.Response.StatusCode == 0 {
		return fmt.Errorf("response status is not specified")
	}
	return nil
}

// NormalPath normalizes path
func (ms *MockScenario) NormalPath(sep uint8) string {
	return NormalizePath(ms.Path, sep)
}

// MockScenarioKeyData defines keys of mock scenario for in-memory store
type MockScenarioKeyData struct {
	// Method for HTTP API
	Method MethodType
	// Name to uniquely identify the scenario
	Name string
	// rePath for the API modified with Regexp
	rePath string
	// Path for the API (original from scenario)
	Path string
	// QueryParams for the API
	QueryParams string
	// ContentType for the response
	ContentType string
	// Contents for request optionally
	Contents string
	// LastUsageTime of key data
	LastUsageTime int64
	// RequestCount for the API
	RequestCount uint64
}

// Equals compares path and query path
func (msd *MockScenarioKeyData) Equals(target *MockScenarioKeyData) error {
	matched, err := regexp.Match(msd.rePath, []byte(target.Path))
	if err != nil {
		return err
	}
	if !matched {
		return fmt.Errorf("path '%s' didn't match '%s'", msd.Path, target.Path)
	}
	if msd.QueryParams != "" && target.QueryParams != "" && !strings.Contains(msd.QueryParams, target.QueryParams) {
		return fmt.Errorf("query-params '%s' didn't match '%s'", msd.QueryParams, target.QueryParams)
	}
	if msd.ContentType != "" && target.ContentType != "" && !strings.Contains(msd.ContentType, target.ContentType) {
		return fmt.Errorf("content-type '%s' didn't match '%s'", msd.ContentType, target.ContentType)
	}
	if msd.Contents != "" && target.Contents != "" && !strings.Contains(msd.Contents, target.Contents) {
		return fmt.Errorf("contents '%s' didn't match '%s'", msd.Contents, target.Contents)
	}
	if target.Name != "" && msd.Name != target.Name {
		return fmt.Errorf("name '%s' didn't match '%s'", msd.Name, target.Name)
	}
	return nil
}

// MatchGroups return match groups for dynamic params in path
func (msd *MockScenarioKeyData) MatchGroups(path string) (res map[string]string) {
	// extract ids
	rawRe := regexp.MustCompile(`(:[\d\w-_]+)`)
	rawParts := rawRe.FindAllStringSubmatch(msd.Path, -1)

	// extract values
	re := regexp.MustCompile(msd.rePath)
	parts := re.FindStringSubmatch(path)

	res = make(map[string]string)
	for i := 1; i < len(parts); i++ {
		if len(rawParts[i-1]) > 0 && len(rawParts[i-1][0]) > 0 {
			res[rawParts[i-1][0][1:]] = parts[i]
		}
	}
	return
}

// Validate scenario
func (msd *MockScenarioKeyData) Validate() error {
	if msd.Method == "" {
		return fmt.Errorf("method is not specified")
	}
	if msd.Path == "" {
		return fmt.Errorf("path is not specified")
	}
	if len(msd.Path) > 200 {
		return fmt.Errorf("path is too long %d", len(msd.Path))
	}
	if matched, err := regexp.Match(`^[\w\d-_\/\\:]+$`, []byte(msd.Path)); err == nil && !matched {
		return fmt.Errorf("path is invalid with special characters %s", msd.Path)
	}
	msd.Path = "/" + NormalizePath(msd.Path, '/')
	if msd.Name == "" {
		return fmt.Errorf("scenario name is not specified")
	}
	if len(msd.Name) > 200 {
		return fmt.Errorf("scenario name is too long %d", len(msd.Name))
	}
	if matched, err := regexp.Match(`^[\w\d-_]+\.?[\w\d-_]+$`, []byte(msd.Name)); err == nil && !matched {
		return fmt.Errorf("scenario name is invalid with special characters %s", msd.Name)
	}
	return nil
}

// String
func (msd *MockScenarioKeyData) String() string {
	return msd.MethodNamePathPrefixKey()
}

// MethodNamePathPrefixKey returns full key for the scenario
func (msd *MockScenarioKeyData) MethodNamePathPrefixKey() string {
	return string(msd.Method) + msd.Name + msd.PathPrefix(1)
}

// PartialMethodPathKey for key by method and first-level path
func (msd *MockScenarioKeyData) PartialMethodPathKey() string {
	return string(msd.Method) + msd.PathPrefix(1)
}

// PathPrefix builds prefix of path
func (msd *MockScenarioKeyData) PathPrefix(max int) string {
	parts := strings.Split(msd.Path, "/")
	if len(parts) <= max {
		return msd.Path
	}

	var buf strings.Builder
	j := 0
	for i := 0; i < len(parts); i++ {
		if parts[i] != "" && j < max {
			buf.WriteString("/" + parts[i])
			j++
		}
	}
	return buf.String()
}

// NormalizeDirPath normalizes dir path
func NormalizeDirPath(path string) string {
	path = NormalizePath(path, os.PathSeparator)
	ndx := strings.Index(path, ":")
	if ndx > 1 {
		path = path[0 : ndx-1]
	}
	return path
}

// NormalizePath normalizes path
func NormalizePath(path string, sepChar uint8) string {
	sep := fmt.Sprintf("%c", sepChar)
	if regexp, err := regexp.Compile(`[\/\\]+`); err == nil {
		path = regexp.ReplaceAllString(path, sep)
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
