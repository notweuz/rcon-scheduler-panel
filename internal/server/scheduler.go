package server

import (
	"context"
	"errors"
	"fmt"
	"rcon-scheduler-panel/internal/config"
	"rcon-scheduler-panel/internal/domain/job"
	"rcon-scheduler-panel/internal/domain/server"
	"time"

	"github.com/notweuz/rcon-scheduler-core/scheduler"
	"github.com/rs/zerolog/log"
)

func NewScheduler(cfg *config.Config) (*scheduler.Scheduler, error) {
	loc, err := time.LoadLocation(cfg.Timezone)
	if err != nil {
		return nil, fmt.Errorf("invalid timezone %s: %w", cfg.Timezone, err)
	}

	s, err := scheduler.New(
		scheduler.WithLocation(loc),
		scheduler.WithLogger(log.Logger),
	)

	if err != nil {
		return nil, fmt.Errorf("failed while creating scheduler: %w", err)
	}
	return s, nil
}

func ResumeScheduler(ctx context.Context, s *scheduler.Scheduler, srvSVC server.Service, jobSVC job.Service) error {
	servers, err := srvSVC.GetAll(ctx)
	if err != nil {
		return fmt.Errorf("get all servers: %w", err)
	}
	if len(servers) == 0 {
		log.Info().Msg("no tasks to resume")
		return nil
	}

	var errs []error
	for _, srv := range servers {
		coreSRV := srv.ToCore()
		if err := s.Register(coreSRV); err != nil {
			errs = append(errs, fmt.Errorf("register server %s: %w", coreSRV.Address, err))
		}
	}

	if err := jobSVC.Resume(ctx); err != nil {
		errs = append(errs, fmt.Errorf("resume jobs: %w", err))
	}

	return errors.Join(errs...)
}
