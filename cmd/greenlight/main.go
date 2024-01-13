// Package main implements greenlight application startup.
package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/denpeshkov/greenlight/internal/http"
	"github.com/denpeshkov/greenlight/internal/multierr"
	"github.com/denpeshkov/greenlight/internal/postgres"
	_ "github.com/lib/pq"
)

// Config represents an application configuration parameters.
type Config struct {
	// HTTP server
	http struct {
		addr         string
		idleTimeout  time.Duration
		readTimeout  time.Duration
		writeTimeout time.Duration
	}
	// PostgreSQL database
	pgDB struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  time.Duration
	}
}

func main() {
	logger := newLogger()

	cfg := Config{}
	err := cfg.parseFlags(os.Args[1:])
	if err != nil {
		logger.Error("flags parsing error: %w", err)
	}

	if err := run(&cfg, logger); err != nil {
		logger.Error("application error", "error", err)
		os.Exit(1)
	}
}

// run executes the program.
func run(cfg *Config, logger *slog.Logger) error {
	db := postgres.NewDB(cfg.pgDB.dsn)
	srv := http.NewServer(cfg.http.addr)

	// Application graceful shutdown
	shutdownErr := make(chan error)
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

		<-quit
		logger.Debug("shutting down HTTP server")
		shutdownErr <- srv.Close()

		logger.Debug("shutting down database")
		shutdownErr <- db.Close()
	}()

	// Setting up DB
	err := db.Open(
		postgres.WithMaxOpenConns(cfg.pgDB.maxOpenConns),
		postgres.WithMaxIdleConns(cfg.pgDB.maxIdleConns),
		postgres.WithMaxIdleTime(cfg.pgDB.maxIdleTime),
	)
	if err != nil {
		return fmt.Errorf("connecting to a database: %w", err)
	}
	logger.Debug("database connection established")

	// Setting up HTTP server
	srv.MovieService = postgres.NewMovieService(db)

	err = srv.Open(
		http.WithIdleTimeout(cfg.http.idleTimeout),
		http.WithReadTimeout(cfg.http.readTimeout),
		http.WithWriteTimeout(cfg.http.writeTimeout),
	)
	if err != nil {
		return fmt.Errorf("HTTP server serve: %w", err)
	}

	// Handle graceful shutdown
	srvErr := <-shutdownErr
	dbErr := <-shutdownErr

	logger.Debug("Errors from shutdown", "srvErr==nil", srvErr == nil, "dbErr==nil", dbErr == nil)
	return multierr.Join(srvErr, dbErr)
}

func (c *Config) parseFlags(args []string) error {
	fs := flag.NewFlagSet("greenlight", flag.ExitOnError)
	// HTTP
	fs.StringVar(&c.http.addr, "addr", ":8080", "address to listen on in format")
	fs.DurationVar(&c.http.idleTimeout, "http-idle-timeout", time.Minute, "HTTP server idle timeout")
	fs.DurationVar(&c.http.readTimeout, "http-read-timeout", 10*time.Second, "HTTP server read timeout")
	fs.DurationVar(&c.http.writeTimeout, "http-write-timeout", 30*time.Second, "HTTP server write timeout")

	//PostgreSQL
	fs.StringVar(&c.pgDB.dsn, "dsn", os.Getenv("POSTGRES_DSN"), "PostgreSQL DSN")
	fs.IntVar(&c.pgDB.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	fs.IntVar(&c.pgDB.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	fs.DurationVar(&c.pgDB.maxIdleTime, "db-max-idle-time", 15*time.Minute, "PostgreSQL max connection idle time")

	return fs.Parse(args)
}

func newLogger() *slog.Logger {
	opts := slog.HandlerOptions{Level: slog.LevelDebug}
	logger := slog.New(slog.NewJSONHandler(os.Stderr, &opts)).With("module", "app")

	return logger
}
