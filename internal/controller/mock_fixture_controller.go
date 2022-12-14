package controller

import (
	"fmt"
	"github.com/bhatti/api-mock-service/internal/utils"
	"net/http"

	"github.com/bhatti/api-mock-service/internal/types"

	"github.com/bhatti/api-mock-service/internal/repository"
	"github.com/bhatti/api-mock-service/internal/web"

	log "github.com/sirupsen/logrus"
)

// MockFixtureController structure
type MockFixtureController struct {
	fixtureRepository repository.MockFixtureRepository
}

// NewMockFixtureController instantiates controller for updating mock-fixtures
func NewMockFixtureController(
	fixtureRepository repository.MockFixtureRepository,
	webserver web.Server) *MockFixtureController {
	ctrl := &MockFixtureController{
		fixtureRepository: fixtureRepository,
	}

	webserver.GET("/_fixtures/:method/fixtures/:path", ctrl.getMockFixtureNames)
	webserver.GET("/_fixtures/:method/:name/:path", ctrl.GetMockFixture)
	webserver.POST("/_fixtures/:method/:name/:path", ctrl.PostMockFixture)
	webserver.DELETE("/_fixtures/:method/:name/:path", ctrl.DeleteMockFixture)
	return ctrl
}

// ********************************* HTTP Handlers ***********************************

// PostMockFixture handler
// swagger:route POST /_fixtures/{method}/{name}/{path} mock-fixtures PostMockFixture
// Creates new mock fixtures based on request body.
// responses:
//
//	200: mockFixtureResponse
func (msc *MockFixtureController) PostMockFixture(c web.APIContext) (err error) {
	var data []byte
	data, c.Request().Body, err = utils.ReadAll(c.Request().Body)
	if err != nil {
		return err
	}
	method, err := types.ToMethod(c.Param("method"))
	if err != nil {
		return err
	}
	name := c.Param("name")
	if name == "" {
		return fmt.Errorf("name not specified")
	}
	path := c.Param("path")
	if path == "" {
		return fmt.Errorf("path not specified")
	}
	log.WithFields(log.Fields{
		"Name": name,
		"Path": path,
	}).Debugf("saving fixture...")
	if err = msc.fixtureRepository.Save(method, name, path, data); err != nil {
		return err
	}
	return c.NoContent(http.StatusOK)
}

// GetMockFixture handler
// swagger:route GET /_fixtures/{method}/{name}/{path} mock-fixtures GetMockFixture
// Finds an existing mock fixtures based on name and path
// responses:
//
//	200: mockFixtureResponse
func (msc *MockFixtureController) GetMockFixture(c web.APIContext) error {
	method, err := types.ToMethod(c.Param("method"))
	if err != nil {
		return err
	}
	name := c.Param("name")
	if name == "" {
		return fmt.Errorf("name not specified")
	}
	path := c.Param("path")
	if path == "" {
		return fmt.Errorf("path not specified")
	}
	log.WithFields(log.Fields{
		"Name": name,
		"Path": path,
	}).Debugf("getting mock fixture...")
	b, err := msc.fixtureRepository.Get(
		method,
		name,
		path)
	if err != nil {
		return err
	}
	return c.Blob(http.StatusOK, "application/binary", b)
}

// swagger:route GET /_fixtures/{method}/fixtures/{path} mock-fixtures getMockFixtureNames
// Returns mock fixture names
// responses:
//
//	200: mockFixtureNamesResponse
func (msc *MockFixtureController) getMockFixtureNames(c web.APIContext) error {
	method, err := types.ToMethod(c.Param("method"))
	if err != nil {
		return err
	}
	path := c.Param("path")
	if path == "" {
		return fmt.Errorf("path not specified")
	}
	log.WithFields(log.Fields{
		"Path": path,
	}).Infof("getting mock fixture names...")
	names, err := msc.fixtureRepository.GetFixtureNames(method, path)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, names)
}

// DeleteMockFixture handler
// swagger:route DELETE /_fixtures/{method}/{name}/{path} mock-fixtures GetMockFixture
// Deletes an existing mock fixtures based on name and path.
// responses:
//
//	200: emptyResponse
func (msc *MockFixtureController) DeleteMockFixture(c web.APIContext) error {
	method, err := types.ToMethod(c.Param("method"))
	if err != nil {
		return err
	}
	name := c.Param("name")
	if name == "" {
		return fmt.Errorf("fixture name not specified")
	}
	path := c.Param("path")
	if path == "" {
		return fmt.Errorf("path not specified")
	}
	log.WithFields(log.Fields{
		"Name": name,
		"Path": path,
	}).Infof("deleting mock fixture...")
	err = msc.fixtureRepository.Delete(method, name, path)
	if err != nil {
		return err
	}
	return c.NoContent(http.StatusOK)
}

// ********************************* Swagger types ***********************************

// swagger:parameters PostMockFixture
// The params for mock-fixture
type mockFixtureCreateParams struct {
	// in:path
	Method string `json:"method"`
	// in:path
	Name string `json:"name"`
	// in:path
	Path string `json:"path"`
	// in:body
	Body []byte
}

// MockFixture body for update
// swagger:response mockFixtureResponse
type mockFixtureResponseBody struct {
	// in:body
	Body []byte
}

// MockFixture names
// swagger:response mockFixtureNamesResponse
type mockFixtureNamesResponseBody struct {
	// in:body
	Body []string
}

// swagger:parameters DeleteMockFixture GetMockFixture
// The parameters for finding mock-fixture by name and path
type mockFixtureIDParams struct {
	// in:path
	Method string `json:"method"`
	// in:path
	Name string `json:"name"`
	// in:path
	Path string `json:"path"`
}

// swagger:parameters getMockFixtureNames
// The parameters for finding mock-fixture names by path
type mockFixtureNamesParams struct {
	// in:path
	Method string `json:"method"`
	// in:path
	Path string `json:"path"`
}

// swagger:response emptyResponse
// Empty response
type emptyResponse struct {
}
