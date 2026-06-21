package pgrepo

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/mykytakuzminov/sprintly-api/internal/domain"
)

func createTestUser(ctx context.Context, db DB) *domain.User {
	user := &domain.User{
		ID:           uuid.New(),
		Email:        uuid.New().String() + "@example.com",
		HashPassword: "hashedpassword",
	}
	_ = NewUserRepository(db).Create(ctx, user)
	return user
}

func TestUserRepo_Create(t *testing.T) {
	withTx(t, func(ctx context.Context, db DB) {
		repo := NewUserRepository(db)

		user := &domain.User{
			ID:           uuid.New(),
			Email:        uuid.New().String() + "@example.com",
			HashPassword: "hashedpassword",
		}

		err := repo.Create(ctx, user)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if user.Role != "member" {
			t.Errorf("expected role member, got %v", user.Role)
		}
		if user.CreatedAt.IsZero() {
			t.Errorf("expected CreatedAt to be set")
		}
		if user.UpdatedAt.IsZero() {
			t.Errorf("expected UpdatedAt to be set")
		}
	})
}

func TestUserRepo_Create_DuplicateEmail(t *testing.T) {
	withTx(t, func(ctx context.Context, db DB) {
		repo := NewUserRepository(db)

		email := uuid.New().String() + "@example.com"

		_ = repo.Create(ctx, &domain.User{
			ID:           uuid.New(),
			Email:        email,
			HashPassword: "hashedpassword",
		})

		err := repo.Create(ctx, &domain.User{
			ID:           uuid.New(),
			Email:        email,
			HashPassword: "hashedpassword",
		})
		if !errors.Is(err, domain.ErrConflict) {
			t.Fatalf("expected ErrConflict, got %v", err)
		}
	})
}

func TestUserRepo_GetByID(t *testing.T) {
	withTx(t, func(ctx context.Context, db DB) {
		repo := NewUserRepository(db)
		user := createTestUser(ctx, db)

		found, err := repo.GetByID(ctx, user.ID)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if found.ID != user.ID {
			t.Errorf("expected ID %v, got %v", user.ID, found.ID)
		}
		if found.Email != user.Email {
			t.Errorf("expected Email %v, got %v", user.Email, found.Email)
		}
	})
}

func TestUserRepo_GetByID_NotFound(t *testing.T) {
	withTx(t, func(ctx context.Context, db DB) {
		repo := NewUserRepository(db)

		_, err := repo.GetByID(ctx, uuid.New())
		if !errors.Is(err, domain.ErrNotFound) {
			t.Fatalf("expected ErrNotFound, got %v", err)
		}
	})
}

func TestUserRepo_GetByEmail(t *testing.T) {
	withTx(t, func(ctx context.Context, db DB) {
		repo := NewUserRepository(db)
		user := createTestUser(ctx, db)

		found, err := repo.GetByEmail(ctx, user.Email)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if found.ID != user.ID {
			t.Errorf("expected ID %v, got %v", user.ID, found.ID)
		}
		if found.Email != user.Email {
			t.Errorf("expected Email %v, got %v", user.Email, found.Email)
		}
	})
}

func TestUserRepo_GetByEmail_NotFound(t *testing.T) {
	withTx(t, func(ctx context.Context, db DB) {
		repo := NewUserRepository(db)

		_, err := repo.GetByEmail(ctx, uuid.New().String()+"@example.com")
		if !errors.Is(err, domain.ErrNotFound) {
			t.Fatalf("expected ErrNotFound, got %v", err)
		}
	})
}

func TestUserRepo_UpdatePassword(t *testing.T) {
	withTx(t, func(ctx context.Context, db DB) {
		repo := NewUserRepository(db)
		user := createTestUser(ctx, db)

		err := repo.UpdatePassword(ctx, user.ID, "newhash")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		found, _ := repo.GetByID(ctx, user.ID)
		if found.HashPassword != "newhash" {
			t.Errorf("expected HashPassword %v, got %v", "newhash", found.HashPassword)
		}
	})
}

func TestUserRepo_Update_NotFound(t *testing.T) {
	withTx(t, func(ctx context.Context, db DB) {
		repo := NewUserRepository(db)

		err := repo.UpdatePassword(ctx, uuid.New(), "oldhash")
		if !errors.Is(err, domain.ErrNotFound) {
			t.Fatalf("expected ErrNotFound, got %v", err)
		}
	})
}

func TestUserRepo_Delete(t *testing.T) {
	withTx(t, func(ctx context.Context, db DB) {
		repo := NewUserRepository(db)
		user := createTestUser(ctx, db)

		err := repo.Delete(ctx, user.ID)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		_, err = repo.GetByID(ctx, user.ID)
		if !errors.Is(err, domain.ErrNotFound) {
			t.Errorf("expected user to be deleted, got %v", err)
		}
	})
}

func TestUserRepo_Delete_NotFound(t *testing.T) {
	withTx(t, func(ctx context.Context, db DB) {
		repo := NewUserRepository(db)

		err := repo.Delete(ctx, uuid.New())
		if !errors.Is(err, domain.ErrNotFound) {
			t.Fatalf("expected ErrNotFound, got %v", err)
		}
	})
}
