package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Column struct {
	ID        uuid.UUID `json:"id"         example:"b3c4d5e6-f7a8-9012-bcde-f12345678901"`
	BoardID   uuid.UUID `json:"board_id"   example:"2af3e6a1-5c4d-4b7e-8f1a-3d9c6b2e0a85"`
	Name      string    `json:"name"       example:"In Progress"`
	Position  uint      `json:"position"   example:"2"`
	CreatedAt time.Time `json:"created_at" example:"2025-01-15T09:00:00Z"`
	UpdatedAt time.Time `json:"updated_at" example:"2025-01-15T09:00:00Z"`
}

type CreateColumnInput struct {
	Name     string `json:"name"     validate:"required,max=100" example:"In Progress"`
	Position uint   `json:"position" validate:"required"         example:"2"`
}

type UpdateColumnInput struct {
	Name     string `json:"name"     validate:"required,max=100" example:"In Progress"`
	Position uint   `json:"position" validate:"required"         example:"2"`
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
