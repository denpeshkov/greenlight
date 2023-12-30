package http

import (
	"encoding/json"
	"net/http"
)

func (s *Server) registerHealthCheckHandlers() {
	s.router.HandleFunc("GET /v1/healthcheck", s.handleHealthCheck)
}

func (s *Server) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	info := HealthInfo{"1.0", "UP"}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(info); err != nil {
		errResp := ErrorResponse{Msg: "Error processing request", err: err}
		s.Error(w, r, http.StatusInternalServerError, errResp)
	}
}

// Application information.
type HealthInfo struct {
	// App's version.
	Version string `json:"version"`
	// Status of the app.
	Status string `json:"status"`
}
