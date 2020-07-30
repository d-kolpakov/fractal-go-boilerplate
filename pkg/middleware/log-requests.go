package middleware

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"time"
)

type RequestLog struct {
	Duration        string      `json:"duration"`
	Request         string      `json:"request"`
	RequestBody     string      `json:"request_body"`
	ResponseBody    string      `json:"response"`
	ResponseSC      int         `json:"response_status_code"`
	ResponseHeaders http.Header `json:"response_headers"`
}

type responseLogWriter struct {
	rw         http.ResponseWriter
	Body       []byte
	StatusCode int
}

func (w *responseLogWriter) Header() http.Header {
	return w.rw.Header()
}

func (w *responseLogWriter) Write(b []byte) (int, error) {
	s, err := w.rw.Write(b)

	if err == nil {
		w.Body = b
	}

	return s, err
}

func (w *responseLogWriter) WriteHeader(statusCode int) {
	w.StatusCode = statusCode
	w.rw.WriteHeader(statusCode)
}

// LogRequests Мидлвайр, который логирует все входящие запросы
//Этот мидлвайр должен использоватся для каждого роута
func (c *Controller) LogRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		t := time.Now()
		bodyBytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			c.L.NewLogEvent().Err(r.Context(), err, "")
		}
		r.Body.Close()
		r.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
		rw := &responseLogWriter{
			rw: w,
		}

		defer func(t time.Time, rw *responseLogWriter) {
			lBody := RequestLog{}

			byteDump, err := httputil.DumpRequest(r, false)

			if err != nil {
				c.L.NewLogEvent().Err(ctx, err, "")
			} else {
				lBody.Request = string(byteDump)

				if len(bodyBytes) > 0 {
					lBody.RequestBody = string(bodyBytes)
				}

				lBody.Duration = fmt.Sprint("%w", time.Since(t))

				lBody.ResponseHeaders = rw.Header()
				lBody.ResponseBody = string(rw.Body)
				lBody.ResponseSC = rw.StatusCode

				c.L.NewLogEvent().Debug(ctx, lBody)
			}
		}(t, rw)
		next.ServeHTTP(rw, r)
	})
}
