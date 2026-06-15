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

type MockTaskService struct {
	createFn             func(ctx context.Context, userID, columnID uuid.UUID, input *domain.CreateTaskInput) (*domain.Task, error)
	getByIDFn            func(ctx context.Context, taskID uuid.UUID) (*domain.Task, error)
	getAllByUserIDFn     func(ctx context.Context, userID uuid.UUID, params *domain.ListParams) ([]*domain.Task, error)
	getAllByColumnIDFn   func(ctx context.Context, columnID uuid.UUID, params *domain.ListParams) ([]*domain.Task, error)
	getAllByAssigneeIDFn func(ctx context.Context, assigneeID uuid.UUID, params *domain.ListParams) ([]*domain.Task, error)
	updateFn             func(ctx context.Context, taskID, userID uuid.UUID, input *domain.UpdateTaskInput) error
	deleteFn             func(ctx context.Context, taskID, userID uuid.UUID) error
}

func (m *MockTaskService) Create(ctx context.Context, userID, columnID uuid.UUID, input *domain.CreateTaskInput) (*domain.Task, error) {
	if m.createFn != nil {
		return m.createFn(ctx, userID, columnID, input)
	}
	return nil, nil
}

func (m *MockTaskService) GetByID(ctx context.Context, taskID uuid.UUID) (*domain.Task, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, taskID)
	}
	return nil, nil
}

func (m *MockTaskService) GetAllByUserID(ctx context.Context, userID uuid.UUID, params *domain.ListParams) ([]*domain.Task, error) {
	if m.getAllByUserIDFn != nil {
		return m.getAllByUserIDFn(ctx, userID, params)
	}
	return []*domain.Task{}, nil
}

func (m *MockTaskService) GetAllByColumnID(ctx context.Context, columnID uuid.UUID, params *domain.ListParams) ([]*domain.Task, error) {
	if m.getAllByColumnIDFn != nil {
		return m.getAllByColumnIDFn(ctx, columnID, params)
	}
	return []*domain.Task{}, nil
}

func (m *MockTaskService) GetAllByAssigneeID(ctx context.Context, assigneeID uuid.UUID, params *domain.ListParams) ([]*domain.Task, error) {
	if m.getAllByAssigneeIDFn != nil {
		return m.getAllByAssigneeIDFn(ctx, assigneeID, params)
	}
	return []*domain.Task{}, nil
}

func (m *MockTaskService) Update(ctx context.Context, taskID, userID uuid.UUID, input *domain.UpdateTaskInput) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, taskID, userID, input)
	}
	return nil
}

func (m *MockTaskService) Delete(ctx context.Context, taskID, userID uuid.UUID) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, taskID, userID)
	}
	return nil
}

func newTestTaskHandler(svc domain.TaskService) *TaskHandler {
	logger, _ := zap.NewDevelopment()
	return NewTaskHandler(svc, createAuth(), logger.Sugar())
}

func newTask(userID, columnID uuid.UUID) *domain.Task {
	return &domain.Task{
		ID:        uuid.New(),
		OwnerID:   userID,
		ColumnID:  columnID,
		Name:      "test task",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func TestTaskHandler_Create_Success(t *testing.T) {
	userID := uuid.New()
	columnID := uuid.New()

	svc := &MockTaskService{
		createFn: func(_ context.Context, uid, cid uuid.UUID, input *domain.CreateTaskInput) (*domain.Task, error) {
			return &domain.Task{
				ID:        uuid.New(),
				OwnerID:   uid,
				ColumnID:  cid,
				Name:      input.Name,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}, nil
		},
	}

	h := newTestTaskHandler(svc)

	body, _ := json.Marshal(domain.CreateTaskInput{Name: "Fix login bug"})

	r := httptest.NewRequest(http.MethodPost, "/columns/"+columnID.String()+"/tasks", bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	r = withUserID(r, userID)
	r = withURLParam(r, "columnID", columnID.String())
	w := httptest.NewRecorder()

	h.Create(w, r)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", w.Code)
	}

	var resp domain.Task
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Name != "Fix login bug" {
		t.Errorf("expected name 'Fix login bug', got %v", resp.Name)
	}
	if resp.OwnerID != userID {
		t.Errorf("expected ownerID %v, got %v", userID, resp.OwnerID)
	}
	if resp.ColumnID != columnID {
		t.Errorf("expected columnID %v, got %v", columnID, resp.ColumnID)
	}
}

func TestTaskHandler_Create_Unauthorized(t *testing.T) {
	h := newTestTaskHandler(&MockTaskService{})

	columnID := uuid.New()
	body, _ := json.Marshal(domain.CreateTaskInput{Name: "Fix login bug"})

	r := httptest.NewRequest(http.MethodPost, "/columns/"+columnID.String()+"/tasks", bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	r = withURLParam(r, "columnID", columnID.String())
	w := httptest.NewRecorder()

	h.Create(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", w.Code)
	}
}

func TestTaskHandler_Create_InvalidColumnUUID(t *testing.T) {
	h := newTestTaskHandler(&MockTaskService{})

	body, _ := json.Marshal(domain.CreateTaskInput{Name: "Fix login bug"})

	r := httptest.NewRequest(http.MethodPost, "/columns/not-a-uuid/tasks", bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	r = withUserID(r, uuid.New())
	r = withURLParam(r, "columnID", "not-a-uuid")
	w := httptest.NewRecorder()

	h.Create(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestTaskHandler_Create_NotJSON(t *testing.T) {
	h := newTestTaskHandler(&MockTaskService{})

	columnID := uuid.New()

	r := httptest.NewRequest(http.MethodPost, "/columns/"+columnID.String()+"/tasks", bytes.NewReader([]byte("not-json")))
	r.Header.Set("Content-Type", "application/json")
	r = withUserID(r, uuid.New())
	r = withURLParam(r, "columnID", columnID.String())
	w := httptest.NewRecorder()

	h.Create(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestTaskHandler_Create_InvalidBody(t *testing.T) {
	svc := &MockTaskService{
		createFn: func(_ context.Context, _, _ uuid.UUID, _ *domain.CreateTaskInput) (*domain.Task, error) {
			return nil, domain.ErrBadRequest
		},
	}

	h := newTestTaskHandler(svc)

	columnID := uuid.New()
	body, _ := json.Marshal(domain.CreateTaskInput{})

	r := httptest.NewRequest(http.MethodPost, "/columns/"+columnID.String()+"/tasks", bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	r = withUserID(r, uuid.New())
	r = withURLParam(r, "columnID", columnID.String())
	w := httptest.NewRecorder()

	h.Create(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestTaskHandler_Create_NotFound(t *testing.T) {
	svc := &MockTaskService{
		createFn: func(_ context.Context, _, _ uuid.UUID, _ *domain.CreateTaskInput) (*domain.Task, error) {
			return nil, domain.ErrNotFound
		},
	}

	h := newTestTaskHandler(svc)

	columnID := uuid.New()
	body, _ := json.Marshal(domain.CreateTaskInput{Name: "Fix login bug"})

	r := httptest.NewRequest(http.MethodPost, "/columns/"+columnID.String()+"/tasks", bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	r = withUserID(r, uuid.New())
	r = withURLParam(r, "columnID", columnID.String())
	w := httptest.NewRecorder()

	h.Create(w, r)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", w.Code)
	}
}

func TestTaskHandler_Create_Forbidden(t *testing.T) {
	svc := &MockTaskService{
		createFn: func(_ context.Context, _, _ uuid.UUID, _ *domain.CreateTaskInput) (*domain.Task, error) {
			return nil, domain.ErrForbidden
		},
	}

	h := newTestTaskHandler(svc)

	columnID := uuid.New()
	body, _ := json.Marshal(domain.CreateTaskInput{Name: "Fix login bug"})

	r := httptest.NewRequest(http.MethodPost, "/columns/"+columnID.String()+"/tasks", bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	r = withUserID(r, uuid.New())
	r = withURLParam(r, "columnID", columnID.String())
	w := httptest.NewRecorder()

	h.Create(w, r)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d", w.Code)
	}
}

func TestTaskHandler_Create_InternalError(t *testing.T) {
	svc := &MockTaskService{
		createFn: func(_ context.Context, _, _ uuid.UUID, _ *domain.CreateTaskInput) (*domain.Task, error) {
			return nil, errors.New("unexpected error")
		},
	}

	h := newTestTaskHandler(svc)

	columnID := uuid.New()
	body, _ := json.Marshal(domain.CreateTaskInput{Name: "Fix login bug"})

	r := httptest.NewRequest(http.MethodPost, "/columns/"+columnID.String()+"/tasks", bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	r = withUserID(r, uuid.New())
	r = withURLParam(r, "columnID", columnID.String())
	w := httptest.NewRecorder()

	h.Create(w, r)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
}

func TestTaskHandler_GetAllByColumnID_Success(t *testing.T) {
	userID := uuid.New()
	columnID := uuid.New()

	svc := &MockTaskService{
		getAllByColumnIDFn: func(_ context.Context, cid uuid.UUID, _ *domain.ListParams) ([]*domain.Task, error) {
			return []*domain.Task{
				newTask(userID, cid),
				newTask(userID, cid),
			}, nil
		},
	}

	h := newTestTaskHandler(svc)

	r := httptest.NewRequest(http.MethodGet, "/columns/"+columnID.String()+"/tasks", nil)
	r = withURLParam(r, "columnID", columnID.String())
	w := httptest.NewRecorder()

	h.GetAllByColumnID(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp []*domain.Task
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(resp) != 2 {
		t.Errorf("expected 2 tasks, got %d", len(resp))
	}
}

func TestTaskHandler_GetAllByColumnID_Empty(t *testing.T) {
	svc := &MockTaskService{
		getAllByColumnIDFn: func(_ context.Context, _ uuid.UUID, _ *domain.ListParams) ([]*domain.Task, error) {
			return []*domain.Task{}, nil
		},
	}

	h := newTestTaskHandler(svc)

	columnID := uuid.New()
	r := httptest.NewRequest(http.MethodGet, "/columns/"+columnID.String()+"/tasks", nil)
	r = withURLParam(r, "columnID", columnID.String())
	w := httptest.NewRecorder()

	h.GetAllByColumnID(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
}

func TestTaskHandler_GetAllByColumnID_InvalidUUID(t *testing.T) {
	h := newTestTaskHandler(&MockTaskService{})

	r := httptest.NewRequest(http.MethodGet, "/columns/not-a-uuid/tasks", nil)
	r = withURLParam(r, "columnID", "not-a-uuid")
	w := httptest.NewRecorder()

	h.GetAllByColumnID(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestTaskHandler_GetAllByColumnID_InternalError(t *testing.T) {
	svc := &MockTaskService{
		getAllByColumnIDFn: func(_ context.Context, _ uuid.UUID, _ *domain.ListParams) ([]*domain.Task, error) {
			return nil, errors.New("unexpected error")
		},
	}

	h := newTestTaskHandler(svc)

	columnID := uuid.New()
	r := httptest.NewRequest(http.MethodGet, "/columns/"+columnID.String()+"/tasks", nil)
	r = withURLParam(r, "columnID", columnID.String())
	w := httptest.NewRecorder()

	h.GetAllByColumnID(w, r)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
}

func TestTaskHandler_GetAllByUserID_Success(t *testing.T) {
	userID := uuid.New()
	columnID := uuid.New()

	svc := &MockTaskService{
		getAllByUserIDFn: func(_ context.Context, uid uuid.UUID, _ *domain.ListParams) ([]*domain.Task, error) {
			return []*domain.Task{
				newTask(uid, columnID),
				newTask(uid, columnID),
			}, nil
		},
	}

	h := newTestTaskHandler(svc)

	r := httptest.NewRequest(http.MethodGet, "/users/me/tasks", nil)
	r = withUserID(r, userID)
	w := httptest.NewRecorder()

	h.GetAllByUserID(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp []*domain.Task
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(resp) != 2 {
		t.Errorf("expected 2 tasks, got %d", len(resp))
	}
}

func TestTaskHandler_GetAllByUserID_Unauthorized(t *testing.T) {
	h := newTestTaskHandler(&MockTaskService{})

	r := httptest.NewRequest(http.MethodGet, "/users/me/tasks", nil)
	w := httptest.NewRecorder()

	h.GetAllByUserID(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", w.Code)
	}
}

func TestTaskHandler_GetAllByUserID_InternalError(t *testing.T) {
	svc := &MockTaskService{
		getAllByUserIDFn: func(_ context.Context, _ uuid.UUID, _ *domain.ListParams) ([]*domain.Task, error) {
			return nil, errors.New("unexpected error")
		},
	}

	h := newTestTaskHandler(svc)

	r := httptest.NewRequest(http.MethodGet, "/users/me/tasks", nil)
	r = withUserID(r, uuid.New())
	w := httptest.NewRecorder()

	h.GetAllByUserID(w, r)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
}

func TestTaskHandler_GetAllByAssigneeID_Success(t *testing.T) {
	assigneeID := uuid.New()
	columnID := uuid.New()

	svc := &MockTaskService{
		getAllByAssigneeIDFn: func(_ context.Context, aid uuid.UUID, _ *domain.ListParams) ([]*domain.Task, error) {
			task := newTask(uuid.New(), columnID)
			task.AssigneeID = &aid
			return []*domain.Task{task}, nil
		},
	}

	h := newTestTaskHandler(svc)

	r := httptest.NewRequest(http.MethodGet, "/users/me/tasks/assigned", nil)
	r = withUserID(r, assigneeID)
	w := httptest.NewRecorder()

	h.GetAllByAssigneeID(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp []*domain.Task
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(resp) != 1 {
		t.Errorf("expected 1 task, got %d", len(resp))
	}
	if *resp[0].AssigneeID != assigneeID {
		t.Errorf("expected assigneeID %v, got %v", assigneeID, resp[0].AssigneeID)
	}
}

func TestTaskHandler_GetAllByAssigneeID_Unauthorized(t *testing.T) {
	h := newTestTaskHandler(&MockTaskService{})

	r := httptest.NewRequest(http.MethodGet, "/users/me/tasks/assigned", nil)
	w := httptest.NewRecorder()

	h.GetAllByAssigneeID(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", w.Code)
	}
}

func TestTaskHandler_GetAllByAssigneeID_InternalError(t *testing.T) {
	svc := &MockTaskService{
		getAllByAssigneeIDFn: func(_ context.Context, _ uuid.UUID, _ *domain.ListParams) ([]*domain.Task, error) {
			return nil, errors.New("unexpected error")
		},
	}

	h := newTestTaskHandler(svc)

	r := httptest.NewRequest(http.MethodGet, "/users/me/tasks/assigned", nil)
	r = withUserID(r, uuid.New())
	w := httptest.NewRecorder()

	h.GetAllByAssigneeID(w, r)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
}

func TestTaskHandler_GetByID_Success(t *testing.T) {
	userID := uuid.New()
	columnID := uuid.New()
	taskID := uuid.New()

	svc := &MockTaskService{
		getByIDFn: func(_ context.Context, id uuid.UUID) (*domain.Task, error) {
			return &domain.Task{
				ID:        id,
				OwnerID:   userID,
				ColumnID:  columnID,
				Name:      "test task",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}, nil
		},
	}

	h := newTestTaskHandler(svc)

	r := httptest.NewRequest(http.MethodGet, "/tasks/"+taskID.String(), nil)
	r = withURLParam(r, "id", taskID.String())
	w := httptest.NewRecorder()

	h.GetByID(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp domain.Task
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.ID != taskID {
		t.Errorf("expected task ID %v, got %v", taskID, resp.ID)
	}
}

func TestTaskHandler_GetByID_InvalidUUID(t *testing.T) {
	h := newTestTaskHandler(&MockTaskService{})

	r := httptest.NewRequest(http.MethodGet, "/tasks/not-a-uuid", nil)
	r = withURLParam(r, "id", "not-a-uuid")
	w := httptest.NewRecorder()

	h.GetByID(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestTaskHandler_GetByID_NotFound(t *testing.T) {
	svc := &MockTaskService{
		getByIDFn: func(_ context.Context, _ uuid.UUID) (*domain.Task, error) {
			return nil, domain.ErrNotFound
		},
	}

	h := newTestTaskHandler(svc)

	taskID := uuid.New()
	r := httptest.NewRequest(http.MethodGet, "/tasks/"+taskID.String(), nil)
	r = withURLParam(r, "id", taskID.String())
	w := httptest.NewRecorder()

	h.GetByID(w, r)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", w.Code)
	}
}

func TestTaskHandler_GetByID_InternalError(t *testing.T) {
	svc := &MockTaskService{
		getByIDFn: func(_ context.Context, _ uuid.UUID) (*domain.Task, error) {
			return nil, errors.New("unexpected error")
		},
	}

	h := newTestTaskHandler(svc)

	taskID := uuid.New()
	r := httptest.NewRequest(http.MethodGet, "/tasks/"+taskID.String(), nil)
	r = withURLParam(r, "id", taskID.String())
	w := httptest.NewRecorder()

	h.GetByID(w, r)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
}

func TestTaskHandler_Update_Success(t *testing.T) {
	svc := &MockTaskService{
		updateFn: func(_ context.Context, _, _ uuid.UUID, _ *domain.UpdateTaskInput) error {
			return nil
		},
	}

	h := newTestTaskHandler(svc)

	taskID := uuid.New()
	body, _ := json.Marshal(domain.UpdateTaskInput{
		ColumnID: uuid.New(),
		Name:     "Updated task",
	})

	r := httptest.NewRequest(http.MethodPatch, "/tasks/"+taskID.String(), bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	r = withUserID(r, uuid.New())
	r = withURLParam(r, "id", taskID.String())
	w := httptest.NewRecorder()

	h.Update(w, r)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", w.Code)
	}
}

func TestTaskHandler_Update_Unauthorized(t *testing.T) {
	h := newTestTaskHandler(&MockTaskService{})

	taskID := uuid.New()
	body, _ := json.Marshal(domain.UpdateTaskInput{
		ColumnID: uuid.New(),
		Name:     "Updated task",
	})

	r := httptest.NewRequest(http.MethodPatch, "/tasks/"+taskID.String(), bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	r = withURLParam(r, "id", taskID.String())
	w := httptest.NewRecorder()

	h.Update(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", w.Code)
	}
}

func TestTaskHandler_Update_InvalidUUID(t *testing.T) {
	h := newTestTaskHandler(&MockTaskService{})

	body, _ := json.Marshal(domain.UpdateTaskInput{
		ColumnID: uuid.New(),
		Name:     "Updated task",
	})

	r := httptest.NewRequest(http.MethodPatch, "/tasks/not-a-uuid", bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	r = withUserID(r, uuid.New())
	r = withURLParam(r, "id", "not-a-uuid")
	w := httptest.NewRecorder()

	h.Update(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestTaskHandler_Update_NotJSON(t *testing.T) {
	h := newTestTaskHandler(&MockTaskService{})

	taskID := uuid.New()

	r := httptest.NewRequest(http.MethodPatch, "/tasks/"+taskID.String(), bytes.NewReader([]byte("not-json")))
	r.Header.Set("Content-Type", "application/json")
	r = withUserID(r, uuid.New())
	r = withURLParam(r, "id", taskID.String())
	w := httptest.NewRecorder()

	h.Update(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestTaskHandler_Update_NotFound(t *testing.T) {
	svc := &MockTaskService{
		updateFn: func(_ context.Context, _, _ uuid.UUID, _ *domain.UpdateTaskInput) error {
			return domain.ErrNotFound
		},
	}

	h := newTestTaskHandler(svc)

	taskID := uuid.New()
	body, _ := json.Marshal(domain.UpdateTaskInput{
		ColumnID: uuid.New(),
		Name:     "Updated task",
	})

	r := httptest.NewRequest(http.MethodPatch, "/tasks/"+taskID.String(), bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	r = withUserID(r, uuid.New())
	r = withURLParam(r, "id", taskID.String())
	w := httptest.NewRecorder()

	h.Update(w, r)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", w.Code)
	}
}

func TestTaskHandler_Update_Forbidden(t *testing.T) {
	svc := &MockTaskService{
		updateFn: func(_ context.Context, _, _ uuid.UUID, _ *domain.UpdateTaskInput) error {
			return domain.ErrForbidden
		},
	}

	h := newTestTaskHandler(svc)

	taskID := uuid.New()
	body, _ := json.Marshal(domain.UpdateTaskInput{
		ColumnID: uuid.New(),
		Name:     "Updated task",
	})

	r := httptest.NewRequest(http.MethodPatch, "/tasks/"+taskID.String(), bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	r = withUserID(r, uuid.New())
	r = withURLParam(r, "id", taskID.String())
	w := httptest.NewRecorder()

	h.Update(w, r)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d", w.Code)
	}
}

func TestTaskHandler_Update_InternalError(t *testing.T) {
	svc := &MockTaskService{
		updateFn: func(_ context.Context, _, _ uuid.UUID, _ *domain.UpdateTaskInput) error {
			return errors.New("unexpected error")
		},
	}

	h := newTestTaskHandler(svc)

	taskID := uuid.New()
	body, _ := json.Marshal(domain.UpdateTaskInput{
		ColumnID: uuid.New(),
		Name:     "Updated task",
	})

	r := httptest.NewRequest(http.MethodPatch, "/tasks/"+taskID.String(), bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	r = withUserID(r, uuid.New())
	r = withURLParam(r, "id", taskID.String())
	w := httptest.NewRecorder()

	h.Update(w, r)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
}

func TestTaskHandler_Delete_Success(t *testing.T) {
	svc := &MockTaskService{
		deleteFn: func(_ context.Context, _, _ uuid.UUID) error {
			return nil
		},
	}

	h := newTestTaskHandler(svc)

	taskID := uuid.New()

	r := httptest.NewRequest(http.MethodDelete, "/tasks/"+taskID.String(), nil)
	r = withUserID(r, uuid.New())
	r = withURLParam(r, "id", taskID.String())
	w := httptest.NewRecorder()

	h.Delete(w, r)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", w.Code)
	}
}

func TestTaskHandler_Delete_Unauthorized(t *testing.T) {
	h := newTestTaskHandler(&MockTaskService{})

	taskID := uuid.New()

	r := httptest.NewRequest(http.MethodDelete, "/tasks/"+taskID.String(), nil)
	r = withURLParam(r, "id", taskID.String())
	w := httptest.NewRecorder()

	h.Delete(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", w.Code)
	}
}

func TestTaskHandler_Delete_InvalidUUID(t *testing.T) {
	h := newTestTaskHandler(&MockTaskService{})

	r := httptest.NewRequest(http.MethodDelete, "/tasks/not-a-uuid", nil)
	r = withUserID(r, uuid.New())
	r = withURLParam(r, "id", "not-a-uuid")
	w := httptest.NewRecorder()

	h.Delete(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestTaskHandler_Delete_NotFound(t *testing.T) {
	svc := &MockTaskService{
		deleteFn: func(_ context.Context, _, _ uuid.UUID) error {
			return domain.ErrNotFound
		},
	}

	h := newTestTaskHandler(svc)

	taskID := uuid.New()

	r := httptest.NewRequest(http.MethodDelete, "/tasks/"+taskID.String(), nil)
	r = withUserID(r, uuid.New())
	r = withURLParam(r, "id", taskID.String())
	w := httptest.NewRecorder()

	h.Delete(w, r)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", w.Code)
	}
}

func TestTaskHandler_Delete_Forbidden(t *testing.T) {
	svc := &MockTaskService{
		deleteFn: func(_ context.Context, _, _ uuid.UUID) error {
			return domain.ErrForbidden
		},
	}

	h := newTestTaskHandler(svc)

	taskID := uuid.New()

	r := httptest.NewRequest(http.MethodDelete, "/tasks/"+taskID.String(), nil)
	r = withUserID(r, uuid.New())
	r = withURLParam(r, "id", taskID.String())
	w := httptest.NewRecorder()

	h.Delete(w, r)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d", w.Code)
	}
}

func TestTaskHandler_Delete_InternalError(t *testing.T) {
	svc := &MockTaskService{
		deleteFn: func(_ context.Context, _, _ uuid.UUID) error {
			return errors.New("unexpected error")
		},
	}

	h := newTestTaskHandler(svc)

	taskID := uuid.New()

	r := httptest.NewRequest(http.MethodDelete, "/tasks/"+taskID.String(), nil)
	r = withUserID(r, uuid.New())
	r = withURLParam(r, "id", taskID.String())
	w := httptest.NewRecorder()

	h.Delete(w, r)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
}
