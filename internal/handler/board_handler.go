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

// CreateBoard godoc
// @Summary     Create a board
// @Description Creates a new board owned by the authenticated user.
// @Tags        boards
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       body body domain.CreateBoardInput true "Board data"
// @Success     201 {object} domain.Board "Board created successfully"
// @Failure     400 {object} handler.ErrorResponse "Invalid request body or validation error"
// @Failure     401 {object} handler.ErrorResponse "Missing or invalid access token"
// @Failure     500 {object} handler.ErrorResponse "Internal server error"
// @Router      /boards [post]
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

// GetAllBoards godoc
// @Summary     Get all boards
// @Description Returns all boards owned by the authenticated user. Supports pagination and sorting.
// @Tags        boards
// @Produce     json
// @Security    BearerAuth
// @Param       limit  query int    false "Number of results per page (default: 20, max: 100)"
// @Param       offset query int    false "Number of results to skip (default: 0)"
// @Param       sort   query string false "Field to sort by: name, created_at, updated_at (default: created_at)"
// @Param       order  query string false "Sort direction: ASC or DESC (default: ASC)"
// @Success     200 {array}  domain.Board "List of boards"
// @Failure     401 {object} handler.ErrorResponse "Missing or invalid access token"
// @Failure     500 {object} handler.ErrorResponse "Internal server error"
// @Router      /boards [get]
func (h *BoardHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	traceID := getTraceID(r, h.logger)

	userID, ok := getUserID(r)
	if !ok {
		logUnauthorizedAccess(h.logger, traceID)
		errorResponse(w, domain.ErrUnauthorized)
		return
	}

	params := parseListParams(r)

	boards, err := h.svc.GetAllByUserID(r.Context(), userID, params)
	if err != nil {
		logUnexpectedError(h.logger, traceID, err)
		errorResponse(w, err)
		return
	}

	logSuccess(h.logger, traceID, "all boards retrieved", "user_id", userID)
	successResponse(w, http.StatusOK, boards)
}

// GetBoardByID godoc
// @Summary     Get board by ID
// @Description Returns a specific board by its unique identifier.
// @Tags        boards
// @Produce     json
// @Security    BearerAuth
// @Param       id path string true "Board ID (UUID)"
// @Success     200 {object} domain.Board "Board data"
// @Failure     400 {object} handler.ErrorResponse "Invalid board ID format"
// @Failure     401 {object} handler.ErrorResponse "Missing or invalid access token"
// @Failure     404 {object} handler.ErrorResponse "Board not found"
// @Failure     500 {object} handler.ErrorResponse "Internal server error"
// @Router      /boards/{id} [get]
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

// UpdateBoard godoc
// @Summary     Update a board
// @Description Updates the name and/or description of a board. Only the board owner can perform this action.
// @Tags        boards
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       id   path string                  true "Board ID (UUID)"
// @Param       body body domain.UpdateBoardInput true "Updated board data"
// @Success     204 "Board updated successfully"
// @Failure     400 {object} handler.ErrorResponse "Invalid board ID format or request body"
// @Failure     401 {object} handler.ErrorResponse "Missing or invalid access token"
// @Failure     403 {object} handler.ErrorResponse "Caller is not the board owner"
// @Failure     404 {object} handler.ErrorResponse "Board not found"
// @Failure     500 {object} handler.ErrorResponse "Internal server error"
// @Router      /boards/{id} [patch]
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

// DeleteBoard godoc
// @Summary     Delete a board
// @Description Deletes a board by ID. Only the board owner can perform this action.
// @Tags        boards
// @Produce     json
// @Security    BearerAuth
// @Param       id path string true "Board ID (UUID)"
// @Success     204 "Board deleted successfully"
// @Failure     400 {object} handler.ErrorResponse "Invalid board ID format"
// @Failure     401 {object} handler.ErrorResponse "Missing or invalid access token"
// @Failure     403 {object} handler.ErrorResponse "Caller is not the board owner"
// @Failure     404 {object} handler.ErrorResponse "Board not found"
// @Failure     500 {object} handler.ErrorResponse "Internal server error"
// @Router      /boards/{id} [delete]
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
