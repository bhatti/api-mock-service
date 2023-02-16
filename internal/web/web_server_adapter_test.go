package web

import (
	"embed"
	"encoding/json"
	"github.com/bhatti/api-mock-service/internal/types"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func adapterHandler(c APIContext) error {
	res := make(map[string]string)
	res["url-path"] = c.Request().URL.Path
	res["http-path"] = c.Param("path")
	res["http-method"] = c.Param("method")
	res["http-name"] = c.Param("name")
	for k, v := range c.Request().URL.Query() {
		res["query-"+k] = v[0]
	}
	return c.JSON(200, res)
}

func Test_ShouldInvokeHTTPRequest(t *testing.T) {
	// GIVEN a server adapter
	adapter := NewWebServerAdapter()
	adapter.GET("/_scenarios", adapterHandler)
	adapter.PUT("/_scenarios/:method/names/:path", adapterHandler)
	adapter.GET("/_scenarios/:method/:name/:path", adapterHandler)
	adapter.PATCH("/_scenarios/:method/:name/:path", adapterHandler)
	adapter.CONNECT("/_scenarios/:method/:name/:path", adapterHandler)
	adapter.HEAD("/_scenarios/:method/:name/:path", adapterHandler)
	adapter.TRACE("/_scenarios/:method/:name/:path", adapterHandler)
	adapter.OPTIONS("/_scenarios/:method/:name/:path", adapterHandler)
	adapter.POST("/_scenarios", adapterHandler)
	adapter.DELETE("/_scenarios/:method/:name/:path", adapterHandler)
	adapter.POST("/_oapi", adapterHandler)
	adapter.Static("", "")
	adapter.Embed(embed.FS{}, "", "")
	adapter.Start("")
	adapter.Stop()

	methodPaths := map[string]string{
		"/_scenarios?a=1&b=1":                     "GET",
		"/_scenarios/POST/names/my/path?a=1&b=1":  "PUT",
		"/_scenarios/POST/myname/my/path?a=2&b=1": "GET",
		"/_scenarios/POST/myname/my/path?a=3&b=1": "PATCH",
		"/_scenarios/POST/myname/my/path?a=4&b=1": "CONNECT",
		"/_scenarios/POST/myname/my/path?a=5&b=1": "HEAD",
		"/_scenarios/POST/myname/my/path?a=6&b=1": "TRACE",
		"/_scenarios/POST/myname/my/path?a=7&b=1": "OPTIONS",
		"/_scenarios?a=8&b=1":                     "POST",
		"/_scenarios/POST/myname/my/path?a=9&b=1": "DELETE",
		"/_oapi?a=10&b=1":                         "POST",
	}
	for path, method := range methodPaths {
		u, err := url.Parse("http://localhost:8080" + path)
		require.NoError(t, err)
		req := &http.Request{
			URL:    u,
			Method: method,
			Header: http.Header{"X1": []string{"val1"}, types.ContentTypeHeader: []string{"json"}},
		}
		res, err := adapter.Invoke(req)
		require.NoError(t, err)
		require.NotNil(t, res)
		params := make(map[string]string)
		b, err := io.ReadAll(res.Body)
		require.NoError(t, err)
		err = json.Unmarshal(b, &params)
		require.NoError(t, err)
		if strings.Contains(path, "POST") {
			require.Equal(t, "POST", params["http-method"])
		}
		if strings.Contains(path, "my/path") {
			require.Equal(t, "my/path", params["http-path"])
		}
		if strings.Contains(path, "myname") {
			require.Equal(t, "myname", params["http-name"])
		}
		require.Equal(t, "1", params["query-b"])
	}
}

func Test_ShouldNotInvokeHTTPRequestWithUnknownPath(t *testing.T) {
	// GIVEN a server adapter
	adapter := NewWebServerAdapter()
	// WHEN using unknown path
	u, err := url.Parse("http://localhost:8080/abc")
	require.NoError(t, err)
	req := &http.Request{
		URL:    u,
		Method: "POST",
		Header: http.Header{"X1": []string{"val1"}, types.ContentTypeHeader: []string{"json"}},
	}
	// THEN it should fail to invoke
	res, err := adapter.Invoke(req)
	require.Error(t, err)
	require.Nil(t, res)
}

func Test_ShouldNotInvokeHTTPRequestWithUnknownMethod(t *testing.T) {
	// GIVEN a server adapter
	adapter := NewWebServerAdapter()
	// WHEN using unknown path
	u, err := url.Parse("http://localhost:8080/abc")
	require.NoError(t, err)
	req := &http.Request{
		URL:    u,
		Method: "XXXX",
		Header: http.Header{"X1": []string{"val1"}, types.ContentTypeHeader: []string{"json"}},
	}
	// THEN it should fail to invoke
	res, err := adapter.Invoke(req)
	require.Error(t, err)
	require.Nil(t, res)
}
