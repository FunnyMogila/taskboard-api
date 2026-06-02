package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"taskboard-api/internal/api"

	"github.com/go-chi/chi/v5"

	"taskboard-api/internal/domain"
)

type TaskService interface {
	Create(ctx context.Context, task domain.Task) (domain.Task, error)
	GetByID(ctx context.Context, id domain.TaskID) (domain.Task, error)
	List(ctx context.Context) ([]domain.Task, error)
	UpdateStatus(ctx context.Context, id domain.TaskID, newStatus domain.TaskStatus) error
	Delete(ctx context.Context, id domain.TaskID) error
	AddComment(ctx context.Context, comment domain.Comment) (domain.Comment, error)
	ListComments(ctx context.Context, taskID domain.TaskID) ([]domain.Comment, error)
}

type TaskHandler struct {
	service TaskService
}

func NewTaskHandler(service TaskService) *TaskHandler {
	return &TaskHandler{service: service}
}

func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req api.CreateTaskRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	var assigneeID *domain.UserID
	if req.AssigneeId != nil {
		id := domain.UserID(*req.AssigneeId)
		assigneeID = &id
	}

	description := ""
	if req.Description != nil {
		description = *req.Description
	}

	task, err := h.service.Create(r.Context(), domain.Task{
		ProjectID:   domain.ProjectID(req.ProjectId),
		AssigneeID:  assigneeID,
		Title:       req.Title,
		Description: description,
	})
	if err != nil {
		handleError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, task)
}

func (h *TaskHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "taskID"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid task id")
		return
	}

	task, err := h.service.GetByID(r.Context(), domain.TaskID(id))
	if err != nil {
		handleError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, task)
}

func (h *TaskHandler) List(w http.ResponseWriter, r *http.Request) {
	tasks, err := h.service.List(r.Context())
	if err != nil {
		handleError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, tasks)
}

func (h *TaskHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "taskID"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid task id")
		return
	}

	var req api.UpdateTaskStatusRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	if err := h.service.UpdateStatus(
		r.Context(),
		domain.TaskID(id),
		domain.TaskStatus(req.Status),
	); err != nil {
		handleError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"status": "updated",
	})
}

func (h *TaskHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "taskID"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid task id")
		return
	}

	if err := h.service.Delete(r.Context(), domain.TaskID(id)); err != nil {
		handleError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"status": "deleted",
	})
}

func (h *TaskHandler) AddComment(w http.ResponseWriter, r *http.Request) {
	taskID, err := strconv.ParseInt(chi.URLParam(r, "taskID"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid task id")
		return
	}

	var req api.CreateCommentRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	comment, err := h.service.AddComment(r.Context(), domain.Comment{
		TaskID:   domain.TaskID(taskID),
		AuthorID: domain.UserID(req.AuthorId),
		Text:     req.Text,
	})
	if err != nil {
		handleError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, comment)
}

func (h *TaskHandler) ListComments(w http.ResponseWriter, r *http.Request) {
	taskID, err := strconv.ParseInt(chi.URLParam(r, "taskID"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid task id")
		return
	}

	comments, err := h.service.ListComments(r.Context(), domain.TaskID(taskID))
	if err != nil {
		handleError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, comments)
}
