package controller

import (
	"encoding/json"
	"fmt"
	"github.com/bhatti/api-mock-service/internal/repository"
	"github.com/bhatti/api-mock-service/internal/types"
	"github.com/bhatti/api-mock-service/internal/utils"
	"github.com/bhatti/api-mock-service/internal/web"
	"net/http"
)

// GroupConfigController structure
type GroupConfigController struct {
	groupConfigRepository repository.GroupConfigRepository
}

// NewGroupConfigController instantiates controller for updating api-scenarios based on OpenAPI v3
func NewGroupConfigController(
	groupConfigRepository repository.GroupConfigRepository,
	webserver web.Server) *GroupConfigController {
	ctrl := &GroupConfigController{
		groupConfigRepository: groupConfigRepository,
	}

	webserver.GET("/_groups/:group/config", ctrl.getGroupConfig)
	webserver.PUT("/_groups/:group/config", ctrl.putGroupConfig)
	return ctrl
}

// ********************************* HTTP Handlers ***********************************

// getGroupConfig handler
// swagger:route GET /_groups/{group}/config group-config getGroupConfig
// Returns group config
// responses:
//
//	200: groupConfigResponse
func (gcc *GroupConfigController) getGroupConfig(c web.APIContext) (err error) {
	group := c.Param("group")
	if group == "" {
		return fmt.Errorf("scenario group not specified in %s", c.Request().URL)
	}
	gc, err := gcc.groupConfigRepository.Load(group)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, gc)
}

// putGroupConfig handler
// swagger:route PUT /_groups/{group}/config group-config putGroupConfig
// Saves group config
// responses:
//
//	200: putGroupConfigResponse
func (gcc *GroupConfigController) putGroupConfig(c web.APIContext) (err error) {
	group := c.Param("group")
	if group == "" {
		return fmt.Errorf("scenario group not specified in %s", c.Request().URL)
	}
	var data []byte
	data, c.Request().Body, err = utils.ReadAll(c.Request().Body)
	if err != nil {
		return err
	}
	gc := &types.GroupConfig{}
	err = json.Unmarshal(data, gc)
	if err != nil {
		return err
	}
	err = gcc.groupConfigRepository.Save(group, gc)
	if err != nil {
		return err
	}
	return c.NoContent(http.StatusOK)
}

// ********************************* Swagger types ***********************************

// The params for getting group-config
// swagger:parameters getGroupConfig
type getGroupsConfigParams struct {
	// in:path
	Group string `json:"group"`
}

// The params for saving group-config
// swagger:parameters putGroupConfig
type putGroupsConfigParams struct {
	// in:path
	Group string `json:"group"`
	// in:body
	Body types.GroupConfig
}

// APIScenario body for getting group-config
// swagger:response groupConfigResponse
type groupConfigResponseBody struct {
	// in:body
	Body types.GroupConfig
}

// APIScenario body for updating group-config
// swagger:response putGroupConfigResponse
type putGroupConfigResponseBody struct {
}
