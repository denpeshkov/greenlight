package postgres

import (
	"context"
	"database/sql"
	"errors"
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
	ctx    context.Context // context
	cancel func()          // context cancel func
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
	if db.DSN == "" {
		return errors.New("data source name (DSN) required")
	}

	if db.db, err = sql.Open("postgres", db.DSN); err != nil {
		return err
	}
	db.ctx, db.cancel = context.WithTimeout(context.Background(), db.opts.ctxTimeout)

	if err = db.db.PingContext(db.ctx); err != nil {
		return err
	}
	return db.Migrate(UP)
}

// Close gracefully shuts down the database.
func (db *DB) Close() error {
	db.cancel()
	err := db.Migrate(DOWN)
	err2 := db.db.Close()
	return multierr.Join(err2, err)
}

// newLogger returns a database logger.
func newLogger() *slog.Logger {
	opts := slog.HandlerOptions{Level: slog.LevelDebug}
	handler := slog.NewJSONHandler(os.Stderr, &opts)

	logger := slog.New(handler).With("module", "postgres")

	return logger
}
