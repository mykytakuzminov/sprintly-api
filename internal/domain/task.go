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

type TaskRepository interface {
	Create(ctx context.Context, task *Task) error
	GetByID(ctx context.Context, id uuid.UUID) (*Task, error)
	GetAllByUserID(ctx context.Context, userID uuid.UUID) ([]*Task, error)
	GetAllByColumnID(ctx context.Context, columnID uuid.UUID) ([]*Task, error)
	GetAllByAssigneeID(ctx context.Context, assigneeID uuid.UUID) ([]*Task, error)
	Update(ctx context.Context, task *Task) error
	Delete(ctx context.Context, id uuid.UUID) error
}
