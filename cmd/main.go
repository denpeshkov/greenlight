package main

import (
	"flag"
	"log/slog"
	"os"

	"github.com/denpeshkov/greenlight/internal/http"
)

func main() {
	logger := newLogger()

	fs := flag.NewFlagSet("greenlight-service", flag.ExitOnError)
	addr := fs.String("addr", ":8080", "`address` to listen on")

	fs.Parse(os.Args[1:])

	srv := http.NewServer()
	srv.Addr = *addr

	err := srv.Start()

	logger.Error("error starting HTTP server", "error", err)
	os.Exit(1)
}

func newLogger() *slog.Logger {
	opts := slog.HandlerOptions{AddSource: true}
	logger := slog.New(slog.NewJSONHandler(os.Stderr, &opts)).With("module", "cmd")

	return logger
}
