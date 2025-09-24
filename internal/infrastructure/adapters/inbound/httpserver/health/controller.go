package health

import (
	"net/http"
	"sync/atomic"

	"github.com/go-chi/chi/v5"
)

type Controller struct {
	isShuttingDown *atomic.Bool
	readyCheck     func() bool
}

func NewController(isShuttingDown *atomic.Bool, readyCheck func() bool) *Controller {
	return &Controller{
		isShuttingDown: isShuttingDown,
		readyCheck:     readyCheck,
	}
}

// RegisterRoutes registers the health check endpoints to the given router
func (h *Controller) RegisterRoutes(router chi.Router) {
    router.Get("/health/live", h.Liveness)
    router.Get("/health/ready", h.Readiness)
}

// Liveness check - verifies if the process is running
//
// @Summary Liveness check
// @Description Returns 200 OK if the service is alive, 503 if shutting down
// @Tags Health
// @Produce plain
// @Success 200 {string} string "ok"
// @Failure 503 {string} string "shutting down"
// @Router /health/live [get]
func (h *Controller) Liveness(w http.ResponseWriter, r *http.Request) {
	if h.isShuttingDown.Load() {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte("shutting down"))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

// Readiness check - verifies if the service is ready to accept traffic
//
// @Summary Readiness check
// @Description Returns 200 OK if ready, 503 if shutting down or not ready
// @Tags Health
// @Produce plain
// @Success 200 {string} string "ready"
// @Failure 503 {string} string "shutting down / not ready"
// @Router /health/ready [get]
func (h *Controller) Readiness(w http.ResponseWriter, r *http.Request) {
	if h.isShuttingDown.Load() {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte("shutting down"))
		return
	}

	// Additional readiness checks (DB, cache, etc.)
	if h.readyCheck != nil && !h.readyCheck() {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte("not ready"))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ready"))
}

