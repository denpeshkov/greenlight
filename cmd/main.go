// Package main implements greenlight application startup.
package main

import (
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/denpeshkov/greenlight/internal/http"
	"github.com/denpeshkov/greenlight/internal/postgres"
	_ "github.com/lib/pq"
)

func main() {
	logger := newLogger()

	if err := run(); err != nil {
		logger.Error("application startup", "error", err)
		os.Exit(1)
	}
}

func run() (err error) {
	fs := flag.NewFlagSet("greenlight-service", flag.ExitOnError)
	addr := fs.String("addr", ":8080", "`address` to listen on")
	dsn := fs.String("dsn", os.Getenv("POSTGRES_DSN"), "PostgreSQL `DSN`")

	fs.Parse(os.Args[1:])

	srv := http.NewServer()
	srv.Addr = *addr

	db := postgres.NewDB(*dsn)
	if err := db.Open(); err != nil {
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
