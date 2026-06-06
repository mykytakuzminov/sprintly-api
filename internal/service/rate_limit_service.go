package service

import (
	"context"
	"errors"
	"time"

	"github.com/mykytakuzminov/task-manager-api/internal/domain"
)

type RateLimitService struct {
	repo             domain.RateLimitRepository
	maxTokens        float64
	refillRatePerSec float64
	ttl              time.Duration
}

func NewRateLimitService(
	repo domain.RateLimitRepository,
	maxTokens float64,
	refillRatePerSec float64,
	ttl time.Duration,
) domain.RateLimitService {
	return &RateLimitService{
		repo:             repo,
		maxTokens:        maxTokens,
		refillRatePerSec: refillRatePerSec, ttl: ttl,
	}
}

func (s *RateLimitService) AllowRequest(ctx context.Context, key string) (bool, error) {
	now := time.Now().Unix()

	bucket, err := s.repo.GetBucket(ctx, key)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			bucket = &domain.LimitationBucket{
				Tokens:      s.maxTokens,
				LastUpdated: now,
			}
		} else {
			return false, nil
		}
	} else {
		elapsedSeconds := now - bucket.LastUpdated
		if elapsedSeconds > 0 {
			generatedTokens := float64(elapsedSeconds) * s.refillRatePerSec
			bucket.Tokens += generatedTokens

			if bucket.Tokens > s.maxTokens {
				bucket.Tokens = s.maxTokens
			}

			bucket.LastUpdated = now
		}
	}

	if bucket.Tokens >= 1.0 {
		bucket.Tokens -= 1.0

		err := s.repo.SaveBucket(ctx, key, bucket, s.ttl)
		if err != nil {
			return false, nil
		}

		return true, nil
	}

	return false, nil
}
