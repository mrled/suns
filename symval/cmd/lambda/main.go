package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/mrled/suns/symval/internal/lambdahandlers/httpapi"
	"github.com/mrled/suns/symval/internal/lambdahandlers/reattestbatch"
	"github.com/mrled/suns/symval/internal/lambdahandlers/streamer"
	"github.com/mrled/suns/symval/internal/logger"
)

func main() {
	// Initialize logger
	log := logger.NewDefaultLogger()
	log = logger.WithExecutable(log, "lambda")
	logger.SetDefault(log)

	// Get the LAMBDA_HANDLER environment variable
	handlerType := os.Getenv("LAMBDA_HANDLER")
	if handlerType == "" {
		log.Error("LAMBDA_HANDLER environment variable is required")
		fmt.Fprintln(os.Stderr, "Error: LAMBDA_HANDLER environment variable is required")
		fmt.Fprintln(os.Stderr, "Valid values: httpapi, reattestbatch, streamer")
		os.Exit(1)
	}

	log.Info("Starting Lambda handler", slog.String("handler", handlerType))

	// Route to the appropriate handler based on LAMBDA_HANDLER
	switch handlerType {
	case "httpapi":
		handler, err := httpapi.NewHandler()
		if err != nil {
			log.Error("Failed to initialize httpapi handler", slog.String("error", err.Error()))
			os.Exit(1)
		}
		lambda.Start(handler.Handle)

	case "reattestbatch":
		handler, err := reattestbatch.NewHandler()
		if err != nil {
			log.Error("Failed to initialize reattestbatch handler", slog.String("error", err.Error()))
			os.Exit(1)
		}
		lambda.Start(handler.Handle)

	case "streamer":
		handler, err := streamer.NewHandler()
		if err != nil {
			log.Error("Failed to initialize streamer handler", slog.String("error", err.Error()))
			os.Exit(1)
		}
		lambda.Start(handler.Handle)

	default:
		log.Error("Invalid LAMBDA_HANDLER value", slog.String("handler", handlerType))
		fmt.Fprintf(os.Stderr, "Error: Invalid LAMBDA_HANDLER value: %s\n", handlerType)
		fmt.Fprintln(os.Stderr, "Valid values: httpapi, reattestbatch, streamer")
		os.Exit(1)
	}
}
