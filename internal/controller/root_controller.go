package controller

import (
	"github.com/bhatti/api-mock-service/internal/contract"
	"github.com/bhatti/api-mock-service/internal/web"
)

// RootController structure
type RootController struct {
	player *contract.ConsumerExecutor
}

// NewRootController instantiates controller for updating mock-scenarios
func NewRootController(
	player *contract.ConsumerExecutor,
	webserver web.Server) *RootController {
	ctrl := &RootController{
		player: player,
	}

	webserver.GET("/:path", ctrl.getRoot)
	webserver.PUT("/:path", ctrl.putRoot)
	webserver.POST("/:path", ctrl.postRoot)
	webserver.DELETE("/:path", ctrl.deleteRoot)
	webserver.CONNECT("/:path", ctrl.connectRoot)
	webserver.HEAD("/:path", ctrl.headRoot)
	webserver.OPTIONS("/:path", ctrl.optionsRoot)
	webserver.PATCH("/:path", ctrl.patchRoot)
	webserver.TRACE("/:path", ctrl.traceRoot)

	webserver.GET("/", ctrl.getRoot)
	webserver.PUT("/", ctrl.putRoot)
	webserver.POST("/", ctrl.postRoot)
	webserver.DELETE("/", ctrl.deleteRoot)
	webserver.CONNECT("/", ctrl.connectRoot)
	webserver.HEAD("/", ctrl.headRoot)
	webserver.OPTIONS("/", ctrl.optionsRoot)
	webserver.PATCH("/", ctrl.patchRoot)
	webserver.TRACE("/", ctrl.traceRoot)

	return ctrl
}

// ********************************* HTTP Handlers ***********************************

// swagger:route POST /{path} mock-play postRoot
// Play scenario from POST request
// responses: returns stubbed response based on API
func (r *RootController) postRoot(c web.APIContext) (err error) {
	return r.player.Execute(c)
}

// swagger:route PUT /{path} mock-play putRoot
// Play scenario from PUT request
// responses: returns stubbed response based on API
func (r *RootController) putRoot(c web.APIContext) (err error) {
	return r.player.Execute(c)
}

// swagger:route GET /{path} mock-play getRoot
// Play scenario from GET request
// responses: returns stubbed response based on API
func (r *RootController) getRoot(c web.APIContext) (err error) {
	return r.player.Execute(c)
}

// swagger:route DELETE /{path} mock-play deleteRoot
// Play scenario from DELETE request
// responses: returns stubbed response based on API
func (r *RootController) deleteRoot(c web.APIContext) (err error) {
	return r.player.Execute(c)
}

// swagger:route CONNECT /{path} mock-play connectRoot
// Play scenario from CONNECT request
// responses: returns stubbed response based on API
func (r *RootController) connectRoot(c web.APIContext) (err error) {
	return r.player.Execute(c)
}

// swagger:route HEAD /{path} mock-play headRoot
// Play scenario from HEAD request
// responses: returns stubbed response based on API
func (r *RootController) headRoot(c web.APIContext) (err error) {
	return r.player.Execute(c)
}

// swagger:route OPTIONS /{path} mock-play optionsRoot
// Play scenario from OPTIONS request
// responses: returns stubbed response based on API
func (r *RootController) optionsRoot(c web.APIContext) (err error) {
	return r.player.Execute(c)
}

// swagger:route PATCH /{path} mock-play patchRoot
// Play scenario from PATCH request
// responses: returns stubbed response based on API
func (r *RootController) patchRoot(c web.APIContext) (err error) {
	return r.player.Execute(c)
}

// swagger:route TRACE /{path} mock-play traceRoot
// Play scenario from TRACE request
// responses: returns stubbed response based on API
func (r *RootController) traceRoot(c web.APIContext) (err error) {
	return r.player.Execute(c)
}

// swagger:parameters postRoot putRoot getRoot deleteRoot
// The parameters for playing APIs by path
type rootPathParams struct {
	// in:path
	Path string `json:"path"`
}
