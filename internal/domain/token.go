package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type TokenRepository interface {
	Set(ctx context.Context, token string, userID uuid.UUID, ttl time.Duration) error
	Get(ctx context.Context, token string) (uuid.UUID, error)
	Delete(ctx context.Context, token string) error
}
