// Package main implements greenlight application startup.
package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/denpeshkov/greenlight/internal/http"
	"github.com/denpeshkov/greenlight/internal/multierr"
	"github.com/denpeshkov/greenlight/internal/postgres"
	_ "github.com/lib/pq"
)

// Config represents an application configuration parameters.
type Config struct {
	http struct {
		addr string
	}
	// PostgreSQL
	pgDB struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  string
	}
}

func (c *Config) parseFlags(args []string) {
	fs := flag.NewFlagSet("greenlight", flag.ExitOnError)
	// HTTP
	fs.StringVar(&c.http.addr, "addr", ":8080", "address to listen on in format")

	//PostgreSQL
	fs.StringVar(&c.pgDB.dsn, "dsn", os.Getenv("POSTGRES_DSN"), "PostgreSQL DSN")
	fs.IntVar(&c.pgDB.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	fs.IntVar(&c.pgDB.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	fs.StringVar(&c.pgDB.maxIdleTime, "db-max-idle-time", "15m", "PostgreSQL max connection idle time")

	fs.Parse(args)
}

func main() {
	logger := newLogger()

	cfg := Config{}
	cfg.parseFlags(os.Args[1:])

	if err := run(&cfg, logger); err != nil {
		logger.Error("application startup", "error", err)
		os.Exit(1)
	}
}

func run(cfg *Config, logger *slog.Logger) (err error) {
	srv := http.NewServer()
	srv.Addr = cfg.http.addr

	maxIdleTime, err := time.ParseDuration(cfg.pgDB.maxIdleTime)
	if err != nil {
		return fmt.Errorf("invalid db-max-idle-time: %w", err)
	}
	db, err := postgres.Open(
		cfg.pgDB.dsn,
		postgres.WithMaxOpenConns(cfg.pgDB.maxOpenConns),
		postgres.WithMaxIdleConns(cfg.pgDB.maxIdleConns),
		postgres.WithMaxIdleTime(maxIdleTime),
	)
	if err != nil {
		return fmt.Errorf("connecting to a database: %w", err)
	}
	defer func() {
		if dbErr := db.Close(); dbErr != nil {
			err = multierr.Join(fmt.Errorf("closing a database connection: %w", dbErr), err)
		}
	}()

	logger.Debug("database connection established")

	srv.MovieService = postgres.NewMovieService(db)

	srvErr := srv.Open()
	if srvErr != nil {
		return fmt.Errorf("starting an HTTP server: %w", srvErr)
	}
	return nil
}

func newLogger() *slog.Logger {
	opts := slog.HandlerOptions{Level: slog.LevelDebug}
	logger := slog.New(slog.NewJSONHandler(os.Stderr, &opts)).With("module", "app")

	return logger
}
