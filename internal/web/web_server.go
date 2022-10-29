package web

import (
	"net/http"
	"os"

	"github.com/bhatti/api-mock-service/internal/types"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// HandlerFunc defines a function to serve HTTP requests.
type HandlerFunc func(APIContext) error

// WrapHandler wraps `http.Handler` into `echo.HandlerFunc`.
func WrapHandler(h http.Handler) HandlerFunc {
	return func(c APIContext) error {
		h.ServeHTTP(c.Response(), c.Request())
		return nil
	}
}

// Server defines methods for binding http methods
type Server interface {
	GET(path string, h HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	POST(path string, h HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	PUT(path string, h HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	DELETE(path string, h HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	CONNECT(path string, h HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	HEAD(path string, h HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	OPTIONS(path string, h HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	PATCH(path string, h HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	TRACE(path string, h HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	AddMiddleware(m echo.MiddlewareFunc)
	Start(address string)
	Static(dir string)
	Stop()
}

// DefaultWebServer defines default web server
type DefaultWebServer struct {
	config *types.Configuration
	e      *echo.Echo
}

// NewDefaultWebServer creates new instance of web server
func NewDefaultWebServer(config *types.Configuration) Server {
	ws := &DefaultWebServer{config: config, e: echo.New()}
	defaultLoggerConfig := middleware.LoggerConfig{
		Skipper: middleware.DefaultSkipper,
		Format: `{"time":"${time_rfc3339_nano}","id":"${id}","remote_ip":"${remote_ip}",` +
			`"host":"${host}","method":"${method}","path":"${path}",` +
			`"status":${status},"error":"${error}","latency":${latency},"latency_human":"${latency_human}"` +
			`,"bytes_in":${bytes_in},"bytes_out":${bytes_out}}` + "\n",
		CustomTimeFormat: "2006-01-02 15:04:05.00000",
	}
	ws.e.Use(middleware.LoggerWithConfig(defaultLoggerConfig))
	ws.e.Use(middleware.Recover())
	ws.e.HTTPErrorHandler = func(err error, c echo.Context) {
		ws.e.DefaultHTTPErrorHandler(err, c)
	}

	//CORS
	ws.e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowMethods: []string{echo.GET, echo.HEAD, echo.PUT, echo.PATCH, echo.POST, echo.DELETE},
	}))
	ws.e.HideBanner = true
	return ws
}

// AddMiddleware adds middleware
func (w *DefaultWebServer) AddMiddleware(m echo.MiddlewareFunc) {
	w.e.Use(m)
}

// GET calls HTTP GET method
func (w *DefaultWebServer) GET(path string, h HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route {
	return w.e.GET(path, func(context echo.Context) error {
		return h(context)
	}, m...)
}

// POST calls HTTP POST method
func (w *DefaultWebServer) POST(path string, h HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route {
	return w.e.POST(path, func(context echo.Context) error {
		return h(context)
	}, m...)
}

// PUT calls HTTP PUT method
func (w *DefaultWebServer) PUT(path string, h HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route {
	return w.e.PUT(path, func(context echo.Context) error {
		return h(context)
	}, m...)
}

// DELETE calls HTTP DELETE method
func (w *DefaultWebServer) DELETE(path string, h HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route {
	return w.e.DELETE(path, func(context echo.Context) error {
		return h(context)
	}, m...)
}

// CONNECT calls HTTP CONNECT method
func (w *DefaultWebServer) CONNECT(path string, h HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route {
	return w.e.CONNECT(path, func(context echo.Context) error {
		return h(context)
	}, m...)
}

// HEAD calls HTTP HEAD method
func (w *DefaultWebServer) HEAD(path string, h HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route {
	return w.e.HEAD(path, func(context echo.Context) error {
		return h(context)
	}, m...)
}

// OPTIONS calls HTTP OPTIONS method
func (w *DefaultWebServer) OPTIONS(path string, h HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route {
	return w.e.OPTIONS(path, func(context echo.Context) error {
		return h(context)
	}, m...)
}

// PATCH calls HTTP PATCH method
func (w *DefaultWebServer) PATCH(path string, h HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route {
	return w.e.PATCH(path, func(context echo.Context) error {
		return h(context)
	}, m...)
}

// TRACE calls HTTP TRACE method
func (w *DefaultWebServer) TRACE(path string, h HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route {
	return w.e.TRACE(path, func(context echo.Context) error {
		return h(context)
	}, m...)
}

// Static - serve assets
func (w *DefaultWebServer) Static(dir string) {
	os.MkdirAll(dir, 0755)
	w.e.Static("/_assets", dir)
}

// Start - starts web server
func (w *DefaultWebServer) Start(address string) {
	w.e.Logger.Fatal(w.e.Start(address))
}

// Stop - stops web server
func (w *DefaultWebServer) Stop() {
	_ = w.e.Close()
}
