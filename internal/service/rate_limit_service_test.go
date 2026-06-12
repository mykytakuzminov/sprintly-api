package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/mykytakuzminov/task-manager-api/internal/domain"
)

type MockRateLimitRepository struct {
	saveBucketFn func(ctx context.Context, key string, bucket *domain.LimitationBucket, ttl time.Duration) error
	getBucketFn  func(ctx context.Context, key string) (*domain.LimitationBucket, error)
}

func (m *MockRateLimitRepository) SaveBucket(
	ctx context.Context,
	key string,
	bucket *domain.LimitationBucket,
	ttl time.Duration,
) error {
	if m.saveBucketFn != nil {
		return m.saveBucketFn(ctx, key, bucket, ttl)
	}
	return nil
}

func (m *MockRateLimitRepository) GetBucket(
	ctx context.Context,
	key string,
) (*domain.LimitationBucket, error) {
	if m.getBucketFn != nil {
		return m.getBucketFn(ctx, key)
	}
	return nil, nil
}

func TestRateLimitService_AllowRequest(t *testing.T) {
	bucket := &domain.LimitationBucket{
		Tokens:      9.0,
		LastUpdated: time.Now().Unix(),
	}

	repo := &MockRateLimitRepository{
		getBucketFn: func(_ context.Context, _ string) (*domain.LimitationBucket, error) {
			return bucket, nil
		},
	}

	svc := NewRateLimitService(repo, 10.0, 1.0, time.Minute)

	allowance, err := svc.AllowRequest(context.Background(), "key")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !allowance {
		t.Errorf("expected true, got false")
	}
}

func TestRateLimitService_AllowRequest_TokenRefill(t *testing.T) {
	bucket := &domain.LimitationBucket{
		Tokens:      0.0,
		LastUpdated: time.Now().Unix() - 3,
	}

	repo := &MockRateLimitRepository{
		getBucketFn: func(_ context.Context, _ string) (*domain.LimitationBucket, error) {
			return bucket, nil
		},
	}

	svc := NewRateLimitService(repo, 10.0, 1.0, time.Minute)

	allowance, err := svc.AllowRequest(context.Background(), "key")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !allowance {
		t.Errorf("expected true, got false")
	}
	if bucket.Tokens < 1.0 || bucket.Tokens > 3.0 {
		t.Errorf("expected tokens between 1.0 and 3.0, got %f", bucket.Tokens)
	}
}

func TestRateLimitService_AllowRequest_TokenCap(t *testing.T) {
	bucket := &domain.LimitationBucket{
		Tokens:      0.0,
		LastUpdated: time.Now().Unix() - 100,
	}

	repo := &MockRateLimitRepository{
		getBucketFn: func(_ context.Context, _ string) (*domain.LimitationBucket, error) {
			return bucket, nil
		},
	}

	svc := NewRateLimitService(repo, 10.0, 1.0, time.Minute)

	allowance, err := svc.AllowRequest(context.Background(), "key")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !allowance {
		t.Errorf("expected true, got false")
	}
	if bucket.Tokens < 8.0 || bucket.Tokens > 9.0 {
		t.Errorf("expected tokens between 8.0 and 9.0, got %f", bucket.Tokens)
	}
}

func TestRateLimitService_AllowRequest_NotFound(t *testing.T) {
	repo := &MockRateLimitRepository{
		getBucketFn: func(_ context.Context, _ string) (*domain.LimitationBucket, error) {
			return nil, domain.ErrNotFound
		},
	}

	svc := NewRateLimitService(repo, 10.0, 1.0, time.Minute)

	allowance, err := svc.AllowRequest(context.Background(), "key")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !allowance {
		t.Errorf("expected true, got false")
	}
}

func TestRateLimitService_AllowRequest_UnexpectedGetBucketError(t *testing.T) {
	repo := &MockRateLimitRepository{
		getBucketFn: func(_ context.Context, _ string) (*domain.LimitationBucket, error) {
			return nil, errors.New("unexpected error")
		},
	}

	svc := NewRateLimitService(repo, 10.0, 1.0, time.Minute)

	allowance, err := svc.AllowRequest(context.Background(), "key")
	if err == nil {
		t.Fatalf("expected unexpected error, got %v", err)
	}
	if allowance {
		t.Errorf("expected false, got true")
	}
}

func TestRateLimitService_AllowRequest_NotEnoughTokens(t *testing.T) {
	bucket := &domain.LimitationBucket{
		Tokens:      0.0,
		LastUpdated: time.Now().Unix(),
	}

	repo := &MockRateLimitRepository{
		getBucketFn: func(_ context.Context, _ string) (*domain.LimitationBucket, error) {
			return bucket, nil
		},
	}

	svc := NewRateLimitService(repo, 10.0, 1.0, time.Minute)

	allowance, err := svc.AllowRequest(context.Background(), "key")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if allowance {
		t.Errorf("expected false, got true")
	}
	if bucket.Tokens >= 1.0 {
		t.Errorf("expected less than 1.0 token, got %f", bucket.Tokens)
	}
}

func TestRateLimitService_AllowRequest_UnexpectedSaveBucketError(t *testing.T) {
	bucket := &domain.LimitationBucket{
		Tokens:      9.0,
		LastUpdated: time.Now().Unix(),
	}

	repo := &MockRateLimitRepository{
		getBucketFn: func(_ context.Context, _ string) (*domain.LimitationBucket, error) {
			return bucket, nil
		},
		saveBucketFn: func(_ context.Context, _ string, _ *domain.LimitationBucket, _ time.Duration) error {
			return errors.New("unexpected error")
		},
	}

	svc := NewRateLimitService(repo, 10.0, 1.0, time.Minute)

	allowance, err := svc.AllowRequest(context.Background(), "key")
	if err == nil {
		t.Fatalf("expected unexpected error, got %v", err)
	}
	if allowance {
		t.Errorf("expected false, got true")
	}
}
