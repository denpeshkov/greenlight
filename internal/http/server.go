package http

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"

	"github.com/denpeshkov/greenlight/internal/greenlight"
)

// Server represents an HTTP server.
type Server struct {
	MovieService greenlight.MovieService

	opts options

	server *http.Server
	router *http.ServeMux
	logger *slog.Logger
}

// NewServer returns a new instance of [Server].
func NewServer(addr string, opts ...Option) *Server {
	s := &Server{
		server: &http.Server{},
		router: http.NewServeMux(),
		logger: newLogger(),
	}
	s.server.Addr = addr

	// Apply options
	for _, opt := range opts {
		opt(&s.opts)
	}

	s.registerHealthCheckHandlers()
	s.registerMovieHandlers()

	return s
}

// Start starts an HTTP server.
func (s *Server) Start() error {
	s.server.Handler = s
	s.server.ErrorLog = slog.NewLogLogger(s.logger.Handler(), slog.LevelError)

	s.server.IdleTimeout = s.opts.idleTimeout
	s.server.ReadTimeout = s.opts.readTimeout
	s.server.WriteTimeout = s.opts.writeTimeout

	err := s.server.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

// Close gracefully shuts down the server.
func (s *Server) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), s.opts.shutdownTimeout)
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

	logger := slog.New(handler)

	return logger
}
