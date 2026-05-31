package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/mykytakuzminov/task-manager-api/internal/auth"
	"github.com/mykytakuzminov/task-manager-api/internal/domain"
	"github.com/mykytakuzminov/task-manager-api/internal/handler/middleware"
)

type BoardHandler struct {
	svc  domain.BoardService
	auth *auth.Auth
}

func NewBoardHandler(svc domain.BoardService, auth *auth.Auth) *BoardHandler {
	return &BoardHandler{svc: svc, auth: auth}
}

func (h *BoardHandler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.AuthMiddleware(h.auth))

	r.Post("/", h.Create)
	r.Get("/", h.GetAll)
	r.Get("/{id}", h.GetByID)
	r.Patch("/{id}", h.Update)
	r.Delete("/{id}", h.Delete)

	return r
}

func (h *BoardHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserID(r)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var input domain.CreateBoardInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	board, err := h.svc.Create(r.Context(), userID, &input)
	if err != nil {
		errorHandler(w, err)
		return
	}

	jsonResponse(w, http.StatusCreated, board)
}

func (h *BoardHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserID(r)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	boards, err := h.svc.GetAllByUserID(r.Context(), userID)
	if err != nil {
		errorHandler(w, err)
		return
	}

	jsonResponse(w, http.StatusOK, boards)
}

func (h *BoardHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	boardID, err := getURLParam(r, "id")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	board, err := h.svc.GetByID(r.Context(), boardID)
	if err != nil {
		errorHandler(w, err)
		return
	}

	jsonResponse(w, http.StatusOK, board)
}

func (h *BoardHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	boardID, err := getURLParam(r, "id")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var input domain.UpdateBoardInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := h.svc.Update(r.Context(), boardID, userID, &input); err != nil {
		errorHandler(w, err)
		return
	}

	jsonResponse(w, http.StatusNoContent, nil)
}

func (h *BoardHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserID(r)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	boardID, err := getURLParam(r, "id")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := h.svc.Delete(r.Context(), boardID, userID); err != nil {
		errorHandler(w, err)
		return
	}

	jsonResponse(w, http.StatusNoContent, nil)
}
