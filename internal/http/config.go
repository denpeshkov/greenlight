package http

import "time"

// Option represents a configuration option for an HTTP.
type Option func(o *options)

// options represents all server options.
type options struct {
	IdleTimeout     time.Duration
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
	MaxRequestBody  int64
}

// WithIdleTimeout sets the idle timeout.
func WithIdleTimeout(t time.Duration) Option {
	return func(o *options) {
		o.IdleTimeout = t
	}
}

// WithReadTimeout sets the read timeout.
func WithReadTimeout(t time.Duration) Option {
	return func(o *options) {
		o.ReadTimeout = t
	}
}

// WithWriteTimeout sets the write timeout.
func WithWriteTimeout(t time.Duration) Option {
	return func(o *options) {
		o.WriteTimeout = t
	}
}

// WithShutdownTimeout sets the shutdown timeout.
func WithShutdownTimeout(t time.Duration) Option {
	return func(o *options) {
		o.ShutdownTimeout = t
	}
}

// WithMaxRequestBody sets the maximum size of the request body in bytes.
func WithMaxRequestBody(sz int64) Option {
	return func(o *options) {
		o.MaxRequestBody = sz
	}
}
