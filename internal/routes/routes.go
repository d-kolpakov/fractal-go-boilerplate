package routes

import (
	"database/sql"
	"encoding/json"
	"github.com/d-kolpakov/fractal-go-boilerplate/internal/handlers"
	"github.com/d-kolpakov/fractal-go-boilerplate/internal/server"
	natsclient "github.com/d-kolpakov/fractal-go-boilerplate/pkg/helpers/natsclirnt"
	"github.com/d-kolpakov/fractal-go-boilerplate/pkg/helpers/stats"
	"github.com/d-kolpakov/fractal-go-boilerplate/pkg/middleware"
	"github.com/d-kolpakov/fractal-go-boilerplate/pkg/monitoring"
	"github.com/d-kolpakov/fractal-go-boilerplate/pkg/probs"
	"github.com/d-kolpakov/logger"
	"github.com/go-chi/chi"
	"net/http"
)

type Routing struct {
	ServiceName      string
	Stan             *natsclient.NatsConnection
	serviceUrlPrefix string
	Stats            *stats.Stats
	L                *logger.Logger
	R                *chi.Mux
	Db               *sql.DB
	AppVersion       string
	registry         []Registry
}

type Registry struct {
	Method string `json:"method"`
	Path   string `json:"path"`
	Policy string `json:"policy"`
}

const (
	Public  = "public"
	Private = "private"
	Root    = "root"

	RegistryQueue = "service-registry"
)

func (route *Routing) InitRouter() error {
	defer route.flushRoutes()
	route.serviceUrlPrefix = "/" + route.ServiceName
	mc := middleware.Controller{
		Db:    route.Db,
		L:     route.L,
		SName: route.ServiceName,
	}

	r := chi.NewRouter()
	route.R = r

	monitoring.AttachMonitoringUrls(route.serviceUrlPrefix, route.R)

	// Application endpoints
	r.Route(route.serviceUrlPrefix, func(r chi.Router) {
		r.Use(mc.ContextRequestMiddleware, mc.LogRequests)

		handler := handlers.Handler{
			L:           route.L,
			AppVersion:  route.AppVersion,
			ServiceName: route.ServiceName,
		}

		route.register(r, http.MethodGet, "/", Public, server.NewHandlerWrapper(route.Stats, handler.HomeRouteHandler, route.L).Process)
		route.register(r, http.MethodGet, "/private", Private, server.NewHandlerWrapper(route.Stats, handler.HomeRouteHandler, route.L).Process)
	})

	return nil
}

func (route *Routing) register(r chi.Router, method, path, policy string, handler http.HandlerFunc) {
	defer func(method, path, policy string) {
		route.registry = append(route.registry, Registry{
			Method: method,
			Path:   path,
			Policy: policy,
		})
	}(method, path, policy)

	r.MethodFunc(method, path, handler)
}

type ServiceRegistryMessage struct {
	ServiceName string     `json:"service_name"`
	Registry    []Registry `json:"registry"`
	Ready       bool       `json:"ready"`
}

func (route *Routing) flushRoutes() {
	msg := &ServiceRegistryMessage{
		ServiceName: route.ServiceName,
		Registry:    route.registry,
		Ready:       true,
	}

	bMsg, err := json.Marshal(msg)

	if err != nil {
		probs.SetReadinessError(err)
	}

	err = route.Stan.SendMessage(RegistryQueue, bMsg)

	if err != nil {
		probs.SetReadinessError(err)
	}
}

func (route *Routing) Deregister() {
	msg := &ServiceRegistryMessage{
		ServiceName: route.ServiceName,
		Ready:       false,
	}

	bMsg, err := json.Marshal(msg)

	if err != nil {
		probs.SetReadinessError(err)
	}

	err = route.Stan.SendMessage(RegistryQueue, bMsg)

	if err != nil {
		probs.SetReadinessError(err)
	}
}
