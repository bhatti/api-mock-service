package types

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"regexp"
	"strings"
)

// MockScenarioKeyData defines keys of mock scenario for in-memory store
type MockScenarioKeyData struct {
	// Method for HTTP API
	Method MethodType `yaml:"method" json:"method"`
	// Name to uniquely identify the scenario
	Name string `yaml:"name" json:"name"`
	// Path for the API (excluding query params)
	Path string `yaml:"path" json:"path"`
	// Group of scenario
	Group string `yaml:"group" json:"group"`
	// Predicate for the request
	Predicate string `yaml:"predicate" json:"predicate"`
	// MatchQueryParams for the API
	MatchQueryParams map[string]string `yaml:"match_query_params" json:"match_query_params"`
	// MatchHeaders for mock response
	MatchHeaders map[string]string `yaml:"match_headers" json:"match_headers"`
	// MatchContents for request optionally
	MatchContents string `yaml:"match_contents" json:"match_contents"`
	// LastUsageTime of key data
	LastUsageTime int64
	// RequestCount for the API
	RequestCount uint64
}

// Equals compares path and query path
func (msd *MockScenarioKeyData) Equals(target *MockScenarioKeyData) error {
	if msd.Method != target.Method {
		return NewNotFoundError(fmt.Sprintf("method '%s' didn't match '%s'", msd.Method, target.Method))
	}
	if msd.Group != "" && target.Group != "" && msd.Group != target.Group {
		return NewNotFoundError(fmt.Sprintf("group '%s' didn't match '%s'", msd.Group, target.Group))
	}
	targetPath := filterURLQueryParams(target.Path)
	rePath := rePath(msd.Path)
	matched, err := regexp.Match(rePath, []byte(targetPath))
	if err != nil {
		return err
	}
	log.WithFields(log.Fields{
		"Group":      msd.Group,
		"Target":     target.String(),
		"This":       msd.String(),
		"TargetPath": targetPath,
		"ThisPath":   msd.Path,
		"RegexPath":  rePath,
		"Matched":    matched,
	}).Debugf("matching path...")
	if !matched {
		return NewNotFoundError(fmt.Sprintf("path '%s' didn't match '%s'", msd.Path, target.Path))
	}
	for k, msdQueryParamVal := range msd.MatchQueryParams {
		targetQueryParamVal := target.MatchQueryParams[k]
		if targetQueryParamVal != msdQueryParamVal &&
			!reMatch(msdQueryParamVal, targetQueryParamVal) {
			return NewValidationError(fmt.Sprintf("request queryParam '%s' didn't match [%v == %v]",
				k, msd.MatchQueryParams, target.MatchQueryParams))
		}
	}

	if msd.MatchContents != "" &&
		!strings.Contains(msd.MatchContents, target.MatchContents) &&
		!reMatch(msd.MatchContents, target.MatchContents) {
		return NewValidationError(fmt.Sprintf("contents '%s' didn't match '%s'",
			msd.MatchContents, target.MatchContents))
	}

	for k, msdHeaderVal := range msd.MatchHeaders {
		targetHeaderVal := getDictValue(k, target.MatchHeaders)
		if targetHeaderVal != msdHeaderVal &&
			!reMatch(msdHeaderVal, targetHeaderVal) {
			return NewValidationError(fmt.Sprintf("%s request header didn't match [%v == %v], all headers %v",
				k, targetHeaderVal, msdHeaderVal, target.MatchHeaders))
		}
	}

	if target.Name != "" && msd.Name != target.Name {
		return NewValidationError(fmt.Sprintf("scenario name '%s' didn't match '%s'",
			msd.Name, target.Name))
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
	msd.Path = "/" + NormalizePath(msd.Path, '/')
	if msd.Name == "" {
		return fmt.Errorf("key scenario name is not specified")
	}
	if len(msd.Name) > 200 {
		return fmt.Errorf("key scenario name is too long %d", len(msd.Name))
	}
	if matched, err := regexp.Match(`^[\w\d-_]+\.?[\w\d-_]+$`, []byte(msd.Name)); err == nil && !matched {
		return fmt.Errorf("key scenario name is invalid with special characters %s", msd.Name)
	}
	return nil
}

// String
func (msd *MockScenarioKeyData) String() string {
	return string(msd.Method) + "|" + msd.Path + "|" + msd.Name
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
