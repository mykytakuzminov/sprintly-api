package domain

import "context"

type AuthTokens struct {
	AccessToken  string
	RefreshToken string
}

type LoginInput struct {
	Email    string `validate:"required,email,max=254"`
	Password string `validate:"required,min=8,max=72"`
}

type AuthService interface {
	Login(ctx context.Context, input *LoginInput) (*AuthTokens, error)
	Logout(ctx context.Context, refreshToken string) error
	Refresh(ctx context.Context, refreshToken string) (string, error)
}
