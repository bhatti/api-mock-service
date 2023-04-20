package pm

import (
	"bytes"
	"fmt"
	"github.com/bhatti/api-mock-service/internal/types"
	"github.com/stretchr/testify/require"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ShouldCreateCollection(t *testing.T) {
	c := CreateCollection("a-name", "a-desc")

	assert.Equal(
		t,
		&PostmanCollection{
			Info: PostmanInfo{
				Name: "a-name",
				Description: PostmanCollectionDescription{
					Content: "a-desc",
				},
			},
		},
		c,
	)
}

func Test_ShouldAddItemIntoCollection(t *testing.T) {
	BasicCollection := CreateCollection("Postman collection", "v2.1.0")
	BasicCollection.AddItem(&PostmanItems{Name: "Item1"})
	BasicCollection.AddItem(&PostmanItems{Name: "Item2"})
	BasicCollection.AddItem(&PostmanItems{Name: "Item3"})

	require.Equal(t,
		[]*PostmanItems{
			{Name: "Item1"},
			{Name: "Item2"},
			{Name: "Item3"},
		},
		BasicCollection.Items,
	)
}

func Test_ShouldAddItemGroupIntoCollection(t *testing.T) {
	BasicCollection := CreateCollection("Postman collection", "v2.1.0")
	BasicCollection.AddItemGroup("new-item-group")
	BasicCollection.AddItemGroup("another-new-item-group")
	require.Equal(t,
		[]*PostmanItems{
			{Name: "new-item-group", Items: make([]*PostmanItems, 0)},
			{Name: "another-new-item-group", Items: make([]*PostmanItems, 0)},
		},
		BasicCollection.Items,
	)
}

func Test_ShouldParseCollection(t *testing.T) {
	cases := []struct {
		scenario           string
		testFile           string
		expectedCollection *PostmanCollection
		expectedError      error
	}{
		{
			"v2.1.0 collection",
			"../../../fixtures/postman.json",
			buildV210Collection(),
			nil,
		},
	}

	for _, tc := range cases {
		file, err := os.Open(tc.testFile)
		require.NoError(t, err)
		c, err := ParseCollection(file)
		require.Equal(t, tc.expectedError, err, tc.scenario)
		require.Equal(t, tc.expectedCollection, c, tc.scenario)
		require.Equal(t, len(tc.expectedCollection.Variables), len(c.Variables), tc.scenario)
	}
}

func Test_ShouldWriteCollection(t *testing.T) {
	cases := []struct {
		scenario       string
		testCollection *PostmanCollection
		expectedFile   string
		expectedError  error
	}{
		{
			"v2.1.0 collection",
			buildV210Collection(),
			"../../../fixtures/postman.json",
			nil,
		},
	}

	for _, tc := range cases {
		var buf bytes.Buffer

		require.NoError(t, tc.testCollection.Write(&buf))

		file, err := os.ReadFile(tc.expectedFile)
		require.NoError(t, err)
		require.Equal(t, string(file), fmt.Sprintf("%s\n", buf.String()), tc.scenario)
	}
}

func Test_ShouldSimplePOSTItem(t *testing.T) {
	c := CreateCollection("Test PostmanCollection", "My Test PostmanCollection")

	file, err := os.Create("postman_collection.json")
	require.NoError(t, err)

	defer func() {
		_ = file.Close()
	}()

	pURL := PostmanURL{
		Raw:      "https://test.com",
		Protocol: "https",
		Host:     []string{"test", "com"},
	}

	headers := []*PostmanHeader{{
		Key:   "h1",
		Value: "h1-value",
	}}

	pBody := PostmanRequestBody{
		Mode:    "raw",
		Raw:     "{\"a\":\"1234\",\"b\":123}",
		Options: &PostmanRequestBodyOptions{PostmanRequestBodyOptionsRaw{Language: "json"}},
	}

	pReq := PostmanRequest{
		Method: types.Post,
		URL:    &pURL,
		Header: headers,
		Body:   &pBody,
	}

	cr := PostmanRequest{
		Method: types.Post,
		URL:    &pURL,
		Header: pReq.Header,
		Body:   pReq.Body,
	}

	item := CreatePostmanItem(PostmanItem{
		Name:    "Test-POST",
		Request: &cr,
	})

	c.AddItemGroup("grp1").AddItem(item)

	err = c.Write(file)
	require.NoError(t, err)

	err = os.Remove("postman_collection.json")
	require.NoError(t, err)
}

func Test_ShouldSimpleGETItem(t *testing.T) {
	c := CreateCollection("Test PostmanCollection", "My Test PostmanCollection")

	file, err := os.Create("postman_collection.json")
	require.NoError(t, err)

	defer func() {
		_ = file.Close()
	}()

	pURL := PostmanURL{
		Raw:      "https://test.com?a=3",
		Protocol: "https",
		Host:     []string{"test", "com"},
		Query: []*PostmanQueryParam{
			{Key: "param1", Value: "value1"},
			{Key: "param2", Value: "value2"},
		},
	}

	headers := []*PostmanHeader{}
	headers = append(headers, &PostmanHeader{
		Key:   "h1",
		Value: "h1-value",
	})
	headers = append(headers, &PostmanHeader{
		Key:   "h2",
		Value: "h2-value",
	})

	pReq := PostmanRequest{
		Method: types.Get,
		URL:    &pURL,
		Header: headers,
	}

	item := CreatePostmanItem(PostmanItem{
		Name:    "Test-GET",
		Request: &pReq,
	})

	c.AddItemGroup("grp1").AddItem(item)

	err = c.Write(file)
	require.NoError(t, err)

	err = os.Remove("postman_collection.json")
	require.NoError(t, err)
}

func buildV210Collection() *PostmanCollection {
	return &PostmanCollection{
		Info: PostmanInfo{
			Name: "Go Collection",
			Description: PostmanCollectionDescription{
				Content: "Awesome description",
			},
			Version: "v2.1.0",
			Schema:  "https://schema.getpostman.com/json/collection/v2.1.0/collection.json",
		},
		Items: []*PostmanItems{
			{
				Name: "This is a folder",
				Items: []*PostmanItems{
					{
						Name: "An item inside a folder",
					},
				},
				Variables: []*PostmanVariable{
					{
						Name:  "api-key",
						Value: "abcd1234",
					},
				},
				Auth: &PostmanAuth{
					Type: types.Bearer,
					Bearer: []*PostmanAuthParam{
						{
							Key:   "token",
							Value: "a-bearer-token",
							Type:  "string",
						},
					},
				},
			},
			{
				Name: "This is a request",
				Request: &PostmanRequest{
					URL: &PostmanURL{
						Raw: "http://www.google.fr",
					},
					Method: types.Get,
				},
				Responses: []*PostmanResponse{
					{
						Name: "This is a response",
						OriginalRequest: &PostmanRequest{
							URL: &PostmanURL{
								Raw: "http://www.google.fr",
							},
							Method: types.Get,
						},
						Status: "OK",
						Code:   200,
						Cookies: []*PostmanCookie{
							{
								Domain: "a-domain",
								Path:   "a-path",
							},
						},
						Headers: &PostmanHeaderList{
							Headers: []*PostmanHeader{
								{
									Key:   "Content-Type",
									Value: "application/json",
								},
							},
						},
						Body: "the-body",
					},
				},
				Events: []*PostmanEvent{
					{
						Listen: PreRequest,
						Script: &PostmanScript{
							Type: "text/javascript",
							Exec: []string{"console.log(\"foo\")"},
						},
					},
				},
			},
			{
				Name: "JSON-RPC Request",
				Request: &PostmanRequest{
					URL: &PostmanURL{
						Raw: "https://gurujsonrpc.appspot.com/guru",
						Variables: []*PostmanVariable{
							{
								Name:  "an-url-variable",
								Value: "an-url-variable-value",
							},
						},
					},
					Auth: &PostmanAuth{
						Type: types.Basic,
						Basic: []*PostmanAuthParam{
							{
								Key:   "password",
								Value: "my-password",
								Type:  "string",
							},
						},
					},
					Method: types.Post,
					Header: []*PostmanHeader{
						{
							Key:   "Content-Type",
							Value: "application/json",
						},
					},
					Body: &PostmanRequestBody{
						Mode:    "raw",
						Raw:     "{\"aKey\":\"a-value\"}",
						Options: &PostmanRequestBodyOptions{PostmanRequestBodyOptionsRaw{Language: "json"}},
					},
				},
			},
			{
				Name:  "An empty folder",
				Items: make([]*PostmanItems, 0),
			},
		},
		Events: []*PostmanEvent{
			{
				Listen: Test,
				Script: &PostmanScript{
					Type: "text/javascript",
					Exec: []string{"console.log(\"bar\")"},
				},
			},
		},
		Auth: &PostmanAuth{
			Type: "bearer",
			Bearer: []*PostmanAuthParam{
				{
					Type:  "string",
					Key:   "token",
					Value: "a-bearer-token",
				},
			},
		},
		Variables: []*PostmanVariable{
			{
				Name:  "a-global-collection-variable",
				Value: "a-global-value",
			},
		},
	}
}
