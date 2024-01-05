// Package main implements greenlight application startup.
package main

import (
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/denpeshkov/greenlight/internal/http"
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

	if err := run(&cfg); err != nil {
		logger.Error("application startup", "error", err)
		os.Exit(1)
	}
}

func run(cfg *Config) (err error) {
	srv := http.NewServer()
	srv.Addr = cfg.http.addr

	db := postgres.NewDB(
		cfg.pgDB.dsn,
	)

	maxIdleTime, err := time.ParseDuration(cfg.pgDB.maxIdleTime)
	if err != nil {
		return fmt.Errorf("invalid db-max-idle-time: %w", err)
	}
	if err := db.Open(
		postgres.WithMaxOpenConns(cfg.pgDB.maxOpenConns),
		postgres.WithMaxIdleConns(cfg.pgDB.maxIdleConns),
		postgres.WithMaxIdleTime(maxIdleTime),
	); err != nil {
		return fmt.Errorf("opening database connection: %w", err)
	}
	defer func() {
		dbErr := db.Close()
		if dbErr != nil {
			err = fmt.Errorf("closing DB connection: %w", dbErr)
		}
		err = errors.Join(err, dbErr)
	}()

	srv.MovieService = postgres.NewMovieService(db)

	srvErr := srv.Start()
	return fmt.Errorf("starting HTTP server: %w", srvErr)
}

func newLogger() *slog.Logger {
	opts := slog.HandlerOptions{AddSource: true, Level: slog.LevelDebug}
	logger := slog.New(slog.NewJSONHandler(os.Stderr, &opts)).With("module", "app")

	return logger
}
