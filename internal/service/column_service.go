package service

import (
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/mykytakuzminov/task-manager-api/internal/domain"
)

type ColumnSvc struct {
	repo      domain.ColumnRepository
	boardRepo domain.BoardRepository
	validate  *validator.Validate
}

func NewColumnService(
	repo domain.ColumnRepository,
	boardRepo domain.BoardRepository,
) domain.ColumnService {
	return &ColumnSvc{
		repo:      repo,
		boardRepo: boardRepo,
		validate:  validator.New(),
	}
}

func (s *ColumnSvc) Create(
	ctx context.Context,
	userID, boardID uuid.UUID,
	input *domain.CreateColumnInput,
) (*domain.Column, error) {
	if err := s.validate.Struct(input); err != nil {
		return nil, domain.ErrBadRequest
	}

	board, err := s.boardRepo.GetByID(ctx, boardID)
	if err != nil {
		return nil, err
	}
	if board.OwnerID != userID {
		return nil, domain.ErrForbidden
	}

	column := &domain.Column{
		ID:       uuid.New(),
		BoardID:  boardID,
		Name:     input.Name,
		Position: input.Position,
	}

	if err := s.repo.Create(ctx, column); err != nil {
		return nil, err
	}

	return column, nil
}

func (s *ColumnSvc) GetByID(
	ctx context.Context,
	columnID uuid.UUID,
) (*domain.Column, error) {
	return s.repo.GetByID(ctx, columnID)
}

func (s *ColumnSvc) GetAllByBoardID(
	ctx context.Context,
	boardID uuid.UUID,
	params *domain.ListParams,
) ([]*domain.Column, error) {
	return s.repo.GetAllByBoardID(ctx, boardID, params)
}

func (s *ColumnSvc) Update(
	ctx context.Context,
	columnID, userID uuid.UUID,
	input *domain.UpdateColumnInput,
) error {
	if err := s.validate.Struct(input); err != nil {
		return domain.ErrBadRequest
	}

	ownerID, err := s.repo.GetOwnerID(ctx, columnID)
	if err != nil {
		return err
	}

	if ownerID != userID {
		return domain.ErrForbidden
	}

	column := &domain.Column{
		ID:       columnID,
		Name:     input.Name,
		Position: input.Position,
	}

	return s.repo.Update(ctx, column)
}

func (s *ColumnSvc) Delete(
	ctx context.Context,
	columnID, userID uuid.UUID,
) error {
	ownerID, err := s.repo.GetOwnerID(ctx, columnID)
	if err != nil {
		return err
	}

	if ownerID != userID {
		return domain.ErrForbidden
	}

	return s.repo.Delete(ctx, columnID)
}
