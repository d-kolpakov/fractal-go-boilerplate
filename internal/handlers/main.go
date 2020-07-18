package handlers

import (
	"fmt"
	"github.com/d-kolpakov/fractal-go-boilerplate/pkg/helpers/response"
	"github.com/d-kolpakov/logger"
	"net/http"
)

// Handler Хендлер дефолтных урлов
type Handler struct {
	L           *logger.Logger
	AppVersion  string
	ServiceName string
}

func (h *Handler) HomeRouteHandler(w http.ResponseWriter, r *http.Request) {
	h.L.NewLogEvent().WithRequest(r).LogWithContext(r.Context(), "HomeRouteHandler")

	response.JSON(w, http.StatusOK, fmt.Sprintf("Hello! This is %s. Version: %s", h.ServiceName, h.AppVersion))
}
