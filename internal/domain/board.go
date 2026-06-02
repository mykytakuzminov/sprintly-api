package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Board struct {
	ID          uuid.UUID `json:"id"          example:"550e8400-e29b-41d4-a716-446655440000"`
	OwnerID     uuid.UUID `json:"owner_id"    example:"550e8400-e29b-41d4-a716-446655440001"`
	Name        string    `json:"name"        example:"My Project"`
	Description *string   `json:"description" example:"Project description"`
	CreatedAt   time.Time `json:"created_at"  example:"2024-01-01T00:00:00Z"`
	UpdatedAt   time.Time `json:"updated_at"  example:"2024-01-01T00:00:00Z"`
}

type CreateBoardInput struct {
	Name        string  `json:"name"        validate:"required,max=100"  example:"My Project"`
	Description *string `json:"description" validate:"omitempty,max=500" example:"Project description"`
}

type UpdateBoardInput struct {
	Name        string  `json:"name"        validate:"required,max=100"  example:"My Project"`
	Description *string `json:"description" validate:"omitempty,max=500" example:"Project description"`
}

type BoardRepository interface {
	Create(ctx context.Context, board *Board) error
	GetByID(ctx context.Context, id uuid.UUID) (*Board, error)
	GetAllByUserID(ctx context.Context, userID uuid.UUID) ([]*Board, error)
	Update(ctx context.Context, board *Board) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type BoardService interface {
	Create(ctx context.Context, userID uuid.UUID, input *CreateBoardInput) (*Board, error)
	GetByID(ctx context.Context, boardID uuid.UUID) (*Board, error)
	GetAllByUserID(ctx context.Context, userID uuid.UUID) ([]*Board, error)
	Update(ctx context.Context, boardID, userID uuid.UUID, input *UpdateBoardInput) error
	Delete(ctx context.Context, boardID, userID uuid.UUID) error
}
