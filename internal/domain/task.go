package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Task struct {
	ID          uuid.UUID
	OwnerID     uuid.UUID
	ColumnID    uuid.UUID
	AssigneeID  *uuid.UUID
	Name        string
	Description *string
	DueDate     *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type CreateTaskInput struct {
	AssigneeID  *uuid.UUID `validate:"omitempty"`
	Name        string     `validate:"required,max=100"`
	Description *string    `validate:"omitempty,max=500"`
	DueDate     *time.Time `validate:"omitempty"`
}

type UpdateTaskInput struct {
	ColumnID    uuid.UUID  `validate:"required"`
	AssigneeID  *uuid.UUID `validate:"omitempty"`
	Name        string     `validate:"required,max=100"`
	Description *string    `validate:"omitempty,max=500"`
	DueDate     *time.Time `validate:"omitempty"`
}

type TaskRepository interface {
	Create(ctx context.Context, task *Task) error
	GetByID(ctx context.Context, id uuid.UUID) (*Task, error)
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
