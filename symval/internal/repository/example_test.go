package repository

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/callista/symval/internal/model"
)

func ExampleMemoryRepository() {
	tmpFile, _ := os.CreateTemp("", "example-*.json")
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	ctx := context.Background()
	repo, _ := NewMemoryRepositoryWithPersistence(tmpPath)

	flip := "flip.example.com"
	data := &model.DomainData{
		ValidateTime: time.Date(2025, 10, 17, 12, 0, 0, 0, time.UTC),
		Owner:        "alice@example.com",
		Domain:       "example.com",
		Flip:         &flip,
		Type:         model.Palindrome,
	}

	repo.Store(ctx, data)

	// Read the JSON file to show format
	content, _ := os.ReadFile(tmpPath)
	fmt.Println(string(content))

	// Output:
	// [
	//   {
	//     "ValidateTime": "2025-10-17T12:00:00Z",
	//     "Owner": "alice@example.com",
	//     "Domain": "example.com",
	//     "Flip": "flip.example.com",
	//     "Type": "palindrome"
	//   }
	// ]
}
