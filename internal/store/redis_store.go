package store

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

const redisOpTimeout = 5 * time.Second

// RedisStore is a Store implementation backed by a Redis-compatible server.
// Entries are serialised as JSON and stored with no expiry by default.
type RedisStore struct {
	client *redis.Client
}

// NewRedisStore creates a RedisStore connected to the given Redis server.
// addr is the host:port of the server (e.g. "localhost:6379").
// password may be empty when authentication is not required.
// db selects the logical Redis database index (usually 0).
func NewRedisStore(addr, password string, db int) *RedisStore {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
	return &RedisStore{client: client}
}

// Ping checks connectivity to the Redis server and returns an error if the
// server is unreachable. Callers may use this at startup to fail fast.
func (r *RedisStore) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), redisOpTimeout)
	defer cancel()
	return r.client.Ping(ctx).Err()
}

// Set serialises entry as JSON and stores it under the given event ID.
func (r *RedisStore) Set(eventID string, entry Entry) {
	data, err := json.Marshal(entry)
	if err != nil {
		log.Printf("ERROR redis store: marshal entry for event %s: %v", eventID, err)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), redisOpTimeout)
	defer cancel()
	if err := r.client.Set(ctx, eventID, data, 0).Err(); err != nil {
		log.Printf("ERROR redis store: SET event %s: %v", eventID, err)
	}
}

// Get retrieves and deserialises the Entry for the given event ID.
// Returns (Entry{}, false) when the key does not exist or on any error.
func (r *RedisStore) Get(eventID string) (Entry, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), redisOpTimeout)
	defer cancel()
	data, err := r.client.Get(ctx, eventID).Bytes()
	if err != nil {
		if !errors.Is(err, redis.Nil) {
			log.Printf("ERROR redis store: GET event %s: %v", eventID, err)
		}
		return Entry{}, false
	}
	var entry Entry
	if err := json.Unmarshal(data, &entry); err != nil {
		log.Printf("ERROR redis store: unmarshal entry for event %s: %v", eventID, err)
		return Entry{}, false
	}
	return entry, true
}

// Delete removes the entry for the given event ID.
func (r *RedisStore) Delete(eventID string) {
	ctx, cancel := context.WithTimeout(context.Background(), redisOpTimeout)
	defer cancel()
	if err := r.client.Del(ctx, eventID).Err(); err != nil {
		log.Printf("ERROR redis store: DEL event %s: %v", eventID, err)
	}
}
