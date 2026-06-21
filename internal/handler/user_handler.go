package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/mykytakuzminov/sprintly-api/internal/domain"
	"go.uber.org/zap"
)

type UserHandler struct {
	svc    domain.UserService
	logger *zap.SugaredLogger
}

func NewUserHandler(
	svc domain.UserService,
	logger *zap.SugaredLogger,
) *UserHandler {
	return &UserHandler{svc: svc, logger: logger}
}

func (h *UserHandler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/me", h.Me)
	r.Patch("/me/password", h.ChangePassword)

	return r
}

func (h *UserHandler) AdminRoutes() chi.Router {
	r := chi.NewRouter()

	r.Get("/{id}", h.GetByID)
	r.Patch("/{id}", h.ChangeRole)
	r.Delete("/{id}", h.Delete)
	r.Get("/all", h.GetAll)

	return r
}

// Register godoc
// @Summary     Register a new user
// @Description Creates a new user account. Email must be unique.
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       body body domain.RegisterInput true "Registration data"
// @Success     201 {object} handler.UserResponse "User created successfully"
// @Failure     400 {object} handler.ErrorResponse "Invalid request body or validation error"
// @Failure     409 {object} handler.ErrorResponse "Email already in use"
// @Failure     500 {object} handler.ErrorResponse "Internal server error"
// @Router      /auth/register [post]
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

// Me godoc
// @Summary     Get current user
// @Description Returns the profile of the currently authenticated user.
// @Tags        users
// @Produce     json
// @Security    BearerAuth
// @Success     200 {object} handler.UserResponse "Current user data"
// @Failure     401 {object} handler.ErrorResponse "Missing or invalid access token"
// @Failure     404 {object} handler.ErrorResponse "User not found"
// @Failure     500 {object} handler.ErrorResponse "Internal server error"
// @Router      /users/me [get]
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

// GetByID godoc
// @Summary     Get user by ID
// @Description Returns the profile of a specific user by their ID. Requires admin role.
// @Tags        admin
// @Produce     json
// @Security    BearerAuth
// @Param       id path string true "User ID (UUID)"
// @Success     200 {object} handler.UserResponse "User data"
// @Failure     401 {object} handler.ErrorResponse "Missing or invalid access token"
// @Failure     403 {object} handler.ErrorResponse "Forbidden: admin role required"
// @Failure     404 {object} handler.ErrorResponse "User not found"
// @Failure     500 {object} handler.ErrorResponse "Internal server error"
// @Router      /admin/users/{id} [get]
func (h *UserHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	traceID := getTraceID(r, h.logger)

	adminID, ok := getUserID(r)
	if !ok {
		logUnauthorizedAccess(h.logger, traceID)
		errorResponse(w, domain.ErrUnauthorized)
		return
	}

	userID, err := getURLParam(r, "id")
	if err != nil {
		logInvalidBody(h.logger, traceID, err)
		errorResponse(w, domain.ErrBadRequest)
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

	logSuccess(h.logger, traceID, "user profile retrieved", "admin_id", adminID, "user_id", userID)
	successResponse(w, http.StatusOK, toUserResponse(user))
}

// GetAll godoc
// @Summary     Get all users
// @Description Returns all users. Requires admin role. Supports pagination and sorting.
// @Tags        admin
// @Produce     json
// @Security    BearerAuth
// @Param       limit    query  int    false "Number of results per page (default: 20, max: 100)"
// @Param       offset   query  int    false "Number of results to skip (default: 0)"
// @Param       sort     query  string false "Field to sort by: email, created_at, updated_at (default: created_at)"
// @Param       order    query  string false "Sort direction: ASC or DESC (default: ASC)"
// @Success     200 {array}  handler.UserResponse "List of users"
// @Failure     401 {object} handler.ErrorResponse "Missing or invalid access token"
// @Failure     403 {object} handler.ErrorResponse "Forbidden: admin role required"
// @Failure     500 {object} handler.ErrorResponse "Internal server error"
// @Router      /admin/users/all [get]
func (h *UserHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	traceID := getTraceID(r, h.logger)

	adminID, ok := getUserID(r)
	if !ok {
		logUnauthorizedAccess(h.logger, traceID)
		errorResponse(w, domain.ErrUnauthorized)
		return
	}

	params := parseListParams(r)

	users, err := h.svc.GetAll(r.Context(), params)
	if err != nil {
		logUnexpectedError(h.logger, traceID, err)
		errorResponse(w, err)
		return
	}

	responses := make([]UserResponse, 0, len(users))
	for _, user := range users {
		responses = append(responses, toUserResponse(user))
	}

	logSuccess(h.logger, traceID, "all users retrieved", "admin_id", adminID)
	successResponse(w, http.StatusOK, responses)
}

// ChangePassword godoc
// @Summary     Change password
// @Description Changes the password of the currently authenticated user. Requires the current password for verification.
// @Tags        users
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       body body domain.ChangePasswordInput true "Old and new passwords"
// @Success     204 "Password changed successfully"
// @Failure     400 {object} handler.ErrorResponse "Invalid request body or validation error"
// @Failure     401 {object} handler.ErrorResponse "Missing or invalid access token or wrong old password"
// @Failure     404 {object} handler.ErrorResponse "User not found"
// @Failure     500 {object} handler.ErrorResponse "Internal server error"
// @Router      /users/me/password [patch]
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

// ChangeRole godoc
// @Summary     Change role
// @Description Changes the role of the given user. Requires admin role.
// @Tags        admin
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       id path string true "User ID (UUID)"
// @Param       body body domain.ChangeRoleInput true "new role"
// @Success     204 "Role changed successfully"
// @Failure     400 {object} handler.ErrorResponse "Invalid request body or validation error"
// @Failure     401 {object} handler.ErrorResponse "Missing or invalid access token"
// @Failure     403 {object} handler.ErrorResponse "Forbidden: admin role required"
// @Failure     404 {object} handler.ErrorResponse "User not found"
// @Failure     500 {object} handler.ErrorResponse "Internal server error"
// @Router      /admin/users/{id} [patch]
func (h *UserHandler) ChangeRole(w http.ResponseWriter, r *http.Request) {
	traceID := getTraceID(r, h.logger)

	adminID, ok := getUserID(r)
	if !ok {
		logUnauthorizedAccess(h.logger, traceID)
		errorResponse(w, domain.ErrUnauthorized)
		return
	}

	userID, err := getURLParam(r, "id")
	if err != nil {
		logInvalidBody(h.logger, traceID, err)
		errorResponse(w, domain.ErrBadRequest)
		return
	}

	var input domain.ChangeRoleInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		logInvalidBody(h.logger, traceID, err)
		errorResponse(w, domain.ErrBadRequest)
		return
	}

	if err := h.svc.ChangeRole(r.Context(), userID, &input); err != nil {
		switch {
		case errors.Is(err, domain.ErrBadRequest):
			logInvalidBody(h.logger, traceID, err)
		case errors.Is(err, domain.ErrNotFound):
			logWarn(h.logger, traceID, "user not found", err)
		default:
			logUnexpectedError(h.logger, traceID, err)
		}
		errorResponse(w, err)
		return
	}

	logSuccess(h.logger, traceID, "role updated", "admin_id", adminID, "user_id", userID)
	noContentResponse(w)
}

// DeleteUser godoc
// @Summary     Delete a user
// @Description Deletes a user by ID. Requires admin role.
// @Tags        admin
// @Produce     json
// @Security    BearerAuth
// @Param       id path string true "User ID (UUID)"
// @Success     204 "User deleted successfully"
// @Failure     400 {object} handler.ErrorResponse "Invalid user ID format"
// @Failure     401 {object} handler.ErrorResponse "Missing or invalid access token"
// @Failure     403 {object} handler.ErrorResponse "Forbidden: admin role required"
// @Failure     404 {object} handler.ErrorResponse "User not found"
// @Failure     500 {object} handler.ErrorResponse "Internal server error"
// @Router      /admin/users/{id} [delete]
func (h *UserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	traceID := getTraceID(r, h.logger)

	adminID, ok := getUserID(r)
	if !ok {
		logUnauthorizedAccess(h.logger, traceID)
		errorResponse(w, domain.ErrUnauthorized)
		return
	}

	userID, err := getURLParam(r, "id")
	if err != nil {
		logInvalidBody(h.logger, traceID, err)
		errorResponse(w, domain.ErrBadRequest)
		return
	}

	if err := h.svc.Delete(r.Context(), userID); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			logWarn(h.logger, traceID, "user not found", err)
		} else {
			logUnexpectedError(h.logger, traceID, err)
		}
		errorResponse(w, err)
		return
	}

	logSuccess(h.logger, traceID, "user deleted", "admin_id", adminID, "user_id", userID)
	noContentResponse(w)
}

type UserResponse struct {
	ID        uuid.UUID `json:"id"         example:"7bc4f9d2-1a3e-4c8f-9b2d-5e7a0c4f1d63"`
	Email     string    `json:"email"      example:"user@example.com"`
	Role      string    `json:"role"       example:"member"`
	CreatedAt time.Time `json:"created_at" example:"2025-01-15T09:00:00Z"`
	UpdatedAt time.Time `json:"updated_at" example:"2025-01-15T09:00:00Z"`
}

func toUserResponse(user *domain.User) UserResponse {
	return UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}
