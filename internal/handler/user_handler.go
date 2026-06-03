package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/mykytakuzminov/task-manager-api/internal/auth"
	"github.com/mykytakuzminov/task-manager-api/internal/domain"
	"github.com/mykytakuzminov/task-manager-api/internal/handler/middleware"
)

type UserHandler struct {
	svc  domain.UserService
	auth *auth.Auth
}

func NewUserHandler(svc domain.UserService, auth *auth.Auth) *UserHandler {
	return &UserHandler{svc: svc, auth: auth}
}

func (h *UserHandler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Post("/register", h.Register)

	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(h.auth))
		r.Get("/me", h.Me)
		r.Patch("/me/password", h.ChangePassword)
	})

	return r
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var input domain.RegisterInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	user, err := h.svc.Register(r.Context(), &input)
	if err != nil {
		errorResponse(w, err)
		return
	}

	jsonResponse(w, http.StatusCreated, toUserResponse(user))
}

func (h *UserHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserID(r)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var input domain.ChangePasswordInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := h.svc.ChangePassword(r.Context(), userID, &input); err != nil {
		errorResponse(w, err)
		return
	}

	jsonResponse(w, http.StatusNoContent, nil)
}

func (h *UserHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserID(r)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	user, err := h.svc.GetByID(r.Context(), userID)
	if err != nil {
		errorResponse(w, err)
		return
	}

	jsonResponse(w, http.StatusOK, toUserResponse(user))
}
