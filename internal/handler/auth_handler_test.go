package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mykytakuzminov/task-manager-api/internal/domain"
	"go.uber.org/zap"
)

type MockAuthService struct {
	loginFn   func(ctx context.Context, input *domain.LoginInput) (*domain.AuthTokens, error)
	logoutFn  func(ctx context.Context, input *domain.LogoutInput) error
	refreshFn func(ctx context.Context, input *domain.RefreshInput) (string, error)
}

func (m *MockAuthService) Login(ctx context.Context, input *domain.LoginInput) (*domain.AuthTokens, error) {
	if m.loginFn != nil {
		return m.loginFn(ctx, input)
	}
	return nil, nil
}

func (m *MockAuthService) Logout(ctx context.Context, input *domain.LogoutInput) error {
	if m.logoutFn != nil {
		return m.logoutFn(ctx, input)
	}
	return nil
}

func (m *MockAuthService) Refresh(ctx context.Context, input *domain.RefreshInput) (string, error) {
	if m.refreshFn != nil {
		return m.refreshFn(ctx, input)
	}
	return "", nil
}

func newTestAuthHandler(svc domain.AuthService) *AuthHandler {
	logger, _ := zap.NewDevelopment()
	return NewAuthHandler(svc, logger.Sugar())
}

func TestAuthHandler_Login_Success(t *testing.T) {
	svc := &MockAuthService{
		loginFn: func(_ context.Context, _ *domain.LoginInput) (*domain.AuthTokens, error) {
			return &domain.AuthTokens{
				AccessToken:  "access",
				RefreshToken: "refresh",
			}, nil
		},
	}

	h := newTestAuthHandler(svc)

	body, _ := json.Marshal(domain.LoginInput{
		Email:    "test@example.com",
		Password: "password123",
	})

	r := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Login(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp domain.AuthTokens
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.AccessToken == "" {
		t.Errorf("expected access token to be set")
	}
	if resp.RefreshToken == "" {
		t.Errorf("expected refresh token to be set")
	}
}

func TestAuthHandler_Login_NotJSON(t *testing.T) {
	h := newTestAuthHandler(&MockAuthService{})

	r := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader([]byte("not-json")))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Login(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestAuthHandler_Login_NotFound(t *testing.T) {
	svc := &MockAuthService{
		loginFn: func(_ context.Context, _ *domain.LoginInput) (*domain.AuthTokens, error) {
			return nil, domain.ErrNotFound
		},
	}

	h := newTestAuthHandler(svc)

	body, _ := json.Marshal(domain.LoginInput{
		Email:    "notexist@example.com",
		Password: "password123",
	})

	r := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Login(w, r)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", w.Code)
	}
}

func TestAuthHandler_Login_InvalidCredentials(t *testing.T) {
	svc := &MockAuthService{
		loginFn: func(_ context.Context, _ *domain.LoginInput) (*domain.AuthTokens, error) {
			return nil, domain.ErrInvalidCredentials
		},
	}

	h := newTestAuthHandler(svc)

	body, _ := json.Marshal(domain.LoginInput{
		Email:    "test@example.com",
		Password: "wrongpassword",
	})

	r := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Login(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", w.Code)
	}
}

func TestAuthHandler_Login_InternalError(t *testing.T) {
	svc := &MockAuthService{
		loginFn: func(_ context.Context, _ *domain.LoginInput) (*domain.AuthTokens, error) {
			return nil, errors.New("unexpected error")
		},
	}

	h := newTestAuthHandler(svc)

	body, _ := json.Marshal(domain.LoginInput{
		Email:    "test@example.com",
		Password: "password123",
	})

	r := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Login(w, r)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
}

func TestAuthHandler_Logout_Success(t *testing.T) {
	svc := &MockAuthService{
		logoutFn: func(_ context.Context, _ *domain.LogoutInput) error {
			return nil
		},
	}

	h := newTestAuthHandler(svc)

	body, _ := json.Marshal(domain.LogoutInput{
		RefreshToken: "refresh",
	})

	r := httptest.NewRequest(http.MethodPost, "/auth/logout", bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Logout(w, r)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", w.Code)
	}
}

func TestAuthHandler_Logout_NotJSON(t *testing.T) {
	h := newTestAuthHandler(&MockAuthService{})

	r := httptest.NewRequest(http.MethodPost, "/auth/logout", bytes.NewReader([]byte("not-json")))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Logout(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestAuthHandler_Logout_InternalError(t *testing.T) {
	svc := &MockAuthService{
		logoutFn: func(_ context.Context, _ *domain.LogoutInput) error {
			return errors.New("unexpected error")
		},
	}

	h := newTestAuthHandler(svc)

	body, _ := json.Marshal(domain.LogoutInput{
		RefreshToken: "refresh",
	})

	r := httptest.NewRequest(http.MethodPost, "/auth/logout", bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Logout(w, r)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
}

func TestAuthHandler_Refresh_Success(t *testing.T) {
	svc := &MockAuthService{
		refreshFn: func(_ context.Context, _ *domain.RefreshInput) (string, error) {
			return "new_access_token", nil
		},
	}

	h := newTestAuthHandler(svc)

	body, _ := json.Marshal(domain.RefreshInput{
		RefreshToken: "refresh",
	})

	r := httptest.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Refresh(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp RefreshResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.AccessToken != "new_access_token" {
		t.Errorf("expected access token 'new_access_token', got %v", resp.AccessToken)
	}
}

func TestAuthHandler_Refresh_NotJSON(t *testing.T) {
	h := newTestAuthHandler(&MockAuthService{})

	r := httptest.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewReader([]byte("not-json")))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Refresh(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestAuthHandler_Refresh_Unauthorized(t *testing.T) {
	svc := &MockAuthService{
		refreshFn: func(_ context.Context, _ *domain.RefreshInput) (string, error) {
			return "", domain.ErrUnauthorized
		},
	}

	h := newTestAuthHandler(svc)

	body, _ := json.Marshal(domain.RefreshInput{
		RefreshToken: "expired_or_invalid",
	})

	r := httptest.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Refresh(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", w.Code)
	}
}

func TestAuthHandler_Refresh_InternalError(t *testing.T) {
	svc := &MockAuthService{
		refreshFn: func(_ context.Context, _ *domain.RefreshInput) (string, error) {
			return "", errors.New("unexpected error")
		},
	}

	h := newTestAuthHandler(svc)

	body, _ := json.Marshal(domain.RefreshInput{
		RefreshToken: "refresh",
	})

	r := httptest.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Refresh(w, r)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
}
