package web

import (
	"bytes"
	"context"
	"fmt"
	"github.com/bhatti/api-mock-service/internal/types"
	"github.com/labstack/echo/v4"
	"io"
	"net/http"
	"sort"
)

// MethodPathHandler mapping
type MethodPathHandler struct {
	method  types.MethodType
	path    string
	handler HandlerFunc
}

// ServerAdapter struct
type ServerAdapter struct {
	handlers []MethodPathHandler
}

// Adapter defines methods to delegate HTTP APIs
type Adapter interface {
	// Invoke http method
	Invoke(request *http.Request) (response *http.Response, err error)
}

// NewWebServerAdapter constructor
func NewWebServerAdapter() *ServerAdapter {
	return &ServerAdapter{
		handlers: make([]MethodPathHandler, 0),
	}
}

// Invoke finds handler in controllers and invokes it if matched
func (a *ServerAdapter) Invoke(request *http.Request) (response *http.Response, err error) {
	method, err := types.ToMethod(request.Method)
	if err != nil {
		return nil, err
	}
	for _, mph := range a.handlers {
		pathParams := types.MatchPathGroups(mph.path, request.URL.Path)
		if pathParams == nil || mph.method != method {
			continue
		}
		response = &http.Response{}
		response.Request = request
		response.TransferEncoding = request.TransferEncoding
		response.Header = make(http.Header)
		response.Body = io.NopCloser(bytes.NewReader([]byte{}))
		ctx := NewDefaultAPIContext(context.Background(), request, response, nil, pathParams)
		err = mph.handler(ctx)
		return
	}
	return nil, fmt.Errorf("not matched %s - %s", method, request.URL.Path)
}

// AddMiddleware middleware
func (a *ServerAdapter) AddMiddleware(_ echo.MiddlewareFunc) {
}

// GET handler calls HTTP GET method
func (a *ServerAdapter) GET(path string, handler HandlerFunc, _ ...echo.MiddlewareFunc) *echo.Route {
	a.appendHandler(MethodPathHandler{method: types.Get, path: path, handler: handler})
	return &echo.Route{}
}

// POST handler calls HTTP POST method
func (a *ServerAdapter) POST(path string, handler HandlerFunc, _ ...echo.MiddlewareFunc) *echo.Route {
	a.appendHandler(MethodPathHandler{method: types.Post, path: path, handler: handler})
	return &echo.Route{}
}

// PUT handler calls HTTP PUT method
func (a *ServerAdapter) PUT(path string, handler HandlerFunc, _ ...echo.MiddlewareFunc) *echo.Route {
	a.appendHandler(MethodPathHandler{method: types.Put, path: path, handler: handler})
	return &echo.Route{}
}

// DELETE handler calls HTTP DELETE method
func (a *ServerAdapter) DELETE(path string, handler HandlerFunc, _ ...echo.MiddlewareFunc) *echo.Route {
	a.appendHandler(MethodPathHandler{method: types.Delete, path: path, handler: handler})
	return &echo.Route{}
}

// CONNECT calls HTTP CONNECT method
func (a *ServerAdapter) CONNECT(path string, handler HandlerFunc, _ ...echo.MiddlewareFunc) *echo.Route {
	a.appendHandler(MethodPathHandler{method: types.Connect, path: path, handler: handler})
	return &echo.Route{}
}

// HEAD calls HTTP HEAD method
func (a *ServerAdapter) HEAD(path string, handler HandlerFunc, _ ...echo.MiddlewareFunc) *echo.Route {
	a.appendHandler(MethodPathHandler{method: types.Head, path: path, handler: handler})
	return &echo.Route{}
}

// OPTIONS calls HTTP OPTIONS method
func (a *ServerAdapter) OPTIONS(path string, handler HandlerFunc, _ ...echo.MiddlewareFunc) *echo.Route {
	a.appendHandler(MethodPathHandler{method: types.Options, path: path, handler: handler})
	return &echo.Route{}
}

// PATCH calls HTTP PATCH method
func (a *ServerAdapter) PATCH(path string, handler HandlerFunc, _ ...echo.MiddlewareFunc) *echo.Route {
	a.appendHandler(MethodPathHandler{method: types.Patch, path: path, handler: handler})
	return &echo.Route{}
}

// TRACE calls HTTP TRACE method
func (a *ServerAdapter) TRACE(path string, handler HandlerFunc, _ ...echo.MiddlewareFunc) *echo.Route {
	a.appendHandler(MethodPathHandler{method: types.Trace, path: path, handler: handler})
	return &echo.Route{}
}

func (a *ServerAdapter) appendHandler(mph MethodPathHandler) {
	a.handlers = append(a.handlers, mph)
	sort.Slice(a.handlers, func(i, j int) bool {
		return len(a.handlers[i].path) > len(a.handlers[j].path)
	})
}

// Start server
func (a *ServerAdapter) Start(string) {
}

// Stop server
func (a *ServerAdapter) Stop() {
}

// Static serve static assets
func (a *ServerAdapter) Static(string) {
}
