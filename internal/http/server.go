package http

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	greenlight "github.com/denpeshkov/greenlight/internal"
)

// Server represents an HTTP server.
type Server struct {
	// Address to listen on, ":http" if empty.
	Addr string

	MovieService greenlight.MovieService

	server *http.Server
	router *http.ServeMux
	logger *slog.Logger
}

// NewServer returns a new instance of [Server].
func NewServer() *Server {
	s := &Server{
		logger: newLogger(),
		server: &http.Server{},
		router: http.NewServeMux(),
	}

	s.registerHealthCheckHandlers()
	s.registerMovieHandlers()

	return s
}

// Start starts a HTTP server.
func (s *Server) Start() error {
	s.server.Addr = s.Addr
	s.server.Handler = s
	s.server.IdleTimeout = time.Minute
	s.server.ReadTimeout = 10 * time.Second
	s.server.WriteTimeout = 30 * time.Second
	s.server.ErrorLog = slog.NewLogLogger(s.logger.Handler(), slog.LevelError)

	s.logger.Debug("starting HTTP server", "addr", s.Addr)

	err := s.server.ListenAndServe()

	s.logger.Error("HTTP server startup", "error", err)

	return err
}

// ServerHTTP handles an HTTP request.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h := s.notFound(s.router)
	h.ServeHTTP(w, r)
}

func newLogger() *slog.Logger {
	opts := slog.HandlerOptions{AddSource: true}
	handler := slog.NewJSONHandler(os.Stderr, &opts)

	logger := slog.New(handler).With("module", "http")

	return logger
}
