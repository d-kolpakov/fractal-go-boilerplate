package middleware

import (
	"context"
	"net/http"
)

func (c *Controller) Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		defer func(ctx context.Context, r *http.Request) {
			if rec := recover(); rec != nil {
				c.L.NewLogEvent().WithRequest(r).Alert(ctx, rec)
			}
		}(ctx, r)

		next.ServeHTTP(w, r)
	})
}
