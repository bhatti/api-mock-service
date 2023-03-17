package repository

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/bhatti/api-mock-service/internal/fuzz"
	"github.com/bhatti/api-mock-service/internal/types/har"
	"github.com/bhatti/api-mock-service/internal/web"
	"gopkg.in/yaml.v3"
	"io"
	"io/fs"
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

// FileAPIScenarioRepository  implements API scenario storage based on local files
type FileAPIScenarioRepository struct {
	mutex            sync.RWMutex
	keysByMethodPath map[string]map[string]*types.APIKeyData
	config           *types.Configuration
	contractDir      string
	historyDir       string
	harDir           string
	maxHistory       int
	debug            bool
}

// NewFileAPIScenarioRepository creates new instance for api scenarios
func NewFileAPIScenarioRepository(
	config *types.Configuration,
) (repo *FileAPIScenarioRepository, err error) {
	contractDir := buildContractsDir(config)
	historyDir := filepath.Join(config.DataDir, "exec_history")
	harDir := filepath.Join(config.DataDir, "har_history")
	if err = mkdir(contractDir); err != nil {
		return nil, err
	}
	if err = mkdir(historyDir); err != nil {
		return nil, err
	}
	if err = mkdir(harDir); err != nil {
		return nil, err
	}
	repo = &FileAPIScenarioRepository{
		config:           config,
		contractDir:      contractDir,
		historyDir:       historyDir,
		harDir:           harDir,
		maxHistory:       config.MaxHistory,
		debug:            config.Debug,
		keysByMethodPath: make(map[string]map[string]*types.APIKeyData),
	}

	err = repo.visit(func(keyData *types.APIKeyData) bool {
		keyMap := repo.keysByMethodPath[keyData.PartialMethodPathKey()]
		if keyMap == nil {
			keyMap = make(map[string]*types.APIKeyData)
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

// GetGroups returns api scenarios groups
func (sr *FileAPIScenarioRepository) GetGroups() (res []string) {
	sr.mutex.RLock()
	defer func() {
		sr.mutex.RUnlock()
	}()
	res = make([]string, 0)
	dupes := make(map[string]bool)
	for _, keyDataMap := range sr.keysByMethodPath {
		for _, keyData := range keyDataMap {
			if keyData.Group != "" && !dupes[keyData.Group] {
				res = append(res, keyData.Group)
				dupes[keyData.Group] = true
			}
		}
	}
	sort.Slice(res, func(i, j int) bool {
		return res[i] < res[j]
	})
	return
}

// GetScenariosNames returns api scenarios for given Method and Path
func (sr *FileAPIScenarioRepository) GetScenariosNames(
	method types.MethodType,
	path string) (scenarioNames []string, err error) {
	scenarioNames = make([]string, 0)
	var files []fs.DirEntry
	dir := sr.buildDir(method, path)

	files, err = os.ReadDir(dir)
	if err != nil {
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

// Save APIScenario
func (sr *FileAPIScenarioRepository) Save(
	scenario *types.APIScenario) (err error) {
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
func (sr *FileAPIScenarioRepository) SaveRaw(input io.ReadCloser) (err error) {
	data, _, err := utils.ReadAll(input)
	if err != nil {
		return err
	}
	keyData, err := unmarshalScenarioKeyData(data)
	if err != nil {
		return err
	}
	return sr.SaveYaml(keyData, data)
}

// SaveYaml saves APIScenario as yaml format
func (sr *FileAPIScenarioRepository) SaveYaml(keyData *types.APIKeyData, payload []byte) (err error) {
	dir := sr.buildDir(keyData.Method, keyData.Path)
	if err = mkdir(dir); err != nil {
		return err
	}
	fileName := sr.buildFileName(keyData.Method, keyData.Name, keyData.Path)
	err = os.WriteFile(fileName, payload, 0644)
	sr.addKeyData(keyData)
	return
}

// LoadRaw loads matching scenario
func (sr *FileAPIScenarioRepository) LoadRaw(
	method types.MethodType,
	name string,
	path string,
) (b []byte, err error) {
	fileName := sr.buildFileName(method, name, path)
	return os.ReadFile(fileName)
}

// Delete removes a job
func (sr *FileAPIScenarioRepository) Delete(
	method types.MethodType,
	scenarioName string,
	path string) error {
	fileName := sr.buildFileName(method, scenarioName, path)
	return os.Remove(fileName)
}

// ListScenarioKeyData returns keys for all scenarios
func (sr *FileAPIScenarioRepository) ListScenarioKeyData(group string) []*types.APIKeyData {
	sr.mutex.RLock()
	defer func() {
		sr.mutex.RUnlock()
	}()
	res := make([]*types.APIKeyData, 0)
	for _, keyDataMap := range sr.keysByMethodPath {
		for _, keyData := range keyDataMap {
			if group == "" || group == keyData.Group {
				res = append(res, keyData)
			}
		}
	}
	sort.Slice(res, func(i, j int) bool {
		if res[i].Path == res[j].Path {
			return res[i].Method < res[j].Method
		}
		return res[i].Path < res[j].Path
	})
	return res

}

// LookupAllByPath finds matching scenarios by path
func (sr *FileAPIScenarioRepository) LookupAllByPath(path string) []*types.APIKeyData {
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	sr.mutex.RLock()
	defer func() {
		sr.mutex.RUnlock()
	}()
	res := make([]*types.APIKeyData, 0)
	for _, keyDataMap := range sr.keysByMethodPath {
		for _, keyData := range keyDataMap {
			if path == keyData.Path {
				res = append(res, keyData)
			}
		}
	}
	sortByUsageTime(res)
	return res
}

// LookupAllByGroup finds matching scenarios by group
func (sr *FileAPIScenarioRepository) LookupAllByGroup(
	group string) []*types.APIKeyData {
	sr.mutex.RLock()
	defer func() {
		sr.mutex.RUnlock()
	}()
	res := make([]*types.APIKeyData, 0)
	for _, keyDataMap := range sr.keysByMethodPath {
		for _, keyData := range keyDataMap {
			if group == keyData.Group {
				res = append(res, keyData)
			}
		}
	}
	sortByUsageTime(res)
	return res
}

// LookupAll finds matching scenarios
func (sr *FileAPIScenarioRepository) LookupAll(
	other *types.APIKeyData,
) (res []*types.APIKeyData, paramMismatchErrors int) {
	sr.mutex.RLock()
	defer func() {
		sr.mutex.RUnlock()
	}()
	res = make([]*types.APIKeyData, 0)
	keyDataMap := sr.keysByMethodPath[other.PartialMethodPathKey()]
	for _, keyData := range keyDataMap {
		if err := keyData.Equals(other); err == nil {
			res = append(res, keyData)
		} else {
			var validationError *types.ValidationError
			if errors.As(err, &validationError) && sr.debug {
				paramMismatchErrors++
				log.WithFields(log.Fields{
					"Other":          other.String(),
					"Actual":         keyData.String(),
					"MismatchParams": paramMismatchErrors,
					"Error":          err,
				}).Infof("mock scenario didn't match for lookup...")
			}
		}
	}
	sortByUsageTime(res)
	return filterScenariosByPredicate(res, other), paramMismatchErrors
}

// Lookup finds top matching scenario
func (sr *FileAPIScenarioRepository) Lookup(
	other *types.APIKeyData,
	inData map[string]any) (scenario *types.APIScenario, err error) {
	matched, paramMismatchErrors := sr.LookupAll(other)
	if len(matched) == 0 {
		if paramMismatchErrors > 0 {
			return nil, types.NewValidationError(fmt.Sprintf("could not match input parameters for API %s", other.String()))
		}
		fileName := sr.buildFileName(other.Method, other.Name, other.Path)
		return nil, types.NewNotFoundError(fmt.Sprintf("could not lookup matching API '%s' [File '%s']",
			other.String(), fileName))
	}
	matched[0].LastUsageTime = time.Now().Unix()
	_ = atomic.AddUint64(&matched[0].RequestCount, 1)

	reqCount := sumRequestCount(matched)
	log.WithFields(log.Fields{
		"Path":              matched[0].Path,
		"Name":              matched[0].Name,
		"Method":            matched[0].Method,
		"RequestCount":      matched[0].RequestCount,
		"TotalRequestCount": reqCount,
		"Timestamp":         matched[0].LastUsageTime,
		"Matched":           len(matched),
	}).Debugf("API scenario found...")

	// Read template file
	dir := sr.buildDir(other.Method, other.Path)
	fileName := sr.buildFileName(matched[0].Method, matched[0].Name, matched[0].Path)
	b, err := os.ReadFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s due to %w", fileName, err)
	}

	data := make(map[string]any)
	for k, v := range inData {
		data[k] = v
	}
	// Find any params for query params and path variables
	for k, v := range matched[0].MatchGroups(other.Path) {
		data[k] = v
	}
	addQueryParams(matched[0].AssertQueryParamsPattern, data)
	addQueryParams(other.AssertQueryParamsPattern, data)
	data[fuzz.RequestCount] = fmt.Sprintf("%d", reqCount)

	scenario, err = unmarshalMockScenario(b, dir, data)
	if err != nil {
		return nil, fmt.Errorf("lookup failed to parse scenario '%s' due to %w", fileName, err)
	}
	scenario.RequestCount = reqCount
	return
}

// HistoryNames returns list of API scenarios names
func (sr *FileAPIScenarioRepository) HistoryNames(group string) (names []string) {
	names = make([]string, 0)
	sanitizedGroup := types.SanitizeNonAlphabet(group, "_")
	files := sr.historyFiles()
	for _, file := range files {
		if group == "" || strings.Contains(file.Name(), group) || strings.Contains(file.Name(), sanitizedGroup) {
			names = append(names, strings.ReplaceAll(file.Name(), types.ScenarioExt, ""))
		}
	}
	return
}

// SaveHistory saves history APIScenario
func (sr *FileAPIScenarioRepository) SaveHistory(
	scenario *types.APIScenario,
	url string,
	host string,
	started time.Time,
	ended time.Time,
) (err error) {
	for name := range web.IgnoredRequestHeaders {
		delete(scenario.Request.Headers, name)
	}
	scenario.Description = fmt.Sprintf("executed started at %s, ended at %s, duration %d millis, target url %s",
		started.UTC().Format(time.RFC3339), ended.UTC().Format(time.RFC3339), ended.UnixMilli()-started.UnixMilli(), url)
	name := types.SanitizeNonAlphabet(scenario.Group, "_") + "_" + scenario.BuildName(string(scenario.Method))
	if err = sr.saveHar(scenario, url, host, started, ended, name); err != nil {
		return err
	}
	if err = sr.saveHistory(scenario, name); err != nil {
		return err
	}

	sr.checkHistoryLimit()
	log.WithFields(log.Fields{
		"Component": "FileScenarioRepository",
		"MaxLimit":  sr.maxHistory,
		"Dir":       sr.historyDir,
		"File":      name,
		"Error":     err,
	}).Debugf("saving history scenario")
	return
}

func (sr *FileAPIScenarioRepository) saveHistory(scenario *types.APIScenario, name string) error {
	b, err := yaml.Marshal(scenario)
	if err != nil {
		return err
	}
	fileName := filepath.Join(sr.historyDir, name+types.ScenarioExt)
	return os.WriteFile(fileName, b, 0644)
}

func (sr *FileAPIScenarioRepository) saveHar(
	scenario *types.APIScenario,
	url string,
	host string,
	started time.Time,
	ended time.Time,
	name string) error {
	harScenario := har.BuildHar(sr.config, scenario, url, host, started, ended)
	b, err := json.Marshal(harScenario)
	if err != nil {
		return err
	}
	fileName := filepath.Join(sr.harDir, name+types.HarExt)
	return os.WriteFile(fileName, b, 0644)
}

// LoadHistory loads scenario
func (sr *FileAPIScenarioRepository) LoadHistory(name string) (*types.APIScenario, error) {
	if !strings.HasSuffix(name, types.ScenarioExt) {
		name = name + types.ScenarioExt
	}
	fileName := filepath.Join(sr.historyDir, name)
	b, err := os.ReadFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s due to %w", fileName, err)
	}

	data := make(map[string]any)

	return unmarshalMockScenario(b, sr.historyDir, data)
}

// LoadHar loads HAR file for the executed history
func (sr *FileAPIScenarioRepository) LoadHar(name string, group string, page int, limit int) ([]har.Har, error) {
	if name != "" {
		res, err := sr.loadHarByName(name)
		if err != nil {
			return nil, err
		}
		return []har.Har{res}, nil
	}
	names := sr.HistoryNames(group)
	harByGroup := make(map[string]*har.Har)
	entries := 0
	for i, name := range names {
		if entries >= limit {
			break
		}
		if page > 0 && i < page*limit {
			continue
		}

		if loadedHar, err := sr.loadHarByName(name); err == nil {
			entries++
			loadedPage := loadedHar.Log.Pages[0]
			prevHar := harByGroup[loadedPage.ID]
			if prevHar == nil {
				harByGroup[loadedPage.ID] = &loadedHar
			} else {
				if loadedPage.StartedDateTime < prevHar.Log.Pages[0].StartedDateTime {
					prevHar.Log.Pages[0].StartedDateTime = loadedPage.StartedDateTime
				}
				for _, entry := range loadedHar.Log.Entries {
					prevHar.Log.Entries = append(prevHar.Log.Entries, entry)
				}
			}
		} else {
			log.WithFields(log.Fields{
				"Component": "FileScenarioRepository",
				"Name":      name,
				"Error":     err,
			}).Warnf("failed to read HAR file")
		}
	}
	res := make([]har.Har, 0)
	for _, next := range harByGroup {
		res = append(res, *next)
	}
	return res, nil
}

func (sr *FileAPIScenarioRepository) loadHarByName(name string) (h har.Har, err error) {
	if !strings.HasSuffix(name, types.HarExt) {
		name = name + types.HarExt
	}
	fileName := filepath.Join(sr.harDir, name)
	var b []byte
	b, err = os.ReadFile(fileName)
	if err != nil {
		return
	}
	err = json.Unmarshal(b, &h)
	if err != nil {
		return
	}
	return
}

// ///////// PRIVATE METHODS //////////////

func (sr *FileAPIScenarioRepository) checkHistoryLimit() {
	infos := sr.historyFiles()
	if len(infos) <= sr.maxHistory {
		return
	}
	for i := len(infos) - 1; i >= sr.maxHistory; i-- {
		harFile := filepath.Join(sr.harDir, strings.ReplaceAll(infos[i].Name(), types.ScenarioExt, types.HarExt))
		err1 := os.Remove(harFile)
		historyFile := filepath.Join(sr.historyDir, infos[i].Name())
		err2 := os.Remove(historyFile)
		log.WithFields(log.Fields{
			"Component":    "FileScenarioRepository",
			"MaxLimit":     sr.maxHistory,
			"InfoSize":     len(infos),
			"Dir":          sr.historyDir,
			"HarFile":      harFile,
			"HistoryFile":  historyFile,
			"I":            i,
			"HarError":     err1,
			"HistoryError": err2,
		}).Debugf("removing old history/har scenario")
	}
}

func (sr *FileAPIScenarioRepository) historyFiles() (infos []fs.FileInfo) {
	files, err := os.ReadDir(sr.historyDir)
	if err != nil {
		log.WithFields(log.Fields{
			"Component": "FileScenarioRepository",
			"Error":     err,
		}).Warnf("failed to read history dir")
		return nil
	}
	for _, file := range files {
		if !file.IsDir() {
			if info, err := file.Info(); err == nil {
				infos = append(infos, info)
			}
		}
	}
	sort.Slice(infos, func(i, j int) bool {
		info1 := infos[i]
		info2 := infos[j]
		return info1.ModTime().After(info2.ModTime())
	})
	return
}

func unmarshalMockScenario(
	b []byte,
	dir string,
	params any) (scenario *types.APIScenario, err error) {
	// parse template
	b, err = fuzz.ParseTemplate(dir, b, params)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template due to %w", err)
	}

	// unmarshal scenario from template output
	scenario = &types.APIScenario{}
	if err = yaml.Unmarshal(b, scenario); err != nil {
		return nil, fmt.Errorf("failed to unmarshal due to %w", err)
	}
	return scenario, nil
}

// visit all scenarios matching properties
func (sr *FileAPIScenarioRepository) visit(
	callback func(keyData *types.APIKeyData) bool) error {
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
			return fmt.Errorf("visit failed to load '%s' due to %w", path, err)
		}
		keyData.LastUsageTime = info.ModTime().Unix()
		if callback(keyData) {
			return errStop
		}
		return
	}

	err := filepath.Walk(sr.contractDir, walkFunc)
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

func (sr *FileAPIScenarioRepository) addKeyData(keyData *types.APIKeyData) {
	sr.mutex.Lock()
	defer func() {
		sr.mutex.Unlock()
	}()
	keyMap := sr.keysByMethodPath[keyData.PartialMethodPathKey()]
	if keyMap == nil {
		keyMap = make(map[string]*types.APIKeyData)
		sr.keysByMethodPath[keyData.PartialMethodPathKey()] = keyMap
	}
	keyMap[keyData.MethodNamePathPrefixKey()] = keyData
	log.WithFields(log.Fields{
		"String":    keyData.String(),
		"Name":      keyData.Name,
		"Method":    keyData.Method,
		"Path":      keyData.Path,
		"Predicate": keyData.Predicate,
		"AllSize":   len(sr.keysByMethodPath),
		"Size":      len(keyMap),
	}).Debugf("registered scenario")
}

func (sr *FileAPIScenarioRepository) buildFileName(
	method types.MethodType,
	scenarioName string,
	path string) string {
	return buildFileName(sr.contractDir, method, scenarioName, path) + types.ScenarioExt
}

func (sr *FileAPIScenarioRepository) buildDir(
	method types.MethodType,
	path string) string {
	return buildDir(sr.contractDir, method, path)
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

func addQueryParams(queryParams map[string]string, data map[string]any) {
	for k, v := range queryParams {
		data[k] = v
	}
}

func unmarshalScenarioKeyData(data []byte) (keyData *types.APIKeyData, err error) {
	rawYaml := string(data)
	ndx := strings.Index(rawYaml, "response:")
	if ndx != -1 {
		rawYaml = rawYaml[0:ndx]
	}
	scenario := &types.APIScenario{}
	err = yaml.Unmarshal([]byte(rawYaml), scenario)
	if err != nil {
		return nil, err
	}
	keyData = scenario.ToKeyData()
	if err := keyData.Validate(); err != nil {
		return nil, err
	}
	return keyData, nil
}

func filterScenariosByPredicate(
	all []*types.APIKeyData, target *types.APIKeyData) (matched []*types.APIKeyData) {
	if len(all) == 0 {
		return all
	}
	sumReqCount := sumRequestCount(all)

	for _, next := range all {
		if utils.MatchScenarioPredicate(next, target, sumReqCount) {
			matched = append(matched, next)
		}
	}
	if len(matched) == 0 {
		return all
	}
	return
}

func sumRequestCount(all []*types.APIKeyData) uint64 {
	sumReqCount := uint64(0)
	for _, next := range all {
		sumReqCount += next.RequestCount
	}
	return sumReqCount
}

func sortByUsageTime(res []*types.APIKeyData) {
	sort.Slice(res, func(i, j int) bool {
		if res[i].LastUsageTime == res[j].LastUsageTime {
			return res[i].Name < res[j].Name
		}
		return res[i].LastUsageTime < res[j].LastUsageTime
	})
}

func buildContractsDir(config *types.Configuration) string {
	contractDir := filepath.Join(config.DataDir, "api_contracts")
	return contractDir
}
