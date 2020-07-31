package middleware

import (
	"net/http"
)

func (c *Controller) CheckXSource(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		source := r.Header.Get("X-Fr-Source")
		if source == "" {
			resp := `{"error":"Empty X-Fr-Source header"}`
			w.Write([]byte(resp))
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		next.ServeHTTP(w, r)
	})
}
