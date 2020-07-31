package routes

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/d-kolpakov/fractal-go-boilerplate/internal/handlers"
	"github.com/d-kolpakov/fractal-go-boilerplate/internal/server"
	"github.com/d-kolpakov/fractal-go-boilerplate/pkg/helpers/stats"
	"github.com/d-kolpakov/fractal-go-boilerplate/pkg/middleware"
	"github.com/d-kolpakov/fractal-go-boilerplate/pkg/monitoring"
	"github.com/d-kolpakov/logger"
	"github.com/go-chi/chi"
	"net/http"
)

type Routing struct {
	ServiceName      string
	serviceUrlPrefix string
	Stats            *stats.Stats
	L                *logger.Logger
	R                *chi.Mux
	Db               *sql.DB
	AppVersion       string
}

func (route *Routing) InitRouter() error {
	route.serviceUrlPrefix = "/" + route.ServiceName
	mc := middleware.Controller{
		Db:    route.Db,
		L:     route.L,
		SName: route.ServiceName,
	}

	r := chi.NewRouter()
	r.Use(mc.Recovery)
	route.R = r

	monitoring.AttachMonitoringUrls(route.serviceUrlPrefix, route.R)

	// Application endpoints
	r.Group(func(r chi.Router) {
		r.Use(mc.ContextRequestMiddleware)
		//r.Use(mc.ContextRequestMiddleware, mc.LogRequests)

		handler := handlers.Handler{
			L:           route.L,
			AppVersion:  route.AppVersion,
			ServiceName: route.ServiceName,
		}

		route.Get(route.serviceUrlPrefix, handler.HomeRouteHandler)
	})

	return nil
}

func (route *Routing) RegisterHandler(path, method string, handler server.WrappedHandlerFunc) {
	route.L.NewLogEvent().
		WithTag("process", "route_registration").
		Debug(context.Background(), fmt.Sprintf("registered %s %s route", method, path))

	wrappedHandler := server.NewHandlerWrapper(route.Stats, handler, route.L)
	switch method {
	case http.MethodGet:
		route.R.Get(path, wrappedHandler.Process)
	case http.MethodPost:
		route.R.Post(path, wrappedHandler.Process)
	case http.MethodDelete:
		route.R.Delete(path, wrappedHandler.Process)
	case http.MethodPatch:
		route.R.Patch(path, wrappedHandler.Process)
	case http.MethodPut:
		route.R.Put(path, wrappedHandler.Process)
	case http.MethodOptions:
		route.R.Options(path, wrappedHandler.Process)
	default:
		panic("unsupported method")
	}
}

func (route *Routing) Get(path string, handler server.WrappedHandlerFunc) {
	route.RegisterHandler(path, http.MethodGet, handler)
}

func (route *Routing) Post(path string, handler server.WrappedHandlerFunc) {
	route.RegisterHandler(path, http.MethodPost, handler)
}

func (route *Routing) Delete(path string, handler server.WrappedHandlerFunc) {
	route.RegisterHandler(path, http.MethodDelete, handler)
}

func (route *Routing) Patch(path string, handler server.WrappedHandlerFunc) {
	route.RegisterHandler(path, http.MethodPatch, handler)
}

func (route *Routing) Put(path string, handler server.WrappedHandlerFunc) {
	route.RegisterHandler(path, http.MethodPut, handler)
}

func (route *Routing) Options(path string, handler server.WrappedHandlerFunc) {
	route.RegisterHandler(path, http.MethodOptions, handler)
}
