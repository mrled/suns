package memrepo

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/mrled/suns/symval/internal/model"
	"github.com/mrled/suns/symval/internal/symgroup"
)

func ExampleMemoryRepository() {
	tmpFile, _ := os.CreateTemp("", "example-*.json")
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	ctx := context.Background()
	repo, _ := NewMemoryRepositoryWithPersistence(tmpPath)

	data := &model.DomainRecord{
		Owner:        "alice@example.com",
		Type:         symgroup.Palindrome,
		Hostname:     "example.com",
		GroupID:      "abc123",
		ValidateTime: time.Date(2025, 10, 17, 12, 0, 0, 0, time.UTC),
	}

	repo.UnconditionalStore(ctx, data)

	// Read the JSON file to show format
	content, _ := os.ReadFile(tmpPath)
	fmt.Println(string(content))

	// Output:
	// [
	//   {
	//     "Owner": "alice@example.com",
	//     "Type": "a",
	//     "Hostname": "example.com",
	//     "GroupID": "abc123",
	//     "ValidateTime": "2025-10-17T12:00:00Z",
	//     "Rev": 1
	//   }
	// ]
}
