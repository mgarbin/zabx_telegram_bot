// Package store provides a thread-safe mapping from a Zabbix trigger ID to
// the Telegram message ID that was sent for that trigger. This allows the bot
// to edit an existing message when a trigger's status changes (e.g.
// PROBLEM → RESOLVED) instead of posting a new one.
//
// Two implementations are available:
//   - MessageStore: in-memory store (default)
//   - RedisStore:   Redis-backed store (enabled when a Redis address is configured)
package store

import "sync"

// Store is the interface implemented by both the in-memory MessageStore and
// the Redis-backed RedisStore.
type Store interface {
	// Set stores an Entry for the given event ID.
	Set(eventID string, entry Entry)
	// Get returns the Entry for the given event ID and a boolean indicating
	// whether the entry exists. Returns false when the key is missing or on
	// any backend error (which is logged internally).
	Get(eventID string) (Entry, bool)
	// Delete removes the entry for the given event ID.
	Delete(eventID string)
}

// Entry holds the data persisted for a single PROBLEM event.
type Entry struct {
	MessageID int
	StartTime string
	Message   string
	Severity  string
}

// MessageStore maps event IDs to Entry values.
type MessageStore struct {
	mu   sync.RWMutex
	data map[string]Entry
}

// New creates and returns an empty MessageStore.
func New() *MessageStore {
	return &MessageStore{data: make(map[string]Entry)}
}

// Set stores an Entry for the given event ID.
func (s *MessageStore) Set(eventID string, entry Entry) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[eventID] = entry
}

// Get returns the Entry for the given event ID, and a boolean indicating
// whether the entry exists.
func (s *MessageStore) Get(eventID string) (Entry, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e, ok := s.data[eventID]
	return e, ok
}

// Delete removes the entry for the given event ID.
func (s *MessageStore) Delete(eventID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, eventID)
}
