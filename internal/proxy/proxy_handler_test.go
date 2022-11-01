package proxy

import (
	"bytes"
	"github.com/bhatti/api-mock-service/internal/repository"
	"github.com/bhatti/api-mock-service/internal/types"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/url"
	"testing"
)

func Test_ShouldNotStartProxyServer(t *testing.T) {
	// GIVEN a mock scenario repository
	scenarioRepository, err := repository.NewFileMockScenarioRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)
	fixtureRepository, err := repository.NewFileFixtureRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)

	handler := NewProxyHandler(-1, scenarioRepository, fixtureRepository)
	require.Error(t, handler.Start())
}

func Test_ShouldNotHandleProxyRequest(t *testing.T) {
	// GIVEN a mock scenario repository
	scenarioRepository, err := repository.NewFileMockScenarioRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)
	fixtureRepository, err := repository.NewFileFixtureRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)

	u, err := url.Parse("http://localhost:8080/path?a=b")
	require.NoError(t, err)
	req := &http.Request{
		URL:    u,
		Method: "POST",
		Header: http.Header{"X1": []string{"val1"}, types.ContentTypeHeader: []string{"json"}},
	}
	handler := NewProxyHandler(8081, scenarioRepository, fixtureRepository)
	_, res := handler.handleRequest(req, nil)
	require.Nil(t, res)
}

func Test_ShouldHandleProxyRequest(t *testing.T) {
	// GIVEN a mock scenario repository
	scenarioRepository, err := repository.NewFileMockScenarioRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)
	fixtureRepository, err := repository.NewFileFixtureRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)

	scenario := buildScenario(types.Post, "todos", "/api/todos", 0)
	require.NoError(t, scenarioRepository.Save(scenario))

	u, err := url.Parse("http://localhost:8080/api/todos?a=3&b=abc")
	require.NoError(t, err)
	req := &http.Request{
		URL:    u,
		Method: "POST",
		Header: http.Header{"X1": []string{"val1"}, types.ContentTypeHeader: []string{"json"}},
	}
	handler := NewProxyHandler(8081, scenarioRepository, fixtureRepository)
	_, res := handler.handleRequest(req, nil)
	require.NotNil(t, res)
}

func Test_ShouldHandleProxyResponseWithoutRequestBody(t *testing.T) {
	// GIVEN a mock scenario repository
	scenarioRepository, err := repository.NewFileMockScenarioRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)
	fixtureRepository, err := repository.NewFileFixtureRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)

	u, err := url.Parse("http://localhost:8080/api/todos?a=b")
	require.NoError(t, err)
	req := &http.Request{
		URL:    u,
		Method: "POST",
		Header: http.Header{"X1": []string{"val1"}, types.ContentTypeHeader: []string{"json"}},
	}
	res := &http.Response{
		Request: req,
	}
	handler := NewProxyHandler(8081, scenarioRepository, fixtureRepository)
	res = handler.handleResponse(res, nil)
	require.NotNil(t, res)
}

func Test_ShouldHandleProxyResponseWithoutRequest(t *testing.T) {
	// GIVEN a mock scenario repository
	scenarioRepository, err := repository.NewFileMockScenarioRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)
	fixtureRepository, err := repository.NewFileFixtureRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)

	res := &http.Response{}
	handler := NewProxyHandler(8081, scenarioRepository, fixtureRepository)
	res = handler.handleResponse(res, nil)
	require.NotNil(t, res)
}

func Test_ShouldHandleProxyResponseWithoutResponseBody(t *testing.T) {
	// GIVEN a mock scenario repository
	scenarioRepository, err := repository.NewFileMockScenarioRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)
	fixtureRepository, err := repository.NewFileFixtureRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)

	u, err := url.Parse("http://localhost:8080/api/todos?a=b")
	require.NoError(t, err)
	req := &http.Request{
		URL:    u,
		Method: "POST",
		Header: http.Header{"X1": []string{"val1"}, types.ContentTypeHeader: []string{"json"}},
		Body:   io.NopCloser(bytes.NewReader([]byte("test"))),
	}
	res := &http.Response{
		Request: req,
	}
	handler := NewProxyHandler(8081, scenarioRepository, fixtureRepository)
	res = handler.handleResponse(res, nil)
	require.NotNil(t, res)
}

func Test_ShouldHandleProxyResponseWithRequestAndResponseBody(t *testing.T) {
	// GIVEN a mock scenario repository
	scenarioRepository, err := repository.NewFileMockScenarioRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)
	fixtureRepository, err := repository.NewFileFixtureRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)

	u, err := url.Parse("http://localhost:8080/api/todos?a=b")
	require.NoError(t, err)
	req := &http.Request{
		URL:    u,
		Method: "POST",
		Header: http.Header{"X1": []string{"val1"}, types.ContentTypeHeader: []string{"json"}},
		Body:   io.NopCloser(bytes.NewReader([]byte("test"))),
	}
	res := &http.Response{
		Request: req,
		Body:    io.NopCloser(bytes.NewReader([]byte("test"))),
		Header:  http.Header{"X1": []string{"val1"}, types.ContentTypeHeader: []string{"json"}},
	}
	handler := NewProxyHandler(8081, scenarioRepository, fixtureRepository)
	res = handler.handleResponse(res, nil)
	require.NotNil(t, res)
	req.Header[types.MockRecordMode] = []string{types.MockRecordModeDisabled}
	res = handler.handleResponse(res, nil)
	require.NotNil(t, res)
}
