// Package service holds business rules. Handlers stay thin; repositories stay
// dumb; everything interesting happens here, unit-testable with a fake
// repository.
package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/Wi1low/go-web-template/internal/model"
	"github.com/Wi1low/go-web-template/internal/repository"
)

// Business errors surfaced to handlers.
var (
	ErrUserNotFound  = errors.New("user not found")
	ErrEmailTaken    = errors.New("email already in use")
	ErrInvalidPaging = errors.New("page must be >= 1 and size in [1, 100]")
)

// CreateUserInput is the payload for Create.
type CreateUserInput struct {
	Name  string
	Email string
}

// UpdateUserInput carries partial updates.
type UpdateUserInput struct {
	Name  *string
	Email *string
}

// UserService is the demo business interface.
type UserService interface {
	Create(ctx context.Context, in CreateUserInput) (*model.User, error)
	Get(ctx context.Context, id uint64) (*model.User, error)
	List(ctx context.Context, page, size int) ([]model.User, int64, error)
	Update(ctx context.Context, id uint64, in UpdateUserInput) (*model.User, error)
	Delete(ctx context.Context, id uint64) error
	FindByEmail(ctx context.Context, email string) (*model.User, error)
}

type userService struct {
	repo repository.UserRepository
}

// NewUserService wires the GORM-backed repository (or a fake in tests).
func NewUserService(repo repository.UserRepository) UserService {
	return &userService{repo: repo}
}

func (s *userService) Create(ctx context.Context, in CreateUserInput) (*model.User, error) {
	if _, err := s.repo.GetByEmail(ctx, in.Email); err == nil {
		return nil, ErrEmailTaken
	} else if !errors.Is(err, repository.ErrNotFound) {
		return nil, fmt.Errorf("check email uniqueness: %w", err)
	}
	u := &model.User{Name: in.Name, Email: in.Email}
	if err := s.repo.Create(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

func (s *userService) Get(ctx context.Context, id uint64) (*model.User, error) {
	u, err := s.repo.GetByID(ctx, id)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, ErrUserNotFound
	}
	return u, err
}

func (s *userService) List(ctx context.Context, page, size int) ([]model.User, int64, error) {
	if page < 1 || size < 1 || size > 100 {
		return nil, 0, ErrInvalidPaging
	}
	return s.repo.List(ctx, (page-1)*size, size)
}

func (s *userService) Update(ctx context.Context, id uint64, in UpdateUserInput) (*model.User, error) {
	u, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if in.Email != nil && *in.Email != u.Email {
		if existing, err := s.repo.GetByEmail(ctx, *in.Email); err == nil && existing.ID != id {
			return nil, ErrEmailTaken
		} else if err != nil && !errors.Is(err, repository.ErrNotFound) {
			return nil, err
		}
		u.Email = *in.Email
	}
	if in.Name != nil {
		u.Name = *in.Name
	}
	if err := s.repo.Update(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

func (s *userService) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	u, err := s.repo.GetByEmail(ctx, email)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, ErrUserNotFound
	}
	return u, err
}

func (s *userService) Delete(ctx context.Context, id uint64) error {
	if err := s.repo.Delete(ctx, id); errors.Is(err, repository.ErrNotFound) {
		return ErrUserNotFound
	} else if err != nil {
		return err
	}
	return nil
}
