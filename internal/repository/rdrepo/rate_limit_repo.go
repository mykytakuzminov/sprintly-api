package rdrepo

import (
	"context"
	"strconv"
	"time"

	"github.com/mykytakuzminov/sprintly-api/internal/domain"
	"github.com/redis/go-redis/v9"
)

type RateLimitRepo struct {
	client *redis.Client
}

func NewRateLimitRepo(client *redis.Client) domain.RateLimitRepository {
	return &RateLimitRepo{client: client}
}

func (r *RateLimitRepo) SaveBucket(
	ctx context.Context,
	key string,
	bucket *domain.LimitationBucket,
	ttl time.Duration,
) error {
	data := map[string]interface{}{
		"tokens":       bucket.Tokens,
		"last_updated": bucket.LastUpdated,
	}

	pipe := r.client.Pipeline()
	pipe.HSet(ctx, key, data)
	pipe.Expire(ctx, key, ttl)
	_, err := pipe.Exec(ctx)
	return err
}

func (r *RateLimitRepo) GetBucket(
	ctx context.Context,
	key string,
) (*domain.LimitationBucket, error) {
	data, err := r.client.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, domain.ErrNotFound
	}

	tokens, err := strconv.ParseFloat(data["tokens"], 64)
	if err != nil {
		return nil, err
	}

	lastUpdated, err := strconv.ParseInt(data["last_updated"], 10, 64)
	if err != nil {
		return nil, err
	}

	return &domain.LimitationBucket{
		Tokens:      tokens,
		LastUpdated: lastUpdated,
	}, nil
}
