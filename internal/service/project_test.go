package service

import (
	"context"
	"errors"
	"taskboard-api/internal/domain"
	"taskboard-api/internal/errs"
	"testing"
)

type fakeProjectRepository struct {
	createFunc    func(ctx context.Context, project domain.Project) (domain.Project, error)
	getByIDFunc   func(ctx context.Context, id domain.ProjectID) (domain.Project, error)
	listFunc      func(ctx context.Context) ([]domain.Project, error)
	closeFunc     func(ctx context.Context, id domain.ProjectID) error
	isMemberFunc  func(ctx context.Context, projectID domain.ProjectID, userID domain.UserID) (bool, error)
	addMemberFunc func(
		ctx context.Context,
		projectID domain.ProjectID,
		userID domain.UserID,
		role domain.ProjectRole,
	) error
}

func (r *fakeProjectRepository) Create(
	ctx context.Context,
	project domain.Project,
) (domain.Project, error) {
	return r.createFunc(ctx, project)
}

func (r *fakeProjectRepository) GetByID(
	ctx context.Context,
	id domain.ProjectID,
) (domain.Project, error) {
	return r.getByIDFunc(ctx, id)
}

func (r *fakeProjectRepository) List(
	ctx context.Context,
) ([]domain.Project, error) {
	return r.listFunc(ctx)
}

func (r *fakeProjectRepository) Close(
	ctx context.Context,
	id domain.ProjectID,
) error {
	return r.closeFunc(ctx, id)
}

func (r *fakeProjectRepository) AddMember(
	ctx context.Context,
	projectID domain.ProjectID,
	userID domain.UserID,
	role domain.ProjectRole,
) error {
	return r.addMemberFunc(ctx, projectID, userID, role)
}

func (r *fakeProjectRepository) IsMember(
	ctx context.Context,
	projectID domain.ProjectID,
	userID domain.UserID,
) (bool, error) {
	if r.isMemberFunc != nil {
		return r.isMemberFunc(ctx, projectID, userID)
	}

	return true, nil
}

func TestProjectService_Create_EmptyName(t *testing.T) {
	repo := &fakeProjectRepository{}
	service := NewProjectService(repo, nil)

	_, err := service.Create(
		context.Background(),
		domain.Project{},
	)

	if !errors.Is(err, errs.ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestProjectService_Create_Success(t *testing.T) {
	repo := &fakeProjectRepository{
		createFunc: func(
			ctx context.Context,
			project domain.Project,
		) (domain.Project, error) {

			project.ID = 1
			return project, nil
		},
	}

	service := NewProjectService(repo, nil)

	project, err := service.Create(
		context.Background(),
		domain.Project{
			Name: "TaskBoard",
		},
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if project.ID != 1 {
		t.Fatalf("expected ID=1, got %d", project.ID)
	}
}

func TestProjectService_GetByID_InvalidID(t *testing.T) {
	repo := &fakeProjectRepository{}
	service := NewProjectService(repo, nil)

	_, err := service.GetByID(context.Background(), 0)

	if !errors.Is(err, errs.ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestProjectService_Close_InvalidID(t *testing.T) {
	repo := &fakeProjectRepository{}
	service := NewProjectService(repo, nil)

	err := service.Close(context.Background(), 0)

	if !errors.Is(err, errs.ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestProjectService_AddMember_InvalidUser(t *testing.T) {
	repo := &fakeProjectRepository{}
	service := NewProjectService(repo, nil)

	err := service.AddMember(
		context.Background(),
		1,
		0,
		domain.ProjectRole("member"),
	)

	if !errors.Is(err, errs.ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}
