package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/mykytakuzminov/task-manager-api/internal/auth"
	"github.com/mykytakuzminov/task-manager-api/internal/domain"
	"go.uber.org/zap"
)

type ColumnHandler struct {
	svc    domain.ColumnService
	auth   *auth.Auth
	logger *zap.SugaredLogger
}

func NewColumnHandler(
	svc domain.ColumnService,
	auth *auth.Auth,
	logger *zap.SugaredLogger,
) *ColumnHandler {
	return &ColumnHandler{svc: svc, auth: auth, logger: logger}
}

func (h *ColumnHandler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/{id}", h.GetByID)
	r.Patch("/{id}", h.Update)
	r.Delete("/{id}", h.Delete)

	return r
}

func (h *ColumnHandler) BoardRoutes() chi.Router {
	r := chi.NewRouter()

	r.Post("/", h.Create)
	r.Get("/", h.GetAll)

	return r
}

func (h *ColumnHandler) Create(w http.ResponseWriter, r *http.Request) {
	traceID := getTraceID(r, h.logger)

	userID, ok := getUserID(r)
	if !ok {
		logUnauthorizedAccess(h.logger, traceID)
		errorResponse(w, domain.ErrUnauthorized)
		return
	}

	boardID, err := getURLParam(r, "boardID")
	if err != nil {
		logInvalidBody(h.logger, traceID, err)
		errorResponse(w, domain.ErrBadRequest)
		return
	}

	var input domain.CreateColumnInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		logInvalidBody(h.logger, traceID, err)
		errorResponse(w, domain.ErrBadRequest)
		return
	}

	column, err := h.svc.Create(r.Context(), userID, boardID, &input)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrBadRequest):
			logInvalidBody(h.logger, traceID, err)
		case errors.Is(err, domain.ErrForbidden):
			logWarn(h.logger, traceID, "access denied", err)
		case errors.Is(err, domain.ErrNotFound):
			logWarn(h.logger, traceID, "board not found", err)
		default:
			logUnexpectedError(h.logger, traceID, err)
		}
		errorResponse(w, err)
		return
	}

	logSuccess(h.logger, traceID, "column created", "user_id", userID, "column_id", column.ID)
	successResponse(w, http.StatusCreated, column)
}

func (h *ColumnHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	traceID := getTraceID(r, h.logger)

	boardID, err := getURLParam(r, "boardID")
	if err != nil {
		logInvalidBody(h.logger, traceID, err)
		errorResponse(w, domain.ErrBadRequest)
		return
	}

	params := parseListParams(r)

	columns, err := h.svc.GetAllByBoardID(r.Context(), boardID, params)
	if err != nil {
		logUnexpectedError(h.logger, traceID, err)
		errorResponse(w, err)
		return
	}

	logSuccess(h.logger, traceID, "all columns retrieved", "board_id", boardID)
	successResponse(w, http.StatusOK, columns)
}

func (h *ColumnHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	traceID := getTraceID(r, h.logger)

	columnID, err := getURLParam(r, "id")
	if err != nil {
		logInvalidBody(h.logger, traceID, err)
		errorResponse(w, domain.ErrBadRequest)
		return
	}

	column, err := h.svc.GetByID(r.Context(), columnID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			logWarn(h.logger, traceID, "column not found", err)
		} else {
			logUnexpectedError(h.logger, traceID, err)
		}
		errorResponse(w, err)
		return
	}

	logSuccess(h.logger, traceID, "column retrieved", "column_id", columnID)
	successResponse(w, http.StatusOK, column)
}

func (h *ColumnHandler) Update(w http.ResponseWriter, r *http.Request) {
	traceID := getTraceID(r, h.logger)

	userID, ok := getUserID(r)
	if !ok {
		logUnauthorizedAccess(h.logger, traceID)
		errorResponse(w, domain.ErrUnauthorized)
		return
	}

	columnID, err := getURLParam(r, "id")
	if err != nil {
		logInvalidBody(h.logger, traceID, err)
		errorResponse(w, domain.ErrBadRequest)
		return
	}

	var input domain.UpdateColumnInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		logInvalidBody(h.logger, traceID, err)
		errorResponse(w, domain.ErrBadRequest)
		return
	}

	if err := h.svc.Update(r.Context(), columnID, userID, &input); err != nil {
		switch {
		case errors.Is(err, domain.ErrBadRequest):
			logInvalidBody(h.logger, traceID, err)
		case errors.Is(err, domain.ErrNotFound):
			logWarn(h.logger, traceID, "column owner not found", err)
		case errors.Is(err, domain.ErrForbidden):
			logWarn(h.logger, traceID, "access denied", err)
		default:
			logUnexpectedError(h.logger, traceID, err)
		}
		errorResponse(w, err)
		return
	}

	logSuccess(h.logger, traceID, "column updated", "user_id", userID, "column_id", columnID)
	noContentResponse(w)
}

func (h *ColumnHandler) Delete(w http.ResponseWriter, r *http.Request) {
	traceID := getTraceID(r, h.logger)

	userID, ok := getUserID(r)
	if !ok {
		logUnauthorizedAccess(h.logger, traceID)
		errorResponse(w, domain.ErrUnauthorized)
		return
	}

	columnID, err := getURLParam(r, "id")
	if err != nil {
		logInvalidBody(h.logger, traceID, err)
		errorResponse(w, domain.ErrBadRequest)
		return
	}

	if err := h.svc.Delete(r.Context(), columnID, userID); err != nil {
		switch {
		case errors.Is(err, domain.ErrNotFound):
			logWarn(h.logger, traceID, "column owner not found", err)
		case errors.Is(err, domain.ErrForbidden):
			logWarn(h.logger, traceID, "access denied", err)
		default:
			logUnexpectedError(h.logger, traceID, err)
		}
		errorResponse(w, err)
		return
	}

	logSuccess(h.logger, traceID, "column deleted", "user_id", userID, "column_id", columnID)
	noContentResponse(w)
}
