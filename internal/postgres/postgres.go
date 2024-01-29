package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/denpeshkov/greenlight/internal/multierr"
)

// DB represents the database connection.
type DB struct {
	// Data source name.
	DSN string

	opts options

	db     *sql.DB
	logger *slog.Logger
}

func NewDB(dsn string, opts ...Option) *DB {
	db := &DB{
		DSN:    dsn,
		logger: newLogger(),
	}

	// Apply options
	for _, opt := range opts {
		opt(&db.opts)
	}

	return db
}

// Open returns a new instance of an established database connection.
func (db *DB) Open() (err error) {
	op := "postgres.DB.Open"

	if db.DSN == "" {
		return errors.New("data source name (DSN) required")
	}

	if db.db, err = sql.Open("postgres", db.DSN); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	db.db.SetMaxOpenConns(db.opts.maxOpenConns)
	db.db.SetMaxIdleConns(db.opts.maxIdleConns)
	db.db.SetConnMaxIdleTime(db.opts.connMaxIdleTime)

	ctx, cancel := context.WithTimeout(context.Background(), db.opts.connTimeout)
	defer cancel()

	if err = db.db.PingContext(ctx); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if err = db.Migrate(UP); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

// Close gracefully shuts down the database.
func (db *DB) Close() error {
	op := "postgres.DB.Close"

	err1 := db.Migrate(DOWN)
	err2 := db.db.Close()
	if err := multierr.Join(err2, err1); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

// newLogger returns a database logger.
func newLogger() *slog.Logger {
	opts := slog.HandlerOptions{Level: slog.LevelDebug}
	handler := slog.NewJSONHandler(os.Stderr, &opts)

	logger := slog.New(handler).With("module", "postgres")

	return logger
}
