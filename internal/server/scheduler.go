package server

import (
	"fmt"
	"rcon-scheduler-panel/internal/config"
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
