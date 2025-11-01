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
	"github.com/mrled/suns/symval/internal/adapter/s3materializedview"
	"github.com/mrled/suns/symval/internal/repository/dynamorepo"
	"github.com/mrled/suns/symval/internal/service/applystream"
)

var (
	dynamoEndpoint   string
	dynamoTable      string
	s3BucketName     string
	s3DataKey        string
	s3Client         *s3.Client
	streamerService  *applystream.Service
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
	// Delegate all processing to the applystream service
	return streamerService.ProcessStreamBatch(ctx, event.Records)
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

	// Note: DynamoDB repository is created but not actively used in the streamer
	// It's kept here for potential future use or debugging purposes
	repo := dynamorepo.NewDynamoRepository(client, dynamoTable)
	log.Printf("DynamoDB repository initialized with table: %s (for reference only)", dynamoTable)
	_ = repo // Suppress unused variable warning

	// Initialize S3 client
	s3Client = s3.NewFromConfig(cfg)
	log.Printf("S3 client initialized for bucket: %s", s3BucketName)

	// Initialize S3 materialized view adapter
	s3View := s3materializedview.New(s3Client, s3BucketName, s3DataKey)
	log.Printf("S3 materialized view initialized for %s/%s", s3BucketName, s3DataKey)

	// Initialize the applystream service
	streamerService = applystream.New(s3View)
	log.Printf("Stream processing service initialized")

	// Start Lambda handler
	log.Printf("Starting DynamoDB Streams Lambda handler with S3 persistence...")
	lambda.Start(handler)
}