package applystream

import (
	"context"
	"fmt"
	"log"

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
	log.Printf("Processing batch of %d records from DynamoDB stream", len(records))

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
			log.Printf("Error processing record %s: %v", record.EventID, err)
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
	log.Printf("Successfully processed %d/%d stream records, S3 file now contains %d records",
		processedCount, len(records), len(allRecords))

	return nil
}

// loadRepository loads the current repository state from S3
func (s *Service) loadRepository(ctx context.Context) (*memrepo.MemoryRepository, error) {
	memRepo, err := s.s3View.Load(ctx)
	if err != nil {
		log.Printf("Error loading repository from S3: %v", err)
		// If file doesn't exist or error occurs, start with empty repository
		log.Printf("Starting with empty repository")
		return memrepo.NewMemoryRepository(), nil
	}
	return memRepo, nil
}

// processRecord processes a single DynamoDB stream record
func (s *Service) processRecord(ctx context.Context, repo model.DomainRepository, record events.DynamoDBEventRecord) error {
	log.Printf("Processing record: EventID=%s, EventName=%s", record.EventID, record.EventName)

	switch record.EventName {
	case "INSERT", "MODIFY":
		return s.handleInsertOrModify(ctx, repo, record)
	case "REMOVE":
		return s.handleRemove(ctx, repo, record)
	default:
		log.Printf("Unknown event type: %s", record.EventName)
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

	log.Printf("Stored/Updated record: GroupID=%s, Hostname=%s", domainRecord.GroupID, domainRecord.Hostname)
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
		log.Printf("Record not found for deletion: GroupID=%s, Hostname=%s", pk, sk)
	} else {
		log.Printf("Removed record: GroupID=%s, Hostname=%s", pk, sk)
	}

	return nil
}