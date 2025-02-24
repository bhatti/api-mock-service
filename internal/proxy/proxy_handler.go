package proxy

import (
	"bytes"
	crand "crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/bhatti/api-mock-service/internal/contract"
	"github.com/bhatti/api-mock-service/internal/repository"
	"github.com/bhatti/api-mock-service/internal/types"
	"github.com/bhatti/api-mock-service/internal/utils"
	"github.com/bhatti/api-mock-service/internal/web"
	"github.com/elazarl/goproxy"
	log "github.com/sirupsen/logrus"
	"io"
	"math/big"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

var acceptAllCerts = &tls.Config{InsecureSkipVerify: true}

//var noProxyClient = &http.Client{Transport: &http.Transport{TLSClientConfig: acceptAllCerts}}

var ignoredResponseHeaders = map[string]struct{}{
	"Access-Control-Expose-Headers":   {},
	"Referrer-Policy":                 {},
	"Report-To":                       {},
	"Strict-Transport":                {},
	"Strict-Origin-When-Cross-Origin": {},
	"Strict-Transport-Security":       {},
	"X-Frame-Options":                 {},
	"X-Content-Type-Options":          {},
	"Timing-Allow-Origin":             {},
}

// Handler structure
type Handler struct {
	config                *types.Configuration
	authAdapter           web.AuthAdapter
	scenarioRepository    repository.APIScenarioRepository
	fixtureRepository     repository.APIFixtureRepository
	groupConfigRepository repository.GroupConfigRepository
	adapter               web.Adapter
}

// NewProxyHandler instantiates controller for updating api-scenarios
func NewProxyHandler(
	config *types.Configuration,
	authAdapter web.AuthAdapter,
	scenarioRepository repository.APIScenarioRepository,
	fixtureRepository repository.APIFixtureRepository,
	groupConfigRepository repository.GroupConfigRepository,
	adapter web.Adapter,
) *Handler {
	return &Handler{
		config:                config,
		authAdapter:           authAdapter,
		scenarioRepository:    scenarioRepository,
		fixtureRepository:     fixtureRepository,
		groupConfigRepository: groupConfigRepository,
		adapter:               adapter,
	}
}

// Start runs the proxy server on a given port
func (h *Handler) Start() error {
	proxy := goproxy.NewProxyHttpServer()

	proxy.OnRequest(h.proxyCondition()).HandleConnect(goproxy.AlwaysMitm)
	proxy.OnRequest(h.proxyCondition()).DoFunc(h.handleRequest)
	proxy.OnResponse(h.proxyCondition()).DoFunc(h.handleResponse)
	proxy.Verbose = false
	acceptAllCerts.VerifyPeerCertificate = func([][]byte, [][]*x509.Certificate) error {
		return nil
	}
	proxy.Tr.TLSClientConfig = acceptAllCerts
	//http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	//err := saveProxyCert()
	//if err != nil {
	//	log.WithFields(log.Fields{
	//		"Error": err,
	//	}).Warnf("failed to create proxy cert")
	//}
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
		}).Warnf("proxy server failed to find existing api scenario")
	}
	// let proxy call real server
	return req, res
}

func (h *Handler) doHandleRequest(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response, error) {
	ctx.UserData = time.Now()
	var err error
	_, req.Body, err = utils.ReadAll(req.Body)
	if err != nil {
		log.WithFields(log.Fields{
			"Path":   req.URL,
			"Method": req.Method,
			"Error":  err,
		}).Warnf("proxy server failed to read request body in handl-request")
		return req, nil, err
	}

	switch req.Body.(type) {
	case utils.ResetReader:
		_ = req.Body.(utils.ResetReader).Reset()
	}

	oldAuth := req.Header.Get(types.AuthorizationHeader)
	awsAuthSig4, awsInfo, err := h.authAdapter.HandleAuth(req)

	if awsAuthSig4 {
		log.WithFields(log.Fields{
			"Component":    "DefaultHTTPClient",
			"URL":          req.URL,
			"Method":       req.Method,
			"OldAuth":      oldAuth,
			"AWSInfo":      awsInfo,
			"Error":        err,
			"AWSKey":       web.GetHeaderParamOrEnvValue(nil, web.AWSAccessKey),
			"HasAWSSecret": web.GetHeaderParamOrEnvValue(nil, web.AWSSecretKey) != "",
			//"Header":       req.Header,
		}).Infof("proxy server checked for aws-request (continue)")
		// changed to not return here
	}

	{
		res, err := h.adapter.Invoke(req)
		if err == nil && res != nil {
			log.WithFields(log.Fields{
				"Host":        req.Host,
				"Path":        req.URL,
				"Method":      req.Method,
				"Headers":     req.Header,
				"AWSAuthSig4": awsAuthSig4,
			}).Debugf("proxy server redirected request to internal controllers")
			req.Header[types.MockRecordMode] = []string{types.MockRecordModeDisabled}
			return req, res, nil
		}
	}

	key, err := web.BuildMockScenarioKeyData(req)
	if err != nil {
		return req, nil, err
	}

	matchedScenario, matchErr := h.scenarioRepository.Lookup(key, nil)
	log.WithFields(log.Fields{
		"Host":            req.Host,
		"Path":            req.URL,
		"Method":          req.Method,
		"MatchedScenario": matchedScenario,
		"AWSAuthSig4":     awsAuthSig4,
		"Error":           matchErr,
		//"Headers":         req.Header,
	}).Infof("proxy server request received [playback=%v]", matchedScenario != nil)
	if matchErr != nil {
		return req, nil, matchErr
	}
	respHeader := make(http.Header)
	respBody, sharedVariables, err := contract.AddMockResponse(
		req,
		req.Header,
		respHeader,
		matchedScenario,
		getStartTime(ctx),
		time.Now(),
		h.config,
		h.scenarioRepository,
		h.fixtureRepository,
		h.groupConfigRepository,
	)
	if err != nil {
		return req, nil, err
	}

	req.Header[types.MockPlayback] = []string{fmt.Sprintf("%v", matchedScenario != nil)}
	req.Header[types.MockRecordMode] = []string{types.MockRecordModeDisabled}

	for k, v := range sharedVariables {
		if strV, ok := v.(string); ok {
			respHeader.Set(k, strV)
		}
	}
	resp := &http.Response{}
	resp.Request = req
	resp.TransferEncoding = req.TransferEncoding
	resp.Header = respHeader
	resp.Header.Set(types.ContentTypeHeader, matchedScenario.Response.ContentType(""))

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
		}).Warnf("proxy server failed to record api scenario")
	}
	return resp
}

func (h *Handler) doHandleResponse(resp *http.Response, ctx *goproxy.ProxyCtx) (*http.Response, error) {
	if resp == nil || resp.Request == nil || len(resp.Request.Header) == 0 ||
		resp.Request.Header.Get(types.MockRecordMode) == types.MockRecordModeDisabled {
		log.WithFields(log.Fields{}).Debugf("proxy server returning canned response")

		return resp, nil
	}

	log.WithFields(log.Fields{
		"Response": resp,
	}).Debugf("proxy server response received")

	var reqBytes []byte
	var err error
	switch resp.Request.Body.(type) {
	case utils.ResetReader:
		_ = resp.Request.Body.(utils.ResetReader).Reset()
	}

	reqBytes, resp.Request.Body, err = utils.ReadAll(resp.Request.Body)
	if err != nil {
		log.WithFields(log.Fields{
			"Path":   resp.Request.URL,
			"Method": resp.Request.Method,
			"Error":  err,
		}).Warnf("proxy server failed to read request body in handle-response")
		return resp, err
	}

	var resBytes []byte
	resBytes, resp.Body, err = utils.ReadAll(resp.Body)
	if err != nil {
		log.WithFields(log.Fields{
			"Path":   resp.Request.URL,
			"Method": resp.Request.Method,
			"Error":  err,
		}).Warnf("proxy server failed to read response body in handle-response")
		return resp, err
	}

	scenario, resContentType, err := saveMockResponse(
		h.config,
		resp.Request.URL,
		resp.Request,
		reqBytes,
		resBytes,
		resp.Header,
		resp.StatusCode,
		resp.Proto,
		getStartTime(ctx),
		time.Now(),
		h.scenarioRepository)
	if err != nil {
		return resp, err
	}
	resp.Body = utils.NopCloser(bytes.NewReader(resBytes))
	resp.Header["Access-Control-Allow-Origin"] = []string{h.config.CORS}
	resp.Header["Access-Control-Allow-Credentials"] = []string{"true"}
	resp.Header["Access-Control-Allow-Methods"] = []string{"GET, POST, DELETE, PUT, PATCH, OPTIONS, HEAD"}
	resp.Header["Access-Control-Allow-Headers"] = []string{"*"}
	resp.Header["Access-Control-Max-Age"] = []string{"1728000"}
	resp.Header["Access-Control-Expose-Headers"] = []string{"Content-Length,Content-Range"}
	agent := h.config.UserAgent
	if resp.Header.Get("Via") != "" {
		agent = agent + " (" + resp.Header.Get("Via") + ")"
	}
	resp.Header["Via"] = []string{agent}
	if resContentType != "" {
		resp.Header[types.ContentTypeHeader] = []string{resContentType}
	} else {
		resp.Header[types.ContentTypeHeader] = []string{"application/json"}
	}
	resp.ContentLength = int64(len(resBytes))
	resp.Header[types.ContentLengthHeader] = []string{fmt.Sprintf("%d", len(resBytes))}
	//resp.Header["Vary"] = []string{"Origin, Accept-Encoding""}
	//resp.Header["Access-Control-Allow-Headers"] = []string{"Content-Type, api_key, Authorization"}
	//resp.Header["Content-Security-Policy"] = []string{"default-src 'self', form-action 'self',script-src 'self'"}

	for k := range ignoredResponseHeaders {
		resp.Header.Del(k)
	}
	log.WithFields(log.Fields{
		"Path":     resp.Request.URL,
		"Method":   resp.Request.Method,
		"Length":   len(resBytes),
		"Scenario": scenario,
		//"ReqHeaders":  resp.Request.Header,
		//"RespHeaders": resp.Header,
	}).Infof("proxy server recorded response")
	resp.Request.Header = make(http.Header) // reset headers for next request in case we are using it.
	return resp, nil
}

func (h *Handler) proxyCondition() goproxy.ReqConditionFunc {
	return func(req *http.Request, _ *goproxy.ProxyCtx) bool {
		if h.config.ProxyURLFilter != "" {
			if matched, err := regexp.Match(h.config.ProxyURLFilter, []byte(req.URL.String())); err == nil && !matched {
				return false
			}
		}
		return !strings.Contains(req.URL.Path, "html") && !strings.Contains(req.URL.Path, "txt")
	}
}

func getStartTime(ctx *goproxy.ProxyCtx) time.Time {
	switch ctx.UserData.(type) {
	case time.Time:
		return ctx.UserData.(time.Time)
	}
	return time.Now()
}

func saveProxyCert() error {
	cert, err := tls.X509KeyPair(goproxy.CA_CERT, goproxy.CA_KEY)
	if err != nil {
		return err
	}
	ca, err := x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		return err
	}
	start := time.Unix(time.Now().Unix()-2592000, 0) // 2592000  = 30 day
	end := time.Unix(time.Now().Unix()+31536000, 0)  // 31536000 = 365 day
	serial := big.NewInt(rand.Int63())
	template := x509.Certificate{
		SerialNumber: serial,
		Issuer:       ca.Subject,
		Subject: pkix.Name{
			Organization: []string{"GoProxy untrusted MITM proxy Inc"},
		},
		NotBefore:             start,
		NotAfter:              end,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}
	caPrivKey, err := rsa.GenerateKey(crand.Reader, 4096)
	if err != nil {
		return err
	}
	derBytes, err := x509.CreateCertificate(crand.Reader, &template, &template, &caPrivKey.PublicKey, caPrivKey)
	if err != nil {
		return err
	}
	certPEM := new(bytes.Buffer)
	err = pem.Encode(certPEM, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	if err != nil {
		return err
	}
	return os.WriteFile("cert.pem", certPEM.Bytes(), 0644)
}
