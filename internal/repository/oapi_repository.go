package repository

import (
	"github.com/getkin/kin-openapi/openapi3"
)

// OAPIRepository defines data store for OpenAPI specs
type OAPIRepository interface {
	// GetNames returns list of mock scenarios names
	GetNames() []string

	// Save saves OAPI spec
	Save(name string, t *openapi3.T) (err error)

	// Load loads OAPI spec
	Load(name string) (*openapi3.T, error)

	// SaveRaw saves raw spec
	SaveRaw(name string, data []byte) (err error)

	// LoadRaw loads raw spec
	LoadRaw(name string) (b []byte, err error)

	// Delete removes an OAPI spec
	Delete(name string) error
}
