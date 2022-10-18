package repository

import "github.com/bhatti/api-mock-service/internal/types"

// MockFixtureRepository defines data store for content for mocking purpose
type MockFixtureRepository interface {
	// Get Content data by id
	Get(
		method types.MethodType,
		name string,
		path string,
	) ([]byte, error)

	// GetFixtureNames returns list of fixture names for given Method and Path
	GetFixtureNames(
		method types.MethodType,
		path string) (names []string, err error)

	// Save Content data
	Save(
		method types.MethodType,
		name string,
		path string,
		contents []byte) (err error)

	// Delete removes a content data
	Delete(
		method types.MethodType,
		name string,
		path string,
	) error
}
