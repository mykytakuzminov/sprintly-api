package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/mykytakuzminov/sprintly-api/internal/domain"
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
	case errors.Is(err, domain.ErrInvalidCredentials):
		status = http.StatusUnauthorized
	case errors.Is(err, domain.ErrBadRequest):
		status = http.StatusBadRequest
	case errors.Is(err, domain.ErrTooManyRequests):
		status = http.StatusTooManyRequests
	default:
		status = http.StatusInternalServerError
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if encErr := json.NewEncoder(w).Encode(ErrorResponse{Error: err.Error()}); encErr != nil {
		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
	}
}

func successResponse(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
	}
}

func noContentResponse(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

type ErrorResponse struct {
	Error string `json:"error" example:"error description"`
}
