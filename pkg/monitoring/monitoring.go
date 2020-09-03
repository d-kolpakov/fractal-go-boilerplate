package monitoring

import (
	"github.com/d-kolpakov/fractal-go-boilerplate/pkg/probs"
	"github.com/go-chi/chi"
	"net/http"
	"net/http/pprof"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// AttachMonitoringUrls Подключение вывода систем мониторинга
func AttachMonitoringUrls(prefix string, router chi.Router) {
	attachMetrics(prefix, router)
	attachProfiler(prefix, router)
	attachProbs(prefix, router)
}

func attachProfiler(prefix string, router chi.Router) {
	router.HandleFunc(prefix+"/debug/pprof/", pprof.Index)
	router.HandleFunc(prefix+"/debug/pprof/cmdline", pprof.Cmdline)
	router.HandleFunc(prefix+"/debug/pprof/profile", pprof.Profile)
	router.HandleFunc(prefix+"/debug/pprof/symbol", pprof.Symbol)

	// Manually add support for paths linked to by index page at /debug/pprof/
	router.Handle(prefix+"/debug/pprof/goroutine", pprof.Handler("goroutine"))
	router.Handle(prefix+"/debug/pprof/heap", pprof.Handler("heap"))
	router.Handle(prefix+"/debug/pprof/threadcreate", pprof.Handler("threadcreate"))
	router.Handle(prefix+"/debug/pprof/block", pprof.Handler("block"))
	router.Handle(prefix+"/debug/pprof/mutex", pprof.Handler("mutex"))
	router.Handle(prefix+"/debug/pprof/allocs", pprof.Handler("allocs"))
	router.Handle(prefix+"/debug/pprof/trace", pprof.Handler("trace"))
}

func attachMetrics(prefix string, router chi.Router) {
	router.HandleFunc(prefix+"/metrics/", func(w http.ResponseWriter, r *http.Request) {
		promhttp.Handler().ServeHTTP(w, r)
	})
}

func attachProbs(prefix string, router chi.Router) {
	router.HandleFunc(prefix+"/_readiness/", probs.Readiness)
	router.HandleFunc(prefix+"/_liveness/", probs.Liveness)
}
