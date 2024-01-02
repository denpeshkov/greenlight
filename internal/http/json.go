package http

import (
	"encoding/json"
	"net/http"
)

func (s *Server) sendResponse(w http.ResponseWriter, r *http.Request, status int, v any) error {
	js, err := json.Marshal(v)
	if err != nil {
		return err
	}
	w.WriteHeader(status)
	w.Write(js)
	return nil
}

func (s *Server) readRequest(w http.ResponseWriter, r *http.Request, dst any) error {
	return json.NewDecoder(r.Body).Decode(&dst)
}
