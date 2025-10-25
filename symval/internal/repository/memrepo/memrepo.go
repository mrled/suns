package memrepo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/mrled/suns/symval/internal/model"
)

// MemoryRepository is an in-memory implementation of DomainRepository optionally backed by a JSON file
type MemoryRepository struct {
	mu       sync.RWMutex
	data     map[string]*model.DomainRecord
	filePath string
}

// makeKey creates a composite key from groupID and hostname
// This matches the DynamoDB schema where PK=groupID and SK=hostname
func makeKey(groupID, hostname string) string {
	return groupID + "#" + hostname
}

// NewMemoryRepository creates a new in-memory repository without persistence.
// Data is stored only in memory and will be lost when the process terminates.
func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		data:     make(map[string]*model.DomainRecord),
		filePath: "",
	}
}

// NewMemoryRepositoryWithPersistence creates a new in-memory repository backed by a JSON file.
// The repository will load existing data from the file on initialization and persist
// all changes (Store, Delete) to the file automatically.
func NewMemoryRepositoryWithPersistence(filePath string) (*MemoryRepository, error) {
	repo := &MemoryRepository{
		data:     make(map[string]*model.DomainRecord),
		filePath: filePath,
	}

	// Create parent directory if it doesn't exist
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	// Try to load existing data from file
	if err := repo.load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	return repo, nil
}

// NewMemoryRepositoryFromJsonString creates a new in-memory repository initialized with data from a JSON string.
// The repository will not be backed by a file and will not persist changes.
// The JSON string should contain an array of DomainRecord objects.
func NewMemoryRepositoryFromJsonString(jsonString string) (*MemoryRepository, error) {
	repo := &MemoryRepository{
		data:     make(map[string]*model.DomainRecord),
		filePath: "",
	}

	// Parse JSON from the string
	if err := repo.loadFromReader(strings.NewReader(jsonString)); err != nil {
		return nil, err
	}

	return repo, nil
}

// loadFromReader reads JSON data from a reader and populates the in-memory data
func (r *MemoryRepository) loadFromReader(reader io.Reader) error {
	var dataSlice []*model.DomainRecord
	if err := json.NewDecoder(reader).Decode(&dataSlice); err != nil {
		return err
	}

	r.data = make(map[string]*model.DomainRecord)
	for _, d := range dataSlice {
		key := makeKey(d.GroupID, d.Hostname)

		// Print a warning if the key already exists.
		// This will not be possible in Dynamo, where a PUT with the same PK and SK will overwrite the existing item.
		if _, exists := r.data[key]; exists {
			fmt.Fprintf(os.Stderr, "Warning: duplicate entry found for GroupID=%s, Hostname=%s (keeping last occurrence)\n", d.GroupID, d.Hostname)
		}

		r.data[key] = d
	}

	return nil
}

// load reads the JSON file and populates the in-memory data
func (r *MemoryRepository) load() error {
	file, err := os.Open(r.filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Check if file is empty
	stat, err := file.Stat()
	if err != nil {
		return err
	}
	if stat.Size() == 0 {
		return nil
	}

	return r.loadFromReader(file)
}

// save writes the in-memory data to the JSON file
// If filePath is empty, this is a no-op
func (r *MemoryRepository) save() error {
	// Skip persistence if no file path is configured
	if r.filePath == "" {
		return nil
	}

	dataSlice := make([]*model.DomainRecord, 0, len(r.data))
	for _, d := range r.data {
		dataSlice = append(dataSlice, d)
	}

	file, err := os.Create(r.filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(dataSlice)
}

// Store saves domain data
func (r *MemoryRepository) Store(ctx context.Context, data *model.DomainRecord) error {
	if data == nil {
		return errors.New("domain data cannot be nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	key := makeKey(data.GroupID, data.Hostname)
	if _, exists := r.data[key]; exists {
		return model.ErrAlreadyExists
	}

	r.data[key] = data
	return r.save()
}

// Get retrieves domain data by group ID and domain name
func (r *MemoryRepository) Get(ctx context.Context, groupID, domain string) (*model.DomainRecord, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	key := makeKey(groupID, domain)
	data, exists := r.data[key]
	if !exists {
		return nil, model.ErrNotFound
	}

	return data, nil
}

// List retrieves all domain data
func (r *MemoryRepository) List(ctx context.Context) ([]*model.DomainRecord, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*model.DomainRecord, 0, len(r.data))
	for _, data := range r.data {
		result = append(result, data)
	}

	return result, nil
}

// Delete removes domain data by group ID and domain name
func (r *MemoryRepository) Delete(ctx context.Context, groupID, domain string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := makeKey(groupID, domain)
	if _, exists := r.data[key]; !exists {
		return model.ErrNotFound
	}

	delete(r.data, key)
	return r.save()
}
