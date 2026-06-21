package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mykytakuzminov/sprintly-api/internal/domain"
	"go.uber.org/zap"
)

type MockUserService struct {
	registerFn       func(ctx context.Context, input *domain.RegisterInput) (*domain.User, error)
	changePasswordFn func(ctx context.Context, userID uuid.UUID, input *domain.ChangePasswordInput) error
	changeRoleFn     func(ctx context.Context, userID uuid.UUID, input *domain.ChangeRoleInput) error
	getByIDFn        func(ctx context.Context, userID uuid.UUID) (*domain.User, error)
	getAllFn         func(ctx context.Context, params *domain.ListParams) ([]*domain.User, error)
	deleteFn         func(ctx context.Context, userID uuid.UUID) error
}

func (m *MockUserService) Register(
	ctx context.Context,
	input *domain.RegisterInput,
) (*domain.User, error) {
	if m.registerFn != nil {
		return m.registerFn(ctx, input)
	}
	return nil, nil
}

func (m *MockUserService) ChangePassword(
	ctx context.Context,
	userID uuid.UUID,
	input *domain.ChangePasswordInput,
) error {
	if m.changePasswordFn != nil {
		return m.changePasswordFn(ctx, userID, input)
	}
	return nil
}

func (m *MockUserService) ChangeRole(
	ctx context.Context,
	userID uuid.UUID,
	input *domain.ChangeRoleInput,
) error {
	if m.changeRoleFn != nil {
		return m.changeRoleFn(ctx, userID, input)
	}
	return nil
}

func (m *MockUserService) GetByID(
	ctx context.Context,
	userID uuid.UUID,
) (*domain.User, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, userID)
	}
	return nil, nil
}

func (m *MockUserService) GetAll(
	ctx context.Context,
	params *domain.ListParams,
) ([]*domain.User, error) {
	if m.getAllFn != nil {
		return m.getAllFn(ctx, params)
	}
	return nil, nil
}

func (m *MockUserService) Delete(ctx context.Context, userID uuid.UUID) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, userID)
	}
	return nil
}

func newTestUserHandler(svc domain.UserService) *UserHandler {
	logger, _ := zap.NewDevelopment()
	return NewUserHandler(svc, logger.Sugar())
}

func withUserID(r *http.Request, userID uuid.UUID) *http.Request {
	ctx := context.WithValue(r.Context(), UserIDKey, userID)
	return r.WithContext(ctx)
}

func TestUserHandler_Register_Success(t *testing.T) {
	svc := &MockUserService{
		registerFn: func(_ context.Context, input *domain.RegisterInput) (*domain.User, error) {
			return &domain.User{
				ID:        uuid.New(),
				Email:     input.Email,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}, nil
		},
	}

	h := newTestUserHandler(svc)

	body, _ := json.Marshal(domain.RegisterInput{
		Email:    "test@example.com",
		Password: "password",
	})

	r := httptest.NewRequest(http.MethodPost, "/users/register", bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Register(w, r)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", w.Code)
	}

	var resp UserResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Email != "test@example.com" {
		t.Errorf("expected email test@example.com, got %v", resp.Email)
	}
}

func TestUserHandler_Register_NotJSON(t *testing.T) {
	svc := &MockUserService{}

	h := newTestUserHandler(svc)

	r := httptest.NewRequest(http.MethodPost, "/users/register", bytes.NewReader([]byte("not-json")))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Register(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestUserHandler_Register_InvalidBody(t *testing.T) {
	svc := &MockUserService{
		registerFn: func(_ context.Context, input *domain.RegisterInput) (*domain.User, error) {
			return nil, domain.ErrBadRequest
		},
	}

	h := newTestUserHandler(svc)

	body, _ := json.Marshal(domain.RegisterInput{})

	r := httptest.NewRequest(http.MethodPost, "/users/register", bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Register(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestUserHandler_Register_AlreadyExists(t *testing.T) {
	svc := &MockUserService{
		registerFn: func(_ context.Context, input *domain.RegisterInput) (*domain.User, error) {
			return nil, domain.ErrConflict
		},
	}

	h := newTestUserHandler(svc)

	body, _ := json.Marshal(domain.RegisterInput{})

	r := httptest.NewRequest(http.MethodPost, "/users/register", bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Register(w, r)

	if w.Code != http.StatusConflict {
		t.Fatalf("expected status 409, got %d", w.Code)
	}
}

func TestUserHandler_Register_InternalError(t *testing.T) {
	svc := &MockUserService{
		registerFn: func(_ context.Context, input *domain.RegisterInput) (*domain.User, error) {
			return nil, errors.New("unexpected error")
		},
	}

	h := newTestUserHandler(svc)

	body, _ := json.Marshal(domain.RegisterInput{})

	r := httptest.NewRequest(http.MethodPost, "/users/register", bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Register(w, r)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
}

func TestUserHandler_ChangePassword_Success(t *testing.T) {
	svc := &MockUserService{
		changePasswordFn: func(_ context.Context, _ uuid.UUID, _ *domain.ChangePasswordInput) error {
			return nil
		},
	}

	h := newTestUserHandler(svc)

	body, _ := json.Marshal(domain.ChangePasswordInput{
		OldPassword: "oldpassword",
		NewPassword: "newpassword",
	})

	r := httptest.NewRequest(http.MethodPatch, "/users/me/password", bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	r = withUserID(r, uuid.New())
	w := httptest.NewRecorder()

	h.ChangePassword(w, r)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", w.Code)
	}
}

func TestUserHandler_ChangePassword_Unauthorized(t *testing.T) {
	h := newTestUserHandler(&MockUserService{})

	body, _ := json.Marshal(domain.ChangePasswordInput{
		OldPassword: "oldpassword",
		NewPassword: "newpassword",
	})

	r := httptest.NewRequest(http.MethodPatch, "/users/me/password", bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.ChangePassword(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", w.Code)
	}
}

func TestUserHandler_ChangePassword_NotJSON(t *testing.T) {
	h := newTestUserHandler(&MockUserService{})

	r := httptest.NewRequest(http.MethodPatch, "/users/me/password", bytes.NewReader([]byte("not-json")))
	r.Header.Set("Content-Type", "application/json")
	r = withUserID(r, uuid.New())
	w := httptest.NewRecorder()

	h.ChangePassword(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestUserHandler_ChangePassword_NotFound(t *testing.T) {
	svc := &MockUserService{
		changePasswordFn: func(_ context.Context, _ uuid.UUID, _ *domain.ChangePasswordInput) error {
			return domain.ErrNotFound
		},
	}

	h := newTestUserHandler(svc)

	body, _ := json.Marshal(domain.ChangePasswordInput{
		OldPassword: "oldpassword",
		NewPassword: "newpassword",
	})

	r := httptest.NewRequest(http.MethodPatch, "/users/me/password", bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	r = withUserID(r, uuid.New())
	w := httptest.NewRecorder()

	h.ChangePassword(w, r)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", w.Code)
	}
}

func TestUserHandler_ChangePassword_InvalidCredentials(t *testing.T) {
	svc := &MockUserService{
		changePasswordFn: func(_ context.Context, _ uuid.UUID, _ *domain.ChangePasswordInput) error {
			return domain.ErrInvalidCredentials
		},
	}

	h := newTestUserHandler(svc)

	body, _ := json.Marshal(domain.ChangePasswordInput{
		OldPassword: "wrongpassword",
		NewPassword: "newpassword",
	})

	r := httptest.NewRequest(http.MethodPatch, "/users/me/password", bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	r = withUserID(r, uuid.New())
	w := httptest.NewRecorder()

	h.ChangePassword(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", w.Code)
	}
}

func TestUserHandler_ChangePassword_InternalError(t *testing.T) {
	svc := &MockUserService{
		changePasswordFn: func(_ context.Context, _ uuid.UUID, _ *domain.ChangePasswordInput) error {
			return errors.New("unexpected error")
		},
	}

	h := newTestUserHandler(svc)

	body, _ := json.Marshal(domain.ChangePasswordInput{
		OldPassword: "oldpassword",
		NewPassword: "newpassword",
	})

	r := httptest.NewRequest(http.MethodPatch, "/users/me/password", bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	r = withUserID(r, uuid.New())
	w := httptest.NewRecorder()

	h.ChangePassword(w, r)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
}

func TestUserHandler_Me_Success(t *testing.T) {
	userID := uuid.New()

	svc := &MockUserService{
		getByIDFn: func(_ context.Context, id uuid.UUID) (*domain.User, error) {
			return &domain.User{
				ID:        id,
				Email:     "test@example.com",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}, nil
		},
	}

	h := newTestUserHandler(svc)

	r := httptest.NewRequest(http.MethodGet, "/users/me", nil)
	r = withUserID(r, userID)
	w := httptest.NewRecorder()

	h.Me(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp UserResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.ID != userID {
		t.Errorf("expected user ID %v, got %v", userID, resp.ID)
	}
	if resp.Email != "test@example.com" {
		t.Errorf("expected email test@example.com, got %v", resp.Email)
	}
}

func TestUserHandler_Me_Unauthorized(t *testing.T) {
	h := newTestUserHandler(&MockUserService{})

	r := httptest.NewRequest(http.MethodGet, "/users/me", nil)
	w := httptest.NewRecorder()

	h.Me(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", w.Code)
	}
}

func TestUserHandler_Me_NotFound(t *testing.T) {
	svc := &MockUserService{
		getByIDFn: func(_ context.Context, _ uuid.UUID) (*domain.User, error) {
			return nil, domain.ErrNotFound
		},
	}

	h := newTestUserHandler(svc)

	r := httptest.NewRequest(http.MethodGet, "/users/me", nil)
	r = withUserID(r, uuid.New())
	w := httptest.NewRecorder()

	h.Me(w, r)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", w.Code)
	}
}

func TestUserHandler_Me_InternalError(t *testing.T) {
	svc := &MockUserService{
		getByIDFn: func(_ context.Context, _ uuid.UUID) (*domain.User, error) {
			return nil, errors.New("unexpected error")
		},
	}

	h := newTestUserHandler(svc)

	r := httptest.NewRequest(http.MethodGet, "/users/me", nil)
	r = withUserID(r, uuid.New())
	w := httptest.NewRecorder()

	h.Me(w, r)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
}
