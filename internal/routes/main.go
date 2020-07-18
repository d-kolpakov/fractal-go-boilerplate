package routes

import (
	"database/sql"
	"github.com/d-kolpakov/fractal-go-boilerplate/internal/handlers"
	"github.com/d-kolpakov/fractal-go-boilerplate/pkg/middleware"
	"github.com/d-kolpakov/fractal-go-boilerplate/pkg/monitoring"
	"github.com/d-kolpakov/logger"
	"github.com/go-chi/chi"
)

type Routing struct {
	ServiceName      string
	serviceUrlPrefix string
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

		r.Get(route.serviceUrlPrefix, handler.HomeRouteHandler)
	})

	return nil
}
