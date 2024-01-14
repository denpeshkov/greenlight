package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/textproto"
)

// sendResponse sends a JSON response with a given status.
// In case of an error, response (and status) is not send and error is returned.
func (s *Server) sendResponse(w http.ResponseWriter, r *http.Request, status int, resp any, headers http.Header) error {
	js, err := json.Marshal(resp)
	if err != nil {
		return fmt.Errorf("marshalling response to JSON: %w", err)
	}
	for k, v := range headers {
		k = textproto.CanonicalMIMEHeaderKey(k)
		w.Header()[k] = v
	}
	w.WriteHeader(status)
	_, err = w.Write(js)
	if err != nil {
		return fmt.Errorf("write data to HTTP connection: %w", err)
	}
	return nil
}

// readRequest decodes a JSON request body to the dst value.
func (s *Server) readRequest(w http.ResponseWriter, r *http.Request, dst any) error {
	err := json.NewDecoder(r.Body).Decode(&dst)
	if err != nil {
		return fmt.Errorf("unmarshalling request to JSON: %w", err)
	}
	return nil
}
