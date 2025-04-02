package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/RaikaSurendra/go_asyncapi/apiserver"
	"github.com/RaikaSurendra/go_asyncapi/config"
	"github.com/RaikaSurendra/go_asyncapi/store"
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	// Load the configuration
	cfg, err := config.New()
	if err != nil {
		return err
	}
	// Create a new logger
	jsonHandler := slog.NewJSONHandler(os.Stdout, nil)
	logger := slog.New(jsonHandler)

	db, err := store.NewPostgresDb(cfg)
	if err != nil {
		return nil
	}
	dataStore := store.New(db)
	jwtManager := apiserver.NewJwtManager(cfg)
	// Create a new API server instance
	apiServer := apiserver.New(cfg, logger, dataStore, jwtManager)
	// Set Context to signal Notify Context
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	// Start the API server
	if err := apiServer.Start(ctx); err != nil {
		return err
	}

	return nil
}
