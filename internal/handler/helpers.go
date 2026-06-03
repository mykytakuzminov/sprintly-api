package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/mykytakuzminov/task-manager-api/internal/handler/middleware"
)

func getUserID(r *http.Request) (uuid.UUID, bool) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	return userID, ok
}

func getURLParam(r *http.Request, key string) (uuid.UUID, error) {
	return uuid.Parse(chi.URLParam(r, key))
}
