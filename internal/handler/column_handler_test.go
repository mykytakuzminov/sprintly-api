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
	"github.com/mykytakuzminov/task-manager-api/internal/domain"
	"go.uber.org/zap"
)

type MockColumnService struct {
	createFn          func(ctx context.Context, userID, boardID uuid.UUID, input *domain.CreateColumnInput) (*domain.Column, error)
	getByIDFn         func(ctx context.Context, columnID uuid.UUID) (*domain.Column, error)
	getAllByBoardIDFn func(ctx context.Context, boardID uuid.UUID, params *domain.ListParams) ([]*domain.Column, error)
	updateFn          func(ctx context.Context, columnID, userID uuid.UUID, input *domain.UpdateColumnInput) error
	deleteFn          func(ctx context.Context, columnID, userID uuid.UUID) error
}

func (m *MockColumnService) Create(ctx context.Context, userID, boardID uuid.UUID, input *domain.CreateColumnInput) (*domain.Column, error) {
	if m.createFn != nil {
		return m.createFn(ctx, userID, boardID, input)
	}
	return nil, nil
}

func (m *MockColumnService) GetByID(ctx context.Context, columnID uuid.UUID) (*domain.Column, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, columnID)
	}
	return nil, nil
}

func (m *MockColumnService) GetAllByBoardID(ctx context.Context, boardID uuid.UUID, params *domain.ListParams) ([]*domain.Column, error) {
	if m.getAllByBoardIDFn != nil {
		return m.getAllByBoardIDFn(ctx, boardID, params)
	}
	return []*domain.Column{}, nil
}

func (m *MockColumnService) Update(ctx context.Context, columnID, userID uuid.UUID, input *domain.UpdateColumnInput) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, columnID, userID, input)
	}
	return nil
}

func (m *MockColumnService) Delete(ctx context.Context, columnID, userID uuid.UUID) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, columnID, userID)
	}
	return nil
}

func newTestColumnHandler(svc domain.ColumnService) *ColumnHandler {
	logger, _ := zap.NewDevelopment()
	return NewColumnHandler(svc, createAuth(), logger.Sugar())
}

func newColumn(boardID uuid.UUID) *domain.Column {
	return &domain.Column{
		ID:        uuid.New(),
		BoardID:   boardID,
		Name:      "test column",
		Position:  1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func TestColumnHandler_Create_Success(t *testing.T) {
	userID := uuid.New()
	boardID := uuid.New()

	svc := &MockColumnService{
		createFn: func(_ context.Context, _, bid uuid.UUID, input *domain.CreateColumnInput) (*domain.Column, error) {
			return &domain.Column{
				ID:        uuid.New(),
				BoardID:   bid,
				Name:      input.Name,
				Position:  input.Position,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}, nil
		},
	}

	h := newTestColumnHandler(svc)

	body, _ := json.Marshal(domain.CreateColumnInput{Name: "In Progress", Position: 1})

	r := httptest.NewRequest(http.MethodPost, "/boards/"+boardID.String()+"/columns", bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	r = withUserID(r, userID)
	r = withURLParam(r, "boardID", boardID.String())
	w := httptest.NewRecorder()

	h.Create(w, r)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", w.Code)
	}

	var resp domain.Column
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Name != "In Progress" {
		t.Errorf("expected name 'In Progress', got %v", resp.Name)
	}
	if resp.BoardID != boardID {
		t.Errorf("expected boardID %v, got %v", boardID, resp.BoardID)
	}
}

func TestColumnHandler_Create_Unauthorized(t *testing.T) {
	h := newTestColumnHandler(&MockColumnService{})

	boardID := uuid.New()
	body, _ := json.Marshal(domain.CreateColumnInput{Name: "In Progress", Position: 1})

	r := httptest.NewRequest(http.MethodPost, "/boards/"+boardID.String()+"/columns", bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	r = withURLParam(r, "boardID", boardID.String())
	w := httptest.NewRecorder()

	h.Create(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", w.Code)
	}
}

func TestColumnHandler_Create_InvalidBoardUUID(t *testing.T) {
	h := newTestColumnHandler(&MockColumnService{})

	body, _ := json.Marshal(domain.CreateColumnInput{Name: "In Progress", Position: 1})

	r := httptest.NewRequest(http.MethodPost, "/boards/not-a-uuid/columns", bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	r = withUserID(r, uuid.New())
	r = withURLParam(r, "boardID", "not-a-uuid")
	w := httptest.NewRecorder()

	h.Create(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestColumnHandler_Create_NotJSON(t *testing.T) {
	h := newTestColumnHandler(&MockColumnService{})

	boardID := uuid.New()

	r := httptest.NewRequest(http.MethodPost, "/boards/"+boardID.String()+"/columns", bytes.NewReader([]byte("not-json")))
	r.Header.Set("Content-Type", "application/json")
	r = withUserID(r, uuid.New())
	r = withURLParam(r, "boardID", boardID.String())
	w := httptest.NewRecorder()

	h.Create(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestColumnHandler_Create_InvalidBody(t *testing.T) {
	svc := &MockColumnService{
		createFn: func(_ context.Context, _, _ uuid.UUID, _ *domain.CreateColumnInput) (*domain.Column, error) {
			return nil, domain.ErrBadRequest
		},
	}

	h := newTestColumnHandler(svc)

	boardID := uuid.New()
	body, _ := json.Marshal(domain.CreateColumnInput{})

	r := httptest.NewRequest(http.MethodPost, "/boards/"+boardID.String()+"/columns", bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	r = withUserID(r, uuid.New())
	r = withURLParam(r, "boardID", boardID.String())
	w := httptest.NewRecorder()

	h.Create(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestColumnHandler_Create_NotFound(t *testing.T) {
	svc := &MockColumnService{
		createFn: func(_ context.Context, _, _ uuid.UUID, _ *domain.CreateColumnInput) (*domain.Column, error) {
			return nil, domain.ErrNotFound
		},
	}

	h := newTestColumnHandler(svc)

	boardID := uuid.New()
	body, _ := json.Marshal(domain.CreateColumnInput{Name: "In Progress", Position: 1})

	r := httptest.NewRequest(http.MethodPost, "/boards/"+boardID.String()+"/columns", bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	r = withUserID(r, uuid.New())
	r = withURLParam(r, "boardID", boardID.String())
	w := httptest.NewRecorder()

	h.Create(w, r)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", w.Code)
	}
}

func TestColumnHandler_Create_Forbidden(t *testing.T) {
	svc := &MockColumnService{
		createFn: func(_ context.Context, _, _ uuid.UUID, _ *domain.CreateColumnInput) (*domain.Column, error) {
			return nil, domain.ErrForbidden
		},
	}

	h := newTestColumnHandler(svc)

	boardID := uuid.New()
	body, _ := json.Marshal(domain.CreateColumnInput{Name: "In Progress", Position: 1})

	r := httptest.NewRequest(http.MethodPost, "/boards/"+boardID.String()+"/columns", bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	r = withUserID(r, uuid.New())
	r = withURLParam(r, "boardID", boardID.String())
	w := httptest.NewRecorder()

	h.Create(w, r)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d", w.Code)
	}
}

func TestColumnHandler_Create_InternalError(t *testing.T) {
	svc := &MockColumnService{
		createFn: func(_ context.Context, _, _ uuid.UUID, _ *domain.CreateColumnInput) (*domain.Column, error) {
			return nil, errors.New("unexpected error")
		},
	}

	h := newTestColumnHandler(svc)

	boardID := uuid.New()
	body, _ := json.Marshal(domain.CreateColumnInput{Name: "In Progress", Position: 1})

	r := httptest.NewRequest(http.MethodPost, "/boards/"+boardID.String()+"/columns", bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	r = withUserID(r, uuid.New())
	r = withURLParam(r, "boardID", boardID.String())
	w := httptest.NewRecorder()

	h.Create(w, r)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
}

func TestColumnHandler_GetAll_Success(t *testing.T) {
	boardID := uuid.New()

	svc := &MockColumnService{
		getAllByBoardIDFn: func(_ context.Context, bid uuid.UUID, _ *domain.ListParams) ([]*domain.Column, error) {
			return []*domain.Column{
				newColumn(bid),
				newColumn(bid),
			}, nil
		},
	}

	h := newTestColumnHandler(svc)

	r := httptest.NewRequest(http.MethodGet, "/boards/"+boardID.String()+"/columns", nil)
	r = withUserID(r, uuid.New())
	r = withURLParam(r, "boardID", boardID.String())
	w := httptest.NewRecorder()

	h.GetAll(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp []*domain.Column
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(resp) != 2 {
		t.Errorf("expected 2 columns, got %d", len(resp))
	}
}

func TestColumnHandler_GetAll_Empty(t *testing.T) {
	svc := &MockColumnService{
		getAllByBoardIDFn: func(_ context.Context, _ uuid.UUID, _ *domain.ListParams) ([]*domain.Column, error) {
			return []*domain.Column{}, nil
		},
	}

	h := newTestColumnHandler(svc)

	boardID := uuid.New()
	r := httptest.NewRequest(http.MethodGet, "/boards/"+boardID.String()+"/columns", nil)
	r = withURLParam(r, "boardID", boardID.String())
	w := httptest.NewRecorder()

	h.GetAll(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
}

func TestColumnHandler_GetAll_InvalidBoardUUID(t *testing.T) {
	h := newTestColumnHandler(&MockColumnService{})

	r := httptest.NewRequest(http.MethodGet, "/boards/not-a-uuid/columns", nil)
	r = withURLParam(r, "boardID", "not-a-uuid")
	w := httptest.NewRecorder()

	h.GetAll(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestColumnHandler_GetAll_InternalError(t *testing.T) {
	svc := &MockColumnService{
		getAllByBoardIDFn: func(_ context.Context, _ uuid.UUID, _ *domain.ListParams) ([]*domain.Column, error) {
			return nil, errors.New("unexpected error")
		},
	}

	h := newTestColumnHandler(svc)

	boardID := uuid.New()
	r := httptest.NewRequest(http.MethodGet, "/boards/"+boardID.String()+"/columns", nil)
	r = withURLParam(r, "boardID", boardID.String())
	w := httptest.NewRecorder()

	h.GetAll(w, r)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
}

func TestColumnHandler_GetByID_Success(t *testing.T) {
	boardID := uuid.New()
	columnID := uuid.New()

	svc := &MockColumnService{
		getByIDFn: func(_ context.Context, id uuid.UUID) (*domain.Column, error) {
			return &domain.Column{
				ID:        id,
				BoardID:   boardID,
				Name:      "test column",
				Position:  1,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}, nil
		},
	}

	h := newTestColumnHandler(svc)

	r := httptest.NewRequest(http.MethodGet, "/columns/"+columnID.String(), nil)
	r = withURLParam(r, "id", columnID.String())
	w := httptest.NewRecorder()

	h.GetByID(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp domain.Column
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.ID != columnID {
		t.Errorf("expected column ID %v, got %v", columnID, resp.ID)
	}
}

func TestColumnHandler_GetByID_InvalidUUID(t *testing.T) {
	h := newTestColumnHandler(&MockColumnService{})

	r := httptest.NewRequest(http.MethodGet, "/columns/not-a-uuid", nil)
	r = withURLParam(r, "id", "not-a-uuid")
	w := httptest.NewRecorder()

	h.GetByID(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestColumnHandler_GetByID_NotFound(t *testing.T) {
	svc := &MockColumnService{
		getByIDFn: func(_ context.Context, _ uuid.UUID) (*domain.Column, error) {
			return nil, domain.ErrNotFound
		},
	}

	h := newTestColumnHandler(svc)

	columnID := uuid.New()
	r := httptest.NewRequest(http.MethodGet, "/columns/"+columnID.String(), nil)
	r = withURLParam(r, "id", columnID.String())
	w := httptest.NewRecorder()

	h.GetByID(w, r)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", w.Code)
	}
}

func TestColumnHandler_GetByID_InternalError(t *testing.T) {
	svc := &MockColumnService{
		getByIDFn: func(_ context.Context, _ uuid.UUID) (*domain.Column, error) {
			return nil, errors.New("unexpected error")
		},
	}

	h := newTestColumnHandler(svc)

	columnID := uuid.New()
	r := httptest.NewRequest(http.MethodGet, "/columns/"+columnID.String(), nil)
	r = withURLParam(r, "id", columnID.String())
	w := httptest.NewRecorder()

	h.GetByID(w, r)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
}

func TestColumnHandler_Update_Success(t *testing.T) {
	svc := &MockColumnService{
		updateFn: func(_ context.Context, _, _ uuid.UUID, _ *domain.UpdateColumnInput) error {
			return nil
		},
	}

	h := newTestColumnHandler(svc)

	columnID := uuid.New()
	body, _ := json.Marshal(domain.UpdateColumnInput{Name: "Done", Position: 2})

	r := httptest.NewRequest(http.MethodPatch, "/columns/"+columnID.String(), bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	r = withUserID(r, uuid.New())
	r = withURLParam(r, "id", columnID.String())
	w := httptest.NewRecorder()

	h.Update(w, r)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", w.Code)
	}
}

func TestColumnHandler_Update_Unauthorized(t *testing.T) {
	h := newTestColumnHandler(&MockColumnService{})

	columnID := uuid.New()
	body, _ := json.Marshal(domain.UpdateColumnInput{Name: "Done", Position: 2})

	r := httptest.NewRequest(http.MethodPatch, "/columns/"+columnID.String(), bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	r = withURLParam(r, "id", columnID.String())
	w := httptest.NewRecorder()

	h.Update(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", w.Code)
	}
}

func TestColumnHandler_Update_InvalidUUID(t *testing.T) {
	h := newTestColumnHandler(&MockColumnService{})

	body, _ := json.Marshal(domain.UpdateColumnInput{Name: "Done", Position: 2})

	r := httptest.NewRequest(http.MethodPatch, "/columns/not-a-uuid", bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	r = withUserID(r, uuid.New())
	r = withURLParam(r, "id", "not-a-uuid")
	w := httptest.NewRecorder()

	h.Update(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestColumnHandler_Update_NotJSON(t *testing.T) {
	h := newTestColumnHandler(&MockColumnService{})

	columnID := uuid.New()

	r := httptest.NewRequest(http.MethodPatch, "/columns/"+columnID.String(), bytes.NewReader([]byte("not-json")))
	r.Header.Set("Content-Type", "application/json")
	r = withUserID(r, uuid.New())
	r = withURLParam(r, "id", columnID.String())
	w := httptest.NewRecorder()

	h.Update(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestColumnHandler_Update_NotFound(t *testing.T) {
	svc := &MockColumnService{
		updateFn: func(_ context.Context, _, _ uuid.UUID, _ *domain.UpdateColumnInput) error {
			return domain.ErrNotFound
		},
	}

	h := newTestColumnHandler(svc)

	columnID := uuid.New()
	body, _ := json.Marshal(domain.UpdateColumnInput{Name: "Done", Position: 2})

	r := httptest.NewRequest(http.MethodPatch, "/columns/"+columnID.String(), bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	r = withUserID(r, uuid.New())
	r = withURLParam(r, "id", columnID.String())
	w := httptest.NewRecorder()

	h.Update(w, r)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", w.Code)
	}
}

func TestColumnHandler_Update_Forbidden(t *testing.T) {
	svc := &MockColumnService{
		updateFn: func(_ context.Context, _, _ uuid.UUID, _ *domain.UpdateColumnInput) error {
			return domain.ErrForbidden
		},
	}

	h := newTestColumnHandler(svc)

	columnID := uuid.New()
	body, _ := json.Marshal(domain.UpdateColumnInput{Name: "Done", Position: 2})

	r := httptest.NewRequest(http.MethodPatch, "/columns/"+columnID.String(), bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	r = withUserID(r, uuid.New())
	r = withURLParam(r, "id", columnID.String())
	w := httptest.NewRecorder()

	h.Update(w, r)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d", w.Code)
	}
}

func TestColumnHandler_Update_InternalError(t *testing.T) {
	svc := &MockColumnService{
		updateFn: func(_ context.Context, _, _ uuid.UUID, _ *domain.UpdateColumnInput) error {
			return errors.New("unexpected error")
		},
	}

	h := newTestColumnHandler(svc)

	columnID := uuid.New()
	body, _ := json.Marshal(domain.UpdateColumnInput{Name: "Done", Position: 2})

	r := httptest.NewRequest(http.MethodPatch, "/columns/"+columnID.String(), bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	r = withUserID(r, uuid.New())
	r = withURLParam(r, "id", columnID.String())
	w := httptest.NewRecorder()

	h.Update(w, r)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
}

func TestColumnHandler_Delete_Success(t *testing.T) {
	svc := &MockColumnService{
		deleteFn: func(_ context.Context, _, _ uuid.UUID) error {
			return nil
		},
	}

	h := newTestColumnHandler(svc)

	columnID := uuid.New()

	r := httptest.NewRequest(http.MethodDelete, "/columns/"+columnID.String(), nil)
	r = withUserID(r, uuid.New())
	r = withURLParam(r, "id", columnID.String())
	w := httptest.NewRecorder()

	h.Delete(w, r)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", w.Code)
	}
}

func TestColumnHandler_Delete_Unauthorized(t *testing.T) {
	h := newTestColumnHandler(&MockColumnService{})

	columnID := uuid.New()

	r := httptest.NewRequest(http.MethodDelete, "/columns/"+columnID.String(), nil)
	r = withURLParam(r, "id", columnID.String())
	w := httptest.NewRecorder()

	h.Delete(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", w.Code)
	}
}

func TestColumnHandler_Delete_InvalidUUID(t *testing.T) {
	h := newTestColumnHandler(&MockColumnService{})

	r := httptest.NewRequest(http.MethodDelete, "/columns/not-a-uuid", nil)
	r = withUserID(r, uuid.New())
	r = withURLParam(r, "id", "not-a-uuid")
	w := httptest.NewRecorder()

	h.Delete(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestColumnHandler_Delete_NotFound(t *testing.T) {
	svc := &MockColumnService{
		deleteFn: func(_ context.Context, _, _ uuid.UUID) error {
			return domain.ErrNotFound
		},
	}

	h := newTestColumnHandler(svc)

	columnID := uuid.New()

	r := httptest.NewRequest(http.MethodDelete, "/columns/"+columnID.String(), nil)
	r = withUserID(r, uuid.New())
	r = withURLParam(r, "id", columnID.String())
	w := httptest.NewRecorder()

	h.Delete(w, r)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", w.Code)
	}
}

func TestColumnHandler_Delete_Forbidden(t *testing.T) {
	svc := &MockColumnService{
		deleteFn: func(_ context.Context, _, _ uuid.UUID) error {
			return domain.ErrForbidden
		},
	}

	h := newTestColumnHandler(svc)

	columnID := uuid.New()

	r := httptest.NewRequest(http.MethodDelete, "/columns/"+columnID.String(), nil)
	r = withUserID(r, uuid.New())
	r = withURLParam(r, "id", columnID.String())
	w := httptest.NewRecorder()

	h.Delete(w, r)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d", w.Code)
	}
}

func TestColumnHandler_Delete_InternalError(t *testing.T) {
	svc := &MockColumnService{
		deleteFn: func(_ context.Context, _, _ uuid.UUID) error {
			return errors.New("unexpected error")
		},
	}

	h := newTestColumnHandler(svc)

	columnID := uuid.New()

	r := httptest.NewRequest(http.MethodDelete, "/columns/"+columnID.String(), nil)
	r = withUserID(r, uuid.New())
	r = withURLParam(r, "id", columnID.String())
	w := httptest.NewRecorder()

	h.Delete(w, r)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
}
