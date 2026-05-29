package service

import (
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/mykytakuzminov/task-manager-api/internal/domain"
)

type TaskSvc struct {
	repo       domain.TaskRepository
	boardRepo  domain.BoardRepository
	columnRepo domain.ColumnRepository
	validate   *validator.Validate
}

func NewTaskService(
	repo domain.TaskRepository,
	boardRepo domain.BoardRepository,
	columnRepo domain.ColumnRepository,
) domain.TaskService {
	return &TaskSvc{
		repo:       repo,
		boardRepo:  boardRepo,
		columnRepo: columnRepo,
		validate:   validator.New(),
	}
}

func (s *TaskSvc) Create(
	ctx context.Context,
	userID, columnID uuid.UUID,
	input *domain.CreateTaskInput,
) (*domain.Task, error) {
	if err := s.validate.Struct(input); err != nil {
		return nil, err
	}

	column, err := s.columnRepo.GetByID(ctx, columnID)
	if err != nil {
		return nil, err
	}

	board, err := s.boardRepo.GetByID(ctx, column.BoardID)
	if err != nil {
		return nil, err
	}
	if board.OwnerID != userID {
		return nil, domain.ErrForbidden
	}

	task := &domain.Task{
		ID:          uuid.New(),
		OwnerID:     userID,
		ColumnID:    columnID,
		AssigneeID:  input.AssigneeID,
		Name:        input.Name,
		Description: input.Description,
		DueDate:     input.DueDate,
	}

	if err := s.repo.Create(ctx, task); err != nil {
		return nil, err
	}

	return task, nil
}

func (s *TaskSvc) GetByID(
	ctx context.Context,
	taskID uuid.UUID,
) (*domain.Task, error) {
	return s.repo.GetByID(ctx, taskID)
}

func (s *TaskSvc) GetAllByUserID(
	ctx context.Context,
	userID uuid.UUID,
) ([]*domain.Task, error) {
	return s.repo.GetAllByUserID(ctx, userID)
}

func (s *TaskSvc) GetAllByColumnID(
	ctx context.Context,
	columnID uuid.UUID,
) ([]*domain.Task, error) {
	return s.repo.GetAllByColumnID(ctx, columnID)
}

func (s *TaskSvc) GetAllByAssigneeID(
	ctx context.Context,
	assigneeID uuid.UUID,
) ([]*domain.Task, error) {
	return s.repo.GetAllByAssigneeID(ctx, assigneeID)
}

func (s *TaskSvc) Update(
	ctx context.Context,
	taskID, userID uuid.UUID,
	input *domain.UpdateTaskInput,
) error {
	if err := s.validate.Struct(input); err != nil {
		return err
	}

	task, err := s.repo.GetByID(ctx, taskID)
	if err != nil {
		return err
	}

	column, err := s.columnRepo.GetByID(ctx, task.ColumnID)
	if err != nil {
		return err
	}

	board, err := s.boardRepo.GetByID(ctx, column.BoardID)
	if err != nil {
		return err
	}

	if board.OwnerID != userID && (task.AssigneeID == nil || *task.AssigneeID != userID) {
		return domain.ErrForbidden
	}

	task.ColumnID = input.ColumnID
	task.AssigneeID = input.AssigneeID
	task.Name = input.Name
	task.Description = input.Description
	task.DueDate = input.DueDate

	return s.repo.Update(ctx, task)
}

func (s *TaskSvc) Delete(
	ctx context.Context,
	taskID, userID uuid.UUID,
) error {
	task, err := s.repo.GetByID(ctx, taskID)
	if err != nil {
		return err
	}

	column, err := s.columnRepo.GetByID(ctx, task.ColumnID)
	if err != nil {
		return err
	}

	board, err := s.boardRepo.GetByID(ctx, column.BoardID)
	if err != nil {
		return err
	}

	if board.OwnerID != userID {
		return domain.ErrForbidden
	}

	return s.repo.Delete(ctx, taskID)
}
