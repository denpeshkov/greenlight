package postgres

import "time"

// Option represents a configuration option for PostgreSQL DB.
type Option func(db *DB)

// WithMaxOpenConns sets the maximum number of open connections to the database.
func WithMaxOpenConns(maxOpenConns int) Option {
	return func(db *DB) {
		db.db.SetMaxOpenConns(maxOpenConns)
	}
}

// WithMaxIdleConns sets the maximum number of connections in the idle connection pool.
func WithMaxIdleConns(maxIdleConns int) Option {
	return func(db *DB) {
		db.db.SetMaxIdleConns(maxIdleConns)
	}
}

// WithConnMaxIdleTime sets the maximum amount of time a connection may be idle.
func WithMaxIdleTime(maxIdleTime time.Duration) Option {
	return func(db *DB) {
		db.db.SetConnMaxIdleTime(maxIdleTime)
	}
}
