package store

import (
	"sync"
	"time"
)

// WebhookEvent represents an intercepted webhook
type WebhookEvent struct {
	ID        string            `json:"id"`
	Method    string            `json:"method"`
	Path      string            `json:"path"`
	Headers   map[string]string `json:"headers"`
	Body      []byte            `json:"body"`     // Stored raw to allow retry
	Payload   interface{}       `json:"payload"`  // Parsed payload for UI display
	Verified  bool              `json:"verified"` // Signature validation status
	Status    string            `json:"status"`   // e.g. "Received", "Forwarded", "Failed"
	Timestamp time.Time         `json:"timestamp"`
}

// MemoryStore holds our captured webhooks
type MemoryStore struct {
	mu     sync.RWMutex
	events []WebhookEvent
}

// NewMemoryStore initializes a new event store
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		events: make([]WebhookEvent, 0),
	}
}

// Add appends a new webhook event to the store
func (s *MemoryStore) Add(event WebhookEvent) {
	s.mu.Lock()
	defer s.mu.Unlock()
	// Insert at the beginning so newest is first in UI
	s.events = append([]WebhookEvent{event}, s.events...)
}

// GetAll returns all captured webhooks
func (s *MemoryStore) GetAll() []WebhookEvent {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.events
}

// GetByID returns a specific webhook event
func (s *MemoryStore) GetByID(id string) (*WebhookEvent, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, event := range s.events {
		if event.ID == id {
			return &event, true
		}
	}
	return nil, false
}
