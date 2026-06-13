package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Column struct {
	ID        uuid.UUID `json:"id"         example:"550e8400-e29b-41d4-a716-446655440000"`
	BoardID   uuid.UUID `json:"board_id"   example:"550e8400-e29b-41d4-a716-446655440001"`
	Name      string    `json:"name"       example:"In Progress"`
	Position  uint      `json:"position"   example:"1"`
	CreatedAt time.Time `json:"created_at" example:"2024-01-01T00:00:00Z"`
	UpdatedAt time.Time `json:"updated_at" example:"2024-01-01T00:00:00Z"`
}

type CreateColumnInput struct {
	Name     string `json:"name"     validate:"required,max=100" example:"In Progress"`
	Position uint   `json:"position" validate:"required"         example:"1"`
}

type UpdateColumnInput struct {
	Name     string `json:"name"     validate:"required,max=100" example:"In Progress"`
	Position uint   `json:"position" validate:"required"         example:"1"`
}

type ColumnRepository interface {
	Create(ctx context.Context, column *Column) error
	GetByID(ctx context.Context, id uuid.UUID) (*Column, error)
	GetOwnerID(ctx context.Context, id uuid.UUID) (uuid.UUID, error)
	GetAllByBoardID(ctx context.Context, boardID uuid.UUID, params *ListParams) ([]*Column, error)
	Update(ctx context.Context, column *Column) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type ColumnService interface {
	Create(ctx context.Context, userID, boardID uuid.UUID, input *CreateColumnInput) (*Column, error)
	GetByID(ctx context.Context, columnID uuid.UUID) (*Column, error)
	GetAllByBoardID(ctx context.Context, boardID uuid.UUID, params *ListParams) ([]*Column, error)
	Update(ctx context.Context, columnID, userID uuid.UUID, input *UpdateColumnInput) error
	Delete(ctx context.Context, columnID, userID uuid.UUID) error
}
