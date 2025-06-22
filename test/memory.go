package test

import (
	"context"
	"sync"
	"time"

	"github.com/fabiohsgomes/go-expert-desafiostec-ratelimiter/pkg/storage"
)

// MemoryStorage implements Storage interface using in-memory maps
type MemoryStorage struct {
	mu       sync.RWMutex
	counts   map[string]countEntry
	blocks   map[string]time.Time
}

type countEntry struct {
	count      int64
	expiration time.Time
}

// NewMemoryStorage creates a new in-memory storage instance
func NewMemoryStorage() storage.Storage {
	return &MemoryStorage{
		counts: make(map[string]countEntry),
		blocks: make(map[string]time.Time),
	}
}

func (m *MemoryStorage) GetRequestCount(ctx context.Context, key string) (int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	entry, exists := m.counts[key]
	if !exists || time.Now().After(entry.expiration) {
		return 0, nil
	}
	return entry.count, nil
}

func (m *MemoryStorage) IncrementRequestCount(ctx context.Context, key string, expiration time.Duration) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	entry, exists := m.counts[key]

	if !exists || now.After(entry.expiration) {
		// Create new entry or reset expired entry
		m.counts[key] = countEntry{
			count:      1,
			expiration: now.Add(expiration),
		}
		return 1, nil
	}

	// Increment existing entry
	entry.count++
	m.counts[key] = entry
	return entry.count, nil
}

func (m *MemoryStorage) IsBlocked(ctx context.Context, key string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	blockTime, exists := m.blocks[key]
	if !exists {
		return false, nil
	}

	if time.Now().After(blockTime) {
		// Block has expired, clean it up
		m.mu.RUnlock()
		m.mu.Lock()
		delete(m.blocks, key)
		m.mu.Unlock()
		m.mu.RLock()
		return false, nil
	}

	return true, nil
}

func (m *MemoryStorage) Block(ctx context.Context, key string, duration time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.blocks[key] = time.Now().Add(duration)
	return nil
}

func (m *MemoryStorage) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.counts = make(map[string]countEntry)
	m.blocks = make(map[string]time.Time)
	return nil
}