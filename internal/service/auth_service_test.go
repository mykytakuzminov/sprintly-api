package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mykytakuzminov/sprintly-api/internal/auth"
	"github.com/mykytakuzminov/sprintly-api/internal/config"
	"github.com/mykytakuzminov/sprintly-api/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

type MockTokenRepository struct {
	setFn    func(ctx context.Context, token string, userID uuid.UUID, ttl time.Duration) error
	getFn    func(ctx context.Context, token string) (uuid.UUID, error)
	deleteFn func(ctx context.Context, token string) error
}

func (m *MockTokenRepository) Set(ctx context.Context, token string, userID uuid.UUID, ttl time.Duration) error {
	if m.setFn != nil {
		return m.setFn(ctx, token, userID, ttl)
	}
	return nil
}

func (m *MockTokenRepository) Get(ctx context.Context, token string) (uuid.UUID, error) {
	if m.getFn != nil {
		return m.getFn(ctx, token)
	}
	return uuid.Nil, nil
}

func (m *MockTokenRepository) Delete(ctx context.Context, token string) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, token)
	}
	return nil
}

func createAuth() *auth.Auth {
	return auth.NewAuth(&config.JWTConfig{
		Secret:     "testsecret",
		AccessTTL:  15 * time.Minute,
		RefreshTTL: 168 * time.Hour,
	})
}

func TestAuthService_Login(t *testing.T) {
	hpwd, _ := bcrypt.GenerateFromPassword([]byte("hashpassword"), bcrypt.MinCost)

	user := &domain.User{
		ID:           uuid.New(),
		Email:        "test@example.com",
		HashPassword: string(hpwd),
		Role:         "member",
	}

	userRepo := &MockUserRepository{
		getByEmailFn: func(_ context.Context, _ string) (*domain.User, error) {
			return user, nil
		},
	}
	tokenRepo := &MockTokenRepository{
		setFn: func(_ context.Context, _ string, _ uuid.UUID, _ time.Duration) error {
			return nil
		},
	}
	auth := createAuth()

	svc := NewAuthService(userRepo, tokenRepo, auth)

	tokens, err := svc.Login(context.Background(), &domain.LoginInput{
		Email:    "test@example.com",
		Password: "hashpassword",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if tokens.AccessToken == "" {
		t.Errorf("expected access token to be set")
	}
	if tokens.RefreshToken == "" {
		t.Errorf("expected refresh token to be set")
	}

	userID, role, err := auth.ParseToken(tokens.AccessToken)
	if err != nil {
		t.Fatalf("expected valid access token, got error: %v", err)
	}
	if role != "member" {
		t.Errorf("expected role member, got %v", role)
	}
	if userID != user.ID {
		t.Errorf("expected user ID %v, got %v", user.ID, userID)
	}

	userID, role, err = auth.ParseToken(tokens.RefreshToken)
	if err != nil {
		t.Fatalf("expected valid refresh token, got error: %v", err)
	}
	if role != "member" {
		t.Errorf("expected role member, got %v", role)
	}
	if userID != user.ID {
		t.Errorf("expected user ID %v, got %v", user.ID, userID)
	}
}

func TestAuthService_Login_NotFound(t *testing.T) {
	userRepo := &MockUserRepository{
		getByEmailFn: func(_ context.Context, _ string) (*domain.User, error) {
			return nil, domain.ErrNotFound
		},
	}
	tokenRepo := &MockTokenRepository{}
	auth := createAuth()

	svc := NewAuthService(userRepo, tokenRepo, auth)

	_, err := svc.Login(context.Background(), &domain.LoginInput{
		Email:    "test@example.com",
		Password: "hashpassword",
	})
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestAuthService_Login_InvalidCredentials(t *testing.T) {
	hpwd, _ := bcrypt.GenerateFromPassword([]byte("hashpassword"), bcrypt.MinCost)

	user := &domain.User{
		ID:           uuid.New(),
		Email:        "test@example.com",
		HashPassword: string(hpwd),
	}

	userRepo := &MockUserRepository{
		getByEmailFn: func(_ context.Context, _ string) (*domain.User, error) {
			return user, nil
		},
	}
	tokenRepo := &MockTokenRepository{}
	auth := createAuth()

	svc := NewAuthService(userRepo, tokenRepo, auth)

	_, err := svc.Login(context.Background(), &domain.LoginInput{
		Email:    "test@example.com",
		Password: "invalidpassword",
	})
	if !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestAuthService_Login_TokenSaveError(t *testing.T) {
	hpwd, _ := bcrypt.GenerateFromPassword([]byte("hashpassword"), bcrypt.MinCost)

	user := &domain.User{
		ID:           uuid.New(),
		Email:        "test@example.com",
		HashPassword: string(hpwd),
	}

	userRepo := &MockUserRepository{
		getByEmailFn: func(_ context.Context, _ string) (*domain.User, error) {
			return user, nil
		},
	}
	tokenRepo := &MockTokenRepository{
		setFn: func(_ context.Context, _ string, _ uuid.UUID, _ time.Duration) error {
			return errors.New("unexpected error")
		},
	}
	auth := createAuth()

	svc := NewAuthService(userRepo, tokenRepo, auth)

	_, err := svc.Login(context.Background(), &domain.LoginInput{
		Email:    "test@example.com",
		Password: "hashpassword",
	})
	if err == nil {
		t.Fatalf("expected unexpected error")
	}
}

func TestAuthService_Refresh(t *testing.T) {
	userID := uuid.New()

	userRepo := &MockUserRepository{
		getByIDFn: func(_ context.Context, _ uuid.UUID) (*domain.User, error) {
			return &domain.User{
				ID:   userID,
				Role: "member",
			}, nil
		},
	}
	tokenRepo := &MockTokenRepository{
		getFn: func(_ context.Context, _ string) (uuid.UUID, error) {
			return userID, nil
		},
	}
	auth := createAuth()

	svc := NewAuthService(userRepo, tokenRepo, auth)

	atoken, err := svc.Refresh(context.Background(), &domain.RefreshInput{
		RefreshToken: "token",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if atoken == "" {
		t.Errorf("expected access token to be set")
	}

	userIDFromToken, _, err := auth.ParseToken(atoken)
	if err != nil {
		t.Fatalf("expected valid access token, got error: %v", err)
	}
	if userID != userIDFromToken {
		t.Errorf("expected user ID %v, got %v", userID, userIDFromToken)
	}
}

func TestAuthService_Refresh_Unauthorized(t *testing.T) {
	userRepo := &MockUserRepository{}
	tokenRepo := &MockTokenRepository{
		getFn: func(_ context.Context, _ string) (uuid.UUID, error) {
			return uuid.Nil, domain.ErrNotFound
		},
	}
	auth := createAuth()

	svc := NewAuthService(userRepo, tokenRepo, auth)

	atoken, err := svc.Refresh(context.Background(), &domain.RefreshInput{
		RefreshToken: "token",
	})
	if !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got %v", err)
	}
	if atoken != "" {
		t.Fatalf("expected empty string, got %v", atoken)
	}
}

func TestAuthService_Refresh_TokenGetError(t *testing.T) {
	userRepo := &MockUserRepository{}
	tokenRepo := &MockTokenRepository{
		getFn: func(_ context.Context, _ string) (uuid.UUID, error) {
			return uuid.Nil, errors.New("unexpected error")
		},
	}
	auth := createAuth()

	svc := NewAuthService(userRepo, tokenRepo, auth)

	atoken, err := svc.Refresh(context.Background(), &domain.RefreshInput{
		RefreshToken: "token",
	})
	if err == nil {
		t.Fatalf("expected unexpected error")
	}
	if atoken != "" {
		t.Fatalf("expected empty string, got %v", atoken)
	}
}
