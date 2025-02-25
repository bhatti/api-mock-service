package contract

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/bhatti/api-mock-service/internal/fuzz"
	"github.com/bhatti/api-mock-service/internal/oapi"
	"github.com/bhatti/api-mock-service/internal/pm"
	"github.com/bhatti/api-mock-service/internal/repository"
	"github.com/bhatti/api-mock-service/internal/types"
	"github.com/bhatti/api-mock-service/internal/web"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
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
	player := NewConsumerExecutor(config, scenarioRepository, fixtureRepository, groupConfigRepository)
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
			types.ETagHeader:          []string{"12"},
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
			types.ETagHeader:          []string{"123"},
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
	specs, _, err := oapi.Parse(context.Background(), &types.Configuration{}, data, dataTempl)

	require.NoError(t, err)
	require.Len(t, specs, 6)
	// AND executor
	player := NewConsumerExecutor(config, scenarioRepository, fixtureRepository, groupConfigRepository)
	for _, spec := range specs {
		if spec.Response.StatusCode != 200 {
			continue
		}
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
	player := NewConsumerExecutor(config, scenarioRepository, fixtureRepository, groupConfigRepository)
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
			types.ETagHeader:        []string{"123"},
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
	player := NewConsumerExecutor(config, scenarioRepository, fixtureRepository, groupConfigRepository)
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
			types.ETagHeader:        []string{"123"},
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
	player := NewConsumerExecutor(config, mockScenarioRepository, fixtureRepository, groupConfigRepository)
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
			types.ETagHeader:        []string{"123"},
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
	player := NewConsumerExecutor(config, mockScenarioRepository, fixtureRepository, groupConfigRepository)
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
			types.ETagHeader:        []string{"123"},
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
	player := NewConsumerExecutor(config, scenarioRepository, fixtureRepository, groupConfigRepository)

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
	err = groupConfigRepository.Save("todos", &types.GroupConfig{
		ChaosEnabled: true,
	})
	require.NoError(t, err)
	player := NewConsumerExecutor(config, scenarioRepository, fixtureRepository, groupConfigRepository)
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
			types.ETagHeader:        []string{"123"},
		},
	})
	// this may fail based on chaos
	_ = player.Execute(ctx)
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
	player := NewConsumerExecutor(config, scenarioRepository, fixtureRepository, groupConfigRepository)
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
			types.ETagHeader:        []string{"123"},
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
	groupConfigRepository, err := repository.NewFileGroupConfigRepository(config)
	require.NoError(t, err)
	reqHeader := http.Header{"X1": []string{"val1"}}
	resHeader := http.Header{"X1": []string{"val1"}}
	matchedScenario := types.BuildTestScenario(types.Post, "name", "/path", 10)
	matchedScenario.Response.ContentsFile = "lines.txt"
	_ = os.MkdirAll("../../mock_tests/api_contracts/path/POST", 0755)
	_ = os.WriteFile("../../mock_tests/api_contracts/path/POST/lines.txt.dat", []byte("test"), 0644)
	req := &http.Request{Body: nil}
	_, _, err = AddMockResponse(
		req,
		reqHeader,
		resHeader,
		matchedScenario,
		time.Now(),
		time.Now(),
		config,
		scenarioRepository,
		fixtureRepository,
		groupConfigRepository,
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
	groupConfigRepository, err := repository.NewFileGroupConfigRepository(config)
	require.NoError(t, err)
	reqHeader := http.Header{"X1": []string{"val1"}, types.ETagHeader: []string{"123"}}
	resHeader := http.Header{"X1": []string{"val1"}}
	matchedScenario := types.BuildTestScenario(types.Post, "name", "/path", 10)
	matchedScenario.Response.ContentsFile = "lines.txt"
	_ = os.MkdirAll("../../mock_tests/api_contracts/path/POST", 0755)
	_ = os.WriteFile("../../mock_tests/api_contracts/path/POST/lines.txt.dat", []byte("test"), 0644)
	u, err := url.Parse("https://jsonplaceholder.typicode.com/api/todos/202?a=123&b=abc")
	require.NoError(t, err)
	req := &http.Request{Body: nil, URL: u}
	_, _, err = AddMockResponse(
		req,
		reqHeader,
		resHeader,
		matchedScenario,
		time.Now(),
		time.Now(),
		config,
		scenarioRepository,
		fixtureRepository,
		groupConfigRepository,
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
	groupConfigRepository, err := repository.NewFileGroupConfigRepository(config)
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
	_, _, err = AddMockResponse(
		req,
		reqHeader,
		resHeader,
		matchedScenario,
		time.Now(),
		time.Now(),
		config,
		scenarioRepository,
		fixtureRepository,
		groupConfigRepository,
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
	groupConfigRepository, err := repository.NewFileGroupConfigRepository(config)
	require.NoError(t, err)
	reqHeader := http.Header{"X1": []string{"val1"}, types.ETagHeader: []string{"123"}}
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
	_, _, err = AddMockResponse(
		req,
		reqHeader,
		resHeader,
		matchedScenario,
		time.Now(),
		time.Now(),
		config,
		scenarioRepository,
		fixtureRepository,
		groupConfigRepository,
	)
	require.NoError(t, err)
}

func Test_ShouldExecutePostmanAPISuite(t *testing.T) {
	// GIVEN configuration and repositories
	config := types.BuildTestConfig()
	scenarioRepository, err := repository.NewFileAPIScenarioRepository(config)
	require.NoError(t, err)
	fixtureRepository, err := repository.NewFileFixtureRepository(config)
	require.NoError(t, err)
	groupConfigRepository, err := repository.NewFileGroupConfigRepository(config)
	require.NoError(t, err)
	accessToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IlRlc3QgVXNlciIsImlhdCI6MTUxNjIzOTAyMn0.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"
	resourceID := "res-12345"

	// Create executor
	player := NewConsumerExecutor(config, scenarioRepository, fixtureRepository, groupConfigRepository)

	// Load and parse the Postman collection
	file, err := os.Open("../../fixtures/postman_basic.json")

	// Parse collection
	collection, err := pm.ParseCollection(file)
	require.NoError(t, err)
	require.Equal(t, "API Testing Suite", collection.Info.Name)

	// Convert to scenarios and store them
	scenarios, vars := pm.ConvertPostmanToScenarios(config, collection, time.Now(), time.Now())
	require.NotEmpty(t, scenarios)
	require.NotEmpty(t, vars.Variables)

	// Save all scenarios
	for _, scenario := range scenarios {
		if strings.Contains(scenario.Name, "Get JWT Token") {
			scenario.Response.Contents = "{\"access_token\": \"" + accessToken + "\"}"
		} else if strings.Contains(scenario.Name, "Update Resource") {
			scenario.Response.Contents = "{\"id\": \"" + resourceID + "\", \"name\": \"updated name\"}"
		} else {
			scenario.Response.Contents = "{\"id\": \"" + resourceID + "\", \"name\": \"test resource\"}"
		}

		err = scenarioRepository.Save(scenario)
		require.NoError(t, err)
	}

	err = scenarioRepository.SaveVariables(vars)
	require.NoError(t, err)

	// Find auth scenario
	var authScenario *types.APIScenario
	for _, s := range scenarios {
		if strings.Contains(s.Name, "Get JWT Token") {
			authScenario = s
			break
		}
	}
	require.NotNil(t, authScenario, "Auth scenario not found")
	// Set up variables for testing
	apiKey := "post-test-api-key-123"
	baseURL = "https://api.example.com"

	// --- Step 1: Test Auth API ---
	t.Run("Auth API Test", func(t *testing.T) {
		// Create mock auth request
		authUrl, err := url.Parse(baseURL + authScenario.Path)
		require.NoError(t, err)

		// Set up headers based on auth scenario request
		headers := http.Header{}
		for k, v := range authScenario.Request.Headers {
			if len(v) > 0 {
				headers.Set(k, replaceVariables(v, map[string]string{
					"api_key": apiKey,
				}))
			}
		}
		headers.Set("Content-Type", "application/json")

		// Create request context
		ctx := web.NewStubContext(&http.Request{
			Method: string(authScenario.Method),
			URL:    authUrl,
			Header: headers,
			Body:   io.NopCloser(strings.NewReader(`{"jws": "signed-payload-123"}`)),
		})

		// Execute auth request
		err = player.Execute(ctx)
		require.NoError(t, err)

		// Verify response
		require.Equal(t, 0, ctx.Response().Status)
		require.Equal(t, "", ctx.Response().Header().Get("Content-Type")) // application/json

		// Parse response to get token
		var authResponse map[string]interface{}
		err = json.Unmarshal(ctx.Result.([]byte), &authResponse)
		require.NoError(t, err)
		require.Equal(t, accessToken, authResponse["access_token"])
	})

	// --- Step 2: Test CRUD APIs ---
	// Find all CRUD scenarios
	var crudScenarios []*types.APIScenario
	for _, s := range scenarios {
		if strings.Contains(s.Group, "CRUD Operations") {
			crudScenarios = append(crudScenarios, s)
		}
	}
	require.NotEmpty(t, crudScenarios, "CRUD scenarios not found")

	// For each CRUD operation
	for _, scenario := range crudScenarios {
		t.Run(fmt.Sprintf("CRUD API Test: %s", scenario.Name), func(t *testing.T) {
			// Replace path variables
			path := replaceVariables(scenario.Path, map[string]string{
				"resource_id": resourceID,
			})

			// Create request URL
			reqUrl, err := url.Parse(baseURL + path)
			require.NoError(t, err)

			// Set up headers
			headers := http.Header{}
			for k, v := range scenario.Request.Headers {
				if len(v) > 0 {
					headers.Set(k, replaceVariables(v, map[string]string{
						"api_key": apiKey,
					}))
				}
			}

			// Add auth header if needed
			if _, ok := scenario.Authentication["bearer"]; ok {
				headers.Set("Authorization", "Bearer "+accessToken)
			}

			headers.Set("Content-Type", "application/json")

			req := &http.Request{
				Method: string(scenario.Method),
				URL:    reqUrl,
				Header: headers,
			}
			// Create request body if needed
			if scenario.Method != types.Get && scenario.Method != types.Delete {
				req.Body = io.NopCloser(strings.NewReader(scenario.Request.Contents))
			}

			// Create request context
			ctx := web.NewStubContext(req)

			// Execute request
			err = player.Execute(ctx)
			require.NoError(t, err)

			// Verify response
			//require.Equal(t, scenario.Response.StatusCode, ctx.Response().Status)
			//require.Equal(t, "application/json", ctx.Response().Header().Get("Content-Type"))

			var responseData map[string]interface{}
			err = json.Unmarshal(ctx.Result.([]byte), &responseData)
			require.NoError(t, err)

			switch scenario.Method {
			case types.Post:
				require.Equal(t, resourceID, responseData["id"])
				require.Equal(t, "test resource", responseData["name"])
			case types.Get:
				require.Equal(t, resourceID, responseData["id"])
				require.Equal(t, "test resource", responseData["name"])
			case types.Patch:
				require.Equal(t, resourceID, responseData["id"])
				require.Equal(t, "updated name", responseData["name"])
			}
		})
	}

	// --- Step 3: Test Full Flow ---
	t.Run("Full API Flow", func(t *testing.T) {
		// Find scenarios for each operation
		var getAuthScenario, createResourceScenario, getResourceScenario, updateResourceScenario, deleteResourceScenario *types.APIScenario

		for _, s := range scenarios {
			if s.Name == "Get JWT Token" || strings.Contains(s.Name, "Get JWT Token") {
				getAuthScenario = s
			} else if strings.Contains(s.Name, "Create Resource") {
				createResourceScenario = s
			} else if strings.Contains(s.Name, "Get Resource") {
				getResourceScenario = s
			} else if strings.Contains(s.Name, "Update Resource") {
				updateResourceScenario = s
			} else if strings.Contains(s.Name, "Delete Resource") {
				deleteResourceScenario = s
			}
		}

		require.NotNil(t, getAuthScenario)
		require.NotNil(t, createResourceScenario)
		require.NotNil(t, getResourceScenario)
		require.NotNil(t, updateResourceScenario)
		require.NotNil(t, deleteResourceScenario)

		// Global context for the flow
		accessToken = ""
		createdResourceID := ""

		// 1. Get Auth Token
		t.Log("Step 1: Authenticate and get token")
		authUrl, _ := url.Parse(baseURL + getAuthScenario.Path)
		authRequestCtx := web.NewStubContext(&http.Request{
			Method: string(getAuthScenario.Method),
			URL:    authUrl,
			Header: http.Header{
				"Content-Type": []string{"application/json"},
				"x-api-key":    []string{apiKey},
			},
			Body: io.NopCloser(strings.NewReader(`{"jws": "signed-payload-123"}`)),
		})

		err = player.Execute(authRequestCtx)
		require.NoError(t, err)

		// Get token from response
		var authResp map[string]interface{}
		require.NoError(t, json.Unmarshal(authRequestCtx.Result.([]byte), &authResp))
		require.Contains(t, authResp, "access_token")
		accessToken = authResp["access_token"].(string)
		require.NotEmpty(t, accessToken)

		// 2. Create Resource
		t.Log("Step 2: Create a new resource")
		createUrl, _ := url.Parse(baseURL + createResourceScenario.Path)
		createRequestCtx := web.NewStubContext(&http.Request{
			Method: string(createResourceScenario.Method),
			URL:    createUrl,
			Header: http.Header{
				"Content-Type":  []string{"application/json"},
				"x-api-key":     []string{apiKey},
				"Authorization": []string{"Bearer " + accessToken},
			},
			Body: io.NopCloser(strings.NewReader(createResourceScenario.Request.Contents)),
		})

		err = player.Execute(createRequestCtx)
		require.NoError(t, err)

		// Get resource ID from response
		var createResp map[string]interface{}
		require.NoError(t, json.Unmarshal(createRequestCtx.Result.([]byte), &createResp))
		require.Contains(t, createResp, "id")

		createdResourceID = createResp["id"].(string)
		require.NotEmpty(t, createdResourceID)

		// 3. Get Resource
		t.Log("Step 3: Retrieve the created resource")
		getPath := strings.Replace(getResourceScenario.Path, "{{resource_id}}", createdResourceID, -1)
		getUrl, _ := url.Parse(baseURL + getPath)
		getRequestCtx := web.NewStubContext(&http.Request{
			Method: string(getResourceScenario.Method),
			URL:    getUrl,
			Header: http.Header{
				"Content-Type":  []string{"application/json"},
				"Authorization": []string{"Bearer " + accessToken},
			},
		})

		err = player.Execute(getRequestCtx)
		require.NoError(t, err)

		// 4. Update Resource
		t.Log("Step 4: Update the resource")
		updatePath := strings.Replace(updateResourceScenario.Path, "{{resource_id}}", createdResourceID, -1)
		updateUrl, _ := url.Parse(baseURL + updatePath)
		updateRequestCtx := web.NewStubContext(&http.Request{
			Method: string(updateResourceScenario.Method),
			URL:    updateUrl,
			Header: http.Header{
				"Content-Type":  []string{"application/json"},
				"Authorization": []string{"Bearer " + accessToken},
			},
			Body: io.NopCloser(strings.NewReader(`{"name": "updated name"}`)),
		})

		// Set up update response
		err = player.Execute(updateRequestCtx)
		require.NoError(t, err)

		// 5. Delete Resource
		t.Log("Step 5: Delete the resource")
		deletePath := strings.Replace(deleteResourceScenario.Path, "{{resource_id}}", createdResourceID, -1)
		deleteUrl, _ := url.Parse(baseURL + deletePath)
		deleteRequestCtx := web.NewStubContext(&http.Request{
			Method: string(deleteResourceScenario.Method),
			URL:    deleteUrl,
			Header: http.Header{
				"Content-Type":  []string{"application/json"},
				"Authorization": []string{"Bearer " + accessToken},
			},
		})

		err = player.Execute(deleteRequestCtx)
		require.NoError(t, err)
		require.Equal(t, 0, deleteRequestCtx.Response().Status) //204
	})

	// --- Step 4: Test Just Create without Auth ---
	t.Run("Create Resource without Auth", func(t *testing.T) {
		var createResourceScenario *types.APIScenario

		for _, s := range scenarios {
			if strings.Contains(s.Name, "Create Resource") {
				createResourceScenario = s
			}
		}

		require.NotNil(t, createResourceScenario)
		createResourceScenario.VariablesFile = ""
		delete(createResourceScenario.Request.Variables, "access_token")
		err = scenarioRepository.Save(createResourceScenario)
		require.NoError(t, err)

		// Global context for the flow
		accessToken = ""
		createdResourceID := ""

		// Create Resource - it should automatically call auth
		t.Log("Create a new resource")
		createUrl, _ := url.Parse(baseURL + createResourceScenario.Path)
		req := &http.Request{
			Method: string(createResourceScenario.Method),
			URL:    createUrl,
			Header: http.Header{
				"Content-Type": []string{"application/json"},
				"x-api-key":    []string{apiKey},
				//"Authorization": []string{"Bearer " + accessToken},
			},
			Body: io.NopCloser(strings.NewReader(createResourceScenario.Request.Contents)),
		}
		require.NotContains(t, req.Header, "access_key")
		createRequestCtx := web.NewStubContext(req)
		err = player.Execute(createRequestCtx)
		require.NoError(t, err)

		// Get resource ID from response
		var createResp map[string]interface{}
		require.NoError(t, json.Unmarshal(createRequestCtx.Result.([]byte), &createResp))
		require.Contains(t, createResp, "id")

		createdResourceID = createResp["id"].(string)
		require.NotEmpty(t, createdResourceID)
		require.Equal(t, accessToken, req.Header.Get("access_key"))
	})
}

// Helper function to replace variables in strings
func replaceVariables(input string, vars map[string]string) string {
	result := input
	for k, v := range vars {
		result = strings.Replace(result, "{{"+k+"}}", v, -1)
	}
	return result
}
