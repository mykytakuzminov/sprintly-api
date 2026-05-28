package service

import (
	"context"
	"errors"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/mykytakuzminov/task-manager-api/internal/domain"
)

type UserService struct {
	repo     domain.UserRepository
	validate *validator.Validate
}

func NewUserService(repo domain.UserRepository) *UserService {
	return &UserService{
		repo:     repo,
		validate: validator.New(),
	}
}

func (s *UserService) Register(
	ctx context.Context,
	input *domain.RegisterInput,
) (*domain.User, error) {
	if err := s.validate.Struct(input); err != nil {
		return nil, err
	}

	hashPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), 12)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		ID:           uuid.New(),
		Email:        input.Email,
		HashPassword: string(hashPassword),
	}

	_, err = s.repo.GetByEmail(ctx, user.Email)
	if err == nil {
		return nil, domain.ErrConflict
	}
	if !errors.Is(err, domain.ErrNotFound) {
		return nil, err
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) ChangePassword(
	ctx context.Context,
	userID uuid.UUID,
	input *domain.ChangePasswordInput,
) error {
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	if err = bcrypt.CompareHashAndPassword(
		[]byte(user.HashPassword),
		[]byte(input.OldPassword),
	); err != nil {
		return domain.ErrForbidden
	}

	hashPassword, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), 12)
	if err != nil {
		return err
	}

	user.HashPassword = string(hashPassword)

	if err = s.repo.Update(ctx, user); err != nil {
		return err
	}

	return nil
}

func (s *UserService) GetByID(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
	return s.repo.GetByID(ctx, userID)
}
