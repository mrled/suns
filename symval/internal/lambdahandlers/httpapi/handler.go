package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/mrled/suns/symval/internal/logger"
	"github.com/mrled/suns/symval/internal/model"
	"github.com/mrled/suns/symval/internal/repository/dynamorepo"
	"github.com/mrled/suns/symval/internal/service/dnsclaims"
	"github.com/mrled/suns/symval/internal/symgroup"
	"github.com/mrled/suns/symval/internal/usecase/attestation"
)

// Handler holds the dependencies for the httpapi Lambda handler
type Handler struct {
	repo          model.DomainRepository
	dnsService    *dnsclaims.Service
	attestUseCase *attestation.AttestationUseCase
	log           *slog.Logger
}

// AttestRequest represents the expected JSON payload for attestation
type AttestRequest struct {
	Owner   string   `json:"owner"`
	Type    string   `json:"type"`
	Domains []string `json:"domains"`
}

// AttestResponse represents the JSON response for attestation
type AttestResponse struct {
	IsValid      bool   `json:"isValid"`
	ExpectedID   string `json:"expectedId"`
	GroupIDCount int    `json:"groupIdCount"`
	ErrorMessage string `json:"errorMessage,omitempty"`
	Message      string `json:"message,omitempty"`
}

// NewHandler creates a new httpapi handler with initialized dependencies
func NewHandler() (*Handler, error) {
	// Initialize logger with executable name for filtering
	log := logger.NewDefaultLogger()
	log = logger.WithExecutable(log, "httpapi")
	logger.SetDefault(log)

	// Optional endpoint override for local development or testing
	dynamoEndpoint := os.Getenv("DYNAMODB_ENDPOINT")
	if dynamoEndpoint != "" {
		log.Info("Using custom DynamoDB endpoint", slog.String("endpoint", dynamoEndpoint))
	} else {
		// When not using a custom endpoint, AWS_REGION is required
		awsRegion := os.Getenv("AWS_REGION")
		if awsRegion == "" {
			return nil, fmt.Errorf("AWS_REGION environment variable is required when DYNAMODB_ENDPOINT is not set")
		}
		log.Info("Using AWS region", slog.String("region", awsRegion))
		log.Info("Using default DynamoDB endpoint discovery")
	}

	dynamoTable := os.Getenv("DYNAMODB_TABLE")
	if dynamoTable == "" {
		return nil, fmt.Errorf("DYNAMODB_TABLE environment variable is required")
	}
	log.Info("Using DynamoDB table", slog.String("table", dynamoTable))

	ctx := context.Background()

	// Load AWS configuration
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Error("Failed to load AWS config", slog.String("error", err.Error()))
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
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

	// Initialize DynamoDB repository
	repo := dynamorepo.NewDynamoRepository(client, dynamoTable)
	log.Info("DynamoDB repository initialized", slog.String("table", dynamoTable))

	// Initialize DNS service
	dnsService := dnsclaims.NewService()
	log.Info("DNS claims service initialized")

	// Initialize attestation use case with DNS service and repository
	attestUseCase := attestation.NewAttestationUseCase(dnsService, repo)
	log.Info("Attestation use case initialized")

	// Verify DynamoDB connection
	records, err := repo.List(ctx)
	if err != nil {
		log.Warn("Failed to list records during startup", slog.String("error", err.Error()))
	} else {
		log.Info("Successfully connected to DynamoDB", slog.Int("record_count", len(records)))
	}

	return &Handler{
		repo:          repo,
		dnsService:    dnsService,
		attestUseCase: attestUseCase,
		log:           log,
	}, nil
}

// Handle processes API Gateway HTTP requests
func (h *Handler) Handle(ctx context.Context, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	// Create a logger with Lambda context
	requestLogger := logger.WithLambda(h.log,
		os.Getenv("AWS_LAMBDA_FUNCTION_NAME"),
		os.Getenv("AWS_LAMBDA_FUNCTION_VERSION"),
		request.RequestContext.RequestID)

	// Log the incoming request details for debugging
	// API Gateway v2 uses different field names
	requestLogger.Info("Incoming request",
		slog.String("method", request.RequestContext.HTTP.Method),
		slog.String("path", request.RequestContext.HTTP.Path),
		slog.String("raw_path", request.RawPath),
		slog.Any("path_parameters", request.PathParameters))

	// For API Gateway v2, the path is in RequestContext.HTTP.Path
	path := request.RequestContext.HTTP.Path
	if path == "" {
		path = request.RawPath
	}

	// Remove the /api prefix if present since we're matching on the API path
	path = strings.TrimPrefix(path, "/api")
	requestLogger.Debug("Processing path", slog.String("path", path))

	// Route based on the path
	// The path should be something like /v1/attest after removing /api prefix
	switch {
	case strings.HasSuffix(path, "/v1/attest") || path == "/v1/attest":
		return h.handleAttest(ctx, request)
	// Add more endpoints here as needed, for example:
	// case strings.HasSuffix(path, "/v1/verify") || path == "/v1/verify":
	//	return h.handleVerify(ctx, request)
	// case strings.HasSuffix(path, "/v1/health") || path == "/v1/health":
	//	return h.handleHealth(ctx, request)
	default:
		requestLogger.Warn("Path not matched", slog.Any("request", request))
		return errorResponseV2(404, fmt.Sprintf("Unknown endpoint: %s", path))
	}
}

func (h *Handler) handleAttest(ctx context.Context, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	// Create a logger with Lambda context for this request
	requestLogger := logger.WithLambda(h.log,
		os.Getenv("AWS_LAMBDA_FUNCTION_NAME"),
		os.Getenv("AWS_LAMBDA_FUNCTION_VERSION"),
		request.RequestContext.RequestID)

	// Log the HTTP method for debugging
	httpMethod := request.RequestContext.HTTP.Method
	requestLogger.Debug("handleAttest called",
		slog.String("method", httpMethod),
		slog.String("path", request.RequestContext.HTTP.Path))

	// Validate HTTP method
	if httpMethod != "POST" {
		requestLogger.Warn("Method validation failed", slog.String("received_method", httpMethod))
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
			return errorResponseV2(400, "invalid symmetry type. "+symgroup.ValidSymmetryTypesText())
		}
	}

	symmetryType := symgroup.SymmetryType(typeCode)

	// Perform attestation
	result, err := h.attestUseCase.Attest(attestReq.Owner, symmetryType, attestReq.Domains)
	if err != nil {
		requestLogger.Error("Attestation failed", slog.String("error", err.Error()))
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
		requestLogger.Error("Failed to marshal response", slog.String("error", err.Error()))
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
