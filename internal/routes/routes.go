package routes

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/d-kolpakov/fractal-go-boilerplate/internal/handlers"
	"github.com/d-kolpakov/fractal-go-boilerplate/internal/server"
	natsclient "github.com/d-kolpakov/fractal-go-boilerplate/pkg/helpers/natsclient"
	"github.com/d-kolpakov/fractal-go-boilerplate/pkg/helpers/stats"
	"github.com/d-kolpakov/fractal-go-boilerplate/pkg/middleware"
	"github.com/d-kolpakov/fractal-go-boilerplate/pkg/monitoring"
	"github.com/d-kolpakov/fractal-go-boilerplate/pkg/probs"
	"github.com/d-kolpakov/logger"
	"github.com/go-chi/chi"
	"math/rand"
	"net/http"
	"time"
)

type Routing struct {
	ServiceName      string
	Stan             *natsclient.NatsConnection
	serviceUrlPrefix string
	Stats            *stats.Stats
	L                *logger.Logger
	R                *chi.Mux
	Db               *sql.DB
	Port             int
	AppVersion       string
	instanceHash     string
	registry         []Registry
}

type Registry struct {
	Method    string `json:"method"`
	Path      string `json:"path"`
	ProxyPath string `json:"proxy_path"`
	Policy    string `json:"policy"`
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

		route.register(r, http.MethodGet, "/"+route.ServiceName, "/", Public, server.NewHandlerWrapper(route.Stats, handler.HomeRouteHandler, route.L).Process)
		route.register(r, http.MethodGet, "/"+route.ServiceName+"/private", "/private", Private, server.NewHandlerWrapper(route.Stats, handler.HomeRouteHandler, route.L).Process)
	})

	return nil
}

func (route *Routing) register(r chi.Router, method, path, proxyPath, policy string, handler http.HandlerFunc) {
	defer func(method, path, policy string) {
		route.registry = append(route.registry, Registry{
			Method:    method,
			Path:      path,
			ProxyPath: proxyPath,
			Policy:    policy,
		})
	}(method, path, policy)

	r.MethodFunc(method, proxyPath, handler)
}

type ServiceRegistryMessage struct {
	ServiceName string     `json:"service_name"`
	BaseHst     string     `json:"base_host"`
	Registry    []Registry `json:"registry"`
	Ready       bool       `json:"ready"`
	Hash        string     `json:"hash"`
	UnixNano    int64      `json:"unix_nano"`
}

func (route *Routing) flushRoutes() {
	rand.Seed(time.Now().UnixNano())
	hash := fmt.Sprintf(route.ServiceName+":%d", rand.Uint64())

	msg := &ServiceRegistryMessage{
		ServiceName: route.ServiceName,
		BaseHst:     fmt.Sprintf("http://127.0.0.1:%d", route.Port),
		Registry:    route.registry,
		Ready:       true,
		UnixNano:    time.Now().UnixNano(),
		Hash:        hash,
	}
	route.instanceHash = hash

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
		UnixNano:    time.Now().UnixNano(),
		Hash:        route.instanceHash,
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
