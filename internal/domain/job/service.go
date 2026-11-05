package job

import (
	"context"
	"errors"
	"fmt"
	"rcon-scheduler-panel/internal/errs"
	"rcon-scheduler-panel/internal/model"
	"sync"

	"github.com/google/uuid"
	"github.com/notweuz/rcon-scheduler-core/scheduler"
)

type Repository interface {
	Create(ctx context.Context, job *model.Job) error
	GetByID(ctx context.Context, id int64, opts ...Option) (*model.Job, error)
	GetByServerID(ctx context.Context, serverID int64, opts ...Option) ([]model.Job, error)
	GetAll(ctx context.Context, opts ...Option) ([]model.Job, error)
	Update(ctx context.Context, job *model.Job) error
	Delete(ctx context.Context, id int64) error
}

type service struct {
	repo Repository
	sch  *scheduler.Scheduler

	mu        sync.Mutex
	scheduled map[int64]scheduledJob
}

type scheduledJob struct {
	serverID int64
	id       uuid.UUID
}

func NewService(repo Repository, sch *scheduler.Scheduler) Service {
	return &service{
		repo:      repo,
		sch:       sch,
		scheduled: make(map[int64]scheduledJob),
	}
}

func (s *service) Create(ctx context.Context, job *model.Job) error {
	if _, err := job.ToCore(); err != nil {
		return errs.ErrBadRequest.WithMessage(err.Error()).WithErr(err)
	}

	if err := s.repo.Create(ctx, job); err != nil {
		switch {
		case errors.Is(err, ErrAlreadyExists):
			return errs.ErrConflict.WithMessage("job already exist").WithErr(err)
		case errors.Is(err, ErrServerNotFound):
			return errs.ErrNotFound.WithMessage("server not found").WithErr(err)
		}
		return err
	}

	created, err := s.repo.GetByID(ctx, job.ID, WithPreloadServer())
	if err != nil {
		return s.rollbackCreate(ctx, job.ID, fmt.Errorf("load created job: %w", err))
	}

	if err := s.schedule(created); err != nil {
		return s.rollbackCreate(ctx, job.ID, err)
	}

	return nil
}

func (s *service) rollbackCreate(ctx context.Context, id int64, cause error) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return errors.Join(cause, fmt.Errorf("rollback job %d: %w", id, err))
	}
	return cause
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

func (s *service) GetAll(ctx context.Context) ([]model.Job, error) {
	jobs, err := s.repo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	if jobs == nil {
		jobs = []model.Job{}
	}

	return jobs, nil
}

func (s *service) Update(ctx context.Context, job *model.Job) error {
	if _, err := job.ToCore(); err != nil {
		return errs.ErrBadRequest.WithMessage(err.Error()).WithErr(err)
	}

	if err := s.repo.Update(ctx, job); err != nil {
		if errors.Is(err, ErrNotFound) {
			return errs.ErrNotFound.WithMessage("job not found").WithErr(err)
		}
		return err
	}

	updated, err := s.repo.GetByID(ctx, job.ID, WithPreloadServer())
	if err != nil {
		return fmt.Errorf("load updated job: %w", err)
	}

	if err := s.unschedule(job.ID); err != nil {
		return err
	}
	if err := s.schedule(updated); err != nil {
		return err
	}

	*job = *updated
	return nil
}

func (s *service) Delete(ctx context.Context, id int64) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		if errors.Is(err, ErrNotFound) {
			return errs.ErrNotFound.WithMessage("job not found").WithErr(err)
		}
		return err
	}

	return s.unschedule(id)
}

func (s *service) Resume(ctx context.Context) error {
	jobs, err := s.repo.GetAll(ctx, WithPreloadServer())
	if err != nil {
		return err
	}
	return s.scheduleAll(jobs)
}

func (s *service) ScheduleServer(ctx context.Context, serverID int64) error {
	jobs, err := s.repo.GetByServerID(ctx, serverID, WithPreloadServer())
	if err != nil {
		return err
	}
	return s.scheduleAll(jobs)
}

func (s *service) UnscheduleServer(serverID int64) error {
	var jobIDs []int64
	s.mu.Lock()
	for jobID, sj := range s.scheduled {
		if sj.serverID == serverID {
			jobIDs = append(jobIDs, jobID)
		}
	}
	s.mu.Unlock()

	var failed []error
	for _, jobID := range jobIDs {
		if err := s.unschedule(jobID); err != nil {
			failed = append(failed, err)
		}
	}
	return errors.Join(failed...)
}

func (s *service) scheduleAll(jobs []model.Job) error {
	var failed []error
	for _, jb := range jobs {
		if err := s.schedule(&jb); err != nil {
			failed = append(failed, err)
		}
	}
	return errors.Join(failed...)
}

func (s *service) schedule(job *model.Job) error {
	coreJob, err := job.ToCore()
	if err != nil {
		return fmt.Errorf("convert job %q to core: %w", job.Name, err)
	}

	id, err := s.sch.AddJob(job.Server.ToCore().Key(), coreJob)
	if err != nil {
		return fmt.Errorf("add job %q to scheduler: %w", job.Name, err)
	}

	s.mu.Lock()
	prev, wasScheduled := s.scheduled[job.ID]
	s.scheduled[job.ID] = scheduledJob{serverID: job.ServerID, id: id}
	s.mu.Unlock()

	if wasScheduled {
		if err := s.sch.RemoveJob(prev.id); err != nil && !errors.Is(err, scheduler.ErrJobNotFound) {
			return fmt.Errorf("remove previous copy of job %d from scheduler: %w", job.ID, err)
		}
	}
	return nil
}

func (s *service) unschedule(jobID int64) error {
	s.mu.Lock()
	sj, ok := s.scheduled[jobID]
	delete(s.scheduled, jobID)
	s.mu.Unlock()

	if !ok {
		return nil
	}

	if err := s.sch.RemoveJob(sj.id); err != nil && !errors.Is(err, scheduler.ErrJobNotFound) {
		return fmt.Errorf("remove job %d from scheduler: %w", jobID, err)
	}
	return nil
}
