package http

import (
	"fmt"
	"net/http"
)

// Error responds with an error and specified status code.
func (s *Server) Error(w http.ResponseWriter, r *http.Request, statusCode int, errResp ErrorResponse) {
	logger := s.logger.With("method", r.Method, "path", r.URL.Path)
	logger.Error(fmt.Sprintf("end user error message: %s", errResp.Msg), "error", errResp.err)

	// In case of an error send a 500 Internal Server Error status code with an empty body
	if err := s.sendResponse(w, r, statusCode, errResp); err != nil {
		logger.Error("sending error response to user", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// ErrorResponse represents an error for an end user.
type ErrorResponse struct {
	// Msg is an error message for an end user.
	Msg string `json:"error_message,omitempty"`
	err error  `json:"-"`
}
