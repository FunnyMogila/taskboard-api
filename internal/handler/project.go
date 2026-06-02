package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"taskboard-api/internal/api"

	"taskboard-api/internal/domain"

	"github.com/go-chi/chi/v5"
)

type ProjectService interface {
	Create(ctx context.Context, project domain.Project) (domain.Project, error)
	GetByID(ctx context.Context, id domain.ProjectID) (domain.Project, error)
	List(ctx context.Context) ([]domain.Project, error)
	Close(ctx context.Context, id domain.ProjectID) error
	AddMember(ctx context.Context, projectID domain.ProjectID, userID domain.UserID, role domain.ProjectRole) error
}

type ProjectHandler struct {
	service ProjectService
}

func NewProjectHandler(service ProjectService) *ProjectHandler {
	return &ProjectHandler{service: service}
}

func (h *ProjectHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req api.CreateProjectRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	description := ""
	if req.Description != nil {
		description = *req.Description
	}

	project, err := h.service.Create(r.Context(), domain.Project{
		Name:        req.Name,
		Description: description,
	})
	if err != nil {
		handleError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, project)
}

func (h *ProjectHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "projectID"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid project id")
		return
	}

	project, err := h.service.GetByID(r.Context(), domain.ProjectID(id))
	if err != nil {
		handleError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, project)
}

func (h *ProjectHandler) List(w http.ResponseWriter, r *http.Request) {
	projects, err := h.service.List(r.Context())
	if err != nil {
		handleError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, projects)
}

func (h *ProjectHandler) Close(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "projectID"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid project id")
		return
	}

	if err := h.service.Close(r.Context(), domain.ProjectID(id)); err != nil {
		handleError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"status": "closed",
	})
}

func (h *ProjectHandler) AddMember(w http.ResponseWriter, r *http.Request) {
	projectID, err := strconv.ParseInt(chi.URLParam(r, "projectID"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid project id")
		return
	}

	var req api.AddProjectMemberRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	if err := h.service.AddMember(
		r.Context(),
		domain.ProjectID(projectID),
		domain.UserID(req.UserId),
		domain.ProjectRole(req.Role),
	); err != nil {
		handleError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{
		"status": "member added",
	})
}
