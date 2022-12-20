package repository

import (
	"bytes"
	"fmt"
	"gopkg.in/yaml.v3"
	"strconv"
	"testing"
	"time"

	"github.com/bhatti/api-mock-service/internal/utils"

	"github.com/bhatti/api-mock-service/internal/types"
	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"
)

const mockPath = "//abc//\\def/123/"

func Test_ShouldRawSaveAndLoadMockScenarios(t *testing.T) {
	// GIVEN a mock scenario repository
	repo, err := NewFileMockScenarioRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)
	// AND mock scenario
	scenario := buildScenario(types.Post, "test1", mockPath, 10)
	b, err := yaml.Marshal(scenario)
	require.NoError(t, err)
	// WHEN saving scenario
	err = repo.SaveRaw(utils.NopCloser(bytes.NewReader(b)))
	// THEN it should succeed
	require.NoError(t, err)

	// AND should return saved scenario
	b, err = repo.LoadRaw(scenario.Method, scenario.Name, scenario.Path)
	require.NoError(t, err)
	require.True(t, len(b) > 0)
}

func Test_ShouldSaveAndGetMockScenarios(t *testing.T) {
	// GIVEN a mock scenario repository
	repo, err := NewFileMockScenarioRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)
	// AND mock scenario
	scenario := buildScenario(types.Post, "test1", mockPath, 10)
	// WHEN saving scenario
	err = repo.Save(scenario)
	// THEN it should succeed
	require.NoError(t, err)

	// AND should return saved scenario
	saved, err := repo.Lookup(scenario.ToKeyData(), nil)
	require.NoError(t, err)
	require.NoError(t, scenario.ToKeyData().Equals(saved.ToKeyData()))
}

func Test_ShouldParsePredicateForNthRequest(t *testing.T) {
	keyData1 := buildScenario(types.Post, "test1", mockPath, 1).ToKeyData()
	keyData2 := buildScenario(types.Post, "test2", mockPath, 2).ToKeyData()
	require.True(t, utils.MatchScenarioPredicate(keyData1, keyData2, 0))
	keyData1.MatchQueryParams = map[string]string{"a": `\d+`, "b": "abc"}
	keyData2.MatchQueryParams = map[string]string{"a": `\d+`, "b": "abc"}
	keyData1.Predicate = `{{NthRequest 3}}`
	require.True(t, utils.MatchScenarioPredicate(keyData1, keyData2, 0))
	require.False(t, utils.MatchScenarioPredicate(keyData1, keyData2, 2))
	require.True(t, utils.MatchScenarioPredicate(keyData1, keyData2, 3))
}

func Test_ShouldNotGetAfterDeletingMockScenarios(t *testing.T) {
	// GIVEN a mock scenario repository
	repo, err := NewFileMockScenarioRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)
	// AND mock scenario
	scenario := buildScenario(types.Post, "test1", mockPath, 20)
	// WHEN saving scenario
	err = repo.Save(scenario)
	// THEN it should succeed
	require.NoError(t, err)

	// AND should return saved scenario
	saved, err := repo.Lookup(scenario.ToKeyData(), make(map[string]any))
	require.NoError(t, err)
	require.NoError(t, scenario.ToKeyData().Equals(saved.ToKeyData()))

	// But WHEN DELETING the mock scenario
	err = repo.Delete(scenario.Method, scenario.Name, scenario.Path)
	require.NoError(t, err)

	// THEN GET should fail
	_, err = repo.Lookup(scenario.ToKeyData(), make(map[string]any))
	require.Error(t, err)
}

func Test_ShouldListMockScenariosNames(t *testing.T) {
	// GIVEN a mock scenario repository
	repo, err := NewFileMockScenarioRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)
	// AND a set of mock scenarios
	for i := 0; i < 10; i++ {
		scenario := buildScenario(types.Post, fmt.Sprintf("test_%d", i), mockPath, 30)
		err = repo.Save(scenario)
		require.NoError(t, err)
	}
	// WHEN listing scenarios
	names, err := repo.GetScenariosNames(types.Post, mockPath)
	require.NoError(t, err)
	for i := 0; i < 10; i++ {
		require.Equal(t, fmt.Sprintf("test_%d", i), names[i])
		err = repo.Delete(types.Post, names[i], mockPath)
		require.NoError(t, err)
	}
}

func Test_ShouldLookupAllByGroup(t *testing.T) {
	// GIVEN a mock scenario repository
	repo, err := NewFileMockScenarioRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)
	// AND a set of mock scenarios
	for i := 0; i < 10; i++ {
		require.NoError(t, repo.Save(buildScenario(types.Get, fmt.Sprintf("todo_%d", i), "/v2/todos", i)))
		require.NoError(t, repo.Save(buildScenario(types.Get, fmt.Sprintf("book_%d", i), "/v2/books", i)))
		require.NoError(t, repo.Save(buildScenario(types.Post, fmt.Sprintf("todo_%d", i), "/v2/todos", i)))
		require.NoError(t, repo.Save(buildScenario(types.Post, fmt.Sprintf("book_%d", i), "/v2/books", i)))
		require.NoError(t, repo.Save(buildScenario(types.Put, fmt.Sprintf("todo_%d", i), "/v2/todos", i)))
		require.NoError(t, repo.Save(buildScenario(types.Put, fmt.Sprintf("book_%d", i), "/v2/books", i)))
	}
	// WHEN looking up by group
	matched := repo.LookupAllByGroup("/v2/todos")
	// THEN it should return it
	assert.Equal(t, 30, len(matched))
	matched = repo.LookupAllByGroup("/v2/books")
	// THEN it should return it
	assert.Equal(t, 30, len(matched))
}

func Test_ShouldLookupPutMockScenarios(t *testing.T) {
	// GIVEN a mock scenario repository
	repo, err := NewFileMockScenarioRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)
	// AND a set of mock scenarios
	for i := 0; i < 10; i++ {
		require.NoError(t, repo.Save(buildScenario(types.Put, fmt.Sprintf("todo_put_%d", i), "/api/todos/:id", i)))
		require.NoError(t, repo.Save(buildScenario(types.Put, fmt.Sprintf("book_put_%d", i), "/api/:topic/books/:id", i)))
	}
	// WHEN looking up todos by POST without criteria
	matched, _ := repo.LookupAll(&types.MockScenarioKeyData{})
	// THEN it should not find it
	assert.Equal(t, 0, len(matched))

	// WHEN looking up todos by PUT with different query param
	matched, _ = repo.LookupAll(&types.MockScenarioKeyData{
		Method:           types.Put,
		Path:             "/api/todos/11",
		MatchQueryParams: map[string]string{"a": "22"},
		MatchHeaders: map[string]string{
			types.ContentTypeHeader: "application/json",
		},
	})
	// THEN it should not find it
	assert.Equal(t, 0, len(matched))

	// WHEN looking up todos by matching path and query params
	matched, _ = repo.LookupAll(&types.MockScenarioKeyData{
		Method:           types.Put,
		Path:             "/api/todos/12",
		MatchQueryParams: map[string]string{"a": "1"},
		MatchHeaders: map[string]string{
			types.ContentTypeHeader: "application/json",
		},
	})
	// THEN it should not find it without headers
	assert.Equal(t, 0, len(matched))

	// WHEN looking up todos by matching path, headers and query params
	matched, _ = repo.LookupAll(&types.MockScenarioKeyData{
		Method:           types.Put,
		Path:             "/api/todos/12",
		MatchQueryParams: map[string]string{"a": "1", "b": "abc"},
		MatchHeaders: map[string]string{
			types.ContentTypeHeader: "application/json",
			"ETag":                  "981",
		},
	})
	// THEN it should find it
	assert.Equal(t, 10, len(matched))
	for i, m := range matched {
		groups := m.MatchGroups(fmt.Sprintf("/api/todos/%d", i))
		require.Equal(t, strconv.Itoa(i), groups["id"])
		assert.Equal(t, 1, len(groups))
	}
	_, err = repo.Lookup(&types.MockScenarioKeyData{
		Method:           types.Put,
		Path:             "/api/todos/12",
		MatchQueryParams: map[string]string{"a": "1", "b": "abc", "n": "0"},
		MatchHeaders: map[string]string{
			types.ContentTypeHeader: "application/json",
			"ETag":                  "981",
		},
	}, make(map[string]any))
	require.NoError(t, err)

	//
	// WHEN looking up books by POST with topic
	matched, _ = repo.LookupAll(&types.MockScenarioKeyData{
		Method:           types.Put,
		Path:             "/api/mytopic/books/11",
		MatchQueryParams: map[string]string{"a": "1", "b": "abc", "n": "0"},
		MatchHeaders: map[string]string{
			types.ContentTypeHeader: "application/json",
			"ETag":                  "981",
		},
	})
	// THEN it should find it
	assert.Equal(t, 10, len(matched))
	for i, m := range matched {
		groups := m.MatchGroups(fmt.Sprintf("/api/mytopic/books/%d", i))
		require.Equal(t, strconv.Itoa(i), groups["id"])
		require.Equal(t, "mytopic", groups["topic"])
		assert.Equal(t, 2, len(groups))
	}
	_, err = repo.Lookup(&types.MockScenarioKeyData{
		Method:           types.Put,
		Path:             "/api/mytopic/books/11",
		MatchQueryParams: map[string]string{"a": "1", "b": "abc", "n": "0"},
		MatchHeaders: map[string]string{
			types.ContentTypeHeader: "application/json",
			"ETag":                  "981",
		},
	}, make(map[string]any))
	require.NoError(t, err)
}

func Test_ShouldListMockScenarios(t *testing.T) {
	// GIVEN a mock scenario repository
	repo, err := NewFileMockScenarioRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)
	// AND a set of mock scenarios
	for i := 0; i < 10; i++ {
		require.NoError(t, repo.Save(buildScenario(types.Get, fmt.Sprintf("todo_post_%d", i), "/v3/api/todos", i)))
		require.NoError(t, repo.Save(buildScenario(types.Post, fmt.Sprintf("todo_post_%d", i), "/v3/api/todos", i)))
		require.NoError(t, repo.Save(buildScenario(types.Delete, fmt.Sprintf("todo_post_%d", i), "/v3/api/todos", i)))
		require.NoError(t, repo.Save(buildScenario(types.Get, fmt.Sprintf("book_post_%d", i), "/v3/api/books", i)))
		require.NoError(t, repo.Save(buildScenario(types.Post, fmt.Sprintf("book_post_%d", i), "/v3/api/books", i)))
		require.NoError(t, repo.Save(buildScenario(types.Delete, fmt.Sprintf("book_post_%d", i), "/v3/api/books", i)))
	}
	// WHEN listing mock scenario
	all := repo.ListScenarioKeyData("")
	// THEN it should succeed
	assert.True(t, len(all) >= 60, fmt.Sprintf("size %d", len(all)))
	for _, next := range all {
		scenario, err := repo.Lookup(next, make(map[string]any))
		require.NoError(t, err)
		require.Equal(t, next.Name, scenario.Name)
	}
	// WHEN listing mock scenario with matching group
	all = repo.ListScenarioKeyData("/v3/api/todos")
	// THEN it should succeed
	assert.True(t, len(all) >= 30, fmt.Sprintf("size %d", len(all)))
	// WHEN listing mock scenario with non-matching group
	all = repo.ListScenarioKeyData("/v3/api/books")
	// THEN it should succeed
	assert.Equal(t, 30, len(all), fmt.Sprintf("size %d", len(all)))
}

func Test_ShouldLookupPostMockScenarios(t *testing.T) {
	// GIVEN a mock scenario repository
	repo, err := NewFileMockScenarioRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)
	// AND a set of mock scenarios
	for i := 0; i < 10; i++ {
		require.NoError(t, repo.Save(buildScenario(types.Post, fmt.Sprintf("todo_post_%d", i), "/v3/todos", i)))
		require.NoError(t, repo.Save(buildScenario(types.Post, fmt.Sprintf("book_post_%d", i), "/v3/:topic/books", i)))
	}
	// WHEN looking up todos by POST without criteria
	matched, _ := repo.LookupAll(&types.MockScenarioKeyData{})
	// THEN it should not find it
	assert.Equal(t, 0, len(matched))

	// WHEN looking up todos by POST with different query param
	matched, _ = repo.LookupAll(&types.MockScenarioKeyData{
		Method:           types.Post,
		Path:             "/v3/todos",
		MatchQueryParams: map[string]string{"a": "11"},
		MatchHeaders: map[string]string{
			types.ContentTypeHeader: "application/json",
		},
	})
	// THEN it should not find it
	assert.Equal(t, 0, len(matched))

	// WHEN looking up todos by matching path and query params
	matched, _ = repo.LookupAll(&types.MockScenarioKeyData{
		Method:           types.Post,
		Path:             "/v3/todos",
		MatchQueryParams: map[string]string{"a": "1"},
		MatchHeaders: map[string]string{
			types.ContentTypeHeader: "application/json",
		},
	})
	// THEN it should not find it
	assert.Equal(t, 0, len(matched))
	// WHEN looking up todos by matching path, headers and query params
	matched, _ = repo.LookupAll(&types.MockScenarioKeyData{
		Method:           types.Post,
		Path:             "/v3/todos",
		MatchQueryParams: map[string]string{"a": "1", "b": "abc"},
		MatchHeaders: map[string]string{
			types.ContentTypeHeader: "application/json",
			"ETag":                  "981",
		},
	})
	// THEN it should find it
	assert.Equal(t, 10, len(matched))
	for _, m := range matched {
		assert.Equal(t, 0, len(m.MatchGroups("/v3/todos")))
	}

	//
	// WHEN looking up books by POST with topic
	matched, _ = repo.LookupAll(&types.MockScenarioKeyData{
		Method:           types.Post,
		Path:             "/v3/mytopic/books",
		MatchQueryParams: map[string]string{"a": "1", "b": "abc"},
		MatchHeaders: map[string]string{
			types.ContentTypeHeader: "application/json",
			"ETag":                  "981",
		},
	})
	// THEN it should find it
	assert.Equal(t, 10, len(matched))
	for _, m := range matched {
		groups := m.MatchGroups("/v3/mytopic/books")
		assert.Equal(t, 1, len(groups))
		require.Equal(t, "mytopic", groups["topic"])
	}
	_, err = repo.Lookup(&types.MockScenarioKeyData{
		Method:           types.Post,
		Path:             "/v3/mytopic/books",
		MatchQueryParams: map[string]string{"a": "1", "b": "abc"},
		MatchHeaders: map[string]string{
			types.ContentTypeHeader: "application/json",
			"ETag":                  "981",
		},
	}, make(map[string]any))
	require.NoError(t, err)
}

func Test_ShouldLookupGetMockScenarios(t *testing.T) {
	// GIVEN a mock scenario repository
	repo, err := NewFileMockScenarioRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)
	// AND a set of mock scenarios
	for i := 0; i < 10; i++ {
		require.NoError(t, repo.Save(buildScenario(types.Get, fmt.Sprintf("todo_get_%d", i), "/api/todos/:id", i)))
		require.NoError(t, repo.Save(buildScenario(types.Get, fmt.Sprintf("book_get_%d", i), "/api/:topic/books/:id", i)))
	}
	// WHEN looking up scenarios with wrong method
	matched, _ := repo.LookupAll(&types.MockScenarioKeyData{
		Method:           types.Patch,
		Path:             "/api/todos/1",
		MatchQueryParams: map[string]string{"a": "1"},
		MatchHeaders: map[string]string{
			types.ContentTypeHeader: "application/json",
		},
	})
	// THEN it should not find it
	assert.Equal(t, 0, len(matched))

	// WHEN looking up scenarios with wrong query params
	matched, _ = repo.LookupAll(&types.MockScenarioKeyData{
		Method:           types.Get,
		Path:             "/api/todos/1",
		MatchQueryParams: map[string]string{"a": "11"},
		MatchHeaders: map[string]string{
			types.ContentTypeHeader: "application/json",
			"ETag":                  "981",
		},
	})
	assert.Equal(t, 0, len(matched))

	// WHEN looking up scenarios with valid params but without headers
	matched, _ = repo.LookupAll(&types.MockScenarioKeyData{
		Method:           types.Get,
		Path:             "/api/todos/2",
		MatchQueryParams: map[string]string{"a": "1"},
		MatchHeaders: map[string]string{
			types.ContentTypeHeader: "application/json",
		},
	})
	// THEN it should not match
	assert.Equal(t, 0, len(matched))

	// WHEN looking up scenarios with valid params and headers
	matched, _ = repo.LookupAll(&types.MockScenarioKeyData{
		Method:           types.Get,
		Path:             "/api/todos/2",
		MatchQueryParams: map[string]string{"a": "1", "b": "abc"},
		MatchHeaders: map[string]string{
			types.ContentTypeHeader: "application/json",
			"ETag":                  "981",
		},
	})
	assert.Equal(t, 10, len(matched))
	for i, m := range matched {
		groups := m.MatchGroups(fmt.Sprintf("/api/todos/%d", i))
		require.Equal(t, strconv.Itoa(i), groups["id"])
		assert.Equal(t, 1, len(groups))
	}

	//
	// WHEN looking up books by POST with topic
	matched, _ = repo.LookupAll(&types.MockScenarioKeyData{
		Method:           types.Get,
		Path:             "/api/mytopic/books/11",
		MatchQueryParams: map[string]string{"a": "1", "b": "abc"},
		MatchHeaders: map[string]string{
			types.ContentTypeHeader: "application/json",
			"ETag":                  "981",
		},
	})
	// THEN it should find it
	assert.Equal(t, 10, len(matched))
	for i, m := range matched {
		groups := m.MatchGroups(fmt.Sprintf("/api/mytopic/books/%d", i))
		require.Equal(t, strconv.Itoa(i), groups["id"])
		require.Equal(t, "mytopic", groups["topic"])
		assert.Equal(t, 2, len(groups))
	}
}

func Test_ShouldLookupDeleteMockScenarios(t *testing.T) {
	// GIVEN a mock scenario repository
	repo, err := NewFileMockScenarioRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)
	// AND a set of mock scenarios
	for i := 0; i < 10; i++ {
		require.NoError(t, repo.Save(buildScenario(types.Delete, fmt.Sprintf("todo_get_%d", i), "/v1/todos/:id", i)))
		require.NoError(t, repo.Save(buildScenario(types.Delete, fmt.Sprintf("book_get_%d", i), "/v1/:topic/books/:id", i)))
	}
	// WHEN looking up scenarios with wrong method
	matched, _ := repo.LookupAll(&types.MockScenarioKeyData{
		Method:           types.Patch,
		Path:             "/v1/todos/1",
		MatchQueryParams: map[string]string{"a": "1"},
		MatchHeaders: map[string]string{
			types.ContentTypeHeader: "application/json",
		},
	})
	// THEN it should not find it
	assert.Equal(t, 0, len(matched))

	// WHEN looking up scenarios with wrong query params
	matched, _ = repo.LookupAll(&types.MockScenarioKeyData{
		Method:           types.Delete,
		Path:             "/v1/todos/1",
		MatchQueryParams: map[string]string{"a": "11"},
		MatchHeaders: map[string]string{
			types.ContentTypeHeader: "application/json",
		},
	})
	assert.Equal(t, 0, len(matched))

	// WHEN looking up scenarios with valid params
	matched, _ = repo.LookupAll(&types.MockScenarioKeyData{
		Method:           types.Delete,
		Path:             "/v1/todos/2",
		MatchQueryParams: map[string]string{"a": "1", "b": "abc"},
		MatchHeaders: map[string]string{
			types.ContentTypeHeader: "application/json",
			"ETag":                  "981",
		},
	})
	assert.Equal(t, 10, len(matched))
	for i, m := range matched {
		groups := m.MatchGroups(fmt.Sprintf("/v1/todos/%d", i))
		require.Equal(t, strconv.Itoa(i), groups["id"])
		assert.Equal(t, 1, len(groups))
	}

	//
	// WHEN looking up books by POST with topic
	matched, _ = repo.LookupAll(&types.MockScenarioKeyData{
		Method:           types.Delete,
		Path:             "/v1/mytopic/books/11",
		MatchQueryParams: map[string]string{"a": "1", "b": "abc"},
		MatchHeaders: map[string]string{
			types.ContentTypeHeader: "application/json",
			"ETag":                  "981",
		},
	})
	// THEN it should find it
	assert.Equal(t, 10, len(matched))
	for i, m := range matched {
		groups := m.MatchGroups(fmt.Sprintf("/v1/mytopic/books/%d", i))
		require.Equal(t, strconv.Itoa(i), groups["id"])
		require.Equal(t, "mytopic", groups["topic"])
		assert.Equal(t, 2, len(groups))
	}
}

func Test_ShouldLookupPutMockScenariosWithPathVariables(t *testing.T) {
	// GIVEN a mock scenario repository
	repo, err := NewFileMockScenarioRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)
	// AND a set of mock scenarios
	for i := 0; i < 10; i++ {
		require.NoError(t, repo.Save(buildScenario(types.Put, fmt.Sprintf("todo_put_%d", i), "/api/todos/{id}", i)))
		require.NoError(t, repo.Save(buildScenario(types.Put, fmt.Sprintf("book_put_%d", i), "/api/{topic}/books/{id}", i)))
	}
	// WHEN looking up todos by POST without criteria
	matched, _ := repo.LookupAll(&types.MockScenarioKeyData{})
	// THEN it should not find it
	assert.Equal(t, 0, len(matched))

	// WHEN looking up todos by PUT with different query param
	matched, _ = repo.LookupAll(&types.MockScenarioKeyData{
		Method:           types.Put,
		Path:             "/api/todos/11",
		MatchQueryParams: map[string]string{"a": "11"},
		MatchHeaders: map[string]string{
			types.ContentTypeHeader: "application/json",
		},
	})
	// THEN it should not find it
	assert.Equal(t, 0, len(matched))

	// WHEN looking up todos by matching path and query params
	matched, _ = repo.LookupAll(&types.MockScenarioKeyData{
		Method:           types.Put,
		Path:             "/api/todos/12",
		MatchQueryParams: map[string]string{"a": "1", "b": "abc"},
		MatchHeaders: map[string]string{
			types.ContentTypeHeader: "application/json",
			"ETag":                  "981",
		},
	})
	// THEN it should find it
	assert.Equal(t, 10, len(matched))
	for i, m := range matched {
		groups := m.MatchGroups(fmt.Sprintf("/api/todos/%d", i+100))
		require.Equal(t, strconv.Itoa(i+100), groups["id"])
		assert.Equal(t, 1, len(groups))
	}
	_, err = repo.Lookup(&types.MockScenarioKeyData{
		Method:           types.Put,
		Path:             "/api/todos/12",
		MatchQueryParams: map[string]string{"a": "1", "b": "abc"},
		MatchHeaders: map[string]string{
			types.ContentTypeHeader: "application/json",
			"ETag":                  "981",
		},
	}, make(map[string]any))
	require.NoError(t, err)

	//
	// WHEN looking up books by POST with topic
	matched, _ = repo.LookupAll(&types.MockScenarioKeyData{
		Method:           types.Put,
		Path:             "/api/mytopic/books/11",
		MatchQueryParams: map[string]string{"a": "1", "b": "abc"},
		MatchHeaders: map[string]string{
			types.ContentTypeHeader: "application/json",
			"ETag":                  "981",
		},
	})
	// THEN it should find it
	assert.Equal(t, 10, len(matched))
	for i, m := range matched {
		groups := m.MatchGroups(fmt.Sprintf("/api/mytopic/books/%d", i))
		require.Equal(t, strconv.Itoa(i), groups["id"])
		require.Equal(t, "mytopic", groups["topic"])
		assert.Equal(t, 2, len(groups))
	}
	_, err = repo.Lookup(&types.MockScenarioKeyData{
		Method:           types.Put,
		Path:             "/api/mytopic/books/11",
		MatchQueryParams: map[string]string{"a": "1", "b": "abc"},
		MatchHeaders: map[string]string{
			types.ContentTypeHeader: "application/json",
		},
	}, make(map[string]any))
	require.Error(t, err)
	_, err = repo.Lookup(&types.MockScenarioKeyData{
		Method:           types.Put,
		Path:             "/api/mytopic/books/11",
		MatchQueryParams: map[string]string{"a": "1", "b": "abc"},
		MatchHeaders: map[string]string{
			types.ContentTypeHeader: "application/json",
			"ETag":                  "981",
		},
	}, make(map[string]any))
	require.NoError(t, err)
}

func buildScenario(method types.MethodType, name string, path string, n int) *types.MockScenario {
	return &types.MockScenario{
		Method:      method,
		Name:        name,
		Path:        path,
		Group:       path,
		Description: name,
		Request: types.MockHTTPRequest{
			MatchQueryParams: map[string]string{"a": `\d+`, "b": "abc"},
			MatchHeaders: map[string]string{
				types.ContentTypeHeader: "application/json",
				"ETag":                  `\d{3}`,
			},
		},
		Response: types.MockHTTPResponse{
			Headers: map[string][]string{
				"ETag":                  {strconv.Itoa(n)},
				types.ContentTypeHeader: {"application/json"},
			},
			Contents:   "test body",
			StatusCode: 200,
		},
		WaitBeforeReply: time.Duration(1) * time.Second,
	}
}
