package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/mykytakuzminov/task-manager-api/internal/domain"
	"github.com/mykytakuzminov/task-manager-api/internal/handler/middleware"
)

func errorHandler(w http.ResponseWriter, err error) {
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
	json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
}

func jsonResponse(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func getUserID(r *http.Request) (uuid.UUID, bool) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	return userID, ok
}

func getURLParam(r *http.Request, key string) (uuid.UUID, error) {
	return uuid.Parse(chi.URLParam(r, key))
}
