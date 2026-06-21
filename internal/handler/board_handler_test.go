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

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/mykytakuzminov/sprintly-api/internal/domain"
	"go.uber.org/zap"
)

type MockBoardService struct {
	createFn         func(ctx context.Context, userID uuid.UUID, input *domain.CreateBoardInput) (*domain.Board, error)
	getByIDFn        func(ctx context.Context, boardID uuid.UUID) (*domain.Board, error)
	getAllByUserIDFn func(ctx context.Context, userID uuid.UUID, params *domain.ListParams) ([]*domain.Board, error)
	updateFn         func(ctx context.Context, boardID, userID uuid.UUID, input *domain.UpdateBoardInput) error
	deleteFn         func(ctx context.Context, boardID, userID uuid.UUID) error
}

func (m *MockBoardService) Create(ctx context.Context, userID uuid.UUID, input *domain.CreateBoardInput) (*domain.Board, error) {
	if m.createFn != nil {
		return m.createFn(ctx, userID, input)
	}
	return nil, nil
}

func (m *MockBoardService) GetByID(ctx context.Context, boardID uuid.UUID) (*domain.Board, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, boardID)
	}
	return nil, nil
}

func (m *MockBoardService) GetAllByUserID(ctx context.Context, userID uuid.UUID, params *domain.ListParams) ([]*domain.Board, error) {
	if m.getAllByUserIDFn != nil {
		return m.getAllByUserIDFn(ctx, userID, params)
	}
	return []*domain.Board{}, nil
}

func (m *MockBoardService) Update(ctx context.Context, boardID, userID uuid.UUID, input *domain.UpdateBoardInput) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, boardID, userID, input)
	}
	return nil
}

func (m *MockBoardService) Delete(ctx context.Context, boardID, userID uuid.UUID) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, boardID, userID)
	}
	return nil
}

func newTestBoardHandler(svc domain.BoardService) *BoardHandler {
	logger, _ := zap.NewDevelopment()
	return NewBoardHandler(svc, logger.Sugar())
}

func withURLParam(r *http.Request, key, value string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, value)
	ctx := context.WithValue(r.Context(), chi.RouteCtxKey, rctx)
	return r.WithContext(ctx)
}

func newBoard(userID uuid.UUID) *domain.Board {
	return &domain.Board{
		ID:        uuid.New(),
		OwnerID:   userID,
		Name:      "test board",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func TestBoardHandler_Create_Success(t *testing.T) {
	userID := uuid.New()

	svc := &MockBoardService{
		createFn: func(_ context.Context, uid uuid.UUID, input *domain.CreateBoardInput) (*domain.Board, error) {
			return &domain.Board{
				ID:        uuid.New(),
				OwnerID:   uid,
				Name:      input.Name,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}, nil
		},
	}

	h := newTestBoardHandler(svc)

	body, _ := json.Marshal(domain.CreateBoardInput{Name: "test board"})

	r := httptest.NewRequest(http.MethodPost, "/boards", bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	r = withUserID(r, userID)
	w := httptest.NewRecorder()

	h.Create(w, r)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", w.Code)
	}

	var resp domain.Board
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Name != "test board" {
		t.Errorf("expected name 'test board', got %v", resp.Name)
	}
	if resp.OwnerID != userID {
		t.Errorf("expected ownerID %v, got %v", userID, resp.OwnerID)
	}
}

func TestBoardHandler_Create_Unauthorized(t *testing.T) {
	h := newTestBoardHandler(&MockBoardService{})

	body, _ := json.Marshal(domain.CreateBoardInput{Name: "test board"})

	r := httptest.NewRequest(http.MethodPost, "/boards", bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Create(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", w.Code)
	}
}

func TestBoardHandler_Create_NotJSON(t *testing.T) {
	h := newTestBoardHandler(&MockBoardService{})

	r := httptest.NewRequest(http.MethodPost, "/boards", bytes.NewReader([]byte("not-json")))
	r.Header.Set("Content-Type", "application/json")
	r = withUserID(r, uuid.New())
	w := httptest.NewRecorder()

	h.Create(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestBoardHandler_Create_InvalidBody(t *testing.T) {
	svc := &MockBoardService{
		createFn: func(_ context.Context, _ uuid.UUID, _ *domain.CreateBoardInput) (*domain.Board, error) {
			return nil, domain.ErrBadRequest
		},
	}

	h := newTestBoardHandler(svc)

	body, _ := json.Marshal(domain.CreateBoardInput{})

	r := httptest.NewRequest(http.MethodPost, "/boards", bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	r = withUserID(r, uuid.New())
	w := httptest.NewRecorder()

	h.Create(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestBoardHandler_Create_InternalError(t *testing.T) {
	svc := &MockBoardService{
		createFn: func(_ context.Context, _ uuid.UUID, _ *domain.CreateBoardInput) (*domain.Board, error) {
			return nil, errors.New("unexpected error")
		},
	}

	h := newTestBoardHandler(svc)

	body, _ := json.Marshal(domain.CreateBoardInput{Name: "test board"})

	r := httptest.NewRequest(http.MethodPost, "/boards", bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	r = withUserID(r, uuid.New())
	w := httptest.NewRecorder()

	h.Create(w, r)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
}

func TestBoardHandler_GetAll_Success(t *testing.T) {
	userID := uuid.New()

	svc := &MockBoardService{
		getAllByUserIDFn: func(_ context.Context, _ uuid.UUID, _ *domain.ListParams) ([]*domain.Board, error) {
			return []*domain.Board{
				newBoard(userID),
				newBoard(userID),
			}, nil
		},
	}

	h := newTestBoardHandler(svc)

	r := httptest.NewRequest(http.MethodGet, "/boards", nil)
	r = withUserID(r, userID)
	w := httptest.NewRecorder()

	h.GetAll(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp []*domain.Board
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(resp) != 2 {
		t.Errorf("expected 2 boards, got %d", len(resp))
	}
}

func TestBoardHandler_GetAll_Empty(t *testing.T) {
	svc := &MockBoardService{
		getAllByUserIDFn: func(_ context.Context, _ uuid.UUID, _ *domain.ListParams) ([]*domain.Board, error) {
			return []*domain.Board{}, nil
		},
	}

	h := newTestBoardHandler(svc)

	r := httptest.NewRequest(http.MethodGet, "/boards", nil)
	r = withUserID(r, uuid.New())
	w := httptest.NewRecorder()

	h.GetAll(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
}

func TestBoardHandler_GetAll_Unauthorized(t *testing.T) {
	h := newTestBoardHandler(&MockBoardService{})

	r := httptest.NewRequest(http.MethodGet, "/boards", nil)
	w := httptest.NewRecorder()

	h.GetAll(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", w.Code)
	}
}

func TestBoardHandler_GetAll_InternalError(t *testing.T) {
	svc := &MockBoardService{
		getAllByUserIDFn: func(_ context.Context, _ uuid.UUID, _ *domain.ListParams) ([]*domain.Board, error) {
			return nil, errors.New("unexpected error")
		},
	}

	h := newTestBoardHandler(svc)

	r := httptest.NewRequest(http.MethodGet, "/boards", nil)
	r = withUserID(r, uuid.New())
	w := httptest.NewRecorder()

	h.GetAll(w, r)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
}

func TestBoardHandler_GetByID_Success(t *testing.T) {
	userID := uuid.New()
	boardID := uuid.New()

	svc := &MockBoardService{
		getByIDFn: func(_ context.Context, id uuid.UUID) (*domain.Board, error) {
			return &domain.Board{
				ID:        id,
				OwnerID:   userID,
				Name:      "test board",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}, nil
		},
	}

	h := newTestBoardHandler(svc)

	r := httptest.NewRequest(http.MethodGet, "/boards/"+boardID.String(), nil)
	r = withUserID(r, userID)
	r = withURLParam(r, "id", boardID.String())
	w := httptest.NewRecorder()

	h.GetByID(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp domain.Board
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.ID != boardID {
		t.Errorf("expected board ID %v, got %v", boardID, resp.ID)
	}
}

func TestBoardHandler_GetByID_InvalidUUID(t *testing.T) {
	h := newTestBoardHandler(&MockBoardService{})

	r := httptest.NewRequest(http.MethodGet, "/boards/not-a-uuid", nil)
	r = withURLParam(r, "id", "not-a-uuid")
	w := httptest.NewRecorder()

	h.GetByID(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestBoardHandler_GetByID_NotFound(t *testing.T) {
	svc := &MockBoardService{
		getByIDFn: func(_ context.Context, _ uuid.UUID) (*domain.Board, error) {
			return nil, domain.ErrNotFound
		},
	}

	h := newTestBoardHandler(svc)

	boardID := uuid.New()
	r := httptest.NewRequest(http.MethodGet, "/boards/"+boardID.String(), nil)
	r = withURLParam(r, "id", boardID.String())
	w := httptest.NewRecorder()

	h.GetByID(w, r)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", w.Code)
	}
}

func TestBoardHandler_GetByID_InternalError(t *testing.T) {
	svc := &MockBoardService{
		getByIDFn: func(_ context.Context, _ uuid.UUID) (*domain.Board, error) {
			return nil, errors.New("unexpected error")
		},
	}

	h := newTestBoardHandler(svc)

	boardID := uuid.New()
	r := httptest.NewRequest(http.MethodGet, "/boards/"+boardID.String(), nil)
	r = withURLParam(r, "id", boardID.String())
	w := httptest.NewRecorder()

	h.GetByID(w, r)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
}

func TestBoardHandler_Update_Success(t *testing.T) {
	svc := &MockBoardService{
		updateFn: func(_ context.Context, _, _ uuid.UUID, _ *domain.UpdateBoardInput) error {
			return nil
		},
	}

	h := newTestBoardHandler(svc)

	boardID := uuid.New()
	body, _ := json.Marshal(domain.UpdateBoardInput{Name: "new name"})

	r := httptest.NewRequest(http.MethodPatch, "/boards/"+boardID.String(), bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	r = withUserID(r, uuid.New())
	r = withURLParam(r, "id", boardID.String())
	w := httptest.NewRecorder()

	h.Update(w, r)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", w.Code)
	}
}

func TestBoardHandler_Update_Unauthorized(t *testing.T) {
	h := newTestBoardHandler(&MockBoardService{})

	boardID := uuid.New()
	body, _ := json.Marshal(domain.UpdateBoardInput{Name: "new name"})

	r := httptest.NewRequest(http.MethodPatch, "/boards/"+boardID.String(), bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	r = withURLParam(r, "id", boardID.String())
	w := httptest.NewRecorder()

	h.Update(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", w.Code)
	}
}

func TestBoardHandler_Update_InvalidUUID(t *testing.T) {
	h := newTestBoardHandler(&MockBoardService{})

	body, _ := json.Marshal(domain.UpdateBoardInput{Name: "new name"})

	r := httptest.NewRequest(http.MethodPatch, "/boards/not-a-uuid", bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	r = withUserID(r, uuid.New())
	r = withURLParam(r, "id", "not-a-uuid")
	w := httptest.NewRecorder()

	h.Update(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestBoardHandler_Update_NotJSON(t *testing.T) {
	h := newTestBoardHandler(&MockBoardService{})

	boardID := uuid.New()

	r := httptest.NewRequest(http.MethodPatch, "/boards/"+boardID.String(), bytes.NewReader([]byte("not-json")))
	r.Header.Set("Content-Type", "application/json")
	r = withUserID(r, uuid.New())
	r = withURLParam(r, "id", boardID.String())
	w := httptest.NewRecorder()

	h.Update(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestBoardHandler_Update_NotFound(t *testing.T) {
	svc := &MockBoardService{
		updateFn: func(_ context.Context, _, _ uuid.UUID, _ *domain.UpdateBoardInput) error {
			return domain.ErrNotFound
		},
	}

	h := newTestBoardHandler(svc)

	boardID := uuid.New()
	body, _ := json.Marshal(domain.UpdateBoardInput{Name: "new name"})

	r := httptest.NewRequest(http.MethodPatch, "/boards/"+boardID.String(), bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	r = withUserID(r, uuid.New())
	r = withURLParam(r, "id", boardID.String())
	w := httptest.NewRecorder()

	h.Update(w, r)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", w.Code)
	}
}

func TestBoardHandler_Update_Forbidden(t *testing.T) {
	svc := &MockBoardService{
		updateFn: func(_ context.Context, _, _ uuid.UUID, _ *domain.UpdateBoardInput) error {
			return domain.ErrForbidden
		},
	}

	h := newTestBoardHandler(svc)

	boardID := uuid.New()
	body, _ := json.Marshal(domain.UpdateBoardInput{Name: "new name"})

	r := httptest.NewRequest(http.MethodPatch, "/boards/"+boardID.String(), bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	r = withUserID(r, uuid.New())
	r = withURLParam(r, "id", boardID.String())
	w := httptest.NewRecorder()

	h.Update(w, r)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d", w.Code)
	}
}

func TestBoardHandler_Update_InternalError(t *testing.T) {
	svc := &MockBoardService{
		updateFn: func(_ context.Context, _, _ uuid.UUID, _ *domain.UpdateBoardInput) error {
			return errors.New("unexpected error")
		},
	}

	h := newTestBoardHandler(svc)

	boardID := uuid.New()
	body, _ := json.Marshal(domain.UpdateBoardInput{Name: "new name"})

	r := httptest.NewRequest(http.MethodPatch, "/boards/"+boardID.String(), bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	r = withUserID(r, uuid.New())
	r = withURLParam(r, "id", boardID.String())
	w := httptest.NewRecorder()

	h.Update(w, r)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
}

func TestBoardHandler_Delete_Success(t *testing.T) {
	svc := &MockBoardService{
		deleteFn: func(_ context.Context, _, _ uuid.UUID) error {
			return nil
		},
	}

	h := newTestBoardHandler(svc)

	boardID := uuid.New()

	r := httptest.NewRequest(http.MethodDelete, "/boards/"+boardID.String(), nil)
	r = withUserID(r, uuid.New())
	r = withURLParam(r, "id", boardID.String())
	w := httptest.NewRecorder()

	h.Delete(w, r)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", w.Code)
	}
}

func TestBoardHandler_Delete_Unauthorized(t *testing.T) {
	h := newTestBoardHandler(&MockBoardService{})

	boardID := uuid.New()

	r := httptest.NewRequest(http.MethodDelete, "/boards/"+boardID.String(), nil)
	r = withURLParam(r, "id", boardID.String())
	w := httptest.NewRecorder()

	h.Delete(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", w.Code)
	}
}

func TestBoardHandler_Delete_InvalidUUID(t *testing.T) {
	h := newTestBoardHandler(&MockBoardService{})

	r := httptest.NewRequest(http.MethodDelete, "/boards/not-a-uuid", nil)
	r = withUserID(r, uuid.New())
	r = withURLParam(r, "id", "not-a-uuid")
	w := httptest.NewRecorder()

	h.Delete(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestBoardHandler_Delete_NotFound(t *testing.T) {
	svc := &MockBoardService{
		deleteFn: func(_ context.Context, _, _ uuid.UUID) error {
			return domain.ErrNotFound
		},
	}

	h := newTestBoardHandler(svc)

	boardID := uuid.New()

	r := httptest.NewRequest(http.MethodDelete, "/boards/"+boardID.String(), nil)
	r = withUserID(r, uuid.New())
	r = withURLParam(r, "id", boardID.String())
	w := httptest.NewRecorder()

	h.Delete(w, r)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", w.Code)
	}
}

func TestBoardHandler_Delete_Forbidden(t *testing.T) {
	svc := &MockBoardService{
		deleteFn: func(_ context.Context, _, _ uuid.UUID) error {
			return domain.ErrForbidden
		},
	}

	h := newTestBoardHandler(svc)

	boardID := uuid.New()

	r := httptest.NewRequest(http.MethodDelete, "/boards/"+boardID.String(), nil)
	r = withUserID(r, uuid.New())
	r = withURLParam(r, "id", boardID.String())
	w := httptest.NewRecorder()

	h.Delete(w, r)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d", w.Code)
	}
}

func TestBoardHandler_Delete_InternalError(t *testing.T) {
	svc := &MockBoardService{
		deleteFn: func(_ context.Context, _, _ uuid.UUID) error {
			return errors.New("unexpected error")
		},
	}

	h := newTestBoardHandler(svc)

	boardID := uuid.New()

	r := httptest.NewRequest(http.MethodDelete, "/boards/"+boardID.String(), nil)
	r = withUserID(r, uuid.New())
	r = withURLParam(r, "id", boardID.String())
	w := httptest.NewRecorder()

	h.Delete(w, r)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
}
