package repository

import (
	"context"
	"fmt"
	"github.com/bhatti/api-mock-service/internal/types"
	"github.com/getkin/kin-openapi/openapi3"
	log "github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const oapiExt = ".json"

// FileOAPIRepository  implements storage for contents using local files
type FileOAPIRepository struct {
	dir string
}

// NewFileOAPIRepository creates new instance for OAPIRepository
func NewFileOAPIRepository(
	config *types.Configuration,
) (*FileOAPIRepository, error) {
	dir := filepath.Join(config.DataDir, "oapi_specs")
	if err := mkdir(dir); err != nil {
		return nil, err
	}
	return &FileOAPIRepository{
		dir: dir,
	}, nil
}

// GetNames returns list of open-api spec names
func (or *FileOAPIRepository) GetNames() (names []string) {
	files, err := os.ReadDir(or.dir)
	if err != nil {
		log.WithFields(log.Fields{
			"Component": "FileOAPIRepository",
			"Error":     err,
		}).Warnf("failed to read open-api specs dir")
		return nil
	}
	names = make([]string, 0)
	for _, file := range files {
		if !file.IsDir() {
			if info, err := file.Info(); err == nil {
				names = append(names, strings.ReplaceAll(info.Name(), oapiExt, ""))
			}
		}
	}
	sort.Slice(names, func(i, j int) bool {
		return names[i] < names[j]
	})
	return
}

// Save saves OAPI spec
func (or *FileOAPIRepository) Save(name string, t *openapi3.T) (err error) {
	b, err := t.MarshalJSON()
	if err != nil {
		return err
	}
	return or.SaveRaw(name, b)

}

// SaveRaw saves raw spec
func (or *FileOAPIRepository) SaveRaw(name string, data []byte) (err error) {
	if name == "" {
		return fmt.Errorf("oapi spec name is not specified")
	}
	fileName := or.buildName(name)
	return os.WriteFile(fileName, data, 0644)
}

// Load loads OAPI spec
func (or *FileOAPIRepository) Load(name string) (*openapi3.T, error) {
	b, err := or.LoadRaw(name)
	if err != nil {
		return nil, err
	}
	loader := &openapi3.Loader{Context: context.Background(), IsExternalRefsAllowed: true}
	return loader.LoadFromData(b)
}

// LoadRaw loads raw spec
func (or *FileOAPIRepository) LoadRaw(name string) (b []byte, err error) {
	fileName := or.buildName(name)
	return os.ReadFile(fileName)
}

// Delete removes an OAPI spec
func (or *FileOAPIRepository) Delete(name string) error {
	fileName := or.buildName(name)
	return os.Remove(fileName)
}

func (or *FileOAPIRepository) buildName(name string) string {
	if !strings.HasSuffix(name, oapiExt) {
		name += oapiExt
	}
	return filepath.Join(or.dir, name)
}
