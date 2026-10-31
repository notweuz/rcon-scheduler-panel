package server

import (
	"context"
	"errors"
	"fmt"
	"rcon-scheduler-panel/internal/config"
	"rcon-scheduler-panel/internal/domain/job"
	"rcon-scheduler-panel/internal/domain/server"
	"rcon-scheduler-panel/internal/model"
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

	allJobs, err := jobSVC.GetAll(ctx)
	if err != nil {
		return fmt.Errorf("get all jobs: %w", err)
	}

	jobsByServer := make(map[int64][]model.Job)
	for _, jb := range allJobs {
		jobsByServer[jb.ServerID] = append(jobsByServer[jb.ServerID], jb)
	}

	var errs []error
	for _, srv := range servers {
		coreSRV := srv.ToCore()
		if err := s.Register(coreSRV); err != nil {
			errs = append(errs, fmt.Errorf("register server %s: %w", coreSRV.Address, err))
			continue
		}

		srvKey := coreSRV.Key()
		for _, jb := range jobsByServer[srv.ID] {
			coreJob, err := jb.ToCore()
			if err != nil {
				errs = append(errs, fmt.Errorf("server %s: convert job %q to core: %w", coreSRV.Address, jb.Name, err))
				continue
			}

			if _, err := s.AddJob(srvKey, coreJob); err != nil {
				errs = append(errs, fmt.Errorf("server %s: add job %q: %w", coreSRV.Address, jb.Name, err))
				continue
			}
		}
	}

	return errors.Join(errs...)
}
