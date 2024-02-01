package http

import (
	"context"
	"errors"
	"fmt"
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
	op := "http.Server.Start"

	handler := s.recoverPanic(s.rateLimit(s.notFound(s.methodNotAllowed(s.router))))

	s.server.Handler = handler
	s.server.ErrorLog = slog.NewLogLogger(s.logger.Handler(), slog.LevelError)

	s.server.IdleTimeout = s.opts.idleTimeout
	s.server.ReadTimeout = s.opts.readTimeout
	s.server.WriteTimeout = s.opts.writeTimeout

	err := s.server.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

// Close gracefully shuts down the server.
func (s *Server) Close() error {
	op := "http.Server.Close"

	ctx, cancel := context.WithTimeout(context.Background(), s.opts.shutdownTimeout)
	defer cancel()

	err := s.server.Shutdown(ctx)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func newLogger() *slog.Logger {
	opts := slog.HandlerOptions{Level: slog.LevelDebug}
	handler := slog.NewJSONHandler(os.Stderr, &opts)

	logger := slog.New(handler)

	return logger
}
