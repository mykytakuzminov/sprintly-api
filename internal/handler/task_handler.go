package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/mykytakuzminov/task-manager-api/internal/auth"
	"github.com/mykytakuzminov/task-manager-api/internal/domain"
	"github.com/mykytakuzminov/task-manager-api/internal/handler/middleware"
	"go.uber.org/zap"
)

type TaskHandler struct {
	svc    domain.TaskService
	auth   *auth.Auth
	logger *zap.SugaredLogger
}

func NewTaskHandler(
	svc domain.TaskService,
	auth *auth.Auth,
	logger *zap.SugaredLogger,
) *TaskHandler {
	return &TaskHandler{svc: svc, auth: auth, logger: logger}
}

func (h *TaskHandler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.AuthMiddleware(h.auth))

	r.Get("/{id}", h.GetByID)
	r.Patch("/{id}", h.Update)
	r.Delete("/{id}", h.Delete)

	return r
}

func (h *TaskHandler) ColumnRoutes() chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.AuthMiddleware(h.auth))

	r.Post("/", h.Create)
	r.Get("/", h.GetAllByColumnID)

	return r
}

func (h *TaskHandler) UserRoutes() chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.AuthMiddleware(h.auth))

	r.Get("/", h.GetAllByUserID)
	r.Get("/assigned", h.GetAllByAssigneeID)

	return r
}

func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	traceID := getTraceID(r, h.logger)

	userID, ok := getUserID(r)
	if !ok {
		logUnauthorizedAccess(h.logger, traceID)
		errorResponse(w, domain.ErrUnauthorized)
		return
	}

	columnID, err := getURLParam(r, "columnID")
	if err != nil {
		logInvalidBody(h.logger, traceID, err)
		errorResponse(w, domain.ErrBadRequest)
		return
	}

	var input domain.CreateTaskInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		logInvalidBody(h.logger, traceID, err)
		errorResponse(w, domain.ErrBadRequest)
		return
	}

	task, err := h.svc.Create(r.Context(), userID, columnID, &input)
	if err != nil {
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

	logSuccess(h.logger, traceID, "task created", "user_id", userID, "task_id", task.ID)
	successResponse(w, http.StatusCreated, task)
}

func (h *TaskHandler) GetAllByColumnID(w http.ResponseWriter, r *http.Request) {
	traceID := getTraceID(r, h.logger)

	columnID, err := getURLParam(r, "columnID")
	if err != nil {
		logInvalidBody(h.logger, traceID, err)
		errorResponse(w, domain.ErrBadRequest)
		return
	}

	tasks, err := h.svc.GetAllByColumnID(r.Context(), columnID)
	if err != nil {
		logUnexpectedError(h.logger, traceID, err)
		errorResponse(w, err)
		return
	}

	logSuccess(h.logger, traceID, "all tasks retrieved", "column_id", columnID)
	successResponse(w, http.StatusOK, tasks)
}

func (h *TaskHandler) GetAllByUserID(w http.ResponseWriter, r *http.Request) {
	traceID := getTraceID(r, h.logger)

	userID, ok := getUserID(r)
	if !ok {
		logUnauthorizedAccess(h.logger, traceID)
		errorResponse(w, domain.ErrUnauthorized)
		return
	}

	tasks, err := h.svc.GetAllByUserID(r.Context(), userID)
	if err != nil {
		logUnexpectedError(h.logger, traceID, err)
		errorResponse(w, err)
		return
	}

	logSuccess(h.logger, traceID, "all tasks retrieved", "user_id", userID)
	successResponse(w, http.StatusOK, tasks)
}

func (h *TaskHandler) GetAllByAssigneeID(w http.ResponseWriter, r *http.Request) {
	traceID := getTraceID(r, h.logger)

	userID, ok := getUserID(r)
	if !ok {
		logUnauthorizedAccess(h.logger, traceID)
		errorResponse(w, domain.ErrUnauthorized)
		return
	}

	tasks, err := h.svc.GetAllByAssigneeID(r.Context(), userID)
	if err != nil {
		logUnexpectedError(h.logger, traceID, err)
		errorResponse(w, err)
		return
	}

	logSuccess(h.logger, traceID, "all tasks retrieved", "assignee_id", userID)
	successResponse(w, http.StatusOK, tasks)
}

func (h *TaskHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	traceID := getTraceID(r, h.logger)

	taskID, err := getURLParam(r, "id")
	if err != nil {
		logInvalidBody(h.logger, traceID, err)
		errorResponse(w, domain.ErrBadRequest)
		return
	}

	task, err := h.svc.GetByID(r.Context(), taskID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			logWarn(h.logger, traceID, "task not found", err)
		} else {
			logUnexpectedError(h.logger, traceID, err)
		}
		errorResponse(w, err)
		return
	}

	logSuccess(h.logger, traceID, "task retrieved", "task_id", taskID)
	successResponse(w, http.StatusOK, task)
}

func (h *TaskHandler) Update(w http.ResponseWriter, r *http.Request) {
	traceID := getTraceID(r, h.logger)

	userID, ok := getUserID(r)
	if !ok {
		logUnauthorizedAccess(h.logger, traceID)
		errorResponse(w, domain.ErrUnauthorized)
		return
	}

	taskID, err := getURLParam(r, "id")
	if err != nil {
		logInvalidBody(h.logger, traceID, err)
		errorResponse(w, domain.ErrBadRequest)
		return
	}

	var input domain.UpdateTaskInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		logInvalidBody(h.logger, traceID, err)
		errorResponse(w, domain.ErrBadRequest)
		return
	}

	if err := h.svc.Update(r.Context(), taskID, userID, &input); err != nil {
		switch {
		case errors.Is(err, domain.ErrBadRequest):
			logInvalidBody(h.logger, traceID, err)
		case errors.Is(err, domain.ErrNotFound):
			logWarn(h.logger, traceID, "task owner not found", err)
		case errors.Is(err, domain.ErrForbidden):
			logWarn(h.logger, traceID, "access denied", err)
		default:
			logUnexpectedError(h.logger, traceID, err)
		}
		errorResponse(w, err)
		return
	}

	logSuccess(h.logger, traceID, "task updated", "user_id", userID, "task_id", taskID)
	noContentResponse(w)
}

func (h *TaskHandler) Delete(w http.ResponseWriter, r *http.Request) {
	traceID := getTraceID(r, h.logger)

	userID, ok := getUserID(r)
	if !ok {
		logUnauthorizedAccess(h.logger, traceID)
		errorResponse(w, domain.ErrUnauthorized)
		return
	}

	taskID, err := getURLParam(r, "id")
	if err != nil {
		logInvalidBody(h.logger, traceID, err)
		errorResponse(w, domain.ErrBadRequest)
		return
	}

	if err := h.svc.Delete(r.Context(), taskID, userID); err != nil {
		switch {
		case errors.Is(err, domain.ErrNotFound):
			logWarn(h.logger, traceID, "task owner not found", err)
		case errors.Is(err, domain.ErrForbidden):
			logWarn(h.logger, traceID, "access denied", err)
		default:
			logUnexpectedError(h.logger, traceID, err)
		}
		errorResponse(w, err)
		return
	}

	logSuccess(h.logger, traceID, "task deleted", "user_id", userID, "task_id", taskID)
	noContentResponse(w)
}
