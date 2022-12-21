package types

import (
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func Test_ShouldValidateProperMockScenario(t *testing.T) {
	// GIVEN a valid mock scenario
	scenario := buildScenario()
	// WHEN validating scenario
	// THEN it should succeed
	require.NoError(t, scenario.Validate())
	require.Equal(t, "path1/test1/abc", scenario.NormalPath('/'))
	require.Equal(t, "a", NormalizePath("a", '_'))
	require.Equal(t, "_", NormalizePath("/", '_'))
	require.True(t, scenario.Digest() != "")
}

func Test_ShouldValidateDotPathForMockScenario(t *testing.T) {
	// GIVEN a valid mock scenario
	scenario := buildScenario()
	scenario.Path = "/2/openapi.json"
	// WHEN validating scenario
	// THEN it should succeed
	require.NoError(t, scenario.Validate())
	require.Equal(t, "2/openapi.json", scenario.NormalPath('/'))
	scenario.Path = "/2018-06-01/runtime/invocation/next"
	require.NoError(t, scenario.Validate())
}

func Test_ShouldSetName(t *testing.T) {
	// GIVEN a valid mock scenario
	scenario := buildScenario()
	scenario.Path = "/2/openapi.json"
	// WHEN setting name
	scenario.SetName("prefix")
	// THEN it should succeed
	require.Contains(t, scenario.Name, "2-openapi.json-200-", scenario.Name)
}

func Test_ShouldNotValidateEmptyMockScenario(t *testing.T) {
	// GIVEN a empty mock scenario
	scenario := &MockScenario{}
	// WHEN validating scenario
	// THEN it should fail
	require.Error(t, scenario.Validate())
	scenario.Method = Get
	require.Error(t, scenario.Validate())
	scenario.Path = "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Faucibus ornare suspendisse sed nisi lacus sed viverra tellus in. Lacus vel facilisis volutpat est velit egestas dui. Neque egestas congue quisque egestas diam in arcu. Risus pretium quam vulputate dignissim suspendisse. Iaculis urna id volutpat lacus laoreet. Viverra mauris in aliquam sem fringilla. Risus ultricies tristique nulla aliquet enim tortor at auctor urna. Feugiat nibh sed pulvinar proin gravida hendrerit lectus. Tempus imperdiet nulla malesuada pellentesque elit eget gravida cum sociis. Integer quis auctor elit sed vulputate mi sit amet mauris. Proin libero nunc consequat interdum varius sit amet mattis vulputate. Arcu ac tortor dignissim convallis aenean."
	require.Error(t, scenario.Validate())
	scenario.Path = "/path1****//\\\\//test1/2///:id"
	require.Error(t, scenario.Validate())
	scenario.Path = "/path1//\\\\//test1/2///:id"
	require.Error(t, scenario.Validate())
	scenario.Name = "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Faucibus ornare suspendisse sed nisi lacus sed viverra tellus in. Lacus vel facilisis volutpat est velit egestas dui. Neque egestas congue quisque egestas diam in arcu. Risus pretium quam vulputate dignissim suspendisse. Iaculis urna id volutpat lacus laoreet. Viverra mauris in aliquam sem fringilla. Risus ultricies tristique nulla aliquet enim tortor at auctor urna. Feugiat nibh sed pulvinar proin gravida hendrerit lectus. Tempus imperdiet nulla malesuada pellentesque elit eget gravida cum sociis. Integer quis auctor elit sed vulputate mi sit amet mauris. Proin libero nunc consequat interdum varius sit amet mattis vulputate. Arcu ac tortor dignissim convallis aenean."
	require.Error(t, scenario.Validate())
	scenario.Name = "test1****"
	require.Error(t, scenario.Validate())
	scenario.Name = "test1"
	scenario.Response.Contents = "test"
	require.NoError(t, scenario.Validate())
	require.Equal(t, "path1/test1/2/:id", scenario.NormalPath('/'))
	require.True(t, scenario.Digest() != "")
	require.Equal(t, "/path1/test1/2/:id", scenario.ToKeyData().Path)
}

func Test_ShouldMatchGroupsInMockScenarioKeyData(t *testing.T) {
	// GIVEN a empty mock scenario
	scenario := MockScenario{
		Name:   "abc*",
		Method: Post,
		Path:   "/v1/category/{cat}/books/{id}",
	}
	keyData := scenario.ToKeyData()
	groups := keyData.MatchGroups("/v1/category/history/books/101")
	require.Equal(t, 2, len(groups))
	require.Equal(t, "history", groups["cat"])
	require.Equal(t, "101", groups["id"])
	require.Equal(t, "POSTabc*/v1/category/{cat}/books/{id}", scenario.String())
	require.Equal(t, "post__v1_category_{cat}_books_{id}", scenario.MethodPath())
	require.Equal(t, "abc", scenario.SafeName())
}

func Test_ShouldMatchGroupsInMockScenarioKeyDataWithColon(t *testing.T) {
	// GIVEN a empty mock scenario
	scenario := MockScenario{
		Path: "/v1/category/:cat/books/:id",
	}
	keyData := scenario.ToKeyData()
	groups := keyData.MatchGroups("/v1/category/history/books/101")
	require.Equal(t, 2, len(groups))
	require.Equal(t, "history", groups["cat"])
	require.Equal(t, "101", groups["id"])
	require.Equal(t, "/v1/category/:cat/books/:id", scenario.String())
}

func Test_ShouldMatchRegex(t *testing.T) {
	require.True(t, reMatch("abc", "abc"))
	require.True(t, reMatch("\\d+", "3"))
}

func Test_ShouldNormalizePath(t *testing.T) {
	require.Equal(t, "path1_id_:id", NormalizePath("/path1/id/:id", '_'))
	require.Equal(t, "path1_id_{id}", NormalizePath("/path1/id/{id}", '_'))
}

func Test_ShouldNormalizeDirPath(t *testing.T) {
	require.Equal(t, "path1/id", NormalizeDirPath("/path1/id/:id"))
	require.Equal(t, "path1/id", NormalizeDirPath("/path1/id/{id}"))
}

func Test_ToMethodShouldValidateMethod(t *testing.T) {
	m, err := ToMethod("get")
	require.NoError(t, err)
	require.Equal(t, Get, m)
	m, err = ToMethod("post")
	require.NoError(t, err)
	require.Equal(t, Post, m)
	m, err = ToMethod("put")
	require.NoError(t, err)
	require.Equal(t, Put, m)
	m, err = ToMethod("delete")
	require.NoError(t, err)
	require.Equal(t, Delete, m)
	m, err = ToMethod("option")
	require.NoError(t, err)
	require.Equal(t, Option, m)
	m, err = ToMethod("head")
	require.NoError(t, err)
	require.Equal(t, Head, m)
	m, err = ToMethod("patch")
	require.NoError(t, err)
	require.Equal(t, Patch, m)
	m, err = ToMethod("connect")
	require.NoError(t, err)
	require.Equal(t, Connect, m)
	m, err = ToMethod("trace")
	require.NoError(t, err)
	require.Equal(t, Trace, m)
	m, err = ToMethod("options")
	require.NoError(t, err)
	require.Equal(t, Options, m)
	_, err = ToMethod("error")
	require.Error(t, err)
}

func buildScenario() *MockScenario {
	scenario := &MockScenario{
		Method:      Post,
		Name:        "scenario",
		Path:        "/path1/\\\\//test1//abc////",
		Description: "",
		Request: MockHTTPRequest{
			MatchQueryParams: map[string]string{
				"a": "1",
				"b": "2",
			},
			MatchHeaders: map[string]string{
				"CTag": "981",
			},
		},
		Response: MockHTTPResponse{
			Headers: map[string][]string{
				"ETag":            {"123"},
				ContentTypeHeader: {"application/json"},
			},
			Contents:   "test body",
			StatusCode: 200,
		},
		WaitBeforeReply: time.Duration(1) * time.Second,
	}
	return scenario
}
