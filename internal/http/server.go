package http

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"time"

	greenlight "github.com/denpeshkov/greenlight/internal"
)

// Server represents an HTTP server.
type Server struct {
	MovieService greenlight.MovieService

	server *http.Server
	router *http.ServeMux
	logger *slog.Logger
}

// NewServer returns a new instance of [Server].
func NewServer(addr string) *Server {
	s := &Server{
		logger: newLogger(),
		server: &http.Server{},
		router: http.NewServeMux(),
	}
	s.server.Addr = addr

	s.registerHealthCheckHandlers()
	s.registerMovieHandlers()

	return s
}

// FIXME maybe use options pattern and remove NewServer()

// Open starts an HTTP server.
func (s *Server) Open(opts ...Option) error {
	s.server.Handler = s
	s.server.ErrorLog = slog.NewLogLogger(s.logger.Handler(), slog.LevelError)

	// Apply options
	for _, opt := range opts {
		opt(s)
	}

	err := s.server.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

// Close gracefully shuts down the server.
func (s *Server) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	return s.server.Shutdown(ctx)
}

// ServerHTTP handles an HTTP request.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h := s.notFound(s.router)
	h.ServeHTTP(w, r)
}

func newLogger() *slog.Logger {
	opts := slog.HandlerOptions{Level: slog.LevelDebug}
	handler := slog.NewJSONHandler(os.Stderr, &opts)

	logger := slog.New(handler).With("module", "http")

	return logger
}
