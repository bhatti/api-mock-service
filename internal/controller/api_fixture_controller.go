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

// APIFixtureController structure
type APIFixtureController struct {
	fixtureRepository repository.APIFixtureRepository
}

// NewAPIFixtureController instantiates controller for updating api-test-fixtures
func NewAPIFixtureController(
	fixtureRepository repository.APIFixtureRepository,
	webserver web.Server) *APIFixtureController {
	ctrl := &APIFixtureController{
		fixtureRepository: fixtureRepository,
	}

	webserver.GET("/_fixtures/:method/fixtures/:path", ctrl.getAPITestFixtureNames)
	webserver.GET("/_fixtures/:method/:name/:path", ctrl.getAPITestFixture)
	webserver.POST("/_fixtures/:method/:name/:path", ctrl.postAPITestFixture)
	webserver.DELETE("/_fixtures/:method/:name/:path", ctrl.deleteAPITestFixture)
	return ctrl
}

// ********************************* HTTP Handlers ***********************************

// postAPITestFixture handler
// swagger:route POST /_fixtures/{method}/{name}/{path} api-test-fixtures postAPITestFixture
// Creates new api-test-fixtures based on request body.
// responses:
//
//	200: apiFixtureResponse
func (msc *APIFixtureController) postAPITestFixture(c web.APIContext) (err error) {
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

// getAPITestFixture handler
// swagger:route GET /_fixtures/{method}/{name}/{path} api-test-fixtures getAPITestFixture
// Finds an existing api-test-fixtures based on name and path
// responses:
//
//	200: apiFixtureResponse
func (msc *APIFixtureController) getAPITestFixture(c web.APIContext) error {
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
	}).Debugf("getting api test-fixture...")
	b, err := msc.fixtureRepository.Get(
		method,
		name,
		path)
	if err != nil {
		return err
	}
	return c.Blob(http.StatusOK, "application/binary", b)
}

// swagger:route GET /_fixtures/{method}/fixtures/{path} api-test-fixtures getAPITestFixtureNames
// Returns api test-fixture names
// responses:
//
//	200: apiFixtureNamesResponse
func (msc *APIFixtureController) getAPITestFixtureNames(c web.APIContext) error {
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
	}).Infof("getting api test-fixture names...")
	names, err := msc.fixtureRepository.GetFixtureNames(method, path)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, names)
}

// deleteAPITestFixture handler
// swagger:route DELETE /_fixtures/{method}/{name}/{path} api-test-fixtures deleteAPITestFixture
// Deletes an existing api-test-fixtures based on name and path.
// responses:
//
//	200: emptyResponse
func (msc *APIFixtureController) deleteAPITestFixture(c web.APIContext) error {
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
	}).Infof("deleting api test-fixture...")
	err = msc.fixtureRepository.Delete(method, name, path)
	if err != nil {
		return err
	}
	return c.NoContent(http.StatusOK)
}

// ********************************* Swagger types ***********************************

// swagger:parameters postAPITestFixture
// The params for api-fixture
type apiFixtureCreateParams struct {
	// in:path
	Method string `json:"method"`
	// in:path
	Name string `json:"name"`
	// in:path
	Path string `json:"path"`
	// in:body
	Body []byte
}

// APIFixture body for update
// swagger:response apiFixtureResponse
type apiFixtureResponseBody struct {
	// in:body
	Body []byte
}

// APIFixture names
// swagger:response apiFixtureNamesResponse
type apiFixtureNamesResponseBody struct {
	// in:body
	Body []string
}

// swagger:parameters deleteAPITestFixture getAPITestFixture
// The parameters for finding api test-fixture by name and path
type apiFixtureIDParams struct {
	// in:path
	Method string `json:"method"`
	// in:path
	Name string `json:"name"`
	// in:path
	Path string `json:"path"`
}

// swagger:parameters getAPITestFixtureNames
// The parameters for finding api test-fixture names by path
type apiFixtureNamesParams struct {
	// in:path
	Method string `json:"method"`
	// in:path
	Path string `json:"path"`
}

// swagger:response emptyResponse
// Empty response
type emptyResponse struct {
}
