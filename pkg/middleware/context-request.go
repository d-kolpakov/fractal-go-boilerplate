package middleware

import (
	"context"
	"github.com/d-kolpakov/fractal-go-boilerplate/pkg/helpers/uuid"
	"net/http"
)

func (m *Controller) ContextRequestMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		ctx := newContextWithRequestID(r.Context(), r)
		ctx = newContextWithClientID(ctx, r)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func newContextWithRequestID(ctx context.Context, req *http.Request) context.Context {
	reqID := req.Header.Get("X-Request-ID")
	if reqID == "" {
		reqID = uuid.GenerateUUID()
	}

	return context.WithValue(ctx, "requestID", reqID)
}

func newContextWithClientID(ctx context.Context, req *http.Request) context.Context {
	clientID := req.Header.Get("X-Client-ID")
	if clientID == "" {
		clientID = "cl" + uuid.GenerateUUID()
	}

	return context.WithValue(ctx, "clientID", clientID)
}
