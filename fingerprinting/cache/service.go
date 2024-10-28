package cache

import (
	"context"
	"time"
)

type VideoFingerprint struct {
	Hashes    []string  `json:"hashes"`
	Timestamp time.Time `json:"timestamp"`
}

// Service represents the cache service interface
type Service interface {
	Set(ctx context.Context, key string, item interface{}, expiration time.Duration) (string, error)
	Get(ctx context.Context, key string) (*VideoFingerprint, error)
}
