package http

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"

	"github.com/denpeshkov/greenlight/internal/greenlight"
	"github.com/denpeshkov/greenlight/internal/multierr"
)

// Server represents an HTTP server.
type Server struct {
	movieService greenlight.MovieService
	userService  greenlight.UserService
	authService  *greenlight.AuthService

	opts options

	server *http.Server
	router *http.ServeMux
	logger *slog.Logger
}

// NewServer returns a new instance of [Server].
func NewServer(addr string, movieService greenlight.MovieService, userService greenlight.UserService, authService *greenlight.AuthService, opts ...Option) *Server {
	s := &Server{
		movieService: movieService,
		userService:  userService,
		authService:  authService,
		server:       &http.Server{},
		router:       http.NewServeMux(),
		logger:       newLogger(),
	}
	s.server.Addr = addr

	// Apply options
	for _, opt := range opts {
		opt(&s.opts)
	}

	s.registerHealthCheckHandlers()
	s.registerMovieHandlers()
	s.registerUserHandlers()
	s.registerAuthHandlers()

	return s
}

// Open starts an HTTP server.
func (s *Server) Open() (err error) {
	defer multierr.Wrap(&err, "http.Server.Start")

	s.server.Handler = s.recoverPanic(s.rateLimit(s.notFound(s.methodNotAllowed(s.router))))
	s.server.ErrorLog = slog.NewLogLogger(s.logger.Handler(), slog.LevelError)

	s.server.IdleTimeout = s.opts.idleTimeout
	s.server.ReadTimeout = s.opts.readTimeout
	s.server.WriteTimeout = s.opts.writeTimeout

	err = s.server.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

// Close gracefully shuts down the server.
func (s *Server) Close() (err error) {
	defer multierr.Wrap(&err, "http.Server.Close")

	ctx, cancel := context.WithTimeout(context.Background(), s.opts.shutdownTimeout)
	defer cancel()

	err = s.server.Shutdown(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (s *Server) handlerFunc(h func(http.ResponseWriter, *http.Request) error) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := h(w, r); err != nil {
			s.Error(w, r, err)
		}
	})
}

func newLogger() *slog.Logger {
	opts := slog.HandlerOptions{Level: slog.LevelDebug}
	handler := slog.NewJSONHandler(os.Stderr, &opts)

	logger := slog.New(handler)

	return logger
}
