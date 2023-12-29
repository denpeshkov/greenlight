package http

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func (s *Server) registerHealthCheckRoutes() {
	s.router.HandleFunc("GET /v1/healthcheck", s.healthCheckHandler)
}

func (s *Server) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	info := HealthInfo{"1.0", "UP"}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(info); err != nil {
		errResp := ErrorResponse{Msg: "Error processing request", err: fmt.Errorf("HealthInfo JSON encoding: %w", err)}
		s.Error(w, r, http.StatusInternalServerError, errResp)
	}
}

type HealthInfo struct {
	// App's version.
	Version string
	// Status of the app.
	Health string
}
