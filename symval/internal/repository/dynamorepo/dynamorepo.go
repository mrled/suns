package dynamorepo

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

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

// UnconditionalStore saves domain data to DynamoDB unconditionally. Returns new rev.
func (r *DynamoRepository) UnconditionalStore(ctx context.Context, data *model.DomainRecord) (int64, error) {
	if data == nil {
		return 0, fmt.Errorf("domain data cannot be nil")
	}

	// Get existing record to determine the new revision
	existing, err := r.Get(ctx, data.GroupID, data.Hostname)
	if err != nil && err != model.ErrNotFound {
		return 0, fmt.Errorf("failed to get existing record: %w", err)
	}

	// Set revision
	if existing != nil {
		data.Rev = existing.Rev + 1
	} else {
		data.Rev = 1
	}

	// Convert domain model to DTO
	dto := FromDomain(data)

	// Marshal the DTO into DynamoDB attribute values
	item, err := attributevalue.MarshalMap(dto)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal domain record: %w", err)
	}

	// Use PutItem without condition to allow overwrites
	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item:      item,
	})

	if err != nil {
		return 0, fmt.Errorf("failed to store domain record: %w", err)
	}

	return data.Rev, nil
}

// Upsert saves domain data with automatic revision increment using UpdateItem. Returns new rev.
func (r *DynamoRepository) Upsert(ctx context.Context, data *model.DomainRecord) (int64, error) {
	if data == nil {
		return 0, fmt.Errorf("domain data cannot be nil")
	}

	// Use UpdateItem with SET to atomically increment revision
	result, err := r.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"pk": &types.AttributeValueMemberS{Value: data.GroupID},
			"sk": &types.AttributeValueMemberS{Value: data.Hostname},
		},
		UpdateExpression: aws.String("SET #owner = :owner, #type = :type, #validateTime = :validateTime, #rev = if_not_exists(#rev, :zero) + :one"),
		ExpressionAttributeNames: map[string]string{
			"#owner":        "Owner",
			"#type":         "Type",
			"#validateTime": "ValidateTime",
			"#rev":          "Rev",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":owner":        &types.AttributeValueMemberS{Value: data.Owner},
			":type":         &types.AttributeValueMemberS{Value: string(data.Type)},
			":validateTime": &types.AttributeValueMemberS{Value: data.ValidateTime.Format(time.RFC3339Nano)},
			":zero":         &types.AttributeValueMemberN{Value: "0"},
			":one":          &types.AttributeValueMemberN{Value: "1"},
		},
		ReturnValues: types.ReturnValueUpdatedNew,
	})

	if err != nil {
		return 0, fmt.Errorf("failed to upsert domain record: %w", err)
	}

	// Extract the new revision from the returned item
	if revAttr, ok := result.Attributes["Rev"]; ok {
		if revNum, ok := revAttr.(*types.AttributeValueMemberN); ok {
			rev, err := strconv.ParseInt(revNum.Value, 10, 64)
			if err != nil {
				return 0, fmt.Errorf("failed to parse revision: %w", err)
			}
			return rev, nil
		}
	}

	return 0, fmt.Errorf("failed to get revision from response")
}

// SetValidationIfUnchanged updates validation time only if revision matches. Returns new rev.
func (r *DynamoRepository) SetValidationIfUnchanged(ctx context.Context, data *model.DomainRecord, snapshotRev int64) (int64, error) {
	if data == nil {
		return 0, fmt.Errorf("domain data cannot be nil")
	}

	// Use UpdateItem with condition expression to check revision
	// Handle missing Rev attribute by treating it as rev 0 for backward compatibility
	result, err := r.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"pk": &types.AttributeValueMemberS{Value: data.GroupID},
			"sk": &types.AttributeValueMemberS{Value: data.Hostname},
		},
		UpdateExpression:    aws.String("SET #validateTime = :validateTime, #rev = if_not_exists(#rev, :zero) + :one"),
		ConditionExpression: aws.String("(attribute_not_exists(#rev) AND :snapshotRev = :zero) OR (#rev = :snapshotRev)"),
		ExpressionAttributeNames: map[string]string{
			"#validateTime": "ValidateTime",
			"#rev":          "Rev",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":validateTime": &types.AttributeValueMemberS{Value: data.ValidateTime.Format(time.RFC3339Nano)},
			":snapshotRev":  &types.AttributeValueMemberN{Value: strconv.FormatInt(snapshotRev, 10)},
			":zero":         &types.AttributeValueMemberN{Value: "0"},
			":one":          &types.AttributeValueMemberN{Value: "1"},
		},
		ReturnValues: types.ReturnValueAllNew,
	})

	if err != nil {
		// Check if the error is a conditional check failure
		var ccfe *types.ConditionalCheckFailedException
		if errors.As(err, &ccfe) {
			return 0, model.ErrRevConflict
		}
		return 0, fmt.Errorf("failed to update validation: %w", err)
	}

	// Extract the new revision from the returned item
	if revAttr, ok := result.Attributes["Rev"]; ok {
		if revNum, ok := revAttr.(*types.AttributeValueMemberN); ok {
			rev, err := strconv.ParseInt(revNum.Value, 10, 64)
			if err != nil {
				return 0, fmt.Errorf("failed to parse revision: %w", err)
			}
			return rev, nil
		}
	}

	return 0, fmt.Errorf("failed to get revision from response")
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

// UnconditionalDelete removes domain data by group ID and hostname from DynamoDB unconditionally
func (r *DynamoRepository) UnconditionalDelete(ctx context.Context, groupID, hostname string) error {
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

// DeleteIfUnchanged removes domain data only if revision matches
func (r *DynamoRepository) DeleteIfUnchanged(ctx context.Context, groupID, hostname string, snapshotRev int64) error {
	// Use ConditionExpression to check both existence and revision
	// Handle missing Rev attribute by treating it as rev 0 for backward compatibility
	_, err := r.client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"pk": &types.AttributeValueMemberS{Value: groupID},
			"sk": &types.AttributeValueMemberS{Value: hostname},
		},
		ConditionExpression: aws.String("attribute_exists(pk) AND attribute_exists(sk) AND ((attribute_not_exists(#rev) AND :snapshotRev = :zero) OR (#rev = :snapshotRev))"),
		ExpressionAttributeNames: map[string]string{
			"#rev": "Rev",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":snapshotRev": &types.AttributeValueMemberN{Value: strconv.FormatInt(snapshotRev, 10)},
			":zero":        &types.AttributeValueMemberN{Value: "0"},
		},
	})

	if err != nil {
		// Check if the error is a conditional check failure
		var ccfe *types.ConditionalCheckFailedException
		if errors.As(err, &ccfe) {
			// Need to check if it's because the item doesn't exist or revision mismatch
			// Try to get the item to distinguish between the two
			existing, getErr := r.Get(ctx, groupID, hostname)
			if getErr == model.ErrNotFound {
				return model.ErrNotFound
			}
			if existing != nil && existing.Rev != snapshotRev {
				return model.ErrRevConflict
			}
			// If we can't determine, return the original error
			return fmt.Errorf("conditional check failed: %w", err)
		}
		return fmt.Errorf("failed to delete domain record: %w", err)
	}

	return nil
}
