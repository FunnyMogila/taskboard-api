package service

import (
	"context"
	"fmt"
	"taskboard-api/internal/audit"
	"taskboard-api/internal/domain"
	"taskboard-api/internal/errs"
)

type ProjectRepository interface {
	Create(ctx context.Context, project domain.Project) (domain.Project, error)
	GetByID(ctx context.Context, id domain.ProjectID) (domain.Project, error)
	List(ctx context.Context) ([]domain.Project, error)
	Close(ctx context.Context, id domain.ProjectID) error

	AddMember(
		ctx context.Context,
		projectID domain.ProjectID,
		userID domain.UserID,
		role domain.ProjectRole,
	) error

	IsMember(
		ctx context.Context,
		projectID domain.ProjectID,
		userID domain.UserID,
	) (bool, error)
}

type ProjectService struct {
	repository ProjectRepository
	audit      AuditPublisher
}

func NewProjectService(repository ProjectRepository, audit AuditPublisher) *ProjectService {
	return &ProjectService{
		repository: repository,
		audit:      audit,
	}
}

func (s *ProjectService) Create(
	ctx context.Context,
	project domain.Project,
) (domain.Project, error) {

	if project.Name == "" {
		return domain.Project{}, errs.ErrInvalidInput
	}

	created, err := s.repository.Create(ctx, project)
	if err != nil {
		return domain.Project{}, err
	}

	if s.audit != nil {
		s.audit.Publish(audit.Event{
			Type:    "project_created",
			Payload: created.Name,
		})
	}

	return created, nil
}

func (s *ProjectService) GetByID(
	ctx context.Context,
	id domain.ProjectID,
) (domain.Project, error) {

	if id <= 0 {
		return domain.Project{}, errs.ErrInvalidInput
	}

	return s.repository.GetByID(ctx, id)
}

func (s *ProjectService) List(
	ctx context.Context,
) ([]domain.Project, error) {

	return s.repository.List(ctx)
}

func (s *ProjectService) Close(
	ctx context.Context,
	id domain.ProjectID,
) error {

	if id <= 0 {
		return errs.ErrInvalidInput
	}

	if err := s.repository.Close(ctx, id); err != nil {
		return err
	}

	if s.audit != nil {
		s.audit.Publish(audit.Event{
			Type:    "project_closed",
			Payload: fmt.Sprintf("project_id=%d", id),
		})
	}

	return nil
}

func (s *ProjectService) AddMember(
	ctx context.Context,
	projectID domain.ProjectID,
	userID domain.UserID,
	role domain.ProjectRole,
) error {

	if projectID <= 0 {
		return errs.ErrInvalidInput
	}

	if userID <= 0 {
		return errs.ErrInvalidInput
	}

	return s.repository.AddMember(
		ctx,
		projectID,
		userID,
		role,
	)
}
