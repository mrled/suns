package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/mrled/suns/symval/internal/model"
	"github.com/mrled/suns/symval/internal/repository/dynamorepo"
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

	// Process each record
	for _, record := range event.Records {
		log.Printf("Processing record: EventID=%s, EventName=%s, EventSource=%s",
			record.EventID, record.EventName, record.EventSourceArn)
	}

	// After processing the stream events, fetch all current records from DynamoDB
	// and save them to S3 in memrepo format
	if err := updateS3DataFile(ctx); err != nil {
		log.Printf("Error updating S3 data file: %v", err)
		return fmt.Errorf("failed to update S3 data file: %w", err)
	}

	log.Printf("Successfully processed %d stream records and updated S3 data file", len(event.Records))
	return nil
}

// updateS3DataFile fetches all records from DynamoDB and saves them to S3 in memrepo format
func updateS3DataFile(ctx context.Context) error {
	// Fetch all records from DynamoDB
	records, err := repo.List(ctx)
	if err != nil {
		return fmt.Errorf("failed to list records from DynamoDB: %w", err)
	}

	log.Printf("Found %d records in DynamoDB", len(records))

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