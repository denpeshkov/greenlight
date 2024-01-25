package http

import (
	"net/http"
	stdhttp "net/http"

	"github.com/denpeshkov/greenlight/internal/greenlight"
)

// Error responds with an error and status code from the error.
func (s *Server) Error(w http.ResponseWriter, r *http.Request, e error) {
	//logger := s.logger.With("method", r.Method, "path", r.URL.Path)
	//logger.Error("Error processing request", "error", err.Error(), )

	// In case of an error send a 500 Internal Server Error status code with an empty body
	if err := s.sendResponse(w, r, ErrorStatusCode(e), ErrorResponse{Err: greenlight.ErrorMessage(e)}, nil); err != nil {
		//logger.Error("sending error response to user", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

var codes = map[string]int{
	greenlight.ErrInternal: stdhttp.StatusInternalServerError,
	greenlight.ErrInvalid:  stdhttp.StatusBadRequest,
	greenlight.ErrNotFound: stdhttp.StatusNotFound,
}

// ErrorStatusCode returns the associated HTTP status code for a WTF error code.
func ErrorStatusCode(e error) int {
	code := greenlight.ErrorCode(e)
	if v, ok := codes[code]; ok {
		return v
	}
	return http.StatusInternalServerError
}

// ErrorResponse represents an error for an end user.
type ErrorResponse struct {
	// Err is an error message for an end user.
	Err string `json:"error,omitempty"`
}
