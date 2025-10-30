package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/mrled/suns/symval/internal/model"
	"github.com/mrled/suns/symval/internal/repository/dynamorepo"
	"github.com/mrled/suns/symval/internal/service/dnsclaims"
	"github.com/mrled/suns/symval/internal/symgroup"
	"github.com/mrled/suns/symval/internal/usecase/attestation"
)

var (
	dynamoEndpoint string
	dynamoTable    string
	repo           model.DomainRepository
	dnsService     *dnsclaims.Service
	attestUseCase  *attestation.AttestationUseCase
)

// AttestRequest represents the expected JSON payload for attestation
type AttestRequest struct {
	Owner   string   `json:"owner"`
	Type    string   `json:"type"`
	Domains []string `json:"domains"`
}

// AttestResponse represents the JSON response for attestation
type AttestResponse struct {
	IsValid      bool     `json:"isValid"`
	ExpectedID   string   `json:"expectedId"`
	GroupIDCount int      `json:"groupIdCount"`
	ErrorMessage string   `json:"errorMessage,omitempty"`
	Message      string   `json:"message,omitempty"`
}

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

func handler(ctx context.Context, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	// Log the incoming request details for debugging
	// API Gateway v2 uses different field names
	log.Printf("Incoming request: Method=%s, Path=%s, RawPath=%s, PathParameters=%+v",
		request.RequestContext.HTTP.Method, request.RequestContext.HTTP.Path,
		request.RawPath, request.PathParameters)

	// For API Gateway v2, the path is in RequestContext.HTTP.Path
	path := request.RequestContext.HTTP.Path
	if path == "" {
		path = request.RawPath
	}

	// Remove the /api prefix if present since we're matching on the API path
	path = strings.TrimPrefix(path, "/api")
	log.Printf("Processing path: %s", path)

	// Route based on the path
	// The path should be something like /v1/attest after removing /api prefix
	switch {
	case strings.HasSuffix(path, "/v1/attest") || path == "/v1/attest":
		return handleAttest(ctx, request)
	// Add more endpoints here as needed, for example:
	// case strings.HasSuffix(path, "/v1/verify") || path == "/v1/verify":
	//	return handleVerify(ctx, request)
	// case strings.HasSuffix(path, "/v1/health") || path == "/v1/health":
	//	return handleHealth(ctx, request)
	default:
		log.Printf("Path not matched. Full request details: %+v", request)
		return errorResponseV2(404, fmt.Sprintf("Unknown endpoint: %s", path))
	}
}

func handleAttest(ctx context.Context, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	// Log the HTTP method for debugging
	httpMethod := request.RequestContext.HTTP.Method
	log.Printf("handleAttest: Method='%s', Path='%s'", httpMethod, request.RequestContext.HTTP.Path)

	// Validate HTTP method
	if httpMethod != "POST" {
		log.Printf("Method validation failed. Received method: %s", httpMethod)
		return errorResponseV2(405, fmt.Sprintf("Method not allowed. Only POST is supported for this endpoint (received: %s)", httpMethod))
	}

	// Parse the request body
	var attestReq AttestRequest
	if err := json.Unmarshal([]byte(request.Body), &attestReq); err != nil {
		return errorResponseV2(400, fmt.Sprintf("Invalid request body: %v", err))
	}

	// Validate required fields
	if attestReq.Owner == "" {
		return errorResponseV2(400, "owner field is required")
	}
	if attestReq.Type == "" {
		return errorResponseV2(400, "type field is required")
	}
	if len(attestReq.Domains) < 1 {
		return errorResponseV2(400, "at least one domain is required")
	}

	// Convert type name to type code (similar to attest command)
	typeName := strings.ToLower(attestReq.Type)
	typeCode, ok := symgroup.TypeNameToCode[typeName]
	if !ok {
		// Check if it's already a valid type code
		if _, codeExists := symgroup.TypeCodeToName[typeName]; codeExists {
			typeCode = typeName
		} else {
			return errorResponseV2(400, "invalid symmetry type. Valid types: palindrome (a), flip180 (b), doubleflip180 (c), mirrortext (d), mirrornames (e), antonymnames (f)")
		}
	}

	symmetryType := symgroup.SymmetryType(typeCode)

	// Perform attestation
	result, err := attestUseCase.Attest(attestReq.Owner, symmetryType, attestReq.Domains)
	if err != nil {
		log.Printf("Attestation failed: %v", err)
		return errorResponseV2(500, fmt.Sprintf("attestation failed: %v", err))
	}

	// Build response
	response := AttestResponse{
		IsValid:      result.IsValid,
		ExpectedID:   result.ExpectedID,
		GroupIDCount: len(result.GroupIDs),
		ErrorMessage: result.ErrorMessage,
	}

	if result.IsValid {
		response.Message = "Attestation PASSED: The domains form a valid symmetric group"
	} else {
		response.Message = "Attestation FAILED"
	}

	// Marshal response to JSON
	responseBody, err := json.Marshal(response)
	if err != nil {
		log.Printf("Failed to marshal response: %v", err)
		return errorResponseV2(500, "failed to generate response")
	}

	return events.APIGatewayV2HTTPResponse{
		StatusCode: 200,
		Body:       string(responseBody),
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}, nil
}

// errorResponseV2 creates a standardized error response for API Gateway v2
func errorResponseV2(statusCode int, message string) (events.APIGatewayV2HTTPResponse, error) {
	errorBody := map[string]string{
		"error": message,
	}
	body, _ := json.Marshal(errorBody)

	return events.APIGatewayV2HTTPResponse{
		StatusCode: statusCode,
		Body:       string(body),
		Headers: map[string]string{
			"Content-Type": "application/json",
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

	// Initialize DynamoDB repository
	repo = dynamorepo.NewDynamoRepository(client, dynamoTable)
	log.Printf("DynamoDB repository initialized with table: %s", dynamoTable)

	// Initialize DNS service
	dnsService = dnsclaims.NewService()
	log.Printf("DNS claims service initialized")

	// Initialize attestation use case with DNS service and repository
	attestUseCase = attestation.NewAttestationUseCase(dnsService, repo)
	log.Printf("Attestation use case initialized")

	// Verify DynamoDB connection
	records, err := repo.List(ctx)
	if err != nil {
		log.Printf("Warning: Failed to list records during startup: %v", err)
	} else {
		log.Printf("Successfully connected to DynamoDB. Found %d records", len(records))
	}

	// Start Lambda handler
	log.Printf("Starting Lambda handler...")
	lambda.Start(handler)
}
