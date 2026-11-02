package user

import (
	"context"
	"errors"
	"rcon-scheduler-panel/internal/errs"
	"rcon-scheduler-panel/internal/model"
)

type Repository interface {
	GetByID(ctx context.Context, id int64) (*model.User, error)
	GetByUsername(ctx context.Context, username string) (*model.User, error)
	Create(ctx context.Context, user *model.User) error
	Count(ctx context.Context) (int64, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) Create(ctx context.Context, u *model.User) error {
	if err := s.repo.Create(ctx, u); err != nil {
		if errors.Is(err, ErrAlreadyExists) {
			return errs.ErrConflict.WithMessage("user already exists").WithErr(err)
		}
		return err
	}
	return nil
}

func (s *service) GetByID(ctx context.Context, id int64) (*model.User, error) {
	u, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, errs.ErrNotFound.WithMessage("user not found").WithErr(err)
		}
		return nil, err
	}
	return u, nil
}

func (s *service) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	u, err := s.repo.GetByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, errs.ErrNotFound.WithMessage("user not found").WithErr(err)
		}
		return nil, err
	}
	return u, nil
}

func (s *service) Count(ctx context.Context) (int64, error) {
	amount, err := s.repo.Count(ctx)
	if err != nil {
		return 0, err
	}
	return amount, nil
}