package repository

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"gopkg.in/yaml.v3"
	"io"
	"io/fs"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/bhatti/api-mock-service/internal/types"
	"github.com/bhatti/api-mock-service/internal/utils"
	log "github.com/sirupsen/logrus"
)

// FileMockScenarioRepository  implements mock scenario storage based on local files
type FileMockScenarioRepository struct {
	mutex            sync.RWMutex
	dir              string
	keysByMethodPath map[string]map[string]*types.MockScenarioKeyData
}

// NewFileMockScenarioRepository creates new instance for mock scenarios
func NewFileMockScenarioRepository(
	config *types.Configuration,
) (repo *FileMockScenarioRepository, err error) {
	if err = mkdir(config.DataDir); err != nil {
		return nil, err
	}
	repo = &FileMockScenarioRepository{
		dir:              config.DataDir,
		keysByMethodPath: make(map[string]map[string]*types.MockScenarioKeyData),
	}

	err = repo.visit(func(keyData *types.MockScenarioKeyData) bool {
		keyMap := repo.keysByMethodPath[keyData.PartialMethodPathKey()]
		if keyMap == nil {
			keyMap = make(map[string]*types.MockScenarioKeyData)
			repo.keysByMethodPath[keyData.PartialMethodPathKey()] = keyMap
		}
		keyMap[keyData.MethodNamePathPrefixKey()] = keyData
		return false
	})

	if err != nil {
		return nil, err
	}
	return
}

// Get MockScenario by Method, Path and Scenario name
func (sr *FileMockScenarioRepository) Get(
	method types.MethodType,
	scenarioName string,
	path string,
	params interface{}) (scenario *types.MockScenario, err error) {
	var b []byte
	fileName := sr.buildFileName(method, scenarioName, path)
	if b, err = os.ReadFile(fileName); err != nil {
		return nil, err
	}
	keyData, err := unmarshalScenarioKeyData(b)
	if err != nil {
		return nil, err
	}
	dir := sr.buildDir(keyData.Method, keyData.Path)
	return unmarshalMockScenario(b, dir, params)
}

// GetScenariosNames returns mock scenarios for given Method and Path
func (sr *FileMockScenarioRepository) GetScenariosNames(
	method types.MethodType,
	path string) (scenarioNames []string, err error) {
	var files []fs.FileInfo
	dir := sr.buildDir(method, path)
	if files, err = ioutil.ReadDir(dir); err != nil {
		return nil, err
	}

	for _, file := range files {
		name := file.Name()
		if strings.HasSuffix(name, types.ScenarioExt) {
			trimSize := len(name) - len(types.ScenarioExt)
			scenarioNames = append(scenarioNames, name[0:trimSize])
		}
	}
	return
}

// Save MockScenario
func (sr *FileMockScenarioRepository) Save(
	scenario *types.MockScenario) (err error) {
	if err = scenario.Validate(); err != nil {
		return err
	}
	var b []byte
	if b, err = yaml.Marshal(scenario); err != nil {
		return err
	}
	return sr.SaveYaml(scenario.ToKeyData(), b)
}

// SaveRaw saves raw data assuming to be yaml format
func (sr *FileMockScenarioRepository) SaveRaw(input io.ReadCloser) (err error) {
	data, err := io.ReadAll(input)
	if err != nil {
		return err
	}
	input.Close()
	keyData, err := unmarshalScenarioKeyData(data)
	if err != nil {
		return err
	}
	return sr.SaveYaml(keyData, data)
}

// SaveYaml saves MockScenario as yaml format
func (sr *FileMockScenarioRepository) SaveYaml(keyData *types.MockScenarioKeyData, payload []byte) (err error) {
	dir := sr.buildDir(keyData.Method, keyData.Path)
	if err = mkdir(dir); err != nil {
		return err
	}
	fileName := sr.buildFileName(keyData.Method, keyData.Name, keyData.Path)
	err = os.WriteFile(fileName, payload, 0644)
	sr.addKeyData(keyData)
	return
}

// Delete removes a job
func (sr *FileMockScenarioRepository) Delete(
	method types.MethodType,
	scenarioName string,
	path string) error {
	fileName := sr.buildFileName(method, scenarioName, path)
	return os.Remove(fileName)
}

// LookupAll finds matching scenarios
func (sr *FileMockScenarioRepository) LookupAll(target *types.MockScenarioKeyData) []*types.MockScenarioKeyData {
	sr.mutex.RLock()
	defer func() {
		sr.mutex.RUnlock()
	}()
	res := make([]*types.MockScenarioKeyData, 0)
	keyDataMap := sr.keysByMethodPath[target.PartialMethodPathKey()]
	for _, keyData := range keyDataMap {
		if keyData.Equals(target) == nil {
			res = append(res, keyData)
		}
	}
	sort.Slice(res, func(i, j int) bool {
		if res[i].LastUsageTime == res[j].LastUsageTime {
			return res[i].Name < res[j].Name
		}
		return res[i].LastUsageTime < res[j].LastUsageTime
	})
	return res
}

// Lookup finds top matching scenario
func (sr *FileMockScenarioRepository) Lookup(target *types.MockScenarioKeyData) (scenario *types.MockScenario, err error) {
	matched := sr.LookupAll(target)
	if len(matched) == 0 {
		return nil, fmt.Errorf("could not lookup matching API %s", target.Path)
	}
	matched[0].LastUsageTime = time.Now().Unix()
	_ = atomic.AddUint64(&matched[0].RequestCount, 1)

	log.WithFields(log.Fields{
		"Path":         matched[0].Path,
		"Name":         matched[0].Name,
		"Method":       matched[0].Method,
		"RequestCount": matched[0].RequestCount,
		"Timestamp":    matched[0].LastUsageTime,
		"Matched":      len(matched),
	}).Infof("API template found...")

	// Read template file
	dir := sr.buildDir(target.Method, target.Path)
	fileName := sr.buildFileName(matched[0].Method, matched[0].Name, matched[0].Path)
	b, err := os.ReadFile(fileName)
	if err != nil {
		return nil, err
	}

	// Find any params for query params and path variables
	params := matched[0].MatchGroups(target.Path)
	if matched[0].QueryParams != "" {
		addQueryParams(matched[0].QueryParams, params)
	}
	if target.QueryParams != "" {
		addQueryParams(target.QueryParams, params)
	}
	params[types.RequestCount] = fmt.Sprintf("%d", matched[0].RequestCount)

	scenario, err = unmarshalMockScenario(b, dir, params)
	scenario.RequestCount = matched[0].RequestCount
	return
}

func unmarshalMockScenario(
	b []byte,
	dir string,
	params interface{}) (scenario *types.MockScenario, err error) {
	// parse template
	b, err = utils.ParseTemplate(dir, b, params)
	if err != nil {
		return nil, err
	}

	// unmarshal scenario from template output
	scenario = &types.MockScenario{}
	if err = yaml.Unmarshal(b, scenario); err != nil {
		return nil, err
	}
	return scenario, nil
}

/////////// PRIVATE METHODS //////////////

// visit all scenarios matching properties
func (sr *FileMockScenarioRepository) visit(
	callback func(keyData *types.MockScenarioKeyData) bool) error {
	var errStop = errors.New("stop")
	var walkFunc = func(path string, info os.FileInfo, err error) (_ error) {
		// handle walking error if any
		if err != nil {
			return err
		}

		// filter by extension
		if filepath.Ext(path) != types.ScenarioExt {
			return
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		keyData, err := unmarshalScenarioKeyData(content)
		if err != nil {
			return err
		}
		if callback(keyData) {
			return errStop
		}
		return
	}

	err := filepath.Walk(sr.dir, walkFunc)
	if err == errStop {
		err = nil
	}
	return err
}

// PRIVATE METHODS

func mkdir(dir string) error {
	_, err := os.Stat(dir)
	if os.IsNotExist(err) {
		return os.MkdirAll(dir, 0755)
	}
	return nil
}

func (sr *FileMockScenarioRepository) addKeyData(keyData *types.MockScenarioKeyData) {
	sr.mutex.Lock()
	defer func() {
		sr.mutex.Unlock()
	}()
	keyMap := sr.keysByMethodPath[keyData.PartialMethodPathKey()]
	if keyMap == nil {
		keyMap = make(map[string]*types.MockScenarioKeyData)
		sr.keysByMethodPath[keyData.PartialMethodPathKey()] = keyMap
	}
	keyMap[keyData.MethodNamePathPrefixKey()] = keyData
	log.WithFields(log.Fields{
		"Name":    keyData.Name,
		"Path":    keyData.Path,
		"AllSize": len(sr.keysByMethodPath),
		"Size":    len(keyMap),
	}).Infof("addKeyData added...")
}

func (sr *FileMockScenarioRepository) buildFileName(
	method types.MethodType,
	scenarioName string,
	path string) string {
	return buildFileName(sr.dir, method, scenarioName, path) + types.ScenarioExt
}

func (sr *FileMockScenarioRepository) buildDir(
	method types.MethodType,
	path string) string {
	return buildDir(sr.dir, method, path)
}

func buildFileName(
	dir string,
	method types.MethodType,
	scenarioName string,
	path string) string {
	return filepath.Join(buildDir(dir, method, path), scenarioName)
}

func buildDir(
	dir string,
	method types.MethodType,
	path string) string {
	return filepath.Join(dir, types.NormalizeDirPath(path), string(method))
}

func addQueryParams(queryParams string, params map[string]string) {
	if dict, err := url.ParseQuery(queryParams); err == nil {
		for k, vals := range dict {
			if len(vals) > 0 {
				params[k] = vals[0]
			}
		}
	}
}

func unmarshalScenarioKeyData(data []byte) (*types.MockScenarioKeyData, error) {
	scanner := bufio.NewScanner(bytes.NewReader(data))
	keyData := &types.MockScenarioKeyData{}
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "response:") {
			break
		}
		if !strings.Contains(line, ":") {
			continue
		}
		ndx := strings.Index(line, ":")
		if ndx == -1 {
			continue
		}
		name := strings.TrimSpace(line[0:ndx])
		val := strings.TrimSpace(line[ndx+1:])
		if name == "method" {
			keyData.Method = types.MethodType(val)
		} else if name == "name" {
			keyData.Name = val
		} else if name == "path" {
			keyData.Path = val
		} else if name == "query_params" {
			keyData.QueryParams = val
		} else if name == "content_type" {
			keyData.ContentType = val
		} else if name == "contents" {
			keyData.Contents = val
		}
	}
	if err := keyData.Validate(); err != nil {
		return nil, err
	}
	return keyData, nil
}
