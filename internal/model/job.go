package model

import (
	"fmt"
	"time"

	corejob "github.com/notweuz/rcon-scheduler-core/scheduler/job"
)

type Job struct {
	ID         int64 `gorm:"primaryKey"`
	ServerID   int64
	Server     *Server
	Name       string
	Cron       string
	LogResults bool

	Steps []Step `gorm:"serializer:json"`
}

type Step struct {
	Command string `json:"command,omitempty"`
	Wait    string `json:"wait,omitempty"`
}

func (Job) TableName() string {
	return "jobs"
}

func (j Job) ToCore() (corejob.Job, error) {
	steps := make([]corejob.Step, len(j.Steps))

	for i, s := range j.Steps {
		step := corejob.Step{Command: s.Command}

		if s.Wait != "" {
			d, err := time.ParseDuration(s.Wait)
			if err != nil {
				return corejob.Job{}, fmt.Errorf("steps[%d]: %w", i, err)
			}
			step.Wait = d
		}
		steps = append(steps, step)
	}
	result := corejob.New(
		j.Name,
		j.Cron,
		steps,
		corejob.WithLogResults(j.LogResults),
	)

	return result, result.Validate()
}
