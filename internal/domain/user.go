package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID
	Email        string
	HashPassword string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type RegisterInput struct {
	Email    string `json:"email"    validate:"required,email,max=254" example:"john@example.com"`
	Password string `json:"password" validate:"required,min=8,max=72"  example:"password123"`
}

type ChangePasswordInput struct {
	OldPassword string `json:"old_password" validate:"required,min=8,max=72" example:"oldpassword123"`
	NewPassword string `json:"new_password" validate:"required,min=8,max=72" example:"newpassword123"`
}

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type UserService interface {
	Register(ctx context.Context, input *RegisterInput) (*User, error)
	ChangePassword(ctx context.Context, userID uuid.UUID, input *ChangePasswordInput) error
	GetByID(ctx context.Context, userID uuid.UUID) (*User, error)
}
