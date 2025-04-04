package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sqs"

	"asyncapi/apiserver"
	"asyncapi/config"
	"asyncapi/store"
)

// main runs the API server.
//
// It loads the configuration, sets up the logger, sets up the store, sets up
// the AWS clients, and starts the server.
//
// If any of these steps fail, it panics with the error.
func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

// run sets up and starts the API server. It loads the configuration, initializes
// the logger, sets up database connections, creates AWS S3 and SQS clients, and
// initializes the API server with these components. It handles system interrupts
// to gracefully shut down the server. If any step fails, it returns an error.

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
	// Set Context to signal Notify Context

	// Set Context to signal Notify Context
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	//setup sqs client code for testing
	sdkConfig, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		return fmt.Errorf("Couldn't load default configuration. Have you set up your AWS account?:%w", err)
	}
	sqsClient := sqs.NewFromConfig(sdkConfig, func(options *sqs.Options) {
		options.BaseEndpoint = aws.String(cfg.ReportsSQSEndpoint)
	})

	s3Client := s3.NewFromConfig(sdkConfig, func(options *s3.Options) {
		options.BaseEndpoint = aws.String(cfg.S3LocalstackEndpoint)
		options.UsePathStyle = true
	})
	// Create a presign client from the s3 client
	s3PresignClient := s3.NewPresignClient(s3Client)
	// Create a new API server instance
	apiServer := apiserver.New(cfg, logger, dataStore, jwtManager, sqsClient, s3PresignClient)
	// Start the API server
	if err := apiServer.Start(ctx); err != nil {
		return err
	}

	return nil
}
