package service

import (
	"context"
	"errors"
	"taskboard-api/internal/mocks"
	"testing"

	"taskboard-api/internal/domain"
	"taskboard-api/internal/errs"

	"github.com/golang/mock/gomock"
)

type fakeUserRepository struct {
	createFunc  func(ctx context.Context, user domain.User) (domain.User, error)
	getByIDFunc func(ctx context.Context, id domain.UserID) (domain.User, error)
}

func (r *fakeUserRepository) Create(ctx context.Context, user domain.User) (domain.User, error) {
	return r.createFunc(ctx, user)
}

func (r *fakeUserRepository) GetByID(ctx context.Context, id domain.UserID) (domain.User, error) {
	return r.getByIDFunc(ctx, id)
}

func TestUserService_Create_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockUserRepository(ctrl)

	input := domain.User{
		Name:  "Miroslav",
		Email: "miroslav@test.com",
	}

	expected := input
	expected.ID = 1

	repo.EXPECT().
		Create(gomock.Any(), input).
		Return(expected, nil)

	service := NewUserService(repo, nil)

	user, err := service.Create(context.Background(), input)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if user.ID != 1 {
		t.Fatalf("expected user ID 1, got %d", user.ID)
	}
}

func TestUserService_Create_EmptyName(t *testing.T) {
	repo := &fakeUserRepository{}
	service := NewUserService(repo, nil)

	_, err := service.Create(context.Background(), domain.User{
		Name:  "",
		Email: "test@test.com",
	})

	if !errors.Is(err, errs.ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestUserService_Create_EmptyEmail(t *testing.T) {
	repo := &fakeUserRepository{}
	service := NewUserService(repo, nil)

	_, err := service.Create(context.Background(), domain.User{
		Name:  "Miroslav",
		Email: "",
	})

	if !errors.Is(err, errs.ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestUserService_GetByID_Success(t *testing.T) {
	repo := &fakeUserRepository{
		getByIDFunc: func(ctx context.Context, id domain.UserID) (domain.User, error) {
			return domain.User{
				ID:    id,
				Name:  "Miroslav",
				Email: "miroslav@test.com",
			}, nil
		},
	}

	service := NewUserService(repo, nil)

	user, err := service.GetByID(context.Background(), 1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if user.ID != 1 {
		t.Fatalf("expected user ID 1, got %d", user.ID)
	}
}

func TestUserService_GetByID_InvalidID(t *testing.T) {
	repo := &fakeUserRepository{}
	service := NewUserService(repo, nil)

	_, err := service.GetByID(context.Background(), 0)

	if !errors.Is(err, errs.ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}
