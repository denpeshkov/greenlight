package http

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/julienschmidt/httprouter"
)

// Server represents an HTTP server.
type Server struct {
	// Logger to use.
	Logger *slog.Logger
	// Address to listen on, ":http" if empty.
	Addr string

	server *http.Server
	router *httprouter.Router
}

// NewServer returns a new HTTP server.
func NewServer() *Server {
	s := &Server{
		Logger: NewLogger(),
		server: &http.Server{},
		router: httprouter.New(),
	}

	s.router.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.Error(w, r, http.StatusNotFound, ErrorResponse{Msg: "Not Found", err: nil})
	})
	s.router.MethodNotAllowed = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.Error(w, r, http.StatusMethodNotAllowed, ErrorResponse{Msg: "Method Not Allowed", err: nil})
	})

	s.registerHealthCheckRoutes()
	s.registerMovieRoutes()

	return s
}

// Start starts a HTTP server.
func (s *Server) Start() error {
	s.server.Addr = s.Addr
	s.server.Handler = s.router
	s.server.IdleTimeout = time.Minute
	s.server.ReadTimeout = 10 * time.Second
	s.server.WriteTimeout = 30 * time.Second
	s.server.ErrorLog = slog.NewLogLogger(s.Logger.Handler(), slog.LevelError)

	s.Logger.Debug("starting HTTP server", "addr", s.Addr)

	err := s.server.ListenAndServe()

	s.Logger.Error("HTTP server startup", "error", err)

	return err
}

func NewLogger() *slog.Logger {
	opts := slog.HandlerOptions{AddSource: true}
	handler := slog.NewJSONHandler(os.Stderr, &opts)

	logger := slog.New(handler).With("module", "http")

	return logger
}
