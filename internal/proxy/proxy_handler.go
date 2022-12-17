package proxy

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/bhatti/api-mock-service/internal/repository"
	"github.com/bhatti/api-mock-service/internal/types"
	"github.com/bhatti/api-mock-service/internal/utils"
	"github.com/bhatti/api-mock-service/internal/web"
	"github.com/elazarl/goproxy"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
)

// Handler structure
type Handler struct {
	config                 *types.Configuration
	mockScenarioRepository repository.MockScenarioRepository
	fixtureRepository      repository.MockFixtureRepository
	adapter                web.Adapter
}

// NewProxyHandler instantiates controller for updating mock-scenarios
func NewProxyHandler(
	config *types.Configuration,
	mockScenarioRepository repository.MockScenarioRepository,
	fixtureRepository repository.MockFixtureRepository,
	adapter web.Adapter,
) *Handler {
	return &Handler{
		config:                 config,
		mockScenarioRepository: mockScenarioRepository,
		fixtureRepository:      fixtureRepository,
		adapter:                adapter,
	}
}

// Start runs the proxy server on a given port
func (h *Handler) Start() error {
	proxy := goproxy.NewProxyHttpServer()
	proxy.OnRequest(proxyCondition()).HandleConnect(goproxy.AlwaysMitm)
	proxy.OnRequest(proxyCondition()).DoFunc(h.handleRequest)
	proxy.OnResponse(proxyCondition()).DoFunc(h.handleResponse)
	proxy.Verbose = false
	return http.ListenAndServe(fmt.Sprintf(":%d", h.config.ProxyPort), proxy)
}

func (h *Handler) handleRequest(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
	req, res, err := h.doHandleRequest(req, ctx)
	var notFoundError *types.NotFoundError
	if err != nil && !errors.As(err, &notFoundError) {
		log.WithFields(log.Fields{
			"Path":   req.URL,
			"Method": req.Method,
			"Error":  err,
		}).Warnf("proxy server failed to fuzz existing mock scenario")
	}
	return req, res
}

func (h *Handler) doHandleRequest(req *http.Request, _ *goproxy.ProxyCtx) (*http.Request, *http.Response, error) {
	res, err := h.adapter.Invoke(req)
	if err == nil && res != nil {
		log.WithFields(log.Fields{
			"Path":    req.URL,
			"Method":  req.Method,
			"Headers": req.Header,
		}).Infof("proxy server redirected request to internal controllers")
		req.Header[types.MockRecordMode] = []string{types.MockRecordModeDisabled}
		return req, res, nil
	}
	key, err := web.BuildMockScenarioKeyData(req)
	if err != nil {
		return req, nil, err
	}

	matchedScenario, err := h.mockScenarioRepository.Lookup(key, nil)
	log.WithFields(log.Fields{
		"Path":            req.URL,
		"Method":          req.Method,
		"Headers":         req.Header,
		"MatchedScenario": matchedScenario,
	}).Infof("proxy server request received")
	if err != nil {
		return req, nil, err
	}
	respHeader := make(http.Header)
	respBody, err := addMockResponse(req.Header, respHeader, matchedScenario, h.fixtureRepository)
	if err != nil {
		return req, nil, err
	}
	req.Header[types.MockRecordMode] = []string{types.MockRecordModeDisabled}

	resp := &http.Response{}
	resp.Request = req
	resp.TransferEncoding = req.TransferEncoding
	resp.Header = respHeader
	resp.Header.Add(types.ContentTypeHeader, matchedScenario.Response.ContentType())
	resp.StatusCode = matchedScenario.Response.StatusCode
	resp.Status = http.StatusText(matchedScenario.Response.StatusCode)
	buf := bytes.NewBuffer(respBody)
	resp.ContentLength = int64(buf.Len())
	resp.Body = io.NopCloser(buf)
	return req, resp, nil
}

func (h *Handler) handleResponse(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
	resp, err := h.doHandleResponse(resp, ctx)
	if err != nil {
		log.WithFields(log.Fields{
			"Path":   resp.Request.URL,
			"Method": resp.Request.Method,
			"Error":  err,
		}).Warnf("proxy server failed to record mock scenario")
	}
	return resp
}

func (h *Handler) doHandleResponse(resp *http.Response, _ *goproxy.ProxyCtx) (*http.Response, error) {
	if resp == nil || resp.Request == nil || len(resp.Request.Header) == 0 ||
		resp.Request.Header.Get(types.MockRecordMode) == types.MockRecordModeDisabled {
		log.WithFields(log.Fields{}).Debugf("proxy server returning canned response")

		return resp, nil
	}

	log.WithFields(log.Fields{
		"Response": resp,
	}).Infof("proxy server response received")

	var reqBytes []byte
	var err error
	switch resp.Request.Body.(type) {
	case utils.ResetReader:
		_ = resp.Request.Body.(utils.ResetReader).Reset()
	}

	reqBytes, resp.Request.Body, err = utils.ReadAll(resp.Request.Body)
	if err != nil {
		return resp, err
	}

	var resBytes []byte
	resBytes, resp.Body, err = utils.ReadAll(resp.Body)
	if err != nil {
		return resp, err
	}

	resContentType, err := saveMockResponse(
		h.config, resp.Request.URL, resp.Request, reqBytes, resBytes, resp.Header, resp.StatusCode, h.mockScenarioRepository)
	if err != nil {
		return resp, err
	}
	resp.Body = io.NopCloser(bytes.NewReader(resBytes))
	resp.Header[types.ContentTypeHeader] = []string{resContentType}
	log.WithFields(log.Fields{
		"Response": resp,
		"Length":   len(resBytes),
		"Headers":  resp.Header,
	}).Infof("proxy server recorded response")
	return resp, nil
}

func proxyCondition() goproxy.ReqConditionFunc {
	return func(_ *http.Request, _ *goproxy.ProxyCtx) bool {
		return true
	}
}
