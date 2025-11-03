package validation

import (
	"testing"

	"github.com/mrled/suns/symval/internal/model"
	"github.com/mrled/suns/symval/internal/symgroup"
)

func TestValidateMirrorText(t *testing.T) {
	records := []*model.DomainRecord{
		{
			Owner:    "duq.xodbox.pub",
			Type:     symgroup.MirrorText,
			Hostname: "example.website",
			GroupID:  "test-group-id",
		},
	}
	valid, err := validateMirrorText(records)
	// Stub returns false, nil
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if valid {
		t.Errorf("Expected valid=false, got true")
	}
}
