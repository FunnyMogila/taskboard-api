package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	api "taskboard-api/internal/api"
	"taskboard-api/internal/domain"
)

type UserService interface {
	Create(ctx context.Context, user domain.User) (domain.User, error)
	GetByID(ctx context.Context, id domain.UserID) (domain.User, error)
}

type UserHandler struct {
	service UserService
}

func NewUserHandler(service UserService) *UserHandler {
	return &UserHandler{
		service: service,
	}
}

func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req api.CreateUserRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	user, err := h.service.Create(r.Context(), domain.User{
		Name:  req.Name,
		Email: req.Email,
	})
	if err != nil {
		handleError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, user)
}

func (h *UserHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	user, err := h.service.GetByID(r.Context(), domain.UserID(id))
	if err != nil {
		handleError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, user)
}
