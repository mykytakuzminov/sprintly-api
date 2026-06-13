package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/mykytakuzminov/task-manager-api/internal/auth"
	"github.com/mykytakuzminov/task-manager-api/internal/domain"
	"go.uber.org/zap"
)

type UserHandler struct {
	svc    domain.UserService
	auth   *auth.Auth
	logger *zap.SugaredLogger
}

func NewUserHandler(
	svc domain.UserService,
	auth *auth.Auth,
	logger *zap.SugaredLogger,
) *UserHandler {
	return &UserHandler{svc: svc, auth: auth, logger: logger}
}

func (h *UserHandler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/me", h.Me)
	r.Patch("/me/password", h.ChangePassword)

	return r
}

func (h *UserHandler) RouteRegister() chi.Router {
	r := chi.NewRouter()

	r.Post("/", h.Register)

	return r
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	traceID := getTraceID(r, h.logger)

	var input domain.RegisterInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		logInvalidBody(h.logger, traceID, err)
		errorResponse(w, domain.ErrBadRequest)
		return
	}

	user, err := h.svc.Register(r.Context(), &input)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrBadRequest):
			logWarn(h.logger, traceID, "invalid registration data", err)
		case errors.Is(err, domain.ErrConflict):
			logWarn(h.logger, traceID, "user already exists", err)
		default:
			logUnexpectedError(h.logger, traceID, err)
		}
		errorResponse(w, err)
		return
	}

	logSuccess(h.logger, traceID, "user created", "user_id", user.ID)
	successResponse(w, http.StatusCreated, toUserResponse(user))
}

func (h *UserHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	traceID := getTraceID(r, h.logger)

	userID, ok := getUserID(r)
	if !ok {
		logUnauthorizedAccess(h.logger, traceID)
		errorResponse(w, domain.ErrUnauthorized)
		return
	}

	var input domain.ChangePasswordInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		logInvalidBody(h.logger, traceID, err)
		errorResponse(w, domain.ErrBadRequest)
		return
	}

	if err := h.svc.ChangePassword(r.Context(), userID, &input); err != nil {
		switch {
		case errors.Is(err, domain.ErrNotFound):
			logWarn(h.logger, traceID, "user not found", err)
		case errors.Is(err, domain.ErrInvalidCredentials):
			logWarn(h.logger, traceID, "wrong credentials", err)
		default:
			logUnexpectedError(h.logger, traceID, err)
		}
		errorResponse(w, err)
		return
	}

	logSuccess(h.logger, traceID, "password updated", "user_id", userID)
	noContentResponse(w)
}

func (h *UserHandler) Me(w http.ResponseWriter, r *http.Request) {
	traceID := getTraceID(r, h.logger)

	userID, ok := getUserID(r)
	if !ok {
		logUnauthorizedAccess(h.logger, traceID)
		errorResponse(w, domain.ErrUnauthorized)
		return
	}

	user, err := h.svc.GetByID(r.Context(), userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			logWarn(h.logger, traceID, "user not found", err)
		} else {
			logUnexpectedError(h.logger, traceID, err)
		}
		errorResponse(w, err)
		return
	}

	logSuccess(h.logger, traceID, "user profile retrieved", "user_id", userID)
	successResponse(w, http.StatusOK, toUserResponse(user))
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
