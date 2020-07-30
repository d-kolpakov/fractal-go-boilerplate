package handlers

import (
	"bytes"
	"fmt"
	"github.com/d-kolpakov/fractal-go-boilerplate/pkg/helpers/response"
	"github.com/d-kolpakov/logger"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
)

// Handler Хендлер дефолтных урлов
type Handler struct {
	L           *logger.Logger
	AppVersion  string
	ServiceName string
}

func (h *Handler) HomeRouteHandler(w http.ResponseWriter, r *http.Request) {
	h.L.NewLogEvent().WithRequest(r).Log(r.Context(), "HomeRouteHandler")

	resp := `{"msg":"%s","dump":"%s","status":%d}`
	msg := fmt.Sprintf("Hello! This is %s. Version: %s", h.ServiceName, h.AppVersion)
	respBytes := []byte(fmt.Sprintf(resp, msg, h.formRequest(r), http.StatusOK))
	response.WriteBody(w, http.StatusOK, respBytes)
}

func (h *Handler) formRequest(r *http.Request) string {
	res := ""

	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return res
	}

	r.Body.Close()
	r.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
	byteDump, err := httputil.DumpRequest(r, false)

	if err != nil {
		return res
	}

	res = string(byteDump)

	return res
}
