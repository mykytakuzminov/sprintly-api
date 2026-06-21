package service

import (
	"context"
	"errors"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/mykytakuzminov/sprintly-api/internal/domain"
)

type UserSvc struct {
	repo     domain.UserRepository
	validate *validator.Validate
}

func NewUserService(repo domain.UserRepository) domain.UserService {
	return &UserSvc{
		repo:     repo,
		validate: validator.New(),
	}
}

func (s *UserSvc) Register(
	ctx context.Context,
	input *domain.RegisterInput,
) (*domain.User, error) {
	if err := s.validate.Struct(input); err != nil {
		return nil, domain.ErrBadRequest
	}

	_, err := s.repo.GetByEmail(ctx, input.Email)
	if err == nil {
		return nil, domain.ErrConflict
	}
	if !errors.Is(err, domain.ErrNotFound) {
		return nil, err
	}

	hpwd, err := bcrypt.GenerateFromPassword([]byte(input.Password), 12)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		ID:           uuid.New(),
		Email:        input.Email,
		HashPassword: string(hpwd),
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserSvc) GetByID(
	ctx context.Context,
	userID uuid.UUID,
) (*domain.User, error) {
	return s.repo.GetByID(ctx, userID)
}

func (s *UserSvc) GetAll(
	ctx context.Context,
	params *domain.ListParams,
) ([]*domain.User, error) {
	return s.repo.GetAll(ctx, params)
}

func (s *UserSvc) ChangePassword(
	ctx context.Context,
	userID uuid.UUID,
	input *domain.ChangePasswordInput,
) error {
	if err := s.validate.Struct(input); err != nil {
		return domain.ErrBadRequest
	}

	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	if err = bcrypt.CompareHashAndPassword(
		[]byte(user.HashPassword),
		[]byte(input.OldPassword),
	); err != nil {
		return domain.ErrInvalidCredentials
	}

	hpwd, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), 12)
	if err != nil {
		return err
	}

	if err = s.repo.UpdatePassword(ctx, user.ID, string(hpwd)); err != nil {
		return err
	}

	return nil
}

func (s *UserSvc) ChangeRole(
	ctx context.Context,
	userID uuid.UUID,
	input *domain.ChangeRoleInput,
) error {
	if err := s.validate.Struct(input); err != nil {
		return domain.ErrBadRequest
	}

	_, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	if err := s.repo.UpdateRole(ctx, userID, input.Role); err != nil {
		return err
	}

	return nil
}
