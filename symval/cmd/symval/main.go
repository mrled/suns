package main

import (
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/mrled/suns/symval/cmd/symval/commands"
	"github.com/mrled/suns/symval/internal/logger"
)

var log *slog.Logger

func init() {
	// Initialize the logger with default configuration from environment variables
	log = logger.NewDefaultLogger()
	// Add the executable name for filtering in log aggregation
	log = logger.WithExecutable(log, "symval")
	// Set as the default logger for any packages that use slog directly
	logger.SetDefault(log)
}

// Main entry point
//
// Expectations of subcommands' RunE functions:
// - Return *commands.UsageError for usage-related errors (exit code 2)
// - Return *commands.ExitError for custom exit codes
// - Return other errors for general failures (exit code 1)
// - Set cmd.SilenceErrors, because we handle error printing here
// - Set cmd.SilenceUsage as appropriate before returning errors (showing only for invalid command line argument errors)
func main() {
	if err := commands.Execute(); err != nil {
		code := 1

		// If it's a usage error, set exit code to 2
		var usageErr *commands.UsageError
		if errors.As(err, &usageErr) {
			code = 2
		}

		// If it's an ExitError, use the specified exit code
		var exitErr *commands.ExitError
		if errors.As(err, &exitErr) {
			code = exitErr.Code
		}

		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(code)
	}
}
