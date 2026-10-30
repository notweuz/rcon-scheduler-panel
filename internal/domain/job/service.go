package job

import (
	"context"
	"errors"
	"rcon-scheduler-panel/internal/errs"
	"rcon-scheduler-panel/internal/model"
)

type Repository interface {
	Create(ctx context.Context, job *model.Job) error
	GetByID(ctx context.Context, id int64) (*model.Job, error)
	GetByServerID(ctx context.Context, serverID int64) ([]model.Job, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{
		repo: repo,
	}
}

func (s *service) Create(ctx context.Context, job *model.Job) error {
	if err := s.repo.Create(ctx, job); err != nil {
		if errors.Is(err, ErrAlreadyExists) {
			return errs.ErrConflict.WithMessage("job already exist").WithErr(err)
		}
		return err
	}
	return nil
}

func (s *service) GetByID(ctx context.Context, id int64) (*model.Job, error) {
	job, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, errs.ErrNotFound.WithMessage("job not found").WithErr(err)
		}
		return nil, err
	}
	return job, nil
}

func (s *service) GetByServerID(ctx context.Context, serverID int64) ([]model.Job, error) {
	jobs, err := s.repo.GetByServerID(ctx, serverID)
	if err != nil {
		return nil, err
	}

	if jobs == nil {
		jobs = []model.Job{}
	}

	return jobs, nil
}