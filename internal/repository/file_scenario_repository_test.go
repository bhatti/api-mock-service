package repository

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/bhatti/api-mock-service/internal/types"
	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"
)

const mockPath = "//abc//\\def/123/"

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
	saved, err := repo.Get(scenario.Method, scenario.Name, scenario.Path, nil)
	require.NoError(t, err)
	require.NoError(t, scenario.Equals(saved))
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
	saved, err := repo.Get(scenario.Method, scenario.Name, scenario.Path, nil)
	require.NoError(t, err)
	require.NoError(t, scenario.Equals(saved))

	// But WHEN DELETING the mock scenario
	err = repo.Delete(scenario.Method, scenario.Name, scenario.Path)
	require.NoError(t, err)

	// THEN GET should fail
	_, err = repo.Get(scenario.Method, scenario.Name, scenario.Path, nil)
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
	matched := repo.LookupAll(&types.MockScenarioKeyData{})
	// THEN it should not find it
	assert.Equal(t, 0, len(matched))

	// WHEN looking up todos by PUT with different query param
	matched = repo.LookupAll(&types.MockScenarioKeyData{
		Method:      types.Put,
		Path:        "/api/todos/11",
		QueryParams: "a=11",
		ContentType: "application",
	})
	// THEN it should not find it
	assert.Equal(t, 0, len(matched))

	// WHEN looking up todos by matching path and query params
	matched = repo.LookupAll(&types.MockScenarioKeyData{
		Method:      types.Put,
		Path:        "/api/todos/12",
		QueryParams: "a=1",
		ContentType: "application",
	})
	// THEN it should find it
	assert.Equal(t, 10, len(matched))
	for i, m := range matched {
		groups := m.MatchGroups(fmt.Sprintf("/api/todos/%d", i))
		require.Equal(t, strconv.Itoa(i), groups["id"])
		assert.Equal(t, 1, len(groups))
	}
	_, err = repo.Lookup(&types.MockScenarioKeyData{
		Method:      types.Put,
		Path:        "/api/todos/12",
		QueryParams: "a=1",
		ContentType: "application",
	})
	require.NoError(t, err)

	//
	// WHEN looking up books by POST with topic
	matched = repo.LookupAll(&types.MockScenarioKeyData{
		Method:      types.Put,
		Path:        "/api/mytopic/books/11",
		QueryParams: "a=1",
		ContentType: "application",
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
		Method:      types.Put,
		Path:        "/api/mytopic/books/11",
		QueryParams: "a=1",
		ContentType: "application",
	})
	require.NoError(t, err)
}

func Test_ShouldLookupPostMockScenarios(t *testing.T) {
	// GIVEN a mock scenario repository
	repo, err := NewFileMockScenarioRepository(&types.Configuration{DataDir: "../../mock_tests"})
	require.NoError(t, err)
	// AND a set of mock scenarios
	for i := 0; i < 10; i++ {
		require.NoError(t, repo.Save(buildScenario(types.Post, fmt.Sprintf("todo_post_%d", i), "/api/todos", i)))
		require.NoError(t, repo.Save(buildScenario(types.Post, fmt.Sprintf("book_post_%d", i), "/api/:topic/books", i)))
	}
	// WHEN looking up todos by POST without criteria
	matched := repo.LookupAll(&types.MockScenarioKeyData{})
	// THEN it should not find it
	assert.Equal(t, 0, len(matched))

	// WHEN looking up todos by POST with different query param
	matched = repo.LookupAll(&types.MockScenarioKeyData{
		Method:      types.Post,
		Path:        "/api/todos",
		QueryParams: "a=11",
		ContentType: "application",
	})
	// THEN it should not find it
	assert.Equal(t, 0, len(matched))

	// WHEN looking up todos by matching path and query params
	matched = repo.LookupAll(&types.MockScenarioKeyData{
		Method:      types.Post,
		Path:        "/api/todos",
		QueryParams: "a=1",
		ContentType: "application",
	})
	// THEN it should find it
	assert.Equal(t, 10, len(matched))
	for _, m := range matched {
		assert.Equal(t, 0, len(m.MatchGroups("/api/todos")))
	}

	//
	// WHEN looking up books by POST with topic
	matched = repo.LookupAll(&types.MockScenarioKeyData{
		Method:      types.Post,
		Path:        "/api/mytopic/books",
		QueryParams: "a=1",
		ContentType: "application",
	})
	// THEN it should find it
	assert.Equal(t, 10, len(matched))
	for _, m := range matched {
		groups := m.MatchGroups("/api/mytopic/books")
		assert.Equal(t, 1, len(groups))
		require.Equal(t, "mytopic", groups["topic"])
	}
	_, err = repo.Lookup(&types.MockScenarioKeyData{
		Method:      types.Post,
		Path:        "/api/mytopic/books",
		QueryParams: "a=1",
		ContentType: "application",
	})
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
	matched := repo.LookupAll(&types.MockScenarioKeyData{
		Method:      types.Patch,
		Path:        "/api/todos/1",
		QueryParams: "a=1",
		ContentType: "application",
	})
	// THEN it should not find it
	assert.Equal(t, 0, len(matched))

	// WHEN looking up scenarios with wrong query params
	matched = repo.LookupAll(&types.MockScenarioKeyData{
		Method:      types.Get,
		Path:        "/api/todos/1",
		QueryParams: "a=11",
		ContentType: "application",
	})
	assert.Equal(t, 0, len(matched))

	// WHEN looking up scenarios with valid params
	matched = repo.LookupAll(&types.MockScenarioKeyData{
		Method:      types.Get,
		Path:        "/api/todos/2",
		QueryParams: "a=1",
		ContentType: "application",
	})
	assert.Equal(t, 10, len(matched))
	for i, m := range matched {
		groups := m.MatchGroups(fmt.Sprintf("/api/todos/%d", i))
		require.Equal(t, strconv.Itoa(i), groups["id"])
		assert.Equal(t, 1, len(groups))
	}

	//
	// WHEN looking up books by POST with topic
	matched = repo.LookupAll(&types.MockScenarioKeyData{
		Method:      types.Get,
		Path:        "/api/mytopic/books/11",
		QueryParams: "a=1",
		ContentType: "application",
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
	matched := repo.LookupAll(&types.MockScenarioKeyData{
		Method:      types.Patch,
		Path:        "/v1/todos/1",
		QueryParams: "a=1",
		ContentType: "application",
	})
	// THEN it should not find it
	assert.Equal(t, 0, len(matched))

	// WHEN looking up scenarios with wrong query params
	matched = repo.LookupAll(&types.MockScenarioKeyData{
		Method:      types.Delete,
		Path:        "/v1/todos/1",
		QueryParams: "a=11",
		ContentType: "application",
	})
	assert.Equal(t, 0, len(matched))

	// WHEN looking up scenarios with valid params
	matched = repo.LookupAll(&types.MockScenarioKeyData{
		Method:      types.Delete,
		Path:        "/v1/todos/2",
		QueryParams: "a=1",
		ContentType: "application",
	})
	assert.Equal(t, 10, len(matched))
	for i, m := range matched {
		groups := m.MatchGroups(fmt.Sprintf("/v1/todos/%d", i))
		require.Equal(t, strconv.Itoa(i), groups["id"])
		assert.Equal(t, 1, len(groups))
	}

	//
	// WHEN looking up books by POST with topic
	matched = repo.LookupAll(&types.MockScenarioKeyData{
		Method:      types.Delete,
		Path:        "/v1/mytopic/books/11",
		QueryParams: "a=1",
		ContentType: "application",
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

func buildScenario(method types.MethodType, name string, path string, n int) *types.MockScenario {
	return &types.MockScenario{
		Method:      method,
		Name:        name,
		Path:        path,
		Description: name,
		Request: types.MockHTTPRequest{
			QueryParams: fmt.Sprintf("a=1&b=2&n=%d", n),
			ContentType: "application/json",
			Headers: map[string][]string{
				"ETag": {"981"},
			},
		},
		Response: types.MockHTTPResponse{
			Headers: map[string][]string{
				"ETag": {"123"},
			},
			ContentType: "application/json",
			Contents:    "test body",
			StatusCode:  200,
		},
		WaitBeforeReply: time.Duration(1) * time.Second,
	}
}
