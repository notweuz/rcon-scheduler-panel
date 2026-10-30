package server

import (
	"context"
	"errors"
	"fmt"
	"rcon-scheduler-panel/internal/errs"
	"rcon-scheduler-panel/internal/model"

	"github.com/notweuz/rcon-scheduler-core/scheduler"
	coreserver "github.com/notweuz/rcon-scheduler-core/server"
)

type Repository interface {
	Create(ctx context.Context, srv *model.Server) error
	GetByID(ctx context.Context, id int64, opts ...Option) (*model.Server, error)
	GetAll(ctx context.Context, opts ...Option) ([]model.Server, error)
}

type service struct {
	repo Repository
	sch  *scheduler.Scheduler
}

func NewService(repo Repository, sch *scheduler.Scheduler) Service {
	return &service{repo: repo, sch: sch}
}

func (s *service) Create(ctx context.Context, srv *model.Server) error {
	if err := s.repo.Create(ctx, srv); err != nil {
		if errors.Is(err, ErrAlreadyExists) {
			return errs.ErrConflict.WithMessage("server already exist").WithErr(err)
		}
		return err
	}

	if err := s.sch.Register(srv.ToCore()); err != nil {
		switch {
		case errors.Is(err, scheduler.ErrServerExists):
			return errs.ErrConflict.
				WithMessage("server already registered").
				WithErr(err)
		case errors.Is(err, coreserver.ErrEmptyAddress),
			errors.Is(err, coreserver.ErrInvalidPort),
			errors.Is(err, scheduler.ErrUnsupportedType):
			return errs.ErrBadRequest.WithMessage(err.Error()).WithErr(err)
		default:
			return fmt.Errorf("register server in scheduler: %w", err)
		}
	}
	
	return nil
}

func (s *service) GetByID(ctx context.Context, id int64) (*model.Server, error) {
	srv, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, errs.ErrNotFound.WithMessage("server not found").WithErr(err)
		}
		return nil, err
	}
	return srv, nil
}

func (s *service) GetAll(ctx context.Context) ([]model.Server, error) {
	servers, err := s.repo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	if servers == nil {
		servers = []model.Server{}
	}

	return servers, nil
}
