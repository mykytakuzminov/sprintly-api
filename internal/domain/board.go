package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Board struct {
	ID          uuid.UUID `json:"id"          example:"2af3e6a1-5c4d-4b7e-8f1a-3d9c6b2e0a85"`
	OwnerID     uuid.UUID `json:"owner_id"    example:"7bc4f9d2-1a3e-4c8f-9b2d-5e7a0c4f1d63"`
	Name        string    `json:"name"        example:"Q3 Product Roadmap"`
	Description *string   `json:"description" example:"Planning board for Q3 product initiatives"`
	CreatedAt   time.Time `json:"created_at"  example:"2025-01-15T09:00:00Z"`
	UpdatedAt   time.Time `json:"updated_at"  example:"2025-01-15T09:00:00Z"`
}

type CreateBoardInput struct {
	Name        string  `json:"name"        validate:"required,max=100"  example:"Q3 Product Roadmap"`
	Description *string `json:"description" validate:"omitempty,max=500" example:"Planning board for Q3 product initiatives"`
}

type UpdateBoardInput struct {
	Name        string  `json:"name"        validate:"required,max=100"  example:"Q3 Product Roadmap"`
	Description *string `json:"description" validate:"omitempty,max=500" example:"Planning board for Q3 product initiatives"`
}

type BoardRepository interface {
	Create(ctx context.Context, board *Board) error
	GetByID(ctx context.Context, id uuid.UUID) (*Board, error)
	GetAllByUserID(ctx context.Context, userID uuid.UUID, params *ListParams) ([]*Board, error)
	Update(ctx context.Context, board *Board) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type BoardService interface {
	Create(ctx context.Context, userID uuid.UUID, input *CreateBoardInput) (*Board, error)
	GetByID(ctx context.Context, boardID uuid.UUID) (*Board, error)
	GetAllByUserID(ctx context.Context, userID uuid.UUID, params *ListParams) ([]*Board, error)
	Update(ctx context.Context, boardID, userID uuid.UUID, input *UpdateBoardInput) error
	Delete(ctx context.Context, boardID, userID uuid.UUID) error
}
