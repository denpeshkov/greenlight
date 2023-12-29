package http

import "net/http"

func (s *Server) MethodNotAllowed(w http.ResponseWriter, r *http.Request) {
	s.Error(w, r, http.StatusMethodNotAllowed, ErrorResponse{Msg: "Method Not Allowed", err: nil})
}
