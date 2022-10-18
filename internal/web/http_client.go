package web

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/bhatti/api-mock-service/internal/types"

	log "github.com/sirupsen/logrus"
)

// HTTPClient defines methods for http get and post methods
type HTTPClient interface {
	Handle(
		ctx context.Context,
		url string,
		method string,
		headers map[string][]string,
		params map[string]string,
		body io.ReadCloser,
	) (int, io.ReadCloser, map[string][]string, error)
}

// DefaultHTTPClient implements HTTPClient
type DefaultHTTPClient struct {
	config *types.Configuration
}

// NewHTTPClient creates structure for HTTPClient
func NewHTTPClient(config *types.Configuration) HTTPClient {
	return &DefaultHTTPClient{config: config}
}

// Handle makes HTTP request
func (w *DefaultHTTPClient) Handle(
	ctx context.Context,
	url string,
	method string,
	headers map[string][]string,
	params map[string]string,
	body io.ReadCloser,
) (statusCode int, respBody io.ReadCloser, respHeader map[string][]string, err error) {
	started := time.Now()
	log.WithFields(log.Fields{
		"Component": "DefaultHTTPClient",
		"Method":    method,
		"URL":       url,
	}).Info("handle BEGIN")

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return 500, nil, make(map[string][]string), err
	}
	statusCode, respBody, respHeader, err = w.execute(req, headers, params)

	elapsed := time.Since(started).String()
	log.WithFields(log.Fields{
		"Component":  "DefaultHTTPClient",
		"URL":        url,
		"Method":     method,
		"StatusCode": statusCode,
		"Elapsed":    elapsed,
		"Error":      err}).Info("handle END")
	return
}

// ////////////////////////////////// PRIVATE METHODS ///////////////////////////////////////////
func (w *DefaultHTTPClient) execute(
	req *http.Request,
	headers map[string][]string,
	params map[string]string) (int, io.ReadCloser, map[string][]string, error) {
	if req == nil {
		return 500, nil, make(map[string][]string), fmt.Errorf("request not specified")
	}
	if len(params) > 0 {
		paramVals := url.Values{}
		for k, v := range params {
			paramVals.Add(k, v)
		}
		req.URL.RawQuery = paramVals.Encode()
	}
	for name, vals := range headers {
		for _, val := range vals {
			req.Header.Add(name, val)
		}
	}
	if w.config.UserAgent != "" {
		req.Header.Set("User-Agent", w.config.UserAgent)
	}

	client := httpClient(w.config)
	resp, err := client.Do(req)
	if err != nil {
		return 500, nil, make(map[string][]string), err
	}
	return resp.StatusCode, resp.Body, resp.Header, nil
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
