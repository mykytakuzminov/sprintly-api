package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Task struct {
	ID          uuid.UUID  `json:"id"          example:"550e8400-e29b-41d4-a716-446655440000"`
	OwnerID     uuid.UUID  `json:"owner_id"    example:"550e8400-e29b-41d4-a716-446655440001"`
	ColumnID    uuid.UUID  `json:"column_id"   example:"550e8400-e29b-41d4-a716-446655440002"`
	AssigneeID  *uuid.UUID `json:"assignee_id" example:"550e8400-e29b-41d4-a716-446655440003"`
	Name        string     `json:"name"        example:"Fix login bug"`
	Description *string    `json:"description" example:"Task description"`
	DueDate     *time.Time `json:"due_date"    example:"2024-12-31T00:00:00Z"`
	CreatedAt   time.Time  `json:"created_at"  example:"2024-01-01T00:00:00Z"`
	UpdatedAt   time.Time  `json:"updated_at"  example:"2024-01-01T00:00:00Z"`
}

type CreateTaskInput struct {
	AssigneeID  *uuid.UUID `json:"assignee_id" validate:"omitempty"         example:"550e8400-e29b-41d4-a716-446655440003"`
	Name        string     `json:"name"        validate:"required,max=100"  example:"Fix login bug"`
	Description *string    `json:"description" validate:"omitempty,max=500" example:"Task description"`
	DueDate     *time.Time `json:"due_date"    validate:"omitempty"         example:"2024-12-31T00:00:00Z"`
}

type UpdateTaskInput struct {
	ColumnID    uuid.UUID  `json:"column_id"   validate:"required"          example:"550e8400-e29b-41d4-a716-446655440002"`
	AssigneeID  *uuid.UUID `json:"assignee_id" validate:"omitempty"         example:"550e8400-e29b-41d4-a716-446655440003"`
	Name        string     `json:"name"        validate:"required,max=100"  example:"Fix login bug"`
	Description *string    `json:"description" validate:"omitempty,max=500" example:"Task description"`
	DueDate     *time.Time `json:"due_date"    validate:"omitempty"         example:"2024-12-31T00:00:00Z"`
}

type TaskRepository interface {
	Create(ctx context.Context, task *Task) error
	GetByID(ctx context.Context, id uuid.UUID) (*Task, error)
	GetOwnerID(ctx context.Context, id uuid.UUID) (uuid.UUID, error)
	GetAllByUserID(ctx context.Context, userID uuid.UUID, params *ListParams) ([]*Task, error)
	GetAllByColumnID(ctx context.Context, columnID uuid.UUID, params *ListParams) ([]*Task, error)
	GetAllByAssigneeID(ctx context.Context, assigneeID uuid.UUID, params *ListParams) ([]*Task, error)
	Update(ctx context.Context, task *Task) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type TaskService interface {
	Create(ctx context.Context, userID, columnID uuid.UUID, input *CreateTaskInput) (*Task, error)
	GetByID(ctx context.Context, taskID uuid.UUID) (*Task, error)
	GetAllByUserID(ctx context.Context, userID uuid.UUID, params *ListParams) ([]*Task, error)
	GetAllByColumnID(ctx context.Context, columnID uuid.UUID, params *ListParams) ([]*Task, error)
	GetAllByAssigneeID(ctx context.Context, assigneeID uuid.UUID, params *ListParams) ([]*Task, error)
	Update(ctx context.Context, taskID, userID uuid.UUID, input *UpdateTaskInput) error
	Delete(ctx context.Context, taskID, userID uuid.UUID) error
}
