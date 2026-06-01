package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Task struct {
	ID          uuid.UUID  `json:"id"`
	OwnerID     uuid.UUID  `json:"owner_id"`
	ColumnID    uuid.UUID  `json:"column_id"`
	AssigneeID  *uuid.UUID `json:"assignee_id"`
	Name        string     `json:"name"`
	Description *string    `json:"description"`
	DueDate     *time.Time `json:"due_date"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type CreateTaskInput struct {
	AssigneeID  *uuid.UUID `json:"assignee_id" validate:"omitempty"`
	Name        string     `json:"name" validate:"required,max=100"`
	Description *string    `json:"description" validate:"omitempty,max=500"`
	DueDate     *time.Time `json:"due_date" validate:"omitempty"`
}

type UpdateTaskInput struct {
	ColumnID    uuid.UUID  `json:"column_id" validate:"required"`
	AssigneeID  *uuid.UUID `json:"assignee_id" validate:"omitempty"`
	Name        string     `json:"name" validate:"required,max=100"`
	Description *string    `json:"description" validate:"omitempty,max=500"`
	DueDate     *time.Time `json:"due_date" validate:"omitempty"`
}

type TaskRepository interface {
	Create(ctx context.Context, task *Task) error
	GetByID(ctx context.Context, id uuid.UUID) (*Task, error)
	GetOwnerID(ctx context.Context, id uuid.UUID) (uuid.UUID, error)
	GetAllByUserID(ctx context.Context, userID uuid.UUID) ([]*Task, error)
	GetAllByColumnID(ctx context.Context, columnID uuid.UUID) ([]*Task, error)
	GetAllByAssigneeID(ctx context.Context, assigneeID uuid.UUID) ([]*Task, error)
	Update(ctx context.Context, task *Task) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type TaskService interface {
	Create(ctx context.Context, userID, columnID uuid.UUID, input *CreateTaskInput) (*Task, error)
	GetByID(ctx context.Context, taskID uuid.UUID) (*Task, error)
	GetAllByUserID(ctx context.Context, userID uuid.UUID) ([]*Task, error)
	GetAllByColumnID(ctx context.Context, columnID uuid.UUID) ([]*Task, error)
	GetAllByAssigneeID(ctx context.Context, assigneeID uuid.UUID) ([]*Task, error)
	Update(ctx context.Context, taskID, userID uuid.UUID, input *UpdateTaskInput) error
	Delete(ctx context.Context, taskID, userID uuid.UUID) error
}
