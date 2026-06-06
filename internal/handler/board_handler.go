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

type BoardHandler struct {
	svc    domain.BoardService
	auth   *auth.Auth
	logger *zap.SugaredLogger
}

func NewBoardHandler(
	svc domain.BoardService,
	auth *auth.Auth,
	logger *zap.SugaredLogger,
) *BoardHandler {
	return &BoardHandler{svc: svc, auth: auth, logger: logger}
}

func (h *BoardHandler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Post("/", h.Create)
	r.Get("/", h.GetAll)
	r.Get("/{id}", h.GetByID)
	r.Patch("/{id}", h.Update)
	r.Delete("/{id}", h.Delete)

	return r
}

func (h *BoardHandler) Create(w http.ResponseWriter, r *http.Request) {
	traceID := getTraceID(r, h.logger)

	userID, ok := getUserID(r)
	if !ok {
		logUnauthorizedAccess(h.logger, traceID)
		errorResponse(w, domain.ErrUnauthorized)
		return
	}

	var input domain.CreateBoardInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		logInvalidBody(h.logger, traceID, err)
		errorResponse(w, domain.ErrBadRequest)
		return
	}

	board, err := h.svc.Create(r.Context(), userID, &input)
	if err != nil {
		if errors.Is(err, domain.ErrBadRequest) {
			logInvalidBody(h.logger, traceID, err)
		} else {
			logUnexpectedError(h.logger, traceID, err)
		}
		errorResponse(w, err)
		return
	}

	logSuccess(h.logger, traceID, "board created", "user_id", userID, "board_id", board.ID)
	successResponse(w, http.StatusCreated, board)
}

func (h *BoardHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	traceID := getTraceID(r, h.logger)

	userID, ok := getUserID(r)
	if !ok {
		logUnauthorizedAccess(h.logger, traceID)
		errorResponse(w, domain.ErrUnauthorized)
		return
	}

	boards, err := h.svc.GetAllByUserID(r.Context(), userID)
	if err != nil {
		logUnexpectedError(h.logger, traceID, err)
		errorResponse(w, err)
		return
	}

	logSuccess(h.logger, traceID, "all boards retrieved", "user_id", userID)
	successResponse(w, http.StatusOK, boards)
}

func (h *BoardHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	traceID := getTraceID(r, h.logger)

	boardID, err := getURLParam(r, "id")
	if err != nil {
		logInvalidBody(h.logger, traceID, err)
		errorResponse(w, domain.ErrBadRequest)
		return
	}

	board, err := h.svc.GetByID(r.Context(), boardID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			logWarn(h.logger, traceID, "board not found", err)
		} else {
			logUnexpectedError(h.logger, traceID, err)
		}
		errorResponse(w, err)
		return
	}

	logSuccess(h.logger, traceID, "board retrieved", "board_id", boardID)
	successResponse(w, http.StatusOK, board)
}

func (h *BoardHandler) Update(w http.ResponseWriter, r *http.Request) {
	traceID := getTraceID(r, h.logger)

	userID, ok := getUserID(r)
	if !ok {
		logUnauthorizedAccess(h.logger, traceID)
		errorResponse(w, domain.ErrUnauthorized)
		return
	}

	boardID, err := getURLParam(r, "id")
	if err != nil {
		logInvalidBody(h.logger, traceID, err)
		errorResponse(w, domain.ErrBadRequest)
		return
	}

	var input domain.UpdateBoardInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		logInvalidBody(h.logger, traceID, err)
		errorResponse(w, domain.ErrBadRequest)
		return
	}

	if err := h.svc.Update(r.Context(), boardID, userID, &input); err != nil {
		switch {
		case errors.Is(err, domain.ErrNotFound):
			logWarn(h.logger, traceID, "board not found", err)
		case errors.Is(err, domain.ErrForbidden):
			logWarn(h.logger, traceID, "access denied", err)
		default:
			logUnexpectedError(h.logger, traceID, err)
		}
		errorResponse(w, err)
		return
	}

	logSuccess(h.logger, traceID, "board updated", "user_id", userID, "board_id", boardID)
	noContentResponse(w)
}

func (h *BoardHandler) Delete(w http.ResponseWriter, r *http.Request) {
	traceID := getTraceID(r, h.logger)

	userID, ok := getUserID(r)
	if !ok {
		logUnauthorizedAccess(h.logger, traceID)
		errorResponse(w, domain.ErrUnauthorized)
		return
	}

	boardID, err := getURLParam(r, "id")
	if err != nil {
		logInvalidBody(h.logger, traceID, err)
		errorResponse(w, domain.ErrBadRequest)
		return
	}

	if err := h.svc.Delete(r.Context(), boardID, userID); err != nil {
		switch {
		case errors.Is(err, domain.ErrNotFound):
			logWarn(h.logger, traceID, "board not found", err)
		case errors.Is(err, domain.ErrForbidden):
			logWarn(h.logger, traceID, "access denied", err)
		default:
			logUnexpectedError(h.logger, traceID, err)
		}
		errorResponse(w, err)
		return
	}

	logSuccess(h.logger, traceID, "board deleted", "user_id", userID, "board_id", boardID)
	noContentResponse(w)
}
