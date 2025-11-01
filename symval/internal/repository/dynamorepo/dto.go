package dynamorepo

import (
	"time"

	"github.com/mrled/suns/symval/internal/model"
	"github.com/mrled/suns/symval/internal/symgroup"
)

// DynamoDTO represents the persistence layer DTO for DynamoDB
// It maps the domain model to DynamoDB's key structure where:
// - PK (partition key) is the GroupID
// - SK (sort key) is the Hostname
type DynamoDTO struct {
	PK           string                `dynamodbav:"pk"` // Partition Key - maps from GroupID
	SK           string                `dynamodbav:"sk"` // Sort Key - maps from Hostname
	Owner        string                `dynamodbav:"Owner"`
	Type         symgroup.SymmetryType `dynamodbav:"Type"`
	ValidateTime time.Time             `dynamodbav:"ValidateTime"`
	Rev          int64                 `dynamodbav:"Rev"` // Monotonically increasing revision number
}

// ToDomain converts a DynamoDTO to a domain model DomainRecord
func (dto *DynamoDTO) ToDomain() *model.DomainRecord {
	return &model.DomainRecord{
		Owner:        dto.Owner,
		Type:         dto.Type,
		Hostname:     dto.SK,
		GroupID:      dto.PK,
		ValidateTime: dto.ValidateTime,
		Rev:          dto.Rev,
	}
}

// FromDomain creates a DynamoDTO from a domain model DomainRecord
func FromDomain(record *model.DomainRecord) *DynamoDTO {
	return &DynamoDTO{
		PK:           record.GroupID,
		SK:           record.Hostname,
		Owner:        record.Owner,
		Type:         record.Type,
		ValidateTime: record.ValidateTime,
		Rev:          record.Rev,
	}
}

// ToDomainList converts a slice of DynamoDTOs to domain model DomainRecords
func ToDomainList(dtos []*DynamoDTO) []*model.DomainRecord {
	records := make([]*model.DomainRecord, len(dtos))
	for i, dto := range dtos {
		records[i] = dto.ToDomain()
	}
	return records
}

// FromDomainList creates a slice of DynamoDTOs from domain model DomainRecords
func FromDomainList(records []*model.DomainRecord) []*DynamoDTO {
	dtos := make([]*DynamoDTO, len(records))
	for i, record := range records {
		dtos[i] = FromDomain(record)
	}
	return dtos
}
