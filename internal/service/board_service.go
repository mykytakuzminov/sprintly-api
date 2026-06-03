package service

import (
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/mykytakuzminov/task-manager-api/internal/domain"
)

type BoardSvc struct {
	repo     domain.BoardRepository
	validate *validator.Validate
}

func NewBoardService(repo domain.BoardRepository) domain.BoardService {
	return &BoardSvc{
		repo:     repo,
		validate: validator.New(),
	}
}

func (s *BoardSvc) Create(
	ctx context.Context,
	userID uuid.UUID,
	input *domain.CreateBoardInput,
) (*domain.Board, error) {
	if err := s.validate.Struct(input); err != nil {
		return nil, domain.ErrBadRequest
	}

	board := &domain.Board{
		ID:          uuid.New(),
		OwnerID:     userID,
		Name:        input.Name,
		Description: input.Description,
	}

	if err := s.repo.Create(ctx, board); err != nil {
		return nil, err
	}

	return board, nil
}

func (s *BoardSvc) GetByID(
	ctx context.Context,
	boardID uuid.UUID,
) (*domain.Board, error) {
	return s.repo.GetByID(ctx, boardID)
}

func (s *BoardSvc) GetAllByUserID(
	ctx context.Context,
	userID uuid.UUID,
) ([]*domain.Board, error) {
	return s.repo.GetAllByUserID(ctx, userID)
}

func (s *BoardSvc) Update(
	ctx context.Context,
	boardID, userID uuid.UUID,
	input *domain.UpdateBoardInput,
) error {
	board, err := s.repo.GetByID(ctx, boardID)
	if err != nil {
		return err
	}

	if board.OwnerID != userID {
		return domain.ErrForbidden
	}

	board.Name = input.Name
	board.Description = input.Description

	return s.repo.Update(ctx, board)
}

func (s *BoardSvc) Delete(
	ctx context.Context,
	boardID, userID uuid.UUID,
) error {
	board, err := s.repo.GetByID(ctx, boardID)
	if err != nil {
		return err
	}

	if board.OwnerID != userID {
		return domain.ErrForbidden
	}

	return s.repo.Delete(ctx, boardID)
}
