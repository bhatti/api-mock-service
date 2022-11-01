package proxy

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/bhatti/api-mock-service/internal/repository"
	"github.com/bhatti/api-mock-service/internal/types"
	"github.com/bhatti/api-mock-service/internal/web"
	"github.com/elazarl/goproxy"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
)

// Handler structure
type Handler struct {
	mockScenarioRepository repository.MockScenarioRepository
	fixtureRepository      repository.MockFixtureRepository
	proxyPort              int
}

// NewProxyHandler instantiates controller for updating mock-scenarios
func NewProxyHandler(
	proxyPort int,
	mockScenarioRepository repository.MockScenarioRepository,
	fixtureRepository repository.MockFixtureRepository,
) *Handler {
	return &Handler{
		proxyPort:              proxyPort,
		mockScenarioRepository: mockScenarioRepository,
		fixtureRepository:      fixtureRepository,
	}
}

func (h *Handler) Start() error {
	proxy := goproxy.NewProxyHttpServer()
	proxy.OnRequest(proxyCondition()).HandleConnect(goproxy.AlwaysMitm)
	proxy.OnRequest(proxyCondition()).DoFunc(h.handleRequest)
	proxy.OnResponse(proxyCondition()).DoFunc(h.handleResponse)
	proxy.Verbose = true
	return http.ListenAndServe(fmt.Sprintf(":%d", h.proxyPort), proxy)
}

func (h *Handler) handleRequest(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
	var validationErr *types.ValidationError
	req, res, err := h.doHandleRequest(req, ctx)
	if err != nil && errors.As(err, &validationErr) {
		log.WithFields(log.Fields{
			"Path":   req.URL,
			"Method": req.Method,
			"Error":  err,
		}).Warnf("proxy server failed to handle mock scenario")
	}
	return req, res
}

func (h *Handler) doHandleRequest(req *http.Request, _ *goproxy.ProxyCtx) (*http.Request, *http.Response, error) {
	log.WithFields(log.Fields{
		"Path":   req.URL,
		"Method": req.Method,
	}).Infof("proxy server request received")
	key, err := web.BuildMockScenarioKeyData(req)
	if err != nil {
		return req, nil, err
	}

	matchedScenario, err := h.mockScenarioRepository.Lookup(key)
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
	resp.Header.Add("Content-Type", matchedScenario.Response.ContentType)
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
	log.WithFields(log.Fields{
		"Response": resp,
	}).Infof("proxy server response received")
	if resp == nil || resp.Request == nil || len(resp.Request.Header) == 0 ||
		resp.Request.Header.Get(types.MockRecordMode) == types.MockRecordModeDisabled {
		return resp, nil
	}
	var reqBody []byte
	var err error
	if resp.Request.Body != nil {
		reqBody, err = io.ReadAll(resp.Request.Body)
		if err != nil {
			return resp, err
		}
	}

	resBytes, resContentType, err := saveMockResponse(
		resp.Request.URL, resp.Request, reqBody, resp.Body, resp.Header, resp.StatusCode, h.mockScenarioRepository)
	if err != nil {
		return resp, err
	}
	resp.Body = io.NopCloser(bytes.NewReader(resBytes))
	resp.Header["Content-Type"] = []string{resContentType}
	return resp, nil
}

func proxyCondition() goproxy.ReqConditionFunc {
	return func(_ *http.Request, _ *goproxy.ProxyCtx) bool {
		return true
	}
}
