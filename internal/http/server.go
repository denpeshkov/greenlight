package http

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"time"
)

// Server represents an HTTP server.
type Server struct {
	// Logger to use.
	Logger *slog.Logger
	// Address to listen on, ":http" if empty.
	Addr string

	server *http.Server
	router *http.ServeMux
}

// NewServer returns a new HTTP server.
func NewServer() *Server {
	s := &Server{
		Logger: NewLogger(),
		server: &http.Server{},
		router: http.NewServeMux(),
	}

	s.registerHealthCheckRoutes()
	s.registerMovieRoutes()

	return s
}

// Start starts a HTTP server.
func (s *Server) Start() error {
	s.server.Addr = s.Addr
	s.server.Handler = http.HandlerFunc(s.ServeHTTP)
	s.server.IdleTimeout = time.Minute
	s.server.ReadTimeout = 10 * time.Second
	s.server.WriteTimeout = 30 * time.Second
	s.server.ErrorLog = slog.NewLogLogger(s.Logger.Handler(), slog.LevelError)

	s.Logger.Debug("starting HTTP server", "addr", s.Addr)

	err := s.server.ListenAndServe()

	s.Logger.Error("HTTP server startup", "error", err)

	return err
}

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
		data, err = json.Marshal(ErrorResponse{Msg: http.StatusText(http.StatusNotFound), err: nil})
		if err != nil {
			return 0, err
		}
	case http.StatusMethodNotAllowed:
		data, err = json.Marshal(ErrorResponse{Msg: http.StatusText(http.StatusMethodNotAllowed), err: nil})
		if err != nil {
			return 0, err
		}
	}
	return w.ResponseWriter.Write(data)
}

// ServeHTTP handles an HTTP request.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	sR := &hijackResponseWriter{ResponseWriter: w, status: http.StatusOK}
	s.router.ServeHTTP(sR, r)
}

func NewLogger() *slog.Logger {
	opts := slog.HandlerOptions{AddSource: true}
	handler := slog.NewJSONHandler(os.Stderr, &opts)

	logger := slog.New(handler).With("module", "http")

	return logger
}
