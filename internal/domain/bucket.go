package domain

import (
	"context"
	"time"
)

type LimitationBucket struct {
	Tokens      float64
	LastUpdated int64
}

type RateLimitRepository interface {
	SaveBucket(ctx context.Context, key string, bucket *LimitationBucket, ttl time.Duration) error
	GetBucket(ctx context.Context, key string) (*LimitationBucket, error)
}

type RateLimitService interface {
	AllowRequest(ctx context.Context, key string) (bool, error)
}
