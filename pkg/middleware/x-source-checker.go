package middleware

import (
	"context"
	"github.com/d-kolpakov/logger"
	"net/http"
)

func (c *Controller) CheckXSource(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		source := r.Header.Get("X-Fr-Source")
		if source == "" {
			resp := `{"error":"\Empty X-Fr-Source header"}`
			w.Write([]byte(resp))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		var key logger.ContextUIDKey = "source"

		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), key, source)))
	})
}
