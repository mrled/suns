package repository

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/mrled/suns/symval/internal/model"
	"github.com/mrled/suns/symval/internal/repository/dynamorepo"
	"github.com/mrled/suns/symval/internal/repository/memrepo"
)

// RepositoryConfig holds configuration for creating a repository
type RepositoryConfig struct {
	// FilePath for JSON file persistence (mutually exclusive with DynamoDB options)
	FilePath string

	// DynamoTable is the DynamoDB table name for persistence
	DynamoTable string

	// DynamoEndpoint is an optional custom DynamoDB endpoint URL
	DynamoEndpoint string
}

// NewRepository creates a DomainRepository based on the provided configuration.
// It returns an error if neither file nor DynamoDB configuration is provided,
// or if repository creation fails.
//
// The function prints informational messages about which persistence mechanism
// is being used to help with debugging and user awareness.
func NewRepository(ctx context.Context, cfg RepositoryConfig) (model.DomainRepository, error) {
	if cfg.DynamoTable != "" {
		// Use DynamoDB persistence
		awsCfg, err := config.LoadDefaultConfig(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to load AWS config: %w", err)
		}

		// Create DynamoDB client
		var client *dynamodb.Client
		if cfg.DynamoEndpoint != "" {
			// Use custom endpoint if specified
			client = dynamodb.NewFromConfig(awsCfg, func(o *dynamodb.Options) {
				o.BaseEndpoint = &cfg.DynamoEndpoint
			})
			fmt.Printf("Using DynamoDB endpoint: %s\n", cfg.DynamoEndpoint)
		} else {
			// Use default endpoint discovery
			client = dynamodb.NewFromConfig(awsCfg)
		}

		repo := dynamorepo.NewDynamoRepository(client, cfg.DynamoTable)
		fmt.Printf("Using DynamoDB table: %s\n", cfg.DynamoTable)
		return repo, nil
	}

	if cfg.FilePath != "" {
		// Use JSON file persistence
		memRepo, err := memrepo.NewMemoryRepositoryWithPersistence(cfg.FilePath)
		if err != nil {
			return nil, fmt.Errorf("failed to create repository: %w", err)
		}
		fmt.Printf("Using JSON persistence: %s\n", cfg.FilePath)
		return memRepo, nil
	}

	return nil, fmt.Errorf("must specify either FilePath or DynamoTable in repository configuration")
}
