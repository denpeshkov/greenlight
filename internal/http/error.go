package http

import (
	"encoding/json"
	"net/http"
)

// Error responds with an error and specified status code.
func (s *Server) Error(w http.ResponseWriter, r *http.Request, statusCode int, errResp ErrorResponse) {
	s.logger.Error(errResp.Msg, "method", r.Method, "path", r.URL.Path, "error", errResp.err)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(errResp); err != nil {
		s.logger.Error("error processing an error response", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// ErrorResponse represents an error for an end user.
type ErrorResponse struct {
	// Msg is an error message for an end user.
	Msg string `json:"error,omitempty"`
	err error  `json:"-"`
}
