package server

import (
	"fmt"
	"github.com/d-kolpakov/fractal-go-boilerplate/pkg/helpers/response"
	"github.com/d-kolpakov/fractal-go-boilerplate/pkg/helpers/stats"
	"github.com/d-kolpakov/logger"
	"net/http"
	"time"
)

type Response struct {
	Response     interface{}
	RealResponse interface{}
	StatusCode   int
	Err          error
}

type WrappedHandlerFunc func(w http.ResponseWriter, r *http.Request) Response

type handlerWrapper struct {
	stats       *stats.Stats
	handlerFunc WrappedHandlerFunc
	l           *logger.Logger
}

func NewHandlerWrapper(stats *stats.Stats, handlerFunc WrappedHandlerFunc, l *logger.Logger) *handlerWrapper {
	return &handlerWrapper{
		stats:       stats,
		handlerFunc: handlerFunc,
		l:           l,
	}
}

func (h *handlerWrapper) Process(w http.ResponseWriter, r *http.Request) {
	t := time.Now()
	url := r.URL.Path

	ctx := r.Context()
	resp := h.handlerFunc(w, r)

	if resp.StatusCode != http.StatusNoContent {
		if resp.RealResponse == nil {
			h.l.NewLogEvent().
				WithTag("kind", "response_err").
				WithTag("process", "request_handling").
				Error(ctx, fmt.Errorf("response must be []byte, nil has been given"))
		} else {
			if bResp, ok := resp.Response.([]byte); ok {
				response.WriteBody(w, resp.StatusCode, bResp)
			} else {
				h.l.NewLogEvent().
					WithTag("kind", "response_err").
					WithTag("process", "request_handling").
					Error(ctx, fmt.Errorf("response must be []byte, %v has been given", resp.Response))
			}
		}
	} else {
		response.WriteBody(w, http.StatusNoContent, nil)
	}

	h.statDuration(url, time.Since(t))
	h.statResponseCode(url, resp.StatusCode)
}

func (h *handlerWrapper) statDuration(url string, t time.Duration) {
	intVal := int64(t)
	h.stats.InsertStat("server.response.duration", &url, nil, &intVal, nil)
}

func (h *handlerWrapper) statResponseCode(url string, status int) {
	intVal := int64(status)
	h.stats.InsertStat("server.response.status", &url, nil, &intVal, nil)
}
