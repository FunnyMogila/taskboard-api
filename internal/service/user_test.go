package service

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"taskboard-api/internal/domain"
	"taskboard-api/internal/errs"
	"taskboard-api/internal/service/mocks"
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
	expected.ID = domain.UserID(1)

	repo.EXPECT().
		Create(gomock.Any(), input).
		Return(expected, nil)

	service := NewUserService(repo, nil)

	user, err := service.Create(context.Background(), input)

	assert.NoError(t, err)
	assert.Equal(t, domain.UserID(1), user.ID)
}

func TestUserService_Create_EmptyName(t *testing.T) {
	repo := &fakeUserRepository{}
	service := NewUserService(repo, nil)

	_, err := service.Create(context.Background(), domain.User{
		Name:  "",
		Email: "test@test.com",
	})

	assert.ErrorIs(t, err, errs.ErrInvalidInput)
}

func TestUserService_Create_EmptyEmail(t *testing.T) {
	repo := &fakeUserRepository{}
	service := NewUserService(repo, nil)

	_, err := service.Create(context.Background(), domain.User{
		Name:  "Miroslav",
		Email: "",
	})

	assert.ErrorIs(t, err, errs.ErrInvalidInput)
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

	user, err := service.GetByID(context.Background(), domain.UserID(1))

	assert.NoError(t, err)
	assert.Equal(t, domain.UserID(1), user.ID)
}

func TestUserService_GetByID_InvalidID(t *testing.T) {
	repo := &fakeUserRepository{}
	service := NewUserService(repo, nil)

	_, err := service.GetByID(context.Background(), 0)

	assert.ErrorIs(t, err, errs.ErrInvalidInput)
}
