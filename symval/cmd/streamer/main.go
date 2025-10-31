package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/mrled/suns/symval/internal/model"
	"github.com/mrled/suns/symval/internal/repository/dynamorepo"
	"github.com/mrled/suns/symval/internal/repository/memrepo"
	"github.com/mrled/suns/symval/internal/symgroup"
)

var (
	dynamoEndpoint string
	dynamoTable    string
	s3BucketName   string
	s3DataKey      string
	repo           model.DomainRepository
	s3Client       *s3.Client
)

func init() {
	// Optional endpoint override for local development or testing
	dynamoEndpoint = os.Getenv("DYNAMODB_ENDPOINT")
	if dynamoEndpoint != "" {
		log.Printf("Using custom DynamoDB endpoint: %s", dynamoEndpoint)
	} else {
		// When not using a custom endpoint, AWS_REGION is required
		awsRegion := os.Getenv("AWS_REGION")
		if awsRegion == "" {
			log.Fatal("AWS_REGION environment variable is required when DYNAMODB_ENDPOINT is not set")
		}
		log.Printf("Using AWS region: %s", awsRegion)
		log.Printf("Using default DynamoDB endpoint discovery")
	}

	dynamoTable = os.Getenv("DYNAMODB_TABLE")
	if dynamoTable == "" {
		log.Fatal("DYNAMODB_TABLE environment variable is required")
	}
	log.Printf("Using DynamoDB table: %s", dynamoTable)

	s3BucketName = os.Getenv("S3_BUCKET")
	if s3BucketName == "" {
		log.Fatal("S3_BUCKET environment variable is required")
	}
	log.Printf("Using S3 bucket: %s", s3BucketName)

	// Use S3_DATA_KEY from environment or default to records/domains.json
	s3DataKey = os.Getenv("S3_DATA_KEY")
	if s3DataKey == "" {
		s3DataKey = "records/domains.json"
	}
	log.Printf("Using S3 key: %s", s3DataKey)
}


func handler(ctx context.Context, event events.DynamoDBEvent) error {
	log.Printf("Processing batch of %d records from DynamoDB stream", len(event.Records))

	// IMPORTANT: This Lambda has reservedConcurrentExecutions=1 in CDK configuration
	// This ensures only one instance runs at a time, making read-modify-write safe
	// without additional locking mechanisms

	// First, fetch the current data from S3 and load into MemoryRepository
	memRepo, err := loadRepositoryFromS3(ctx)
	if err != nil {
		log.Printf("Error loading repository from S3: %v", err)
		// If file doesn't exist or error, start with empty repository
		memRepo = memrepo.NewMemoryRepository()
	}

	// Process each stream event using the repository methods
	for _, record := range event.Records {
		log.Printf("Processing record: EventID=%s, EventName=%s", record.EventID, record.EventName)

		switch record.EventName {
		case "INSERT", "MODIFY":
			// For INSERT and MODIFY, store the record using repository
			if domainRecord := convertStreamToDomainRecord(record); domainRecord != nil {
				if err := memRepo.Store(ctx, domainRecord); err != nil {
					log.Printf("Error storing record: %v", err)
				} else {
					log.Printf("Stored/Updated record: GroupID=%s, Hostname=%s", domainRecord.GroupID, domainRecord.Hostname)
				}
			}

		case "REMOVE":
			// For REMOVE, delete using repository
			pk := extractStringAttribute(record.Change.Keys, "pk")
			sk := extractStringAttribute(record.Change.Keys, "sk")
			if pk != "" && sk != "" {
				if err := memRepo.Delete(ctx, pk, sk); err != nil {
					if err != model.ErrNotFound {
						log.Printf("Error deleting record: %v", err)
					}
				} else {
					log.Printf("Removed record: GroupID=%s, Hostname=%s", pk, sk)
				}
			}

		default:
			log.Printf("Unknown event type: %s", record.EventName)
		}
	}

	// Save the updated repository data to S3
	if err := saveRepositoryToS3(ctx, memRepo); err != nil {
		log.Printf("Error saving repository to S3: %v", err)
		return fmt.Errorf("failed to save repository to S3: %w", err)
	}

	// Get final count for logging
	allRecords, _ := memRepo.List(ctx)
	log.Printf("Successfully processed %d stream records, S3 file now contains %d records",
		len(event.Records), len(allRecords))
	return nil
}

// loadRepositoryFromS3 loads data from S3 into a new MemoryRepository
func loadRepositoryFromS3(ctx context.Context) (*memrepo.MemoryRepository, error) {
	result, err := s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &s3BucketName,
		Key:    &s3DataKey,
	})

	if err != nil {
		// Check if the error is because the file doesn't exist
		// In that case, we return an empty repository
		return nil, fmt.Errorf("failed to get object from S3: %w", err)
	}
	defer result.Body.Close()

	// Read the body
	bodyBytes, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read S3 object body: %w", err)
	}

	// Create a new MemoryRepository from the JSON string
	repo, err := memrepo.NewMemoryRepositoryFromJsonString(string(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create repository from JSON: %w", err)
	}

	return repo, nil
}

// saveRepositoryToS3 saves the repository data to S3
func saveRepositoryToS3(ctx context.Context, repo *memrepo.MemoryRepository) error {
	// Get all records from the repository
	records, err := repo.List(ctx)
	if err != nil {
		return fmt.Errorf("failed to list records from repository: %w", err)
	}

	// Marshal to JSON in memrepo format (array of DomainRecord)
	jsonData, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal records: %w", err)
	}

	// Upload to S3 with appropriate headers for public access
	_, err = s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:       &s3BucketName,
		Key:          &s3DataKey,
		Body:         bytes.NewReader(jsonData),
		ContentType:  stringPtr("application/json"),
		CacheControl: stringPtr("max-age=60"), // Cache for 1 minute
	})

	if err != nil {
		return fmt.Errorf("failed to upload to S3: %w", err)
	}

	log.Printf("Successfully updated data file at %s/%s with %d records", s3BucketName, s3DataKey, len(records))
	return nil
}

// convertStreamToDomainRecord converts a DynamoDB stream record to a DomainRecord
func convertStreamToDomainRecord(record events.DynamoDBEventRecord) *model.DomainRecord {
	// We need the NewImage for INSERT/MODIFY events
	if record.Change.NewImage == nil {
		return nil
	}

	// Extract fields from the NewImage
	domainRecord := &model.DomainRecord{}

	// GroupID comes from pk
	if pk, ok := record.Change.NewImage["pk"]; ok && pk.DataType() == events.DataTypeString {
		domainRecord.GroupID = pk.String()
	}

	// Hostname comes from sk
	if sk, ok := record.Change.NewImage["sk"]; ok && sk.DataType() == events.DataTypeString {
		domainRecord.Hostname = sk.String()
	}

	// Owner
	if owner, ok := record.Change.NewImage["Owner"]; ok && owner.DataType() == events.DataTypeString {
		domainRecord.Owner = owner.String()
	}

	// Type
	if typeField, ok := record.Change.NewImage["Type"]; ok && typeField.DataType() == events.DataTypeString {
		domainRecord.Type = symgroup.SymmetryType(typeField.String())
	}

	// ValidateTime
	if validateTime, ok := record.Change.NewImage["ValidateTime"]; ok && validateTime.DataType() == events.DataTypeString {
		// Parse the time string
		if t, err := time.Parse(time.RFC3339, validateTime.String()); err == nil {
			domainRecord.ValidateTime = t
		}
	}

	// Validate we have the minimum required fields
	if domainRecord.GroupID == "" || domainRecord.Hostname == "" {
		return nil
	}

	return domainRecord
}

// extractStringAttribute extracts a string value from DynamoDB attribute map
func extractStringAttribute(attrs map[string]events.DynamoDBAttributeValue, key string) string {
	if attr, ok := attrs[key]; ok {
		if attr.DataType() == events.DataTypeString {
			return attr.String()
		}
	}
	return ""
}

// stringPtr returns a pointer to a string
func stringPtr(s string) *string {
	return &s
}

func main() {
	ctx := context.Background()

	// Load AWS configuration
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatalf("Failed to load AWS config: %v", err)
	}

	// Create DynamoDB client
	var client *dynamodb.Client
	if dynamoEndpoint != "" {
		// Use custom endpoint if specified
		client = dynamodb.NewFromConfig(cfg, func(o *dynamodb.Options) {
			o.BaseEndpoint = &dynamoEndpoint
		})
		log.Printf("DynamoDB client configured with custom endpoint: %s", dynamoEndpoint)
	} else {
		// Use default endpoint discovery
		client = dynamodb.NewFromConfig(cfg)
		log.Printf("DynamoDB client configured with default endpoint discovery")
	}

	// Initialize DynamoDB repository (optional, in case we need to query the table)
	repo = dynamorepo.NewDynamoRepository(client, dynamoTable)
	log.Printf("DynamoDB repository initialized with table: %s", dynamoTable)

	// Initialize S3 client
	s3Client = s3.NewFromConfig(cfg)
	log.Printf("S3 client initialized for bucket: %s", s3BucketName)

	// Start Lambda handler
	log.Printf("Starting DynamoDB Streams Lambda handler with S3 persistence...")
	lambda.Start(handler)
}