package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/mykytakuzminov/sprintly-api/internal/auth"
	"github.com/mykytakuzminov/sprintly-api/internal/domain"
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

	r.Get("/{id}", h.GetByID)
	r.Patch("/{id}", h.Update)
	r.Delete("/{id}", h.Delete)

	return r
}

func (h *TaskHandler) ColumnRoutes() chi.Router {
	r := chi.NewRouter()

	r.Post("/", h.Create)
	r.Get("/", h.GetAllByColumnID)

	return r
}

func (h *TaskHandler) UserRoutes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", h.GetAllByUserID)
	r.Get("/assigned", h.GetAllByAssigneeID)

	return r
}

// CreateTask godoc
// @Summary     Create a task
// @Description Creates a new task inside a column. Only the owner of the parent board can create tasks.
// @Tags        tasks
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       columnID path string                 true "Column ID (UUID)"
// @Param       body     body domain.CreateTaskInput true "Task data"
// @Success     201 {object} domain.Task "Task created successfully"
// @Failure     400 {object} handler.ErrorResponse "Invalid column ID format, request body, or validation error"
// @Failure     401 {object} handler.ErrorResponse "Missing or invalid access token"
// @Failure     403 {object} handler.ErrorResponse "Caller is not the board owner"
// @Failure     404 {object} handler.ErrorResponse "Column not found"
// @Failure     500 {object} handler.ErrorResponse "Internal server error"
// @Router      /columns/{columnID}/tasks [post]
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

// GetAllTasksByColumnID godoc
// @Summary     Get all tasks in a column
// @Description Returns all tasks belonging to the specified column. Supports pagination and sorting.
// @Tags        tasks
// @Produce     json
// @Security    BearerAuth
// @Param       columnID path   string true  "Column ID (UUID)"
// @Param       limit    query  int    false "Number of results per page (default: 20, max: 100)"
// @Param       offset   query  int    false "Number of results to skip (default: 0)"
// @Param       sort     query  string false "Field to sort by: name, due_date, created_at, updated_at (default: created_at)"
// @Param       order    query  string false "Sort direction: ASC or DESC (default: ASC)"
// @Success     200 {array}  domain.Task "List of tasks"
// @Failure     400 {object} handler.ErrorResponse "Invalid column ID format"
// @Failure     401 {object} handler.ErrorResponse "Missing or invalid access token"
// @Failure     500 {object} handler.ErrorResponse "Internal server error"
// @Router      /columns/{columnID}/tasks [get]
func (h *TaskHandler) GetAllByColumnID(w http.ResponseWriter, r *http.Request) {
	traceID := getTraceID(r, h.logger)

	columnID, err := getURLParam(r, "columnID")
	if err != nil {
		logInvalidBody(h.logger, traceID, err)
		errorResponse(w, domain.ErrBadRequest)
		return
	}

	params := parseListParams(r)

	tasks, err := h.svc.GetAllByColumnID(r.Context(), columnID, params)
	if err != nil {
		logUnexpectedError(h.logger, traceID, err)
		errorResponse(w, err)
		return
	}

	logSuccess(h.logger, traceID, "all tasks retrieved", "column_id", columnID)
	successResponse(w, http.StatusOK, tasks)
}

// GetAllTasksByUserID godoc
// @Summary     Get all tasks created by current user
// @Description Returns all tasks where owner_id matches the authenticated user. Supports pagination and sorting.
// @Tags        tasks
// @Produce     json
// @Security    BearerAuth
// @Param       limit  query int    false "Number of results per page (default: 20, max: 100)"
// @Param       offset query int    false "Number of results to skip (default: 0)"
// @Param       sort   query string false "Field to sort by: name, due_date, created_at, updated_at (default: created_at)"
// @Param       order  query string false "Sort direction: ASC or DESC (default: ASC)"
// @Success     200 {array}  domain.Task "List of tasks"
// @Failure     401 {object} handler.ErrorResponse "Missing or invalid access token"
// @Failure     500 {object} handler.ErrorResponse "Internal server error"
// @Router      /users/me/tasks [get]
func (h *TaskHandler) GetAllByUserID(w http.ResponseWriter, r *http.Request) {
	traceID := getTraceID(r, h.logger)

	userID, ok := getUserID(r)
	if !ok {
		logUnauthorizedAccess(h.logger, traceID)
		errorResponse(w, domain.ErrUnauthorized)
		return
	}

	params := parseListParams(r)

	tasks, err := h.svc.GetAllByUserID(r.Context(), userID, params)
	if err != nil {
		logUnexpectedError(h.logger, traceID, err)
		errorResponse(w, err)
		return
	}

	logSuccess(h.logger, traceID, "all tasks retrieved", "user_id", userID)
	successResponse(w, http.StatusOK, tasks)
}

// GetAllTasksByAssigneeID godoc
// @Summary     Get all tasks assigned to current user
// @Description Returns all tasks where assignee_id matches the authenticated user. Supports pagination and sorting.
// @Tags        tasks
// @Produce     json
// @Security    BearerAuth
// @Param       limit  query int    false "Number of results per page (default: 20, max: 100)"
// @Param       offset query int    false "Number of results to skip (default: 0)"
// @Param       sort   query string false "Field to sort by: name, due_date, created_at, updated_at (default: created_at)"
// @Param       order  query string false "Sort direction: ASC or DESC (default: ASC)"
// @Success     200 {array}  domain.Task "List of assigned tasks"
// @Failure     401 {object} handler.ErrorResponse "Missing or invalid access token"
// @Failure     500 {object} handler.ErrorResponse "Internal server error"
// @Router      /users/me/tasks/assigned [get]
func (h *TaskHandler) GetAllByAssigneeID(w http.ResponseWriter, r *http.Request) {
	traceID := getTraceID(r, h.logger)

	userID, ok := getUserID(r)
	if !ok {
		logUnauthorizedAccess(h.logger, traceID)
		errorResponse(w, domain.ErrUnauthorized)
		return
	}

	params := parseListParams(r)

	tasks, err := h.svc.GetAllByAssigneeID(r.Context(), userID, params)
	if err != nil {
		logUnexpectedError(h.logger, traceID, err)
		errorResponse(w, err)
		return
	}

	logSuccess(h.logger, traceID, "all tasks retrieved", "assignee_id", userID)
	successResponse(w, http.StatusOK, tasks)
}

// GetTaskByID godoc
// @Summary     Get task by ID
// @Description Returns a specific task by its unique identifier.
// @Tags        tasks
// @Produce     json
// @Security    BearerAuth
// @Param       id path string true "Task ID (UUID)"
// @Success     200 {object} domain.Task "Task data"
// @Failure     400 {object} handler.ErrorResponse "Invalid task ID format"
// @Failure     401 {object} handler.ErrorResponse "Missing or invalid access token"
// @Failure     404 {object} handler.ErrorResponse "Task not found"
// @Failure     500 {object} handler.ErrorResponse "Internal server error"
// @Router      /tasks/{id} [get]
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

// UpdateTask godoc
// @Summary     Update a task
// @Description Updates task fields including column, assignee, name, description and due date.
// @Description The column_id can be used to move the task to another column.
// @Description Only the task owner (board owner) can update it.
// @Tags        tasks
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       id   path string                 true "Task ID (UUID)"
// @Param       body body domain.UpdateTaskInput true "Updated task data"
// @Success     204 "Task updated successfully"
// @Failure     400 {object} handler.ErrorResponse "Invalid task ID format, request body, or validation error"
// @Failure     401 {object} handler.ErrorResponse "Missing or invalid access token"
// @Failure     403 {object} handler.ErrorResponse "Caller is not the task owner"
// @Failure     404 {object} handler.ErrorResponse "Task not found"
// @Failure     500 {object} handler.ErrorResponse "Internal server error"
// @Router      /tasks/{id} [patch]
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

// DeleteTask godoc
// @Summary     Delete a task
// @Description Deletes a task by ID. Only the task owner (board owner) can perform this action.
// @Tags        tasks
// @Produce     json
// @Security    BearerAuth
// @Param       id path string true "Task ID (UUID)"
// @Success     204 "Task deleted successfully"
// @Failure     400 {object} handler.ErrorResponse "Invalid task ID format"
// @Failure     401 {object} handler.ErrorResponse "Missing or invalid access token"
// @Failure     403 {object} handler.ErrorResponse "Caller is not the task owner"
// @Failure     404 {object} handler.ErrorResponse "Task not found"
// @Failure     500 {object} handler.ErrorResponse "Internal server error"
// @Router      /tasks/{id} [delete]
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
