package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/mrled/suns/symval/internal/model"
	"github.com/mrled/suns/symval/internal/repository/dynamorepo"
)

var (
	dynamoEndpoint string
	dynamoTable    string
	repo           model.DomainRepository
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
}

// StreamRecord represents a processed record from the stream
type StreamRecord struct {
	EventName string
	Keys      map[string]types.AttributeValue
	NewImage  map[string]types.AttributeValue
	OldImage  map[string]types.AttributeValue
}

func handler(ctx context.Context, event events.DynamoDBEvent) error {
	log.Printf("Processing batch of %d records from DynamoDB stream", len(event.Records))

	for _, record := range event.Records {
		// Log the event details
		log.Printf("Processing record: EventID=%s, EventName=%s, EventSource=%s",
			record.EventID, record.EventName, record.EventSourceArn)

		// Process based on event type
		switch record.EventName {
		case "INSERT":
			if err := handleInsert(ctx, record); err != nil {
				log.Printf("Error handling INSERT event: %v", err)
				// Return error to fail the batch and retry
				return fmt.Errorf("failed to handle INSERT: %w", err)
			}

		case "MODIFY":
			if err := handleModify(ctx, record); err != nil {
				log.Printf("Error handling MODIFY event: %v", err)
				// Return error to fail the batch and retry
				return fmt.Errorf("failed to handle MODIFY: %w", err)
			}

		case "REMOVE":
			if err := handleRemove(ctx, record); err != nil {
				log.Printf("Error handling REMOVE event: %v", err)
				// Return error to fail the batch and retry
				return fmt.Errorf("failed to handle REMOVE: %w", err)
			}

		default:
			log.Printf("Unknown event type: %s", record.EventName)
		}
	}

	log.Printf("Successfully processed %d records", len(event.Records))
	return nil
}

func handleInsert(ctx context.Context, record events.DynamoDBEventRecord) error {
	// Extract key attributes
	pk := extractStringAttribute(record.Change.Keys, "pk")
	sk := extractStringAttribute(record.Change.Keys, "sk")

	log.Printf("INSERT: pk=%s, sk=%s", pk, sk)

	// Extract and log new image data
	if record.Change.NewImage != nil {
		newImageJSON, _ := json.Marshal(record.Change.NewImage)
		log.Printf("New image: %s", string(newImageJSON))

		// Here you can add custom logic for handling inserts
		// For example, sending notifications, updating caches, etc.
	}

	return nil
}

func handleModify(ctx context.Context, record events.DynamoDBEventRecord) error {
	// Extract key attributes
	pk := extractStringAttribute(record.Change.Keys, "pk")
	sk := extractStringAttribute(record.Change.Keys, "sk")

	log.Printf("MODIFY: pk=%s, sk=%s", pk, sk)

	// Log old and new images for comparison
	if record.Change.OldImage != nil {
		oldImageJSON, _ := json.Marshal(record.Change.OldImage)
		log.Printf("Old image: %s", string(oldImageJSON))
	}

	if record.Change.NewImage != nil {
		newImageJSON, _ := json.Marshal(record.Change.NewImage)
		log.Printf("New image: %s", string(newImageJSON))

		// Here you can add custom logic for handling modifications
		// For example, detecting specific changes and triggering actions
	}

	return nil
}

func handleRemove(ctx context.Context, record events.DynamoDBEventRecord) error {
	// Extract key attributes
	pk := extractStringAttribute(record.Change.Keys, "pk")
	sk := extractStringAttribute(record.Change.Keys, "sk")

	log.Printf("REMOVE: pk=%s, sk=%s", pk, sk)

	// Log old image data before deletion
	if record.Change.OldImage != nil {
		oldImageJSON, _ := json.Marshal(record.Change.OldImage)
		log.Printf("Removed image: %s", string(oldImageJSON))

		// Here you can add custom logic for handling deletions
		// For example, archiving data, cleaning up related resources, etc.
	}

	return nil
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

	// Start Lambda handler
	log.Printf("Starting DynamoDB Streams Lambda handler...")
	lambda.Start(handler)
}