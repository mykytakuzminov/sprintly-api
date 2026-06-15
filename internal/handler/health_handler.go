package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/mykytakuzminov/sprintly-api/internal/domain"
	"go.uber.org/zap"
)

type HealthHandler struct {
	svc    domain.HealthService
	logger *zap.SugaredLogger
}

func NewHealthHandler(svc domain.HealthService, logger *zap.SugaredLogger) *HealthHandler {
	return &HealthHandler{svc: svc, logger: logger}
}

func (h *HealthHandler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", h.Health)

	return r
}

// Health godoc
// @Summary     Health check
// @Description Returns the health status of the service including database and Redis connectivity.
// @Tags        health
// @Produce     json
// @Success     200 {object} domain.HealthStats "Service is healthy"
// @Failure     503 {object} domain.HealthStats "Service is degraded — one or more dependencies are unavailable"
// @Router      /health [get]
func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	traceID := getTraceID(r, h.logger)

	healthStats := h.svc.Check(r.Context())

	code := http.StatusOK
	if healthStats.Status == "degraded" {
		code = http.StatusServiceUnavailable
	}

	logSuccess(h.logger, traceID, "health checked", "health_status", healthStats.Status)
	successResponse(w, code, healthStats)
}
