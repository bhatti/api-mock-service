package types

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/bhatti/api-mock-service/internal/fuzz"
	log "github.com/sirupsen/logrus"
)

// MockScenarioKeyData defines keys of mock scenario for in-memory store
type MockScenarioKeyData struct {
	// Method for HTTP API
	Method MethodType `yaml:"method" json:"method"`
	// Name to uniquely identify the scenario
	Name string `yaml:"name" json:"name"`
	// Path for the API (excluding query params)
	Path string `yaml:"path" json:"path"`
	// Order of scenario
	Order int `yaml:"order" json:"order"`
	// Group of scenario
	Group string `yaml:"group" json:"group"`
	// Tags of scenario
	Tags []string `yaml:"tags" json:"tags"`
	// Predicate for the request
	Predicate string `yaml:"predicate" json:"predicate"`
	// AssertQueryParamsPattern for the API
	AssertQueryParamsPattern map[string]string `yaml:"assert_query_params_pattern" json:"assert_query_params_pattern"`
	// AssertHeadersPattern for mock response
	AssertHeadersPattern map[string]string `yaml:"assert_headers_pattern" json:"assert_headers_pattern"`
	// AssertContentsPattern for request optionally
	AssertContentsPattern string `yaml:"assert_contents_pattern" json:"assert_contents_pattern"`
	// LastUsageTime of key data
	LastUsageTime int64
	// RequestCount for the API
	RequestCount uint64
}

// Equals compares path and query path
func (msd *MockScenarioKeyData) Equals(other *MockScenarioKeyData) error {
	if msd.Method != other.Method {
		return NewNotFoundError(fmt.Sprintf("method '%s' didn't match '%s'", msd.Method, other.Method))
	}
	if msd.Group != "" && other.Group != "" && msd.Group != other.Group {
		return NewNotFoundError(fmt.Sprintf("group '%s' didn't match '%s'", msd.Group, other.Group))
	}
	otherPath := filterURLQueryParams(other.Path)
	rePath := rePath(msd.Path)
	matched, err := regexp.Match(rePath, []byte(otherPath))
	if err != nil {
		return err
	}
	log.WithFields(log.Fields{
		"Group":     msd.Group,
		"Order":     msd.Order,
		"Other":     other.String(),
		"This":      msd.String(),
		"OtherPath": otherPath,
		"ThisPath":  msd.Path,
		"RegexPath": rePath,
		"Matched":   matched,
	}).Debugf("matching path...")
	if !matched {
		return NewNotFoundError(fmt.Sprintf("path '%s' didn't match '%s'", msd.Path, other.Path))
	}
	for k, msdQueryParamVal := range msd.AssertQueryParamsPattern {
		targetQueryParamVal := other.AssertQueryParamsPattern[k]
		if targetQueryParamVal != msdQueryParamVal &&
			!reMatch(msdQueryParamVal, targetQueryParamVal) {
			return NewValidationError(fmt.Sprintf("request queryParam '%s' didn't match [%v == %v]",
				k, msd.AssertQueryParamsPattern, other.AssertQueryParamsPattern))
		}
	}

	if msd.AssertContentsPattern != "" &&
		!strings.Contains(msd.AssertContentsPattern, other.AssertContentsPattern) &&
		!reMatch(msd.AssertContentsPattern, other.AssertContentsPattern) {
		if other.AssertContentsPattern == "" {
			return NewValidationError(fmt.Sprintf("contents '%s' didn't match '%s'",
				msd.AssertContentsPattern, other.AssertContentsPattern))
		}
		regex := make(map[string]string)
		err := json.Unmarshal([]byte(msd.AssertContentsPattern), &regex)
		if err != nil {
			return fmt.Errorf("failed to unmarshal contents '%s' regex due to %w", msd.AssertContentsPattern, err)
		}
		matchContents, err := fuzz.UnmarshalArrayOrObject([]byte(other.AssertContentsPattern))
		if err != nil {
			return fmt.Errorf("failed to unmarshal other contents '%s' regex due to %w", other.AssertContentsPattern, err)
		}
		err = fuzz.ValidateRegexMap(matchContents, regex)
		if err != nil {
			return NewValidationError(fmt.Sprintf("contents '%s' didn't match '%s' due to %s",
				msd.AssertContentsPattern, other.AssertContentsPattern, err))
		}
	}

	for k, msdHeaderVal := range msd.AssertHeadersPattern {
		targetHeaderVal := getDictValue(k, other.AssertHeadersPattern)
		if targetHeaderVal != msdHeaderVal &&
			!reMatch(msdHeaderVal, targetHeaderVal) {
			return NewValidationError(fmt.Sprintf("%s request header didn't match [%v == %v], all headers %v",
				k, targetHeaderVal, msdHeaderVal, other.AssertHeadersPattern))
		}
	}

	if len(msd.Tags) > 0 && len(other.Tags) > 0 {
		strMap := toStringMap(msd.Tags)
		for _, tag := range other.Tags {
			if !strMap[strings.ToUpper(tag)] {
				return NewValidationError(fmt.Sprintf("%s request tag didn't match %v, all tags %v",
					tag, msd.Tags, other.Tags))
			}
		}
	}

	if other.Name != "" && msd.Name != other.Name {
		return NewValidationError(fmt.Sprintf("scenario name '%s' didn't match '%s'",
			msd.Name, other.Name))
	}
	return nil
}

// MatchGroups return match groups for dynamic params in path
func (msd *MockScenarioKeyData) MatchGroups(path string) map[string]string {
	return MatchPathGroups(msd.Path, path)
}

// MatchPathGroups return match groups for dynamic params in path
func MatchPathGroups(rawPath string, targetPath string) (res map[string]string) {
	rePath := rePath(rawPath)
	matched, err := regexp.Match(rePath, []byte(targetPath))
	if err != nil {
		return
	}
	if !matched {
		return
	}

	res = make(map[string]string)

	// extract dynamic properties using :id or {id} format
	var rawParts [][]string
	if strings.Contains(rawPath, ":") {
		rawRe := regexp.MustCompile(`(:[\d\w-_]+)`)
		rawParts = rawRe.FindAllStringSubmatch(rawPath, -1)
	} else if strings.Contains(rawPath, "{") && strings.Contains(rawPath, "}") {
		rawRe := regexp.MustCompile(`(\{[\d\w-_]+)`)
		rawParts = rawRe.FindAllStringSubmatch(rawPath, -1)
	} else {
		return
	}

	// extract values
	re := regexp.MustCompile(rePath)
	parts := re.FindStringSubmatch(targetPath)

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
		return fmt.Errorf("key method is not specified")
	}
	if msd.Path == "" {
		return fmt.Errorf("key path is not specified")
	}
	if len(msd.Path) > 200 {
		return fmt.Errorf("key path is too long %d", len(msd.Path))
	}
	if matched, err := regexp.Match(`^[\w\d\.\-_\/\\:{}]+$`, []byte(msd.Path)); err == nil && !matched {
		return fmt.Errorf("key path is invalid with special characters '%s'", msd.Path)
	}
	msd.Path = NormalizePath(msd.Path, '/')
	if !strings.HasPrefix(msd.Path, "/") {
		msd.Path = "/" + msd.Path
	}
	if msd.Name == "" {
		return fmt.Errorf("key scenario name is not specified")
	}
	if len(msd.Name) > 200 {
		return fmt.Errorf("key scenario name is too long %d", len(msd.Name))
	}
	if matched, err := regexp.Match(`^[\w\d-_\.]+$`, []byte(msd.Name)); err == nil && !matched {
		return fmt.Errorf("key scenario name is invalid with special characters %s", msd.Name)
	}
	return nil
}

// String
func (msd *MockScenarioKeyData) String() string {
	return string(msd.Method) + "|" + msd.Path + "|" + msd.Name
}

// MethodPath helper method
func (msd *MockScenarioKeyData) MethodPath() string {
	return strings.ToLower(string(msd.Method)) + "_" + SanitizeNonAlphabet(msd.Path, "_") // replace slash
}

// SafeName strips invalid characters
func (msd *MockScenarioKeyData) SafeName() string {
	return SanitizeNonAlphabet(msd.Name, "")
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

func filterURLQueryParams(rawPath string) string {
	ndx := strings.Index(rawPath, "?")
	if ndx != -1 {
		return rawPath[0:ndx]
	}
	return rawPath
}

func rePath(rawPath string) (rePath string) {
	rawPath = filterURLQueryParams(rawPath)

	targetPattern := `([^/]*)`
	rePath = rawPath
	if strings.Contains(rePath, ":") {
		re := regexp.MustCompile(`:[\w\d-_]+`)
		rePath = re.ReplaceAllString(rawPath, targetPattern)
	} else if strings.Contains(rePath, "{") && strings.Contains(rePath, "}") {
		re := regexp.MustCompile(`{[\w\d-_]+}`)
		rePath = re.ReplaceAllString(rawPath, targetPattern)
	}
	ndx := strings.LastIndex(rePath, targetPattern)
	if ndx != -1 {
		rePath = fmt.Sprintf("%s(.+)%s", rePath[0:ndx], rePath[ndx+len(targetPattern):])
	}
	if len(rePath) > 0 {
		rePath += `$`
	}
	return
}

func getDictValue(name string, dict map[string]string) string {
	name = strings.ToUpper(name)
	for k, v := range dict {
		if strings.ToUpper(k) == name {
			return v
		}
	}
	return ""
}

func toStringMap(arr []string) (res map[string]bool) {
	res = make(map[string]bool)
	for _, e := range arr {
		res[strings.ToUpper(e)] = true
	}
	return
}
