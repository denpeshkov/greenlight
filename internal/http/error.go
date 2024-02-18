package http

import (
	"errors"
	"net/http"

	"github.com/denpeshkov/greenlight/internal/greenlight"
)

// Error responds with an error and status code from the error.
func (s *Server) Error(w http.ResponseWriter, r *http.Request, e error) {
	code := ErrorStatusCode(e)
	if code == http.StatusInternalServerError {
		s.logger.Error("Error processing request", "method", r.Method, "path", r.URL.Path, "error", e.Error())
	}
	errResp := ErrorBody(e)

	if err := s.sendResponse(w, r, code, errResp, nil); err != nil {
		s.logger.Error("Sending error response", "error", err, "error_resp", errResp)
		// In case of an error send a 500 Internal Server Error status code with an empty body
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func ErrorStatusCode(err error) int {
	switch {
	case errors.As(err, new(*greenlight.NotFoundError)):
		return http.StatusNotFound
	case errors.As(err, new(*greenlight.InvalidError)):
		return http.StatusUnprocessableEntity
	case errors.As(err, new(*greenlight.ConflictError)):
		return http.StatusConflict
	case errors.As(err, new(*greenlight.RateLimitError)):
		return http.StatusTooManyRequests
	case errors.As(err, new(*greenlight.UnauthorizedError)):
		return http.StatusUnauthorized
	case errors.As(err, new(*greenlight.InternalError)):
		fallthrough
	default:
		return http.StatusInternalServerError
	}
}

func ErrorBody(err error) any {
	var nfErr *greenlight.NotFoundError
	var invErr *greenlight.InvalidError
	var intErr *greenlight.InternalError
	var cftErr *greenlight.ConflictError
	var rateErr *greenlight.RateLimitError
	var unErr *greenlight.UnauthorizedError

	switch {
	case errors.As(err, &nfErr):
		return ErrorResponse{Msg: nfErr.Msg}
	case errors.As(err, &invErr):
		vs := invErr.Violations()
		m := make(map[string]string, len(vs))
		for k, v := range vs {
			m[k] = v.Error()
		}
		return ValidationErrorResponse{Msg: invErr.Msg, Fields: m}
	case errors.As(err, &cftErr):
		return ErrorResponse{Msg: cftErr.Msg}
	case errors.As(err, &rateErr):
		return ErrorResponse{Msg: rateErr.Msg}
	case errors.As(err, &unErr):
		return ErrorResponse{Msg: unErr.Msg}
	case errors.As(err, &intErr):
		fallthrough
	default:
		return ErrorResponse{Msg: "Server error."}
	}
}

// ErrorResponse represents an error for an end user.
type ErrorResponse struct {
	// Msg is an error message for an end user.
	Msg string `json:"message,omitempty"`
}

type ValidationErrorResponse struct {
	Msg    string            `json:"message,omitempty"`
	Fields map[string]string `json:"invalid_fields,omitempty"`
}
