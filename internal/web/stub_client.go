package web

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
)

// StubHTTPResponse defines stub response
type StubHTTPResponse struct {
	Filename      string
	Bytes         []byte
	Status        int
	Error         error
	sleepDuration time.Duration
}

// NewStubHTTPResponseError creates stubbed response with error
func NewStubHTTPResponseError(status int, sleep time.Duration, err error) *StubHTTPResponse {
	return &StubHTTPResponse{Status: status, sleepDuration: sleep, Error: err}
}

// NewStubHTTPResponse creates stubbed response
func NewStubHTTPResponse(status int, unk any) *StubHTTPResponse {
	if unk == nil {
		return &StubHTTPResponse{Status: status}
	}
	switch unk.(type) {
	case string:
		if _, err := os.Stat(unk.(string)); err == nil {
			return &StubHTTPResponse{Status: status, Filename: unk.(string)}
		}
		return &StubHTTPResponse{Status: status, Bytes: []byte(unk.(string))}
	default:
		b, err := json.Marshal(unk)
		if err != nil {
			panic(fmt.Errorf("failed to serialize %v due to error %w", unk, err))
		}
		if _, err := os.Stat(string(b)); err == nil {
			return &StubHTTPResponse{Status: status, Filename: string(b)}
		}
		return &StubHTTPResponse{Status: status, Bytes: b}
	}
}

// StubHTTPClient implements HTTPClient for stubbed response
type StubHTTPClient struct {
	mappingByMethodURL map[string]*StubHTTPResponse
}

// NewStubHTTPClient - creates structure for HTTPClient
func NewStubHTTPClient() *StubHTTPClient {
	return &StubHTTPClient{
		mappingByMethodURL: make(map[string]*StubHTTPResponse),
	}
}

// AddMapping adds mapping for stub response
func (w *StubHTTPClient) AddMapping(method string, url string, resp *StubHTTPResponse) {
	w.mappingByMethodURL[method+url] = resp
}

// getMapping finds mapping for stub response
func (w *StubHTTPClient) getMapping(method string, url string) *StubHTTPResponse {
	return w.mappingByMethodURL[method+url]
}

// Handle makes HTTP request
func (w *StubHTTPClient) Handle(
	_ context.Context,
	url string,
	method string,
	_ map[string][]string,
	_ map[string]string,
	_ io.ReadCloser) (int, io.ReadCloser, map[string][]string, error) {
	if url == "" {
		return 500, nil, nil, fmt.Errorf("url is not specified")
	}
	if method == "" {
		return 500, nil, nil, fmt.Errorf("method is not specified")
	}
	log.WithFields(log.Fields{"component": "stub-web", "url": url, "method": method}).Debugf("BEGIN")
	resp := w.getMapping(method, url)
	if resp == nil {
		return 404, nil, nil, fmt.Errorf("couldn't find URL '%s' method '%s' in mapping: %v", url, method, w.mappingByMethodURL)
	}
	if resp.sleepDuration > 0 {
		time.Sleep(resp.sleepDuration)
	}
	if len(resp.Bytes) > 0 {
		return resp.Status, io.NopCloser(bytes.NewReader(resp.Bytes)), nil, resp.Error
	}
	if resp.Error != nil {
		return resp.Status, nil, nil, resp.Error
	}
	b, err := os.ReadFile(resp.Filename)
	if err != nil {
		return 404, nil, nil, fmt.Errorf("error reading file %v for url %v due to %w", resp.Filename, url, err)
	}
	return resp.Status, io.NopCloser(bytes.NewReader(b)), nil, resp.Error
}
