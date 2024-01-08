package http

import "time"

// Option represents a configuration option for an HTTP server.
type Option func(s *Server)

// WithIdleTimeout sets the idle timeout.
func WithIdleTimeout(idleTimeout time.Duration) Option {
	return func(s *Server) {
		s.server.IdleTimeout = idleTimeout
	}
}

// WithReadTimeout sets the read timeout.
func WithReadTimeout(readTimeout time.Duration) Option {
	return func(s *Server) {
		s.server.ReadTimeout = readTimeout
	}
}

// WithWriteTimeout sets the write timeout.
func WithWriteTimeout(writeTimeout time.Duration) Option {
	return func(s *Server) {
		s.server.WriteTimeout = writeTimeout
	}
}
