package middleware

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"time"
)

type RequestLog struct {
	Duration    time.Duration `json:"duration"`
	Request     string        `json:"request"`
	RequestBody string        `json:"request_body"`
	Response    string        `json:"response"`
}

// LogRequests Мидлвайр, который логирует все входящие запросы
//Этот мидлвайр должен использоватся для каждого роута
func (c *Controller) LogRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		t := time.Now()
		bodyBytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			c.L.ErrWithContext(r.Context(), err, "")
		}
		r.Body.Close()
		r.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

		defer func() {
			lBody := RequestLog{}

			byteDump, err := httputil.DumpRequest(r, false)

			if err != nil {
				c.L.ErrWithContext(ctx, err, "")
			} else {
				lBody.Request = string(byteDump)

				if len(bodyBytes) > 0 {
					lBody.RequestBody = string(bodyBytes)
				}

				lBody.Duration = time.Since(t)
				c.L.DebugWithContext(ctx, lBody)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
