package validation

import (
	"context"
	"testing"

	"github.com/callista/symval/internal/model"
)

func TestService_Validate_Success(t *testing.T) {
	service := NewService()
	ctx := context.Background()

	// Valid group: all fields match and groupID is correct
	data := []*model.DomainData{
		{
			Owner:    "alice@example.com",
			Type:     model.Palindrome,
			Hostname: "example.com",
			GroupID:  "v1:palindrome:ddhIziTf/kTYyc/vnrux+C84XVmM3twmGEJ5wPrUA4c=",
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

	// Valid group with multiple hostnames
	data := []*model.DomainData{
		{
			Owner:    "alice@example.com",
			Type:     model.Palindrome,
			Hostname: "example.com",
			GroupID:  "v1:palindrome:z6RsiCWP6vkX8TbKrzwt8sTVAObs79zVOoj9ibZaGyU=",
		},
		{
			Owner:    "alice@example.com",
			Type:     model.Palindrome,
			Hostname: "flip.example.com",
			GroupID:  "v1:palindrome:z6RsiCWP6vkX8TbKrzwt8sTVAObs79zVOoj9ibZaGyU=",
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
			Type:     model.Palindrome,
			Hostname: "example.com",
			GroupID:  "v1:palindrome:z6RsiCWP6vkX8TbKrzwt8sTVAObs79zVOoj9ibZaGyU=",
		},
		{
			Owner:    "bob@example.com", // Different owner
			Type:     model.Palindrome,
			Hostname: "flip.example.com",
			GroupID:  "v1:palindrome:z6RsiCWP6vkX8TbKrzwt8sTVAObs79zVOoj9ibZaGyU=",
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
			Type:     model.Palindrome,
			Hostname: "example.com",
			GroupID:  "v1:palindrome:z6RsiCWP6vkX8TbKrzwt8sTVAObs79zVOoj9ibZaGyU=",
		},
		{
			Owner:    "alice@example.com",
			Type:     model.Flip180, // Different type
			Hostname: "flip.example.com",
			GroupID:  "v1:palindrome:z6RsiCWP6vkX8TbKrzwt8sTVAObs79zVOoj9ibZaGyU=",
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
			Type:     model.Palindrome,
			Hostname: "example.com",
			GroupID:  "v1:palindrome:z6RsiCWP6vkX8TbKrzwt8sTVAObs79zVOoj9ibZaGyU=",
		},
		{
			Owner:    "alice@example.com",
			Type:     model.Palindrome,
			Hostname: "flip.example.com",
			GroupID:  "v1:palindrome:different-group-id", // Different groupID
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
			Type:     model.Palindrome,
			Hostname: "example.com",
			GroupID:  "v1:palindrome:wrong-hash-value",
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
