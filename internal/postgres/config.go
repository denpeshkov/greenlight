package postgres

import "time"

type options struct {
	maxOpenConns    int
	maxIdleConns    int
	connMaxIdleTime time.Duration
	connTimeout     time.Duration
	queryTimeout    time.Duration
}

// Option represents a configuration option for PostgreSQL*options.
type Option func(db *options)

// WithMaxOpenConns sets the maximum number of open connections to the database.
func WithMaxOpenConns(maxOpenConns int) Option {
	return func(opts *options) {
		opts.maxOpenConns = maxOpenConns
	}
}

// WithMaxIdleConns sets the maximum number of connections in the idle connection pool.
func WithMaxIdleConns(maxIdleConns int) Option {
	return func(opts *options) {
		opts.maxIdleConns = maxIdleConns
	}
}

// WithConnMaxIdleTime sets the maximum amount of time a connection may be idle.
func WithMaxIdleTime(maxIdleTime time.Duration) Option {
	return func(opts *options) {
		opts.connMaxIdleTime = maxIdleTime
	}
}

// WithConnectionTimeout sets the timeout for establishing the connection.
func WithConnectionTimeout(connTimeout time.Duration) Option {
	return func(opts *options) {
		opts.connTimeout = connTimeout
	}
}

// WithQueryTimeout sets the query execution timeout.
func WithQueryTimeout(queryTimeout time.Duration) Option {
	return func(opts *options) {
		opts.queryTimeout = queryTimeout
	}
}
