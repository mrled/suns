package reattestbatch

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/mrled/suns/symval/internal/adapter/s3materializedview"
	"github.com/mrled/suns/symval/internal/logger"
	"github.com/mrled/suns/symval/internal/repository/dynamorepo"
	"github.com/mrled/suns/symval/internal/service/dnsclaims"
	"github.com/mrled/suns/symval/internal/usecase/reattest"
)

// Handler holds the dependencies for the reattestbatch Lambda handler
type Handler struct {
	log              *slog.Logger
	dynamoRepo       *dynamorepo.DynamoRepository
	s3View           *s3materializedview.S3MaterializedView
	dynamoTable      string
	s3BucketName     string
	s3DataKey        string
	gracePeriodHours int
}

// NewHandler creates a new reattestbatch handler with initialized dependencies
func NewHandler() (*Handler, error) {
	// Initialize logger with executable name for filtering
	log := logger.NewDefaultLogger()
	log = logger.WithExecutable(log, "reattestbatch")
	logger.SetDefault(log)

	// Get environment variables
	dynamoTable := os.Getenv("DYNAMODB_TABLE")
	if dynamoTable == "" {
		return nil, fmt.Errorf("DYNAMODB_TABLE environment variable is required")
	}
	log.Info("Using DynamoDB table", slog.String("table", dynamoTable))

	s3BucketName := os.Getenv("S3_BUCKET")
	if s3BucketName == "" {
		return nil, fmt.Errorf("S3_BUCKET environment variable is required")
	}
	log.Info("Using S3 bucket", slog.String("bucket", s3BucketName))

	// Use S3_DATA_KEY from environment or default to records/domains.json
	s3DataKey := os.Getenv("S3_DATA_KEY")
	if s3DataKey == "" {
		s3DataKey = "records/domains.json"
	}
	log.Info("Using S3 key", slog.String("key", s3DataKey))

	gracePeriodHours := 72

	return &Handler{
		log:              log,
		dynamoTable:      dynamoTable,
		s3BucketName:     s3BucketName,
		s3DataKey:        s3DataKey,
		gracePeriodHours: gracePeriodHours,
	}, nil
}

// Handle processes scheduled Lambda events for batch re-attestation
func (h *Handler) Handle(ctx context.Context, event map[string]interface{}) error {
	// Create a logger with Lambda context
	requestLogger := logger.WithLambda(h.log,
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
	h.dynamoRepo = dynamorepo.NewDynamoRepository(dynamoClient, h.dynamoTable)

	// Initialize S3 client and materialized view
	s3Client := s3.NewFromConfig(cfg)
	h.s3View = s3materializedview.New(s3Client, h.s3BucketName, h.s3DataKey)

	// Load current data from S3
	memRepo, err := h.s3View.Load(ctx)
	if err != nil {
		requestLogger.Error("Failed to load data from S3",
			slog.Bool("notify", true),
			slog.String("error", err.Error()))
		return fmt.Errorf("failed to load data from S3: %w", err)
	}

	// Create DNS service for attestation
	dnsService := dnsclaims.NewService()

	// Create reattest use case with DynamoDB support
	reattestUC := reattest.NewReattestUseCaseWithDynamo(dnsService, memRepo, h.dynamoRepo)
	reattestUC.SetGracePeriod(h.gracePeriodHours)

	// Perform re-attestation and update/delete as needed
	results, stats, err := reattestUC.ReattestAllAndUpdate(ctx)
	if err != nil {
		requestLogger.Error("Failed to re-attest and update groups",
			slog.Bool("notify", true),
			slog.String("error", err.Error()))
		return fmt.Errorf("failed to re-attest and update groups: %w", err)
	}

	// Log details for each result if needed
	for _, result := range results {
		groupLogger := requestLogger.With(
			slog.String("group_id", result.GroupID),
			slog.String("owner", result.Owner),
			slog.String("type", result.Type),
			slog.Int("record_count", len(result.Records)))

		if result.IsValid {
			groupLogger.Info("Group attestation succeeded")
		} else {
			// Check if group was within grace period or deleted
			var oldestValidation time.Time
			for _, record := range result.Records {
				if oldestValidation.IsZero() || record.ValidateTime.Before(oldestValidation) {
					oldestValidation = record.ValidateTime
				}
			}
			hoursSinceValidation := time.Since(oldestValidation).Hours()

			if hoursSinceValidation > float64(h.gracePeriodHours) {
				groupLogger.Warn("Group attestation failed, grace period exceeded (deleted)",
					slog.String("error", result.ErrorMessage),
					slog.Float64("hours_since_validation", hoursSinceValidation),
					slog.Int("grace_period_hours", h.gracePeriodHours))
			} else {
				groupLogger.Info("Group attestation failed, within grace period (skipped)",
					slog.String("error", result.ErrorMessage),
					slog.Float64("hours_since_validation", hoursSinceValidation),
					slog.Int("grace_period_hours", h.gracePeriodHours))
			}
		}
	}

	requestLogger.Info("Re-attestation completed",
		slog.Int("groups_processed", stats.GroupsProcessed),
		slog.Int("records_updated", stats.RecordsUpdated),
		slog.Int("records_deleted", stats.RecordsDeleted),
		slog.Int("records_skipped", stats.RecordsSkipped),
		slog.Int("errors", stats.Errors))

	return nil
}
