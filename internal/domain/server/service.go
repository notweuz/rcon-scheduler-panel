package server

import (
	"context"
	"errors"
	"fmt"
	"rcon-scheduler-panel/internal/errs"
	"rcon-scheduler-panel/internal/model"
	"time"

	"github.com/notweuz/rcon-scheduler-core/scheduler"
	coreserver "github.com/notweuz/rcon-scheduler-core/server"
	"github.com/rs/zerolog/log"
)

// fix core blocking server for ~20s if it's not available on command execution
const executeTimeout = 5 * time.Second

type Repository interface {
	Create(ctx context.Context, srv *model.Server) error
	GetByID(ctx context.Context, id int64, opts ...Option) (*model.Server, error)
	GetAll(ctx context.Context, opts ...Option) ([]model.Server, error)
	Delete(ctx context.Context, id int64, opts ...Option) (*model.Server, error)
	Update(ctx context.Context, srv *model.Server) error
}

type JobScheduler interface {
	ScheduleServer(ctx context.Context, serverID int64) error
	UnscheduleServer(serverID int64) error
}

type service struct {
	repo Repository
	sch  *scheduler.Scheduler
	jobs JobScheduler
}

func NewService(repo Repository, sch *scheduler.Scheduler, jobs JobScheduler) Service {
	return &service{repo: repo, sch: sch, jobs: jobs}
}

func (s *service) Create(ctx context.Context, srv *model.Server) error {
	if err := s.repo.Create(ctx, srv); err != nil {
		if errors.Is(err, ErrAlreadyExists) {
			return errs.ErrConflict.WithMessage("server already exist").WithErr(err)
		}
		return err
	}

	created, err := s.repo.GetByID(ctx, srv.ID, WithPreloadProtocol())
	if err != nil {
		return s.rollbackCreate(ctx, srv.ID, fmt.Errorf("load created server: %w", err))
	}
	srv.Protocol = created.Protocol

	if err := s.register(srv); err != nil {
		return s.rollbackCreate(ctx, srv.ID, err)
	}

	return nil
}

func (s *service) rollbackCreate(ctx context.Context, id int64, cause error) error {
	if _, err := s.repo.Delete(ctx, id); err != nil {
		return errors.Join(cause, fmt.Errorf("rollback server %d: %w", id, err))
	}
	return cause
}

func (s *service) GetByID(ctx context.Context, id int64) (*model.Server, error) {
	return s.getByID(ctx, id, WithPreloadProtocol())
}

func (s *service) GetAll(ctx context.Context) ([]model.Server, error) {
	servers, err := s.repo.GetAll(ctx, WithPreloadProtocol())
	if err != nil {
		return nil, err
	}

	if servers == nil {
		servers = []model.Server{}
	}

	return servers, nil
}

func (s *service) Delete(ctx context.Context, id int64) error {
	srv, err := s.repo.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return errs.ErrNotFound.WithMessage("server not found").WithErr(err)
		}
		return err
	}

	if err := s.jobs.UnscheduleServer(id); err != nil {
		return fmt.Errorf("unschedule jobs of server %d: %w", id, err)
	}

	if err := s.sch.Unregister(srv.ToCore().Key()); err != nil {
		if errors.Is(err, scheduler.ErrServerNotFound) {
			log.Warn().Err(err).Int64("server_id", id).Msg("deleted server wasn't registered in scheduler")
			return nil
		}
		return fmt.Errorf("unregister server in scheduler: %w", err)
	}

	return nil
}

func (s *service) Update(ctx context.Context, srv *model.Server) error {
	old, err := s.repo.GetByID(ctx, srv.ID, WithPreloadProtocol())
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return errs.ErrNotFound.WithMessage("server not found").WithErr(err)
		}
		return err
	}

	if srv.Password == "" {
		srv.Password = old.Password
	}

	if err := s.repo.Update(ctx, srv); err != nil {
		switch {
		case errors.Is(err, ErrNotFound):
			return errs.ErrNotFound.WithMessage("server not found").WithErr(err)
		case errors.Is(err, ErrProtocolNotFound):
			return errs.ErrNotFound.WithMessage("protocol not found").WithErr(err)
		}
		return err
	}

	updated, err := s.repo.GetByID(ctx, srv.ID, WithPreloadProtocol())
	if err != nil {
		return s.rollbackUpdate(ctx, old, fmt.Errorf("load updated server: %w", err))
	}

	if err := s.jobs.UnscheduleServer(old.ID); err != nil {
		return fmt.Errorf("unschedule jobs of server %d: %w", old.ID, err)
	}
	if err := s.sch.Unregister(old.ToCore().Key()); err != nil && !errors.Is(err, scheduler.ErrServerNotFound) {
		return fmt.Errorf("unregister server in scheduler: %w", err)
	}

	if err := s.register(updated); err != nil {
		return s.rollbackUpdate(ctx, old, err)
	}
	if err := s.jobs.ScheduleServer(ctx, updated.ID); err != nil {
		return fmt.Errorf("schedule jobs of server %d: %w", updated.ID, err)
	}

	*srv = *updated
	return nil
}

func (s *service) Execute(ctx context.Context, id int64, command string) (string, error) {
	srv, err := s.getByID(ctx, id)
	if err != nil {
		return "", fmt.Errorf("execute command for %d server: %w", id, err)
	}

	ctx, cancel := context.WithTimeout(ctx, executeTimeout)
	defer cancel()

	response, err := s.sch.Execute(ctx, srv.ToCore().Key(), command)
	if err != nil {
		if errors.Is(err, scheduler.ErrServerNotFound) {
			return "", fmt.Errorf("execute command on server %d: %w", id, err)
		}
		return "", errs.ErrBadGateway.
			WithMessage("game server is unavailable or rejected the command").
			WithErr(err)
	}

	return response, nil
}

func (s *service) rollbackUpdate(ctx context.Context, old *model.Server, cause error) error {
	failed := []error{cause}
	if err := s.repo.Update(ctx, old); err != nil {
		failed = append(failed, fmt.Errorf("rollback server %d: %w", old.ID, err))
	}
	if err := s.sch.Register(old.ToCore()); err != nil && !errors.Is(err, scheduler.ErrServerExists) {
		failed = append(failed, fmt.Errorf("register old server %d: %w", old.ID, err))
	}
	if err := s.jobs.ScheduleServer(ctx, old.ID); err != nil {
		failed = append(failed, fmt.Errorf("schedule jobs of old server %d: %w", old.ID, err))
	}
	return errors.Join(failed...)
}

func (s *service) register(srv *model.Server) error {
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

func (s *service) getByID(ctx context.Context, id int64, opts ...Option) (*model.Server, error) {
	srv, err := s.repo.GetByID(ctx, id, opts...)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, errs.ErrNotFound.WithMessage("server not found").WithErr(err)
		}
		return nil, err
	}
	return srv, nil
}
