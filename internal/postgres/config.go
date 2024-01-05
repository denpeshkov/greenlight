package postgres

import "time"

// Options represents a configuration option for PostgreSQL DB.
type Option func(opts *options) error

type options struct {
	maxOpenConns int
	maxIdleConns int
	maxIdleTime  time.Duration
}

func WithMaxOpenConns(maxOpenConns int) Option {
	return func(opts *options) error {
		opts.maxOpenConns = maxOpenConns
		return nil
	}
}

func WithMaxIdleConns(maxIdleConns int) Option {
	return func(opts *options) error {
		opts.maxIdleConns = maxIdleConns
		return nil
	}
}

func WithMaxIdleTime(maxIdleTime time.Duration) Option {
	return func(opts *options) error {
		opts.maxIdleTime = maxIdleTime
		return nil
	}
}
