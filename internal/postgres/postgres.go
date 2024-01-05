package postgres

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"os"
	"time"
)

// DB represents the database connection.
type DB struct {
	// Data source name.
	DSN string

	db     *sql.DB
	ctx    context.Context // context
	cancel func()          // context cancel func
	logger *slog.Logger
}

// NewDB returns a new instance of [DB].
func NewDB(dsn string) *DB {
	db := &DB{
		DSN:    dsn,
		logger: newLogger(),
	}
	db.ctx, db.cancel = context.WithTimeout(context.Background(), 10*time.Second)

	return db
}

// Open opens the database connection.
func (db *DB) Open(opts ...Option) (err error) {
	if db.DSN == "" {
		return errors.New("data source name (DSN) required")
	}

	if db.db, err = sql.Open("postgres", db.DSN); err != nil {
		return err
	}

	var options options
	for _, o := range opts {
		o(&options)
	}
	db.db.SetMaxOpenConns(options.maxOpenConns)
	db.db.SetMaxIdleConns(options.maxIdleConns)
	db.db.SetConnMaxIdleTime(options.maxIdleTime)

	if err = db.db.PingContext(db.ctx); err != nil {
		return err
	}

	db.logger.Debug("database connection established")

	return nil
}

func (db *DB) Close() error {
	db.cancel()
	return db.db.Close()
}

func newLogger() *slog.Logger {
	opts := slog.HandlerOptions{AddSource: true}
	handler := slog.NewJSONHandler(os.Stderr, &opts)

	logger := slog.New(handler).With("module", "postgres")

	return logger
}
