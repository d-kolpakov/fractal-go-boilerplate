package middleware

import (
	"context"
	"net/http"
)

func (c *Controller) Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		defer func(ctx context.Context) {
			if r := recover(); r != nil {
				c.L.AlertWithContext(ctx, r)
			}
		}(ctx)

		next.ServeHTTP(w, r)
	})
}
