package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/mykytakuzminov/task-manager-api/internal/auth"
	"github.com/mykytakuzminov/task-manager-api/internal/domain"
	"github.com/mykytakuzminov/task-manager-api/internal/handler/middleware"
)

type TaskHandler struct {
	svc  domain.TaskService
	auth *auth.Auth
}

func NewTaskHandler(svc domain.TaskService, auth *auth.Auth) *TaskHandler {
	return &TaskHandler{svc: svc, auth: auth}
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
	userID, ok := getUserID(r)
	if !ok {
		errorResponse(w, domain.ErrUnauthorized)
		return
	}

	columnID, err := getURLParam(r, "columnID")
	if err != nil {
		errorResponse(w, domain.ErrBadRequest)
		return
	}

	var input domain.CreateTaskInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		errorResponse(w, domain.ErrBadRequest)
		return
	}

	task, err := h.svc.Create(r.Context(), userID, columnID, &input)
	if err != nil {
		errorResponse(w, err)
		return
	}

	successResponse(w, http.StatusCreated, task)
}

func (h *TaskHandler) GetAllByColumnID(w http.ResponseWriter, r *http.Request) {
	columnID, err := getURLParam(r, "columnID")
	if err != nil {
		errorResponse(w, domain.ErrBadRequest)
		return
	}

	tasks, err := h.svc.GetAllByColumnID(r.Context(), columnID)
	if err != nil {
		errorResponse(w, err)
		return
	}

	successResponse(w, http.StatusOK, tasks)
}

func (h *TaskHandler) GetAllByUserID(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserID(r)
	if !ok {
		errorResponse(w, domain.ErrUnauthorized)
		return
	}

	tasks, err := h.svc.GetAllByUserID(r.Context(), userID)
	if err != nil {
		errorResponse(w, err)
		return
	}

	successResponse(w, http.StatusOK, tasks)
}

func (h *TaskHandler) GetAllByAssigneeID(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserID(r)
	if !ok {
		errorResponse(w, domain.ErrUnauthorized)
		return
	}

	tasks, err := h.svc.GetAllByAssigneeID(r.Context(), userID)
	if err != nil {
		errorResponse(w, err)
		return
	}

	successResponse(w, http.StatusOK, tasks)
}

func (h *TaskHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	taskID, err := getURLParam(r, "id")
	if err != nil {
		errorResponse(w, domain.ErrBadRequest)
		return
	}

	task, err := h.svc.GetByID(r.Context(), taskID)
	if err != nil {
		errorResponse(w, err)
		return
	}

	successResponse(w, http.StatusOK, task)
}

func (h *TaskHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserID(r)
	if !ok {
		errorResponse(w, domain.ErrUnauthorized)
		return
	}

	taskID, err := getURLParam(r, "id")
	if err != nil {
		errorResponse(w, domain.ErrBadRequest)
		return
	}

	var input domain.UpdateTaskInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		errorResponse(w, domain.ErrBadRequest)
		return
	}

	if err := h.svc.Update(r.Context(), taskID, userID, &input); err != nil {
		errorResponse(w, err)
		return
	}

	noContentResponse(w)
}

func (h *TaskHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserID(r)
	if !ok {
		errorResponse(w, domain.ErrUnauthorized)
		return
	}

	taskID, err := getURLParam(r, "id")
	if err != nil {
		errorResponse(w, domain.ErrBadRequest)
		return
	}

	if err := h.svc.Delete(r.Context(), taskID, userID); err != nil {
		errorResponse(w, err)
		return
	}

	noContentResponse(w)
}
