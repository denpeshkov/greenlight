package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/textproto"
	"strings"

	"github.com/denpeshkov/greenlight/internal/greenlight"
)

// sendResponse sends a JSON response with a given status.
// In case of an error, response (and status) is not send and error is returned.
func (s *Server) sendResponse(w http.ResponseWriter, r *http.Request, status int, resp any, headers http.Header) error {
	op := "http.Server.sendResponse"

	js, err := json.Marshal(resp)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	for k, v := range headers {
		k = textproto.CanonicalMIMEHeaderKey(k)
		w.Header()[k] = v
	}
	w.WriteHeader(status)
	_, err = w.Write(js)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

// readRequest decodes a JSON request body to the dst value.
func (s *Server) readRequest(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, s.opts.maxRequestBody)

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(&dst); err != nil {
		switch {
		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return greenlight.NewInvalidError("JSON body contains unknown key %s", fieldName)
		default:
			return greenlight.NewInvalidError("JSON body format is incorrect")
		}
	}
	return nil
}
