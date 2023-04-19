package controller

import (
	"bytes"
	"encoding/json"
	"github.com/bhatti/api-mock-service/internal/repository"
	"github.com/bhatti/api-mock-service/internal/types"
	"github.com/bhatti/api-mock-service/internal/web"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"testing"
	"time"
)

func Test_InitializeSwaggerStructsForGroupConfigController(t *testing.T) {
	_ = getGroupsConfigParams{}
	_ = putGroupsConfigParams{}
	_ = groupConfigResponseBody{}
	_ = putGroupConfigResponseBody{}
}

func Test_ShouldFailGetGroupConfigWithoutGroup(t *testing.T) {
	config := types.BuildTestConfig()
	// GIVEN repository and controller for mock scenario
	repo, err := repository.NewFileGroupConfigRepository(config)
	require.NoError(t, err)
	webServer := web.NewStubWebServer()
	ctrl := NewGroupConfigController(repo, webServer)
	data := []byte("test data")
	require.NoError(t, err)
	reader := io.NopCloser(bytes.NewReader(data))
	ctx := web.NewStubContext(&http.Request{Body: reader})

	// WHEN fetching group config without group
	err = ctrl.getGroupConfig(ctx)
	// THEN it should fail
	require.Error(t, err)
}

func Test_ShouldFailGetGroupConfigWithoutConfig(t *testing.T) {
	config := types.BuildTestConfig()
	// GIVEN repository and controller for mock scenario
	repo, err := repository.NewFileGroupConfigRepository(config)
	require.NoError(t, err)
	webServer := web.NewStubWebServer()
	ctrl := NewGroupConfigController(repo, webServer)
	data := []byte("test data")
	require.NoError(t, err)
	reader := io.NopCloser(bytes.NewReader(data))
	ctx := web.NewStubContext(&http.Request{Body: reader})
	ctx.Params["group"] = "test5"

	// WHEN fetching group config without group
	err = ctrl.getGroupConfig(ctx)
	// THEN it should fail
	require.Error(t, err)
}

func Test_ShouldGetGroupConfigWithGroup(t *testing.T) {
	config := types.BuildTestConfig()
	// GIVEN repository and controller for mock scenario
	repo, err := repository.NewFileGroupConfigRepository(config)
	require.NoError(t, err)
	err = repo.Save("test1", &types.GroupConfig{})
	require.NoError(t, err)
	webServer := web.NewStubWebServer()
	ctrl := NewGroupConfigController(repo, webServer)
	data := []byte("test data")
	require.NoError(t, err)
	reader := io.NopCloser(bytes.NewReader(data))
	ctx := web.NewStubContext(&http.Request{Body: reader})
	ctx.Params["group"] = "test1"

	// WHEN fetching group config without group
	err = ctrl.getGroupConfig(ctx)
	// THEN it should not fail
	require.NoError(t, err)
}

func Test_ShouldFailPutGroupConfigWithoutGroup(t *testing.T) {
	config := types.BuildTestConfig()
	// GIVEN repository and controller for mock scenario
	repo, err := repository.NewFileGroupConfigRepository(config)
	require.NoError(t, err)
	webServer := web.NewStubWebServer()
	ctrl := NewGroupConfigController(repo, webServer)
	data := []byte("test data")
	require.NoError(t, err)
	reader := io.NopCloser(bytes.NewReader(data))
	ctx := web.NewStubContext(&http.Request{Body: reader})

	// WHEN saving group config without group
	err = ctrl.putGroupConfig(ctx)
	// THEN it should fail
	require.Error(t, err)
}

func Test_ShouldPutGroupConfig(t *testing.T) {
	config := types.BuildTestConfig()
	// GIVEN repository and controller for mock scenario
	repo, err := repository.NewFileGroupConfigRepository(config)
	require.NoError(t, err)
	webServer := web.NewStubWebServer()
	ctrl := NewGroupConfigController(repo, webServer)
	gc := &types.GroupConfig{
		Variables:                        map[string]string{"v1": "val"},
		MeanTimeBetweenAdditionalLatency: 8,
		MeanTimeBetweenFailure:           4,
		MaxAdditionalLatency:             time.Second * 3,
	}
	data, err := json.Marshal(gc)
	require.NoError(t, err)
	require.NoError(t, err)
	reader := io.NopCloser(bytes.NewReader(data))
	ctx := web.NewStubContext(&http.Request{Body: reader})
	ctx.Params["group"] = "test2"

	// WHEN saving group config with group
	err = ctrl.putGroupConfig(ctx)
	// THEN it should succeed
	require.NoError(t, err)

	// WHEN fetching group config without group
	err = ctrl.getGroupConfig(ctx)
	// THEN it should not fail
	require.NoError(t, err)
}
