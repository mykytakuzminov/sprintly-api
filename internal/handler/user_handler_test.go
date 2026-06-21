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
	"github.com/mykytakuzminov/sprintly-api/internal/auth"
	"github.com/mykytakuzminov/sprintly-api/internal/config"
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

func createAuth() *auth.Auth {
	return auth.NewAuth(&config.JWTConfig{
		Secret:     "testsecret",
		AccessTTL:  15 * time.Minute,
		RefreshTTL: 168 * time.Hour,
	})
}

func newTestUserHandler(svc domain.UserService) *UserHandler {
	logger, _ := zap.NewDevelopment()
	return NewUserHandler(svc, createAuth(), logger.Sugar())
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

func TestUserHandler_GetByID_Success(t *testing.T) {
	userID := uuid.New()
	adminID := uuid.New()

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

	r := httptest.NewRequest(http.MethodGet, "/admin/users/"+userID.String(), nil)
	r = withUserID(r, adminID)
	r = withURLParam(r, "id", userID.String())
	w := httptest.NewRecorder()

	h.GetByID(w, r)

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
}

func TestUserHandler_GetByID_Unauthorized(t *testing.T) {
	h := newTestUserHandler(&MockUserService{})

	userID := uuid.New()
	r := httptest.NewRequest(http.MethodGet, "/admin/users/"+userID.String(), nil)
	r = withURLParam(r, "id", userID.String())
	w := httptest.NewRecorder()

	h.GetByID(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", w.Code)
	}
}

func TestUserHandler_GetByID_InvalidUUID(t *testing.T) {
	h := newTestUserHandler(&MockUserService{})

	r := httptest.NewRequest(http.MethodGet, "/admin/users/not-a-uuid", nil)
	r = withUserID(r, uuid.New())
	r = withURLParam(r, "id", "not-a-uuid")
	w := httptest.NewRecorder()

	h.GetByID(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestUserHandler_GetByID_NotFound(t *testing.T) {
	svc := &MockUserService{
		getByIDFn: func(_ context.Context, _ uuid.UUID) (*domain.User, error) {
			return nil, domain.ErrNotFound
		},
	}

	h := newTestUserHandler(svc)

	userID := uuid.New()
	r := httptest.NewRequest(http.MethodGet, "/admin/users/"+userID.String(), nil)
	r = withUserID(r, uuid.New())
	r = withURLParam(r, "id", userID.String())
	w := httptest.NewRecorder()

	h.GetByID(w, r)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", w.Code)
	}
}

func TestUserHandler_GetByID_InternalError(t *testing.T) {
	svc := &MockUserService{
		getByIDFn: func(_ context.Context, _ uuid.UUID) (*domain.User, error) {
			return nil, errors.New("unexpected error")
		},
	}

	h := newTestUserHandler(svc)

	userID := uuid.New()
	r := httptest.NewRequest(http.MethodGet, "/admin/users/"+userID.String(), nil)
	r = withUserID(r, uuid.New())
	r = withURLParam(r, "id", userID.String())
	w := httptest.NewRecorder()

	h.GetByID(w, r)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
}

func TestUserHandler_GetAll_Success(t *testing.T) {
	svc := &MockUserService{
		getAllFn: func(_ context.Context, _ *domain.ListParams) ([]*domain.User, error) {
			return []*domain.User{
				{ID: uuid.New(), Email: "a@example.com", CreatedAt: time.Now(), UpdatedAt: time.Now()},
				{ID: uuid.New(), Email: "b@example.com", CreatedAt: time.Now(), UpdatedAt: time.Now()},
			}, nil
		},
	}

	h := newTestUserHandler(svc)

	r := httptest.NewRequest(http.MethodGet, "/admin/users/all", nil)
	r = withUserID(r, uuid.New())
	w := httptest.NewRecorder()

	h.GetAll(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp []UserResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(resp) != 2 {
		t.Errorf("expected 2 users, got %d", len(resp))
	}
}

func TestUserHandler_GetAll_Empty(t *testing.T) {
	svc := &MockUserService{
		getAllFn: func(_ context.Context, _ *domain.ListParams) ([]*domain.User, error) {
			return []*domain.User{}, nil
		},
	}

	h := newTestUserHandler(svc)

	r := httptest.NewRequest(http.MethodGet, "/admin/users/all", nil)
	r = withUserID(r, uuid.New())
	w := httptest.NewRecorder()

	h.GetAll(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
}

func TestUserHandler_GetAll_Unauthorized(t *testing.T) {
	h := newTestUserHandler(&MockUserService{})

	r := httptest.NewRequest(http.MethodGet, "/admin/users/all", nil)
	w := httptest.NewRecorder()

	h.GetAll(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", w.Code)
	}
}

func TestUserHandler_GetAll_InternalError(t *testing.T) {
	svc := &MockUserService{
		getAllFn: func(_ context.Context, _ *domain.ListParams) ([]*domain.User, error) {
			return nil, errors.New("unexpected error")
		},
	}

	h := newTestUserHandler(svc)

	r := httptest.NewRequest(http.MethodGet, "/admin/users/all", nil)
	r = withUserID(r, uuid.New())
	w := httptest.NewRecorder()

	h.GetAll(w, r)

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

func TestUserHandler_ChangeRole_Success(t *testing.T) {
	svc := &MockUserService{
		changeRoleFn: func(_ context.Context, _ uuid.UUID, _ *domain.ChangeRoleInput) error {
			return nil
		},
	}

	h := newTestUserHandler(svc)

	userID := uuid.New()
	body, _ := json.Marshal(domain.ChangeRoleInput{Role: "admin"})

	r := httptest.NewRequest(http.MethodPatch, "/admin/users/"+userID.String(), bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	r = withUserID(r, uuid.New())
	r = withURLParam(r, "id", userID.String())
	w := httptest.NewRecorder()

	h.ChangeRole(w, r)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", w.Code)
	}
}

func TestUserHandler_ChangeRole_Unauthorized(t *testing.T) {
	h := newTestUserHandler(&MockUserService{})

	userID := uuid.New()
	body, _ := json.Marshal(domain.ChangeRoleInput{Role: "admin"})

	r := httptest.NewRequest(http.MethodPatch, "/admin/users/"+userID.String(), bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	r = withURLParam(r, "id", userID.String())
	w := httptest.NewRecorder()

	h.ChangeRole(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", w.Code)
	}
}

func TestUserHandler_ChangeRole_InvalidUUID(t *testing.T) {
	h := newTestUserHandler(&MockUserService{})

	body, _ := json.Marshal(domain.ChangeRoleInput{Role: "admin"})

	r := httptest.NewRequest(http.MethodPatch, "/admin/users/not-a-uuid", bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	r = withUserID(r, uuid.New())
	r = withURLParam(r, "id", "not-a-uuid")
	w := httptest.NewRecorder()

	h.ChangeRole(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestUserHandler_ChangeRole_NotJSON(t *testing.T) {
	h := newTestUserHandler(&MockUserService{})

	userID := uuid.New()

	r := httptest.NewRequest(http.MethodPatch, "/admin/users/"+userID.String(), bytes.NewReader([]byte("not-json")))
	r.Header.Set("Content-Type", "application/json")
	r = withUserID(r, uuid.New())
	r = withURLParam(r, "id", userID.String())
	w := httptest.NewRecorder()

	h.ChangeRole(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestUserHandler_ChangeRole_InvalidBody(t *testing.T) {
	svc := &MockUserService{
		changeRoleFn: func(_ context.Context, _ uuid.UUID, _ *domain.ChangeRoleInput) error {
			return domain.ErrBadRequest
		},
	}

	h := newTestUserHandler(svc)

	userID := uuid.New()
	body, _ := json.Marshal(domain.ChangeRoleInput{Role: "superadmin"})

	r := httptest.NewRequest(http.MethodPatch, "/admin/users/"+userID.String(), bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	r = withUserID(r, uuid.New())
	r = withURLParam(r, "id", userID.String())
	w := httptest.NewRecorder()

	h.ChangeRole(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestUserHandler_ChangeRole_NotFound(t *testing.T) {
	svc := &MockUserService{
		changeRoleFn: func(_ context.Context, _ uuid.UUID, _ *domain.ChangeRoleInput) error {
			return domain.ErrNotFound
		},
	}

	h := newTestUserHandler(svc)

	userID := uuid.New()
	body, _ := json.Marshal(domain.ChangeRoleInput{Role: "admin"})

	r := httptest.NewRequest(http.MethodPatch, "/admin/users/"+userID.String(), bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	r = withUserID(r, uuid.New())
	r = withURLParam(r, "id", userID.String())
	w := httptest.NewRecorder()

	h.ChangeRole(w, r)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", w.Code)
	}
}

func TestUserHandler_ChangeRole_InternalError(t *testing.T) {
	svc := &MockUserService{
		changeRoleFn: func(_ context.Context, _ uuid.UUID, _ *domain.ChangeRoleInput) error {
			return errors.New("unexpected error")
		},
	}

	h := newTestUserHandler(svc)

	userID := uuid.New()
	body, _ := json.Marshal(domain.ChangeRoleInput{Role: "admin"})

	r := httptest.NewRequest(http.MethodPatch, "/admin/users/"+userID.String(), bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	r = withUserID(r, uuid.New())
	r = withURLParam(r, "id", userID.String())
	w := httptest.NewRecorder()

	h.ChangeRole(w, r)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
}

func TestUserHandler_Delete_Success(t *testing.T) {
	svc := &MockUserService{
		deleteFn: func(_ context.Context, _ uuid.UUID) error {
			return nil
		},
	}

	h := newTestUserHandler(svc)

	userID := uuid.New()

	r := httptest.NewRequest(http.MethodDelete, "/admin/users/"+userID.String(), nil)
	r = withUserID(r, uuid.New())
	r = withURLParam(r, "id", userID.String())
	w := httptest.NewRecorder()

	h.Delete(w, r)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", w.Code)
	}
}

func TestUserHandler_Delete_Unauthorized(t *testing.T) {
	h := newTestUserHandler(&MockUserService{})

	userID := uuid.New()

	r := httptest.NewRequest(http.MethodDelete, "/admin/users/"+userID.String(), nil)
	r = withURLParam(r, "id", userID.String())
	w := httptest.NewRecorder()

	h.Delete(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", w.Code)
	}
}

func TestUserHandler_Delete_InvalidUUID(t *testing.T) {
	h := newTestUserHandler(&MockUserService{})

	r := httptest.NewRequest(http.MethodDelete, "/admin/users/not-a-uuid", nil)
	r = withUserID(r, uuid.New())
	r = withURLParam(r, "id", "not-a-uuid")
	w := httptest.NewRecorder()

	h.Delete(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestUserHandler_Delete_NotFound(t *testing.T) {
	svc := &MockUserService{
		deleteFn: func(_ context.Context, _ uuid.UUID) error {
			return domain.ErrNotFound
		},
	}

	h := newTestUserHandler(svc)

	userID := uuid.New()

	r := httptest.NewRequest(http.MethodDelete, "/admin/users/"+userID.String(), nil)
	r = withUserID(r, uuid.New())
	r = withURLParam(r, "id", userID.String())
	w := httptest.NewRecorder()

	h.Delete(w, r)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", w.Code)
	}
}

func TestUserHandler_Delete_InternalError(t *testing.T) {
	svc := &MockUserService{
		deleteFn: func(_ context.Context, _ uuid.UUID) error {
			return errors.New("unexpected error")
		},
	}

	h := newTestUserHandler(svc)

	userID := uuid.New()

	r := httptest.NewRequest(http.MethodDelete, "/admin/users/"+userID.String(), nil)
	r = withUserID(r, uuid.New())
	r = withURLParam(r, "id", userID.String())
	w := httptest.NewRecorder()

	h.Delete(w, r)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
}
