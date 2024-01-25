package http

import "time"

// Option represents a configuration option for an HTTP.
type Option func(o *options)

// options represents all server options.
type options struct {
	idleTimeout     time.Duration
	readTimeout     time.Duration
	writeTimeout    time.Duration
	shutdownTimeout time.Duration
	maxRequestBody  int64
}

// WithIdleTimeout sets the idle timeout.
func WithIdleTimeout(t time.Duration) Option {
	return func(o *options) {
		o.idleTimeout = t
	}
}

// WithReadTimeout sets the read timeout.
func WithReadTimeout(t time.Duration) Option {
	return func(o *options) {
		o.readTimeout = t
	}
}

// WithWriteTimeout sets the write timeout.
func WithWriteTimeout(t time.Duration) Option {
	return func(o *options) {
		o.writeTimeout = t
	}
}

// WithShutdownTimeout sets the shutdown timeout.
func WithShutdownTimeout(t time.Duration) Option {
	return func(o *options) {
		o.shutdownTimeout = t
	}
}

// WithMaxRequestBody sets the maximum size of the request body in bytes.
func WithMaxRequestBody(sz int64) Option {
	return func(o *options) {
		o.maxRequestBody = sz
	}
}
