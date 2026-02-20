// Package store provides a thread-safe in-memory mapping from a Zabbix
// trigger ID to the Telegram message ID that was sent for that trigger.
// This allows the bot to edit an existing message when a trigger's status
// changes (e.g. PROBLEM → RESOLVED) instead of posting a new one.
package store

import "sync"

// MessageStore maps trigger IDs to Telegram message IDs.
type MessageStore struct {
	mu   sync.RWMutex
	data map[string]int
}

// New creates and returns an empty MessageStore.
func New() *MessageStore {
	return &MessageStore{data: make(map[string]int)}
}

// Set stores the Telegram message ID for the given trigger ID.
func (s *MessageStore) Set(triggerID string, messageID int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[triggerID] = messageID
}

// Get returns the Telegram message ID for the given trigger ID, and a boolean
// indicating whether the entry exists.
func (s *MessageStore) Get(triggerID string) (int, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	id, ok := s.data[triggerID]
	return id, ok
}

// Delete removes the entry for the given trigger ID.
func (s *MessageStore) Delete(triggerID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, triggerID)
}
