package types

import (
	"net/http"
	"strconv"
	"time"
)

func buildScenario() *APIScenario {
	scenario := &APIScenario{
		Method:         Post,
		Name:           "scenario",
		Path:           "/path1/\\\\//test1//abc////",
		Description:    "",
		Group:          "test-group",
		Tags:           []string{"tag1", "tag2"},
		Authentication: make(map[string]APIAuthorization),
		Request: APIRequest{
			Headers: make(map[string]string),
			AssertQueryParamsPattern: map[string]string{
				"a": "1",
				"b": "2",
			},
			AssertHeadersPattern: map[string]string{
				"CTag": "981",
			},
			Variables: make(map[string]string),
		},
		Response: APIResponse{
			Headers: map[string][]string{
				ETagHeader:        {"123"},
				ContentTypeHeader: {"application/json"},
			},
			Contents:   "test body",
			StatusCode: 200,
		},
		WaitBeforeReply: time.Duration(1) * time.Second,
	}
	return scenario
}

// BuildTestScenario helper method
func BuildTestScenario(method MethodType, name string, path string, n int) *APIScenario {
	return &APIScenario{
		Method:      method,
		Name:        name,
		Path:        path,
		Group:       path,
		Description: name,
		Request: APIRequest{
			HTTPVersion:              "1.1",
			QueryParams:              make(map[string]string),
			PostParams:               make(map[string]string),
			Headers:                  make(map[string]string),
			AssertQueryParamsPattern: map[string]string{"a": `\d+`, "b": "abc"},
			AssertHeadersPattern: map[string]string{
				ContentTypeHeader: "application/json",
				ETagHeader:        `\d{3}`,
			},
		},
		Response: APIResponse{
			HTTPVersion: "1.1",
			Headers: http.Header{
				ETagHeader:        {strconv.Itoa(n)},
				ContentTypeHeader: {"application/json"},
			},
			Contents:   "test body",
			StatusCode: 200,
		},
		WaitBeforeReply: time.Duration(1) * time.Second,
	}
}
