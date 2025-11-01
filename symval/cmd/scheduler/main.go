package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/mrled/suns/symval/internal/adapter/s3materializedview"
	"github.com/mrled/suns/symval/internal/logger"
	"github.com/mrled/suns/symval/internal/model"
	"github.com/mrled/suns/symval/internal/repository/dynamorepo"
	"github.com/mrled/suns/symval/internal/service/dnsclaims"
	"github.com/mrled/suns/symval/internal/usecase/reattest"
)

var (
	log              *slog.Logger
	dynamoRepo       *dynamorepo.DynamoRepository
	s3View           *s3materializedview.S3MaterializedView
	dynamoTable      string
	s3BucketName     string
	s3DataKey        string
	gracePeriodHours = 72
)

func init() {
	// Initialize logger with executable name for filtering
	log = logger.NewDefaultLogger()
	log = logger.WithExecutable(log, "scheduler")
	logger.SetDefault(log)

	// Get environment variables
	dynamoTable = os.Getenv("DYNAMODB_TABLE")
	if dynamoTable == "" {
		log.Error("DYNAMODB_TABLE environment variable is required")
		os.Exit(1)
	}
	log.Info("Using DynamoDB table", slog.String("table", dynamoTable))

	s3BucketName = os.Getenv("S3_BUCKET")
	if s3BucketName == "" {
		log.Error("S3_BUCKET environment variable is required")
		os.Exit(1)
	}
	log.Info("Using S3 bucket", slog.String("bucket", s3BucketName))

	// Use S3_DATA_KEY from environment or default to records/domains.json
	s3DataKey = os.Getenv("S3_DATA_KEY")
	if s3DataKey == "" {
		s3DataKey = "records/domains.json"
	}
	log.Info("Using S3 key", slog.String("key", s3DataKey))
}

func handler(ctx context.Context, event map[string]interface{}) error {
	// Create a logger with Lambda context
	requestLogger := logger.WithLambda(log,
		os.Getenv("AWS_LAMBDA_FUNCTION_NAME"),
		os.Getenv("AWS_LAMBDA_FUNCTION_VERSION"),
		"") // No request ID for scheduled events

	requestLogger.Info("Scheduled Lambda triggered", slog.Any("event", event))

	// Initialize AWS clients
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		requestLogger.Error("Failed to load AWS config", slog.String("error", err.Error()))
		return fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Initialize DynamoDB client and repository
	dynamoClient := dynamodb.NewFromConfig(cfg)
	dynamoRepo = dynamorepo.NewDynamoRepository(dynamoClient, dynamoTable)

	// Initialize S3 client and materialized view
	s3Client := s3.NewFromConfig(cfg)
	s3View = s3materializedview.New(s3Client, s3BucketName, s3DataKey)

	// Load current data from S3
	memRepo, err := s3View.Load(ctx)
	if err != nil {
		requestLogger.Error("Failed to load data from S3",
			slog.Bool("notify", true),
			slog.String("error", err.Error()))
		return fmt.Errorf("failed to load data from S3: %w", err)
	}

	// Get all records from the materialized view
	allRecords, err := memRepo.List(ctx)
	if err != nil {
		requestLogger.Error("Failed to list records from memory repo",
			slog.Bool("notify", true),
			slog.String("error", err.Error()))
		return fmt.Errorf("failed to list records: %w", err)
	}

	// Group records by GroupID for validation
	groupedRecords := model.GroupByGroupID(allRecords)

	requestLogger.Info("Processing groups for re-attestation",
		slog.Int("total_groups", len(groupedRecords)),
		slog.Int("total_records", len(allRecords)))

	// Create DNS service for attestation
	dnsService := dnsclaims.NewService()

	// Create reattest use case
	reattestUC := reattest.NewReattestUseCase(dnsService, memRepo)

	// Perform re-attestation to check validity
	results, err := reattestUC.ReattestAll(ctx)
	if err != nil {
		requestLogger.Error("Failed to re-attest groups",
			slog.Bool("notify", true),
			slog.String("error", err.Error()))
		return fmt.Errorf("failed to re-attest groups: %w", err)
	}

	successCount := 0
	failureCount := 0
	deleteCount := 0
	skipCount := 0

	// Process each attestation result
	for _, result := range results {
		groupRecords := groupedRecords[result.GroupID]
		groupLogger := requestLogger.With(
			slog.String("group_id", result.GroupID),
			slog.String("owner", result.Owner),
			slog.String("type", result.Type),
			slog.Int("record_count", len(groupRecords)))

		if result.IsValid {
			// Attestation succeeded - update all records in the group with current timestamp
			groupLogger.Info("Group attestation succeeded, updating records")

			for _, record := range groupRecords {
				// Keep the snapshot revision for conditional update
				snapshotRev := record.Rev
				record.ValidateTime = time.Now()
				if _, err := dynamoRepo.SetValidationIfUnchanged(ctx, record, snapshotRev); err != nil {
					if err == model.ErrRevConflict {
						groupLogger.Warn("Record changed during validation, skipping",
							slog.String("hostname", record.Hostname))
						skipCount++
					} else {
						groupLogger.Error("Failed to update record in DynamoDB",
							slog.Bool("notify", true),
							slog.String("hostname", record.Hostname),
							slog.String("error", err.Error()))
						failureCount++
					}
				} else {
					successCount++
				}
			}
		} else {
			// Attestation failed - check grace period
			groupLogger.Warn("Group attestation failed",
				slog.String("error", result.ErrorMessage))

			// Get the oldest validation time from the group
			var oldestValidation time.Time
			for _, record := range groupRecords {
				if oldestValidation.IsZero() || record.ValidateTime.Before(oldestValidation) {
					oldestValidation = record.ValidateTime
				}
			}

			hoursSinceValidation := time.Since(oldestValidation).Hours()

			if hoursSinceValidation > float64(gracePeriodHours) {
				// Grace period exceeded - delete all records in the group
				groupLogger.Warn("Grace period exceeded, deleting group",
					slog.Float64("hours_since_validation", hoursSinceValidation),
					slog.Int("grace_period_hours", gracePeriodHours))

				for _, record := range groupRecords {
					if err := dynamoRepo.DeleteIfUnchanged(ctx, result.GroupID, record.Hostname, record.Rev); err != nil {
						if err == model.ErrRevConflict {
							groupLogger.Warn("Record changed during deletion, skipping",
								slog.String("hostname", record.Hostname))
							skipCount++
						} else {
							groupLogger.Error("Failed to delete record from DynamoDB",
								slog.Bool("notify", true),
								slog.String("hostname", record.Hostname),
								slog.String("error", err.Error()))
							failureCount++
						}
					} else {
						deleteCount++
					}
				}
			} else {
				// Within grace period - do nothing
				groupLogger.Info("Within grace period, skipping",
					slog.Float64("hours_since_validation", hoursSinceValidation),
					slog.Int("grace_period_hours", gracePeriodHours))
				skipCount += len(groupRecords)
			}
		}
	}

	requestLogger.Info("Re-attestation completed",
		slog.Int("groups_processed", len(results)),
		slog.Int("records_updated", successCount),
		slog.Int("records_deleted", deleteCount),
		slog.Int("records_skipped", skipCount),
		slog.Int("errors", failureCount))

	return nil
}

func main() {
	log.Info("Starting scheduled Lambda handler")
	lambda.Start(handler)
}
