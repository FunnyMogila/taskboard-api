package service

import (
	"context"
	"fmt"

	"taskboard-api/internal/audit"
	"taskboard-api/internal/domain"
	"taskboard-api/internal/errs"
)

type taskRepository interface {
	Create(ctx context.Context, task domain.Task) (domain.Task, error)
	GetByID(ctx context.Context, id domain.TaskID) (domain.Task, error)
	List(ctx context.Context) ([]domain.Task, error)
	UpdateStatus(ctx context.Context, id domain.TaskID, status domain.TaskStatus) error
	Delete(ctx context.Context, id domain.TaskID) error
}

type commentRepository interface {
	Create(ctx context.Context, comment domain.Comment) (domain.Comment, error)
	ListByTask(ctx context.Context, taskID domain.TaskID) ([]domain.Comment, error)
}

type TaskService struct {
	taskRepository    taskRepository
	commentRepository commentRepository
	projectRepository projectRepository
	audit             auditPublisher
}

func NewTaskService(
	taskRepository taskRepository,
	commentRepository commentRepository,
	projectRepository projectRepository,
	audit auditPublisher,
) *TaskService {
	return &TaskService{
		taskRepository:    taskRepository,
		commentRepository: commentRepository,
		projectRepository: projectRepository,
		audit:             audit,
	}
}

func (s *TaskService) Create(ctx context.Context, task domain.Task) (domain.Task, error) {
	if task.ProjectID <= 0 {
		return domain.Task{}, errs.ErrInvalidInput
	}

	if task.Title == "" {
		return domain.Task{}, errs.ErrInvalidInput
	}

	project, err := s.projectRepository.GetByID(ctx, task.ProjectID)
	if err != nil {
		return domain.Task{}, err
	}

	if project.Status == domain.ProjectStatusClosed {
		return domain.Task{}, errs.ErrProjectClosed
	}

	if task.AssigneeID != nil {
		isMember, err := s.projectRepository.IsMember(ctx, task.ProjectID, *task.AssigneeID)
		if err != nil {
			return domain.Task{}, err
		}

		if !isMember {
			return domain.Task{}, errs.ErrForbidden
		}
	}

	created, err := s.taskRepository.Create(ctx, task)
	if err != nil {
		return domain.Task{}, err
	}

	if s.audit != nil {
		s.audit.Publish(audit.Event{
			Type:    "task_created",
			Payload: created.Title,
		})
	}

	return created, nil
}

func (s *TaskService) GetByID(ctx context.Context, id domain.TaskID) (domain.Task, error) {
	if id <= 0 {
		return domain.Task{}, errs.ErrInvalidInput
	}

	return s.taskRepository.GetByID(ctx, id)
}

func (s *TaskService) List(ctx context.Context) ([]domain.Task, error) {
	return s.taskRepository.List(ctx)
}

func (s *TaskService) UpdateStatus(
	ctx context.Context,
	id domain.TaskID,
	newStatus domain.TaskStatus,
) error {
	if id <= 0 {
		return errs.ErrInvalidInput
	}

	task, err := s.taskRepository.GetByID(ctx, id)
	if err != nil {
		return err
	}

	project, err := s.projectRepository.GetByID(ctx, task.ProjectID)
	if err != nil {
		return err
	}

	if task.Status == newStatus {
		return nil
	}

	if project.Status == domain.ProjectStatusClosed {
		return errs.ErrProjectClosed
	}

	if !isAllowedTaskStatusTransition(task.Status, newStatus) {
		return errs.ErrInvalidTransition
	}

	if err := s.taskRepository.UpdateStatus(ctx, id, newStatus); err != nil {
		return err
	}

	if s.audit != nil {
		s.audit.Publish(audit.Event{
			Type:    "task_status_updated",
			Payload: fmt.Sprintf("task_id=%d status=%s", id, newStatus),
		})
	}

	return nil
}

func (s *TaskService) Delete(ctx context.Context, id domain.TaskID) error {
	if id <= 0 {
		return errs.ErrInvalidInput
	}

	return s.taskRepository.Delete(ctx, id)
}

func (s *TaskService) AddComment(
	ctx context.Context,
	comment domain.Comment,
) (domain.Comment, error) {
	if comment.TaskID <= 0 || comment.AuthorID <= 0 || comment.Text == "" {
		return domain.Comment{}, errs.ErrInvalidInput
	}

	task, err := s.taskRepository.GetByID(ctx, comment.TaskID)
	if err != nil {
		return domain.Comment{}, err
	}

	if task.Status == domain.TaskStatusDone || task.Status == domain.TaskStatusCancelled {
		return domain.Comment{}, errs.ErrInvalidTransition
	}

	created, err := s.commentRepository.Create(ctx, comment)
	if err != nil {
		return domain.Comment{}, err
	}

	if s.audit != nil {
		s.audit.Publish(audit.Event{
			Type:    "comment_created",
			Payload: fmt.Sprintf("task_id=%d", comment.TaskID),
		})
	}

	return created, nil
}

func (s *TaskService) ListComments(
	ctx context.Context,
	taskID domain.TaskID,
) ([]domain.Comment, error) {
	if taskID <= 0 {
		return nil, errs.ErrInvalidInput
	}

	return s.commentRepository.ListByTask(ctx, taskID)
}

func isAllowedTaskStatusTransition(
	currentStatus domain.TaskStatus,
	newStatus domain.TaskStatus,
) bool {
	allowedTransitions := map[domain.TaskStatus][]domain.TaskStatus{
		domain.TaskStatusNew: {
			domain.TaskStatusInProgress,
			domain.TaskStatusCancelled,
		},
		domain.TaskStatusInProgress: {
			domain.TaskStatusDone,
			domain.TaskStatusCancelled,
		},
		domain.TaskStatusDone:      {},
		domain.TaskStatusCancelled: {},
	}

	allowedStatuses, ok := allowedTransitions[currentStatus]
	if !ok {
		return false
	}

	for _, status := range allowedStatuses {
		if status == newStatus {
			return true
		}
	}

	return false
}
