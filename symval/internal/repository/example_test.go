package repository

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/mrled/suns/symval/internal/model"
)

func ExampleMemoryRepository() {
	tmpFile, _ := os.CreateTemp("", "example-*.json")
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	ctx := context.Background()
	repo, _ := NewMemoryRepositoryWithPersistence(tmpPath)

	data := &model.DomainData{
		Owner:        "alice@example.com",
		Type:         model.Palindrome,
		Hostname:     "example.com",
		GroupID:      "abc123",
		ValidateTime: time.Date(2025, 10, 17, 12, 0, 0, 0, time.UTC),
	}

	repo.Store(ctx, data)

	// Read the JSON file to show format
	content, _ := os.ReadFile(tmpPath)
	fmt.Println(string(content))

	// Output:
	// [
	//   {
	//     "Owner": "alice@example.com",
	//     "Type": "palindrome",
	//     "Hostname": "example.com",
	//     "GroupID": "abc123",
	//     "ValidateTime": "2025-10-17T12:00:00Z"
	//   }
	// ]
}
