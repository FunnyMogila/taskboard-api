package service

import (
	"context"
	"errors"
	"testing"

	"taskboard-api/internal/domain"
	"taskboard-api/internal/errs"
)

type fakeTaskRepository struct {
	createFunc       func(ctx context.Context, task domain.Task) (domain.Task, error)
	getByIDFunc      func(ctx context.Context, id domain.TaskID) (domain.Task, error)
	listFunc         func(ctx context.Context) ([]domain.Task, error)
	updateStatusFunc func(ctx context.Context, id domain.TaskID, status domain.TaskStatus) error
	deleteFunc       func(ctx context.Context, id domain.TaskID) error
}

func (r *fakeTaskRepository) Create(ctx context.Context, task domain.Task) (domain.Task, error) {
	if r.createFunc != nil {
		return r.createFunc(ctx, task)
	}
	return task, nil
}

func (r *fakeTaskRepository) GetByID(ctx context.Context, id domain.TaskID) (domain.Task, error) {
	if r.getByIDFunc != nil {
		return r.getByIDFunc(ctx, id)
	}
	return domain.Task{}, nil
}

func (r *fakeTaskRepository) List(ctx context.Context) ([]domain.Task, error) {
	if r.listFunc != nil {
		return r.listFunc(ctx)
	}
	return nil, nil
}

func (r *fakeTaskRepository) UpdateStatus(ctx context.Context, id domain.TaskID, status domain.TaskStatus) error {
	if r.updateStatusFunc != nil {
		return r.updateStatusFunc(ctx, id, status)
	}
	return nil
}

func (r *fakeTaskRepository) Delete(ctx context.Context, id domain.TaskID) error {
	if r.deleteFunc != nil {
		return r.deleteFunc(ctx, id)
	}
	return nil
}

type fakeCommentRepository struct {
	createFunc     func(ctx context.Context, comment domain.Comment) (domain.Comment, error)
	listByTaskFunc func(ctx context.Context, taskID domain.TaskID) ([]domain.Comment, error)
}

func (r *fakeCommentRepository) Create(ctx context.Context, comment domain.Comment) (domain.Comment, error) {
	if r.createFunc != nil {
		return r.createFunc(ctx, comment)
	}
	return comment, nil
}

func (r *fakeCommentRepository) ListByTask(ctx context.Context, taskID domain.TaskID) ([]domain.Comment, error) {
	if r.listByTaskFunc != nil {
		return r.listByTaskFunc(ctx, taskID)
	}
	return nil, nil
}

func TestTaskService_Create_ClosedProject(t *testing.T) {
	projectRepo := &fakeProjectRepository{
		getByIDFunc: func(ctx context.Context, id domain.ProjectID) (domain.Project, error) {
			return domain.Project{
				ID:     id,
				Status: domain.ProjectStatusClosed,
			}, nil
		},
	}

	service := NewTaskService(
		&fakeTaskRepository{},
		&fakeCommentRepository{},
		projectRepo,
		nil,
	)

	_, err := service.Create(context.Background(), domain.Task{
		ProjectID: 1,
		Title:     "Test task",
	})

	if !errors.Is(err, errs.ErrProjectClosed) {
		t.Fatalf("expected ErrProjectClosed, got %v", err)
	}
}

func TestTaskService_Create_NotProjectMember(t *testing.T) {
	projectRepo := &fakeProjectRepository{
		getByIDFunc: func(ctx context.Context, id domain.ProjectID) (domain.Project, error) {
			return domain.Project{
				ID:     id,
				Status: domain.ProjectStatusActive,
			}, nil
		},
		isMemberFunc: func(ctx context.Context, projectID domain.ProjectID, userID domain.UserID) (bool, error) {
			return false, nil
		},
	}

	service := NewTaskService(
		&fakeTaskRepository{},
		&fakeCommentRepository{},
		projectRepo,
		nil,
	)

	assigneeID := domain.UserID(10)

	_, err := service.Create(context.Background(), domain.Task{
		ProjectID:  1,
		AssigneeID: &assigneeID,
		Title:      "Test task",
	})

	if !errors.Is(err, errs.ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestTaskService_UpdateStatus_InvalidTransition(t *testing.T) {
	taskRepo := &fakeTaskRepository{
		getByIDFunc: func(ctx context.Context, id domain.TaskID) (domain.Task, error) {
			return domain.Task{
				ID:        id,
				ProjectID: 1,
				Status:    domain.TaskStatusDone,
			}, nil
		},
	}

	projectRepo := &fakeProjectRepository{
		getByIDFunc: func(ctx context.Context, id domain.ProjectID) (domain.Project, error) {
			return domain.Project{
				ID:     id,
				Status: domain.ProjectStatusActive,
			}, nil
		},
	}

	service := NewTaskService(
		taskRepo,
		&fakeCommentRepository{},
		projectRepo,
		nil,
	)

	err := service.UpdateStatus(context.Background(), 1, domain.TaskStatusInProgress)

	if !errors.Is(err, errs.ErrInvalidTransition) {
		t.Fatalf("expected ErrInvalidTransition, got %v", err)
	}
}

func TestTaskService_AddComment_DoneTask(t *testing.T) {
	taskRepo := &fakeTaskRepository{
		getByIDFunc: func(ctx context.Context, id domain.TaskID) (domain.Task, error) {
			return domain.Task{
				ID:     id,
				Status: domain.TaskStatusDone,
			}, nil
		},
	}

	service := NewTaskService(
		taskRepo,
		&fakeCommentRepository{},
		&fakeProjectRepository{},
		nil,
	)

	_, err := service.AddComment(context.Background(), domain.Comment{
		TaskID:   1,
		AuthorID: 1,
		Text:     "Comment",
	})

	if !errors.Is(err, errs.ErrInvalidTransition) {
		t.Fatalf("expected ErrInvalidTransition, got %v", err)
	}
}
