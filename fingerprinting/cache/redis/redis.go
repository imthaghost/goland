package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/imthaghost/goland/fingerprinting/cache"
	"github.com/imthaghost/goland/fingerprinting/config"

	"github.com/redis/go-redis/v9"
)

// Redis struct represents a Redis cache
type Redis struct {
	Cache *redis.Client
}

const (
	DefaultExpiration = 24 * time.Hour
)

func New(config config.Config) *Redis {
	address := config.RedisConfig.Host + ":" + config.RedisConfig.Port
	client := redis.NewClient(&redis.Options{
		Addr:     address,
		Password: config.RedisConfig.Password,
		DB:       0,
	})
	return &Redis{Cache: client}
}

// Set stores serialized items in the cache.
func (r *Redis) Set(ctx context.Context, key string, item interface{}, expiration time.Duration) (string, error) {
	if expiration == 0 {
		expiration = DefaultExpiration
	}

	itemJSON, err := json.Marshal(item)
	if err != nil {
		return "", fmt.Errorf("failed to marshal item: %w", err)
	}

	_, err = r.Cache.Set(ctx, key, itemJSON, expiration).Result()
	if err != nil {
		return "", fmt.Errorf("failed to store in Redis: %w", err)
	}

	return key, nil
}

// Get retrieves items and deserializes them into the expected type.
func (r *Redis) Get(ctx context.Context, key string) (*cache.VideoFingerprint, error) {
	item, err := r.Cache.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("redis get error: %w", err)
	}

	var fingerprint cache.VideoFingerprint
	if err := json.Unmarshal([]byte(item), &fingerprint); err != nil {
		return nil, fmt.Errorf("failed to unmarshal fingerprint: %w", err)
	}

	return &fingerprint, nil
}

// InvalidatePattern invalidates cache entries matching a specific pattern.
func (r *Redis) InvalidatePattern(ctx context.Context, pattern string) error {
	keys, err := r.Cache.Keys(ctx, pattern).Result()
	if err != nil {
		return fmt.Errorf("failed to retrieve keys: %w", err)
	}

	for _, key := range keys {
		if err := r.Cache.Del(ctx, key).Err(); err != nil {
			return fmt.Errorf("failed to delete key: %w", err)
		}
	}
	return nil
}
