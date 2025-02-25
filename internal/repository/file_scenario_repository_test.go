package repository

import (
	"bytes"
	"fmt"
	"github.com/bhatti/api-mock-service/internal/utils"
	"gopkg.in/yaml.v3"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/bhatti/api-mock-service/internal/types"
	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"
)

const apiPath = "//abc//\\def/123/"

func Test_ShouldRawSaveAndLoadMockScenarios(t *testing.T) {
	// GIVEN a mock scenario repository
	repo, err := NewFileAPIScenarioRepository(types.BuildTestConfig())
	require.NoError(t, err)
	// AND mock scenario
	scenario := types.BuildTestScenario(types.Post, "test1", apiPath, 10)
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
	repo, err := NewFileAPIScenarioRepository(types.BuildTestConfig())
	require.NoError(t, err)
	// AND mock scenario
	scenario := types.BuildTestScenario(types.Post, "test1", apiPath, 10)
	// WHEN saving scenario
	err = repo.Save(scenario)
	// THEN it should succeed
	require.NoError(t, err)

	// AND should return saved scenario
	saved, err := repo.Lookup(scenario.ToKeyData(), nil)
	require.NoError(t, err)
	require.NoError(t, scenario.ToKeyData().Equals(saved.ToKeyData()))

	t.Run("Save Variables", func(t *testing.T) {
		apiVars := &types.APIVariables{
			Name:      "common-test-name",
			Variables: map[string]string{"gk1": "v1", "gk2": "v2"},
		}

		scenario.VariablesFile = apiVars.Name
		err = repo.Save(scenario)
		require.NoError(t, err)

		err = repo.SaveVariables(apiVars)
		require.NoError(t, err)

		// AND should return saved scenario
		loaded, err := repo.Lookup(scenario.ToKeyData(), nil)
		require.NoError(t, err)
		require.Equal(t, "v1", loaded.Request.Variables["gk1"])
		require.Equal(t, "v2", loaded.Request.Variables["gk2"])
	})
}

func Test_ShouldNotGetAfterDeletingMockScenarios(t *testing.T) {
	// GIVEN a mock scenario repository
	repo, err := NewFileAPIScenarioRepository(types.BuildTestConfig())
	require.NoError(t, err)
	// AND mock scenario
	scenario := types.BuildTestScenario(types.Post, "test1", apiPath, 20)
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

func Test_ShouldListMockScenariosGroups(t *testing.T) {
	// GIVEN a mock scenario repository
	repo, err := NewFileAPIScenarioRepository(types.BuildTestConfig())
	require.NoError(t, err)
	// AND a set of mock scenarios
	for i := 0; i < 10; i++ {
		scenario := types.BuildTestScenario(types.Post, fmt.Sprintf("test_%d", i), apiPath, 30)
		scenario.Group = fmt.Sprintf("test-group-%v", i%2 == 0)
		err = repo.Save(scenario)
		require.NoError(t, err)
	}
	// WHEN listing scenario groups
	groups := repo.GetGroups()
	require.True(t, len(groups) >= 2)
}

func Test_ShouldListMockScenariosByPath(t *testing.T) {
	// GIVEN a mock scenario repository
	repo, err := NewFileAPIScenarioRepository(types.BuildTestConfig())
	require.NoError(t, err)
	// AND a set of mock scenarios
	for i := 0; i < 10; i++ {
		scenario := types.BuildTestScenario(types.Post, fmt.Sprintf("test_%d", i), apiPath, 30)
		scenario.Group = fmt.Sprintf("test-group-%v", i%2 == 0)
		scenario.Path = fmt.Sprintf("/api/v1/%v", i%2 == 0)
		err = repo.Save(scenario)
		require.NoError(t, err)
	}
	// WHEN listing scenarios by path
	scenarios := repo.LookupAllByPath("api/v1/true")
	// THEN it should return matching scenarios
	require.Equal(t, 5, len(scenarios))
}

func Test_ShouldListMockScenariosNames(t *testing.T) {
	// GIVEN a mock scenario repository
	repo, err := NewFileAPIScenarioRepository(types.BuildTestConfig())
	require.NoError(t, err)
	// AND a set of mock scenarios
	for i := 0; i < 10; i++ {
		scenario := types.BuildTestScenario(types.Post, fmt.Sprintf("test_%d", i), apiPath, 30)
		err = repo.Save(scenario)
		require.NoError(t, err)
	}
	// WHEN listing scenarios
	names, err := repo.GetScenariosNames(types.Post, apiPath)
	require.NoError(t, err)
	for i := 0; i < 10; i++ {
		require.Equal(t, fmt.Sprintf("test_%d", i), names[i])
		err = repo.Delete(types.Post, names[i], apiPath)
		require.NoError(t, err)
	}
}

func Test_ShouldLookupAllByGroup(t *testing.T) {
	// GIVEN a mock scenario repository
	repo, err := NewFileAPIScenarioRepository(types.BuildTestConfig())
	require.NoError(t, err)
	// AND a set of mock scenarios
	for i := 0; i < 10; i++ {
		require.NoError(t, repo.Save(types.BuildTestScenario(types.Get, fmt.Sprintf("todo_%d", i), "/v2/todos", i)))
		require.NoError(t, repo.Save(types.BuildTestScenario(types.Get, fmt.Sprintf("book_%d", i), "/v2/books", i)))
		require.NoError(t, repo.Save(types.BuildTestScenario(types.Post, fmt.Sprintf("todo_%d", i), "/v2/todos", i)))
		require.NoError(t, repo.Save(types.BuildTestScenario(types.Post, fmt.Sprintf("book_%d", i), "/v2/books", i)))
		require.NoError(t, repo.Save(types.BuildTestScenario(types.Put, fmt.Sprintf("todo_%d", i), "/v2/todos", i)))
		require.NoError(t, repo.Save(types.BuildTestScenario(types.Put, fmt.Sprintf("book_%d", i), "/v2/books", i)))
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
	repo, err := NewFileAPIScenarioRepository(types.BuildTestConfig())
	require.NoError(t, err)
	// AND a set of mock scenarios
	for i := 0; i < 10; i++ {
		require.NoError(t, repo.Save(types.BuildTestScenario(types.Put, fmt.Sprintf("todo_put_%d", i), "/api/todos/:id", i)))
		require.NoError(t, repo.Save(types.BuildTestScenario(types.Put, fmt.Sprintf("book_put_%d", i), "/api/:topic/books/:id", i)))
	}
	// WHEN looking up todos by POST without criteria
	matched, _, _, _ := repo.LookupAll(&types.APIKeyData{})
	// THEN it should not find it
	assert.Equal(t, 0, len(matched))

	// WHEN looking up todos by PUT with different query param
	matched, _, _, _ = repo.LookupAll(&types.APIKeyData{
		Method:                   types.Put,
		Path:                     "/api/todos/11",
		AssertQueryParamsPattern: map[string]string{"a": "22"},
		AssertHeadersPattern: map[string]string{
			types.ContentTypeHeader: "application/json",
		},
	})
	// THEN it should not find it
	assert.Equal(t, 0, len(matched))

	// WHEN looking up todos by matching path and query params
	matched, _, _, _ = repo.LookupAll(&types.APIKeyData{
		Method:                   types.Put,
		Path:                     "/api/todos/12",
		AssertQueryParamsPattern: map[string]string{"a": "1"},
		AssertHeadersPattern: map[string]string{
			types.ContentTypeHeader: "application/json",
		},
	})
	// THEN it should not find it without headers
	assert.Equal(t, 0, len(matched))

	// WHEN looking up todos by matching path, headers and query params
	matched, _, _, _ = repo.LookupAll(&types.APIKeyData{
		Method:                   types.Put,
		Path:                     "/api/todos/12",
		AssertQueryParamsPattern: map[string]string{"a": "1", "b": "abc"},
		AssertHeadersPattern: map[string]string{
			types.ContentTypeHeader: "application/json",
			types.ETagHeader:        "981",
		},
	})
	// THEN it should find it
	assert.Equal(t, 10, len(matched))
	for i, m := range matched {
		groups := m.MatchGroups(fmt.Sprintf("/api/todos/%d", i))
		require.Equal(t, strconv.Itoa(i), groups["id"])
		assert.Equal(t, 1, len(groups))
	}
	_, err = repo.Lookup(&types.APIKeyData{
		Method:                   types.Put,
		Path:                     "/api/todos/12",
		AssertQueryParamsPattern: map[string]string{"a": "1", "b": "abc", "n": "0"},
		AssertHeadersPattern: map[string]string{
			types.ContentTypeHeader: "application/json",
			types.ETagHeader:        "981",
		},
	}, make(map[string]any))
	require.NoError(t, err)

	//
	// WHEN looking up books by POST with topic
	matched, _, _, _ = repo.LookupAll(&types.APIKeyData{
		Method:                   types.Put,
		Path:                     "/api/mytopic/books/11",
		AssertQueryParamsPattern: map[string]string{"a": "1", "b": "abc", "n": "0"},
		AssertHeadersPattern: map[string]string{
			types.ContentTypeHeader: "application/json",
			types.ETagHeader:        "981",
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
	_, err = repo.Lookup(&types.APIKeyData{
		Method:                   types.Put,
		Path:                     "/api/mytopic/books/11",
		AssertQueryParamsPattern: map[string]string{"a": "1", "b": "abc", "n": "0"},
		AssertHeadersPattern: map[string]string{
			types.ContentTypeHeader: "application/json",
			types.ETagHeader:        "981",
		},
	}, make(map[string]any))
	require.NoError(t, err)
}

func Test_ShouldListMockScenarios(t *testing.T) {
	// GIVEN a mock scenario repository
	repo, err := NewFileAPIScenarioRepository(types.BuildTestConfig())
	require.NoError(t, err)
	// AND a set of mock scenarios
	for i := 0; i < 10; i++ {
		require.NoError(t, repo.Save(types.BuildTestScenario(types.Get, fmt.Sprintf("todo_post_%d", i), "/v3/api/todos", i)))
		require.NoError(t, repo.Save(types.BuildTestScenario(types.Post, fmt.Sprintf("todo_post_%d", i), "/v3/api/todos", i)))
		require.NoError(t, repo.Save(types.BuildTestScenario(types.Delete, fmt.Sprintf("todo_post_%d", i), "/v3/api/todos", i)))
		require.NoError(t, repo.Save(types.BuildTestScenario(types.Get, fmt.Sprintf("book_post_%d", i), "/v3/api/books", i)))
		require.NoError(t, repo.Save(types.BuildTestScenario(types.Post, fmt.Sprintf("book_post_%d", i), "/v3/api/books", i)))
		require.NoError(t, repo.Save(types.BuildTestScenario(types.Delete, fmt.Sprintf("book_post_%d", i), "/v3/api/books", i)))
	}
	// WHEN listing mock scenario
	all := repo.ListScenarioKeyData("Twitter API v2_V2.21")
	// THEN it should succeed
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
	repo, err := NewFileAPIScenarioRepository(types.BuildTestConfig())
	require.NoError(t, err)
	// AND a set of mock scenarios
	for i := 0; i < 10; i++ {
		require.NoError(t, repo.Save(types.BuildTestScenario(types.Post, fmt.Sprintf("todo_post_%d", i), "/v3/todos", i)))
		require.NoError(t, repo.Save(types.BuildTestScenario(types.Post, fmt.Sprintf("book_post_%d", i), "/v3/:topic/books", i)))
	}
	// WHEN looking up todos by POST without criteria
	matched, _, _, _ := repo.LookupAll(&types.APIKeyData{})
	// THEN it should not find it
	assert.Equal(t, 0, len(matched))

	// WHEN looking up todos by POST with different query param
	matched, _, _, _ = repo.LookupAll(&types.APIKeyData{
		Method:                   types.Post,
		Path:                     "/v3/todos",
		AssertQueryParamsPattern: map[string]string{"a": "11"},
		AssertHeadersPattern: map[string]string{
			types.ContentTypeHeader: "application/json",
		},
	})
	// THEN it should not find it
	assert.Equal(t, 0, len(matched))

	// WHEN looking up todos by matching path and query params
	matched, _, _, _ = repo.LookupAll(&types.APIKeyData{
		Method:                   types.Post,
		Path:                     "/v3/todos",
		AssertQueryParamsPattern: map[string]string{"a": "1"},
		AssertHeadersPattern: map[string]string{
			types.ContentTypeHeader: "application/json",
		},
	})
	// THEN it should not find it
	assert.Equal(t, 0, len(matched))
	// WHEN looking up todos by matching path, headers and query params
	matched, _, _, _ = repo.LookupAll(&types.APIKeyData{
		Method:                   types.Post,
		Path:                     "/v3/todos",
		AssertQueryParamsPattern: map[string]string{"a": "1", "b": "abc"},
		AssertHeadersPattern: map[string]string{
			types.ContentTypeHeader: "application/json",
			types.ETagHeader:        "981",
		},
	})
	// THEN it should find it
	assert.Equal(t, 10, len(matched))
	for _, m := range matched {
		assert.Equal(t, 0, len(m.MatchGroups("/v3/todos")))
	}

	//
	// WHEN looking up books by POST with topic
	matched, _, _, _ = repo.LookupAll(&types.APIKeyData{
		Method:                   types.Post,
		Path:                     "/v3/mytopic/books",
		AssertQueryParamsPattern: map[string]string{"a": "1", "b": "abc"},
		AssertHeadersPattern: map[string]string{
			types.ContentTypeHeader: "application/json",
			types.ETagHeader:        "981",
		},
	})
	// THEN it should find it
	assert.Equal(t, 10, len(matched))
	for _, m := range matched {
		groups := m.MatchGroups("/v3/mytopic/books")
		assert.Equal(t, 1, len(groups))
		require.Equal(t, "mytopic", groups["topic"])
	}
	_, err = repo.Lookup(&types.APIKeyData{
		Method:                   types.Post,
		Path:                     "/v3/mytopic/books",
		AssertQueryParamsPattern: map[string]string{"a": "1", "b": "abc"},
		AssertHeadersPattern: map[string]string{
			types.ContentTypeHeader: "application/json",
			types.ETagHeader:        "981",
		},
	}, make(map[string]any))
	require.NoError(t, err)
}

func Test_ShouldLookupGetMockScenarios(t *testing.T) {
	// GIVEN a mock scenario repository
	repo, err := NewFileAPIScenarioRepository(types.BuildTestConfig())
	require.NoError(t, err)
	// AND a set of mock scenarios
	for i := 0; i < 10; i++ {
		require.NoError(t, repo.Save(types.BuildTestScenario(types.Get, fmt.Sprintf("todo_get_%d", i), "/api/todos/:id", i)))
		require.NoError(t, repo.Save(types.BuildTestScenario(types.Get, fmt.Sprintf("book_get_%d", i), "/api/:topic/books/:id", i)))
	}
	// WHEN looking up scenarios with wrong method
	matched, _, _, _ := repo.LookupAll(&types.APIKeyData{
		Method:                   types.Patch,
		Path:                     "/api/todos/1",
		AssertQueryParamsPattern: map[string]string{"a": "1"},
		AssertHeadersPattern: map[string]string{
			types.ContentTypeHeader: "application/json",
		},
	})
	// THEN it should not find it
	assert.Equal(t, 0, len(matched))

	// WHEN looking up scenarios with wrong query params
	matched, _, _, _ = repo.LookupAll(&types.APIKeyData{
		Method:                   types.Get,
		Path:                     "/api/todos/1",
		AssertQueryParamsPattern: map[string]string{"a": "11"},
		AssertHeadersPattern: map[string]string{
			types.ContentTypeHeader: "application/json",
			types.ETagHeader:        "981",
		},
	})
	assert.Equal(t, 0, len(matched))

	// WHEN looking up scenarios with valid params but without headers
	matched, _, _, _ = repo.LookupAll(&types.APIKeyData{
		Method:                   types.Get,
		Path:                     "/api/todos/2",
		AssertQueryParamsPattern: map[string]string{"a": "1"},
		AssertHeadersPattern: map[string]string{
			types.ContentTypeHeader: "application/json",
		},
	})
	// THEN it should not match
	assert.Equal(t, 0, len(matched))

	// WHEN looking up scenarios with valid params and headers
	matched, _, _, _ = repo.LookupAll(&types.APIKeyData{
		Method:                   types.Get,
		Path:                     "/api/todos/2",
		AssertQueryParamsPattern: map[string]string{"a": "1", "b": "abc"},
		AssertHeadersPattern: map[string]string{
			types.ContentTypeHeader: "application/json",
			types.ETagHeader:        "981",
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
	matched, _, _, _ = repo.LookupAll(&types.APIKeyData{
		Method:                   types.Get,
		Path:                     "/api/mytopic/books/11",
		AssertQueryParamsPattern: map[string]string{"a": "1", "b": "abc"},
		AssertHeadersPattern: map[string]string{
			types.ContentTypeHeader: "application/json",
			types.ETagHeader:        "981",
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
	repo, err := NewFileAPIScenarioRepository(types.BuildTestConfig())
	require.NoError(t, err)
	// AND a set of mock scenarios
	for i := 0; i < 10; i++ {
		require.NoError(t, repo.Save(types.BuildTestScenario(types.Delete, fmt.Sprintf("todo_get_%d", i), "/v1/todos/:id", i)))
		require.NoError(t, repo.Save(types.BuildTestScenario(types.Delete, fmt.Sprintf("book_get_%d", i), "/v1/:topic/books/:id", i)))
	}
	// WHEN looking up scenarios with wrong method
	matched, _, _, _ := repo.LookupAll(&types.APIKeyData{
		Method:                   types.Patch,
		Path:                     "/v1/todos/1",
		AssertQueryParamsPattern: map[string]string{"a": "1"},
		AssertHeadersPattern: map[string]string{
			types.ContentTypeHeader: "application/json",
		},
	})
	// THEN it should not find it
	assert.Equal(t, 0, len(matched))

	// WHEN looking up scenarios with wrong query params
	matched, _, _, _ = repo.LookupAll(&types.APIKeyData{
		Method:                   types.Delete,
		Path:                     "/v1/todos/1",
		AssertQueryParamsPattern: map[string]string{"a": "11"},
		AssertHeadersPattern: map[string]string{
			types.ContentTypeHeader: "application/json",
		},
	})
	assert.Equal(t, 0, len(matched))

	// WHEN looking up scenarios with valid params
	matched, _, _, _ = repo.LookupAll(&types.APIKeyData{
		Method:                   types.Delete,
		Path:                     "/v1/todos/2",
		AssertQueryParamsPattern: map[string]string{"a": "1", "b": "abc"},
		AssertHeadersPattern: map[string]string{
			types.ContentTypeHeader: "application/json",
			types.ETagHeader:        "981",
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
	matched, _, _, _ = repo.LookupAll(&types.APIKeyData{
		Method:                   types.Delete,
		Path:                     "/v1/mytopic/books/11",
		AssertQueryParamsPattern: map[string]string{"a": "1", "b": "abc"},
		AssertHeadersPattern: map[string]string{
			types.ContentTypeHeader: "application/json",
			types.ETagHeader:        "981",
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
	repo, err := NewFileAPIScenarioRepository(types.BuildTestConfig())
	require.NoError(t, err)
	// AND a set of mock scenarios
	for i := 0; i < 10; i++ {
		require.NoError(t, repo.Save(types.BuildTestScenario(types.Put, fmt.Sprintf("todo_put_%d", i), "/api/todos/{id}", i)))
		require.NoError(t, repo.Save(types.BuildTestScenario(types.Put, fmt.Sprintf("book_put_%d", i), "/api/{topic}/books/{id}", i)))
	}
	// WHEN looking up todos by POST without criteria
	matched, _, _, _ := repo.LookupAll(&types.APIKeyData{})
	// THEN it should not find it
	assert.Equal(t, 0, len(matched))

	// WHEN looking up todos by PUT with different query param
	matched, _, _, _ = repo.LookupAll(&types.APIKeyData{
		Method:                   types.Put,
		Path:                     "/api/todos/11",
		AssertQueryParamsPattern: map[string]string{"a": "11"},
		AssertHeadersPattern: map[string]string{
			types.ContentTypeHeader: "application/json",
		},
	})
	// THEN it should not find it
	assert.Equal(t, 0, len(matched))

	// WHEN looking up todos by matching path and query params
	matched, _, _, _ = repo.LookupAll(&types.APIKeyData{
		Method:                   types.Put,
		Path:                     "/api/todos/12",
		AssertQueryParamsPattern: map[string]string{"a": "1", "b": "abc"},
		AssertHeadersPattern: map[string]string{
			types.ContentTypeHeader: "application/json",
			types.ETagHeader:        "981",
		},
	})
	// THEN it should find it
	assert.Equal(t, 10, len(matched))
	for i, m := range matched {
		groups := m.MatchGroups(fmt.Sprintf("/api/todos/%d", i+100))
		require.Equal(t, strconv.Itoa(i+100), groups["id"])
		assert.Equal(t, 1, len(groups))
	}
	_, err = repo.Lookup(&types.APIKeyData{
		Method:                   types.Put,
		Path:                     "/api/todos/12",
		AssertQueryParamsPattern: map[string]string{"a": "1", "b": "abc"},
		AssertHeadersPattern: map[string]string{
			types.ContentTypeHeader: "application/json",
			types.ETagHeader:        "981",
		},
	}, make(map[string]any))
	require.NoError(t, err)

	//
	// WHEN looking up books by POST with topic
	matched, _, _, _ = repo.LookupAll(&types.APIKeyData{
		Method:                   types.Put,
		Path:                     "/api/mytopic/books/11",
		AssertQueryParamsPattern: map[string]string{"a": "1", "b": "abc"},
		AssertHeadersPattern: map[string]string{
			types.ContentTypeHeader: "application/json",
			types.ETagHeader:        "981",
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
	_, err = repo.Lookup(&types.APIKeyData{
		Method:                   types.Put,
		Path:                     "/api/mytopic/books/11",
		AssertQueryParamsPattern: map[string]string{"a": "1", "b": "abc"},
		AssertHeadersPattern: map[string]string{
			types.ContentTypeHeader: "application/json",
		},
	}, make(map[string]any))
	require.Error(t, err)
	_, err = repo.Lookup(&types.APIKeyData{
		Method:                   types.Put,
		Path:                     "/api/mytopic/books/11",
		AssertQueryParamsPattern: map[string]string{"a": "1", "b": "abc"},
		AssertHeadersPattern: map[string]string{
			types.ContentTypeHeader: "application/json",
			types.ETagHeader:        "981",
		},
	}, make(map[string]any))
	require.NoError(t, err)
}

func Test_ShouldLoadSaveScenariosHistory(t *testing.T) {
	// GIVEN a mock scenario repository
	config := types.BuildTestConfig()
	repo, err := NewFileAPIScenarioRepository(config)
	require.NoError(t, err)
	// AND a set of mock scenarios
	started := time.Now()
	groups := []string{
		fmt.Sprintf("%d-save-history-group-first", started.UnixMilli()),
		fmt.Sprintf("%d-save-history-group-second", started.UnixMilli()),
	}
	u, err := url.Parse("https://localhost:8080")
	require.NoError(t, err)
	for i := 0; i < 200; i++ {
		scenario := types.BuildTestScenario(types.Post, fmt.Sprintf("save-history_%d", i), "/path", i)
		scenario.Group = groups[i%2]
		scenario.Path = fmt.Sprintf("/v123/api/%s/%d", apiPath, i)
		scenario.Request.QueryParams["i"] = fmt.Sprintf("%d", started.UnixMilli()+int64(i))
		scenario.Request.Headers["I"] = fmt.Sprintf("%d", started.UnixMilli()+int64(i))
		err = repo.SaveHistory(scenario, u.String(), time.Now(), time.Now().Add(time.Second))
		require.NoError(t, err)
	}
	names := repo.HistoryNames("unknown")
	require.Equal(t, 0, len(names))
	names = repo.HistoryNames("")
	require.True(t, len(names) >= 200)
	names = repo.HistoryNames(groups[0])
	require.Equal(t, 100, len(names))
	names = repo.HistoryNames(groups[1])
	require.Equal(t, 100, len(names))
	for _, name := range names {
		scenarios, err := repo.LoadHistory(name, "", 0, 100)
		require.NoError(t, err)
		for _, scenario := range scenarios {
			require.Contains(t, name, scenario.Group)
			loaded, err := repo.loadHistoryByName(name)
			require.NoError(t, err)
			require.Contains(t, loaded.Group, scenario.Group)
		}
	}
	for i := 0; i < 6; i++ {
		scenarios, err := repo.LoadHistory("", groups[0], i, 20)
		require.NoError(t, err)
		if i == 5 {
			require.Equal(t, 0, len(scenarios))
		} else {
			require.Equal(t, 20, len(scenarios), fmt.Sprintf("page %d", i))
		}
	}
}

func Test_ShouldLoadSaveScenariosHistoryWithLimit(t *testing.T) {
	// GIVEN a mock scenario repository
	config := types.BuildTestConfig()
	config.MaxHistory = 10
	repo, err := NewFileAPIScenarioRepository(config)
	require.NoError(t, err)
	// AND a set of mock scenarios
	started := time.Now()
	groups := []string{
		fmt.Sprintf("%d-save-history-group-first", started.UnixMilli()),
		fmt.Sprintf("%d-save-history-group-second", started.UnixMilli()),
	}
	u, err := url.Parse("https://localhost:8080")
	require.NoError(t, err)
	for i := 0; i < 200; i++ {
		scenario := types.BuildTestScenario(types.Post, fmt.Sprintf("save-history_%d", i), "/path", i)
		scenario.Group = groups[i%2]
		scenario.Path = fmt.Sprintf("/v789/api/%s/%d", apiPath, i)
		scenario.Request.QueryParams["i"] = fmt.Sprintf("%d", started.UnixMilli()+int64(i))
		scenario.Request.Headers["I"] = fmt.Sprintf("%d", started.UnixMilli()+int64(i))
		err = repo.SaveHistory(scenario, u.String(), time.Now(), time.Now().Add(time.Second))
		require.NoError(t, err)
	}
	names := repo.HistoryNames("unknown")
	require.Equal(t, 0, len(names))
	names = repo.HistoryNames("")
	require.True(t, len(names) >= 10)
	names = repo.HistoryNames(groups[0])
	require.Equal(t, 5, len(names))
	names = repo.HistoryNames(groups[1])
	require.Equal(t, 5, len(names))
	scenarios, err := repo.LoadHistory("", "", 0, 100)
	require.NoError(t, err)
	require.Equal(t, 10, len(scenarios)) // max 10
	for i := 0; i < 2; i++ {
		scenarios, err := repo.LoadHistory("", groups[0], i, 20)
		require.NoError(t, err)
		if i == 1 {
			require.Equal(t, 0, len(scenarios))
		} else {
			require.Equal(t, 5, len(scenarios), fmt.Sprintf("page %d", i))
		}
	}
}

func Test_ShouldConvertHar(t *testing.T) {
	config := types.BuildTestConfig()
	// GIVEN repository for mock scenario
	mockScenarioRepository, err := NewFileAPIScenarioRepository(config)
	require.NoError(t, err)
	// AND a valid scenario
	scenario := types.BuildTestScenario(types.Post, "test-name", "/path", 0)
	scenario.Group = "archive-group"
	scenario.Request.Headers = map[string]string{
		types.ContentTypeHeader: "application/json 1.1",
	}
	scenario.Request.QueryParams = map[string]string{
		"abc": "123",
	}
	err = mockScenarioRepository.Save(scenario)
	require.NoError(t, err)
	u, err := url.Parse("https://localhost:8080" + scenario.Path)
	require.NoError(t, err)
	err = mockScenarioRepository.SaveHistory(scenario, u.String(), time.Now(), time.Now().Add(time.Second))
	require.NoError(t, err)
}
