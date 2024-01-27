package http

import (
	"encoding/json"
	"net/http"
)

// hijackResponseWriter records status of the HTTP response.
type hijackResponseWriter struct {
	http.ResponseWriter
	status int
}

func (w *hijackResponseWriter) WriteHeader(statusCode int) {
	w.status = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *hijackResponseWriter) Write(data []byte) (n int, err error) {
	switch w.status {
	case http.StatusNotFound:
		data, err = json.Marshal(ErrorResponse{Msg: http.StatusText(http.StatusNotFound)})
	case http.StatusMethodNotAllowed:
		data, err = json.Marshal(ErrorResponse{Msg: http.StatusText(http.StatusMethodNotAllowed)})
	}
	if err != nil {
		return 0, err
	}
	return w.ResponseWriter.Write(data)
}

// notFound returns a request handler that handles [http.StatusNotFound] and [http.StatusMethodNotAllowed] status codes.
func (s *Server) notFound(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hW := &hijackResponseWriter{ResponseWriter: w, status: http.StatusOK}
		h.ServeHTTP(hW, r)
	})
}
