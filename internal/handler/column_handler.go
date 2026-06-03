package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/mykytakuzminov/task-manager-api/internal/auth"
	"github.com/mykytakuzminov/task-manager-api/internal/domain"
	"github.com/mykytakuzminov/task-manager-api/internal/handler/middleware"
)

type ColumnHandler struct {
	svc  domain.ColumnService
	auth *auth.Auth
}

func NewColumnHandler(svc domain.ColumnService, auth *auth.Auth) *ColumnHandler {
	return &ColumnHandler{svc: svc, auth: auth}
}

func (h *ColumnHandler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.AuthMiddleware(h.auth))

	r.Get("/{id}", h.GetByID)
	r.Patch("/{id}", h.Update)
	r.Delete("/{id}", h.Delete)

	return r
}

func (h *ColumnHandler) BoardRoutes() chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.AuthMiddleware(h.auth))

	r.Post("/", h.Create)
	r.Get("/", h.GetAll)

	return r
}

func (h *ColumnHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserID(r)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	boardID, err := getURLParam(r, "boardID")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var input domain.CreateColumnInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	column, err := h.svc.Create(r.Context(), userID, boardID, &input)
	if err != nil {
		errorResponse(w, err)
		return
	}

	jsonResponse(w, http.StatusCreated, column)
}

func (h *ColumnHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	boardID, err := getURLParam(r, "boardID")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	columns, err := h.svc.GetAllByBoardID(r.Context(), boardID)
	if err != nil {
		errorResponse(w, err)
		return
	}

	jsonResponse(w, http.StatusOK, columns)
}

func (h *ColumnHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	columnID, err := getURLParam(r, "id")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	column, err := h.svc.GetByID(r.Context(), columnID)
	if err != nil {
		errorResponse(w, err)
		return
	}

	jsonResponse(w, http.StatusOK, column)
}

func (h *ColumnHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserID(r)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	columnID, err := getURLParam(r, "id")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var input domain.UpdateColumnInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := h.svc.Update(r.Context(), columnID, userID, &input); err != nil {
		errorResponse(w, err)
		return
	}

	jsonResponse(w, http.StatusNoContent, nil)
}

func (h *ColumnHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserID(r)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	columnID, err := getURLParam(r, "id")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := h.svc.Delete(r.Context(), columnID, userID); err != nil {
		errorResponse(w, err)
		return
	}

	jsonResponse(w, http.StatusNoContent, nil)
}
