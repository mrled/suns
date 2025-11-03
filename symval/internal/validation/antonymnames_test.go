package validation

import (
	"testing"

	"github.com/mrled/suns/symval/internal/model"
	"github.com/mrled/suns/symval/internal/symgroup"
)

func TestValidateAntonymNames(t *testing.T) {
	records := []*model.DomainRecord{
		{
			Owner:    "example.com",
			Type:     symgroup.AntonymNames,
			Hostname: "example.website",
			GroupID:  "test-group-id",
		},
		{
			Owner:    "example.com",
			Type:     symgroup.AntonymNames,
			Hostname: "example.email",
			GroupID:  "test-group-id",
		},
	}
	valid, err := validateAntonymNames(records)
	// Stub returns false, nil
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if valid {
		t.Errorf("Expected valid=false, got true")
	}
}
