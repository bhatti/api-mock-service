package web

import (
	"bytes"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/url"
	"testing"
)

func Test_ShouldNotBuildMockScenarioKeyDataWithoutMethod(t *testing.T) {
	request := &http.Request{}
	request.Header = make(http.Header)
	request.Body = io.NopCloser(bytes.NewReader([]byte{}))
	request.URL, _ = url.Parse("http://localhost/path?x=a")
	_, err := BuildMockScenarioKeyData(request)
	require.Error(t, err)
}

func Test_ShouldBuildMockScenarioKeyData(t *testing.T) {
	request := &http.Request{}
	request.Header = http.Header{"one": {}, "two": {"x"}}
	request.Body = io.NopCloser(bytes.NewReader([]byte{}))
	request.URL, _ = url.Parse("http://localhost/path?x=a")
	request.Method = "POST"
	scenario, err := BuildMockScenarioKeyData(request)
	require.NoError(t, err)
	require.Equal(t, "/path", scenario.Path)
}
