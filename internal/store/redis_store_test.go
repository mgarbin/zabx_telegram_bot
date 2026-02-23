package store_test

import (
	"sync"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/mgarbin/zabbix-telegram-event-correlator/internal/store"
)

// startMiniRedis starts a miniredis server and returns its address.
// The server is automatically closed when the test ends.
func startMiniRedis(t *testing.T) string {
	t.Helper()
	s, err := miniredis.Run()
	if err != nil {
		t.Fatalf("starting miniredis: %v", err)
	}
	t.Cleanup(s.Close)
	return s.Addr()
}

func TestRedisSetAndGet(t *testing.T) {
	addr := startMiniRedis(t)
	s := store.NewRedisStore(addr, "", 0)

	s.Set("trigger-1", store.Entry{MessageID: 42, StartTime: "2024-01-01 00:00:00 UTC", Message: "details", Severity: "HIGH"})
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
	if e.Severity != "HIGH" {
		t.Fatalf("expected Severity 'HIGH', got %q", e.Severity)
	}
}

func TestRedisGetMissing(t *testing.T) {
	addr := startMiniRedis(t)
	s := store.NewRedisStore(addr, "", 0)

	_, ok := s.Get("nonexistent")
	if ok {
		t.Fatal("expected Get to return false for a missing key")
	}
}

func TestRedisDelete(t *testing.T) {
	addr := startMiniRedis(t)
	s := store.NewRedisStore(addr, "", 0)

	s.Set("trigger-1", store.Entry{MessageID: 99})
	s.Delete("trigger-1")

	_, ok := s.Get("trigger-1")
	if ok {
		t.Fatal("expected entry to be absent after Delete")
	}
}

func TestRedisDeleteMissing(t *testing.T) {
	addr := startMiniRedis(t)
	s := store.NewRedisStore(addr, "", 0)
	// Should not panic.
	s.Delete("does-not-exist")
}

func TestRedisConcurrentAccess(t *testing.T) {
	addr := startMiniRedis(t)
	s := store.NewRedisStore(addr, "", 0)
	var wg sync.WaitGroup

	for i := 0; i < 50; i++ {
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

// TestRedisStoreImplementsStore verifies at compile time that *RedisStore
// satisfies the Store interface.
func TestRedisStoreImplementsStore(t *testing.T) {
	var _ store.Store = (*store.RedisStore)(nil)
}

// TestMessageStoreImplementsStore verifies at compile time that *MessageStore
// satisfies the Store interface.
func TestMessageStoreImplementsStore(t *testing.T) {
	var _ store.Store = (*store.MessageStore)(nil)
}
