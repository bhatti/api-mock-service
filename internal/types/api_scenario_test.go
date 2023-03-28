package types

import (
	"github.com/stretchr/testify/require"
	"net/http"
	"net/url"
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

func Test_ShouldAssertRequestContentsPatternOrContent(t *testing.T) {
	// GIVEN a valid mock scenario
	scenario := buildScenario()
	scenario.Request.Contents = "content"
	scenario.Request.ExampleContents = "example"
	require.Equal(t, "example", scenario.Request.AssertContentsPatternOrContent())
	scenario.Request.ExampleContents = ""
	scenario.Request.AssertContentsPattern = "assert"
	require.Equal(t, "assert", scenario.Request.AssertContentsPatternOrContent())
	scenario.Request.ExampleContents = ""
	scenario.Request.AssertContentsPattern = ""
	require.Equal(t, "content", scenario.Request.AssertContentsPatternOrContent())
}

func Test_ShouldAssertResponseContentsPatternOrContent(t *testing.T) {
	// GIVEN a valid mock scenario
	scenario := buildScenario()
	scenario.Response.Contents = "content"
	scenario.Response.ExampleContents = "example"
	require.Equal(t, "example", scenario.Response.AssertContentsPatternOrContent())
	scenario.Response.ExampleContents = ""
	scenario.Response.AssertContentsPattern = "assert"
	require.Equal(t, "assert", scenario.Response.AssertContentsPatternOrContent())
	scenario.Response.ExampleContents = ""
	scenario.Response.AssertContentsPattern = ""
	require.Equal(t, "content", scenario.Response.AssertContentsPatternOrContent())
}

func Test_ShouldBuildURL(t *testing.T) {
	// GIVEN a valid mock scenario
	scenario := buildScenario()
	scenario.Path = "/path1/abc"
	require.Equal(t, "/path1/abc", scenario.BuildURL(""))
	require.Equal(t, "xyz/path1/abc", scenario.BuildURL("xyz"))
}

func Test_ShouldAddAWSHeaders(t *testing.T) {
	// GIVEN a valid mock scenario
	scenario := buildScenario()
	scenario.addAWSHeaders()
	require.Equal(t, 3, len(scenario.Authentication))
	require.Equal(t, "apiKey", scenario.Authentication["aws.auth.sigv4"].Type)
	require.Equal(t, "apiKey", scenario.Authentication["smithy.scenario.httpApiKeyAuth"].Type)
	require.Equal(t, "http", scenario.Authentication["bearerAuth"].Type)
}

func Test_ShouldAddAuthHeaders(t *testing.T) {
	// GIVEN a valid mock scenario
	scenario := buildScenario()
	scenario.addAuthHeaders()
	require.Equal(t, 2, len(scenario.Authentication))
	require.Equal(t, "http", scenario.Authentication["basicAuth"].Type)
	require.Equal(t, "http", scenario.Authentication["bearerAuth"].Type)
}

func Test_ShouldGetRequestAuthHeader(t *testing.T) {
	// GIVEN a valid mock scenario
	scenario := BuildTestScenario(Get, "name", "/path", 1)
	// WHEN fetching content type
	require.Equal(t, "", scenario.Request.AuthHeader())
	scenario.Request.Headers["authorization"] = "abc"
	require.Equal(t, "abc", scenario.Request.AuthHeader())
}

func Test_ShouldSanitizeRegexValue(t *testing.T) {
	require.Equal(t, "__1", SanitizeRegexValue("__1"))
	require.Equal(t, "1", SanitizeRegexValue("(1)"))
	require.Equal(t, "1", SanitizeRegexValue("1"))
}

func Test_ShouldGetRequestContentType(t *testing.T) {
	// GIVEN a valid mock scenario
	scenario := buildScenario()
	// WHEN fetching content type
	require.Equal(t, "", scenario.Request.ContentType(""))
	scenario.Request.Headers[ContentTypeHeader] = "abc"
	require.Equal(t, "abc", scenario.Request.ContentType(""))
}

func Test_ShouldGetResponseContentType(t *testing.T) {
	// GIVEN a valid mock scenario
	scenario := buildScenario()
	// WHEN fetching content type
	require.Equal(t, "application/json", scenario.Response.ContentType(""))
	scenario.Response.Headers[ContentTypeHeader] = []string{"abc"}
	require.Equal(t, "abc", scenario.Response.ContentType(""))
}

func Test_ShouldGetRequestTarget(t *testing.T) {
	// GIVEN a valid mock scenario
	scenario := buildScenario()
	scenario.Path = "/api/v1/"
	// WHEN fetching target
	require.Equal(t, "", scenario.Request.TargetHeader())
	scenario.Request.Headers["my-target"] = "abc"
	require.Equal(t, "abc", scenario.Request.TargetHeader())
	require.Equal(t, "post__api_v1_abc", scenario.MethodPathTarget())
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

func Test_ShouldGetURL(t *testing.T) {
	// GIVEN a valid mock scenario
	u, err := url.Parse("http://localhost")
	require.NoError(t, err)
	scenario := buildScenario()
	scenario.Path = "/2/openapi.json"
	// WHEN setting name
	scenario.SetName("prefix")
	require.False(t, scenario.HasURL())
	_, err = scenario.GetURL("http://localhost")
	require.NoError(t, err)
	_, err = scenario.GetNetURL(u)
	require.NoError(t, err)
	scenario.BaseURL = u.String()
	require.True(t, scenario.HasURL())
	_, err = scenario.GetURL("")
	require.NoError(t, err)
}

func Test_ShouldSanitizeSpecialChar(t *testing.T) {
	require.Equal(t, "Delete Rules by rule ID", sanitizeSpecialChars("Delete Rules (by rule ID)", ""))
	require.Equal(t, "Users followed Lists", sanitizeSpecialChars("User's followed Lists", ""))
}

func Test_ShouldGetStartEndTime(t *testing.T) {
	// GIVEN a valid mock scenario
	scenario := buildScenario()
	require.Equal(t, int64(0), scenario.GetMillisTime())
	scenario.StartTime = time.Now()
	scenario.EndTime = time.Now().Add(time.Second)
	require.Equal(t, int64(1000), scenario.GetMillisTime())
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

func Test_ShouldSetNameWithPathVariables(t *testing.T) {
	// GIVEN a valid mock scenario
	scenario := buildScenario()
	scenario.Path = "{var}/2/openapi.json"
	// WHEN setting name
	scenario.SetName("prefix")
	// THEN it should succeed
	require.Contains(t, scenario.Name, "prefix-200-")
}

func Test_ShouldNotValidateEmptyMockScenario(t *testing.T) {
	// GIVEN a empty mock scenario
	scenario := &APIScenario{}
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
	scenario.Name = ""
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
	scenario := APIScenario{
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
	require.Equal(t, "post__v1_category_cat_books_id", scenario.MethodPath())
	require.Equal(t, "abc", scenario.SafeName())
}

func Test_ShouldMatchGroupsInMockScenarioKeyDataWithColon(t *testing.T) {
	// GIVEN a empty mock scenario
	scenario := APIScenario{
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

func Test_ShouldBuildScenarioFromHTTPWithNilURL(t *testing.T) {
	config := BuildTestConfig()
	_, err := BuildScenarioFromHTTP(
		config,
		"prefix",
		nil,
		"POST",
		"group",
		"1.1",
		"1.1",
		[]byte("body"),
		[]byte("body"),
		map[string][]string{"q": {"val"}},
		map[string][]string{"p": {"val"}},
		map[string][]string{"h": {"val"}},
		"application/json",
		map[string][]string{"h": {"val"}},
		"application/json",
		200,
		time.Now(),
		time.Now())
	require.Error(t, err)
}

func Test_ShouldBuildScenarioFromHTTP(t *testing.T) {
	u, err := url.Parse("http://localhost:8000")
	require.NoError(t, err)
	config := BuildTestConfig()
	scenario, err := BuildScenarioFromHTTP(
		config,
		"prefix",
		u,
		"POST",
		"group",
		"1.1",
		"1.1",
		[]byte("body"),
		[]byte("body"),
		map[string][]string{"q": {"val"}},
		map[string][]string{"p": {"val"}},
		map[string][]string{"Content-Type": {"application/json"}},
		"application/json",
		map[string][]string{"Content-Type": {"application/json"}},
		"application/json",
		200,
		time.Now(),
		time.Now())
	require.NoError(t, err)
	err = scenario.Request.Assert(
		map[string]string{"q": "val"},
		map[string]string{"p": "val"},
		map[string][]string{"h": {"val"}},
		[]byte("body"),
		make(map[string]any))
	require.Error(t, err)
	err = scenario.Request.Assert(
		map[string]string{"q": "val"},
		map[string]string{"p": "val"},
		map[string][]string{"Content-Type": {"application/json"}},
		[]byte("body"),
		make(map[string]any))
	require.NoError(t, err)
	err = scenario.Response.Assert(
		map[string][]string{"h": {"val"}},
		[]byte("body"),
		map[string]any{"elapsed": 1, "status": 200})
	require.Error(t, err)
	err = scenario.Response.Assert(
		map[string][]string{"Content-Type": {"application/json"}},
		[]byte("body"),
		map[string]any{"elapsed": 1, "status": 200})
	require.NoError(t, err)

	templateParams, queryParams, postParams, reqHeaders := scenario.Request.BuildTemplateParams(
		&http.Request{URL: u}, map[string]string{"path": "val"},
		map[string][]string{"h": {"val"}}, map[string]any{"x": 1})
	require.Equal(t, 5, len(templateParams))
	require.Equal(t, 2, len(queryParams))
	require.Equal(t, 1, len(postParams))
	require.Equal(t, 2, len(reqHeaders))
}

func Test_ShouldNormalizeGroup(t *testing.T) {
	require.Equal(t, "root", NormalizeGroup("", ""))
	require.Equal(t, "title", NormalizeGroup("title", ""))
	require.Equal(t, "path1", NormalizeGroup("", "/path1/{test}"))
	require.Equal(t, "path1", NormalizeGroup("", "/path1/:test"))
	require.Equal(t, "path1_path2", NormalizeGroup("", "/path1/path2"))
	require.Equal(t, "root", NormalizeGroup("", "/"))
}

func buildScenario() *APIScenario {
	scenario := &APIScenario{
		Method:         Post,
		Name:           "scenario",
		Path:           "/path1/\\\\//test1//abc////",
		Description:    "",
		Group:          "test-group",
		Tags:           []string{"tag1", "tag2"},
		Authentication: make(map[string]APIAuthorization),
		Request: APIRequest{
			Headers: make(map[string]string),
			AssertQueryParamsPattern: map[string]string{
				"a": "1",
				"b": "2",
			},
			AssertHeadersPattern: map[string]string{
				"CTag": "981",
			},
			Variables: make(map[string]string),
		},
		Response: APIResponse{
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
