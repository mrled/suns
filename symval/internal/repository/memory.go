package repository

import (
	"context"
	"errors"
	"sync"

	"github.com/callista/symval/internal/model"
)

// MemoryRepository is an in-memory implementation of DomainRepository
type MemoryRepository struct {
	mu   sync.RWMutex
	data map[string]*model.DomainData
}

// NewMemoryRepository creates a new in-memory repository
func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		data: make(map[string]*model.DomainData),
	}
}

// Store saves domain data
func (r *MemoryRepository) Store(ctx context.Context, data *model.DomainData) error {
	if data == nil {
		return errors.New("domain data cannot be nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.data[data.Domain]; exists {
		return ErrAlreadyExists
	}

	r.data[data.Domain] = data
	return nil
}

// Get retrieves domain data by domain name
func (r *MemoryRepository) Get(ctx context.Context, domain string) (*model.DomainData, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	data, exists := r.data[domain]
	if !exists {
		return nil, ErrNotFound
	}

	return data, nil
}

// List retrieves all domain data
func (r *MemoryRepository) List(ctx context.Context) ([]*model.DomainData, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*model.DomainData, 0, len(r.data))
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
	return nil
}
