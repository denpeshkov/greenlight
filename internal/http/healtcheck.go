package http

import (
	"fmt"
	"net/http"
)

func (s *Server) registerHealthCheckHandlers() {
	s.router.HandleFunc("GET /v1/healthcheck", s.handleHealthCheck)
}

// handleHealthCheck handles requests to get application information (status).
func (s *Server) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	op := "http.Server.handleHealthCheck"

	info := HealthInfo{"1.0", "UP"}
	w.Header().Set("Content-Type", "application/json")
	if err := s.sendResponse(w, r, http.StatusOK, info, nil); err != nil {
		s.Error(w, r, fmt.Errorf("%s: %w", op, err))
		return
	}
}

// Application information.
type HealthInfo struct {
	// App's version.
	Version string `json:"version"`
	// Status of the app.
	Status string `json:"status"`
}
