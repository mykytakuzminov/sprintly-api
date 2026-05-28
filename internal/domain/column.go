package domain

import (
	"context"

	"github.com/google/uuid"
)

type Column struct {
	ID       uuid.UUID
	BoardID  uuid.UUID
	Name     string
	Position uint
}

type ColumnRepository interface {
	Create(ctx context.Context, column *Column) error
	GetByID(ctx context.Context, id uuid.UUID) (*Column, error)
	GetAllByBoardID(ctx context.Context, boardID uuid.UUID) ([]*Column, error)
	Update(ctx context.Context, column *Column) error
	Delete(ctx context.Context, id uuid.UUID) error
}
