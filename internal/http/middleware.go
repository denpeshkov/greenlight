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

type notFoundResponseWriter struct {
	hijackResponseWriter
}

func (w *notFoundResponseWriter) Write(data []byte) (n int, err error) {
	if w.hijackResponseWriter.status == http.StatusNotFound {
		data, err = json.Marshal(ErrorResponse{Msg: http.StatusText(http.StatusNotFound)})
	}
	if err != nil {
		return 0, err
	}
	return w.hijackResponseWriter.Write(data)
}

// notFound returns a request handler that handles [http.StatusNotFound] status code.
func (s *Server) notFound(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hw := hijackResponseWriter{ResponseWriter: w, status: http.StatusOK}
		nfw := &notFoundResponseWriter{hijackResponseWriter: hw}
		h.ServeHTTP(nfw, r)
	})
}

type methodNotAllowedResponseWriter struct {
	hijackResponseWriter
}

func (w *methodNotAllowedResponseWriter) Write(data []byte) (n int, err error) {
	if w.hijackResponseWriter.status == http.StatusMethodNotAllowed {
		data, err = json.Marshal(ErrorResponse{Msg: http.StatusText(http.StatusMethodNotAllowed)})
	}
	if err != nil {
		return 0, err
	}
	return w.hijackResponseWriter.Write(data)
}

// methodNotAllowed returns a request handler that handles [http.StatusMethodNotAllowed] status code.
func (s *Server) methodNotAllowed(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hw := hijackResponseWriter{ResponseWriter: w, status: http.StatusOK}
		mrw := &methodNotAllowedResponseWriter{hijackResponseWriter: hw}
		h.ServeHTTP(mrw, r)
	})
}
