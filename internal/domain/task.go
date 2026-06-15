package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Task struct {
	ID          uuid.UUID  `json:"id"          example:"c4d5e6f7-a8b9-0123-cdef-123456789012"`
	OwnerID     uuid.UUID  `json:"owner_id"    example:"7bc4f9d2-1a3e-4c8f-9b2d-5e7a0c4f1d63"`
	ColumnID    uuid.UUID  `json:"column_id"   example:"b3c4d5e6-f7a8-9012-bcde-f12345678901"`
	AssigneeID  *uuid.UUID `json:"assignee_id" example:"9ec632eb-2345-5678-90bc-def012345678"`
	Name        string     `json:"name"        example:"Fix authentication bug"`
	Description *string    `json:"description" example:"Token expiration is not handled correctly on the client side"`
	DueDate     *time.Time `json:"due_date"    example:"2025-06-30T23:59:59Z"`
	CreatedAt   time.Time  `json:"created_at"  example:"2025-01-15T09:00:00Z"`
	UpdatedAt   time.Time  `json:"updated_at"  example:"2025-01-15T09:00:00Z"`
}

type CreateTaskInput struct {
	AssigneeID  *uuid.UUID `json:"assignee_id" validate:"omitempty"         example:"9ec632eb-2345-5678-90bc-def012345678"`
	Name        string     `json:"name"        validate:"required,max=100"  example:"Fix authentication bug"`
	Description *string    `json:"description" validate:"omitempty,max=500" example:"Token expiration is not handled correctly on the client side"`
	DueDate     *time.Time `json:"due_date"    validate:"omitempty"         example:"2025-06-30T23:59:59Z"`
}

type UpdateTaskInput struct {
	ColumnID    uuid.UUID  `json:"column_id"   validate:"required"          example:"b3c4d5e6-f7a8-9012-bcde-f12345678901"`
	AssigneeID  *uuid.UUID `json:"assignee_id" validate:"omitempty"         example:"9ec632eb-2345-5678-90bc-def012345678"`
	Name        string     `json:"name"        validate:"required,max=100"  example:"Fix authentication bug"`
	Description *string    `json:"description" validate:"omitempty,max=500" example:"Token expiration is not handled correctly on the client side"`
	DueDate     *time.Time `json:"due_date"    validate:"omitempty"         example:"2025-06-30T23:59:59Z"`
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
