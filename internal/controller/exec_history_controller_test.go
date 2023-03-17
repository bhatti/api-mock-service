package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/bhatti/api-mock-service/internal/types/archive"
	"io"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/bhatti/api-mock-service/internal/repository"
	"github.com/bhatti/api-mock-service/internal/types"
	"github.com/bhatti/api-mock-service/internal/web"

	"github.com/stretchr/testify/require"
)

func Test_InitializeSwaggerStructsForExecHistoryoController(t *testing.T) {
	_ = execHistoryHarResponse{}
	_ = execHistoryNamesResponseBody{}
	_ = execHistoryNamesParams{}
}

func Test_ShouldGetExecutionHistoryNames(t *testing.T) {
	config := types.BuildTestConfig()
	// GIVEN repository and controller for mock scenario
	mockScenarioRepository, err := repository.NewFileAPIScenarioRepository(config)
	require.NoError(t, err)
	webServer := web.NewStubWebServer()
	ctrl := NewExecHistoryController(mockScenarioRepository, webServer)
	data := []byte("test data")
	require.NoError(t, err)
	reader := io.NopCloser(bytes.NewReader(data))
	u, err := url.Parse("http://localhost:8080?a=1&b=abc")
	require.NoError(t, err)
	ctx := web.NewStubContext(&http.Request{Body: reader, URL: u})
	ctx.Request().Header = http.Header{"Auth": []string{"0123456789"}}
	scenario := buildScenario(types.Post, "test1", "/path1", 1)
	err = mockScenarioRepository.SaveHistory(scenario, u, time.Now(), time.Now())
	require.NoError(t, err)

	// WHEN getting mock scenario groups
	err = ctrl.getExecHistoryNames(ctx)
	// THEN it should not fail
	require.NoError(t, err)
	names := ctx.Result.([]string)
	require.True(t, len(names) > 0)
}

func Test_ShouldGetExecutionHistoryHar(t *testing.T) {
	config := types.BuildTestConfig()
	// GIVEN repository and controller for mock scenario
	mockScenarioRepository, err := repository.NewFileAPIScenarioRepository(config)
	require.NoError(t, err)
	webServer := web.NewStubWebServer()
	ctrl := NewExecHistoryController(mockScenarioRepository, webServer)
	require.NoError(t, err)
	u, err := url.Parse("https://localhost:8080")
	require.NoError(t, err)
	for i := 0; i < 120; i++ {
		scenario := buildScenario(types.Post, "test1", fmt.Sprintf("/users/path/%d", i), i)
		if i%2 == 0 {
			scenario.Group = "exec-1"
		} else {
			scenario.Group = "exec-2"
		}
		err = mockScenarioRepository.SaveHistory(scenario, u, time.Now(), time.Now())
		require.NoError(t, err)
	}

	for i := 0; i < 4; i++ {
		reader := io.NopCloser(bytes.NewReader([]byte("test data")))
		u, err := url.Parse(fmt.Sprintf("http://localhost:8080?a=1&b=abc&page=%d&group=%s", i, "exec-1"))
		require.NoError(t, err)
		ctx := web.NewStubContext(&http.Request{Body: reader, URL: u})
		ctx.Request().Header = http.Header{"Auth": []string{"0123456789"}}
		// WHEN getting mock scenario groups
		err = ctrl.getExecHistoryHar(ctx)
		// THEN it should not fail
		require.NoError(t, err)
		res := ctx.Result.([]archive.Har)
		require.Equal(t, 2, len(res), fmt.Sprintf("i=%d", i))
		require.Equal(t, 10, len(res[0].Log.Entries), fmt.Sprintf("i=%d", i))
		require.Equal(t, 10, len(res[1].Log.Entries), fmt.Sprintf("i=%d", i))
	}
}

func Test_ShouldSaveHarContents(t *testing.T) {
	config := types.BuildTestConfig()
	// GIVEN repository and controller for mock scenario
	mockScenarioRepository, err := repository.NewFileAPIScenarioRepository(config)
	require.NoError(t, err)
	webServer := web.NewStubWebServer()
	ctrl := NewExecHistoryController(mockScenarioRepository, webServer)
	scenario := buildScenario(types.Post, "test1", "/path1", 1)
	//u, err := url.Parse("http://localhost:8080" + scenario.Path + "?a=1&b=abc")
	u, err := url.Parse("http://localhost:8080?a=1&b=abc")
	require.NoError(t, err)
	har := archive.BuildHar(config, scenario, u, time.Now(), time.Now().Add(time.Second))
	b, err := json.Marshal(har)
	require.NoError(t, err)
	reader := io.NopCloser(bytes.NewReader(b))
	ctx := web.NewStubContext(&http.Request{Body: reader, Method: string(scenario.Method), URL: u})
	ctx.Request().Header = http.Header{"Auth": []string{"0123456789"}, "Content-Type": []string{"application/json"}}

	// WHEN creating mock scenario
	err = ctrl.postExecHistoryHar(ctx)

	// THEN it should return saved scenario
	require.NoError(t, err)
}
