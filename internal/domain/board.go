package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Board struct {
	ID          uuid.UUID `json:"id"`
	OwnerID     uuid.UUID `json:"owner_id"`
	Name        string    `json:"name"`
	Description *string   `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateBoardInput struct {
	Name        string  `json:"name"        validate:"required,max=100"`
	Description *string `json:"description" validate:"omitempty,max=500"`
}

type UpdateBoardInput struct {
	Name        string  `json:"name"        validate:"required,max=100"`
	Description *string `kson:"description" validate:"omitempty,max=500"`
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
