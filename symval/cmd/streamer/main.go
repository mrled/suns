package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/mrled/suns/symval/internal/adapter/s3materializedview"
	"github.com/mrled/suns/symval/internal/logger"
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
	log              *slog.Logger
)

func init() {
	// Initialize logger with executable name for filtering
	log = logger.NewDefaultLogger()
	log = logger.WithExecutable(log, "streamer")
	logger.SetDefault(log)

	// Optional endpoint override for local development or testing
	dynamoEndpoint = os.Getenv("DYNAMODB_ENDPOINT")
	if dynamoEndpoint != "" {
		log.Info("Using custom DynamoDB endpoint", slog.String("endpoint", dynamoEndpoint))
	} else {
		// When not using a custom endpoint, AWS_REGION is required
		awsRegion := os.Getenv("AWS_REGION")
		if awsRegion == "" {
			log.Error("AWS_REGION environment variable is required when DYNAMODB_ENDPOINT is not set")
			os.Exit(1)
		}
		log.Info("Using AWS region", slog.String("region", awsRegion))
		log.Info("Using default DynamoDB endpoint discovery")
	}

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


func handler(ctx context.Context, event events.DynamoDBEvent) error {
	// Delegate all processing to the applystream service
	err := streamerService.ProcessStreamBatch(ctx, event.Records)
	if err != nil {
		log.Error("Stream processing failed",
			slog.String("error", err.Error()),
			slog.Bool("notify", true))
	}
	return err
}


func main() {
	ctx := context.Background()

	// Load AWS configuration
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Error("Failed to load AWS config", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// Create DynamoDB client
	var client *dynamodb.Client
	if dynamoEndpoint != "" {
		// Use custom endpoint if specified
		client = dynamodb.NewFromConfig(cfg, func(o *dynamodb.Options) {
			o.BaseEndpoint = &dynamoEndpoint
		})
		log.Info("DynamoDB client configured", slog.String("endpoint", dynamoEndpoint))
	} else {
		// Use default endpoint discovery
		client = dynamodb.NewFromConfig(cfg)
		log.Info("DynamoDB client configured with default endpoint discovery")
	}

	// Note: DynamoDB repository is created but not actively used in the streamer
	// It's kept here for potential future use or debugging purposes
	repo := dynamorepo.NewDynamoRepository(client, dynamoTable)
	log.Info("DynamoDB repository initialized (for reference only)", slog.String("table", dynamoTable))
	_ = repo // Suppress unused variable warning

	// Initialize S3 client
	s3Client = s3.NewFromConfig(cfg)
	log.Info("S3 client initialized", slog.String("bucket", s3BucketName))

	// Initialize S3 materialized view adapter
	s3View := s3materializedview.New(s3Client, s3BucketName, s3DataKey)
	log.Info("S3 materialized view initialized",
		slog.String("bucket", s3BucketName),
		slog.String("key", s3DataKey))

	// Initialize the applystream service
	streamerService = applystream.New(s3View)
	log.Info("Stream processing service initialized")

	// Start Lambda handler
	log.Info("Starting DynamoDB Streams Lambda handler with S3 persistence")
	lambda.Start(handler)
}