package applystream

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/aws/aws-lambda-go/events"
	"github.com/mrled/suns/symval/internal/adapter/dynamostream"
	"github.com/mrled/suns/symval/internal/adapter/s3materializedview"
	"github.com/mrled/suns/symval/internal/model"
	"github.com/mrled/suns/symval/internal/repository/memrepo"
)

// Service handles DynamoDB stream processing and S3 materialized view updates
type Service struct {
	s3View *s3materializedview.S3MaterializedView
}

// New creates a new applystream service
func New(s3View *s3materializedview.S3MaterializedView) *Service {
	return &Service{
		s3View: s3View,
	}
}

// ProcessStreamBatch processes a batch of DynamoDB stream records
// This method ensures thread-safe processing by:
// 1. Loading the current state from S3
// 2. Applying all stream changes to an in-memory repository
// 3. Saving the updated state back to S3
//
// IMPORTANT: This assumes reservedConcurrentExecutions=1 in Lambda configuration
// to ensure only one instance runs at a time, making read-modify-write safe
func (s *Service) ProcessStreamBatch(ctx context.Context, records []events.DynamoDBEventRecord) error {
	slog.Info("Processing batch from DynamoDB stream", slog.Int("record_count", len(records)))

	// Load current data from S3 into memory repository
	memRepo, err := s.loadRepository(ctx)
	if err != nil {
		return fmt.Errorf("failed to load repository: %w", err)
	}

	// Process each stream record
	processedCount := 0
	for _, record := range records {
		if err := s.processRecord(ctx, memRepo, record); err != nil {
			// Log error but continue processing other records
			slog.Error("Error processing record",
				slog.String("event_id", record.EventID),
				slog.String("error", err.Error()))
			continue
		}
		processedCount++
	}

	// Save updated repository back to S3
	if err := s.s3View.Save(ctx, memRepo); err != nil {
		return fmt.Errorf("failed to save repository to S3: %w", err)
	}

	// Log final statistics
	allRecords, _ := memRepo.List(ctx)
	slog.Info("Successfully processed stream batch",
		slog.Int("processed", processedCount),
		slog.Int("total", len(records)),
		slog.Int("s3_record_count", len(allRecords)))

	return nil
}

// loadRepository loads the current repository state from S3
func (s *Service) loadRepository(ctx context.Context) (*memrepo.MemoryRepository, error) {
	memRepo, err := s.s3View.Load(ctx)
	if err != nil {
		slog.Warn("Error loading repository from S3", slog.String("error", err.Error()))
		// If file doesn't exist or error occurs, start with empty repository
		slog.Info("Starting with empty repository")
		return memrepo.NewMemoryRepository(), nil
	}
	return memRepo, nil
}

// processRecord processes a single DynamoDB stream record
func (s *Service) processRecord(ctx context.Context, repo model.DomainRepository, record events.DynamoDBEventRecord) error {
	slog.Debug("Processing record",
		slog.String("event_id", record.EventID),
		slog.String("event_name", record.EventName))

	switch record.EventName {
	case "INSERT", "MODIFY":
		return s.handleInsertOrModify(ctx, repo, record)
	case "REMOVE":
		return s.handleRemove(ctx, repo, record)
	default:
		slog.Warn("Unknown event type", slog.String("event_name", record.EventName))
		return fmt.Errorf("unknown event type: %s", record.EventName)
	}
}

// handleInsertOrModify handles INSERT and MODIFY stream events
func (s *Service) handleInsertOrModify(ctx context.Context, repo model.DomainRepository, record events.DynamoDBEventRecord) error {
	// Convert the DynamoDB stream record to domain model
	domainRecord, err := dynamostream.ConvertToDomainRecord(record.Change.NewImage)
	if err != nil {
		return fmt.Errorf("failed to convert stream record: %w", err)
	}

	// Store the record in the repository
	if err := repo.Store(ctx, domainRecord); err != nil {
		return fmt.Errorf("failed to store record: %w", err)
	}

	slog.Debug("Stored/Updated record",
		slog.String("group_id", domainRecord.GroupID),
		slog.String("hostname", domainRecord.Hostname))
	return nil
}

// handleRemove handles REMOVE stream events
func (s *Service) handleRemove(ctx context.Context, repo model.DomainRepository, record events.DynamoDBEventRecord) error {
	// Extract the keys from the stream record
	pk := dynamostream.ExtractStringAttribute(record.Change.Keys, "pk")
	sk := dynamostream.ExtractStringAttribute(record.Change.Keys, "sk")

	if pk == "" || sk == "" {
		return fmt.Errorf("missing required keys: pk=%s, sk=%s", pk, sk)
	}

	// Delete the record from the repository
	if err := repo.Delete(ctx, pk, sk); err != nil {
		if err != model.ErrNotFound {
			return fmt.Errorf("failed to delete record: %w", err)
		}
		// Record not found is not an error for delete operations
		slog.Debug("Record not found for deletion",
			slog.String("group_id", pk),
			slog.String("hostname", sk))
	} else {
		slog.Debug("Removed record",
			slog.String("group_id", pk),
			slog.String("hostname", sk))
	}

	return nil
}