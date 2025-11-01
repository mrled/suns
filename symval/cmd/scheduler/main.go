package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/mrled/suns/symval/internal/logger"
)

var log *slog.Logger

func init() {
	// Initialize logger with executable name for filtering
	log = logger.NewDefaultLogger()
	log = logger.WithExecutable(log, "scheduler")
	logger.SetDefault(log)
}

func handler(ctx context.Context, event map[string]interface{}) error {
	// Create a logger with Lambda context
	requestLogger := logger.WithLambda(log,
		os.Getenv("AWS_LAMBDA_FUNCTION_NAME"),
		os.Getenv("AWS_LAMBDA_FUNCTION_VERSION"),
		"") // No request ID for scheduled events

	requestLogger.Info("Scheduled Lambda triggered", slog.Any("event", event))
	return nil
}

func main() {
	log.Info("Starting scheduled Lambda handler")
	lambda.Start(handler)
}
