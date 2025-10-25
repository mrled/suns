package main

import (
	"context"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/mrled/suns/symval/internal/model"
	"github.com/mrled/suns/symval/internal/repository/dynamorepo"
)

var (
	dynamoURL string
	repo      model.DomainRepository
)

func init() {
	dynamoURL = os.Getenv("DYNAMO_URL")
	if dynamoURL == "" {
		log.Fatal("DYNAMO_URL environment variable is required")
	}
	log.Printf("Using DynamoDB URL: %s", dynamoURL)
}

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       "Hello from webhook Lambda!",
		Headers: map[string]string{
			"Content-Type": "text/plain",
		},
	}, nil
}

func main() {
	ctx := context.Background()

	// Load AWS configuration
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatalf("Failed to load AWS config: %v", err)
	}

	// Create DynamoDB client
	client := dynamodb.NewFromConfig(cfg, func(o *dynamodb.Options) {
		o.BaseEndpoint = &dynamoURL
	})

	// Initialize repository (TODO: implement dynamorepo.New())
	_ = client // Will be used when dynamorepo is implemented
	repo = nil // Placeholder until dynamorepo.New() is implemented
	log.Printf("DynamoDB repository stub: %s", dynamorepo.NotImplemented)

	// Stub operation: List records to verify connection
	if repo != nil {
		records, err := repo.List(ctx)
		if err != nil {
			log.Printf("Warning: Failed to list records: %v", err)
		} else {
			log.Printf("Successfully connected to DynamoDB. Found %d records", len(records))
		}
	} else {
		log.Printf("Skipping connection test - repository not yet implemented")
	}

	// Start Lambda handler
	lambda.Start(handler)
}
