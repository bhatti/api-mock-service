package controller

import (
	"github.com/bhatti/api-mock-service/internal/proxy"
	"github.com/bhatti/api-mock-service/internal/web"
)

// APIProxyController structure
type APIProxyController struct {
	recorder *proxy.Recorder
}

// NewAPIProxyController instantiates controller for updating api-scenarios
func NewAPIProxyController(
	recorder *proxy.Recorder,
	webserver web.Server) *APIProxyController {
	ctrl := &APIProxyController{
		recorder: recorder,
	}

	webserver.GET("/_proxy", ctrl.getAPIProxy)
	webserver.PUT("/_proxy", ctrl.putAPIProxy)
	webserver.POST("/_proxy", ctrl.postAPIProxy)
	webserver.DELETE("/_proxy", ctrl.deleteAPIProxy)
	return ctrl
}

// ********************************* HTTP Handlers ***********************************

// swagger:route POST /_proxy api-proxy postAPIProxy
// Records scenario from POST request
// responses: returns original response based on API
func (msc *APIProxyController) postAPIProxy(c web.APIContext) (err error) {
	return msc.recorder.Handle(c)
}

// swagger:route PUT /_proxy api-proxy putAPIProxy
// Records scenario from PUT request
// responses: returns original response based on API
func (msc *APIProxyController) putAPIProxy(c web.APIContext) (err error) {
	return msc.recorder.Handle(c)
}

// swagger:route GET /_proxy api-proxy getAPIProxy
// Records scenario from GET request
// responses: returns original response based on API
func (msc *APIProxyController) getAPIProxy(c web.APIContext) (err error) {
	return msc.recorder.Handle(c)
}

// swagger:route DELETE /_proxy api-proxy deleteAPIProxy
// Records scenario from DELETE request
// responses: returns original response based on API
func (msc *APIProxyController) deleteAPIProxy(c web.APIContext) (err error) {
	return msc.recorder.Handle(c)
}
