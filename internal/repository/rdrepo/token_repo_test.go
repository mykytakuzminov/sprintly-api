package rdrepo

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mykytakuzminov/sprintly-api/internal/domain"
)

func TestTokenRepo_Set(t *testing.T) {
	defer teardown(t)

	repo := NewTokenRepository(testClient)
	ctx := context.Background()

	err := repo.Set(ctx, "token", uuid.New(), time.Minute)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	_, err = repo.Get(ctx, "token")
	if errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected token to be set, got %v", err)
	}
}

func TestTokenRepo_Get(t *testing.T) {
	defer teardown(t)

	repo := NewTokenRepository(testClient)
	ctx := context.Background()
	userID := uuid.New()

	_ = repo.Set(ctx, "token", userID, time.Minute)

	foundID, err := repo.Get(ctx, "token")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if foundID != userID {
		t.Errorf("expected ID %v, got %v", userID, foundID)
	}
}

func TestTokenRepo_Get_NotFound(t *testing.T) {
	defer teardown(t)

	repo := NewTokenRepository(testClient)
	ctx := context.Background()

	_, err := repo.Get(ctx, "token")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestTokenRepo_Delete(t *testing.T) {
	defer teardown(t)

	repo := NewTokenRepository(testClient)
	ctx := context.Background()

	_ = repo.Set(ctx, "token", uuid.New(), time.Minute)

	err := repo.Delete(ctx, "token")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	_, err = repo.Get(ctx, "token")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected token to be deleted, got %v", err)
	}
}
