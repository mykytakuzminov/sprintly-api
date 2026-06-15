package rdrepo

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/mykytakuzminov/sprintly-api/internal/domain"
	"github.com/redis/go-redis/v9"
)

type TokenRepo struct {
	client *redis.Client
}

func NewTokenRepository(client *redis.Client) *TokenRepo {
	return &TokenRepo{client: client}
}

func (r *TokenRepo) Set(
	ctx context.Context,
	token string,
	userID uuid.UUID,
	ttl time.Duration,
) error {
	return r.client.Set(ctx, token, userID.String(), ttl).Err()
}

func (r *TokenRepo) Get(ctx context.Context, token string) (uuid.UUID, error) {
	val, err := r.client.Get(ctx, token).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return uuid.Nil, domain.ErrNotFound
		}
		return uuid.Nil, err
	}

	return uuid.Parse(val)
}

func (r *TokenRepo) Delete(ctx context.Context, token string) error {
	return r.client.Del(ctx, token).Err()
}
