package repository

import (
	"encoding/json"
	"fmt"
	"github.com/bhatti/api-mock-service/internal/types"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const groupConfigExt = ".json"

const globalName = "global"
const rootName = "root"

// FileGroupConfigRepository  implements storage for contents using local files
type FileGroupConfigRepository struct {
	dir string
}

// NewFileGroupConfigRepository creates new instance for GroupConfigRepository
func NewFileGroupConfigRepository(
	config *types.Configuration,
) (*FileGroupConfigRepository, error) {
	dir := filepath.Join(config.DataDir, "groups")
	if err := mkdir(dir); err != nil {
		return nil, err
	}
	return &FileGroupConfigRepository{
		dir: dir,
	}, nil
}

// Variables returns variables for given name
func (gcr *FileGroupConfigRepository) Variables(name string) map[string]string {
	names := gcr.getNames()
	res := make(map[string]string)
	for _, next := range names {
		if next == globalName || next == rootName {
			gcr.addVariables(next, res)
		}
	}
	for _, next := range names {
		if strings.HasPrefix(name, next) {
			gcr.addVariables(next, res)
		}
	}
	return res
}

// Save saves GroupConfig spec
func (gcr *FileGroupConfigRepository) Save(name string, gc *types.GroupConfig) (err error) {
	if name == "" {
		return fmt.Errorf("oapi spec name is not specified")
	}
	b, err := json.MarshalIndent(gc, "", "  ")
	if err != nil {
		return err
	}
	fileName := gcr.buildName(name)
	return os.WriteFile(fileName, b, 0644)
}

// Load loads GroupConfig spec
func (gcr *FileGroupConfigRepository) Load(name string) (*types.GroupConfig, error) {
	fileName := gcr.buildName(name)
	b, err := os.ReadFile(fileName)
	if err != nil {
		return nil, err
	}
	gc := &types.GroupConfig{}
	err = json.Unmarshal(b, gc)
	if err != nil {
		return nil, err
	}
	return gc, nil
}

// Delete removes an GroupConfig spec
func (gcr *FileGroupConfigRepository) Delete(name string) error {
	fileName := gcr.buildName(name)
	return os.Remove(fileName)
}

func (gcr *FileGroupConfigRepository) buildName(name string) string {
	if !strings.HasSuffix(name, groupConfigExt) {
		name += groupConfigExt
	}
	return filepath.Join(gcr.dir, name)
}

func (gcr *FileGroupConfigRepository) getNames() (names []string) {
	files, err := os.ReadDir(gcr.dir)
	if err != nil {
		return nil
	}
	names = make([]string, 0)
	for _, file := range files {
		if !file.IsDir() {
			if info, err := file.Info(); err == nil {
				names = append(names, strings.ReplaceAll(info.Name(), groupConfigExt, ""))
			}
		}
	}
	sort.Slice(names, func(i, j int) bool {
		return names[i] < names[j]
	})
	return
}

func (gcr *FileGroupConfigRepository) addVariables(name string, res map[string]string) {
	if gc, err := gcr.Load(name); err == nil {
		for k, v := range gc.Variables {
			res[k] = v
		}
	}
}
