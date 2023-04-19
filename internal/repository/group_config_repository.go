package repository

import (
	"github.com/bhatti/api-mock-service/internal/types"
)

// GroupConfigRepository defines data store for group config
type GroupConfigRepository interface {
	// Variables returns variables for given name
	Variables(name string) map[string]string

	// Save saves group config
	Save(name string, gc *types.GroupConfig) (err error)

	// Load loads group config
	Load(name string) (*types.GroupConfig, error)

	// Delete removes group config
	Delete(name string) error
}
