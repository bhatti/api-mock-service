package controller

import (
	"github.com/bhatti/api-mock-service/internal/proxy"
	"github.com/bhatti/api-mock-service/internal/web"
)

// RootController structure
type RootController struct {
	player *proxy.Player
}

// NewRootController instantiates controller for updating mock-scenarios
func NewRootController(
	player *proxy.Player,
	webserver web.Server) *RootController {
	ctrl := &RootController{
		player: player,
	}

	webserver.GET("/:path", ctrl.getRoot)
	webserver.PUT("/:path", ctrl.putRoot)
	webserver.POST("/:path", ctrl.postRoot)
	webserver.DELETE("/:path", ctrl.deleteRoot)
	return ctrl
}

// ********************************* HTTP Handlers ***********************************

// swagger:route POST /_proxy mock-proxy postRoot
// Records scenario from POST request
// responses: returns original response based on API
func (r *RootController) postRoot(c web.APIContext) (err error) {
	return r.player.Handle(c)
}

// swagger:route PUT /_proxy mock-proxy putRoot
// Records scenario from PUT request
// responses: returns original response based on API
func (r *RootController) putRoot(c web.APIContext) (err error) {
	return r.player.Handle(c)
}

// swagger:route GET /_proxy mock-proxy getRoot
// Records scenario from GET request
// responses: returns original response based on API
func (r *RootController) getRoot(c web.APIContext) (err error) {
	return r.player.Handle(c)
}

// swagger:route DELETE /_proxy mock-proxy deleteRoot
// Records scenario from DELETE request
// responses: returns original response based on API
func (r *RootController) deleteRoot(c web.APIContext) (err error) {
	return r.player.Handle(c)
}

// swagger:parameters postRoot putRoot getRoot deleteRoot
// The parameters for playing APIs by path
type rootPathParams struct {
	// in:path
	Path string `json:"path"`
}
