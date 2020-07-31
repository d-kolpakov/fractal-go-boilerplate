package server

import (
	"context"
	"fmt"
	"github.com/d-kolpakov/fractal-go-boilerplate/pkg/helpers/response"
	"github.com/d-kolpakov/fractal-go-boilerplate/pkg/helpers/stats"
	"github.com/d-kolpakov/logger"
	"github.com/dhnikolas/configo"
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
	ctxWithoutTimeout := r.Context()
	ctx, done := context.WithTimeout(ctxWithoutTimeout, time.Duration(configo.EnvInt("app-request-timeout", 20))*time.Second)
	resp := Response{}
	go func(ctx, ctxWithoutTimeout context.Context) {
		defer func(ctx context.Context, r *http.Request, url string) {
			if rec := recover(); rec != nil {
				h.l.NewLogEvent().
					WithTag("kind", "panic").
					WithTag("process", "request_handling").
					WithRequest(r).
					Alert(ctx, rec)
				h.statPanic(ctxWithoutTimeout, url)
			}
		}(ctxWithoutTimeout, r, url)

		resp = h.handlerFunc(w, r)
		done()
	}(ctx, ctxWithoutTimeout)

	status := resp.StatusCode
	select {
	case _ = <-ctx.Done():
		deadline, _ := ctx.Deadline()
		if time.Now().After(deadline) {
			resp := fmt.Sprintf(`{"error":"timeout", "detail":"request timeout %ds"}`, configo.EnvInt("app-request-timeout", 20))
			h.l.NewLogEvent().
				WithTag("kind", "timeout").
				WithTag("process", "request_handling").
				WithRequest(r).
				Error(ctx, "request timeout")

			status = http.StatusBadRequest
			h.statTimeout(ctxWithoutTimeout, url)
			response.WriteBody(w, http.StatusBadRequest, []byte(resp))
		} else {
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
		}

		h.statDuration(ctxWithoutTimeout, url, time.Since(t))
		h.statResponseCode(ctxWithoutTimeout, url, status)
	}
}

func (h *handlerWrapper) statDuration(ctx context.Context, url string, t time.Duration) {
	intVal := int64(t)
	h.stats.InsertStat("server.response.duration", &url, nil, &intVal, nil, h.getXFrSourceFromCtx(ctx), h.getRequestIDFromCtx(ctx))
}

func (h *handlerWrapper) statResponseCode(ctx context.Context, url string, status int) {
	intVal := int64(status)
	h.stats.InsertStat("server.response.status", &url, nil, &intVal, nil, h.getXFrSourceFromCtx(ctx), h.getRequestIDFromCtx(ctx))
}

func (h *handlerWrapper) statTimeout(ctx context.Context, url string) {
	h.stats.InsertStat("server.response.timeout", &url, nil, nil, nil, h.getXFrSourceFromCtx(ctx), h.getRequestIDFromCtx(ctx))
}

func (h *handlerWrapper) statPanic(ctx context.Context, url string) {
	h.stats.InsertStat("server.panic", &url, nil, nil, nil, h.getXFrSourceFromCtx(ctx), h.getRequestIDFromCtx(ctx))
}

func (h *handlerWrapper) getXFrSourceFromCtx(ctx context.Context) *string {
	var key logger.ContextUIDKey = "source"
	var res *string
	source := ctx.Value(key)
	if source != nil {
		sourceString, ok := source.(string)
		if ok {
			res = &sourceString
		}
	}

	return res
}

func (h *handlerWrapper) getRequestIDFromCtx(ctx context.Context) *string {
	var key logger.ContextUIDKey = "requestID"
	var res *string
	id := ctx.Value(key)
	if id != nil {
		idString, ok := id.(string)
		if ok {
			res = &idString
		}
	}

	return res
}
