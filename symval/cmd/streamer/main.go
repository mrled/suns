package main

import (
	"context"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/mrled/suns/symval/internal/adapter/dynamostream"
	"github.com/mrled/suns/symval/internal/adapter/s3materializedview"
	"github.com/mrled/suns/symval/internal/model"
	"github.com/mrled/suns/symval/internal/repository/dynamorepo"
	"github.com/mrled/suns/symval/internal/repository/memrepo"
)

var (
	dynamoEndpoint string
	dynamoTable    string
	s3BucketName   string
	s3DataKey      string
	repo           model.DomainRepository
	s3Client       *s3.Client
	s3View         *s3materializedview.S3MaterializedView
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
	memRepo, err := s3View.Load(ctx)
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
			domainRecord, err := dynamostream.ConvertToDomainRecord(record.Change.NewImage)
			if err != nil {
				log.Printf("Error converting stream record: %v", err)
				continue
			}

			if err := memRepo.Store(ctx, domainRecord); err != nil {
				log.Printf("Error storing record: %v", err)
			} else {
				log.Printf("Stored/Updated record: GroupID=%s, Hostname=%s", domainRecord.GroupID, domainRecord.Hostname)
			}

		case "REMOVE":
			// For REMOVE, delete using repository
			pk := dynamostream.ExtractStringAttribute(record.Change.Keys, "pk")
			sk := dynamostream.ExtractStringAttribute(record.Change.Keys, "sk")
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
	if err := s3View.Save(ctx, memRepo); err != nil {
		log.Printf("Error saving repository to S3: %v", err)
		return err
	}

	// Get final count for logging
	allRecords, _ := memRepo.List(ctx)
	log.Printf("Successfully processed %d stream records, S3 file now contains %d records",
		len(event.Records), len(allRecords))
	return nil
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

	// Initialize S3 materialized view adapter
	s3View = s3materializedview.New(s3Client, s3BucketName, s3DataKey)
	log.Printf("S3 materialized view initialized for %s/%s", s3BucketName, s3DataKey)

	// Start Lambda handler
	log.Printf("Starting DynamoDB Streams Lambda handler with S3 persistence...")
	lambda.Start(handler)
}