package middleware

import (
	"context"
	"github.com/d-kolpakov/fractal-go-boilerplate/pkg/helpers/uuid"
	"github.com/d-kolpakov/logger"
	"net/http"
)

func (m *Controller) ContextRequestMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		ctx := m.newContextWithRequestID(r.Context(), r)
		ctx = m.newContextWithClientID(ctx, r)
		ctx = m.newContextWithServiceName(ctx, r)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *Controller) newContextWithRequestID(ctx context.Context, req *http.Request) context.Context {
	reqID := req.Header.Get("X-Request-ID")
	if reqID == "" {
		reqID = uuid.GenerateUUID()
	}

	var key logger.ContextUIDKey = "requestID"
	return context.WithValue(ctx, key, reqID)
}

func (m *Controller) newContextWithClientID(ctx context.Context, req *http.Request) context.Context {
	clientID := req.Header.Get("X-Client-ID")
	if clientID == "" {
		clientID = "cl" + uuid.GenerateUUID()
	}

	var key logger.ContextUIDKey = "clientID"
	return context.WithValue(ctx, key, clientID)
}

func (m *Controller) newContextWithServiceName(ctx context.Context, req *http.Request) context.Context {
	source := req.Header.Get("X-Fr-Source")

	var key logger.ContextUIDKey = "source"
	return context.WithValue(ctx, key, source)
}
