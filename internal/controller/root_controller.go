package controller

import (
	"github.com/bhatti/api-mock-service/internal/contract"
	"github.com/bhatti/api-mock-service/internal/web"
)

// RootController structure
type RootController struct {
	consumerExecutor *contract.ConsumerExecutor
}

// NewRootController instantiates controller for updating api-scenarios
func NewRootController(
	consumerExecutor *contract.ConsumerExecutor,
	webserver web.Server) *RootController {
	ctrl := &RootController{
		consumerExecutor: consumerExecutor,
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

// swagger:route POST /{path} consumer-contract postRoot
// Play API scenario for previously defined API contract based on POST request.
// responses: returns stubbed response based on API
func (r *RootController) postRoot(c web.APIContext) (err error) {
	return r.consumerExecutor.Execute(c)
}

// swagger:route PUT /{path} consumer-contract putRoot
// Play API scenario for previously defined API contract based on PUT request.
// responses: returns stubbed response based on API
func (r *RootController) putRoot(c web.APIContext) (err error) {
	return r.consumerExecutor.Execute(c)
}

// swagger:route GET /{path} consumer-contract getRoot
// Play API scenario for previously defined API contract based on GET request.
// responses: returns stubbed response based on API
func (r *RootController) getRoot(c web.APIContext) (err error) {
	return r.consumerExecutor.Execute(c)
}

// swagger:route DELETE /{path} consumer-contract deleteRoot
// Play API scenario for previously defined API contract based on DELETE request.
// responses: returns stubbed response based on API
func (r *RootController) deleteRoot(c web.APIContext) (err error) {
	return r.consumerExecutor.Execute(c)
}

// swagger:route CONNECT /{path} consumer-contract connectRoot
// Play API scenario for previously defined API contract based on CONNECT request.
// responses: returns stubbed response based on API
func (r *RootController) connectRoot(c web.APIContext) (err error) {
	return r.consumerExecutor.Execute(c)
}

// swagger:route HEAD /{path} consumer-contract headRoot
// Play API scenario for previously defined API contract based on HEAD request.
// responses: returns stubbed response based on API
func (r *RootController) headRoot(c web.APIContext) (err error) {
	return r.consumerExecutor.Execute(c)
}

// swagger:route OPTIONS /{path} consumer-contract optionsRoot
// Play API scenario for previously defined API contract based on OPTIONS request.
// responses: returns stubbed response based on API
func (r *RootController) optionsRoot(c web.APIContext) (err error) {
	return r.consumerExecutor.Execute(c)
}

// swagger:route PATCH /{path} consumer-contract patchRoot
// Play API scenario for previously defined API contract based on PATCH request.
// responses: returns stubbed response based on API
func (r *RootController) patchRoot(c web.APIContext) (err error) {
	return r.consumerExecutor.Execute(c)
}

// swagger:route TRACE /{path} consumer-contract traceRoot
// Play API scenario for previously defined API contract based on TRACE request.
// responses: returns stubbed response based on API
func (r *RootController) traceRoot(c web.APIContext) (err error) {
	return r.consumerExecutor.Execute(c)
}

// swagger:parameters postRoot putRoot getRoot deleteRoot postRoot putRoot getRoot connectRoot headRoot optionsRoot patchRoot traceRoot
// The parameters for consumer-based API testing by path
type rootPathParams struct {
	// in:path
	Path string `json:"path"`
}
