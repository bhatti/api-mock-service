package repository

import (
	"github.com/bhatti/api-mock-service/internal/fuzz"
	"io/fs"
	"io/ioutil"
	"os"
	"strings"

	"github.com/bhatti/api-mock-service/internal/types"
)

// FileMockFixtureRepository  implements storage for contents using local files
type FileMockFixtureRepository struct {
	dir string
}

// NewFileFixtureRepository creates new instance for content repository
func NewFileFixtureRepository(
	config *types.Configuration,
) (*FileMockFixtureRepository, error) {
	dir := buildContractsDir(config)
	if err := mkdir(dir); err != nil {
		return nil, err
	}
	return &FileMockFixtureRepository{
		dir: dir,
	}, nil
}

// Get contents by id
func (cr *FileMockFixtureRepository) Get(
	method types.MethodType,
	name string,
	path string) ([]byte, error) {
	fileName := cr.buildFileName(method, name, path)
	return os.ReadFile(fileName)
}

// GetFixtureNames returns list of fixture names for given Method and Path
func (cr *FileMockFixtureRepository) GetFixtureNames(
	method types.MethodType,
	path string) (names []string, err error) {
	var files []fs.FileInfo
	dir := cr.buildDir(method, path)
	if files, err = ioutil.ReadDir(dir); err != nil {
		return nil, err
	}

	for _, file := range files {
		name := file.Name()
		if strings.HasSuffix(name, fuzz.FixtureDataExt) {
			trimSize := len(name) - len(fuzz.FixtureDataExt)
			names = append(names, name[0:trimSize])
		}
	}
	return
}

// Save contents
func (cr *FileMockFixtureRepository) Save(
	method types.MethodType,
	name string,
	path string,
	content []byte) error {
	dir := cr.buildDir(method, path)
	if err := mkdir(dir); err != nil {
		return err
	}
	fileName := cr.buildFileName(method, name, path)
	return os.WriteFile(fileName, content, 0644)
}

// Delete removes a job
func (cr *FileMockFixtureRepository) Delete(
	method types.MethodType,
	name string,
	path string) error {
	fileName := cr.buildFileName(method, name, path)
	return os.Remove(fileName)
}

// PRIVATE METHODS
func (cr *FileMockFixtureRepository) buildFileName(
	method types.MethodType,
	scenarioName string,
	path string) string {
	return buildFileName(cr.dir, method, scenarioName, path) + fuzz.FixtureDataExt
}

func (cr *FileMockFixtureRepository) buildDir(
	method types.MethodType,
	path string) string {
	return buildDir(cr.dir, method, path)
}
