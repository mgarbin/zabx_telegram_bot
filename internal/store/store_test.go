package store_test

import (
	"sync"
	"testing"

	"github.com/mgarbin/zabbix-telegram-event-correlator/internal/store"
)

func TestSetAndGet(t *testing.T) {
	s := store.New()

	s.Set("trigger-1", store.Entry{MessageID: 42, StartTime: "2024-01-01 00:00:00 UTC", Message: "details"})
	e, ok := s.Get("trigger-1")
	if !ok {
		t.Fatal("expected entry to exist after Set")
	}
	if e.MessageID != 42 {
		t.Fatalf("expected message ID 42, got %d", e.MessageID)
	}
	if e.StartTime != "2024-01-01 00:00:00 UTC" {
		t.Fatalf("expected StartTime to be preserved, got %q", e.StartTime)
	}
	if e.Message != "details" {
		t.Fatalf("expected Message to be preserved, got %q", e.Message)
	}
}

func TestGetMissing(t *testing.T) {
	s := store.New()

	_, ok := s.Get("nonexistent")
	if ok {
		t.Fatal("expected Get to return false for a missing key")
	}
}

func TestDelete(t *testing.T) {
	s := store.New()

	s.Set("trigger-1", store.Entry{MessageID: 99})
	s.Delete("trigger-1")

	_, ok := s.Get("trigger-1")
	if ok {
		t.Fatal("expected entry to be absent after Delete")
	}
}

func TestDeleteMissing(t *testing.T) {
	s := store.New()
	// Should not panic.
	s.Delete("does-not-exist")
}

func TestConcurrentAccess(t *testing.T) {
	s := store.New()
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			key := "trigger"
			s.Set(key, store.Entry{MessageID: n})
			s.Get(key)
			s.Delete(key)
		}(i)
	}
	wg.Wait()
}
