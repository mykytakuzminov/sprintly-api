package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/mykytakuzminov/sprintly-api/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

type MockUserRepository struct {
	createFn         func(ctx context.Context, user *domain.User) error
	getByIDFn        func(ctx context.Context, id uuid.UUID) (*domain.User, error)
	getByEmailFn     func(ctx context.Context, email string) (*domain.User, error)
	getAll           func(ctx context.Context, params *domain.ListParams) ([]*domain.User, error)
	updatePasswordFn func(ctx context.Context, userID uuid.UUID, hashPassword string) error
	updateRoleFn     func(ctx context.Context, userID uuid.UUID, role string) error
	deleteFn         func(ctx context.Context, id uuid.UUID) error
}

func (m *MockUserRepository) Create(ctx context.Context, user *domain.User) error {
	if m.createFn != nil {
		return m.createFn(ctx, user)
	}
	return nil
}

func (m *MockUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	if m.getByEmailFn != nil {
		return m.getByEmailFn(ctx, email)
	}
	return nil, nil
}

func (m *MockUserRepository) GetAll(ctx context.Context, params *domain.ListParams) ([]*domain.User, error) {
	if m.getAll != nil {
		return m.getAll(ctx, params)
	}
	return nil, nil
}

func (m *MockUserRepository) UpdatePassword(ctx context.Context, userID uuid.UUID, hashPassword string) error {
	if m.updatePasswordFn != nil {
		return m.updatePasswordFn(ctx, userID, hashPassword)
	}
	return nil
}

func (m *MockUserRepository) UpdateRole(ctx context.Context, userID uuid.UUID, role string) error {
	if m.updateRoleFn != nil {
		return m.updateRoleFn(ctx, userID, role)
	}
	return nil
}

func (m *MockUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
}

func TestUserService_Register(t *testing.T) {
	repo := &MockUserRepository{
		getByEmailFn: func(_ context.Context, _ string) (*domain.User, error) {
			return nil, domain.ErrNotFound
		},
		createFn: func(_ context.Context, _ *domain.User) error {
			return nil
		},
	}

	svc := NewUserService(repo)

	user, err := svc.Register(context.Background(), &domain.RegisterInput{
		Email:    "test@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if user.Email != "test@example.com" {
		t.Errorf("expected email %v, got %v", "test@example.com", user.Email)
	}
}

func TestUserService_Register_InvalidRequestBody(t *testing.T) {
	svc := NewUserService(&MockUserRepository{})

	_, err := svc.Register(context.Background(), &domain.RegisterInput{
		Email:    "test",
		Password: "pass",
	})
	if !errors.Is(err, domain.ErrBadRequest) {
		t.Fatalf("expected ErrBadRequest, got %v", err)
	}
}

func TestUserService_Register_EmptyFields(t *testing.T) {
	svc := NewUserService(&MockUserRepository{})

	_, err := svc.Register(context.Background(), &domain.RegisterInput{
		Email:    "",
		Password: "",
	})
	if !errors.Is(err, domain.ErrBadRequest) {
		t.Fatalf("expected ErrBadRequest, got %v", err)
	}
}

func TestUserService_Register_DuplicateEmail(t *testing.T) {
	repo := &MockUserRepository{
		getByEmailFn: func(_ context.Context, _ string) (*domain.User, error) {
			return nil, nil
		},
	}

	svc := NewUserService(repo)

	_, err := svc.Register(context.Background(), &domain.RegisterInput{
		Email:    "test@example.com",
		Password: "password123",
	})
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("expected ErrConflict, got %v", err)
	}
}

func TestUserService_GetByID(t *testing.T) {
	userID := uuid.New()

	repo := &MockUserRepository{
		getByIDFn: func(_ context.Context, id uuid.UUID) (*domain.User, error) {
			return &domain.User{ID: id, Email: "test@example.com"}, nil
		},
	}

	svc := NewUserService(repo)

	user, err := svc.GetByID(context.Background(), userID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if user.ID != userID {
		t.Errorf("expected ID %v, got %v", userID, user.ID)
	}
}

func TestUserService_GetByID_NotFound(t *testing.T) {
	repo := &MockUserRepository{
		getByIDFn: func(_ context.Context, _ uuid.UUID) (*domain.User, error) {
			return nil, domain.ErrNotFound
		},
	}

	svc := NewUserService(repo)

	_, err := svc.GetByID(context.Background(), uuid.New())
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestUserService_GetAll(t *testing.T) {
	repo := &MockUserRepository{
		getAll: func(_ context.Context, _ *domain.ListParams) ([]*domain.User, error) {
			return []*domain.User{
				{ID: uuid.New(), Email: "a@example.com"},
				{ID: uuid.New(), Email: "b@example.com"},
			}, nil
		},
	}

	svc := NewUserService(repo)

	users, err := svc.GetAll(context.Background(), &domain.ListParams{Limit: 20, Offset: 0})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(users) != 2 {
		t.Errorf("expected 2 users, got %v", len(users))
	}
}

func TestUserService_GetAll_Empty(t *testing.T) {
	repo := &MockUserRepository{
		getAll: func(_ context.Context, _ *domain.ListParams) ([]*domain.User, error) {
			return []*domain.User{}, nil
		},
	}

	svc := NewUserService(repo)

	users, err := svc.GetAll(context.Background(), &domain.ListParams{Limit: 20, Offset: 0})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(users) != 0 {
		t.Errorf("expected 0 users, got %v", len(users))
	}
}

func TestUserService_GetAll_InternalError(t *testing.T) {
	repo := &MockUserRepository{
		getAll: func(_ context.Context, _ *domain.ListParams) ([]*domain.User, error) {
			return nil, errors.New("unexpected error")
		},
	}

	svc := NewUserService(repo)

	_, err := svc.GetAll(context.Background(), &domain.ListParams{Limit: 20, Offset: 0})
	if err == nil {
		t.Fatalf("expected unexpected error")
	}
}

func TestUserService_ChangePassword(t *testing.T) {
	hpwd, _ := bcrypt.GenerateFromPassword([]byte("hashpassword"), bcrypt.MinCost)

	repo := &MockUserRepository{
		getByIDFn: func(_ context.Context, _ uuid.UUID) (*domain.User, error) {
			return &domain.User{
				HashPassword: string(hpwd),
			}, nil
		},
		updatePasswordFn: func(_ context.Context, _ uuid.UUID, _ string) error {
			return nil
		},
	}

	svc := NewUserService(repo)

	err := svc.ChangePassword(context.Background(), uuid.New(), &domain.ChangePasswordInput{
		OldPassword: "hashpassword",
		NewPassword: "newpassword",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestUserService_ChangePassword_NotFound(t *testing.T) {
	repo := &MockUserRepository{
		getByIDFn: func(_ context.Context, _ uuid.UUID) (*domain.User, error) {
			return nil, domain.ErrNotFound
		},
	}

	svc := NewUserService(repo)

	err := svc.ChangePassword(context.Background(), uuid.New(), &domain.ChangePasswordInput{
		OldPassword: "hashpassword",
		NewPassword: "newpassword",
	})
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestUserService_ChangePassword_InvalidRequestBody(t *testing.T) {
	svc := NewUserService(&MockUserRepository{})

	err := svc.ChangePassword(context.Background(), uuid.New(), &domain.ChangePasswordInput{
		OldPassword: "short",
		NewPassword: "short",
	})
	if !errors.Is(err, domain.ErrBadRequest) {
		t.Fatalf("expected ErrBadRequest, got %v", err)
	}
}

func TestUserService_ChangePassword_InvalidCredentials(t *testing.T) {
	hpwd, _ := bcrypt.GenerateFromPassword([]byte("hashpassword"), bcrypt.MinCost)

	repo := &MockUserRepository{
		getByIDFn: func(_ context.Context, _ uuid.UUID) (*domain.User, error) {
			return &domain.User{
				HashPassword: string(hpwd),
			}, nil
		},
	}

	svc := NewUserService(repo)

	err := svc.ChangePassword(context.Background(), uuid.New(), &domain.ChangePasswordInput{
		OldPassword: "password",
		NewPassword: "newpassword",
	})
	if !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestUserService_ChangeRole(t *testing.T) {
	userID := uuid.New()

	repo := &MockUserRepository{
		getByIDFn: func(_ context.Context, id uuid.UUID) (*domain.User, error) {
			return &domain.User{ID: id, Role: "member"}, nil
		},
		updateRoleFn: func(_ context.Context, _ uuid.UUID, _ string) error {
			return nil
		},
	}

	svc := NewUserService(repo)

	err := svc.ChangeRole(context.Background(), userID, &domain.ChangeRoleInput{Role: "admin"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestUserService_ChangeRole_InvalidRequestBody(t *testing.T) {
	svc := NewUserService(&MockUserRepository{})

	err := svc.ChangeRole(context.Background(), uuid.New(), &domain.ChangeRoleInput{Role: "superadmin"})
	if !errors.Is(err, domain.ErrBadRequest) {
		t.Fatalf("expected ErrBadRequest, got %v", err)
	}
}

func TestUserService_ChangeRole_NotFound(t *testing.T) {
	repo := &MockUserRepository{
		getByIDFn: func(_ context.Context, _ uuid.UUID) (*domain.User, error) {
			return nil, domain.ErrNotFound
		},
	}

	svc := NewUserService(repo)

	err := svc.ChangeRole(context.Background(), uuid.New(), &domain.ChangeRoleInput{Role: "admin"})
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestUserService_ChangeRole_InternalError(t *testing.T) {
	repo := &MockUserRepository{
		getByIDFn: func(_ context.Context, id uuid.UUID) (*domain.User, error) {
			return &domain.User{ID: id}, nil
		},
		updateRoleFn: func(_ context.Context, _ uuid.UUID, _ string) error {
			return errors.New("unexpected error")
		},
	}

	svc := NewUserService(repo)

	err := svc.ChangeRole(context.Background(), uuid.New(), &domain.ChangeRoleInput{Role: "admin"})
	if err == nil {
		t.Fatalf("expected unexpected error")
	}
}

func TestUserService_Delete(t *testing.T) {
	userID := uuid.New()

	repo := &MockUserRepository{
		getByIDFn: func(_ context.Context, id uuid.UUID) (*domain.User, error) {
			return &domain.User{ID: id}, nil
		},
		deleteFn: func(_ context.Context, _ uuid.UUID) error {
			return nil
		},
	}

	svc := NewUserService(repo)

	err := svc.Delete(context.Background(), userID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestUserService_Delete_NotFound(t *testing.T) {
	repo := &MockUserRepository{
		getByIDFn: func(_ context.Context, _ uuid.UUID) (*domain.User, error) {
			return nil, domain.ErrNotFound
		},
	}

	svc := NewUserService(repo)

	err := svc.Delete(context.Background(), uuid.New())
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestUserService_Delete_InternalError(t *testing.T) {
	repo := &MockUserRepository{
		getByIDFn: func(_ context.Context, id uuid.UUID) (*domain.User, error) {
			return &domain.User{ID: id}, nil
		},
		deleteFn: func(_ context.Context, _ uuid.UUID) error {
			return errors.New("unexpected error")
		},
	}

	svc := NewUserService(repo)

	err := svc.Delete(context.Background(), uuid.New())
	if err == nil {
		t.Fatalf("expected unexpected error")
	}
}
