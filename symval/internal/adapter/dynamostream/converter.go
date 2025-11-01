package dynamostream

import (
	"fmt"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/mrled/suns/symval/internal/model"
	"github.com/mrled/suns/symval/internal/symgroup"
)

// ConvertToDomainRecord converts a DynamoDB NewImage map to a DomainRecord
func ConvertToDomainRecord(newImage map[string]events.DynamoDBAttributeValue) (*model.DomainRecord, error) {
	if newImage == nil {
		return nil, fmt.Errorf("newImage is nil")
	}

	// Extract fields from the NewImage
	domainRecord := &model.DomainRecord{}

	// GroupID comes from pk
	if pk, ok := newImage["pk"]; ok && pk.DataType() == events.DataTypeString {
		domainRecord.GroupID = pk.String()
	}

	// Hostname comes from sk
	if sk, ok := newImage["sk"]; ok && sk.DataType() == events.DataTypeString {
		domainRecord.Hostname = sk.String()
	}

	// Owner - required
	if owner, ok := newImage["Owner"]; ok && owner.DataType() == events.DataTypeString {
		domainRecord.Owner = owner.String()
	} else {
		return nil, fmt.Errorf("missing required field: Owner")
	}

	// Type - required
	if typeField, ok := newImage["Type"]; ok && typeField.DataType() == events.DataTypeString {
		domainRecord.Type = symgroup.SymmetryType(typeField.String())
	} else {
		return nil, fmt.Errorf("missing required field: Type")
	}

	// ValidateTime - required and must be valid RFC3339
	if validateTime, ok := newImage["ValidateTime"]; ok && validateTime.DataType() == events.DataTypeString {
		t, err := time.Parse(time.RFC3339, validateTime.String())
		if err != nil {
			return nil, fmt.Errorf("invalid ValidateTime format: %w", err)
		}
		domainRecord.ValidateTime = t
	} else {
		return nil, fmt.Errorf("missing required field: ValidateTime")
	}

	// Validate we have the primary key fields
	if domainRecord.GroupID == "" {
		return nil, fmt.Errorf("missing required field: GroupID (pk)")
	}
	if domainRecord.Hostname == "" {
		return nil, fmt.Errorf("missing required field: Hostname (sk)")
	}

	return domainRecord, nil
}

// ExtractStringAttribute extracts a string value from DynamoDB attribute map
func ExtractStringAttribute(attrs map[string]events.DynamoDBAttributeValue, key string) string {
	if attr, ok := attrs[key]; ok {
		if attr.DataType() == events.DataTypeString {
			return attr.String()
		}
	}
	return ""
}
