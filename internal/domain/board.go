package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Board struct {
	ID          uuid.UUID
	OwnerID     uuid.UUID
	Name        string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type BoardRepository interface {
	Create(ctx context.Context, board *Board) error
	GetByID(ctx context.Context, id uuid.UUID) (*Board, error)
	GetAllByUserID(ctx context.Context, userID uuid.UUID) ([]*Board, error)
	Update(ctx context.Context, board *Board) error
	Delete(ctx context.Context, id uuid.UUID) error
}
