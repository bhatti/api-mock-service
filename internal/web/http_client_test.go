package web

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/bhatti/api-mock-service/internal/types"
)

const todoURL = "https://jsonplaceholder.typicode.com/todos"

func newTest() HTTPClient {
	c := types.Configuration{}
	return NewHTTPClient(&c)
}

func Test_ShouldRealGet(t *testing.T) {
	w := newTest()
	_, _, _, err := w.Handle(
		context.Background(),
		"GET",
		todoURL+"/1",
		map[string][]string{"key": {"value"}},
		make(map[string]string),
		nil)
	if err != nil {
		t.Logf("unexpected response Get error " + err.Error())
	}
}

func Test_ShouldRealDelete(t *testing.T) {
	w := newTest()
	_, _, _, err := w.Handle(
		context.Background(),
		"DELETE",
		todoURL+"/1",
		map[string][]string{"key": {"value"}},
		make(map[string]string),
		nil)
	if err != nil {
		t.Logf("unexpected response Delete error " + err.Error())
	}
}

func Test_ShouldRealDeleteBody(t *testing.T) {
	w := newTest()
	body := io.ReadCloser(io.NopCloser(bytes.NewReader([]byte("hello"))))
	_, _, _, err := w.Handle(
		context.Background(),
		"DELETE",
		todoURL+"/1",
		map[string][]string{"key": {"value"}},
		make(map[string]string),
		body)
	if err != nil {
		t.Logf("unexpected response Delete error " + err.Error())
	}
}

func Test_ShouldRealPostError(t *testing.T) {
	w := newTest()
	_, _, _, err := w.Handle(
		context.Background(),
		"POST",
		todoURL+"_____",
		map[string][]string{"key": {"value"}},
		map[string]string{},
		nil)
	if err == nil {
		t.Logf("expected response Post error ")
	}
}

func Test_ShouldRealPost(t *testing.T) {
	w := newTest()
	_, _, _, err := w.Handle(
		context.Background(),
		"POST",
		todoURL,
		map[string][]string{"key": {"value"}},
		map[string]string{},
		nil)
	if err != nil {
		t.Logf("unexpected response Post error " + err.Error())
	}
}

func Test_ShouldRealPostForm(t *testing.T) {
	w := newTest()
	_, _, _, err := w.Handle(
		context.Background(),
		"POST",
		todoURL,
		map[string][]string{"key": {"value"}},
		map[string]string{"name": "value"},
		nil)
	if err != nil {
		t.Logf("unexpected response Post error " + err.Error())
	}
}

func Test_ShouldRealPostBody(t *testing.T) {
	w := newTest()
	body := io.ReadCloser(io.NopCloser(bytes.NewReader([]byte("hello"))))
	_, _, _, err := w.Handle(
		context.Background(),
		"POST",
		todoURL,
		map[string][]string{"key": {"value"}},
		map[string]string{},
		body)
	if err != nil {
		t.Logf("unexpected response Post error " + err.Error())
	}
}
