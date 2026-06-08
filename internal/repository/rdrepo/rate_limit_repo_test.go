package rdrepo

import (
	"context"
	"testing"
	"time"

	"github.com/mykytakuzminov/task-manager-api/internal/domain"
	"github.com/redis/go-redis/v9"
)

func createTestBucket(ctx context.Context, client *redis.Client) *domain.LimitationBucket {
	bucket := &domain.LimitationBucket{
		Tokens:      10.0,
		LastUpdated: time.Now().Unix(),
	}

	_ = NewRateLimitRepo(client).SaveBucket(ctx, "key", bucket, time.Minute)
	return bucket
}

func TestRateLimitRepo_SaveBucket(t *testing.T) {
	defer teardown(t)

	repo := NewRateLimitRepo(testClient)
	ctx := context.Background()

	bucket := &domain.LimitationBucket{
		Tokens:      10.0,
		LastUpdated: time.Now().Unix(),
	}

	err := repo.SaveBucket(ctx, "key", bucket, time.Minute)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestRateLimitRepo_GetBucket(t *testing.T) {
	defer teardown(t)

	repo := NewRateLimitRepo(testClient)
	ctx := context.Background()
	bucket := createTestBucket(ctx, testClient)

	found, err := repo.GetBucket(ctx, "key")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if found.Tokens != bucket.Tokens {
		t.Errorf("expected the same amount of tokens")
	}
}

func TestRateLimitRepo_GetBucket_NotFound(t *testing.T) {
	defer teardown(t)

	repo := NewRateLimitRepo(testClient)
	ctx := context.Background()

	_, err := repo.GetBucket(ctx, "key")
	if err == nil {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
