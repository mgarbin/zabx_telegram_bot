package store_test

import (
	"sync"
	"testing"

	"github.com/mgarbin/zabx_telegram_bot/internal/store"
)

func TestSetAndGet(t *testing.T) {
	s := store.New()

	s.Set("trigger-1", 42)
	id, ok := s.Get("trigger-1")
	if !ok {
		t.Fatal("expected entry to exist after Set")
	}
	if id != 42 {
		t.Fatalf("expected message ID 42, got %d", id)
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

	s.Set("trigger-1", 99)
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
			s.Set(key, n)
			s.Get(key)
			s.Delete(key)
		}(i)
	}
	wg.Wait()
}
