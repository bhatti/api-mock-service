package web

import (
	"github.com/bhatti/api-mock-service/internal/types"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func Test_ShouldNotHandleAuth(t *testing.T) {
	adapter := NewAuthAdapter(types.BuildTestConfig())
	req := &http.Request{
		Header: http.Header{
			"Content-Type": []string{"application/json; charset=UTF-8"},
		},
	}
	ok, _, err := adapter.HandleAuth(req)
	assert.NoError(t, err)
	assert.False(t, ok)
}

func Test_ShouldHandleBearerToken(t *testing.T) {
	config := types.BuildTestConfig()
	config.AuthBearerToken = "aa"
	adapter := NewAuthAdapter(config)
	req := &http.Request{
		Header: http.Header{
			"Content-Type": []string{"application/json; charset=UTF-8"},
		},
	}
	ok, _, err := adapter.HandleAuth(req)
	assert.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, "Bearer aa", req.Header.Get(types.AuthorizationHeader))
}

func Test_ShouldHandleBasicAuth(t *testing.T) {
	config := types.BuildTestConfig()
	adapter := NewAuthAdapter(config)
	req := &http.Request{
		Header: http.Header{
			"Content-Type": []string{"application/json; charset=UTF-8"},
		},
	}
	ok, _, err := adapter.HandleAuth(req)
	assert.NoError(t, err)
	assert.False(t, ok)

	config.BasicAuth.Username = "aa"
	config.BasicAuth.Password = ""
	adapter = NewAuthAdapter(config)
	ok, _, err = adapter.HandleAuth(req)
	assert.NoError(t, err)
	assert.False(t, ok)

	config.BasicAuth.Username = ""
	config.BasicAuth.Password = "bb"
	adapter = NewAuthAdapter(config)
	ok, _, err = adapter.HandleAuth(req)
	assert.NoError(t, err)
	assert.False(t, ok)

	config.BasicAuth.Username = "aa"
	config.BasicAuth.Password = "bb"
	ok, _, err = adapter.HandleAuth(req)
	assert.NoError(t, err)
	assert.True(t, ok)
}
