package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/mykytakuzminov/task-manager-api/internal/domain"
	"go.uber.org/zap"
)

type AuthHandler struct {
	svc    domain.AuthService
	logger *zap.SugaredLogger
}

func NewAuthHandler(svc domain.AuthService, logger *zap.SugaredLogger) *AuthHandler {
	return &AuthHandler{svc: svc, logger: logger}
}

func (h *AuthHandler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Post("/login", h.Login)
	r.Post("/logout", h.Logout)
	r.Post("/refresh", h.Refresh)

	return r
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	traceID := getTraceID(r, h.logger)

	var input domain.LoginInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		logInvalidBody(h.logger, traceID, err)
		errorResponse(w, domain.ErrBadRequest)
		return
	}

	tokens, err := h.svc.Login(r.Context(), &input)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			logWarn(h.logger, traceID, "wrong credentials", err)
		} else {
			logUnexpectedError(h.logger, traceID, err)
		}
		errorResponse(w, err)
		return
	}

	logSuccess(h.logger, traceID, "user authenticated")
	successResponse(w, http.StatusOK, tokens)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	traceID := getTraceID(r, h.logger)

	var input domain.LogoutInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		logInvalidBody(h.logger, traceID, err)
		errorResponse(w, domain.ErrBadRequest)
		return
	}

	if err := h.svc.Logout(r.Context(), &input); err != nil {
		logUnexpectedError(h.logger, traceID, err)
		errorResponse(w, err)
		return
	}

	logSuccess(h.logger, traceID, "user logged out")
	noContentResponse(w)
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	traceID := getTraceID(r, h.logger)

	var input domain.RefreshInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		logInvalidBody(h.logger, traceID, err)
		errorResponse(w, domain.ErrBadRequest)
		return
	}

	token, err := h.svc.Refresh(r.Context(), &input)
	if err != nil {
		if errors.Is(err, domain.ErrUnauthorized) {
			logWarn(h.logger, traceID, "user not authorized", err)
		} else {
			logUnexpectedError(h.logger, traceID, err)
		}
		errorResponse(w, err)
		return
	}

	logSuccess(h.logger, traceID, "access token refreshed")
	successResponse(w, http.StatusOK, toRefreshResponse(token))
}

type RefreshResponse struct {
	AccessToken string `json:"access_token" example:"access_token"`
}

func toRefreshResponse(token string) RefreshResponse {
	return RefreshResponse{AccessToken: token}
}
