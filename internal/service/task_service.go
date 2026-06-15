package service

import (
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/mykytakuzminov/sprintly-api/internal/domain"
)

type TaskSvc struct {
	repo       domain.TaskRepository
	columnRepo domain.ColumnRepository
	validate   *validator.Validate
}

func NewTaskService(
	repo domain.TaskRepository,
	columnRepo domain.ColumnRepository,
) domain.TaskService {
	return &TaskSvc{
		repo:       repo,
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
		return nil, domain.ErrBadRequest
	}

	ownerID, err := s.columnRepo.GetOwnerID(ctx, columnID)
	if err != nil {
		return nil, err
	}

	if ownerID != userID {
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
	params *domain.ListParams,
) ([]*domain.Task, error) {
	return s.repo.GetAllByUserID(ctx, userID, params)
}

func (s *TaskSvc) GetAllByColumnID(
	ctx context.Context,
	columnID uuid.UUID,
	params *domain.ListParams,
) ([]*domain.Task, error) {
	return s.repo.GetAllByColumnID(ctx, columnID, params)
}

func (s *TaskSvc) GetAllByAssigneeID(
	ctx context.Context,
	assigneeID uuid.UUID,
	params *domain.ListParams,
) ([]*domain.Task, error) {
	return s.repo.GetAllByAssigneeID(ctx, assigneeID, params)
}

func (s *TaskSvc) Update(
	ctx context.Context,
	taskID, userID uuid.UUID,
	input *domain.UpdateTaskInput,
) error {
	if err := s.validate.Struct(input); err != nil {
		return domain.ErrBadRequest
	}

	ownerID, err := s.repo.GetOwnerID(ctx, taskID)
	if err != nil {
		return err
	}

	if ownerID != userID {
		return domain.ErrForbidden
	}

	task := &domain.Task{
		ID:          taskID,
		ColumnID:    input.ColumnID,
		AssigneeID:  input.AssigneeID,
		Name:        input.Name,
		Description: input.Description,
		DueDate:     input.DueDate,
	}

	return s.repo.Update(ctx, task)
}

func (s *TaskSvc) Delete(
	ctx context.Context,
	taskID, userID uuid.UUID,
) error {
	ownerID, err := s.repo.GetOwnerID(ctx, taskID)
	if err != nil {
		return err
	}

	if ownerID != userID {
		return domain.ErrForbidden
	}

	return s.repo.Delete(ctx, taskID)
}
