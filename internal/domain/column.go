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

type CreateColumnInput struct {
	Name     string `validate:"required,max=100"`
	Position uint   `validate:"required"`
}

type UpdateColumnInput struct {
	Name     string `validate:"required,max=100"`
	Position uint   `validate:"required"`
}

type ColumnRepository interface {
	Create(ctx context.Context, column *Column) error
	GetByID(ctx context.Context, id uuid.UUID) (*Column, error)
	GetAllByBoardID(ctx context.Context, boardID uuid.UUID) ([]*Column, error)
	Update(ctx context.Context, column *Column) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type ColumnService interface {
	Create(ctx context.Context, userID, boardID uuid.UUID, input *CreateColumnInput) (*Column, error)
	GetByID(ctx context.Context, columnID uuid.UUID) (*Column, error)
	GetAllByBoardID(ctx context.Context, boardID uuid.UUID) ([]*Column, error)
	Update(ctx context.Context, columnID, userID uuid.UUID, input *UpdateColumnInput) error
	Delete(ctx context.Context, columnID, userID uuid.UUID) error
}
