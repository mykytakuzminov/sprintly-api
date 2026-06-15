package domain

import "context"

type AuthTokens struct {
	AccessToken  string `json:"access_token"  example:"<access_token>"`
	RefreshToken string `json:"refresh_token" example:"<refresh_token>"`
}

type LoginInput struct {
	Email    string `json:"email"    validate:"required,email,max=254" example:"user@example.com"`
	Password string `json:"password" validate:"required,min=8,max=72"  example:"password"`
}

type LogoutInput struct {
	RefreshToken string `json:"refresh_token" validate:"required" example:"<refresh_token>"`
}

type RefreshInput struct {
	RefreshToken string `json:"refresh_token" validate:"required" example:"<refresh_token>"`
}

type AuthService interface {
	Login(ctx context.Context, input *LoginInput) (*AuthTokens, error)
	Logout(ctx context.Context, input *LogoutInput) error
	Refresh(ctx context.Context, input *RefreshInput) (string, error)
}
