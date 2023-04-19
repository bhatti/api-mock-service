package contract

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/bhatti/api-mock-service/internal/fuzz"
	"github.com/bhatti/api-mock-service/internal/oapi"
	"github.com/bhatti/api-mock-service/internal/repository"
	"github.com/bhatti/api-mock-service/internal/types"
	"github.com/bhatti/api-mock-service/internal/web"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	"io"
	"net/http"
	"net/url"
	"os"
	"testing"
	"time"
)

func Test_ShouldLookupPutMockScenarios(t *testing.T) {
	config := types.BuildTestConfig()
	// GIVEN a mock scenario repository
	scenarioRepository, err := repository.NewFileAPIScenarioRepository(config)
	require.NoError(t, err)
	fixtureRepository, err := repository.NewFileFixtureRepository(config)
	require.NoError(t, err)
	groupConfigRepository, err := repository.NewFileGroupConfigRepository(config)
	require.NoError(t, err)
	player := NewConsumerExecutor(scenarioRepository, fixtureRepository, groupConfigRepository)
	// AND a set of mock scenarios
	for i := 0; i < 3; i++ {
		scenario := types.BuildTestScenario(types.Put, fmt.Sprintf("todo_put_%d", i), "/api/todos/:id", i)
		require.NoError(t, scenarioRepository.Save(scenario))
	}
	u, err := url.Parse("https://jsonplaceholder.typicode.com/blah")
	// WHEN looking up non-existing API
	ctx := web.NewStubContext(
		&http.Request{
			Method: "PUT",
			URL:    u,
			Header: http.Header{
				types.MockWaitBeforeReply: []string{"1"},
				types.MockResponseStatus:  []string{"0"},
				types.ContentTypeHeader:   []string{"application/json"},
			},
		},
	)
	// THEN it should not find it
	err = player.Execute(ctx)
	require.Error(t, err)

	u, err = url.Parse("https://jsonplaceholder.typicode.com/api/todos/202?a=123&b=abc")
	require.NoError(t, err)
	// WHEN looking up todos by PUT with different query param
	ctx = web.NewStubContext(&http.Request{
		Method: "PUT",
		URL:    u,
		Header: http.Header{
			types.MockWaitBeforeReply: []string{"1"},
			types.MockResponseStatus:  []string{"0"},
			types.ContentTypeHeader:   []string{"application/json"},
			"ETag":                    []string{"12"},
		},
	})
	err = player.Execute(ctx)
	// THEN it should not find it due to missing ETag regex \d{3}
	require.Error(t, err)

	// WHEN looking up todos by PUT with different query param
	ctx = web.NewStubContext(&http.Request{
		Method: "PUT",
		URL:    u,
		Header: http.Header{
			types.MockWaitBeforeReply: []string{"1"},
			types.MockResponseStatus:  []string{"0"},
			types.ContentTypeHeader:   []string{"application/json"},
			"ETag":                    []string{"123"},
		},
	})
	err = player.Execute(ctx)
	require.NoError(t, err)
	// THEN it should find it
	saved := ctx.Result.([]byte)
	require.Equal(t, "test body", string(saved))
}

func Test_ShouldExecuteDescribeAPI(t *testing.T) {
	config := types.BuildTestConfig()
	// GIVEN a mock scenario repository
	scenarioRepository, err := repository.NewFileAPIScenarioRepository(config)
	require.NoError(t, err)
	fixtureRepository, err := repository.NewFileFixtureRepository(config)
	require.NoError(t, err)
	groupConfigRepository, err := repository.NewFileGroupConfigRepository(config)
	require.NoError(t, err)
	data, err := os.ReadFile("../../fixtures/oapi/describe-job.json")
	require.NoError(t, err)
	dataTempl := fuzz.NewDataTemplateRequest(false, 1, 1)
	specs, err := oapi.Parse(context.Background(), data, dataTempl)
	require.NoError(t, err)
	require.Equal(t, 1, len(specs))
	// AND executor
	player := NewConsumerExecutor(scenarioRepository, fixtureRepository, groupConfigRepository)
	for _, spec := range specs {
		scenario, err := spec.BuildMockScenario(dataTempl)
		require.NoError(t, err)
		require.True(t, scenario.Request.Headers["x-api-key"] != "")
		_, err = yaml.Marshal(scenario)
		require.NoError(t, err)
		require.NoError(t, scenarioRepository.Save(scenario))

		// WHEN executing with GET API
		u, err := url.Parse("https://localhost:8080/v1/describe/123")
		require.NoError(t, err)
		ctx := web.NewStubContext(&http.Request{
			Method: "GET",
			URL:    u,
			Header: http.Header{
				types.ContentTypeHeader:   []string{"application/json"},
				types.AuthorizationHeader: []string{"123456789"},
			},
		})
		err = player.Execute(ctx)
		require.NoError(t, err)
		specAgain := oapi.ScenarioToOpenAPI(spec.Title, "", scenario)
		j, err := specAgain.MarshalJSON()
		require.NoError(t, err)
		require.True(t, len(j) > 0)
	}
}

func Test_ShouldLookupPostMockScenarios(t *testing.T) {
	config := types.BuildTestConfig()
	// GIVEN a mock scenario repository
	scenarioRepository, err := repository.NewFileAPIScenarioRepository(config)
	require.NoError(t, err)
	fixtureRepository, err := repository.NewFileFixtureRepository(config)
	require.NoError(t, err)
	groupConfigRepository, err := repository.NewFileGroupConfigRepository(config)
	require.NoError(t, err)
	player := NewConsumerExecutor(scenarioRepository, fixtureRepository, groupConfigRepository)
	// AND a set of mock scenarios
	for i := 0; i < 3; i++ {
		require.NoError(t, scenarioRepository.Save(types.BuildTestScenario(types.Post, fmt.Sprintf("book_post_%d", i), "/api/:topic/books/:id", i)))
	}

	// WHEN matching partial url without id
	u, err := url.Parse("https://books.com/api/scifi/books?a=123&b=abc")
	require.NoError(t, err)
	// WHEN looking up todos by POST with different query param
	ctx := web.NewStubContext(&http.Request{
		Method: "POST",
		URL:    u,
		Header: http.Header{
			types.ContentTypeHeader: []string{"application/json"},
		},
	})
	err = player.Execute(ctx)
	// THEN it should not find it
	require.Error(t, err)

	// WHEN matching complete url with id
	u, err = url.Parse("https://books.com/api/scifi/books/13?a=123&b=abc")
	require.NoError(t, err)
	// WHEN looking up todos by POST with different query param
	ctx = web.NewStubContext(&http.Request{
		Method: "POST",
		URL:    u,
		Header: http.Header{
			types.ContentTypeHeader: []string{"application/json"},
			"ETag":                  []string{"123"},
		},
	})
	err = player.Execute(ctx)
	require.NoError(t, err)
	// THEN it should find it
	saved := ctx.Result.([]byte)
	require.Equal(t, "test body", string(saved))
}

func Test_ShouldLookupGetMockScenarios(t *testing.T) {
	config := types.BuildTestConfig()
	// GIVEN a mock scenario repository
	scenarioRepository, err := repository.NewFileAPIScenarioRepository(config)
	require.NoError(t, err)
	fixtureRepository, err := repository.NewFileFixtureRepository(config)
	require.NoError(t, err)
	groupConfigRepository, err := repository.NewFileGroupConfigRepository(config)
	require.NoError(t, err)
	player := NewConsumerExecutor(scenarioRepository, fixtureRepository, groupConfigRepository)
	// AND a set of mock scenarios
	for i := 0; i < 3; i++ {
		require.NoError(t, scenarioRepository.Save(types.BuildTestScenario(types.Get, fmt.Sprintf("books_get_%d", i), "/api/books/:topic/:id", i)))
	}
	// WHEN looking up non-existing API
	u, err := url.Parse("https://books.com/v2/topic/business/blah")
	require.NoError(t, err)
	ctx := web.NewStubContext(&http.Request{
		Method: "GET",
		URL:    u,
		Header: http.Header{
			types.ContentTypeHeader: []string{"application/json"},
		},
	})
	// THEN it should not find it
	err = player.Execute(ctx)
	require.Error(t, err)

	u, err = url.Parse("https://books.com/api/books/topic/business/202?a=123&b=abc")
	require.NoError(t, err)
	// WHEN looking up todos by PUT with different query param
	ctx = web.NewStubContext(&http.Request{
		Method: "GET",
		URL:    u,
		Header: http.Header{
			types.ContentTypeHeader: []string{"application/json"},
			"ETag":                  []string{"123"},
		},
	})
	err = player.Execute(ctx)
	require.NoError(t, err)
	// THEN it should find it
	saved := ctx.Result.([]byte)
	require.Equal(t, "test body", string(saved))
}

func Test_ShouldLookupDeleteMockScenarios(t *testing.T) {
	config := types.BuildTestConfig()
	// GIVEN a mock scenario repository and player
	mockScenarioRepository, err := repository.NewFileAPIScenarioRepository(config)
	require.NoError(t, err)
	fixtureRepository, err := repository.NewFileFixtureRepository(config)
	require.NoError(t, err)
	groupConfigRepository, err := repository.NewFileGroupConfigRepository(config)
	require.NoError(t, err)
	player := NewConsumerExecutor(mockScenarioRepository, fixtureRepository, groupConfigRepository)
	// AND a set of mock scenarios
	for i := 0; i < 3; i++ {
		require.NoError(t, mockScenarioRepository.Save(types.BuildTestScenario(types.Delete, fmt.Sprintf("books_delete_%d", i), "/api/books/:topic/:id", i)))
	}
	// WHEN looking up non-existing API
	u, err := url.Parse("https://books.com/business/books/202")
	ctx := web.NewStubContext(&http.Request{
		Method: "DELETE",
		URL:    u,
		Header: http.Header{
			types.ContentTypeHeader: []string{"application/json"},
		},
	})
	// THEN it should not find it
	err = player.Execute(ctx)
	require.Error(t, err)

	// WHEN looking up todos by PUT with different query param
	u, err = url.Parse("https://books.com/api/books/topic/business/202?a=123&b=abc")
	ctx = web.NewStubContext(&http.Request{
		Method: "DELETE",
		URL:    u,
		Header: http.Header{
			types.ContentTypeHeader: []string{"application/json"},
			"ETag":                  []string{"123"},
		},
	})
	err = player.Execute(ctx)
	require.NoError(t, err)
	// THEN it should find it
	saved := ctx.Result.([]byte)
	require.Equal(t, "test body", string(saved))
}

func Test_ShouldLookupDeleteMockScenariosWithBraces(t *testing.T) {
	config := types.BuildTestConfig()
	// GIVEN a mock scenario repository and player
	mockScenarioRepository, err := repository.NewFileAPIScenarioRepository(config)
	require.NoError(t, err)
	fixtureRepository, err := repository.NewFileFixtureRepository(config)
	require.NoError(t, err)
	groupConfigRepository, err := repository.NewFileGroupConfigRepository(config)
	require.NoError(t, err)
	player := NewConsumerExecutor(mockScenarioRepository, fixtureRepository, groupConfigRepository)
	// AND a set of mock scenarios
	for i := 0; i < 3; i++ {
		require.NoError(t, mockScenarioRepository.Save(types.BuildTestScenario(types.Delete, fmt.Sprintf("books_delete_%d", i), "/api/books/{topic}/{id}", i)))
	}
	// WHEN looking up non-existing API
	u, err := url.Parse("https://books.com/business/books/202")
	ctx := web.NewStubContext(&http.Request{
		Method: "DELETE",
		URL:    u,
		Header: http.Header{
			types.ContentTypeHeader: []string{"application/json"},
		},
	})
	// THEN it should not find it
	err = player.Execute(ctx)
	require.Error(t, err)

	// WHEN looking up todos by PUT with different query param
	u, err = url.Parse("https://books.com/api/books/topic/business/202?a=123&b=abc")
	ctx = web.NewStubContext(&http.Request{
		Method: "DELETE",
		URL:    u,
		Header: http.Header{
			types.ContentTypeHeader: []string{"application/json"},
			"ETag":                  []string{"123"},
		},
	})
	err = player.Execute(ctx)
	require.NoError(t, err)
	// THEN it should find it
	saved := ctx.Result.([]byte)
	require.Equal(t, "test body", string(saved))
}

func Test_ShouldGenerateGetCustomerResponse(t *testing.T) {
	config := types.BuildTestConfig()
	// GIVEN a mock scenario loaded from YAML
	b, err := os.ReadFile("../../fixtures/get_customer.yaml")
	require.NoError(t, err)

	// AND a mock scenario repository
	scenarioRepository, err := repository.NewFileAPIScenarioRepository(config)
	require.NoError(t, err)
	fixtureRepository, err := repository.NewFileFixtureRepository(config)
	require.NoError(t, err)
	groupConfigRepository, err := repository.NewFileGroupConfigRepository(config)
	require.NoError(t, err)
	player := NewConsumerExecutor(scenarioRepository, fixtureRepository, groupConfigRepository)

	b, err = fuzz.ParseTemplate("../../fixtures", b, map[string]any{"id": "123"})
	require.NoError(t, err)
	scenario := types.APIScenario{}
	// AND it should return valid mock scenario
	err = yaml.Unmarshal(b, &scenario)
	require.NoError(t, err)
	// AND a set of mock scenarios
	require.NoError(t, scenarioRepository.Save(&scenario))
	u, err := url.Parse("http://localhost/customers/123")
	// WHEN looking up non-existing API
	ctx := web.NewStubContext(
		&http.Request{
			Method: "GET",
			URL:    u,
			Header: http.Header{
				types.ContentTypeHeader: []string{"application/json"},
			},
		},
	)
	// THEN it should not find it
	err = player.Execute(ctx)
	require.NoError(t, err)

	b = ctx.Result.([]byte)
	obj := make(map[string]any)
	err = json.Unmarshal(b, &obj)
	require.NoError(t, err)
	require.Contains(t, obj["email"], "@")
}

func Test_ShouldLookupPutMockScenariosWithChaos(t *testing.T) {
	config := types.BuildTestConfig()
	// GIVEN a mock scenario repository
	scenarioRepository, err := repository.NewFileAPIScenarioRepository(config)
	require.NoError(t, err)
	fixtureRepository, err := repository.NewFileFixtureRepository(config)
	require.NoError(t, err)
	groupConfigRepository, err := repository.NewFileGroupConfigRepository(config)
	require.NoError(t, err)
	groupConfigRepository.Save("todos", &types.GroupConfig{
		ChaosEnabled: true,
	})
	player := NewConsumerExecutor(scenarioRepository, fixtureRepository, groupConfigRepository)
	// AND a set of mock scenarios
	for i := 0; i < 3; i++ {
		scenario := types.BuildTestScenario(types.Put, fmt.Sprintf("todo_put_%d", i), "/api/todos/{id}", i)
		scenario.Group = "todos"
		require.NoError(t, scenarioRepository.Save(scenario))
	}
	u, err := url.Parse("https://jsonplaceholder.typicode.com/blah")
	// WHEN looking up non-existing API
	ctx := web.NewStubContext(
		&http.Request{
			Method: "PUT",
			URL:    u,
			Header: http.Header{
				types.ContentTypeHeader: []string{"application/json"},
			},
		},
	)
	// THEN it should not find it
	err = player.Execute(ctx)
	require.Error(t, err)

	u, err = url.Parse("https://jsonplaceholder.typicode.com/api/todos/202?a=123&b=abc")
	require.NoError(t, err)
	// WHEN looking up todos by PUT with different query param
	ctx = web.NewStubContext(&http.Request{
		Method: "PUT",
		URL:    u,
		Header: http.Header{
			types.ContentTypeHeader: []string{"application/json"},
			"ETag":                  []string{"123"},
		},
	})
	// this may fail based on chaos
	player.Execute(ctx)
}

func Test_ShouldLookupPutMockScenariosWithBraces(t *testing.T) {
	config := types.BuildTestConfig()
	// GIVEN a mock scenario repository
	scenarioRepository, err := repository.NewFileAPIScenarioRepository(config)
	require.NoError(t, err)
	fixtureRepository, err := repository.NewFileFixtureRepository(config)
	require.NoError(t, err)
	groupConfigRepository, err := repository.NewFileGroupConfigRepository(config)
	require.NoError(t, err)
	player := NewConsumerExecutor(scenarioRepository, fixtureRepository, groupConfigRepository)
	// AND a set of mock scenarios
	for i := 0; i < 3; i++ {
		require.NoError(t, scenarioRepository.Save(types.BuildTestScenario(types.Put, fmt.Sprintf("todo_put_%d", i), "/api/todos/{id}", i)))
	}
	u, err := url.Parse("https://jsonplaceholder.typicode.com/blah")
	// WHEN looking up non-existing API
	ctx := web.NewStubContext(
		&http.Request{
			Method: "PUT",
			URL:    u,
			Header: http.Header{
				types.ContentTypeHeader: []string{"application/json"},
			},
		},
	)
	// THEN it should not find it
	err = player.Execute(ctx)
	require.Error(t, err)

	u, err = url.Parse("https://jsonplaceholder.typicode.com/api/todos/202?a=123&b=abc")
	require.NoError(t, err)
	// WHEN looking up todos by PUT with different query param
	ctx = web.NewStubContext(&http.Request{
		Method: "PUT",
		URL:    u,
		Header: http.Header{
			types.ContentTypeHeader: []string{"application/json"},
			"ETag":                  []string{"123"},
		},
	})
	err = player.Execute(ctx)
	require.NoError(t, err)
	// THEN it should find it
	saved := ctx.Result.([]byte)
	require.Equal(t, "test body", string(saved))
}

func Test_ShouldAddMockResponseWithNilRequestWithoutQueryParams(t *testing.T) {
	config := types.BuildTestConfig()
	// GIVEN a mock scenario and fixture repository
	scenarioRepository, err := repository.NewFileAPIScenarioRepository(config)
	require.NoError(t, err)
	fixtureRepository, err := repository.NewFileFixtureRepository(config)
	require.NoError(t, err)
	reqHeader := http.Header{"X1": []string{"val1"}}
	resHeader := http.Header{"X1": []string{"val1"}}
	matchedScenario := types.BuildTestScenario(types.Post, "name", "/path", 10)
	matchedScenario.Response.ContentsFile = "lines.txt"
	_ = os.MkdirAll("../../mock_tests/api_contracts/path/POST", 0755)
	_ = os.WriteFile("../../mock_tests/api_contracts/path/POST/lines.txt.dat", []byte("test"), 0644)
	req := &http.Request{Body: nil}
	_, err = AddMockResponse(
		req,
		reqHeader,
		resHeader,
		matchedScenario,
		time.Now(),
		time.Now(),
		scenarioRepository,
		fixtureRepository,
	)
	require.Error(t, err)
	require.Contains(t, err.Error(), `didn't match required request query param 'a' with regex '\d+'`)
}

func Test_ShouldAddMockResponseWithNilRequest(t *testing.T) {
	config := types.BuildTestConfig()
	// GIVEN a mock scenario and fixture repository
	scenarioRepository, err := repository.NewFileAPIScenarioRepository(config)
	require.NoError(t, err)
	fixtureRepository, err := repository.NewFileFixtureRepository(config)
	require.NoError(t, err)
	reqHeader := http.Header{"X1": []string{"val1"}, "ETag": []string{"123"}}
	resHeader := http.Header{"X1": []string{"val1"}}
	matchedScenario := types.BuildTestScenario(types.Post, "name", "/path", 10)
	matchedScenario.Response.ContentsFile = "lines.txt"
	_ = os.MkdirAll("../../mock_tests/api_contracts/path/POST", 0755)
	_ = os.WriteFile("../../mock_tests/api_contracts/path/POST/lines.txt.dat", []byte("test"), 0644)
	u, err := url.Parse("https://jsonplaceholder.typicode.com/api/todos/202?a=123&b=abc")
	require.NoError(t, err)
	req := &http.Request{Body: nil, URL: u}
	_, err = AddMockResponse(
		req,
		reqHeader,
		resHeader,
		matchedScenario,
		time.Now(),
		time.Now(),
		scenarioRepository,
		fixtureRepository,
	)
	require.NoError(t, err)
}

func Test_ShouldNotAddMockResponseWithoutQueryParams(t *testing.T) {
	config := types.BuildTestConfig()
	// GIVEN a mock scenario and fixture repository
	scenarioRepository, err := repository.NewFileAPIScenarioRepository(config)
	require.NoError(t, err)
	fixtureRepository, err := repository.NewFileFixtureRepository(config)
	require.NoError(t, err)
	reqHeader := http.Header{"X1": []string{"val1"}}
	resHeader := http.Header{"X1": []string{"val1"}}
	matchedScenario := types.BuildTestScenario(types.Post, "name", "/path", 10)
	matchedScenario.Response.ContentsFile = "lines.txt"
	_ = os.MkdirAll("../../mock_tests/api_contracts/path/POST", 0755)
	_ = os.WriteFile("../../mock_tests/api_contracts/path/POST/lines.txt.dat", []byte("test"), 0644)
	data := []byte("test data")
	reader := io.NopCloser(bytes.NewReader(data))
	req := &http.Request{Body: reader}
	_, err = AddMockResponse(
		req,
		reqHeader,
		resHeader,
		matchedScenario,
		time.Now(),
		time.Now(),
		scenarioRepository,
		fixtureRepository,
	)
	require.Error(t, err)
}

func Test_ShouldAddMockResponseWithRequest(t *testing.T) {
	config := types.BuildTestConfig()
	// GIVEN a mock scenario and fixture repository
	scenarioRepository, err := repository.NewFileAPIScenarioRepository(config)
	require.NoError(t, err)
	fixtureRepository, err := repository.NewFileFixtureRepository(config)
	require.NoError(t, err)
	reqHeader := http.Header{"X1": []string{"val1"}, "ETag": []string{"123"}}
	resHeader := http.Header{"X1": []string{"val1"}}
	matchedScenario := types.BuildTestScenario(types.Post, "name", "/path", 10)
	matchedScenario.Response.ContentsFile = "lines.txt"
	_ = os.MkdirAll("../../mock_tests/api_contracts/path/POST", 0755)
	_ = os.WriteFile("../../mock_tests/api_contracts/path/POST/lines.txt.dat", []byte("test"), 0644)
	data := []byte("test data")
	reader := io.NopCloser(bytes.NewReader(data))
	u, _ := url.Parse("http://localhost:8080?a=123&b=abcd")
	req := &http.Request{
		Body: reader,
		URL:  u,
	}
	_, err = AddMockResponse(
		req,
		reqHeader,
		resHeader,
		matchedScenario,
		time.Now(),
		time.Now(),
		scenarioRepository,
		fixtureRepository,
	)
	require.NoError(t, err)
}
