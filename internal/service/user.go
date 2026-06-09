package service

import (
	"context"

	"taskboard-api/internal/audit"
	"taskboard-api/internal/domain"
	"taskboard-api/internal/errs"
)

type userRepository interface {
	Create(ctx context.Context, user domain.User) (domain.User, error)
	GetByID(ctx context.Context, id domain.UserID) (domain.User, error)
}

type UserService struct {
	repository userRepository
	audit      auditPublisher
}

func NewUserService(repository userRepository, audit auditPublisher) *UserService {
	return &UserService{
		repository: repository,
		audit:      audit,
	}
}

func (s *UserService) Create(ctx context.Context, user domain.User) (domain.User, error) {
	if user.Name == "" {
		return domain.User{}, errs.ErrInvalidInput
	}

	if user.Email == "" {
		return domain.User{}, errs.ErrInvalidInput
	}

	created, err := s.repository.Create(ctx, user)
	if err != nil {
		return domain.User{}, err
	}

	if s.audit != nil {
		s.audit.Publish(audit.Event{
			Type:    "user_created",
			Payload: created.Email,
		})
	}

	return created, nil
}

func (s *UserService) GetByID(ctx context.Context, id domain.UserID) (domain.User, error) {
	if id <= 0 {
		return domain.User{}, errs.ErrInvalidInput
	}

	return s.repository.GetByID(ctx, id)
}
