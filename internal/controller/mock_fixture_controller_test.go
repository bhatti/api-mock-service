package controller

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/bhatti/api-mock-service/internal/repository"
	"github.com/bhatti/api-mock-service/internal/types"
	"github.com/bhatti/api-mock-service/internal/web"

	"github.com/stretchr/testify/require"
)

func Test_InitializeSwaggerStructsForMockFixtureController(t *testing.T) {
	_ = mockFixtureNamesResponseBody{}
	_ = mockFixtureNamesParams{}
	_ = mockFixtureCreateParams{}
	_ = mockFixtureResponseBody{}
	_ = mockFixtureIDParams{}
	_ = emptyResponse{}
}

func Test_ShouldFailPostFixtureWithoutNameOrPath(t *testing.T) {
	// GIVEN repository and controller for mock fixture
	mockFixtureRepository, err := repository.NewFileFixtureRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)
	webServer := web.NewStubWebServer()
	ctrl := NewMockFixtureController(mockFixtureRepository, webServer)
	data := []byte("test data")
	require.NoError(t, err)
	reader := io.NopCloser(bytes.NewReader(data))
	ctx := web.NewStubContext(&http.Request{Body: reader})

	// WHEN creating mock fixture without name and path
	err = ctrl.PostMockFixture(ctx)

	// THEN it should fail
	require.Error(t, err)
	ctx.Params["method"] = "POST"
	err = ctrl.PostMockFixture(ctx)
	require.Error(t, err)

	// AND it should fail again
	require.Error(t, err)
	ctx.Params["name"] = "data1"
	err = ctrl.PostMockFixture(ctx)
	require.Error(t, err)
}

func Test_ShouldFailGetFixtureWithoutNameOrPath(t *testing.T) {
	// GIVEN repository and controller for mock fixture
	mockFixtureRepository, err := repository.NewFileFixtureRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)
	webServer := web.NewStubWebServer()
	ctrl := NewMockFixtureController(mockFixtureRepository, webServer)
	data := []byte("test data")
	require.NoError(t, err)
	reader := io.NopCloser(bytes.NewReader(data))
	ctx := web.NewStubContext(&http.Request{Body: reader})

	// WHEN getting mock fixture without name and path
	err = ctrl.GetMockFixture(ctx)
	// THEN it should fail
	require.Error(t, err)

	// AND it should fail given method but without name
	ctx.Params["method"] = "GET"
	err = ctrl.GetMockFixture(ctx)
	require.Error(t, err)

	// AND it should fail to post given method but without name
	err = ctrl.PostMockFixture(ctx)
	require.Error(t, err)

	// AND it should fail again
	require.Error(t, err)
	ctx.Params["name"] = "data1"
	err = ctrl.GetMockFixture(ctx)
	require.Error(t, err)
}

func Test_ShouldFailGetFixtureNamesWithoutNameOrPath(t *testing.T) {
	// GIVEN repository and controller for mock fixture
	mockFixtureRepository, err := repository.NewFileFixtureRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)
	webServer := web.NewStubWebServer()
	ctrl := NewMockFixtureController(mockFixtureRepository, webServer)
	data := []byte("test data")
	require.NoError(t, err)
	reader := io.NopCloser(bytes.NewReader(data))
	ctx := web.NewStubContext(&http.Request{Body: reader})

	// WHEN getting mock fixture without name and path
	err = ctrl.getMockFixtureNames(ctx)
	// THEN it should fail
	require.Error(t, err)

	// AND it should fail given method but without name
	ctx.Params["method"] = "GET"
	err = ctrl.getMockFixtureNames(ctx)
	require.Error(t, err)

	// AND it should fail again with name
	require.Error(t, err)
	ctx.Params["name"] = "data1"
	err = ctrl.getMockFixtureNames(ctx)
	require.Error(t, err)
}

func Test_ShouldFailDeleteFixtureWithoutNameOrPath(t *testing.T) {
	// GIVEN repository and controller for mock fixture
	mockFixtureRepository, err := repository.NewFileFixtureRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)
	webServer := web.NewStubWebServer()
	ctrl := NewMockFixtureController(mockFixtureRepository, webServer)
	data := []byte("test data")
	require.NoError(t, err)
	reader := io.NopCloser(bytes.NewReader(data))
	ctx := web.NewStubContext(&http.Request{Body: reader})

	// WHEN deleting mock fixture without name and path
	err = ctrl.DeleteMockFixture(ctx)

	// THEN it should fail
	require.Error(t, err)
	ctx.Params["method"] = "DELETE"
	err = ctrl.DeleteMockFixture(ctx)
	require.Error(t, err)

	// AND it should fail again
	require.Error(t, err)
	ctx.Params["name"] = "data1"
	err = ctrl.DeleteMockFixture(ctx)
	require.Error(t, err)
}

func Test_ShouldCreateAndGetMockFixture(t *testing.T) {
	// GIVEN repository and controller for mock fixture
	mockFixtureRepository, err := repository.NewFileFixtureRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)
	webServer := web.NewStubWebServer()
	ctrl := NewMockFixtureController(mockFixtureRepository, webServer)
	data := []byte("test data")
	require.NoError(t, err)
	reader := io.NopCloser(bytes.NewReader(data))
	ctx := web.NewStubContext(&http.Request{Body: reader})

	// WHEN creating mock fixture
	ctx.Params["method"] = "POST"
	ctx.Params["name"] = "data1"
	ctx.Params["path"] = "/ghi/klm"
	err = ctrl.PostMockFixture(ctx)

	// THEN it should return saved fixture
	require.NoError(t, err)

	// WHEN getting mock scenario by path
	err = ctrl.GetMockFixture(ctx)

	// THEN it should not fail
	require.NoError(t, err)
}

func Test_ShouldCreateAndGetMockFixtureNames(t *testing.T) {
	// GIVEN repository and controller for mock fixtures
	mockFixtureRepository, err := repository.NewFileFixtureRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)
	webServer := web.NewStubWebServer()
	ctrl := NewMockFixtureController(mockFixtureRepository, webServer)
	for i := 0; i < 10; i++ {
		data := []byte(fmt.Sprintf("abc_%d", i))
		reader := io.NopCloser(bytes.NewReader(data))
		ctx := web.NewStubContext(&http.Request{Body: reader})
		ctx.Params["method"] = "GET"
		ctx.Params["name"] = fmt.Sprintf("data_%d", i)
		ctx.Params["path"] = "/qfc/klm"

		// WHEN creating mock fixture
		err = ctrl.PostMockFixture(ctx)
		// THEN it should return saved scenario
		require.NoError(t, err)
	}

	// WHEN getting mock fixture by path
	ctx := web.NewStubContext(&http.Request{})
	ctx.Params["method"] = "GET"
	ctx.Params["path"] = "/qfc/klm"
	err = ctrl.getMockFixtureNames(ctx)

	// THEN it should not fail
	require.NoError(t, err)
	names := ctx.Result.([]string)
	require.Equal(t, 10, len(names))
}

func Test_ShouldCreateAndDeleteMockFixture(t *testing.T) {
	// GIVEN repository and controller for mock scenario
	mockFixtureRepository, err := repository.NewFileFixtureRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)
	webServer := web.NewStubWebServer()
	ctrl := NewMockFixtureController(mockFixtureRepository, webServer)

	// WHEN creating mock scenario
	data := []byte("test data")
	reader := io.NopCloser(bytes.NewReader(data))
	ctx := web.NewStubContext(&http.Request{Body: reader})
	ctx.Params["method"] = "DELETE"
	ctx.Params["name"] = "data1"
	ctx.Params["path"] = "/ghi/klm"
	err = ctrl.PostMockFixture(ctx)

	// THEN it should succeed
	require.NoError(t, err)

	// WHEN deleting mock scenario
	err = ctrl.DeleteMockFixture(ctx)
	// THEN it should succeed
	require.NoError(t, err)

	// AND get API should fail
	err = ctrl.GetMockFixture(ctx)

	// THEN it should not fail
	require.Error(t, err)
}
