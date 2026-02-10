package database

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisClient wraps go-redis with helpers for sessions, cache, and rate limiting.
type RedisClient struct {
	*redis.Client
}

// NewRedisClient creates a Redis client with optional connection retry.
func NewRedisClient(redisURL string) (*RedisClient, error) {
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("redis url: %w", err)
	}
	client := redis.NewClient(opt)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// Retry with backoff
	for i := 0; i < 5; i++ {
		if err := client.Ping(ctx).Err(); err == nil {
			return &RedisClient{Client: client}, nil
		}
		time.Sleep(time.Duration(i+1) * time.Second)
	}
	// Graceful: return client anyway so app can start (e.g. cache optional)
	if err := client.Ping(ctx).Err(); err != nil {
		log.Printf("warning: redis unavailable: %v; continuing without cache", err)
	}
	return &RedisClient{Client: client}, nil
}

// SetWithExpiry sets a key with TTL.
func (r *RedisClient) SetWithExpiry(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return r.Client.Set(ctx, key, value, ttl).Err()
}

// Get returns the string value.
func (r *RedisClient) Get(ctx context.Context, key string) (string, error) {
	return r.Client.Get(ctx, key).Result()
}

// GetJSON gets and unmarshals JSON into v.
func (r *RedisClient) GetJSON(ctx context.Context, key string, v interface{}) error {
	s, err := r.Client.Get(ctx, key).Result()
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(s), v)
}

// SetJSON marshals v and sets with optional TTL (0 = no expiry).
func (r *RedisClient) SetJSON(ctx context.Context, key string, v interface{}, ttl time.Duration) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return r.Client.Set(ctx, key, b, ttl).Err()
}

// Delete removes a key.
func (r *RedisClient) Delete(ctx context.Context, keys ...string) error {
	return r.Client.Del(ctx, keys...).Err()
}

// Exists returns whether the key exists.
func (r *RedisClient) Exists(ctx context.Context, key string) (bool, error) {
	n, err := r.Client.Exists(ctx, key).Result()
	return n > 0, err
}

// DeleteByPattern deletes keys matching pattern (e.g. "notebook:*"). Use with care.
func (r *RedisClient) DeleteByPattern(ctx context.Context, pattern string) error {
	iter := r.Client.Scan(ctx, 0, pattern, 100).Iterator()
	var keys []string
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if err := iter.Err(); err != nil {
		return err
	}
	if len(keys) > 0 {
		return r.Client.Del(ctx, keys...).Err()
	}
	return nil
}

// Increment increments a key and returns the new value.
func (r *RedisClient) Increment(ctx context.Context, key string) (int64, error) {
	return r.Client.Incr(ctx, key).Result()
}

// Expire sets TTL on a key.
func (r *RedisClient) Expire(ctx context.Context, key string, ttl time.Duration) error {
	return r.Client.Expire(ctx, key, ttl).Err()
}
