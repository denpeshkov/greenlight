package http

import (
	"expvar"
	"net/http"

	"github.com/denpeshkov/greenlight/internal/greenlight"
	"github.com/denpeshkov/greenlight/internal/multierr"
)

func (s *Server) registerHealthCheckHandlers() {
	s.router.Handle("GET /v1/healthcheck", s.handlerFunc(s.handleHealthCheck))
	s.router.Handle("GET /v1/metrics", expvar.Handler())
}

// handleHealthCheck handles requests to get application information (status).
func (s *Server) handleHealthCheck(w http.ResponseWriter, r *http.Request) (err error) {
	defer multierr.Wrap(&err, "http.Server.handleHealthCheck")

	info := HealthInfo{greenlight.AppVersion, "up"}
	w.Header().Set("Content-Type", "application/json")
	if err := s.sendResponse(w, r, http.StatusOK, info, nil); err != nil {
		return err
	}
	return nil
}

// Application information.
type HealthInfo struct {
	// App's version.
	Version string `json:"version"`
	// Status of the app.
	Status string `json:"status"`
}
