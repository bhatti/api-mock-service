package web

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/bhatti/api-mock-service/internal/types"
	"github.com/bhatti/api-mock-service/internal/utils"
	log "github.com/sirupsen/logrus"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// Authorization constant
const Authorization = "Authorization"

// AWSAccessKey constant
const AWSAccessKey = "AWS_ACCESS_KEY_ID"

// AWSSecretKey constant
const AWSSecretKey = "AWS_SECRET_ACCESS_KEY"

// AWSSecurityToken constant
const AWSSecurityToken = "AWS_SECURITY_TOKEN"

// AWSSessionToken constant
const AWSSessionToken = "AWS_SESSION_TOKEN"

var internalParamKeys = []string{AWSAccessKey, AWSSecretKey, AWSSecurityToken, AWSSessionToken}

// HTTPClient defines methods for http get and post methods
type HTTPClient interface {
	Handle(
		ctx context.Context,
		url string,
		method string,
		headers map[string][]string,
		params map[string]string,
		body io.ReadCloser,
	) (int, string, io.ReadCloser, map[string][]string, error)
}

// DefaultHTTPClient implements HTTPClient
type DefaultHTTPClient struct {
	config    *types.Configuration
	awsSigner AWSSigner
}

// NewHTTPClient creates structure for HTTPClient
func NewHTTPClient(config *types.Configuration, awsSigner AWSSigner) *DefaultHTTPClient {
	return &DefaultHTTPClient{
		config:    config,
		awsSigner: awsSigner,
	}
}

// Handle makes HTTP request
func (w *DefaultHTTPClient) Handle(
	ctx context.Context,
	url string,
	method string,
	headers map[string][]string,
	params map[string]string,
	body io.ReadCloser,
) (statusCode int, httpVersion string, respBody io.ReadCloser, respHeader map[string][]string, err error) {
	started := time.Now()
	log.WithFields(log.Fields{
		"Component": "DefaultHTTPClient",
		"Method":    method,
		"URL":       url,
	}).Debug("handle BEGIN")
	var bodyB []byte
	bodyB, body, err = utils.ReadAll(body)

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return 500, "", nil, make(map[string][]string), err
	}
	req.ContentLength = int64(len(bodyB))
	statusCode, httpVersion, respBody, respHeader, err = w.execute(req, headers, params)

	elapsed := time.Since(started).String()
	log.WithFields(log.Fields{
		"Component":  "DefaultHTTPClient",
		"URL":        url,
		"Method":     method,
		"StatusCode": statusCode,
		"Elapsed":    elapsed,
		"Error":      err}).Debug("handle END")
	return
}

// ////////////////////////////////// PRIVATE METHODS ///////////////////////////////////////////
func (w *DefaultHTTPClient) execute(
	req *http.Request,
	headers map[string][]string,
	params map[string]string) (int, string, io.ReadCloser, map[string][]string, error) {
	if req == nil {
		return 500, "", nil, make(map[string][]string), fmt.Errorf("request not specified")
	}
	internalKeyMap := make(map[string]string)
	if w.config.UserAgent != "" {
		req.Header.Set("User-Agent", w.config.UserAgent)
	}
	if len(params) > 0 {
		paramVals := url.Values{}
		for k, v := range params {
			if isInternalParamKeys(k) {
				internalKeyMap[strings.ToUpper(k)] = v
			} else {
				paramVals.Add(k, v)
			}
		}
		req.URL.RawQuery = paramVals.Encode()
	}
	for name, vals := range headers {
		for _, val := range vals {
			if isInternalParamKeys(name) {
				internalKeyMap[strings.ToUpper(name)] = val
			} else {
				req.Header.Add(name, val)
			}
		}
	}

	staticCredentials := credentials.NewStaticCredentials(
		GetHeaderParamOrEnvValue(internalKeyMap, AWSAccessKey),
		GetHeaderParamOrEnvValue(internalKeyMap, AWSSecretKey),
		GetHeaderParamOrEnvValue(internalKeyMap, AWSSecurityToken, AWSSessionToken),
	)
	awsAuthSig4, awsInfo, awsErr := w.awsSigner.AWSSign(req, staticCredentials)
	headers = req.Header
	client := httpClient(w.config)
	resp, err := client.Do(req)

	if err != nil {
		log.WithFields(log.Fields{
			"Component":   "DefaultHTTPClient",
			"URL":         req.URL,
			"Method":      req.Method,
			"Headers":     req.Header,
			"Params":      params,
			"AWSAuthSig4": awsAuthSig4,
			"AWSInfo":     awsInfo,
			"AWSError":    awsErr,
			"Error":       err,
		}).Warnf("failed to invoke http client")
	} else {
		log.WithFields(log.Fields{
			"Component":   "DefaultHTTPClient",
			"URL":         req.URL,
			"Method":      req.Method,
			"StatusCode":  resp.StatusCode,
			"Status":      resp.Status,
			"Headers":     req.Header,
			"Params":      params,
			"AWSAuthSig4": awsAuthSig4,
			"AWSInfo":     awsInfo,
			"AWSError":    awsErr,
			"Error":       err,
		}).Infof("invoked http client")
	}

	if err != nil {
		return 500, "", nil, make(map[string][]string), err
	}
	return resp.StatusCode, resp.Proto, resp.Body, resp.Header, nil
}

// GetHeaderParamOrEnvValue searches key in map or env variables
func GetHeaderParamOrEnvValue(params map[string]string, names ...string) string {
	for _, name := range names {
		if len(params[name]) > 0 {
			return params[name]
		}
		if len(os.Getenv(name)) > 0 {
			return os.Getenv(name)
		}
	}
	return ""
}

func getLocalIPAddresses() []string {
	ips := make([]string, 0)
	interfaces, err := net.Interfaces()
	if err != nil {
		return ips
	}
	// handle err
	for _, i := range interfaces {
		addresses, err := i.Addrs()
		if err != nil {
			return ips
		}
		for _, addr := range addresses {
			switch v := addr.(type) {
			case *net.IPNet:
				ips = append(ips, v.IP.String())
			case *net.IPAddr:
				ips = append(ips, v.IP.String())
			}
		}
	}
	return ips
}

func getRemoteIPAddressFromURL(targetURL string) string {
	hostIP := ""
	u, err := url.Parse(targetURL)
	if err == nil {
		addr, err := net.LookupIP(u.Host)
		if err == nil {
			hostIP = ""
			for i, a := range addr {
				if i > 0 {
					hostIP = hostIP + " "
				}
				hostIP = hostIP + a.String()
			}
		}
	}
	return hostIP
}

func getProxyEnv() map[string]string {
	proxies := make(map[string]string)
	proxies["HTTP_PROXY"] = os.Getenv("HTTP_PROXY")
	proxies["HTTPS_PROXY"] = os.Getenv("HTTPS_PROXY")
	proxies["NO_PROXY"] = os.Getenv("NO_PROXY")
	return proxies
}

func httpClient(config *types.Configuration) *http.Client {
	if config.ProxyURL == "" {
		return &http.Client{}
	}
	proxyURL, err := url.Parse(config.ProxyURL)
	if err != nil {
		log.WithFields(log.Fields{
			"Component": "DefaultHTTPClient",
			"IP":        getRemoteIPAddressFromURL(config.ProxyURL),
			"Error":     err,
			"Proxy":     config.ProxyURL}).Warn("Failed to parse proxy header")
		return &http.Client{}
	}

	headers := make(http.Header, 0)
	headers.Set("User-Agent", config.UserAgent)

	//adding the proxy settings to the Transport object
	transport := &http.Transport{
		Proxy:              http.ProxyURL(proxyURL),
		ProxyConnectHeader: headers,
	}

	log.WithFields(log.Fields{
		"Component": "DefaultHTTPClient",
		"LocalIP":   getLocalIPAddresses(),
		"EnvProxy":  getProxyEnv(),
		"Proxy":     proxyURL}).Info("Http client using proxy")
	return &http.Client{
		Transport: transport,
	}
}

func isInternalParamKeys(k string) bool {
	for _, next := range internalParamKeys {
		if strings.EqualFold(next, k) {
			return true
		}
	}
	return false
}
