package repository

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"

	"github.com/mrled/suns/symval/internal/model"
)

// MemoryRepository is an in-memory implementation of DomainRepository optionally backed by a JSON file
type MemoryRepository struct {
	mu       sync.RWMutex
	data     map[string]*model.DomainRecord
	filePath string
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

	var dataSlice []*model.DomainRecord
	if err := json.NewDecoder(file).Decode(&dataSlice); err != nil {
		return err
	}

	r.data = make(map[string]*model.DomainRecord)
	for _, d := range dataSlice {
		r.data[d.Hostname] = d
	}

	return nil
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

	if _, exists := r.data[data.Hostname]; exists {
		return ErrAlreadyExists
	}

	r.data[data.Hostname] = data
	return r.save()
}

// Get retrieves domain data by domain name
func (r *MemoryRepository) Get(ctx context.Context, domain string) (*model.DomainRecord, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	data, exists := r.data[domain]
	if !exists {
		return nil, ErrNotFound
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

// Delete removes domain data by domain name
func (r *MemoryRepository) Delete(ctx context.Context, domain string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.data[domain]; !exists {
		return ErrNotFound
	}

	delete(r.data, domain)
	return r.save()
}
