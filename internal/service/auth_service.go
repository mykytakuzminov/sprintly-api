package service

import (
	"context"
	"errors"

	"github.com/mykytakuzminov/task-manager-api/internal/auth"
	"github.com/mykytakuzminov/task-manager-api/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

type AuthSvc struct {
	userRepo  domain.UserRepository
	tokenRepo domain.TokenRepository
	auth      *auth.Auth
}

func NewAuthService(
	userRepo domain.UserRepository,
	tokenRepo domain.TokenRepository,
	auth *auth.Auth,
) domain.AuthService {
	return &AuthSvc{
		userRepo:  userRepo,
		tokenRepo: tokenRepo,
		auth:      auth,
	}
}

func (s *AuthSvc) Login(
	ctx context.Context,
	input *domain.LoginInput,
) (*domain.AuthTokens, error) {
	user, err := s.userRepo.GetByEmail(ctx, input.Email)
	if err != nil {
		return nil, err
	}

	if err = bcrypt.CompareHashAndPassword(
		[]byte(user.HashPassword),
		[]byte(input.Password),
	); err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	atoken, err := s.auth.GenerateAccessToken(user.ID)
	if err != nil {
		return nil, err
	}

	rtoken, err := s.auth.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, err
	}

	if err := s.tokenRepo.Set(
		ctx,
		rtoken,
		user.ID,
		s.auth.RefreshTTL(),
	); err != nil {
		return nil, err
	}

	tokens := &domain.AuthTokens{
		AccessToken:  atoken,
		RefreshToken: rtoken,
	}

	return tokens, nil
}

func (s *AuthSvc) Logout(
	ctx context.Context,
	input *domain.LogoutInput,
) error {
	return s.tokenRepo.Delete(ctx, input.RefreshToken)
}

func (s *AuthSvc) Refresh(
	ctx context.Context,
	input *domain.RefreshInput,
) (string, error) {
	userID, err := s.tokenRepo.Get(ctx, input.RefreshToken)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return "", domain.ErrUnauthorized
		}
		return "", err
	}

	atoken, err := s.auth.GenerateAccessToken(userID)
	if err != nil {
		return "", err
	}

	return atoken, nil
}
