package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mykytakuzminov/task-manager-api/internal/domain"
)

type MockTaskRepository struct {
	createFn           func(ctx context.Context, task *domain.Task) error
	getByIDFn          func(ctx context.Context, id uuid.UUID) (*domain.Task, error)
	getOwnerIDFn       func(ctx context.Context, id uuid.UUID) (uuid.UUID, error)
	getAllByUserIDFn   func(ctx context.Context, userID uuid.UUID) ([]*domain.Task, error)
	getAllByColumnIDFn func(ctx context.Context, columnID uuid.UUID) ([]*domain.Task, error)
	getAllByAssigneeFn func(ctx context.Context, assigneeID uuid.UUID) ([]*domain.Task, error)
	updateFn           func(ctx context.Context, task *domain.Task) error
	deleteFn           func(ctx context.Context, id uuid.UUID) error
}

func (m *MockTaskRepository) Create(ctx context.Context, task *domain.Task) error {
	if m.createFn != nil {
		return m.createFn(ctx, task)
	}
	return nil
}

func (m *MockTaskRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Task, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *MockTaskRepository) GetOwnerID(ctx context.Context, id uuid.UUID) (uuid.UUID, error) {
	if m.getOwnerIDFn != nil {
		return m.getOwnerIDFn(ctx, id)
	}
	return uuid.Nil, nil
}

func (m *MockTaskRepository) GetAllByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Task, error) {
	if m.getAllByUserIDFn != nil {
		return m.getAllByUserIDFn(ctx, userID)
	}
	return []*domain.Task{}, nil
}

func (m *MockTaskRepository) GetAllByColumnID(ctx context.Context, columnID uuid.UUID) ([]*domain.Task, error) {
	if m.getAllByColumnIDFn != nil {
		return m.getAllByColumnIDFn(ctx, columnID)
	}
	return []*domain.Task{}, nil
}

func (m *MockTaskRepository) GetAllByAssigneeID(ctx context.Context, assigneeID uuid.UUID) ([]*domain.Task, error) {
	if m.getAllByAssigneeFn != nil {
		return m.getAllByAssigneeFn(ctx, assigneeID)
	}
	return []*domain.Task{}, nil
}

func (m *MockTaskRepository) Update(ctx context.Context, task *domain.Task) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, task)
	}
	return nil
}

func (m *MockTaskRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
}

func TestTaskService_Create(t *testing.T) {
	userID := uuid.New()
	columnID := uuid.New()

	columnRepo := &MockColumnRepository{
		getOwnerIDFn: func(_ context.Context, _ uuid.UUID) (uuid.UUID, error) {
			return userID, nil
		},
	}
	taskRepo := &MockTaskRepository{
		createFn: func(_ context.Context, _ *domain.Task) error {
			return nil
		},
	}

	svc := NewTaskService(taskRepo, columnRepo)

	task, err := svc.Create(context.Background(), userID, columnID, &domain.CreateTaskInput{
		Name: "Fix login bug",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if task.Name != "Fix login bug" {
		t.Errorf("expected task name %v, got %v", "Fix login bug", task.Name)
	}
	if task.ColumnID != columnID {
		t.Errorf("expected ColumnID %v, got %v", columnID, task.ColumnID)
	}
	if task.OwnerID != userID {
		t.Errorf("expected OwnerID %v, got %v", userID, task.OwnerID)
	}
}

func TestTaskService_Create_WithOptionalFields(t *testing.T) {
	userID := uuid.New()
	columnID := uuid.New()
	assigneeID := uuid.New()
	desc := "description"
	dueDate := time.Now().Add(24 * time.Hour)

	columnRepo := &MockColumnRepository{
		getOwnerIDFn: func(_ context.Context, _ uuid.UUID) (uuid.UUID, error) {
			return userID, nil
		},
	}
	taskRepo := &MockTaskRepository{
		createFn: func(_ context.Context, _ *domain.Task) error {
			return nil
		},
	}

	svc := NewTaskService(taskRepo, columnRepo)

	task, err := svc.Create(context.Background(), userID, columnID, &domain.CreateTaskInput{
		Name:        "Fix login bug",
		AssigneeID:  &assigneeID,
		Description: &desc,
		DueDate:     &dueDate,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if task.AssigneeID == nil || *task.AssigneeID != assigneeID {
		t.Errorf("expected AssigneeID %v, got %v", assigneeID, task.AssigneeID)
	}
	if task.Description == nil || *task.Description != desc {
		t.Errorf("expected Description %v, got %v", desc, task.Description)
	}
	if task.DueDate == nil || !task.DueDate.Equal(dueDate) {
		t.Errorf("expected DueDate %v, got %v", dueDate, task.DueDate)
	}
}

func TestTaskService_Create_InvalidBodyRequest(t *testing.T) {
	svc := NewTaskService(&MockTaskRepository{}, &MockColumnRepository{})

	_, err := svc.Create(context.Background(), uuid.New(), uuid.New(), &domain.CreateTaskInput{
		Name: strings.Repeat("a", 101),
	})
	if !errors.Is(err, domain.ErrBadRequest) {
		t.Fatalf("expected ErrBadRequest, got %v", err)
	}
}

func TestTaskService_Create_EmptyName(t *testing.T) {
	svc := NewTaskService(&MockTaskRepository{}, &MockColumnRepository{})

	_, err := svc.Create(context.Background(), uuid.New(), uuid.New(), &domain.CreateTaskInput{
		Name: "",
	})
	if !errors.Is(err, domain.ErrBadRequest) {
		t.Fatalf("expected ErrBadRequest, got %v", err)
	}
}

func TestTaskService_Create_NotFound(t *testing.T) {
	columnRepo := &MockColumnRepository{
		getOwnerIDFn: func(_ context.Context, _ uuid.UUID) (uuid.UUID, error) {
			return uuid.Nil, domain.ErrNotFound
		},
	}

	svc := NewTaskService(&MockTaskRepository{}, columnRepo)

	_, err := svc.Create(context.Background(), uuid.New(), uuid.New(), &domain.CreateTaskInput{
		Name: "Fix login bug",
	})
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestTaskService_Create_AccessDenied(t *testing.T) {
	columnRepo := &MockColumnRepository{
		getOwnerIDFn: func(_ context.Context, _ uuid.UUID) (uuid.UUID, error) {
			return uuid.New(), nil
		},
	}

	svc := NewTaskService(&MockTaskRepository{}, columnRepo)

	_, err := svc.Create(context.Background(), uuid.New(), uuid.New(), &domain.CreateTaskInput{
		Name: "Fix login bug",
	})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestTaskService_Update(t *testing.T) {
	userID := uuid.New()
	taskID := uuid.New()
	columnID := uuid.New()

	taskRepo := &MockTaskRepository{
		getOwnerIDFn: func(_ context.Context, _ uuid.UUID) (uuid.UUID, error) {
			return userID, nil
		},
		updateFn: func(_ context.Context, _ *domain.Task) error {
			return nil
		},
	}

	svc := NewTaskService(taskRepo, &MockColumnRepository{})

	err := svc.Update(context.Background(), taskID, userID, &domain.UpdateTaskInput{
		ColumnID: columnID,
		Name:     "Updated task",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestTaskService_Update_InvalidBodyRequest(t *testing.T) {
	svc := NewTaskService(&MockTaskRepository{}, &MockColumnRepository{})

	err := svc.Update(context.Background(), uuid.New(), uuid.New(), &domain.UpdateTaskInput{
		Name: strings.Repeat("a", 101),
	})
	if !errors.Is(err, domain.ErrBadRequest) {
		t.Fatalf("expected ErrBadRequest, got %v", err)
	}
}

func TestTaskService_Update_NotFound(t *testing.T) {
	taskRepo := &MockTaskRepository{
		getOwnerIDFn: func(_ context.Context, _ uuid.UUID) (uuid.UUID, error) {
			return uuid.Nil, domain.ErrNotFound
		},
	}

	svc := NewTaskService(taskRepo, &MockColumnRepository{})

	err := svc.Update(context.Background(), uuid.New(), uuid.New(), &domain.UpdateTaskInput{
		ColumnID: uuid.New(),
		Name:     "Updated task",
	})
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestTaskService_Update_AccessDenied(t *testing.T) {
	taskRepo := &MockTaskRepository{
		getOwnerIDFn: func(_ context.Context, _ uuid.UUID) (uuid.UUID, error) {
			return uuid.New(), nil
		},
	}

	svc := NewTaskService(taskRepo, &MockColumnRepository{})

	err := svc.Update(context.Background(), uuid.New(), uuid.New(), &domain.UpdateTaskInput{
		ColumnID: uuid.New(),
		Name:     "Updated task",
	})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestTaskService_Delete(t *testing.T) {
	userID := uuid.New()

	taskRepo := &MockTaskRepository{
		getOwnerIDFn: func(_ context.Context, _ uuid.UUID) (uuid.UUID, error) {
			return userID, nil
		},
		deleteFn: func(_ context.Context, _ uuid.UUID) error {
			return nil
		},
	}

	svc := NewTaskService(taskRepo, &MockColumnRepository{})

	err := svc.Delete(context.Background(), uuid.New(), userID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestTaskService_Delete_NotFound(t *testing.T) {
	taskRepo := &MockTaskRepository{
		getOwnerIDFn: func(_ context.Context, _ uuid.UUID) (uuid.UUID, error) {
			return uuid.Nil, domain.ErrNotFound
		},
	}

	svc := NewTaskService(taskRepo, &MockColumnRepository{})

	err := svc.Delete(context.Background(), uuid.New(), uuid.New())
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestTaskService_Delete_AccessDenied(t *testing.T) {
	taskRepo := &MockTaskRepository{
		getOwnerIDFn: func(_ context.Context, _ uuid.UUID) (uuid.UUID, error) {
			return uuid.New(), nil
		},
	}

	svc := NewTaskService(taskRepo, &MockColumnRepository{})

	err := svc.Delete(context.Background(), uuid.New(), uuid.New())
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}
