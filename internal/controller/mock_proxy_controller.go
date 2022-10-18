package controller

import (
	"github.com/bhatti/api-mock-service/internal/proxy"
	"github.com/bhatti/api-mock-service/internal/web"
)

// MockProxyController structure
type MockProxyController struct {
	recorder *proxy.Recorder
}

// NewMockProxyController instantiates controller for updating mock-scenarios
func NewMockProxyController(
	recorder *proxy.Recorder,
	webserver web.Server) *MockProxyController {
	ctrl := &MockProxyController{
		recorder: recorder,
	}

	webserver.GET("/_proxy", ctrl.getMockProxy)
	webserver.PUT("/_proxy", ctrl.putMockProxy)
	webserver.POST("/_proxy", ctrl.postMockProxy)
	webserver.DELETE("/_proxy", ctrl.deleteMockProxy)
	return ctrl
}

// ********************************* HTTP Handlers ***********************************

// swagger:route POST /_proxy mock-proxy postMockProxy
// Records scenario from POST request
// responses: returns original response based on API
func (msc *MockProxyController) postMockProxy(c web.APIContext) (err error) {
	return msc.recorder.Handle(c)
}

// swagger:route PUT /_proxy mock-proxy putMockProxy
// Records scenario from PUT request
// responses: returns original response based on API
func (msc *MockProxyController) putMockProxy(c web.APIContext) (err error) {
	return msc.recorder.Handle(c)
}

// swagger:route GET /_proxy mock-proxy getMockProxy
// Records scenario from GET request
// responses: returns original response based on API
func (msc *MockProxyController) getMockProxy(c web.APIContext) (err error) {
	return msc.recorder.Handle(c)
}

// swagger:route DELETE /_proxy mock-proxy deleteMockProxy
// Records scenario from DELETE request
// responses: returns original response based on API
func (msc *MockProxyController) deleteMockProxy(c web.APIContext) (err error) {
	return msc.recorder.Handle(c)
}
