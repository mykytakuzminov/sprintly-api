package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/mykytakuzminov/task-manager-api/internal/domain"
)

func errorResponse(w http.ResponseWriter, err error) {
	var status int

	switch {
	case errors.Is(err, domain.ErrNotFound):
		status = http.StatusNotFound
	case errors.Is(err, domain.ErrConflict):
		status = http.StatusConflict
	case errors.Is(err, domain.ErrForbidden):
		status = http.StatusForbidden
	case errors.Is(err, domain.ErrUnauthorized):
		status = http.StatusUnauthorized
	default:
		status = http.StatusInternalServerError
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(ErrorResponse{Error: err.Error()})
}

func jsonResponse(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

type RefreshResponse struct {
	AccessToken string `json:"access_token" example:"access_token"`
}

func toRefreshResponse(token string) RefreshResponse {
	return RefreshResponse{AccessToken: token}
}

type UserResponse struct {
	ID        uuid.UUID `json:"id"         example:"550e8400-e29b-41d4-a716-446655440000"`
	Email     string    `json:"email"      example:"john@example.com"`
	CreatedAt time.Time `json:"created_at" example:"2024-01-01T00:00:00Z"`
	UpdatedAt time.Time `json:"updated_at" example:"2024-01-01T00:00:00Z"`
}

func toUserResponse(user *domain.User) UserResponse {
	return UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}

type ErrorResponse struct {
	Error string `json:"error" example:"error description"`
}
