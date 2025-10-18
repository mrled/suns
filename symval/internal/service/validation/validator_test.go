package validation

import (
	"context"
	"testing"

	"github.com/mrled/suns/symval/internal/model"
	"github.com/mrled/suns/symval/internal/symgroup"
)

func TestService_Validate_Success(t *testing.T) {
	service := NewService()
	ctx := context.Background()

	// Valid group: all fields match and groupID is correct
	// Using "aba" which is a palindrome
	data := []*model.DomainData{
		{
			Owner:    "alice@example.com",
			Type:     symgroup.Palindrome,
			Hostname: "aba",
			GroupID:  "v1:a:/42YGfwOEr8NJIkuRZh+JJoo3Og2qFytYOKOqqjG2XY=:4SStzOH7L4jh6nmcPQgghF7TQ+bHOeVBMfyzpW5Lwb0=",
		},
	}

	valid, err := service.Validate(ctx, data)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if !valid {
		t.Errorf("Expected valid=true, got false")
	}
}

func TestService_Validate_MultipleHostnames(t *testing.T) {
	service := NewService()
	ctx := context.Background()

	// Test with MirrorNames type which requires exactly two hostnames that are mirror pairs
	data := []*model.DomainData{
		{
			Owner:    "alice@example.com",
			Type:     symgroup.MirrorNames,
			Hostname: "a.b.com",
			GroupID:  "v1:e:/42YGfwOEr8NJIkuRZh+JJoo3Og2qFytYOKOqqjG2XY=:SGjit3PbOdrHhyHGQZzNBkgwB2bYLJ1ZDNqkPJW728c=",
		},
		{
			Owner:    "alice@example.com",
			Type:     symgroup.MirrorNames,
			Hostname: "com.b.a",
			GroupID:  "v1:e:/42YGfwOEr8NJIkuRZh+JJoo3Og2qFytYOKOqqjG2XY=:SGjit3PbOdrHhyHGQZzNBkgwB2bYLJ1ZDNqkPJW728c=",
		},
	}

	valid, err := service.Validate(ctx, data)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if !valid {
		t.Errorf("Expected valid=true, got false")
	}
}

func TestService_Validate_EmptyList(t *testing.T) {
	service := NewService()
	ctx := context.Background()

	data := []*model.DomainData{}

	valid, err := service.Validate(ctx, data)
	if err == nil {
		t.Error("Expected error for empty list, got nil")
	}
	if valid {
		t.Error("Expected valid=false for empty list")
	}
}

func TestService_Validate_OwnerMismatch(t *testing.T) {
	service := NewService()
	ctx := context.Background()

	data := []*model.DomainData{
		{
			Owner:    "alice@example.com",
			Type:     symgroup.Palindrome,
			Hostname: "example.com",
			GroupID:  "v1:a:z6RsiCWP6vkX8TbKrzwt8sTVAObs79zVOoj9ibZaGyU=",
		},
		{
			Owner:    "bob@example.com", // Different owner
			Type:     symgroup.Palindrome,
			Hostname: "flip.example.com",
			GroupID:  "v1:a:z6RsiCWP6vkX8TbKrzwt8sTVAObs79zVOoj9ibZaGyU=",
		},
	}

	valid, err := service.Validate(ctx, data)
	if err == nil {
		t.Error("Expected error for owner mismatch, got nil")
	}
	if valid {
		t.Error("Expected valid=false for owner mismatch")
	}
}

func TestService_Validate_TypeMismatch(t *testing.T) {
	service := NewService()
	ctx := context.Background()

	data := []*model.DomainData{
		{
			Owner:    "alice@example.com",
			Type:     symgroup.Palindrome,
			Hostname: "example.com",
			GroupID:  "v1:a:z6RsiCWP6vkX8TbKrzwt8sTVAObs79zVOoj9ibZaGyU=",
		},
		{
			Owner:    "alice@example.com",
			Type:     symgroup.Flip180, // Different type
			Hostname: "flip.example.com",
			GroupID:  "v1:a:z6RsiCWP6vkX8TbKrzwt8sTVAObs79zVOoj9ibZaGyU=",
		},
	}

	valid, err := service.Validate(ctx, data)
	if err == nil {
		t.Error("Expected error for type mismatch, got nil")
	}
	if valid {
		t.Error("Expected valid=false for type mismatch")
	}
}

func TestService_Validate_GroupIDMismatch(t *testing.T) {
	service := NewService()
	ctx := context.Background()

	data := []*model.DomainData{
		{
			Owner:    "alice@example.com",
			Type:     symgroup.Palindrome,
			Hostname: "example.com",
			GroupID:  "v1:a:z6RsiCWP6vkX8TbKrzwt8sTVAObs79zVOoj9ibZaGyU=",
		},
		{
			Owner:    "alice@example.com",
			Type:     symgroup.Palindrome,
			Hostname: "flip.example.com",
			GroupID:  "v1:a:different-group-id", // Different groupID
		},
	}

	valid, err := service.Validate(ctx, data)
	if err == nil {
		t.Error("Expected error for groupID mismatch, got nil")
	}
	if valid {
		t.Error("Expected valid=false for groupID mismatch")
	}
}

func TestService_Validate_InvalidGroupID(t *testing.T) {
	service := NewService()
	ctx := context.Background()

	// GroupID doesn't match the calculated value
	data := []*model.DomainData{
		{
			Owner:    "alice@example.com",
			Type:     symgroup.Palindrome,
			Hostname: "example.com",
			GroupID:  "v1:a:wrong-hash-value",
		},
	}

	valid, err := service.Validate(ctx, data)
	if err == nil {
		t.Error("Expected error for invalid groupID, got nil")
	}
	if valid {
		t.Error("Expected valid=false for invalid groupID")
	}
}

func TestService_ValidateBase_Success(t *testing.T) {
	service := NewService()
	ctx := context.Background()

	data := []*model.DomainData{
		{
			Owner:    "alice@example.com",
			Type:     symgroup.Palindrome,
			Hostname: "aba",
			GroupID:  "v1:a:/42YGfwOEr8NJIkuRZh+JJoo3Og2qFytYOKOqqjG2XY=:4SStzOH7L4jh6nmcPQgghF7TQ+bHOeVBMfyzpW5Lwb0=",
		},
	}

	owner, groupID, symmetryType, err := service.ValidateBase(ctx, data)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if owner != "alice@example.com" {
		t.Errorf("Expected owner 'alice@example.com', got '%s'", owner)
	}
	if groupID != "v1:a:/42YGfwOEr8NJIkuRZh+JJoo3Og2qFytYOKOqqjG2XY=:4SStzOH7L4jh6nmcPQgghF7TQ+bHOeVBMfyzpW5Lwb0=" {
		t.Errorf("Expected specific groupID, got '%s'", groupID)
	}
	if symmetryType != symgroup.Palindrome {
		t.Errorf("Expected type 'a', got '%s'", symmetryType)
	}
}

func TestService_Validate_AllSymmetryTypes(t *testing.T) {
	service := NewService()
	ctx := context.Background()

	tests := []struct {
		name         string
		symmetryType symgroup.SymmetryType
		hostnames    []string
		groupID      string
	}{
		{"Palindrome", symgroup.Palindrome, []string{"aba"}, "v1:a:/42YGfwOEr8NJIkuRZh+JJoo3Og2qFytYOKOqqjG2XY=:4SStzOH7L4jh6nmcPQgghF7TQ+bHOeVBMfyzpW5Lwb0="},
		{"Flip180", symgroup.Flip180, []string{"example.com"}, "v1:b:/42YGfwOEr8NJIkuRZh+JJoo3Og2qFytYOKOqqjG2XY=:o3mm9u6vuaVeN4wRgDTidR5oL6ufLTCrE9ISVYbOGUc="},
		{"DoubleFlip180", symgroup.DoubleFlip180, []string{"example.com"}, "v1:c:/42YGfwOEr8NJIkuRZh+JJoo3Og2qFytYOKOqqjG2XY=:o3mm9u6vuaVeN4wRgDTidR5oL6ufLTCrE9ISVYbOGUc="},
		{"MirrorText", symgroup.MirrorText, []string{"example.com"}, "v1:d:/42YGfwOEr8NJIkuRZh+JJoo3Og2qFytYOKOqqjG2XY=:o3mm9u6vuaVeN4wRgDTidR5oL6ufLTCrE9ISVYbOGUc="},
		{"MirrorNames", symgroup.MirrorNames, []string{"a.b.com", "com.b.a"}, "v1:e:/42YGfwOEr8NJIkuRZh+JJoo3Og2qFytYOKOqqjG2XY=:SGjit3PbOdrHhyHGQZzNBkgwB2bYLJ1ZDNqkPJW728c="},
		{"AntonymNames", symgroup.AntonymNames, []string{"example.com"}, "v1:f:/42YGfwOEr8NJIkuRZh+JJoo3Og2qFytYOKOqqjG2XY=:o3mm9u6vuaVeN4wRgDTidR5oL6ufLTCrE9ISVYbOGUc="},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := make([]*model.DomainData, len(tt.hostnames))
			for i, hostname := range tt.hostnames {
				data[i] = &model.DomainData{
					Owner:    "alice@example.com",
					Type:     tt.symmetryType,
					Hostname: hostname,
					GroupID:  tt.groupID,
				}
			}

			valid, err := service.Validate(ctx, data)
			if err != nil {
				t.Errorf("Expected no error for %s, got: %v", tt.name, err)
			}
			if !valid {
				t.Errorf("Expected valid=true for %s, got false", tt.name)
			}
		})
	}
}

func TestService_Validate_UnknownSymmetryType(t *testing.T) {
	service := NewService()
	ctx := context.Background()

	// Use an unknown symmetry type
	data := []*model.DomainData{
		{
			Owner:    "alice@example.com",
			Type:     symgroup.SymmetryType("unknown"),
			Hostname: "example.com",
			GroupID:  "v1:unknown:ddhIziTf/kTYyc/vnrux+C84XVmM3twmGEJ5wPrUA4c=",
		},
	}

	valid, err := service.Validate(ctx, data)
	if err == nil {
		t.Error("Expected error for unknown symmetry type, got nil")
	}
	if valid {
		t.Error("Expected valid=false for unknown symmetry type")
	}
}
