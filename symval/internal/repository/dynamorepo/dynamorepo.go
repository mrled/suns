package dynamorepo

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/mrled/suns/symval/internal/model"
)

// DynamoRepository is a DynamoDB implementation of DomainRepository
type DynamoRepository struct {
	client    *dynamodb.Client
	tableName string
}

// NewDynamoRepository creates a new DynamoDB-backed repository
func NewDynamoRepository(client *dynamodb.Client, tableName string) *DynamoRepository {
	return &DynamoRepository{
		client:    client,
		tableName: tableName,
	}
}

// Store saves domain data to DynamoDB
// Uses group ID as the PK and hostname as the SK
// If the record already exists, it updates the timestamp
func (r *DynamoRepository) Store(ctx context.Context, data *model.DomainRecord) error {
	if data == nil {
		return fmt.Errorf("domain data cannot be nil")
	}

	// Convert domain model to DTO
	dto := FromDomain(data)

	// Marshal the DTO into DynamoDB attribute values
	item, err := attributevalue.MarshalMap(dto)
	if err != nil {
		return fmt.Errorf("failed to marshal domain record: %w", err)
	}

	// Use PutItem without condition to allow overwrites (updating timestamp)
	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item:      item,
	})

	if err != nil {
		return fmt.Errorf("failed to store domain record: %w", err)
	}

	return nil
}

// Get retrieves domain data by group ID and hostname from DynamoDB
func (r *DynamoRepository) Get(ctx context.Context, groupID, hostname string) (*model.DomainRecord, error) {
	result, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"pk": &types.AttributeValueMemberS{Value: groupID},
			"sk": &types.AttributeValueMemberS{Value: hostname},
		},
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get domain record: %w", err)
	}

	if result.Item == nil {
		return nil, model.ErrNotFound
	}

	var dto DynamoDTO
	if err := attributevalue.UnmarshalMap(result.Item, &dto); err != nil {
		return nil, fmt.Errorf("failed to unmarshal domain record: %w", err)
	}

	return dto.ToDomain(), nil
}

// List retrieves all domain data from DynamoDB
func (r *DynamoRepository) List(ctx context.Context) ([]*model.DomainRecord, error) {
	// Use Scan to retrieve all items
	// Note: For production use with large tables, consider using pagination
	result, err := r.client.Scan(ctx, &dynamodb.ScanInput{
		TableName: aws.String(r.tableName),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to scan domain records: %w", err)
	}

	var dtos []*DynamoDTO
	for _, item := range result.Items {
		var dto DynamoDTO
		if err := attributevalue.UnmarshalMap(item, &dto); err != nil {
			return nil, fmt.Errorf("failed to unmarshal domain record: %w", err)
		}
		dtos = append(dtos, &dto)
	}

	return ToDomainList(dtos), nil
}

// Delete removes domain data by group ID and hostname from DynamoDB
func (r *DynamoRepository) Delete(ctx context.Context, groupID, hostname string) error {
	// Use ConditionExpression to ensure the item exists before deleting
	// This matches the behavior of MemoryRepository.Delete which returns ErrNotFound
	_, err := r.client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"pk": &types.AttributeValueMemberS{Value: groupID},
			"sk": &types.AttributeValueMemberS{Value: hostname},
		},
		ConditionExpression: aws.String("attribute_exists(pk) AND attribute_exists(sk)"),
	})

	if err != nil {
		// Check if the error is a conditional check failure (item doesn't exist)
		var ccfe *types.ConditionalCheckFailedException
		if errors.As(err, &ccfe) {
			return model.ErrNotFound
		}
		return fmt.Errorf("failed to delete domain record: %w", err)
	}

	return nil
}
