package s3materializedview

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/mrled/suns/symval/internal/repository/memrepo"
)

// S3MaterializedView handles loading and saving repository data to S3
type S3MaterializedView struct {
	s3Client     *s3.Client
	bucketName   string
	key          string
	contentType  string
	cacheControl string
}

// New creates a new S3MaterializedView adapter
func New(s3Client *s3.Client, bucketName, key string) *S3MaterializedView {
	return &S3MaterializedView{
		s3Client:     s3Client,
		bucketName:   bucketName,
		key:          key,
		contentType:  "application/json",
		cacheControl: "max-age=60", // Cache for 1 minute
	}
}

// Load loads data from S3 into a new MemoryRepository
func (s *S3MaterializedView) Load(ctx context.Context) (*memrepo.MemoryRepository, error) {
	result, err := s.s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &s.bucketName,
		Key:    &s.key,
	})

	if err != nil {
		// Check if the error is because the file doesn't exist
		// In that case, we return an empty repository
		return nil, fmt.Errorf("failed to get object from S3: %w", err)
	}
	defer result.Body.Close()

	// Read the body
	bodyBytes, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read S3 object body: %w", err)
	}

	// Create a new MemoryRepository from the JSON string
	repo, err := memrepo.NewMemoryRepositoryFromJsonString(string(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create repository from JSON: %w", err)
	}

	return repo, nil
}

// Save saves the repository data to S3
func (s *S3MaterializedView) Save(ctx context.Context, repo *memrepo.MemoryRepository) error {
	// Get all records from the repository
	records, err := repo.List(ctx)
	if err != nil {
		return fmt.Errorf("failed to list records from repository: %w", err)
	}

	// Marshal to JSON in memrepo format (array of DomainRecord)
	jsonData, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal records: %w", err)
	}

	// Upload to S3 with appropriate headers for public access
	_, err = s.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:       &s.bucketName,
		Key:          &s.key,
		Body:         bytes.NewReader(jsonData),
		ContentType:  stringPtr(s.contentType),
		CacheControl: stringPtr(s.cacheControl),
	})

	if err != nil {
		return fmt.Errorf("failed to upload to S3: %w", err)
	}

	slog.Info("Successfully updated S3 data file",
		slog.String("bucket", s.bucketName),
		slog.String("key", s.key),
		slog.Int("record_count", len(records)))
	return nil
}

// stringPtr returns a pointer to a string
func stringPtr(s string) *string {
	return &s
}
